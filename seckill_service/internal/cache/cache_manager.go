package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// CacheManager 缓存管理器
// CacheManager cache manager
type CacheManager struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *CacheConfig
}

// CacheConfig 缓存配置
// CacheConfig cache configuration
type CacheConfig struct {
	DefaultTTL      time.Duration `json:"default_ttl"`
	ActivityTTL     time.Duration `json:"activity_ttl"`
	StockTTL        time.Duration `json:"stock_ttl"`
	UserLimitTTL    time.Duration `json:"user_limit_ttl"`
	RateLimitTTL    time.Duration `json:"rate_limit_ttl"`
	RefreshInterval time.Duration `json:"refresh_interval"`
}

// CacheKeys 缓存键常量
// CacheKeys cache key constants
type CacheKeys struct {
	ActivityPrefix   string
	StockPrefix      string
	UserLimitPrefix  string
	RateLimitPrefix  string
	MetricsPrefix    string
	HealthPrefix     string
}

// DefaultCacheKeys 默认缓存键
// DefaultCacheKeys default cache keys
var DefaultCacheKeys = &CacheKeys{
	ActivityPrefix:  "seckill:activity:",
	StockPrefix:     "seckill:stock:",
	UserLimitPrefix: "seckill:user_limit:",
	RateLimitPrefix: "seckill:rate_limit:",
	MetricsPrefix:   "seckill:metrics:",
	HealthPrefix:    "seckill:health:",
}

// CacheMetrics 缓存指标
// CacheMetrics cache metrics
type CacheMetrics struct {
	HitCount    int64   `json:"hit_count"`
	MissCount   int64   `json:"miss_count"`
	HitRate     float64 `json:"hit_rate"`
	SetCount    int64   `json:"set_count"`
	DeleteCount int64   `json:"delete_count"`
	ErrorCount  int64   `json:"error_count"`
	LastUpdated time.Time `json:"last_updated"`
}

// NewCacheManager 创建缓存管理器
// NewCacheManager creates cache manager
func NewCacheManager(redisClient redis.Client, config *CacheConfig, log logger.Logger) *CacheManager {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &CacheManager{
		redisClient: redisClient,
		logger:      log,
		config:      config,
	}
}

// SetActivity 设置活动缓存
// SetActivity sets activity cache
func (c *CacheManager) SetActivity(ctx context.Context, activityID string, data interface{}) error {
	key := DefaultCacheKeys.ActivityPrefix + activityID
	return c.setWithTTL(ctx, key, data, c.config.ActivityTTL)
}

// GetActivity 获取活动缓存
// GetActivity gets activity cache
func (c *CacheManager) GetActivity(ctx context.Context, activityID string, dest interface{}) error {
	key := DefaultCacheKeys.ActivityPrefix + activityID
	return c.get(ctx, key, dest)
}

// SetStock 设置库存缓存
// SetStock sets stock cache
func (c *CacheManager) SetStock(ctx context.Context, activityID string, stock int) error {
	key := DefaultCacheKeys.StockPrefix + activityID
	return c.setWithTTL(ctx, key, stock, c.config.StockTTL)
}

// GetStock 获取库存缓存
// GetStock gets stock cache
func (c *CacheManager) GetStock(ctx context.Context, activityID string) (int, error) {
	key := DefaultCacheKeys.StockPrefix + activityID
	var stock int
	err := c.get(ctx, key, &stock)
	return stock, err
}

// DecrStock 原子递减库存
// DecrStock atomically decrements stock
func (c *CacheManager) DecrStock(ctx context.Context, activityID string) (int64, error) {
	key := DefaultCacheKeys.StockPrefix + activityID
	return c.redisClient.Decr(ctx, key)
}

// SetUserLimit 设置用户限购
// SetUserLimit sets user purchase limit
func (c *CacheManager) SetUserLimit(ctx context.Context, userID, activityID string, count int) error {
	key := fmt.Sprintf("%s%s:%s", DefaultCacheKeys.UserLimitPrefix, userID, activityID)
	return c.setWithTTL(ctx, key, count, c.config.UserLimitTTL)
}

// GetUserLimit 获取用户限购
// GetUserLimit gets user purchase limit
func (c *CacheManager) GetUserLimit(ctx context.Context, userID, activityID string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", DefaultCacheKeys.UserLimitPrefix, userID, activityID)
	var count int
	err := c.get(ctx, key, &count)
	return count, err
}

// IncrUserLimit 增加用户购买计数
// IncrUserLimit increments user purchase count
func (c *CacheManager) IncrUserLimit(ctx context.Context, userID, activityID string) (int64, error) {
	key := fmt.Sprintf("%s%s:%s", DefaultCacheKeys.UserLimitPrefix, userID, activityID)
	
	// 先设置过期时间，再递增
	// Set expiration first, then increment
	exists, err := c.redisClient.Exists(ctx, key)
	if err != nil {
		return 0, err
	}
	
	if exists == 0 {
		// 键不存在，先设置初始值和过期时间
		// Key doesn't exist, set initial value and expiration
		if err := c.redisClient.Set(ctx, key, 0, c.config.UserLimitTTL); err != nil {
			return 0, err
		}
	}
	
	return c.redisClient.Incr(ctx, key)
}

// SetRateLimit 设置限流
// SetRateLimit sets rate limit
func (c *CacheManager) SetRateLimit(ctx context.Context, key string, count int) error {
	rateLimitKey := DefaultCacheKeys.RateLimitPrefix + key
	return c.setWithTTL(ctx, rateLimitKey, count, c.config.RateLimitTTL)
}

// GetRateLimit 获取限流
// GetRateLimit gets rate limit
func (c *CacheManager) GetRateLimit(ctx context.Context, key string) (int, error) {
	rateLimitKey := DefaultCacheKeys.RateLimitPrefix + key
	var count int
	err := c.get(ctx, rateLimitKey, &count)
	return count, err
}

// InvalidateActivity 失效活动缓存
// InvalidateActivity invalidates activity cache
func (c *CacheManager) InvalidateActivity(ctx context.Context, activityID string) error {
	keys := []string{
		DefaultCacheKeys.ActivityPrefix + activityID,
		DefaultCacheKeys.StockPrefix + activityID,
	}
	
	return c.redisClient.Del(ctx, keys...)
}

// RefreshActivity 刷新活动缓存
// RefreshActivity refreshes activity cache
func (c *CacheManager) RefreshActivity(ctx context.Context, activityID string, data interface{}) error {
	// 先删除旧缓存
	// Delete old cache first
	if err := c.InvalidateActivity(ctx, activityID); err != nil {
		c.logger.Warn("Failed to invalidate activity cache",
			logger.String("activity_id", activityID),
			logger.Error(err))
	}
	
	// 设置新缓存
	// Set new cache
	return c.SetActivity(ctx, activityID, data)
}

// GetMetrics 获取缓存指标
// GetMetrics gets cache metrics
func (c *CacheManager) GetMetrics(ctx context.Context) (*CacheMetrics, error) {
	key := DefaultCacheKeys.MetricsPrefix + "stats"
	
	var metrics CacheMetrics
	err := c.get(ctx, key, &metrics)
	if err != nil {
		// 如果没有指标数据，返回空指标
		// If no metrics data, return empty metrics
		return &CacheMetrics{
			LastUpdated: time.Now(),
		}, nil
	}
	
	return &metrics, nil
}

// UpdateMetrics 更新缓存指标
// UpdateMetrics updates cache metrics
func (c *CacheManager) UpdateMetrics(ctx context.Context, metrics *CacheMetrics) error {
	key := DefaultCacheKeys.MetricsPrefix + "stats"
	metrics.LastUpdated = time.Now()
	
	// 计算命中率
	// Calculate hit rate
	total := metrics.HitCount + metrics.MissCount
	if total > 0 {
		metrics.HitRate = float64(metrics.HitCount) / float64(total)
	}
	
	return c.setWithTTL(ctx, key, metrics, 24*time.Hour)
}

// setWithTTL 设置带过期时间的缓存
// setWithTTL sets cache with TTL
func (c *CacheManager) setWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	return c.redisClient.Set(ctx, key, string(data), ttl)
}

// get 获取缓存
// get gets cache
func (c *CacheManager) get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.redisClient.Get(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(data), dest)
}

// DefaultCacheConfig 默认缓存配置
// DefaultCacheConfig default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		DefaultTTL:      1 * time.Hour,
		ActivityTTL:     24 * time.Hour,
		StockTTL:        1 * time.Hour,
		UserLimitTTL:    24 * time.Hour,
		RateLimitTTL:    1 * time.Minute,
		RefreshInterval: 5 * time.Minute,
	}
}
