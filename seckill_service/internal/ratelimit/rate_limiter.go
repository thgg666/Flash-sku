package ratelimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
)

// LimitLevel 限流级别
// LimitLevel rate limit level
type LimitLevel string

const (
	LimitLevelGlobal LimitLevel = "global" // 全局限流
	LimitLevelIP     LimitLevel = "ip"     // IP限流
	LimitLevelUser   LimitLevel = "user"   // 用户限流
)

// RateLimiter 多级限流器
// RateLimiter multi-level rate limiter
type RateLimiter struct {
	globalBucket *TokenBucket           // 全局限流桶
	ipBuckets    sync.Map               // IP限流桶 map[string]*TokenBucket
	userBuckets  sync.Map               // 用户限流桶 map[string]*TokenBucket
	configs      map[LimitLevel]*TokenBucketConfig // 配置
	logger       logger.Logger          // 日志器
	cleanupTicker *time.Ticker          // 清理定时器
	stopCleanup   chan struct{}         // 停止清理信号
}

// RateLimiterConfig 限流器配置
// RateLimiterConfig rate limiter configuration
type RateLimiterConfig struct {
	GlobalConfig *TokenBucketConfig `json:"global_config"`
	IPConfig     *TokenBucketConfig `json:"ip_config"`
	UserConfig   *TokenBucketConfig `json:"user_config"`
	CleanupInterval time.Duration   `json:"cleanup_interval"` // 清理间隔
}

// LimitResult 限流结果
// LimitResult rate limit result
type LimitResult struct {
	Allowed     bool       `json:"allowed"`      // 是否允许
	Level       LimitLevel `json:"level"`        // 触发的限流级别
	Reason      string     `json:"reason"`       // 限流原因
	RetryAfter  int64      `json:"retry_after"`  // 建议重试时间(秒)
	RemainingTokens int64  `json:"remaining_tokens"` // 剩余令牌数
}

// NewRateLimiter 创建新的限流器
// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig, log logger.Logger) *RateLimiter {
	if config == nil {
		config = GetDefaultRateLimiterConfig()
	}

	rl := &RateLimiter{
		globalBucket: NewTokenBucket(config.GlobalConfig),
		configs: map[LimitLevel]*TokenBucketConfig{
			LimitLevelGlobal: config.GlobalConfig,
			LimitLevelIP:     config.IPConfig,
			LimitLevelUser:   config.UserConfig,
		},
		logger:      log,
		stopCleanup: make(chan struct{}),
	}

	// 启动清理协程
	// Start cleanup goroutine
	cleanupInterval := config.CleanupInterval
	if cleanupInterval == 0 {
		cleanupInterval = 5 * time.Minute // 默认5分钟清理一次
	}
	
	rl.cleanupTicker = time.NewTicker(cleanupInterval)
	go rl.cleanupRoutine()

	return rl
}

// Allow 检查是否允许请求 (检查所有级别)
// Allow checks if request is allowed (checks all levels)
func (rl *RateLimiter) Allow(ip, userID string) *LimitResult {
	// 1. 检查全局限流
	// Check global rate limit
	if !rl.globalBucket.Allow() {
		return &LimitResult{
			Allowed:         false,
			Level:           LimitLevelGlobal,
			Reason:          "Global rate limit exceeded",
			RetryAfter:      1, // 建议1秒后重试
			RemainingTokens: rl.globalBucket.GetTokens(),
		}
	}

	// 2. 检查IP限流
	// Check IP rate limit
	if ip != "" {
		ipBucket := rl.getOrCreateIPBucket(ip)
		if !ipBucket.Allow() {
			return &LimitResult{
				Allowed:         false,
				Level:           LimitLevelIP,
				Reason:          fmt.Sprintf("IP rate limit exceeded for %s", ip),
				RetryAfter:      10, // IP限流建议10秒后重试
				RemainingTokens: ipBucket.GetTokens(),
			}
		}
	}

	// 3. 检查用户限流
	// Check user rate limit
	if userID != "" {
		userBucket := rl.getOrCreateUserBucket(userID)
		if !userBucket.Allow() {
			return &LimitResult{
				Allowed:         false,
				Level:           LimitLevelUser,
				Reason:          fmt.Sprintf("User rate limit exceeded for %s", userID),
				RetryAfter:      60, // 用户限流建议60秒后重试
				RemainingTokens: userBucket.GetTokens(),
			}
		}
	}

	// 所有检查通过
	// All checks passed
	return &LimitResult{
		Allowed:         true,
		Level:           "",
		Reason:          "",
		RetryAfter:      0,
		RemainingTokens: rl.globalBucket.GetTokens(),
	}
}

// AllowLevel 检查特定级别的限流
// AllowLevel checks rate limit for specific level
func (rl *RateLimiter) AllowLevel(level LimitLevel, key string) *LimitResult {
	switch level {
	case LimitLevelGlobal:
		if rl.globalBucket.Allow() {
			return &LimitResult{
				Allowed:         true,
				Level:           level,
				RemainingTokens: rl.globalBucket.GetTokens(),
			}
		}
		return &LimitResult{
			Allowed:         false,
			Level:           level,
			Reason:          "Global rate limit exceeded",
			RetryAfter:      1,
			RemainingTokens: rl.globalBucket.GetTokens(),
		}

	case LimitLevelIP:
		bucket := rl.getOrCreateIPBucket(key)
		if bucket.Allow() {
			return &LimitResult{
				Allowed:         true,
				Level:           level,
				RemainingTokens: bucket.GetTokens(),
			}
		}
		return &LimitResult{
			Allowed:         false,
			Level:           level,
			Reason:          fmt.Sprintf("IP rate limit exceeded for %s", key),
			RetryAfter:      10,
			RemainingTokens: bucket.GetTokens(),
		}

	case LimitLevelUser:
		bucket := rl.getOrCreateUserBucket(key)
		if bucket.Allow() {
			return &LimitResult{
				Allowed:         true,
				Level:           level,
				RemainingTokens: bucket.GetTokens(),
			}
		}
		return &LimitResult{
			Allowed:         false,
			Level:           level,
			Reason:          fmt.Sprintf("User rate limit exceeded for %s", key),
			RetryAfter:      60,
			RemainingTokens: bucket.GetTokens(),
		}

	default:
		return &LimitResult{
			Allowed: false,
			Level:   level,
			Reason:  "Unknown limit level",
		}
	}
}

// getOrCreateIPBucket 获取或创建IP限流桶
// getOrCreateIPBucket gets or creates IP rate limit bucket
func (rl *RateLimiter) getOrCreateIPBucket(ip string) *TokenBucket {
	if bucket, ok := rl.ipBuckets.Load(ip); ok {
		return bucket.(*TokenBucket)
	}

	// 创建新的IP桶
	// Create new IP bucket
	newBucket := NewTokenBucket(rl.configs[LimitLevelIP])
	actual, loaded := rl.ipBuckets.LoadOrStore(ip, newBucket)
	
	if loaded {
		return actual.(*TokenBucket)
	}
	
	rl.logger.Debug("Created new IP bucket",
		logger.String("ip", ip),
		logger.Int64("capacity", newBucket.GetCapacity()),
		logger.Int64("refill_rate", newBucket.GetRefillRate()))
	
	return newBucket
}

// getOrCreateUserBucket 获取或创建用户限流桶
// getOrCreateUserBucket gets or creates user rate limit bucket
func (rl *RateLimiter) getOrCreateUserBucket(userID string) *TokenBucket {
	if bucket, ok := rl.userBuckets.Load(userID); ok {
		return bucket.(*TokenBucket)
	}

	// 创建新的用户桶
	// Create new user bucket
	newBucket := NewTokenBucket(rl.configs[LimitLevelUser])
	actual, loaded := rl.userBuckets.LoadOrStore(userID, newBucket)
	
	if loaded {
		return actual.(*TokenBucket)
	}
	
	rl.logger.Debug("Created new user bucket",
		logger.String("user_id", userID),
		logger.Int64("capacity", newBucket.GetCapacity()),
		logger.Int64("refill_rate", newBucket.GetRefillRate()))
	
	return newBucket
}

// UpdateConfig 动态更新配置
// UpdateConfig dynamically updates configuration
func (rl *RateLimiter) UpdateConfig(level LimitLevel, config *TokenBucketConfig) error {
	switch level {
	case LimitLevelGlobal:
		rl.globalBucket.UpdateConfig(config)
		rl.configs[level] = config
		rl.logger.Info("Updated global rate limit config",
			logger.Int64("capacity", config.Capacity),
			logger.Int64("refill_rate", config.RefillRate))

	case LimitLevelIP:
		rl.configs[level] = config
		// 更新所有现有的IP桶
		// Update all existing IP buckets
		rl.ipBuckets.Range(func(key, value interface{}) bool {
			bucket := value.(*TokenBucket)
			bucket.UpdateConfig(config)
			return true
		})
		rl.logger.Info("Updated IP rate limit config",
			logger.Int64("capacity", config.Capacity),
			logger.Int64("refill_rate", config.RefillRate))

	case LimitLevelUser:
		rl.configs[level] = config
		// 更新所有现有的用户桶
		// Update all existing user buckets
		rl.userBuckets.Range(func(key, value interface{}) bool {
			bucket := value.(*TokenBucket)
			bucket.UpdateConfig(config)
			return true
		})
		rl.logger.Info("Updated user rate limit config",
			logger.Int64("capacity", config.Capacity),
			logger.Int64("refill_rate", config.RefillRate))

	default:
		return fmt.Errorf("unknown limit level: %s", level)
	}

	return nil
}

// cleanupRoutine 清理过期的桶
// cleanupRoutine cleans up expired buckets
func (rl *RateLimiter) cleanupRoutine() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup 清理长时间未使用的桶
// cleanup removes long unused buckets
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	cleanupThreshold := 10 * time.Minute // 10分钟未使用则清理

	// 清理IP桶
	// Cleanup IP buckets
	ipCount := 0
	rl.ipBuckets.Range(func(key, value interface{}) bool {
		bucket := value.(*TokenBucket)
		stats := bucket.GetStats()
		
		// 如果桶已满且超过阈值时间未使用，则删除
		// Remove if bucket is full and unused for threshold time
		if stats.CurrentTokens == stats.Capacity && 
		   now.Sub(stats.LastRefill) > cleanupThreshold {
			rl.ipBuckets.Delete(key)
			rl.logger.Debug("Cleaned up IP bucket", logger.String("ip", key.(string)))
		} else {
			ipCount++
		}
		return true
	})

	// 清理用户桶
	// Cleanup user buckets
	userCount := 0
	rl.userBuckets.Range(func(key, value interface{}) bool {
		bucket := value.(*TokenBucket)
		stats := bucket.GetStats()
		
		if stats.CurrentTokens == stats.Capacity && 
		   now.Sub(stats.LastRefill) > cleanupThreshold {
			rl.userBuckets.Delete(key)
			rl.logger.Debug("Cleaned up user bucket", logger.String("user_id", key.(string)))
		} else {
			userCount++
		}
		return true
	})

	rl.logger.Debug("Rate limiter cleanup completed",
		logger.Int("active_ip_buckets", ipCount),
		logger.Int("active_user_buckets", userCount))
}

// Close 关闭限流器
// Close shuts down the rate limiter
func (rl *RateLimiter) Close() {
	close(rl.stopCleanup)
}

// GetDefaultRateLimiterConfig 获取默认限流器配置
// GetDefaultRateLimiterConfig returns default rate limiter configuration
func GetDefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		GlobalConfig:    GetDefaultConfig("global"),
		IPConfig:        GetDefaultConfig("ip"),
		UserConfig:      GetDefaultConfig("user"),
		CleanupInterval: 5 * time.Minute,
	}
}
