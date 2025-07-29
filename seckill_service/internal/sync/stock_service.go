package sync

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// StockSyncService 库存同步服务
// StockSyncService stock synchronization service
type StockSyncService struct {
	syncer  *StockSyncer
	logger  logger.Logger
	config  *StockSyncConfig
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex
	metrics *StockSyncMetrics
}

// StockSyncMetrics 库存同步指标
// StockSyncMetrics stock sync metrics
type StockSyncMetrics struct {
	mu               sync.RWMutex
	TotalSynced      int64            `json:"total_synced"`
	SuccessCount     int64            `json:"success_count"`
	ErrorCount       int64            `json:"error_count"`
	ConflictCount    int64            `json:"conflict_count"`
	LastSyncTime     time.Time        `json:"last_sync_time"`
	AvgSyncDuration  time.Duration    `json:"avg_sync_duration"`
	LastSyncDuration time.Duration    `json:"last_sync_duration"`
	ConflictsByType  map[string]int64 `json:"conflicts_by_type"`
}

// NewStockSyncService 创建库存同步服务
// NewStockSyncService creates stock sync service
func NewStockSyncService(db *sql.DB, redisClient redis.Client, config *StockSyncConfig, log logger.Logger) *StockSyncService {
	if config == nil {
		config = DefaultStockSyncConfig()
	}

	syncer := NewStockSyncer(db, redisClient, config, log)
	ctx, cancel := context.WithCancel(context.Background())

	return &StockSyncService{
		syncer: syncer,
		logger: log,
		config: config,
		ctx:    ctx,
		cancel: cancel,
		metrics: &StockSyncMetrics{
			ConflictsByType: make(map[string]int64),
		},
	}
}

// Start 启动库存同步服务
// Start starts stock sync service
func (s *StockSyncService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.logger.Info("Starting stock sync service",
		logger.Duration("interval", s.config.SyncInterval))

	// 立即执行一次同步
	// Execute sync immediately
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.performSync(); err != nil {
			s.logger.Error("Initial stock sync failed", logger.Error(err))
		}
	}()

	// 启动定期同步
	// Start periodic sync
	s.wg.Add(1)
	go s.periodicSync()

	return nil
}

// Stop 停止库存同步服务
// Stop stops stock sync service
func (s *StockSyncService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping stock sync service")
	s.running = false
	s.cancel()
	s.wg.Wait()

	s.logger.Info("Stock sync service stopped")
	return nil
}

// periodicSync 定期同步
// periodicSync periodic synchronization
func (s *StockSyncService) periodicSync() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.performSync(); err != nil {
				s.logger.Error("Periodic stock sync failed", logger.Error(err))
			}
		case <-s.ctx.Done():
			return
		}
	}
}

// performSync 执行同步
// performSync performs synchronization
func (s *StockSyncService) performSync() error {
	startTime := time.Now()
	s.logger.Debug("Starting stock sync")

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	results, err := s.syncer.SyncAllStocks(ctx)
	duration := time.Since(startTime)

	// 更新指标
	// Update metrics
	s.updateMetrics(results, duration)

	if err != nil {
		s.logger.Error("Stock sync failed",
			logger.Error(err),
			logger.Duration("duration", duration))
		return err
	}

	s.logger.Debug("Stock sync completed",
		logger.Int("results_count", len(results)),
		logger.Duration("duration", duration))

	return nil
}

// SyncNow 立即同步
// SyncNow synchronizes immediately
func (s *StockSyncService) SyncNow() error {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return fmt.Errorf("stock sync service is not running")
	}
	s.mu.RUnlock()

	s.logger.Info("Manual stock sync requested")
	return s.performSync()
}

// SyncActivity 同步指定活动的库存
// SyncActivity synchronizes specific activity stock
func (s *StockSyncService) SyncActivity(activityID string) (*StockSyncResult, error) {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return nil, fmt.Errorf("stock sync service is not running")
	}
	s.mu.RUnlock()

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	result, err := s.syncer.SyncSingleStock(ctx, activityID)
	if err != nil {
		s.logger.Error("Failed to sync activity stock",
			logger.String("activity_id", activityID),
			logger.Error(err))
		return result, err
	}

	// 更新指标
	// Update metrics
	s.updateSingleMetrics(result)

	s.logger.Info("Activity stock synchronized",
		logger.String("activity_id", activityID),
		logger.Bool("success", result.Success))

	return result, nil
}

// updateMetrics 更新指标
// updateMetrics updates metrics
func (s *StockSyncService) updateMetrics(results []*StockSyncResult, duration time.Duration) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.LastSyncTime = time.Now()
	s.metrics.LastSyncDuration = duration

	// 计算平均同步时间
	// Calculate average sync duration
	if s.metrics.TotalSynced == 0 {
		s.metrics.AvgSyncDuration = duration
	} else {
		totalDuration := time.Duration(s.metrics.TotalSynced) * s.metrics.AvgSyncDuration
		s.metrics.AvgSyncDuration = (totalDuration + duration) / time.Duration(s.metrics.TotalSynced+1)
	}

	// 统计结果
	// Count results
	for _, result := range results {
		s.metrics.TotalSynced++
		if result.Success {
			s.metrics.SuccessCount++
		} else {
			s.metrics.ErrorCount++
		}

		if result.ConflictType != "" {
			s.metrics.ConflictCount++
			s.metrics.ConflictsByType[result.ConflictType]++
		}
	}
}

// updateSingleMetrics 更新单个指标
// updateSingleMetrics updates single metrics
func (s *StockSyncService) updateSingleMetrics(result *StockSyncResult) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalSynced++
	if result.Success {
		s.metrics.SuccessCount++
	} else {
		s.metrics.ErrorCount++
	}

	if result.ConflictType != "" {
		s.metrics.ConflictCount++
		s.metrics.ConflictsByType[result.ConflictType]++
	}
}

// GetMetrics 获取同步指标
// GetMetrics gets sync metrics
func (s *StockSyncService) GetMetrics() StockSyncMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// 复制冲突类型统计
	// Copy conflict type stats
	conflictsByType := make(map[string]int64)
	for k, v := range s.metrics.ConflictsByType {
		conflictsByType[k] = v
	}

	return StockSyncMetrics{
		TotalSynced:      s.metrics.TotalSynced,
		SuccessCount:     s.metrics.SuccessCount,
		ErrorCount:       s.metrics.ErrorCount,
		ConflictCount:    s.metrics.ConflictCount,
		LastSyncTime:     s.metrics.LastSyncTime,
		AvgSyncDuration:  s.metrics.AvgSyncDuration,
		LastSyncDuration: s.metrics.LastSyncDuration,
		ConflictsByType:  conflictsByType,
	}
}

// IsRunning 检查服务是否运行中
// IsRunning checks if service is running
func (s *StockSyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// ResetMetrics 重置指标
// ResetMetrics resets metrics
func (s *StockSyncService) ResetMetrics() {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalSynced = 0
	s.metrics.SuccessCount = 0
	s.metrics.ErrorCount = 0
	s.metrics.ConflictCount = 0
	s.metrics.ConflictsByType = make(map[string]int64)

	s.logger.Info("Stock sync metrics reset")
}
