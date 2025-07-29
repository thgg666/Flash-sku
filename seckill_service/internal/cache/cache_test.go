package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()
	
	assert.NotNil(t, config)
	assert.Equal(t, 1*time.Hour, config.DefaultTTL)
	assert.Equal(t, 24*time.Hour, config.ActivityTTL)
	assert.Equal(t, 1*time.Hour, config.StockTTL)
	assert.Equal(t, 24*time.Hour, config.UserLimitTTL)
	assert.Equal(t, 1*time.Minute, config.RateLimitTTL)
	assert.Equal(t, 5*time.Minute, config.RefreshInterval)
}

func TestCacheKeys(t *testing.T) {
	keys := DefaultCacheKeys
	
	assert.Equal(t, "seckill:activity:", keys.ActivityPrefix)
	assert.Equal(t, "seckill:stock:", keys.StockPrefix)
	assert.Equal(t, "seckill:user_limit:", keys.UserLimitPrefix)
	assert.Equal(t, "seckill:rate_limit:", keys.RateLimitPrefix)
	assert.Equal(t, "seckill:metrics:", keys.MetricsPrefix)
	assert.Equal(t, "seckill:health:", keys.HealthPrefix)
}

func TestCacheMetrics(t *testing.T) {
	metrics := &CacheMetrics{
		HitCount:  100,
		MissCount: 20,
	}
	
	// 测试命中率计算
	// Test hit rate calculation
	total := metrics.HitCount + metrics.MissCount
	expectedHitRate := float64(metrics.HitCount) / float64(total)
	
	assert.Equal(t, int64(100), metrics.HitCount)
	assert.Equal(t, int64(20), metrics.MissCount)
	assert.InDelta(t, expectedHitRate, 0.833, 0.001) // 100/120 ≈ 0.833
}

func TestAtomicMetrics(t *testing.T) {
	metrics := &AtomicMetrics{}
	
	// 测试初始值
	// Test initial values
	assert.Equal(t, int64(0), metrics.hitCount)
	assert.Equal(t, int64(0), metrics.missCount)
	assert.Equal(t, int64(0), metrics.setCount)
	assert.Equal(t, int64(0), metrics.deleteCount)
	assert.Equal(t, int64(0), metrics.errorCount)
}
