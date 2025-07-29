package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/activity"
	"github.com/flashsku/seckill/internal/api"
	"github.com/flashsku/seckill/internal/cache"
	"github.com/flashsku/seckill/internal/seckill"
)

// MockSeckillRedisClient 模拟秒杀Redis客户端
// MockSeckillRedisClient mock seckill Redis client
type MockSeckillRedisClient struct {
	stock      int64
	userLimits map[string]int64
	purchases  map[string]int64
	mutex      sync.RWMutex
}

func NewMockSeckillRedisClient(initialStock int64) *MockSeckillRedisClient {
	return &MockSeckillRedisClient{
		stock:      initialStock,
		userLimits: make(map[string]int64),
		purchases:  make(map[string]int64),
	}
}

func (m *MockSeckillRedisClient) Get(ctx context.Context, key string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 模拟库存查询
	// Mock stock query
	if key == "seckill:stock:activity123" {
		return fmt.Sprintf("%d", m.stock), nil
	}

	// 模拟活动信息
	// Mock activity info
	if key == "seckill:activity:activity123" {
		futureTime := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
		pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		return fmt.Sprintf(`{
			"id":"activity123",
			"name":"Performance Test Activity",
			"status":"active",
			"start_time":"%s",
			"end_time":"%s",
			"total_stock":10000,
			"sold_count":%d,
			"price":99.99,
			"user_limit":5
		}`, pastTime, futureTime, 10000-m.stock), nil
	}

	// 模拟用户购买记录
	// Mock user purchase records
	if key[:len("seckill:user:purchases:")] == "seckill:user:purchases:" {
		userID := key[len("seckill:user:purchases:"):]
		if purchases, exists := m.purchases[userID]; exists {
			return fmt.Sprintf("%d", purchases), nil
		}
		return "0", nil
	}

	return "", fmt.Errorf("key not found")
}

func (m *MockSeckillRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (m *MockSeckillRedisClient) Del(ctx context.Context, keys ...string) error {
	return nil
}

func (m *MockSeckillRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 1, nil
}

func (m *MockSeckillRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 模拟用户购买计数增加
	// Mock user purchase count increment
	if key[:len("seckill:user:purchases:")] == "seckill:user:purchases:" {
		userID := key[len("seckill:user:purchases:"):]
		m.purchases[userID]++
		return m.purchases[userID], nil
	}

	return 1, nil
}

func (m *MockSeckillRedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockSeckillRedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return value, nil
}

func (m *MockSeckillRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return time.Hour, nil
}

func (m *MockSeckillRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *MockSeckillRedisClient) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 模拟Lua脚本执行 - 原子性库存扣减
	// Mock Lua script execution - atomic stock deduction
	if m.stock > 0 {
		m.stock--
		return []interface{}{int64(1), "success", map[string]interface{}{"new_stock": m.stock}}, nil
	}

	return []interface{}{int64(0), "out_of_stock"}, nil
}

func (m *MockSeckillRedisClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	return "mock_sha1", nil
}

func (m *MockSeckillRedisClient) EvalSHA(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return m.Eval(ctx, "", keys, args...)
}

func (m *MockSeckillRedisClient) ScriptExists(ctx context.Context, sha1 string) (bool, error) {
	return true, nil
}

func (m *MockSeckillRedisClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockSeckillRedisClient) Close() error {
	return nil
}

// setupSeckillTestHandler 设置秒杀测试处理器
// setupSeckillTestHandler sets up seckill test handler
func setupSeckillTestHandler(initialStock int64) (*api.SeckillHandler, *MockSeckillRedisClient) {
	mockRedis := NewMockSeckillRedisClient(initialStock)
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
	handler := api.NewSeckillHandler(seckillService, activityValidator, metricsCollector, mockLogger)

	return handler, mockRedis
}

// SeckillRequest 秒杀请求结构
// SeckillRequest seckill request structure
type SeckillRequest struct {
	UserID         string `json:"user_id"`
	PurchaseAmount int    `json:"purchase_amount"`
	UserLimit      int    `json:"user_limit,omitempty"`
}

// TestSeckillPerformance 测试秒杀性能
// TestSeckillPerformance tests seckill performance
func TestSeckillPerformance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 初始库存
	// Initial stock
	const initialStock = 1000
	const numRequests = 2000 // 超过库存数量，测试并发控制
	const targetQPS = 1000

	handler, mockRedis := setupSeckillTestHandler(initialStock)
	router := gin.New()
	router.POST("/api/v1/seckill/:activity_id", handler.ProcessSeckill)

	var (
		successCount  int64
		failureCount  int64
		totalDuration time.Duration
		wg            sync.WaitGroup
		startTime     = time.Now()
	)

	t.Logf("Starting seckill performance test with %d requests, initial stock: %d", numRequests, initialStock)

	// 并发执行秒杀请求
	// Execute seckill requests concurrently
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			// 创建请求
			// Create request
			reqBody := SeckillRequest{
				UserID:         fmt.Sprintf("user_%d", requestID),
				PurchaseAmount: 1,
				UserLimit:      5,
			}

			jsonBody, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/seckill/activity123", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Request-ID", fmt.Sprintf("perf-test-%d", requestID))

			// 执行请求
			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 统计结果
			// Count results
			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failureCount, 1)
				// 打印前几个失败的详细信息
				// Print details of first few failures
				if requestID < 5 {
					t.Logf("Request %d failed with status %d: %s", requestID, w.Code, w.Body.String())
				}
			}
		}(i)
	}

	// 等待所有请求完成
	// Wait for all requests to complete
	wg.Wait()
	totalDuration = time.Since(startTime)

	// 计算统计数据
	// Calculate statistics
	actualQPS := float64(numRequests) / totalDuration.Seconds()
	successRate := float64(successCount) / float64(numRequests)

	// 获取最终库存
	// Get final stock
	finalStock := mockRedis.stock

	// 输出结果
	// Output results
	t.Logf("Seckill Performance Test Results:")
	t.Logf("  Total Requests: %d", numRequests)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Failure Count: %d", failureCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Total Duration: %v", totalDuration)
	t.Logf("  Actual QPS: %.0f", actualQPS)
	t.Logf("  Initial Stock: %d", initialStock)
	t.Logf("  Final Stock: %d", finalStock)
	t.Logf("  Stock Consumed: %d", initialStock-int(finalStock))

	// 验证性能指标
	// Verify performance metrics
	assert.Greater(t, actualQPS, float64(targetQPS), "QPS should be > %d", targetQPS)
	assert.Equal(t, int64(initialStock), successCount+finalStock, "Success count + final stock should equal initial stock")
	assert.GreaterOrEqual(t, finalStock, int64(0), "Final stock should not be negative")

	// 验证没有超卖
	// Verify no overselling
	expectedFailures := int64(numRequests) - int64(initialStock)
	if expectedFailures > 0 {
		assert.GreaterOrEqual(t, failureCount, expectedFailures, "Should have appropriate failure count when requests exceed stock")
	}
}

// BenchmarkSeckill 秒杀基准测试
// BenchmarkSeckill seckill benchmark test
func BenchmarkSeckill(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 使用大库存避免库存不足
	// Use large stock to avoid stock shortage
	handler, _ := setupSeckillTestHandler(int64(b.N * 2))
	router := gin.New()
	router.POST("/api/v1/seckill/:activity_id", handler.ProcessSeckill)

	reqBody := SeckillRequest{
		UserID:         "benchmark_user",
		PurchaseAmount: 1,
		UserLimit:      5,
	}

	jsonBody, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/seckill/activity123", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", fmt.Sprintf("bench-%d", i))

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusConflict {
			b.Errorf("Unexpected status code: %d", w.Code)
		}
	}
}

// TestSeckillStressTest 秒杀压力测试
// TestSeckillStressTest seckill stress test
func TestSeckillStressTest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 压力测试参数
	// Stress test parameters
	const (
		initialStock         = 500
		numGoroutines        = 100
		requestsPerGoroutine = 50
		totalRequests        = numGoroutines * requestsPerGoroutine
	)

	handler, mockRedis := setupSeckillTestHandler(initialStock)
	router := gin.New()
	router.POST("/api/v1/seckill/:activity_id", handler.ProcessSeckill)

	var (
		successCount int64
		errorCount   int64
		wg           sync.WaitGroup
		startTime    = time.Now()
	)

	t.Logf("Starting stress test: %d goroutines, %d requests each, total: %d",
		numGoroutines, requestsPerGoroutine, totalRequests)

	// 启动多个goroutine并发执行
	// Start multiple goroutines for concurrent execution
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for r := 0; r < requestsPerGoroutine; r++ {
				reqBody := SeckillRequest{
					UserID:         fmt.Sprintf("stress_user_%d_%d", goroutineID, r),
					PurchaseAmount: 1,
					UserLimit:      5,
				}

				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest("POST", "/api/v1/seckill/activity123", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Request-ID", fmt.Sprintf("stress-%d-%d", goroutineID, r))

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
					// 打印前几个失败的详细信息
					// Print details of first few failures
					if goroutineID == 0 && r < 3 {
						t.Logf("Stress test request %d-%d failed with status %d: %s", goroutineID, r, w.Code, w.Body.String())
					}
				}
			}
		}(g)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 计算结果
	// Calculate results
	qps := float64(totalRequests) / duration.Seconds()
	successRate := float64(successCount) / float64(totalRequests)
	finalStock := mockRedis.stock

	t.Logf("Stress Test Results:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Requests per Goroutine: %d", requestsPerGoroutine)
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Success Count: %d", successCount)
	t.Logf("  Error Count: %d", errorCount)
	t.Logf("  Success Rate: %.2f%%", successRate*100)
	t.Logf("  Duration: %v", duration)
	t.Logf("  QPS: %.0f", qps)
	t.Logf("  Initial Stock: %d", initialStock)
	t.Logf("  Final Stock: %d", finalStock)
	t.Logf("  Stock Consumed: %d", initialStock-int(finalStock))

	// 验证结果
	// Verify results
	assert.Greater(t, qps, 1000.0, "QPS should be > 1000")
	assert.Equal(t, int64(initialStock), successCount+finalStock, "No overselling should occur")
	assert.GreaterOrEqual(t, finalStock, int64(0), "Stock should not be negative")
}

// TestSeckillLatencyDistribution 测试秒杀延迟分布
// TestSeckillLatencyDistribution tests seckill latency distribution
func TestSeckillLatencyDistribution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const numRequests = 1000
	handler, _ := setupSeckillTestHandler(numRequests)
	router := gin.New()
	router.POST("/api/v1/seckill/:activity_id", handler.ProcessSeckill)

	latencies := make([]time.Duration, numRequests)
	var wg sync.WaitGroup

	t.Logf("Testing latency distribution with %d requests", numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			reqBody := SeckillRequest{
				UserID:         fmt.Sprintf("latency_user_%d", requestID),
				PurchaseAmount: 1,
				UserLimit:      5,
			}

			jsonBody, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/seckill/activity123", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			start := time.Now()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			latencies[requestID] = time.Since(start)
		}(i)
	}

	wg.Wait()

	// 计算延迟统计
	// Calculate latency statistics
	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := totalLatency / time.Duration(numRequests)

	// 简单排序计算百分位数
	// Simple sort to calculate percentiles
	for i := 0; i < len(latencies)-1; i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	p50 := latencies[int(float64(numRequests)*0.5)]
	p95 := latencies[int(float64(numRequests)*0.95)]
	p99 := latencies[int(float64(numRequests)*0.99)]

	t.Logf("Latency Distribution Results:")
	t.Logf("  Average: %v", avgLatency)
	t.Logf("  P50: %v", p50)
	t.Logf("  P95: %v", p95)
	t.Logf("  P99: %v", p99)
	t.Logf("  Min: %v", latencies[0])
	t.Logf("  Max: %v", latencies[numRequests-1])

	// 验证延迟要求
	// Verify latency requirements
	assert.Less(t, p99, 100*time.Millisecond, "P99 latency should be < 100ms")
	assert.Less(t, p95, 50*time.Millisecond, "P95 latency should be < 50ms")
	assert.Less(t, avgLatency, 25*time.Millisecond, "Average latency should be < 25ms")
}
