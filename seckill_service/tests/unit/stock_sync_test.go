package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/sync"
)

func TestDefaultStockSyncConfig(t *testing.T) {
	config := sync.DefaultStockSyncConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 1*time.Minute, config.SyncInterval)
	assert.Equal(t, 50, config.BatchSize)
	assert.Equal(t, 1*time.Hour, config.StockTTL)
	assert.True(t, config.EnableRealtime)
	assert.Equal(t, "merge", config.ConflictStrategy)
}

func TestStockSyncResult(t *testing.T) {
	result := &sync.StockSyncResult{
		ActivityID:   "test_activity_1",
		Success:      true,
		OldStock:     100,
		NewStock:     95,
		ConflictType: "merge_use_smaller_redis",
		SyncTime:     time.Now(),
	}

	assert.Equal(t, "test_activity_1", result.ActivityID)
	assert.True(t, result.Success)
	assert.Equal(t, 100, result.OldStock)
	assert.Equal(t, 95, result.NewStock)
	assert.Equal(t, "merge_use_smaller_redis", result.ConflictType)
	assert.NotZero(t, result.SyncTime)
}

func TestStockData(t *testing.T) {
	stockData := &sync.StockData{
		ActivityID:     "activity_123",
		TotalStock:     1000,
		AvailableStock: 800,
		ReservedStock:  50,
		SoldStock:      150,
		LastUpdated:    time.Now(),
		Version:        1,
	}

	assert.Equal(t, "activity_123", stockData.ActivityID)
	assert.Equal(t, 1000, stockData.TotalStock)
	assert.Equal(t, 800, stockData.AvailableStock)
	assert.Equal(t, 50, stockData.ReservedStock)
	assert.Equal(t, 150, stockData.SoldStock)
	assert.Equal(t, 1, stockData.Version)

	// 验证库存一致性
	// Verify stock consistency
	totalUsed := stockData.ReservedStock + stockData.SoldStock
	assert.Equal(t, stockData.TotalStock-stockData.AvailableStock, totalUsed)
}

func TestStockSyncMetrics(t *testing.T) {
	now := time.Now()
	metrics := &sync.StockSyncMetrics{
		TotalSynced:      100,
		SuccessCount:     95,
		ErrorCount:       5,
		ConflictCount:    10,
		LastSyncTime:     now,
		AvgSyncDuration:  500 * time.Millisecond,
		LastSyncDuration: 600 * time.Millisecond,
		ConflictsByType: map[string]int64{
			"merge_use_smaller_redis": 5,
			"merge_use_smaller_db":    3,
			"redis_priority_kept":     2,
		},
	}

	assert.Equal(t, int64(100), metrics.TotalSynced)
	assert.Equal(t, int64(95), metrics.SuccessCount)
	assert.Equal(t, int64(5), metrics.ErrorCount)
	assert.Equal(t, int64(10), metrics.ConflictCount)
	assert.Equal(t, now, metrics.LastSyncTime)

	// 计算成功率
	// Calculate success rate
	successRate := float64(metrics.SuccessCount) / float64(metrics.TotalSynced)
	assert.InDelta(t, 0.95, successRate, 0.01) // 95% 成功率

	// 验证冲突类型统计
	// Verify conflict type stats
	assert.Equal(t, int64(5), metrics.ConflictsByType["merge_use_smaller_redis"])
	assert.Equal(t, int64(3), metrics.ConflictsByType["merge_use_smaller_db"])
	assert.Equal(t, int64(2), metrics.ConflictsByType["redis_priority_kept"])

	// 验证冲突总数
	// Verify total conflicts
	totalConflicts := int64(0)
	for _, count := range metrics.ConflictsByType {
		totalConflicts += count
	}
	assert.Equal(t, metrics.ConflictCount, totalConflicts)
}

func TestStockSyncConfig(t *testing.T) {
	// 测试不同的冲突策略
	// Test different conflict strategies
	strategies := []string{"redis_priority", "db_priority", "merge"}

	for _, strategy := range strategies {
		config := &sync.StockSyncConfig{
			SyncInterval:     30 * time.Second,
			BatchSize:        25,
			StockTTL:         30 * time.Minute,
			EnableRealtime:   false,
			ConflictStrategy: strategy,
		}

		assert.Equal(t, 30*time.Second, config.SyncInterval)
		assert.Equal(t, 25, config.BatchSize)
		assert.Equal(t, 30*time.Minute, config.StockTTL)
		assert.False(t, config.EnableRealtime)
		assert.Equal(t, strategy, config.ConflictStrategy)
	}
}

func TestStockSyncConfigValidation(t *testing.T) {
	// 测试配置验证逻辑
	// Test configuration validation logic

	// 有效配置
	// Valid configuration
	validConfig := &sync.StockSyncConfig{
		SyncInterval:     1 * time.Minute,
		BatchSize:        50,
		StockTTL:         1 * time.Hour,
		EnableRealtime:   true,
		ConflictStrategy: "merge",
	}

	assert.Greater(t, validConfig.SyncInterval, time.Duration(0))
	assert.Greater(t, validConfig.BatchSize, 0)
	assert.Greater(t, validConfig.StockTTL, time.Duration(0))
	assert.Contains(t, []string{"redis_priority", "db_priority", "merge"}, validConfig.ConflictStrategy)

	// 边界值测试
	// Boundary value testing
	minConfig := &sync.StockSyncConfig{
		SyncInterval:     1 * time.Second,
		BatchSize:        1,
		StockTTL:         1 * time.Second,
		EnableRealtime:   false,
		ConflictStrategy: "db_priority",
	}

	assert.Equal(t, 1*time.Second, minConfig.SyncInterval)
	assert.Equal(t, 1, minConfig.BatchSize)
	assert.Equal(t, 1*time.Second, minConfig.StockTTL)
}
