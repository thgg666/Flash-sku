package ratelimit

import (
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
)

// RateLimitMetrics 限流指标
// RateLimitMetrics rate limit metrics
type RateLimitMetrics struct {
	// 总体指标
	// Overall metrics
	TotalRequests    int64 `json:"total_requests"`
	AllowedRequests  int64 `json:"allowed_requests"`
	BlockedRequests  int64 `json:"blocked_requests"`
	
	// 按级别统计
	// Statistics by level
	GlobalBlocked    int64 `json:"global_blocked"`
	IPBlocked        int64 `json:"ip_blocked"`
	UserBlocked      int64 `json:"user_blocked"`
	
	// 性能指标
	// Performance metrics
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	MaxResponseTime  float64 `json:"max_response_time_ms"`
	
	// 时间窗口
	// Time window
	WindowStart      time.Time `json:"window_start"`
	WindowEnd        time.Time `json:"window_end"`
	
	// 计算字段
	// Calculated fields
	AllowRate        float64 `json:"allow_rate"`
	BlockRate        float64 `json:"block_rate"`
}

// MetricsCollector 指标收集器
// MetricsCollector metrics collector
type MetricsCollector struct {
	metrics     *RateLimitMetrics
	mutex       sync.RWMutex
	logger      logger.Logger
	windowSize  time.Duration
	
	// 响应时间统计
	// Response time statistics
	responseTimes []float64
	maxSamples    int
}

// NewMetricsCollector 创建指标收集器
// NewMetricsCollector creates metrics collector
func NewMetricsCollector(windowSize time.Duration, log logger.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics: &RateLimitMetrics{
			WindowStart: time.Now(),
		},
		logger:        log,
		windowSize:    windowSize,
		responseTimes: make([]float64, 0, 1000),
		maxSamples:    1000,
	}
}

// RecordRequest 记录请求
// RecordRequest records a request
func (mc *MetricsCollector) RecordRequest(result *LimitResult, responseTime time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.metrics.TotalRequests++
	
	responseTimeMs := float64(responseTime.Nanoseconds()) / 1e6
	
	if result.Allowed {
		mc.metrics.AllowedRequests++
	} else {
		mc.metrics.BlockedRequests++
		
		// 按级别统计
		// Count by level
		switch result.Level {
		case LimitLevelGlobal:
			mc.metrics.GlobalBlocked++
		case LimitLevelIP:
			mc.metrics.IPBlocked++
		case LimitLevelUser:
			mc.metrics.UserBlocked++
		}
	}
	
	// 记录响应时间
	// Record response time
	mc.recordResponseTime(responseTimeMs)
	
	// 更新计算字段
	// Update calculated fields
	mc.updateCalculatedFields()
}

// recordResponseTime 记录响应时间
// recordResponseTime records response time
func (mc *MetricsCollector) recordResponseTime(responseTimeMs float64) {
	// 添加新的响应时间
	// Add new response time
	if len(mc.responseTimes) >= mc.maxSamples {
		// 移除最旧的样本
		// Remove oldest sample
		mc.responseTimes = mc.responseTimes[1:]
	}
	mc.responseTimes = append(mc.responseTimes, responseTimeMs)
	
	// 更新最大响应时间
	// Update max response time
	if responseTimeMs > mc.metrics.MaxResponseTime {
		mc.metrics.MaxResponseTime = responseTimeMs
	}
}

// updateCalculatedFields 更新计算字段
// updateCalculatedFields updates calculated fields
func (mc *MetricsCollector) updateCalculatedFields() {
	if mc.metrics.TotalRequests > 0 {
		mc.metrics.AllowRate = float64(mc.metrics.AllowedRequests) / float64(mc.metrics.TotalRequests)
		mc.metrics.BlockRate = float64(mc.metrics.BlockedRequests) / float64(mc.metrics.TotalRequests)
	}
	
	// 计算平均响应时间
	// Calculate average response time
	if len(mc.responseTimes) > 0 {
		var sum float64
		for _, rt := range mc.responseTimes {
			sum += rt
		}
		mc.metrics.AvgResponseTime = sum / float64(len(mc.responseTimes))
	}
	
	mc.metrics.WindowEnd = time.Now()
}

// GetMetrics 获取当前指标
// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() *RateLimitMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	// 返回指标的副本
	// Return a copy of metrics
	metricsCopy := *mc.metrics
	return &metricsCopy
}

// Reset 重置指标
// Reset resets metrics
func (mc *MetricsCollector) Reset() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.metrics = &RateLimitMetrics{
		WindowStart: time.Now(),
	}
	mc.responseTimes = mc.responseTimes[:0]
	
	mc.logger.Info("Rate limit metrics reset")
}

// GetSummary 获取指标摘要
// GetSummary returns metrics summary
func (mc *MetricsCollector) GetSummary() map[string]interface{} {
	metrics := mc.GetMetrics()
	
	duration := metrics.WindowEnd.Sub(metrics.WindowStart)
	
	summary := map[string]interface{}{
		"window_duration_seconds": duration.Seconds(),
		"total_requests":          metrics.TotalRequests,
		"requests_per_second":     float64(metrics.TotalRequests) / duration.Seconds(),
		"allow_rate_percent":      metrics.AllowRate * 100,
		"block_rate_percent":      metrics.BlockRate * 100,
		"avg_response_time_ms":    metrics.AvgResponseTime,
		"max_response_time_ms":    metrics.MaxResponseTime,
		"blocked_by_level": map[string]int64{
			"global": metrics.GlobalBlocked,
			"ip":     metrics.IPBlocked,
			"user":   metrics.UserBlocked,
		},
	}
	
	return summary
}

// InstrumentedRateLimiter 带指标的限流器
// InstrumentedRateLimiter rate limiter with metrics
type InstrumentedRateLimiter struct {
	limiter   RateLimiterInterface
	collector *MetricsCollector
	logger    logger.Logger
}

// NewInstrumentedRateLimiter 创建带指标的限流器
// NewInstrumentedRateLimiter creates rate limiter with metrics
func NewInstrumentedRateLimiter(limiter RateLimiterInterface, collector *MetricsCollector, log logger.Logger) *InstrumentedRateLimiter {
	return &InstrumentedRateLimiter{
		limiter:   limiter,
		collector: collector,
		logger:    log,
	}
}

// Allow 检查是否允许请求（带指标记录）
// Allow checks if request is allowed (with metrics recording)
func (irl *InstrumentedRateLimiter) Allow(ip, userID string) *LimitResult {
	startTime := time.Now()
	
	result := irl.limiter.Allow(ip, userID)
	
	responseTime := time.Since(startTime)
	irl.collector.RecordRequest(result, responseTime)
	
	// 记录详细日志
	// Log detailed information
	if !result.Allowed {
		irl.logger.Warn("Request blocked by rate limiter",
			logger.String("ip", ip),
			logger.String("user_id", userID),
			logger.String("level", string(result.Level)),
			logger.String("reason", result.Reason),
			logger.Int64("retry_after", result.RetryAfter),
			logger.Duration("response_time", responseTime))
	} else {
		irl.logger.Debug("Request allowed by rate limiter",
			logger.String("ip", ip),
			logger.String("user_id", userID),
			logger.Int64("remaining_tokens", result.RemainingTokens),
			logger.Duration("response_time", responseTime))
	}
	
	return result
}

// GetMetrics 获取指标
// GetMetrics returns metrics
func (irl *InstrumentedRateLimiter) GetMetrics() *RateLimitMetrics {
	return irl.collector.GetMetrics()
}

// GetSummary 获取指标摘要
// GetSummary returns metrics summary
func (irl *InstrumentedRateLimiter) GetSummary() map[string]interface{} {
	return irl.collector.GetSummary()
}

// ResetMetrics 重置指标
// ResetMetrics resets metrics
func (irl *InstrumentedRateLimiter) ResetMetrics() {
	irl.collector.Reset()
}

// PerformanceAlert 性能告警
// PerformanceAlert performance alert
type PerformanceAlert struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	Timestamp   time.Time `json:"timestamp"`
}

// CheckAlerts 检查性能告警
// CheckAlerts checks for performance alerts
func (irl *InstrumentedRateLimiter) CheckAlerts() []PerformanceAlert {
	metrics := irl.GetMetrics()
	var alerts []PerformanceAlert
	
	// 检查阻塞率告警
	// Check block rate alert
	if metrics.BlockRate > 0.1 { // 阻塞率超过10%
		alerts = append(alerts, PerformanceAlert{
			Type:         "HIGH_BLOCK_RATE",
			Message:      "Rate limit block rate is too high",
			Threshold:    0.1,
			CurrentValue: metrics.BlockRate,
			Timestamp:    time.Now(),
		})
	}
	
	// 检查响应时间告警
	// Check response time alert
	if metrics.AvgResponseTime > 10.0 { // 平均响应时间超过10ms
		alerts = append(alerts, PerformanceAlert{
			Type:         "HIGH_RESPONSE_TIME",
			Message:      "Rate limiter response time is too high",
			Threshold:    10.0,
			CurrentValue: metrics.AvgResponseTime,
			Timestamp:    time.Now(),
		})
	}
	
	return alerts
}
