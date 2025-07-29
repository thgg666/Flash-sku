package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/cache"
)

func TestDefaultMetricsConfig(t *testing.T) {
	config := cache.DefaultMetricsConfig()
	
	assert.NotNil(t, config)
	assert.Equal(t, 30*time.Second, config.CollectInterval)
	assert.Equal(t, 24*time.Hour, config.RetentionPeriod)
	assert.False(t, config.EnableDetailedLog)
	
	// 检查告警阈值
	// Check alert thresholds
	assert.Equal(t, 0.8, config.AlertThresholds.LowHitRate)
	assert.Equal(t, 0.05, config.AlertThresholds.HighErrorRate)
	assert.Equal(t, 10, config.AlertThresholds.LowStockAlert)
	assert.Equal(t, 100*time.Millisecond, config.AlertThresholds.HighLatency)
}

func TestStockMetric(t *testing.T) {
	now := time.Now()
	
	metric := &cache.StockMetric{
		ActivityID:   "activity_123",
		CurrentStock: 50,
		InitialStock: 100,
		SoldCount:    50,
		LastUpdated:  now,
		UpdateCount:  10,
	}
	
	assert.Equal(t, "activity_123", metric.ActivityID)
	assert.Equal(t, 50, metric.CurrentStock)
	assert.Equal(t, 100, metric.InitialStock)
	assert.Equal(t, 50, metric.SoldCount)
	assert.Equal(t, now, metric.LastUpdated)
	assert.Equal(t, int64(10), metric.UpdateCount)
	
	// 验证库存一致性
	// Verify stock consistency
	assert.Equal(t, metric.InitialStock-metric.SoldCount, metric.CurrentStock)
}

func TestActivityMetric(t *testing.T) {
	now := time.Now()
	
	metric := &cache.ActivityMetric{
		ActivityID:      "activity_456",
		RequestCount:    1000,
		SuccessCount:    950,
		FailureCount:    50,
		LastAccess:      now,
		AvgResponseTime: 50 * time.Millisecond,
	}
	
	assert.Equal(t, "activity_456", metric.ActivityID)
	assert.Equal(t, int64(1000), metric.RequestCount)
	assert.Equal(t, int64(950), metric.SuccessCount)
	assert.Equal(t, int64(50), metric.FailureCount)
	assert.Equal(t, now, metric.LastAccess)
	assert.Equal(t, 50*time.Millisecond, metric.AvgResponseTime)
	
	// 验证请求计数一致性
	// Verify request count consistency
	assert.Equal(t, metric.RequestCount, metric.SuccessCount+metric.FailureCount)
	
	// 计算成功率
	// Calculate success rate
	successRate := float64(metric.SuccessCount) / float64(metric.RequestCount)
	assert.InDelta(t, 0.95, successRate, 0.01) // 95% 成功率
}

func TestMetricsSnapshot(t *testing.T) {
	now := time.Now()
	
	snapshot := &cache.MetricsSnapshot{
		Timestamp:        now,
		HitRate:          0.85,
		ErrorRate:        0.02,
		OperationsPerSec: 1000.5,
		AvgLatency:       25 * time.Millisecond,
		StockSummary: &cache.StockSummary{
			TotalActivities: 10,
			LowStockCount:   2,
			OutOfStockCount: 1,
			TotalSoldCount:  500,
			AvgStockLevel:   75.5,
		},
		ActivitySummary: &cache.ActivitySummary{
			ActiveCount:     5,
			TotalRequests:   5000,
			SuccessRate:     0.96,
			AvgResponseTime: 30 * time.Millisecond,
		},
		Alerts: []cache.Alert{
			{
				Type:      "hit_rate",
				Level:     "warning",
				Message:   "Hit rate below threshold",
				Value:     0.75,
				Threshold: 0.8,
				Timestamp: now,
			},
		},
	}
	
	assert.Equal(t, now, snapshot.Timestamp)
	assert.Equal(t, 0.85, snapshot.HitRate)
	assert.Equal(t, 0.02, snapshot.ErrorRate)
	assert.Equal(t, 1000.5, snapshot.OperationsPerSec)
	assert.Equal(t, 25*time.Millisecond, snapshot.AvgLatency)
	assert.NotNil(t, snapshot.StockSummary)
	assert.NotNil(t, snapshot.ActivitySummary)
	assert.Len(t, snapshot.Alerts, 1)
}

func TestStockSummary(t *testing.T) {
	summary := &cache.StockSummary{
		TotalActivities: 20,
		LowStockCount:   3,
		OutOfStockCount: 2,
		TotalSoldCount:  1500,
		AvgStockLevel:   62.5,
	}
	
	assert.Equal(t, 20, summary.TotalActivities)
	assert.Equal(t, 3, summary.LowStockCount)
	assert.Equal(t, 2, summary.OutOfStockCount)
	assert.Equal(t, 1500, summary.TotalSoldCount)
	assert.Equal(t, 62.5, summary.AvgStockLevel)
	
	// 验证库存状态分布
	// Verify stock status distribution
	normalStockCount := summary.TotalActivities - summary.LowStockCount - summary.OutOfStockCount
	assert.Equal(t, 15, normalStockCount) // 20 - 3 - 2 = 15
}

func TestActivitySummary(t *testing.T) {
	summary := &cache.ActivitySummary{
		ActiveCount:     8,
		TotalRequests:   10000,
		SuccessRate:     0.94,
		AvgResponseTime: 45 * time.Millisecond,
	}
	
	assert.Equal(t, 8, summary.ActiveCount)
	assert.Equal(t, int64(10000), summary.TotalRequests)
	assert.Equal(t, 0.94, summary.SuccessRate)
	assert.Equal(t, 45*time.Millisecond, summary.AvgResponseTime)
	
	// 计算成功和失败请求数
	// Calculate success and failure request counts
	successRequests := int64(float64(summary.TotalRequests) * summary.SuccessRate)
	failureRequests := summary.TotalRequests - successRequests
	
	assert.Equal(t, int64(9400), successRequests)
	assert.Equal(t, int64(600), failureRequests)
}

func TestAlert(t *testing.T) {
	now := time.Now()
	
	alert := cache.Alert{
		Type:      "error_rate",
		Level:     "error",
		Message:   "Error rate too high",
		Value:     0.08,
		Threshold: 0.05,
		Timestamp: now,
	}
	
	assert.Equal(t, "error_rate", alert.Type)
	assert.Equal(t, "error", alert.Level)
	assert.Equal(t, "Error rate too high", alert.Message)
	assert.Equal(t, 0.08, alert.Value)
	assert.Equal(t, 0.05, alert.Threshold)
	assert.Equal(t, now, alert.Timestamp)
	
	// 验证告警条件
	// Verify alert condition
	assert.Greater(t, alert.Value, alert.Threshold)
}

func TestAlertLevels(t *testing.T) {
	// 测试不同级别的告警
	// Test different alert levels
	levels := []string{"warning", "error", "critical"}
	
	for _, level := range levels {
		alert := cache.Alert{
			Type:      "test",
			Level:     level,
			Message:   "Test alert",
			Value:     1.0,
			Threshold: 0.5,
			Timestamp: time.Now(),
		}
		
		assert.Equal(t, level, alert.Level)
		assert.Contains(t, []string{"warning", "error", "critical"}, alert.Level)
	}
}

func TestCacheMetricsData(t *testing.T) {
	now := time.Now()
	
	data := &cache.CacheMetricsData{
		HitCount:       1000,
		MissCount:      200,
		SetCount:       500,
		DeleteCount:    50,
		ErrorCount:     10,
		AvgLatency:     30 * time.Millisecond,
		MaxLatency:     100 * time.Millisecond,
		MinLatency:     5 * time.Millisecond,
		TotalLatency:   45 * time.Second,
		OperationCount: 1500,
		LastUpdated:    now,
		StartTime:      now.Add(-1 * time.Hour),
		StockMetrics:   make(map[string]*cache.StockMetric),
		ActivityMetrics: make(map[string]*cache.ActivityMetric),
	}
	
	assert.Equal(t, int64(1000), data.HitCount)
	assert.Equal(t, int64(200), data.MissCount)
	assert.Equal(t, int64(500), data.SetCount)
	assert.Equal(t, int64(50), data.DeleteCount)
	assert.Equal(t, int64(10), data.ErrorCount)
	assert.Equal(t, 30*time.Millisecond, data.AvgLatency)
	assert.Equal(t, 100*time.Millisecond, data.MaxLatency)
	assert.Equal(t, 5*time.Millisecond, data.MinLatency)
	assert.Equal(t, 45*time.Second, data.TotalLatency)
	assert.Equal(t, int64(1500), data.OperationCount)
	assert.Equal(t, now, data.LastUpdated)
	assert.NotNil(t, data.StockMetrics)
	assert.NotNil(t, data.ActivityMetrics)
	
	// 计算命中率
	// Calculate hit rate
	totalRequests := data.HitCount + data.MissCount
	hitRate := float64(data.HitCount) / float64(totalRequests)
	assert.InDelta(t, 0.833, hitRate, 0.001) // 1000/(1000+200) ≈ 0.833
	
	// 验证平均延迟计算
	// Verify average latency calculation
	expectedAvgLatency := data.TotalLatency / time.Duration(data.OperationCount)
	assert.Equal(t, expectedAvgLatency, data.AvgLatency)
}

func TestAlertThresholds(t *testing.T) {
	thresholds := cache.AlertThresholds{
		LowHitRate:    0.75,
		HighErrorRate: 0.1,
		LowStockAlert: 5,
		HighLatency:   200 * time.Millisecond,
	}
	
	assert.Equal(t, 0.75, thresholds.LowHitRate)
	assert.Equal(t, 0.1, thresholds.HighErrorRate)
	assert.Equal(t, 5, thresholds.LowStockAlert)
	assert.Equal(t, 200*time.Millisecond, thresholds.HighLatency)
	
	// 验证阈值合理性
	// Verify threshold reasonableness
	assert.Greater(t, thresholds.LowHitRate, 0.0)
	assert.Less(t, thresholds.LowHitRate, 1.0)
	assert.Greater(t, thresholds.HighErrorRate, 0.0)
	assert.Less(t, thresholds.HighErrorRate, 1.0)
	assert.Greater(t, thresholds.LowStockAlert, 0)
	assert.Greater(t, thresholds.HighLatency, time.Duration(0))
}
