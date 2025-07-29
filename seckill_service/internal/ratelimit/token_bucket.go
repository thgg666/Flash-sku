package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器
// TokenBucket token bucket rate limiter
type TokenBucket struct {
	capacity    int64     // 桶容量 (最大令牌数)
	tokens      int64     // 当前令牌数
	refillRate  int64     // 每秒补充速率 (tokens/second)
	lastRefill  time.Time // 上次补充时间
	mutex       sync.Mutex // 并发安全锁
}

// TokenBucketConfig 令牌桶配置
// TokenBucketConfig token bucket configuration
type TokenBucketConfig struct {
	Capacity   int64 `json:"capacity"`    // 桶容量
	RefillRate int64 `json:"refill_rate"` // 每秒补充速率
}

// NewTokenBucket 创建新的令牌桶
// NewTokenBucket creates a new token bucket
func NewTokenBucket(config *TokenBucketConfig) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		capacity:   config.Capacity,
		tokens:     config.Capacity, // 初始时桶是满的
		refillRate: config.RefillRate,
		lastRefill: now,
	}
}

// Allow 尝试获取一个令牌
// Allow attempts to consume one token
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 尝试获取N个令牌
// AllowN attempts to consume N tokens
func (tb *TokenBucket) AllowN(n int64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// 补充令牌
	// Refill tokens
	tb.refill()

	// 检查是否有足够的令牌
	// Check if there are enough tokens
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}

	return false
}

// refill 补充令牌 (内部方法，调用前需要加锁)
// refill replenishes tokens (internal method, requires lock)
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// 计算应该补充的令牌数
	// Calculate tokens to add
	tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate
	
	if tokensToAdd > 0 {
		// 更新令牌数，不超过容量
		// Update tokens, not exceeding capacity
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}
}

// GetTokens 获取当前令牌数 (用于监控)
// GetTokens returns current token count (for monitoring)
func (tb *TokenBucket) GetTokens() int64 {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.refill()
	return tb.tokens
}

// GetCapacity 获取桶容量
// GetCapacity returns bucket capacity
func (tb *TokenBucket) GetCapacity() int64 {
	return tb.capacity
}

// GetRefillRate 获取补充速率
// GetRefillRate returns refill rate
func (tb *TokenBucket) GetRefillRate() int64 {
	return tb.refillRate
}

// UpdateConfig 动态更新配置
// UpdateConfig dynamically updates configuration
func (tb *TokenBucket) UpdateConfig(config *TokenBucketConfig) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	// 补充令牌到当前时间
	// Refill tokens to current time
	tb.refill()
	
	// 更新配置
	// Update configuration
	tb.capacity = config.Capacity
	tb.refillRate = config.RefillRate
	
	// 调整当前令牌数不超过新容量
	// Adjust current tokens not to exceed new capacity
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
}

// Reset 重置令牌桶 (清空所有令牌)
// Reset resets the token bucket (empties all tokens)
func (tb *TokenBucket) Reset() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.tokens = 0
	tb.lastRefill = time.Now()
}

// Fill 填满令牌桶
// Fill fills the token bucket to capacity
func (tb *TokenBucket) Fill() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.tokens = tb.capacity
	tb.lastRefill = time.Now()
}

// TokenBucketStats 令牌桶统计信息
// TokenBucketStats token bucket statistics
type TokenBucketStats struct {
	Capacity     int64     `json:"capacity"`
	CurrentTokens int64    `json:"current_tokens"`
	RefillRate   int64     `json:"refill_rate"`
	LastRefill   time.Time `json:"last_refill"`
	Utilization  float64   `json:"utilization"` // 利用率 (1 - tokens/capacity)
}

// GetStats 获取统计信息
// GetStats returns statistics
func (tb *TokenBucket) GetStats() *TokenBucketStats {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	tb.refill()
	
	utilization := 1.0 - float64(tb.tokens)/float64(tb.capacity)
	if utilization < 0 {
		utilization = 0
	}
	
	return &TokenBucketStats{
		Capacity:      tb.capacity,
		CurrentTokens: tb.tokens,
		RefillRate:    tb.refillRate,
		LastRefill:    tb.lastRefill,
		Utilization:   utilization,
	}
}

// min 返回两个int64中的较小值
// min returns the smaller of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// DefaultTokenBucketConfigs 默认令牌桶配置
// DefaultTokenBucketConfigs default token bucket configurations
var DefaultTokenBucketConfigs = map[string]*TokenBucketConfig{
	"global": {
		Capacity:   1000, // 全局限流: 1000 QPS
		RefillRate: 1000,
	},
	"ip": {
		Capacity:   10, // IP限流: 10 QPS
		RefillRate: 10,
	},
	"user": {
		Capacity:   1, // 用户限流: 1 QPS
		RefillRate: 1,
	},
}

// GetDefaultConfig 获取默认配置
// GetDefaultConfig returns default configuration
func GetDefaultConfig(configType string) *TokenBucketConfig {
	if config, exists := DefaultTokenBucketConfigs[configType]; exists {
		// 返回配置的副本，避免修改原始配置
		// Return a copy to avoid modifying original config
		return &TokenBucketConfig{
			Capacity:   config.Capacity,
			RefillRate: config.RefillRate,
		}
	}
	
	// 默认配置
	// Default configuration
	return &TokenBucketConfig{
		Capacity:   100,
		RefillRate: 100,
	}
}
