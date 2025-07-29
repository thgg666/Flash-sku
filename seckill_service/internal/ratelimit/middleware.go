package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/flashsku/seckill/pkg/logger"
)

// RateLimitMiddleware 限流中间件接口
// RateLimitMiddleware rate limit middleware interface
type RateLimitMiddleware interface {
	Handler() gin.HandlerFunc
	GetStats(c *gin.Context)
	UpdateConfig(c *gin.Context)
}

// GinRateLimitMiddleware Gin限流中间件
// GinRateLimitMiddleware Gin rate limit middleware
type GinRateLimitMiddleware struct {
	limiter RateLimiterInterface
	logger  logger.Logger
	config  *MiddlewareConfig
}

// RateLimiterInterface 限流器接口
// RateLimiterInterface rate limiter interface
type RateLimiterInterface interface {
	Allow(ip, userID string) *LimitResult
}

// MiddlewareConfig 中间件配置
// MiddlewareConfig middleware configuration
type MiddlewareConfig struct {
	SkipPaths      []string `json:"skip_paths"`       // 跳过限流的路径
	EnableUserID   bool     `json:"enable_user_id"`   // 是否启用用户ID限流
	UserIDHeader   string   `json:"user_id_header"`   // 用户ID请求头
	EnableLogging  bool     `json:"enable_logging"`   // 是否启用日志
	CustomResponse bool     `json:"custom_response"`  // 是否使用自定义响应格式
}

// RateLimitResponse 限流响应
// RateLimitResponse rate limit response
type RateLimitResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code"`
	RetryAfter int64  `json:"retry_after"`
	Timestamp  int64  `json:"timestamp"`
}

// NewGinRateLimitMiddleware 创建Gin限流中间件
// NewGinRateLimitMiddleware creates Gin rate limit middleware
func NewGinRateLimitMiddleware(limiter RateLimiterInterface, config *MiddlewareConfig, log logger.Logger) *GinRateLimitMiddleware {
	if config == nil {
		config = GetDefaultMiddlewareConfig()
	}

	return &GinRateLimitMiddleware{
		limiter: limiter,
		logger:  log,
		config:  config,
	}
}

// Handler 返回Gin中间件处理函数
// Handler returns Gin middleware handler function
func (m *GinRateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过限流
		// Check if should skip rate limiting
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 获取客户端IP
		// Get client IP
		clientIP := m.getClientIP(c)

		// 获取用户ID (如果启用)
		// Get user ID (if enabled)
		var userID string
		if m.config.EnableUserID {
			userID = m.getUserID(c)
		}

		// 执行限流检查
		// Perform rate limit check
		result := m.limiter.Allow(clientIP, userID)

		// 记录限流日志
		// Log rate limit result
		if m.config.EnableLogging {
			m.logRateLimit(c, clientIP, userID, result)
		}

		// 设置响应头
		// Set response headers
		m.setRateLimitHeaders(c, result)

		if !result.Allowed {
			// 限流触发，返回错误响应
			// Rate limit triggered, return error response
			m.handleRateLimitExceeded(c, result)
			return
		}

		// 继续处理请求
		// Continue processing request
		c.Next()
	}
}

// shouldSkip 检查是否应该跳过限流
// shouldSkip checks if rate limiting should be skipped
func (m *GinRateLimitMiddleware) shouldSkip(c *gin.Context) bool {
	path := c.Request.URL.Path
	
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
	}
	
	return false
}

// getClientIP 获取客户端IP
// getClientIP gets client IP address
func (m *GinRateLimitMiddleware) getClientIP(c *gin.Context) string {
	// 优先从X-Forwarded-For获取
	// Prefer X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// 其次从X-Real-IP获取
	// Then try X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// 最后使用RemoteAddr
	// Finally use RemoteAddr
	return c.ClientIP()
}

// getUserID 获取用户ID
// getUserID gets user ID
func (m *GinRateLimitMiddleware) getUserID(c *gin.Context) string {
	// 从请求头获取
	// Get from header
	if m.config.UserIDHeader != "" {
		if userID := c.GetHeader(m.config.UserIDHeader); userID != "" {
			return userID
		}
	}
	
	// 从JWT token获取 (如果存在)
	// Get from JWT token (if exists)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	
	return ""
}

// logRateLimit 记录限流日志
// logRateLimit logs rate limit result
func (m *GinRateLimitMiddleware) logRateLimit(c *gin.Context, clientIP, userID string, result *LimitResult) {
	fields := []logger.Field{
		logger.String("method", c.Request.Method),
		logger.String("path", c.Request.URL.Path),
		logger.String("client_ip", clientIP),
		logger.Bool("allowed", result.Allowed),
		logger.Int64("remaining_tokens", result.RemainingTokens),
	}
	
	if userID != "" {
		fields = append(fields, logger.String("user_id", userID))
	}
	
	if !result.Allowed {
		fields = append(fields,
			logger.String("limit_level", string(result.Level)),
			logger.String("reason", result.Reason),
			logger.Int64("retry_after", result.RetryAfter))
		
		m.logger.Warn("Rate limit exceeded", fields...)
	} else {
		m.logger.Debug("Rate limit check passed", fields...)
	}
}

// setRateLimitHeaders 设置限流相关的响应头
// setRateLimitHeaders sets rate limit related response headers
func (m *GinRateLimitMiddleware) setRateLimitHeaders(c *gin.Context, result *LimitResult) {
	// 设置剩余令牌数
	// Set remaining tokens
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(result.RemainingTokens, 10))
	
	if !result.Allowed {
		// 设置重试时间
		// Set retry after
		c.Header("Retry-After", strconv.FormatInt(result.RetryAfter, 10))
		c.Header("X-RateLimit-Limit-Level", string(result.Level))
	}
}

// handleRateLimitExceeded 处理限流超出的情况
// handleRateLimitExceeded handles rate limit exceeded case
func (m *GinRateLimitMiddleware) handleRateLimitExceeded(c *gin.Context, result *LimitResult) {
	if m.config.CustomResponse {
		// 使用自定义响应格式
		// Use custom response format
		response := RateLimitResponse{
			Success:    false,
			Message:    result.Reason,
			ErrorCode:  "RATE_LIMIT_EXCEEDED",
			RetryAfter: result.RetryAfter,
			Timestamp:  time.Now().Unix(),
		}
		
		c.JSON(http.StatusTooManyRequests, response)
	} else {
		// 使用标准响应格式
		// Use standard response format
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Rate limit exceeded",
			"message":     result.Reason,
			"retry_after": result.RetryAfter,
		})
	}
	
	c.Abort()
}

// GetStats 获取限流统计信息的处理函数
// GetStats handler for getting rate limit statistics
func (m *GinRateLimitMiddleware) GetStats(c *gin.Context) {
	// 这里可以实现获取统计信息的逻辑
	// Implementation for getting statistics can be added here
	c.JSON(http.StatusOK, gin.H{
		"message": "Rate limit stats endpoint",
		"status":  "not implemented",
	})
}

// UpdateConfig 更新配置的处理函数
// UpdateConfig handler for updating configuration
func (m *GinRateLimitMiddleware) UpdateConfig(c *gin.Context) {
	// 这里可以实现动态更新配置的逻辑
	// Implementation for dynamic configuration update can be added here
	c.JSON(http.StatusOK, gin.H{
		"message": "Rate limit config update endpoint",
		"status":  "not implemented",
	})
}

// GetDefaultMiddlewareConfig 获取默认中间件配置
// GetDefaultMiddlewareConfig returns default middleware configuration
func GetDefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/ping",
		},
		EnableUserID:   true,
		UserIDHeader:   "X-User-ID",
		EnableLogging:  true,
		CustomResponse: true,
	}
}

// RateLimitConfig 限流配置结构
// RateLimitConfig rate limit configuration structure
type RateLimitConfig struct {
	Global *TokenBucketConfig `json:"global"`
	IP     *TokenBucketConfig `json:"ip"`
	User   *TokenBucketConfig `json:"user"`
}

// GetRateLimitConfig 获取当前限流配置
// GetRateLimitConfig returns current rate limit configuration
func GetRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Global: GetDefaultConfig("global"),
		IP:     GetDefaultConfig("ip"),
		User:   GetDefaultConfig("user"),
	}
}
