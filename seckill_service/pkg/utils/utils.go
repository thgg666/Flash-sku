package utils

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

// GenerateID 生成唯一ID
// GenerateID generates unique ID
func GenerateID() string {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(10000)
	return fmt.Sprintf("%d%04d", timestamp, random)
}

// GenerateOrderID 生成订单ID
// GenerateOrderID generates order ID
func GenerateOrderID(userID, activityID string) string {
	timestamp := time.Now().Format("20060102150405")
	hash := MD5Hash(userID + activityID + strconv.FormatInt(time.Now().UnixNano(), 10))
	return fmt.Sprintf("ORD%s%s", timestamp, hash[:8])
}

// MD5Hash 计算MD5哈希
// MD5Hash calculates MD5 hash
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// GetGoroutineCount 获取当前协程数量
// GetGoroutineCount gets current goroutine count
func GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// GetMemoryUsage 获取内存使用情况
// GetMemoryUsage gets memory usage
func GetMemoryUsage() (uint64, uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc, m.Sys
}

// FormatBytes 格式化字节数
// FormatBytes formats bytes
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// MinInt 返回两个整数中的最小值
// MinInt returns minimum of two integers
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt 返回两个整数中的最大值
// MaxInt returns maximum of two integers
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt64 返回两个64位整数中的最小值
// MinInt64 returns minimum of two int64
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两个64位整数中的最大值
// MaxInt64 returns maximum of two int64
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// IsValidActivityID 验证活动ID格式
// IsValidActivityID validates activity ID format
func IsValidActivityID(activityID string) bool {
	if len(activityID) == 0 || len(activityID) > 50 {
		return false
	}
	// TODO: 添加更严格的格式验证
	// TODO: Add stricter format validation
	return true
}

// IsValidUserID 验证用户ID格式
// IsValidUserID validates user ID format
func IsValidUserID(userID string) bool {
	if len(userID) == 0 || len(userID) > 50 {
		return false
	}
	// TODO: 添加更严格的格式验证
	// TODO: Add stricter format validation
	return true
}

// GetClientIP 获取客户端IP地址
// GetClientIP gets client IP address
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	if xRealIP != "" {
		return xRealIP
	}
	if xForwardedFor != "" {
		return xForwardedFor
	}
	return remoteAddr
}

// SafeString 安全字符串处理，防止注入
// SafeString safe string processing to prevent injection
func SafeString(s string) string {
	// TODO: 实现字符串安全处理
	// TODO: Implement string safety processing
	return s
}
