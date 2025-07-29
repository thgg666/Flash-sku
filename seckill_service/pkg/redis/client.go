package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Client Redis客户端接口
// Client Redis client interface
type Client interface {
	// Get 获取值
	// Get gets value
	Get(ctx context.Context, key string) (string, error)

	// Set 设置值
	// Set sets value
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Del 删除键
	// Del deletes key
	Del(ctx context.Context, keys ...string) error

	// Exists 检查键是否存在
	// Exists checks if key exists
	Exists(ctx context.Context, keys ...string) (int64, error)

	// TTL 获取键的过期时间
	// TTL gets key expiration time
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Incr 递增
	// Incr increments
	Incr(ctx context.Context, key string) (int64, error)

	// Decr 递减
	// Decr decrements
	Decr(ctx context.Context, key string) (int64, error)

	// IncrBy 按指定值递增
	// IncrBy increments by specified value
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// Expire 设置键过期时间
	// Expire sets key expiration time
	Expire(ctx context.Context, key string, expiration time.Duration) error

	// Eval 执行Lua脚本
	// Eval executes Lua script
	Eval(ctx context.Context, script string, keys []string, args ...any) (any, error)

	// ScriptLoad 加载Lua脚本
	// ScriptLoad loads Lua script
	ScriptLoad(ctx context.Context, script string) (string, error)

	// EvalSHA 执行已加载的Lua脚本
	// EvalSHA executes loaded Lua script
	EvalSHA(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error)

	// ScriptExists 检查脚本是否存在
	// ScriptExists checks if script exists
	ScriptExists(ctx context.Context, sha1 string) (bool, error)

	// Ping 检查连接
	// Ping checks connection
	Ping(ctx context.Context) error

	// Close 关闭连接
	// Close closes connection
	Close() error
}

// Config Redis配置
// Config Redis configuration
type Config struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

// RedisClient Redis客户端实现
// RedisClient Redis client implementation
type RedisClient struct {
	client *redis.Client
	config *Config
}

// NewClient 创建新的Redis客户端
// NewClient creates new Redis client
func NewClient(config *Config) (Client, error) {
	// 创建Redis客户端选项
	// Create Redis client options
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	// 创建Redis客户端
	// Create Redis client
	rdb := redis.NewClient(opts)

	// 测试连接
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	client := &RedisClient{
		client: rdb,
		config: config,
	}

	return client, nil
}

// Get 获取值
// Get gets value
func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set 设置值
// Set sets value
func (c *RedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Del 删除键
// Del deletes key
func (c *RedisClient) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
// Exists checks if key exists
func (c *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// TTL 获取键的过期时间
// TTL gets key expiration time
func (c *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Incr 递增
// Incr increments
func (c *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Decr 递减
// Decr decrements
func (c *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// Eval 执行Lua脚本
// Eval executes Lua script
func (c *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

// ScriptLoad 加载Lua脚本
// ScriptLoad loads Lua script
func (c *RedisClient) ScriptLoad(ctx context.Context, script string) (string, error) {
	return c.client.ScriptLoad(ctx, script).Result()
}

// EvalSHA 执行已加载的Lua脚本
// EvalSHA executes loaded Lua script
func (c *RedisClient) EvalSHA(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptExists 检查脚本是否存在
// ScriptExists checks if script exists
func (c *RedisClient) ScriptExists(ctx context.Context, sha1 string) (bool, error) {
	results, err := c.client.ScriptExists(ctx, sha1).Result()
	if err != nil {
		return false, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return false, nil
}

// IncrBy 按指定值递增
// IncrBy increments by specified value
func (c *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Expire 设置键过期时间
// Expire sets key expiration time
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// Ping 检查连接
// Ping checks connection
func (c *RedisClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close 关闭连接
// Close closes connection
func (c *RedisClient) Close() error {
	return c.client.Close()
}
