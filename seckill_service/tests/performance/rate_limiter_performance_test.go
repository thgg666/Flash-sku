package performance

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/ratelimit"
)

// TestRateLimiterPerformance 测试限流器性能
// TestRateLimiterPerformance tests rate limiter performance
func TestRateLimiterPerformance(t *testing.T) {
	mockLogger := &MockLogger{}

	// 配置高性能限流器
	// Configure high-performance rate limiter
	config := &ratelimit.RateLimiterConfig{
		GlobalConfig: &ratelimit.TokenBucketConfig{
			Capacity:   10000,
			RefillRate: 5000,
		},
		IPConfig: &ratelimit.TokenBucketConfig{
			Capacity:   1000,
			RefillRate: 500,
		},
		UserConfig: &ratelimit.TokenBucketConfig{
			Capacity:   100,
			RefillRate: 50,
		},
		CleanupInterval: 5 * time.Minute,
	}

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	// 创建指标收集器
	// Create metrics collector
	metricsCollector := ratelimit.NewMetricsCollector(time.Minute, mockLogger)
	instrumentedLimiter := ratelimit.NewInstrumentedRateLimiter(limiter, metricsCollector, mockLogger)

	// 性能测试参数
	// Performance test parameters
	const (
		numGoroutines        = 100
		requestsPerGoroutine = 100
		totalRequests        = numGoroutines * requestsPerGoroutine
	)

	var (
		allowedCount int64
		blockedCount int64
		totalLatency int64
		wg           sync.WaitGroup
	)

	t.Logf("Starting rate limiter performance test: %d goroutines, %d requests each",
		numGoroutines, requestsPerGoroutine)

	startTime := time.Now()

	// 启动并发测试
	// Start concurrent test
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for r := 0; r < requestsPerGoroutine; r++ {
				ip := fmt.Sprintf("192.168.%d.%d", goroutineID%256, r%256)
				userID := fmt.Sprintf("user_%d_%d", goroutineID, r)

				reqStart := time.Now()
				result := instrumentedLimiter.Allow(ip, userID)
				latency := time.Since(reqStart)

				atomic.AddInt64(&totalLatency, latency.Nanoseconds())

				if result.Allowed {
					atomic.AddInt64(&allowedCount, 1)
				} else {
					atomic.AddInt64(&blockedCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// 计算性能指标
	// Calculate performance metrics
	qps := float64(totalRequests) / totalDuration.Seconds()
	avgLatency := time.Duration(totalLatency / int64(totalRequests))
	allowRate := float64(allowedCount) / float64(totalRequests)

	// 获取限流器指标
	// Get rate limiter metrics
	metrics := instrumentedLimiter.GetMetrics()
	summary := instrumentedLimiter.GetSummary()

	// 输出结果
	// Output results
	t.Logf("Rate Limiter Performance Test Results:")
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Allowed: %d", allowedCount)
	t.Logf("  Blocked: %d", blockedCount)
	t.Logf("  Allow Rate: %.2f%%", allowRate*100)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  QPS: %.0f", qps)
	t.Logf("  Average Latency: %v", avgLatency)
	t.Logf("  Metrics Avg Response Time: %.2fms", metrics.AvgResponseTime)
	t.Logf("  Metrics Max Response Time: %.2fms", metrics.MaxResponseTime)

	// 验证性能指标
	// Verify performance metrics
	assert.Greater(t, qps, 10000.0, "QPS should be > 10,000")
	assert.Less(t, avgLatency, 1*time.Millisecond, "Average latency should be < 1ms")
	assert.Greater(t, allowRate, 0.5, "Allow rate should be > 50%")

	// 输出详细指标摘要
	// Output detailed metrics summary
	t.Logf("Detailed Metrics Summary:")
	for key, value := range summary {
		t.Logf("  %s: %v", key, value)
	}

	// 检查性能告警
	// Check performance alerts
	alerts := instrumentedLimiter.CheckAlerts()
	if len(alerts) > 0 {
		t.Logf("Performance Alerts:")
		for _, alert := range alerts {
			t.Logf("  %s: %s (threshold: %.2f, current: %.2f)",
				alert.Type, alert.Message, alert.Threshold, alert.CurrentValue)
		}
	} else {
		t.Logf("No performance alerts detected")
	}
}

// BenchmarkRateLimiter 限流器基准测试
// BenchmarkRateLimiter rate limiter benchmark test
func BenchmarkRateLimiter(b *testing.B) {
	mockLogger := &MockLogger{}

	config := &ratelimit.RateLimiterConfig{
		GlobalConfig: &ratelimit.TokenBucketConfig{
			Capacity:   int64(b.N * 2),
			RefillRate: int64(b.N),
		},
		IPConfig: &ratelimit.TokenBucketConfig{
			Capacity:   1000,
			RefillRate: 500,
		},
		UserConfig: &ratelimit.TokenBucketConfig{
			Capacity:   100,
			RefillRate: 50,
		},
	}

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ip := fmt.Sprintf("192.168.1.%d", i%256)
			userID := fmt.Sprintf("user_%d", i%1000)

			result := limiter.Allow(ip, userID)

			// 避免编译器优化
			// Prevent compiler optimization
			_ = result.Allowed

			i++
		}
	})
}

// TestRateLimiterMemoryUsage 测试限流器内存使用
// TestRateLimiterMemoryUsage tests rate limiter memory usage
func TestRateLimiterMemoryUsage(t *testing.T) {
	mockLogger := &MockLogger{}
	limiter := ratelimit.NewRateLimiter(nil, mockLogger)
	defer limiter.Close()

	// 创建大量不同的IP和用户
	// Create many different IPs and users
	const numEntities = 10000

	t.Logf("Creating %d unique IP and user buckets", numEntities)

	startTime := time.Now()

	for i := 0; i < numEntities; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		userID := fmt.Sprintf("user_%d", i)

		// 触发桶的创建
		// Trigger bucket creation
		limiter.Allow(ip, userID)
	}

	creationTime := time.Since(startTime)

	t.Logf("Created %d buckets in %v", numEntities*2, creationTime)
	t.Logf("Average creation time per bucket: %v", creationTime/time.Duration(numEntities*2))

	// 测试访问性能
	// Test access performance
	accessStart := time.Now()

	for i := 0; i < 1000; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		userID := fmt.Sprintf("user_%d", i)

		result := limiter.Allow(ip, userID)
		_ = result.Allowed
	}

	accessTime := time.Since(accessStart)

	t.Logf("1000 accesses to existing buckets took %v", accessTime)
	t.Logf("Average access time: %v", accessTime/1000)

	// 验证性能要求
	// Verify performance requirements
	avgCreationTime := creationTime / time.Duration(numEntities*2)
	avgAccessTime := accessTime / 1000

	assert.Less(t, avgCreationTime, 100*time.Microsecond, "Bucket creation should be < 100µs")
	assert.Less(t, avgAccessTime, 10*time.Microsecond, "Bucket access should be < 10µs")
}

// TestRateLimiterConcurrentAccess 测试限流器并发访问
// TestRateLimiterConcurrentAccess tests rate limiter concurrent access
func TestRateLimiterConcurrentAccess(t *testing.T) {
	mockLogger := &MockLogger{}

	config := &ratelimit.RateLimiterConfig{
		GlobalConfig: &ratelimit.TokenBucketConfig{
			Capacity:   1000,
			RefillRate: 500,
		},
		IPConfig: &ratelimit.TokenBucketConfig{
			Capacity:   10,
			RefillRate: 5,
		},
		UserConfig: &ratelimit.TokenBucketConfig{
			Capacity:   5,
			RefillRate: 2,
		},
	}

	limiter := ratelimit.NewRateLimiter(config, mockLogger)
	defer limiter.Close()

	const (
		numGoroutines        = 50
		requestsPerGoroutine = 20
		sharedIP             = "192.168.1.1"
		sharedUser           = "shared_user"
	)

	var (
		allowedCount int64
		blockedCount int64
		wg           sync.WaitGroup
	)

	t.Logf("Testing concurrent access: %d goroutines accessing shared resources", numGoroutines)

	startTime := time.Now()

	// 所有goroutine访问相同的IP和用户，测试并发安全性
	// All goroutines access same IP and user to test concurrent safety
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for r := 0; r < requestsPerGoroutine; r++ {
				result := limiter.Allow(sharedIP, sharedUser)

				if result.Allowed {
					atomic.AddInt64(&allowedCount, 1)
				} else {
					atomic.AddInt64(&blockedCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalRequests := int64(numGoroutines * requestsPerGoroutine)

	t.Logf("Concurrent Access Test Results:")
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Allowed: %d", allowedCount)
	t.Logf("  Blocked: %d", blockedCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  QPS: %.0f", float64(totalRequests)/duration.Seconds())

	// 验证并发安全性 - 允许的请求数不应超过桶容量
	// Verify concurrent safety - allowed requests should not exceed bucket capacity
	maxAllowed := config.IPConfig.Capacity + config.UserConfig.Capacity // 最严格的限制
	assert.LessOrEqual(t, allowedCount, maxAllowed,
		"Allowed requests should not exceed the strictest limit")

	// 验证总数正确
	// Verify total count is correct
	assert.Equal(t, totalRequests, allowedCount+blockedCount,
		"Total requests should equal allowed + blocked")
}
