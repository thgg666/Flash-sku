package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
)

// CacheMonitor 缓存监控器
// CacheMonitor cache monitor
type CacheMonitor struct {
	cacheManager *CacheManager
	logger       logger.Logger
	metrics      *AtomicMetrics
	config       *MonitorConfig
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	running      bool
	mu           sync.RWMutex
}

// AtomicMetrics 原子指标
// AtomicMetrics atomic metrics
type AtomicMetrics struct {
	hitCount    int64
	missCount   int64
	setCount    int64
	deleteCount int64
	errorCount  int64
}

// MonitorConfig 监控配置
// MonitorConfig monitor configuration
type MonitorConfig struct {
	UpdateInterval time.Duration `json:"update_interval"`
	AlertThreshold float64       `json:"alert_threshold"` // 命中率告警阈值
	EnableAlerts   bool          `json:"enable_alerts"`
}

// NewCacheMonitor 创建缓存监控器
// NewCacheMonitor creates cache monitor
func NewCacheMonitor(cacheManager *CacheManager, config *MonitorConfig, log logger.Logger) *CacheMonitor {
	if config == nil {
		config = &MonitorConfig{
			UpdateInterval: 30 * time.Second,
			AlertThreshold: 0.8,
			EnableAlerts:   true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &CacheMonitor{
		cacheManager: cacheManager,
		logger:       log,
		metrics:      &AtomicMetrics{},
		ctx:          ctx,
		cancel:       cancel,
		config:       config,
	}
}

// Start 启动监控
// Start starts monitoring
func (m *CacheMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	m.logger.Info("Starting cache monitor")

	m.wg.Add(1)
	go m.monitorLoop()
}

// Stop 停止监控
// Stop stops monitoring
func (m *CacheMonitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	m.logger.Info("Stopping cache monitor")
	m.running = false
	m.cancel()
	m.wg.Wait()
	m.logger.Info("Cache monitor stopped")
}

// monitorLoop 监控循环
// monitorLoop monitoring loop
func (m *CacheMonitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateMetrics()
		case <-m.ctx.Done():
			return
		}
	}
}

// updateMetrics 更新指标
// updateMetrics updates metrics
func (m *CacheMonitor) updateMetrics() {
	// 获取当前指标
	// Get current metrics
	hitCount := atomic.LoadInt64(&m.metrics.hitCount)
	missCount := atomic.LoadInt64(&m.metrics.missCount)
	setCount := atomic.LoadInt64(&m.metrics.setCount)
	deleteCount := atomic.LoadInt64(&m.metrics.deleteCount)
	errorCount := atomic.LoadInt64(&m.metrics.errorCount)

	// 创建指标对象
	// Create metrics object
	metrics := &CacheMetrics{
		HitCount:    hitCount,
		MissCount:   missCount,
		SetCount:    setCount,
		DeleteCount: deleteCount,
		ErrorCount:  errorCount,
		LastUpdated: time.Now(),
	}

	// 计算命中率
	// Calculate hit rate
	total := hitCount + missCount
	if total > 0 {
		metrics.HitRate = float64(hitCount) / float64(total)
	}

	// 更新到缓存
	// Update to cache
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	if err := m.cacheManager.UpdateMetrics(ctx, metrics); err != nil {
		m.logger.Error("Failed to update cache metrics", logger.Error(err))
		atomic.AddInt64(&m.metrics.errorCount, 1)
	}

	// 记录指标日志
	// Log metrics
	m.logger.Debug("Cache metrics updated",
		logger.Int64("hit_count", hitCount),
		logger.Int64("miss_count", missCount),
		logger.Float64("hit_rate", metrics.HitRate),
		logger.Int64("error_count", errorCount))

	// 检查告警
	// Check alerts
	m.checkAlerts(metrics)
}

// checkAlerts 检查告警
// checkAlerts checks alerts
func (m *CacheMonitor) checkAlerts(metrics *CacheMetrics) {
	if !m.config.EnableAlerts {
		return
	}

	// 命中率过低告警
	// Low hit rate alert
	if metrics.HitRate < m.config.AlertThreshold && metrics.HitCount+metrics.MissCount > 100 {
		m.logger.Warn("Cache hit rate is low",
			logger.Float64("hit_rate", metrics.HitRate),
			logger.Float64("threshold", m.config.AlertThreshold),
			logger.Int64("total_requests", metrics.HitCount+metrics.MissCount))
	}

	// 错误率过高告警
	// High error rate alert
	totalOps := metrics.HitCount + metrics.MissCount + metrics.SetCount + metrics.DeleteCount
	if totalOps > 0 {
		errorRate := float64(metrics.ErrorCount) / float64(totalOps)
		if errorRate > 0.05 { // 5% 错误率
			m.logger.Warn("Cache error rate is high",
				logger.Float64("error_rate", errorRate),
				logger.Int64("error_count", metrics.ErrorCount),
				logger.Int64("total_ops", totalOps))
		}
	}
}

// RecordHit 记录缓存命中
// RecordHit records cache hit
func (m *CacheMonitor) RecordHit() {
	atomic.AddInt64(&m.metrics.hitCount, 1)
}

// RecordMiss 记录缓存未命中
// RecordMiss records cache miss
func (m *CacheMonitor) RecordMiss() {
	atomic.AddInt64(&m.metrics.missCount, 1)
}

// RecordSet 记录缓存设置
// RecordSet records cache set
func (m *CacheMonitor) RecordSet() {
	atomic.AddInt64(&m.metrics.setCount, 1)
}

// RecordDelete 记录缓存删除
// RecordDelete records cache delete
func (m *CacheMonitor) RecordDelete() {
	atomic.AddInt64(&m.metrics.deleteCount, 1)
}

// RecordError 记录缓存错误
// RecordError records cache error
func (m *CacheMonitor) RecordError() {
	atomic.AddInt64(&m.metrics.errorCount, 1)
}

// GetCurrentMetrics 获取当前指标
// GetCurrentMetrics gets current metrics
func (m *CacheMonitor) GetCurrentMetrics() *CacheMetrics {
	hitCount := atomic.LoadInt64(&m.metrics.hitCount)
	missCount := atomic.LoadInt64(&m.metrics.missCount)
	setCount := atomic.LoadInt64(&m.metrics.setCount)
	deleteCount := atomic.LoadInt64(&m.metrics.deleteCount)
	errorCount := atomic.LoadInt64(&m.metrics.errorCount)

	metrics := &CacheMetrics{
		HitCount:    hitCount,
		MissCount:   missCount,
		SetCount:    setCount,
		DeleteCount: deleteCount,
		ErrorCount:  errorCount,
		LastUpdated: time.Now(),
	}

	// 计算命中率
	// Calculate hit rate
	total := hitCount + missCount
	if total > 0 {
		metrics.HitRate = float64(hitCount) / float64(total)
	}

	return metrics
}

// ResetMetrics 重置指标
// ResetMetrics resets metrics
func (m *CacheMonitor) ResetMetrics() {
	atomic.StoreInt64(&m.metrics.hitCount, 0)
	atomic.StoreInt64(&m.metrics.missCount, 0)
	atomic.StoreInt64(&m.metrics.setCount, 0)
	atomic.StoreInt64(&m.metrics.deleteCount, 0)
	atomic.StoreInt64(&m.metrics.errorCount, 0)

	m.logger.Info("Cache metrics reset")
}

// IsRunning 检查是否运行中
// IsRunning checks if running
func (m *CacheMonitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
