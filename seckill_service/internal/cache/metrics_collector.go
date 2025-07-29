package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// MetricsCollector 指标收集器
// MetricsCollector metrics collector
type MetricsCollector struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *MetricsConfig
	metrics     *CacheMetricsData
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
}

// MetricsConfig 指标配置
// MetricsConfig metrics configuration
type MetricsConfig struct {
	CollectInterval   time.Duration   `json:"collect_interval"`
	RetentionPeriod   time.Duration   `json:"retention_period"`
	EnableDetailedLog bool            `json:"enable_detailed_log"`
	AlertThresholds   AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds 告警阈值
// AlertThresholds alert thresholds
type AlertThresholds struct {
	LowHitRate    float64       `json:"low_hit_rate"`    // 低命中率阈值
	HighErrorRate float64       `json:"high_error_rate"` // 高错误率阈值
	LowStockAlert int           `json:"low_stock_alert"` // 低库存告警
	HighLatency   time.Duration `json:"high_latency"`    // 高延迟告警
}

// CacheMetricsData 缓存指标数据
// CacheMetricsData cache metrics data
type CacheMetricsData struct {
	mu sync.RWMutex

	// 基础指标
	// Basic metrics
	HitCount    int64 `json:"hit_count"`
	MissCount   int64 `json:"miss_count"`
	SetCount    int64 `json:"set_count"`
	DeleteCount int64 `json:"delete_count"`
	ErrorCount  int64 `json:"error_count"`

	// 性能指标
	// Performance metrics
	AvgLatency     time.Duration `json:"avg_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	TotalLatency   time.Duration `json:"total_latency"`
	OperationCount int64         `json:"operation_count"`

	// 业务指标
	// Business metrics
	StockMetrics    map[string]*StockMetric    `json:"stock_metrics"`
	ActivityMetrics map[string]*ActivityMetric `json:"activity_metrics"`

	// 时间戳
	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// StockMetric 库存指标
// StockMetric stock metric
type StockMetric struct {
	ActivityID   string    `json:"activity_id"`
	CurrentStock int       `json:"current_stock"`
	InitialStock int       `json:"initial_stock"`
	SoldCount    int       `json:"sold_count"`
	LastUpdated  time.Time `json:"last_updated"`
	UpdateCount  int64     `json:"update_count"`
}

// ActivityMetric 活动指标
// ActivityMetric activity metric
type ActivityMetric struct {
	ActivityID      string        `json:"activity_id"`
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	FailureCount    int64         `json:"failure_count"`
	LastAccess      time.Time     `json:"last_access"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}

// MetricsSnapshot 指标快照
// MetricsSnapshot metrics snapshot
type MetricsSnapshot struct {
	Timestamp        time.Time        `json:"timestamp"`
	HitRate          float64          `json:"hit_rate"`
	ErrorRate        float64          `json:"error_rate"`
	OperationsPerSec float64          `json:"operations_per_sec"`
	AvgLatency       time.Duration    `json:"avg_latency"`
	StockSummary     *StockSummary    `json:"stock_summary"`
	ActivitySummary  *ActivitySummary `json:"activity_summary"`
	Alerts           []Alert          `json:"alerts"`
}

// StockSummary 库存汇总
// StockSummary stock summary
type StockSummary struct {
	TotalActivities int     `json:"total_activities"`
	LowStockCount   int     `json:"low_stock_count"`
	OutOfStockCount int     `json:"out_of_stock_count"`
	TotalSoldCount  int     `json:"total_sold_count"`
	AvgStockLevel   float64 `json:"avg_stock_level"`
}

// ActivitySummary 活动汇总
// ActivitySummary activity summary
type ActivitySummary struct {
	ActiveCount     int           `json:"active_count"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRate     float64       `json:"success_rate"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}

// Alert 告警
// Alert alert
type Alert struct {
	Type      string    `json:"type"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// NewMetricsCollector 创建指标收集器
// NewMetricsCollector creates metrics collector
func NewMetricsCollector(redisClient redis.Client, config *MetricsConfig, log logger.Logger) *MetricsCollector {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MetricsCollector{
		redisClient: redisClient,
		logger:      log,
		config:      config,
		metrics: &CacheMetricsData{
			StockMetrics:    make(map[string]*StockMetric),
			ActivityMetrics: make(map[string]*ActivityMetric),
			StartTime:       time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动指标收集
// Start starts metrics collection
func (c *MetricsCollector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return
	}

	c.running = true
	c.logger.Info("Starting cache metrics collector",
		logger.Duration("interval", c.config.CollectInterval))

	c.wg.Add(1)
	go c.collectLoop()
}

// Stop 停止指标收集
// Stop stops metrics collection
func (c *MetricsCollector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.logger.Info("Stopping cache metrics collector")
	c.running = false
	c.cancel()
	c.wg.Wait()
	c.logger.Info("Cache metrics collector stopped")
}

// collectLoop 收集循环
// collectLoop collection loop
func (c *MetricsCollector) collectLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.collectMetrics()
		case <-c.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集指标
// collectMetrics collects metrics
func (c *MetricsCollector) collectMetrics() {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	c.logger.Debug("Collecting cache metrics")

	// 收集库存指标
	// Collect stock metrics
	c.collectStockMetrics(ctx)

	// 收集活动指标
	// Collect activity metrics
	c.collectActivityMetrics(ctx)

	// 更新时间戳
	// Update timestamp
	c.metrics.mu.Lock()
	c.metrics.LastUpdated = time.Now()
	c.metrics.mu.Unlock()

	// 生成快照和告警
	// Generate snapshot and alerts
	snapshot := c.GenerateSnapshot()
	c.processAlerts(snapshot.Alerts)

	if c.config.EnableDetailedLog {
		c.logDetailedMetrics(snapshot)
	}
}

// collectStockMetrics 收集库存指标
// collectStockMetrics collects stock metrics
func (c *MetricsCollector) collectStockMetrics(ctx context.Context) {
	// 获取所有库存键
	// Get all stock keys
	pattern := "seckill:stock:*"
	// 注意：这里简化实现，实际应该使用SCAN命令
	// Note: Simplified implementation, should use SCAN command in practice

	c.logger.Debug("Collecting stock metrics", logger.String("pattern", pattern))

	// 这里应该实现具体的库存指标收集逻辑
	// Should implement specific stock metrics collection logic here
}

// collectActivityMetrics 收集活动指标
// collectActivityMetrics collects activity metrics
func (c *MetricsCollector) collectActivityMetrics(ctx context.Context) {
	// 获取所有活动键
	// Get all activity keys
	pattern := "seckill:activity:*"

	c.logger.Debug("Collecting activity metrics", logger.String("pattern", pattern))

	// 这里应该实现具体的活动指标收集逻辑
	// Should implement specific activity metrics collection logic here
}

// RecordHit 记录缓存命中
// RecordHit records cache hit
func (c *MetricsCollector) RecordHit(latency time.Duration) {
	atomic.AddInt64(&c.metrics.HitCount, 1)
	c.updateLatency(latency)
}

// RecordMiss 记录缓存未命中
// RecordMiss records cache miss
func (c *MetricsCollector) RecordMiss(latency time.Duration) {
	atomic.AddInt64(&c.metrics.MissCount, 1)
	c.updateLatency(latency)
}

// RecordSet 记录缓存设置
// RecordSet records cache set
func (c *MetricsCollector) RecordSet(latency time.Duration) {
	atomic.AddInt64(&c.metrics.SetCount, 1)
	c.updateLatency(latency)
}

// RecordError 记录错误
// RecordError records error
func (c *MetricsCollector) RecordError() {
	atomic.AddInt64(&c.metrics.ErrorCount, 1)
}

// updateLatency 更新延迟指标
// updateLatency updates latency metrics
func (c *MetricsCollector) updateLatency(latency time.Duration) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.OperationCount++
	c.metrics.TotalLatency += latency

	if c.metrics.OperationCount == 1 {
		c.metrics.MinLatency = latency
		c.metrics.MaxLatency = latency
	} else {
		if latency < c.metrics.MinLatency {
			c.metrics.MinLatency = latency
		}
		if latency > c.metrics.MaxLatency {
			c.metrics.MaxLatency = latency
		}
	}

	c.metrics.AvgLatency = c.metrics.TotalLatency / time.Duration(c.metrics.OperationCount)
}

// GenerateSnapshot 生成指标快照
// GenerateSnapshot generates metrics snapshot
func (c *MetricsCollector) GenerateSnapshot() *MetricsSnapshot {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Alerts:    make([]Alert, 0),
	}

	// 计算命中率
	// Calculate hit rate
	totalRequests := c.metrics.HitCount + c.metrics.MissCount
	if totalRequests > 0 {
		snapshot.HitRate = float64(c.metrics.HitCount) / float64(totalRequests)
	}

	// 计算错误率
	// Calculate error rate
	totalOps := c.metrics.HitCount + c.metrics.MissCount + c.metrics.SetCount + c.metrics.DeleteCount
	if totalOps > 0 {
		snapshot.ErrorRate = float64(c.metrics.ErrorCount) / float64(totalOps)
	}

	// 计算每秒操作数
	// Calculate operations per second
	duration := time.Since(c.metrics.StartTime).Seconds()
	if duration > 0 {
		snapshot.OperationsPerSec = float64(totalOps) / duration
	}

	snapshot.AvgLatency = c.metrics.AvgLatency

	// 生成库存汇总
	// Generate stock summary
	snapshot.StockSummary = c.generateStockSummary()

	// 生成活动汇总
	// Generate activity summary
	snapshot.ActivitySummary = c.generateActivitySummary()

	// 检查告警
	// Check alerts
	snapshot.Alerts = c.checkAlerts(snapshot)

	return snapshot
}

// generateStockSummary 生成库存汇总
// generateStockSummary generates stock summary
func (c *MetricsCollector) generateStockSummary() *StockSummary {
	summary := &StockSummary{}

	totalStock := 0
	for _, metric := range c.metrics.StockMetrics {
		summary.TotalActivities++
		summary.TotalSoldCount += metric.SoldCount
		totalStock += metric.CurrentStock

		if metric.CurrentStock == 0 {
			summary.OutOfStockCount++
		} else if metric.CurrentStock <= c.config.AlertThresholds.LowStockAlert {
			summary.LowStockCount++
		}
	}

	if summary.TotalActivities > 0 {
		summary.AvgStockLevel = float64(totalStock) / float64(summary.TotalActivities)
	}

	return summary
}

// generateActivitySummary 生成活动汇总
// generateActivitySummary generates activity summary
func (c *MetricsCollector) generateActivitySummary() *ActivitySummary {
	summary := &ActivitySummary{}

	var totalResponseTime time.Duration

	for _, metric := range c.metrics.ActivityMetrics {
		summary.ActiveCount++
		summary.TotalRequests += metric.RequestCount
		totalResponseTime += metric.AvgResponseTime
	}

	if summary.TotalRequests > 0 {
		totalSuccess := int64(0)
		for _, metric := range c.metrics.ActivityMetrics {
			totalSuccess += metric.SuccessCount
		}
		summary.SuccessRate = float64(totalSuccess) / float64(summary.TotalRequests)
	}

	if summary.ActiveCount > 0 {
		summary.AvgResponseTime = totalResponseTime / time.Duration(summary.ActiveCount)
	}

	return summary
}

// checkAlerts 检查告警
// checkAlerts checks alerts
func (c *MetricsCollector) checkAlerts(snapshot *MetricsSnapshot) []Alert {
	var alerts []Alert
	now := time.Now()

	// 检查命中率告警
	// Check hit rate alert
	if snapshot.HitRate < c.config.AlertThresholds.LowHitRate {
		alerts = append(alerts, Alert{
			Type:      "hit_rate",
			Level:     "warning",
			Message:   "Cache hit rate is below threshold",
			Value:     snapshot.HitRate,
			Threshold: c.config.AlertThresholds.LowHitRate,
			Timestamp: now,
		})
	}

	// 检查错误率告警
	// Check error rate alert
	if snapshot.ErrorRate > c.config.AlertThresholds.HighErrorRate {
		alerts = append(alerts, Alert{
			Type:      "error_rate",
			Level:     "error",
			Message:   "Cache error rate is above threshold",
			Value:     snapshot.ErrorRate,
			Threshold: c.config.AlertThresholds.HighErrorRate,
			Timestamp: now,
		})
	}

	// 检查延迟告警
	// Check latency alert
	if snapshot.AvgLatency > c.config.AlertThresholds.HighLatency {
		alerts = append(alerts, Alert{
			Type:      "latency",
			Level:     "warning",
			Message:   "Average latency is above threshold",
			Value:     float64(snapshot.AvgLatency.Milliseconds()),
			Threshold: float64(c.config.AlertThresholds.HighLatency.Milliseconds()),
			Timestamp: now,
		})
	}

	// 检查库存告警
	// Check stock alerts
	if snapshot.StockSummary.LowStockCount > 0 {
		alerts = append(alerts, Alert{
			Type:      "low_stock",
			Level:     "warning",
			Message:   fmt.Sprintf("%d activities have low stock", snapshot.StockSummary.LowStockCount),
			Value:     float64(snapshot.StockSummary.LowStockCount),
			Threshold: 0,
			Timestamp: now,
		})
	}

	if snapshot.StockSummary.OutOfStockCount > 0 {
		alerts = append(alerts, Alert{
			Type:      "out_of_stock",
			Level:     "critical",
			Message:   fmt.Sprintf("%d activities are out of stock", snapshot.StockSummary.OutOfStockCount),
			Value:     float64(snapshot.StockSummary.OutOfStockCount),
			Threshold: 0,
			Timestamp: now,
		})
	}

	return alerts
}

// processAlerts 处理告警
// processAlerts processes alerts
func (c *MetricsCollector) processAlerts(alerts []Alert) {
	for _, alert := range alerts {
		switch alert.Level {
		case "critical":
			c.logger.Error("Cache critical alert",
				logger.String("type", alert.Type),
				logger.String("message", alert.Message),
				logger.Float64("value", alert.Value),
				logger.Float64("threshold", alert.Threshold))
		case "error":
			c.logger.Error("Cache error alert",
				logger.String("type", alert.Type),
				logger.String("message", alert.Message),
				logger.Float64("value", alert.Value),
				logger.Float64("threshold", alert.Threshold))
		case "warning":
			c.logger.Warn("Cache warning alert",
				logger.String("type", alert.Type),
				logger.String("message", alert.Message),
				logger.Float64("value", alert.Value),
				logger.Float64("threshold", alert.Threshold))
		}
	}
}

// logDetailedMetrics 记录详细指标
// logDetailedMetrics logs detailed metrics
func (c *MetricsCollector) logDetailedMetrics(snapshot *MetricsSnapshot) {
	c.logger.Info("Cache metrics snapshot",
		logger.Float64("hit_rate", snapshot.HitRate),
		logger.Float64("error_rate", snapshot.ErrorRate),
		logger.Float64("ops_per_sec", snapshot.OperationsPerSec),
		logger.Duration("avg_latency", snapshot.AvgLatency),
		logger.Int("total_activities", snapshot.StockSummary.TotalActivities),
		logger.Int("low_stock_count", snapshot.StockSummary.LowStockCount),
		logger.Int("alerts_count", len(snapshot.Alerts)))
}

// GetCurrentMetrics 获取当前指标
// GetCurrentMetrics gets current metrics
func (c *MetricsCollector) GetCurrentMetrics() *CacheMetricsData {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	// 创建副本
	// Create copy
	metrics := &CacheMetricsData{
		HitCount:        atomic.LoadInt64(&c.metrics.HitCount),
		MissCount:       atomic.LoadInt64(&c.metrics.MissCount),
		SetCount:        atomic.LoadInt64(&c.metrics.SetCount),
		DeleteCount:     atomic.LoadInt64(&c.metrics.DeleteCount),
		ErrorCount:      atomic.LoadInt64(&c.metrics.ErrorCount),
		AvgLatency:      c.metrics.AvgLatency,
		MaxLatency:      c.metrics.MaxLatency,
		MinLatency:      c.metrics.MinLatency,
		TotalLatency:    c.metrics.TotalLatency,
		OperationCount:  c.metrics.OperationCount,
		LastUpdated:     c.metrics.LastUpdated,
		StartTime:       c.metrics.StartTime,
		StockMetrics:    make(map[string]*StockMetric),
		ActivityMetrics: make(map[string]*ActivityMetric),
	}

	// 复制映射
	// Copy maps
	for k, v := range c.metrics.StockMetrics {
		stockCopy := *v
		metrics.StockMetrics[k] = &stockCopy
	}
	for k, v := range c.metrics.ActivityMetrics {
		activityCopy := *v
		metrics.ActivityMetrics[k] = &activityCopy
	}

	return metrics
}

// ExportMetrics 导出指标
// ExportMetrics exports metrics
func (c *MetricsCollector) ExportMetrics() ([]byte, error) {
	snapshot := c.GenerateSnapshot()
	return json.Marshal(snapshot)
}

// IsRunning 检查是否运行中
// IsRunning checks if running
func (c *MetricsCollector) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// DefaultMetricsConfig 默认指标配置
// DefaultMetricsConfig default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		CollectInterval:   30 * time.Second,
		RetentionPeriod:   24 * time.Hour,
		EnableDetailedLog: false,
		AlertThresholds: AlertThresholds{
			LowHitRate:    0.8,
			HighErrorRate: 0.05,
			LowStockAlert: 10,
			HighLatency:   100 * time.Millisecond,
		},
	}
}
