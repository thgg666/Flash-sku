package unit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flashsku/seckill/internal/ratelimit"
	"github.com/flashsku/seckill/pkg/logger"
)

// MockLogger 模拟日志器
// MockLogger mock logger
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...logger.Field)  {}
func (m *MockLogger) Info(msg string, fields ...logger.Field)   {}
func (m *MockLogger) Warn(msg string, fields ...logger.Field)   {}
func (m *MockLogger) Error(msg string, fields ...logger.Field)  {}
func (m *MockLogger) Fatal(msg string, fields ...logger.Field)  {}
func (m *MockLogger) With(fields ...logger.Field) logger.Logger { return m }

func TestTokenBucket_BasicFunctionality(t *testing.T) {
	// 测试基本的令牌桶功能
	// Test basic token bucket functionality
	config := &ratelimit.TokenBucketConfig{
		Capacity:   10,
		RefillRate: 5, // 每秒补充5个令牌
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 初始状态应该是满的
	// Initial state should be full
	assert.Equal(t, int64(10), bucket.GetTokens())
	assert.Equal(t, int64(10), bucket.GetCapacity())
	assert.Equal(t, int64(5), bucket.GetRefillRate())

	// 消费令牌
	// Consume tokens
	assert.True(t, bucket.Allow())
	assert.Equal(t, int64(9), bucket.GetTokens())

	// 消费多个令牌
	// Consume multiple tokens
	assert.True(t, bucket.AllowN(5))
	assert.Equal(t, int64(4), bucket.GetTokens())

	// 尝试消费超过可用令牌数
	// Try to consume more than available
	assert.False(t, bucket.AllowN(5))
	assert.Equal(t, int64(4), bucket.GetTokens()) // 令牌数不应该改变
}

func TestTokenBucket_Refill(t *testing.T) {
	// 测试令牌补充功能
	// Test token refill functionality
	config := &ratelimit.TokenBucketConfig{
		Capacity:   10,
		RefillRate: 2, // 每秒补充2个令牌
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 消费所有令牌
	// Consume all tokens
	assert.True(t, bucket.AllowN(10))
	assert.Equal(t, int64(0), bucket.GetTokens())

	// 等待1秒，应该补充2个令牌
	// Wait 1 second, should refill 2 tokens
	time.Sleep(1100 * time.Millisecond) // 稍微多等一点避免时间精度问题

	// 检查令牌是否补充
	// Check if tokens are refilled
	tokens := bucket.GetTokens()
	assert.GreaterOrEqual(t, tokens, int64(2))
	assert.LessOrEqual(t, tokens, int64(3)) // 允许一些时间误差

	// 再等待1秒
	// Wait another second
	time.Sleep(1100 * time.Millisecond)
	tokens = bucket.GetTokens()
	assert.GreaterOrEqual(t, tokens, int64(4))
	assert.LessOrEqual(t, tokens, int64(5))
}

func TestTokenBucket_CapacityLimit(t *testing.T) {
	// 测试容量限制
	// Test capacity limit
	config := &ratelimit.TokenBucketConfig{
		Capacity:   5,
		RefillRate: 10, // 高补充速率
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 等待足够长时间让令牌补充
	// Wait long enough for token refill
	time.Sleep(2 * time.Second)

	// 令牌数不应该超过容量
	// Token count should not exceed capacity
	assert.Equal(t, int64(5), bucket.GetTokens())
}

func TestTokenBucket_UpdateConfig(t *testing.T) {
	// 测试动态配置更新
	// Test dynamic configuration update
	config := &ratelimit.TokenBucketConfig{
		Capacity:   10,
		RefillRate: 5,
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 消费一些令牌
	// Consume some tokens
	assert.True(t, bucket.AllowN(5))
	assert.Equal(t, int64(5), bucket.GetTokens())

	// 更新配置 - 减少容量
	// Update config - reduce capacity
	newConfig := &ratelimit.TokenBucketConfig{
		Capacity:   3,
		RefillRate: 2,
	}

	bucket.UpdateConfig(newConfig)

	// 令牌数应该调整到新容量
	// Token count should adjust to new capacity
	assert.Equal(t, int64(3), bucket.GetTokens())
	assert.Equal(t, int64(3), bucket.GetCapacity())
	assert.Equal(t, int64(2), bucket.GetRefillRate())
}

func TestTokenBucket_ResetAndFill(t *testing.T) {
	// 测试重置和填充功能
	// Test reset and fill functionality
	config := &ratelimit.TokenBucketConfig{
		Capacity:   10,
		RefillRate: 5,
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 消费一些令牌
	// Consume some tokens
	assert.True(t, bucket.AllowN(7))
	assert.Equal(t, int64(3), bucket.GetTokens())

	// 重置桶
	// Reset bucket
	bucket.Reset()
	assert.Equal(t, int64(0), bucket.GetTokens())

	// 填满桶
	// Fill bucket
	bucket.Fill()
	assert.Equal(t, int64(10), bucket.GetTokens())
}

func TestTokenBucket_Stats(t *testing.T) {
	// 测试统计信息
	// Test statistics
	config := &ratelimit.TokenBucketConfig{
		Capacity:   10,
		RefillRate: 5,
	}

	bucket := ratelimit.NewTokenBucket(config)

	// 消费一些令牌
	// Consume some tokens
	assert.True(t, bucket.AllowN(6))

	stats := bucket.GetStats()
	assert.Equal(t, int64(10), stats.Capacity)
	assert.Equal(t, int64(4), stats.CurrentTokens)
	assert.Equal(t, int64(5), stats.RefillRate)
	assert.Equal(t, 0.6, stats.Utilization) // (10-4)/10 = 0.6
	assert.False(t, stats.LastRefill.IsZero())
}

func TestRateLimiter_MultiLevel(t *testing.T) {
	// 测试多级限流器
	// Test multi-level rate limiter
	mockLogger := &MockLogger{}
	config := ratelimit.GetDefaultRateLimiterConfig()

	// 设置较小的限制便于测试，但要考虑多级限流的影响
	// Set smaller limits for testing, considering multi-level impact
	config.GlobalConfig.Capacity = 10 // 增加全局容量
	config.GlobalConfig.RefillRate = 1
	config.IPConfig.Capacity = 10 // 增加IP容量，避免IP限流干扰
	config.IPConfig.RefillRate = 1
	config.UserConfig.Capacity = 10 // 增加用户容量，避免用户限流干扰
	config.UserConfig.RefillRate = 1

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	// 测试全局限流 - 使用不同的IP和用户避免其他级别限流
	// Test global rate limiting - use different IPs and users to avoid other level limits
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i+1)
		user := fmt.Sprintf("user%d", i+1)
		result := limiter.Allow(ip, user)
		assert.True(t, result.Allowed, "Request %d should be allowed", i+1)
	}

	// 第11个请求应该被全局限流阻止
	// 11th request should be blocked by global rate limit
	result := limiter.Allow("192.168.1.100", "user100")
	assert.False(t, result.Allowed)
	assert.Equal(t, ratelimit.LimitLevelGlobal, result.Level)
}

func TestRateLimiter_IPLevel(t *testing.T) {
	// 测试IP级别限流
	// Test IP level rate limiting
	mockLogger := &MockLogger{}
	config := ratelimit.GetDefaultRateLimiterConfig()

	// 设置全局限制很高，IP限制较低
	// Set high global limit, low IP limit
	config.GlobalConfig.Capacity = 100
	config.IPConfig.Capacity = 2
	config.UserConfig.Capacity = 10

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	// IP1的前2个请求应该通过
	// First 2 requests from IP1 should pass
	result1 := limiter.Allow("192.168.1.1", "user1")
	assert.True(t, result1.Allowed)

	result2 := limiter.Allow("192.168.1.1", "user2")
	assert.True(t, result2.Allowed)

	// IP1的第3个请求应该被IP限流阻止
	// 3rd request from IP1 should be blocked by IP rate limit
	result3 := limiter.Allow("192.168.1.1", "user3")
	assert.False(t, result3.Allowed)
	assert.Equal(t, ratelimit.LimitLevelIP, result3.Level)

	// 不同IP的请求应该通过
	// Request from different IP should pass
	result4 := limiter.Allow("192.168.1.2", "user4")
	assert.True(t, result4.Allowed)
}

func TestRateLimiter_UserLevel(t *testing.T) {
	// 测试用户级别限流
	// Test user level rate limiting
	mockLogger := &MockLogger{}
	config := ratelimit.GetDefaultRateLimiterConfig()

	// 设置全局和IP限制很高，用户限制较低
	// Set high global and IP limits, low user limit
	config.GlobalConfig.Capacity = 100
	config.IPConfig.Capacity = 100
	config.UserConfig.Capacity = 1

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	// 用户1的第1个请求应该通过
	// First request from user1 should pass
	result1 := limiter.Allow("192.168.1.1", "user1")
	assert.True(t, result1.Allowed)

	// 用户1的第2个请求应该被用户限流阻止
	// 2nd request from user1 should be blocked by user rate limit
	result2 := limiter.Allow("192.168.1.2", "user1") // 不同IP，相同用户
	assert.False(t, result2.Allowed)
	assert.Equal(t, ratelimit.LimitLevelUser, result2.Level)

	// 不同用户的请求应该通过
	// Request from different user should pass
	result3 := limiter.Allow("192.168.1.1", "user2")
	assert.True(t, result3.Allowed)
}

func TestRateLimiter_AllowLevel(t *testing.T) {
	// 测试特定级别的限流检查
	// Test specific level rate limit check
	mockLogger := &MockLogger{}
	config := ratelimit.GetDefaultRateLimiterConfig()
	config.IPConfig.Capacity = 2

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	// 测试IP级别限流
	// Test IP level rate limiting
	result1 := limiter.AllowLevel(ratelimit.LimitLevelIP, "192.168.1.1")
	assert.True(t, result1.Allowed)

	result2 := limiter.AllowLevel(ratelimit.LimitLevelIP, "192.168.1.1")
	assert.True(t, result2.Allowed)

	result3 := limiter.AllowLevel(ratelimit.LimitLevelIP, "192.168.1.1")
	assert.False(t, result3.Allowed)
	assert.Equal(t, ratelimit.LimitLevelIP, result3.Level)
}

func TestRateLimiter_UpdateConfig(t *testing.T) {
	// 测试动态配置更新
	// Test dynamic configuration update
	mockLogger := &MockLogger{}
	limiter := ratelimit.NewRateLimiter(nil, mockLogger)
	defer limiter.Close()

	// 更新IP配置
	// Update IP configuration
	newIPConfig := &ratelimit.TokenBucketConfig{
		Capacity:   1,
		RefillRate: 1,
	}

	err := limiter.UpdateConfig(ratelimit.LimitLevelIP, newIPConfig)
	require.NoError(t, err)

	// 验证新配置生效
	// Verify new configuration takes effect
	result1 := limiter.AllowLevel(ratelimit.LimitLevelIP, "test-ip")
	assert.True(t, result1.Allowed)

	result2 := limiter.AllowLevel(ratelimit.LimitLevelIP, "test-ip")
	assert.False(t, result2.Allowed)

	// 测试无效级别
	// Test invalid level
	err = limiter.UpdateConfig("invalid", newIPConfig)
	assert.Error(t, err)
}
