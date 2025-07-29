package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// RedisRateLimiter 基于Redis的分布式限流器
// RedisRateLimiter Redis-based distributed rate limiter
type RedisRateLimiter struct {
	client  redis.Client
	config  *RateLimiterConfig
	logger  logger.Logger
	scripts map[string]string // Lua脚本缓存
}

// NewRedisRateLimiter 创建Redis限流器
// NewRedisRateLimiter creates Redis rate limiter
func NewRedisRateLimiter(client redis.Client, config *RateLimiterConfig, log logger.Logger) *RedisRateLimiter {
	if config == nil {
		config = GetDefaultRateLimiterConfig()
	}

	rl := &RedisRateLimiter{
		client:  client,
		config:  config,
		logger:  log,
		scripts: make(map[string]string),
	}

	// 预加载Lua脚本
	// Preload Lua scripts
	rl.loadScripts()

	return rl
}

// Allow 检查是否允许请求
// Allow checks if request is allowed
func (rl *RedisRateLimiter) Allow(ctx context.Context, ip, userID string) *LimitResult {
	// 1. 检查全局限流
	// Check global rate limit
	globalResult := rl.allowLevel(ctx, LimitLevelGlobal, "global", rl.config.GlobalConfig)
	if !globalResult.Allowed {
		return globalResult
	}

	// 2. 检查IP限流
	// Check IP rate limit
	if ip != "" {
		ipResult := rl.allowLevel(ctx, LimitLevelIP, ip, rl.config.IPConfig)
		if !ipResult.Allowed {
			return ipResult
		}
	}

	// 3. 检查用户限流
	// Check user rate limit
	if userID != "" {
		userResult := rl.allowLevel(ctx, LimitLevelUser, userID, rl.config.UserConfig)
		if !userResult.Allowed {
			return userResult
		}
	}

	return &LimitResult{
		Allowed:         true,
		RemainingTokens: globalResult.RemainingTokens,
	}
}

// allowLevel 检查特定级别的限流
// allowLevel checks rate limit for specific level
func (rl *RedisRateLimiter) allowLevel(ctx context.Context, level LimitLevel, key string, config *TokenBucketConfig) *LimitResult {
	redisKey := rl.buildRedisKey(level, key)

	// 使用Lua脚本原子性地检查和更新令牌桶
	// Use Lua script to atomically check and update token bucket
	result, err := rl.client.Eval(ctx, rl.scripts["token_bucket"], []string{redisKey},
		config.Capacity, config.RefillRate, time.Now().Unix(), 1)

	if err != nil {
		rl.logger.Error("Redis rate limit check failed",
			logger.String("level", string(level)),
			logger.String("key", key),
			logger.Error(err))

		// 发生错误时允许请求通过 (fail-open策略)
		// Allow request on error (fail-open strategy)
		return &LimitResult{
			Allowed: true,
			Level:   level,
			Reason:  "Rate limiter error, allowing request",
		}
	}

	// 解析Lua脚本返回结果
	// Parse Lua script result
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) < 3 {
		rl.logger.Error("Invalid rate limit result format",
			logger.String("level", string(level)),
			logger.String("key", key))
		return &LimitResult{Allowed: true}
	}

	allowed, _ := resultArray[0].(int64)
	remainingTokens, _ := resultArray[1].(int64)
	retryAfter, _ := resultArray[2].(int64)

	if allowed == 1 {
		return &LimitResult{
			Allowed:         true,
			Level:           level,
			RemainingTokens: remainingTokens,
		}
	}

	return &LimitResult{
		Allowed:         false,
		Level:           level,
		Reason:          fmt.Sprintf("%s rate limit exceeded for %s", level, key),
		RetryAfter:      retryAfter,
		RemainingTokens: remainingTokens,
	}
}

// GetStats 获取限流统计信息
// GetStats returns rate limit statistics
func (rl *RedisRateLimiter) GetStats(ctx context.Context, level LimitLevel, key string) (*TokenBucketStats, error) {
	redisKey := rl.buildRedisKey(level, key)

	// 获取令牌桶数据
	// Get token bucket data
	data, err := rl.client.Get(ctx, redisKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit stats: %w", err)
	}

	if data == "" {
		// 桶不存在，返回默认状态
		// Bucket doesn't exist, return default state
		config := rl.getConfigForLevel(level)
		return &TokenBucketStats{
			Capacity:      config.Capacity,
			CurrentTokens: config.Capacity,
			RefillRate:    config.RefillRate,
			LastRefill:    time.Now(),
			Utilization:   0.0,
		}, nil
	}

	var bucketData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &bucketData); err != nil {
		return nil, fmt.Errorf("failed to parse bucket data: %w", err)
	}

	tokens, _ := bucketData["tokens"].(float64)
	lastRefill, _ := bucketData["last_refill"].(float64)
	config := rl.getConfigForLevel(level)

	utilization := 1.0 - tokens/float64(config.Capacity)
	if utilization < 0 {
		utilization = 0
	}

	return &TokenBucketStats{
		Capacity:      config.Capacity,
		CurrentTokens: int64(tokens),
		RefillRate:    config.RefillRate,
		LastRefill:    time.Unix(int64(lastRefill), 0),
		Utilization:   utilization,
	}, nil
}

// Reset 重置限流状态
// Reset resets rate limit state
func (rl *RedisRateLimiter) Reset(ctx context.Context, level LimitLevel, key string) error {
	redisKey := rl.buildRedisKey(level, key)
	return rl.client.Del(ctx, redisKey)
}

// UpdateConfig 动态更新配置
// UpdateConfig dynamically updates configuration
func (rl *RedisRateLimiter) UpdateConfig(level LimitLevel, config *TokenBucketConfig) error {
	switch level {
	case LimitLevelGlobal:
		rl.config.GlobalConfig = config
	case LimitLevelIP:
		rl.config.IPConfig = config
	case LimitLevelUser:
		rl.config.UserConfig = config
	default:
		return fmt.Errorf("unknown limit level: %s", level)
	}

	rl.logger.Info("Updated Redis rate limit config",
		logger.String("level", string(level)),
		logger.Int64("capacity", config.Capacity),
		logger.Int64("refill_rate", config.RefillRate))

	return nil
}

// buildRedisKey 构建Redis键
// buildRedisKey builds Redis key
func (rl *RedisRateLimiter) buildRedisKey(level LimitLevel, key string) string {
	return fmt.Sprintf("rate_limit:%s:%s", level, key)
}

// getConfigForLevel 获取指定级别的配置
// getConfigForLevel returns configuration for specified level
func (rl *RedisRateLimiter) getConfigForLevel(level LimitLevel) *TokenBucketConfig {
	switch level {
	case LimitLevelGlobal:
		return rl.config.GlobalConfig
	case LimitLevelIP:
		return rl.config.IPConfig
	case LimitLevelUser:
		return rl.config.UserConfig
	default:
		return GetDefaultConfig("global")
	}
}

// loadScripts 加载Lua脚本
// loadScripts loads Lua scripts
func (rl *RedisRateLimiter) loadScripts() {
	// 令牌桶算法Lua脚本
	// Token bucket algorithm Lua script
	rl.scripts["token_bucket"] = `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local requested = tonumber(ARGV[4])
		
		-- 获取当前桶状态
		-- Get current bucket state
		local bucket_data = redis.call('GET', key)
		local tokens = capacity
		local last_refill = now
		
		if bucket_data then
			local data = cjson.decode(bucket_data)
			tokens = data.tokens or capacity
			last_refill = data.last_refill or now
		end
		
		-- 计算需要补充的令牌
		-- Calculate tokens to refill
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed * refill_rate)
		
		if tokens_to_add > 0 then
			tokens = math.min(capacity, tokens + tokens_to_add)
			last_refill = now
		end
		
		-- 检查是否有足够的令牌
		-- Check if there are enough tokens
		local allowed = 0
		local retry_after = 0
		
		if tokens >= requested then
			tokens = tokens - requested
			allowed = 1
		else
			-- 计算建议重试时间
			-- Calculate suggested retry time
			retry_after = math.ceil((requested - tokens) / refill_rate)
		end
		
		-- 保存桶状态
		-- Save bucket state
		local new_data = cjson.encode({
			tokens = tokens,
			last_refill = last_refill
		})
		
		redis.call('SET', key, new_data, 'EX', 3600) -- 1小时过期
		
		return {allowed, tokens, retry_after}
	`

	rl.logger.Info("Loaded Redis rate limiter Lua scripts")
}

// Close 关闭Redis限流器
// Close shuts down Redis rate limiter
func (rl *RedisRateLimiter) Close() error {
	// Redis客户端由外部管理，这里不需要关闭
	// Redis client is managed externally, no need to close here
	rl.logger.Info("Redis rate limiter closed")
	return nil
}
