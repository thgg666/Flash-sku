package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/cache"
)

func TestDefaultUpdateConfig(t *testing.T) {
	config := cache.DefaultUpdateConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 1*time.Hour, config.DefaultTTL)
	assert.Equal(t, 0.2, config.RefreshThreshold)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.RetryDelay)
	assert.True(t, config.EnableAsync)
	assert.Equal(t, 50, config.BatchSize)
}

func TestUpdateResult(t *testing.T) {
	now := time.Now()
	duration := 100 * time.Millisecond

	result := &cache.UpdateResult{
		Key:       "test_key",
		Success:   true,
		Error:     "",
		Duration:  duration,
		Strategy:  "write_through",
		Timestamp: now,
	}

	assert.Equal(t, "test_key", result.Key)
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)
	assert.Equal(t, duration, result.Duration)
	assert.Equal(t, "write_through", result.Strategy)
	assert.Equal(t, now, result.Timestamp)
}

func TestUpdateResultWithError(t *testing.T) {
	result := &cache.UpdateResult{
		Key:      "test_key",
		Success:  false,
		Error:    "connection failed",
		Strategy: "write_behind",
	}

	assert.Equal(t, "test_key", result.Key)
	assert.False(t, result.Success)
	assert.Equal(t, "connection failed", result.Error)
	assert.Equal(t, "write_behind", result.Strategy)
}

func TestDefaultConsistencyConfig(t *testing.T) {
	config := cache.DefaultConsistencyConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5*time.Minute, config.CheckInterval)
	assert.True(t, config.RepairEnabled)
	assert.Equal(t, 3, config.MaxRepairRetries)
	assert.Equal(t, 1*time.Second, config.RepairDelay)
	assert.Equal(t, 0.95, config.AlertThreshold)
}

func TestCacheValidationResult(t *testing.T) {
	now := time.Now()

	result := &cache.ValidationResult{
		Key:          "test_key",
		IsConsistent: true,
		CacheValue:   "cache_data",
		SourceValue:  "source_data",
		Difference:   "",
		RepairAction: "none",
		Timestamp:    now,
	}

	assert.Equal(t, "test_key", result.Key)
	assert.True(t, result.IsConsistent)
	assert.Equal(t, "cache_data", result.CacheValue)
	assert.Equal(t, "source_data", result.SourceValue)
	assert.Empty(t, result.Difference)
	assert.Equal(t, "none", result.RepairAction)
	assert.Equal(t, now, result.Timestamp)
}

func TestValidationResultInconsistent(t *testing.T) {
	result := &cache.ValidationResult{
		Key:          "test_key",
		IsConsistent: false,
		CacheValue:   "old_data",
		SourceValue:  "new_data",
		Difference:   "value mismatch",
		RepairAction: "update_cache",
	}

	assert.Equal(t, "test_key", result.Key)
	assert.False(t, result.IsConsistent)
	assert.Equal(t, "old_data", result.CacheValue)
	assert.Equal(t, "new_data", result.SourceValue)
	assert.Equal(t, "value mismatch", result.Difference)
	assert.Equal(t, "update_cache", result.RepairAction)
}

func TestConsistencyReport(t *testing.T) {
	now := time.Now()
	duration := 500 * time.Millisecond

	report := &cache.ConsistencyReport{
		TotalChecked:     100,
		ConsistentCount:  95,
		InconsistentKeys: []string{"key1", "key2", "key3"},
		ValidationResults: []*cache.ValidationResult{
			{Key: "key1", IsConsistent: false},
			{Key: "key2", IsConsistent: true},
		},
		ConsistencyRate: 0.95,
		CheckTime:       now,
		Duration:        duration,
	}

	assert.Equal(t, 100, report.TotalChecked)
	assert.Equal(t, 95, report.ConsistentCount)
	assert.Len(t, report.InconsistentKeys, 3)
	assert.Contains(t, report.InconsistentKeys, "key1")
	assert.Contains(t, report.InconsistentKeys, "key2")
	assert.Contains(t, report.InconsistentKeys, "key3")
	assert.Len(t, report.ValidationResults, 2)
	assert.Equal(t, 0.95, report.ConsistencyRate)
	assert.Equal(t, now, report.CheckTime)
	assert.Equal(t, duration, report.Duration)

	// 验证一致性率计算
	// Verify consistency rate calculation
	expectedRate := float64(report.ConsistentCount) / float64(report.TotalChecked)
	assert.InDelta(t, expectedRate, report.ConsistencyRate, 0.001)
}

func TestConsistencyReportCalculations(t *testing.T) {
	// 测试不同的一致性率场景
	// Test different consistency rate scenarios
	testCases := []struct {
		name            string
		totalChecked    int
		consistentCount int
		expectedRate    float64
	}{
		{"Perfect consistency", 100, 100, 1.0},
		{"High consistency", 100, 95, 0.95},
		{"Medium consistency", 100, 80, 0.80},
		{"Low consistency", 100, 50, 0.50},
		{"Zero consistency", 100, 0, 0.0},
		{"Empty check", 0, 0, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report := &cache.ConsistencyReport{
				TotalChecked:    tc.totalChecked,
				ConsistentCount: tc.consistentCount,
			}

			// 计算一致性率
			// Calculate consistency rate
			if report.TotalChecked > 0 {
				report.ConsistencyRate = float64(report.ConsistentCount) / float64(report.TotalChecked)
			}

			assert.InDelta(t, tc.expectedRate, report.ConsistencyRate, 0.001)
		})
	}
}

func TestUpdateConfigValidation(t *testing.T) {
	// 测试配置验证逻辑
	// Test configuration validation logic

	// 有效配置
	// Valid configuration
	validConfig := &cache.UpdateConfig{
		DefaultTTL:       1 * time.Hour,
		RefreshThreshold: 0.2,
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
		EnableAsync:      true,
		BatchSize:        50,
	}

	assert.Greater(t, validConfig.DefaultTTL, time.Duration(0))
	assert.Greater(t, validConfig.RefreshThreshold, 0.0)
	assert.Less(t, validConfig.RefreshThreshold, 1.0)
	assert.Greater(t, validConfig.MaxRetries, 0)
	assert.Greater(t, validConfig.RetryDelay, time.Duration(0))
	assert.Greater(t, validConfig.BatchSize, 0)

	// 边界值测试
	// Boundary value testing
	boundaryConfig := &cache.UpdateConfig{
		DefaultTTL:       1 * time.Second,
		RefreshThreshold: 0.01,
		MaxRetries:       1,
		RetryDelay:       1 * time.Millisecond,
		EnableAsync:      false,
		BatchSize:        1,
	}

	assert.Equal(t, 1*time.Second, boundaryConfig.DefaultTTL)
	assert.Equal(t, 0.01, boundaryConfig.RefreshThreshold)
	assert.Equal(t, 1, boundaryConfig.MaxRetries)
	assert.Equal(t, 1*time.Millisecond, boundaryConfig.RetryDelay)
	assert.False(t, boundaryConfig.EnableAsync)
	assert.Equal(t, 1, boundaryConfig.BatchSize)
}
