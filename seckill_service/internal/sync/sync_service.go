package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// SyncService 同步服务
// SyncService synchronization service
type SyncService struct {
	syncer  *DataSyncer
	logger  logger.Logger
	config  *SyncConfig
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex
}

// NewSyncService 创建同步服务
// NewSyncService creates synchronization service
func NewSyncService(config *SyncConfig, redisClient redis.Client, log logger.Logger) (*SyncService, error) {
	syncer, err := NewDataSyncer(config, redisClient, log)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &SyncService{
		syncer: syncer,
		logger: log,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start 启动同步服务
// Start starts synchronization service
func (s *SyncService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.logger.Info("Starting sync service",
		logger.Duration("interval", s.config.SyncInterval))

	// 立即执行一次同步
	// Execute sync immediately
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.syncer.SyncAllActivities(s.ctx); err != nil {
			s.logger.Error("Initial sync failed", logger.Error(err))
		}
	}()

	// 启动定期同步
	// Start periodic sync
	s.wg.Add(1)
	go s.periodicSync()

	return nil
}

// Stop 停止同步服务
// Stop stops synchronization service
func (s *SyncService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("Stopping sync service")
	s.running = false
	s.cancel()
	s.wg.Wait()

	if err := s.syncer.Close(); err != nil {
		s.logger.Error("Failed to close syncer", logger.Error(err))
		return err
	}

	s.logger.Info("Sync service stopped")
	return nil
}

// periodicSync 定期同步
// periodicSync periodic synchronization
func (s *SyncService) periodicSync() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performSync()
		case <-s.ctx.Done():
			return
		}
	}
}

// performSync 执行同步
// performSync performs synchronization
func (s *SyncService) performSync() {
	s.logger.Debug("Starting periodic sync")

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	if err := s.syncer.SyncAllActivities(ctx); err != nil {
		s.logger.Error("Periodic sync failed", logger.Error(err))
	} else {
		s.logger.Debug("Periodic sync completed successfully")
	}
}

// SyncNow 立即同步
// SyncNow synchronizes immediately
func (s *SyncService) SyncNow() error {
	s.mu.RLock()
	if !s.running {
		s.mu.RUnlock()
		return fmt.Errorf("sync service is not running")
	}
	s.mu.RUnlock()

	s.logger.Info("Manual sync requested")

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	return s.syncer.SyncAllActivities(ctx)
}

// GetMetrics 获取同步指标
// GetMetrics gets synchronization metrics
func (s *SyncService) GetMetrics(ctx context.Context) (*SyncMetrics, error) {
	metricsKey := "seckill:sync:metrics"

	metricsJSON, err := s.syncer.redisClient.Get(ctx, metricsKey)
	if err != nil {
		return nil, err
	}

	var metrics SyncMetrics
	if err := json.Unmarshal([]byte(metricsJSON), &metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}

// IsRunning 检查服务是否运行中
// IsRunning checks if service is running
func (s *SyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// DefaultSyncConfig 默认同步配置
// DefaultSyncConfig default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		SyncInterval:    5 * time.Minute,
		BatchSize:       100,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		CacheExpiration: 24 * time.Hour,
	}
}
