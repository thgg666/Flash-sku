package model

import "time"

// Response 统一响应结构体
// Response unified response structure
type Response struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SeckillResponse 秒杀响应
// SeckillResponse seckill response
type SeckillResponse struct {
	Success    bool   `json:"success"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	OrderID    string `json:"order_id,omitempty"`
	ActivityID string `json:"activity_id"`
}

// StockResponse 库存响应
// StockResponse stock response
type StockResponse struct {
	ActivityID   string `json:"activity_id"`
	CurrentStock int    `json:"current_stock"`
	TotalStock   int    `json:"total_stock"`
}

// HealthResponse 健康检查响应
// HealthResponse health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Service   string            `json:"service"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricsResponse 性能指标响应
// MetricsResponse metrics response
type MetricsResponse struct {
	RequestCount     int64   `json:"request_count"`
	SuccessCount     int64   `json:"success_count"`
	ErrorCount       int64   `json:"error_count"`
	AvgResponseTime  float64 `json:"avg_response_time"`
	GoroutineCount   int     `json:"goroutine_count"`
	MemoryUsage      uint64  `json:"memory_usage"`
	Timestamp        time.Time `json:"timestamp"`
}

// ErrorResponse 错误响应
// ErrorResponse error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// 常用响应代码
// Common response codes
const (
	CodeSuccess         = "SUCCESS"
	CodeInternalError   = "INTERNAL_ERROR"
	CodeInvalidParam    = "INVALID_PARAM"
	CodeRateLimit       = "RATE_LIMIT"
	CodeActivityInvalid = "ACTIVITY_INVALID"
	CodeSoldOut         = "SOLD_OUT"
	CodeAlreadyBought   = "ALREADY_BOUGHT"
	CodeUserNotFound    = "USER_NOT_FOUND"
)

// NewSuccessResponse 创建成功响应
// NewSuccessResponse creates success response
func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:      CodeSuccess,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse 创建错误响应
// NewErrorResponse creates error response
func NewErrorResponse(code, message string) *Response {
	return &Response{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}
}
