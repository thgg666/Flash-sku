package main

import (
	"fmt"
	"time"

	"github.com/flashsku/seckill/internal/ratelimit"
	"github.com/flashsku/seckill/pkg/logger"
)

// MockLogger 模拟日志器
// MockLogger mock logger
type MockLogger struct{}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	fmt.Printf("[DEBUG] %s\n", msg)
}
func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	fmt.Printf("[INFO] %s\n", msg)
}
func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	fmt.Printf("[WARN] %s\n", msg)
}
func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	fmt.Printf("[ERROR] %s\n", msg)
}
func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	fmt.Printf("[FATAL] %s\n", msg)
}
func (m *MockLogger) With(fields ...logger.Field) logger.Logger { return m }

func main() {
	fmt.Println("=== 令牌桶限流器示例 ===")
	fmt.Println("=== Token Bucket Rate Limiter Example ===")

	// 创建模拟日志器
	// Create mock logger
	mockLogger := &MockLogger{}

	// 示例1: 基础令牌桶
	// Example 1: Basic token bucket
	fmt.Println("\n1. 基础令牌桶测试 (Basic Token Bucket Test)")
	basicTokenBucketExample()

	// 示例2: 多级限流器
	// Example 2: Multi-level rate limiter
	fmt.Println("\n2. 多级限流器测试 (Multi-level Rate Limiter Test)")
	multiLevelRateLimiterExample(mockLogger)

	// 示例3: 动态配置更新
	// Example 3: Dynamic configuration update
	fmt.Println("\n3. 动态配置更新测试 (Dynamic Configuration Update Test)")
	dynamicConfigExample(mockLogger)

	fmt.Println("\n=== 示例完成 ===")
	fmt.Println("=== Examples Completed ===")
}

func basicTokenBucketExample() {
	// 创建令牌桶：容量5，每秒补充2个令牌
	// Create token bucket: capacity 5, refill 2 tokens per second
	config := &ratelimit.TokenBucketConfig{
		Capacity:   5,
		RefillRate: 2,
	}

	bucket := ratelimit.NewTokenBucket(config)

	fmt.Printf("初始令牌数: %d\n", bucket.GetTokens())

	// 快速消费令牌
	// Quickly consume tokens
	for i := 0; i < 7; i++ {
		allowed := bucket.Allow()
		tokens := bucket.GetTokens()
		fmt.Printf("请求 %d: %v, 剩余令牌: %d\n", i+1, allowed, tokens)
	}

	// 等待令牌补充
	// Wait for token refill
	fmt.Println("等待2秒令牌补充...")
	time.Sleep(2 * time.Second)

	fmt.Printf("补充后令牌数: %d\n", bucket.GetTokens())

	// 再次尝试请求
	// Try requests again
	for i := 0; i < 3; i++ {
		allowed := bucket.Allow()
		tokens := bucket.GetTokens()
		fmt.Printf("补充后请求 %d: %v, 剩余令牌: %d\n", i+1, allowed, tokens)
	}
}

func multiLevelRateLimiterExample(log logger.Logger) {
	// 创建多级限流器配置
	// Create multi-level rate limiter configuration
	config := &ratelimit.RateLimiterConfig{
		GlobalConfig: &ratelimit.TokenBucketConfig{
			Capacity:   10,
			RefillRate: 5,
		},
		IPConfig: &ratelimit.TokenBucketConfig{
			Capacity:   3,
			RefillRate: 1,
		},
		UserConfig: &ratelimit.TokenBucketConfig{
			Capacity:   2,
			RefillRate: 1,
		},
		CleanupInterval: 1 * time.Minute,
	}

	limiter := ratelimit.NewRateLimiter(config, log)
	defer limiter.Close()

	// 模拟不同用户和IP的请求
	// Simulate requests from different users and IPs
	testCases := []struct {
		ip     string
		userID string
		desc   string
	}{
		{"192.168.1.1", "user1", "用户1从IP1"},
		{"192.168.1.1", "user1", "用户1从IP1 (重复)"},
		{"192.168.1.1", "user2", "用户2从IP1"},
		{"192.168.1.1", "user3", "用户3从IP1 (应该被IP限流)"},
		{"192.168.1.2", "user1", "用户1从IP2 (应该被用户限流)"},
		{"192.168.1.2", "user4", "用户4从IP2"},
	}

	for i, tc := range testCases {
		result := limiter.Allow(tc.ip, tc.userID)
		status := "✅ 允许"
		if !result.Allowed {
			status = fmt.Sprintf("❌ 拒绝 (%s)", result.Level)
		}

		fmt.Printf("请求 %d - %s: %s\n", i+1, tc.desc, status)
		if !result.Allowed {
			fmt.Printf("  原因: %s, 建议重试: %ds\n", result.Reason, result.RetryAfter)
		}
	}
}

func dynamicConfigExample(log logger.Logger) {
	// 创建限流器
	// Create rate limiter
	limiter := ratelimit.NewRateLimiter(nil, log)
	defer limiter.Close()

	fmt.Println("初始配置测试:")

	// 测试初始配置
	// Test initial configuration
	for i := 0; i < 3; i++ {
		result := limiter.AllowLevel(ratelimit.LimitLevelIP, "test-ip")
		fmt.Printf("  请求 %d: %v\n", i+1, result.Allowed)
	}

	// 动态更新IP配置 - 更严格的限制
	// Dynamically update IP configuration - stricter limit
	newConfig := &ratelimit.TokenBucketConfig{
		Capacity:   1,
		RefillRate: 1,
	}

	fmt.Println("\n更新IP配置为更严格的限制 (容量: 1, 速率: 1)")
	err := limiter.UpdateConfig(ratelimit.LimitLevelIP, newConfig)
	if err != nil {
		fmt.Printf("配置更新失败: %v\n", err)
		return
	}

	fmt.Println("更新后测试:")

	// 测试新配置
	// Test new configuration
	for i := 0; i < 3; i++ {
		result := limiter.AllowLevel(ratelimit.LimitLevelIP, "new-test-ip")
		fmt.Printf("  请求 %d: %v\n", i+1, result.Allowed)
	}

	// 等待令牌补充
	// Wait for token refill
	fmt.Println("\n等待1秒令牌补充...")
	time.Sleep(1 * time.Second)

	result := limiter.AllowLevel(ratelimit.LimitLevelIP, "new-test-ip")
	fmt.Printf("补充后请求: %v\n", result.Allowed)
}
