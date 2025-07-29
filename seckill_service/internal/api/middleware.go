package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
// RateLimitConfig rate limit configuration
type RateLimitConfig struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// RequestIDMiddleware 请求ID中间件
// RequestIDMiddleware request ID middleware
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// LoggerMiddleware 日志中间件
// LoggerMiddleware logger middleware
func LoggerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		// Process request
		c.Next()

		// 记录日志
		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Info("API request",
			logger.String("method", method),
			logger.String("path", path),
			logger.String("client_ip", clientIP),
			logger.Int("status_code", statusCode),
			logger.Duration("latency", latency),
			logger.String("request_id", c.GetString("request_id")))
	}
}

// CORSMiddleware CORS中间件
// CORSMiddleware CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityMiddleware 安全中间件
// SecurityMiddleware security middleware
func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		// Set security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件
// RateLimitMiddleware rate limit middleware
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	// 这里应该实现真正的限流逻辑
	// Should implement real rate limiting logic here
	return func(c *gin.Context) {
		// 简化的限流检查
		// Simplified rate limit check
		clientIP := c.ClientIP()

		// 这里应该使用Redis或内存存储来跟踪请求频率
		// Should use Redis or memory store to track request frequency
		_ = clientIP

		// 暂时不做限流
		// No rate limiting for now
		c.Next()
	}
}

// AuthMiddleware 认证中间件
// AuthMiddleware authentication middleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查Authorization头
		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"message":    "Authorization header required",
				"error_code": "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// 简单的Bearer token检查
		// Simple Bearer token check
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"message":    "Invalid authorization format",
				"error_code": "INVALID_AUTH",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":    false,
				"message":    "Token required",
				"error_code": "TOKEN_REQUIRED",
			})
			c.Abort()
			return
		}

		// 这里应该验证token的有效性
		// Should validate token validity here
		// 暂时跳过验证
		// Skip validation for now

		c.Set("token", token)
		c.Next()
	}
}

// ValidationMiddleware 参数验证中间件
// ValidationMiddleware parameter validation middleware
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证活动ID格式
		// Validate activity ID format
		if activityID := c.Param("activity_id"); activityID != "" {
			if !isValidActivityID(activityID) {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":    false,
					"message":    "Invalid activity ID format",
					"error_code": "INVALID_ACTIVITY_ID",
				})
				c.Abort()
				return
			}
		}

		// 验证用户ID格式
		// Validate user ID format
		if userID := c.Param("user_id"); userID != "" {
			if !isValidUserID(userID) {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":    false,
					"message":    "Invalid user ID format",
					"error_code": "INVALID_USER_ID",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RecoveryMiddleware 恢复中间件
// RecoveryMiddleware recovery middleware
func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Error("Panic recovered",
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
			logger.String("client_ip", c.ClientIP()),
			logger.String("panic", fmt.Sprintf("%v", recovered)))

		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"message":    "Internal server error",
			"error_code": "PANIC_RECOVERED",
			"request_id": c.GetString("request_id"),
		})
	})
}

// MetricsMiddleware 指标中间件
// MetricsMiddleware metrics middleware
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 记录指标
		// Record metrics
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		path := c.FullPath()
		method := c.Request.Method

		// 这里应该将指标发送到监控系统
		// Should send metrics to monitoring system here
		_ = duration
		_ = statusCode
		_ = path
		_ = method
	}
}

// generateRequestID 生成请求ID
// generateRequestID generates request ID
func generateRequestID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("req_%d", timestamp)
}

// isValidActivityID 验证活动ID格式
// isValidActivityID validates activity ID format
func isValidActivityID(activityID string) bool {
	// 简单的格式验证
	// Simple format validation
	if len(activityID) < 1 || len(activityID) > 50 {
		return false
	}

	// 检查是否包含非法字符
	// Check for illegal characters
	for _, char := range activityID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// isValidUserID 验证用户ID格式
// isValidUserID validates user ID format
func isValidUserID(userID string) bool {
	// 简单的格式验证
	// Simple format validation
	if len(userID) < 1 || len(userID) > 50 {
		return false
	}

	// 检查是否为纯数字或字母数字组合
	// Check if it's pure numbers or alphanumeric combination
	for _, char := range userID {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return false
		}
	}

	return true
}

// ParseIntParam 解析整数参数
// ParseIntParam parses integer parameter
func ParseIntParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// ParseBoolParam 解析布尔参数
// ParseBoolParam parses boolean parameter
func ParseBoolParam(c *gin.Context, key string, defaultValue bool) bool {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
