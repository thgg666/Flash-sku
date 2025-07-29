package performance

import (
	"context"
	"fmt"
	"time"

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

// MockRedisClient 模拟Redis客户端
// MockRedisClient mock Redis client
type MockRedisClient struct{}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	// 模拟库存数据
	// Mock stock data
	if key == "seckill:stock:activity123" {
		return "150", nil
	}
	if key == "seckill:activity:activity123" {
		// 创建一个未来结束时间的活动
		// Create an activity with future end time
		futureTime := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
		pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
		return fmt.Sprintf(`{
			"id":"activity123",
			"name":"Test Activity",
			"status":"active",
			"start_time":"%s",
			"end_time":"%s",
			"total_stock":1000,
			"sold_count":850,
			"price":99.99,
			"user_limit":5
		}`, pastTime, futureTime), nil
	}
	return "", fmt.Errorf("key not found")
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	return nil
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return 1, nil
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockRedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return 1, nil
}

func (m *MockRedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return value, nil
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return time.Hour, nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	// 模拟Lua脚本执行结果
	// Mock Lua script execution result
	return []interface{}{int64(1), "success", map[string]interface{}{"new_stock": int64(149)}}, nil
}

func (m *MockRedisClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	return "mock_sha1", nil
}

func (m *MockRedisClient) EvalSHA(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return []interface{}{int64(1), "success"}, nil
}

func (m *MockRedisClient) ScriptExists(ctx context.Context, sha1 string) (bool, error) {
	return true, nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	return nil
}

func (m *MockRedisClient) Close() error {
	return nil
}
