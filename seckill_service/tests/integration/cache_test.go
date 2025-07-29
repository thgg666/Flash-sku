package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flashsku/seckill/internal/cache"
	"github.com/flashsku/seckill/internal/sync"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

func TestCacheManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建Redis客户端
	// Create Redis client
	redisConfig := &redis.Config{
		Host:     "localhost",
		Port:     "6379",
		Database: 1, // 使用测试数据库
	}

	redisClient, err := redis.NewClient(redisConfig)
	require.NoError(t, err)
	defer redisClient.Close()

	// 创建日志器
	// Create logger
	logConfig := &logger.Config{
		Level:  logger.DEBUG,
		Format: "text",
		Output: "stdout",
	}
	log := logger.NewLogger(logConfig)

	// 创建缓存管理器
	// Create cache manager
	cacheConfig := cache.DefaultCacheConfig()
	cacheManager := cache.NewCacheManager(redisClient, cacheConfig, log)

	ctx := context.Background()

	// 测试活动缓存
	// Test activity cache
	t.Run("ActivityCache", func(t *testing.T) {
		activityID := "test_activity_1"
		activityData := map[string]interface{}{
			"id":    activityID,
			"name":  "Test Activity",
			"stock": 100,
		}

		// 设置活动缓存
		// Set activity cache
		err := cacheManager.SetActivity(ctx, activityID, activityData)
		assert.NoError(t, err)

		// 获取活动缓存
		// Get activity cache
		var retrievedData map[string]interface{}
		err = cacheManager.GetActivity(ctx, activityID, &retrievedData)
		assert.NoError(t, err)
		assert.Equal(t, activityData["name"], retrievedData["name"])
	})

	// 测试库存缓存
	// Test stock cache
	t.Run("StockCache", func(t *testing.T) {
		activityID := "test_activity_2"
		initialStock := 100

		// 设置库存
		// Set stock
		err := cacheManager.SetStock(ctx, activityID, initialStock)
		assert.NoError(t, err)

		// 获取库存
		// Get stock
		stock, err := cacheManager.GetStock(ctx, activityID)
		assert.NoError(t, err)
		assert.Equal(t, initialStock, stock)

		// 递减库存
		// Decrement stock
		newStock, err := cacheManager.DecrStock(ctx, activityID)
		assert.NoError(t, err)
		assert.Equal(t, int64(initialStock-1), newStock)
	})

	// 测试用户限购
	// Test user limit
	t.Run("UserLimit", func(t *testing.T) {
		userID := "test_user_1"
		activityID := "test_activity_3"

		// 增加用户购买计数
		// Increment user purchase count
		count, err := cacheManager.IncrUserLimit(ctx, userID, activityID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// 再次增加
		// Increment again
		count, err = cacheManager.IncrUserLimit(ctx, userID, activityID)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// 获取用户限购
		// Get user limit
		userCount, err := cacheManager.GetUserLimit(ctx, userID, activityID)
		assert.NoError(t, err)
		assert.Equal(t, 2, userCount)
	})

	// 测试缓存失效
	// Test cache invalidation
	t.Run("CacheInvalidation", func(t *testing.T) {
		activityID := "test_activity_4"

		// 设置活动和库存
		// Set activity and stock
		err := cacheManager.SetActivity(ctx, activityID, map[string]string{"name": "Test"})
		assert.NoError(t, err)
		err = cacheManager.SetStock(ctx, activityID, 50)
		assert.NoError(t, err)

		// 失效缓存
		// Invalidate cache
		err = cacheManager.InvalidateActivity(ctx, activityID)
		assert.NoError(t, err)

		// 验证缓存已失效
		// Verify cache is invalidated
		var data map[string]string
		err = cacheManager.GetActivity(ctx, activityID, &data)
		assert.Error(t, err) // 应该返回错误，因为缓存已失效
	})
}

func TestCacheMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建Redis客户端
	// Create Redis client
	redisConfig := &redis.Config{
		Host:     "localhost",
		Port:     "6379",
		Database: 1,
	}

	redisClient, err := redis.NewClient(redisConfig)
	require.NoError(t, err)
	defer redisClient.Close()

	// 创建日志器
	// Create logger
	logConfig := &logger.Config{
		Level:  logger.DEBUG,
		Format: "text",
		Output: "stdout",
	}
	log := logger.NewLogger(logConfig)

	// 创建缓存管理器和监控器
	// Create cache manager and monitor
	cacheConfig := cache.DefaultCacheConfig()
	cacheManager := cache.NewCacheManager(redisClient, cacheConfig, log)

	monitorConfig := &cache.MonitorConfig{
		UpdateInterval: 100 * time.Millisecond,
		AlertThreshold: 0.8,
		EnableAlerts:   true,
	}
	monitor := cache.NewCacheMonitor(cacheManager, monitorConfig, log)

	// 启动监控
	// Start monitoring
	monitor.Start()
	defer monitor.Stop()

	// 模拟缓存操作
	// Simulate cache operations
	for i := 0; i < 10; i++ {
		monitor.RecordHit()
	}
	for i := 0; i < 3; i++ {
		monitor.RecordMiss()
	}
	for i := 0; i < 5; i++ {
		monitor.RecordSet()
	}

	// 等待监控更新
	// Wait for monitor update
	time.Sleep(200 * time.Millisecond)

	// 获取指标
	// Get metrics
	metrics := monitor.GetCurrentMetrics()
	assert.Equal(t, int64(10), metrics.HitCount)
	assert.Equal(t, int64(3), metrics.MissCount)
	assert.Equal(t, int64(5), metrics.SetCount)
	assert.Greater(t, metrics.HitRate, 0.7) // 10/(10+3) ≈ 0.77

	t.Logf("Cache metrics: Hit=%d, Miss=%d, HitRate=%.2f",
		metrics.HitCount, metrics.MissCount, metrics.HitRate)
}

func TestDataSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 注意：这个测试需要实际的PostgreSQL数据库
	// Note: This test requires actual PostgreSQL database
	t.Skip("Skipping data sync test - requires PostgreSQL setup")

	// 创建Redis客户端
	// Create Redis client
	redisConfig := &redis.Config{
		Host:     "localhost",
		Port:     "6379",
		Database: 1,
	}

	redisClient, err := redis.NewClient(redisConfig)
	require.NoError(t, err)
	defer redisClient.Close()

	// 创建日志器
	// Create logger
	logConfig := &logger.Config{
		Level:  logger.DEBUG,
		Format: "text",
		Output: "stdout",
	}
	log := logger.NewLogger(logConfig)

	// 创建同步配置
	// Create sync configuration
	syncConfig := sync.DefaultSyncConfig()
	syncConfig.DatabaseURL = "postgres://user:pass@localhost/flashsku_test?sslmode=disable"

	// 创建同步服务
	// Create sync service
	syncService, err := sync.NewSyncService(syncConfig, redisClient, log)
	require.NoError(t, err)
	defer syncService.Stop()

	// 启动同步服务
	// Start sync service
	err = syncService.Start()
	assert.NoError(t, err)

	// 等待同步完成
	// Wait for sync completion
	time.Sleep(2 * time.Second)

	// 验证同步状态
	// Verify sync status
	assert.True(t, syncService.IsRunning())

	// 手动触发同步
	// Manually trigger sync
	err = syncService.SyncNow()
	assert.NoError(t, err)
}
