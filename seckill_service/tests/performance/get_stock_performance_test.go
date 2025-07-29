package performance

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/activity"
	"github.com/flashsku/seckill/internal/api"
	"github.com/flashsku/seckill/internal/cache"
	"github.com/flashsku/seckill/internal/seckill"
)

// setupTestHandler 设置测试处理器
// setupTestHandler sets up test handler
func setupTestHandler() *api.SeckillHandler {
	mockRedis := &MockRedisClient{}
	mockLogger := &MockLogger{}

	// 创建秒杀服务
	// Create seckill service
	seckillConfig := seckill.DefaultSeckillConfig()
	seckillService := seckill.NewSeckillService(mockRedis, seckillConfig, nil, nil, mockLogger)

	// 创建活动验证器
	// Create activity validator
	validatorConfig := activity.DefaultValidatorConfig()
	activityValidator := activity.NewActivityValidator(mockRedis, validatorConfig, mockLogger)

	// 创建指标收集器
	// Create metrics collector
	metricsConfig := &cache.MetricsConfig{
		CollectInterval: time.Minute,
		RetentionPeriod: time.Hour,
	}
	metricsCollector := cache.NewMetricsCollector(mockRedis, metricsConfig, mockLogger)

	// 创建处理器
	// Create handler
	return api.NewSeckillHandler(seckillService, activityValidator, metricsCollector, mockLogger)
}

// TestGetStockPerformance 测试GetStock API性能
// TestGetStockPerformance tests GetStock API performance
func TestGetStockPerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := setupTestHandler()
	router := gin.New()
	router.GET("/api/v1/seckill/stock/:activity_id", handler.GetStock)

	// 性能测试参数
	// Performance test parameters
	const (
		numRequests = 1000
		maxLatency  = 50 * time.Millisecond // 目标延迟 < 50ms
	)

	var totalDuration time.Duration
	var successCount int
	latencies := make([]time.Duration, numRequests)

	t.Logf("Starting performance test with %d requests", numRequests)

	for i := 0; i < numRequests; i++ {
		// 创建请求
		// Create request
		req, _ := http.NewRequest("GET", "/api/v1/seckill/stock/activity123", nil)
		req.Header.Set("X-Request-ID", fmt.Sprintf("test-req-%d", i))

		// 记录开始时间
		// Record start time
		startTime := time.Now()

		// 执行请求
		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 计算延迟
		// Calculate latency
		latency := time.Since(startTime)
		latencies[i] = latency
		totalDuration += latency

		// 检查响应
		// Check response
		if w.Code == http.StatusOK {
			successCount++
		} else if i < 5 { // 只打印前5个错误
			t.Logf("Request %d failed with status %d: %s", i, w.Code, w.Body.String())
		}
	}

	// 计算统计数据
	// Calculate statistics
	avgLatency := totalDuration / time.Duration(numRequests)
	successRate := float64(successCount) / float64(numRequests)

	// 计算P99延迟
	// Calculate P99 latency
	// 简单排序找到P99
	// Simple sort to find P99
	for i := 0; i < len(latencies)-1; i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
	p99Index := int(float64(numRequests) * 0.99)
	p99Latency := latencies[p99Index]

	// 输出结果
	// Output results
	t.Logf("Performance Test Results:")
	t.Logf("  Total Requests: %d", numRequests)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Average Latency: %v", avgLatency)
	t.Logf("  P99 Latency: %v", p99Latency)
	t.Logf("  Total Duration: %v", totalDuration)

	// 验证性能指标
	// Verify performance metrics
	assert.Greater(t, successRate, 0.99, "Success rate should be > 99%")
	assert.Less(t, p99Latency, maxLatency, "P99 latency should be < 50ms")
	assert.Less(t, avgLatency, maxLatency/2, "Average latency should be < 25ms")

	// 计算QPS
	// Calculate QPS
	qps := float64(numRequests) / totalDuration.Seconds()
	t.Logf("  Estimated QPS: %.0f", qps)

	// 验证QPS目标 (这里是单线程测试，实际QPS会更高)
	// Verify QPS target (this is single-threaded test, actual QPS would be higher)
	assert.Greater(t, qps, 1000.0, "QPS should be > 1000 (single-threaded)")
}

// BenchmarkGetStock 基准测试
// BenchmarkGetStock benchmark test
func BenchmarkGetStock(b *testing.B) {
	gin.SetMode(gin.TestMode)

	handler := setupTestHandler()
	router := gin.New()
	router.GET("/api/v1/seckill/stock/:activity_id", handler.GetStock)

	req, _ := http.NewRequest("GET", "/api/v1/seckill/stock/activity123", nil)
	req.Header.Set("X-Request-ID", "benchmark-test")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Errorf("Expected status 200, got %d", w.Code)
		}
	}
}
