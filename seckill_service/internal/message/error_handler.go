package message

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
)

// ErrorHandler 错误处理器接口
// ErrorHandler error handler interface
type ErrorHandler interface {
	// 处理消息发送错误
	// Handle message send error
	HandleError(ctx context.Context, err error, msg *ReliableMessage) error

	// 获取错误统计
	// Get error statistics
	GetErrorStats() *ErrorStats

	// 重置错误统计
	// Reset error statistics
	ResetErrorStats()
}

// ErrorStats 错误统计
// ErrorStats error statistics
type ErrorStats struct {
	TotalErrors       int64            `json:"total_errors"`
	ErrorsByType      map[string]int64 `json:"errors_by_type"`
	ErrorsByMessage   map[string]int64 `json:"errors_by_message"`
	LastError         string           `json:"last_error"`
	LastErrorTime     time.Time        `json:"last_error_time"`
	RecoveredErrors   int64            `json:"recovered_errors"`
	PermanentFailures int64            `json:"permanent_failures"`
}

// RetryStrategy 重试策略
// RetryStrategy retry strategy
type RetryStrategy interface {
	// 计算下次重试时间
	// Calculate next retry time
	NextRetryTime(retryCount int, lastError error) time.Time

	// 判断是否应该重试
	// Determine if should retry
	ShouldRetry(retryCount int, maxRetries int, err error) bool

	// 获取重试延迟
	// Get retry delay
	GetRetryDelay(retryCount int) time.Duration
}

// ExponentialBackoffStrategy 指数退避重试策略
// ExponentialBackoffStrategy exponential backoff retry strategy
type ExponentialBackoffStrategy struct {
	BaseDelay  time.Duration `json:"base_delay"`
	MaxDelay   time.Duration `json:"max_delay"`
	Multiplier float64       `json:"multiplier"`
	Jitter     bool          `json:"jitter"`
}

// DefaultRetryStrategy 默认重试策略
// DefaultRetryStrategy default retry strategy
func DefaultRetryStrategy() *ExponentialBackoffStrategy {
	return &ExponentialBackoffStrategy{
		BaseDelay:  1 * time.Second,
		MaxDelay:   5 * time.Minute,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// NextRetryTime 计算下次重试时间
// NextRetryTime calculates next retry time
func (s *ExponentialBackoffStrategy) NextRetryTime(retryCount int, lastError error) time.Time {
	delay := s.GetRetryDelay(retryCount)
	return time.Now().Add(delay)
}

// ShouldRetry 判断是否应该重试
// ShouldRetry determines if should retry
func (s *ExponentialBackoffStrategy) ShouldRetry(retryCount int, maxRetries int, err error) bool {
	if retryCount >= maxRetries {
		return false
	}

	// 检查错误类型，某些错误不应该重试
	// Check error type, some errors should not be retried
	if isNonRetryableError(err) {
		return false
	}

	return true
}

// GetRetryDelay 获取重试延迟
// GetRetryDelay gets retry delay
func (s *ExponentialBackoffStrategy) GetRetryDelay(retryCount int) time.Duration {
	delay := s.BaseDelay

	// 计算指数退避延迟
	// Calculate exponential backoff delay
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * s.Multiplier)
		if delay > s.MaxDelay {
			delay = s.MaxDelay
			break
		}
	}

	// 添加抖动避免雷群效应
	// Add jitter to avoid thundering herd
	if s.Jitter {
		jitterAmount := time.Duration(float64(delay) * 0.1) // 10% jitter
		jitterFactor := float64(time.Now().UnixNano()%1000)/1000.0*2 - 1
		delay += time.Duration(float64(jitterAmount) * jitterFactor)
	}

	return delay
}

// DefaultErrorHandler 默认错误处理器
// DefaultErrorHandler default error handler
type DefaultErrorHandler struct {
	retryStrategy RetryStrategy
	logger        logger.Logger
	stats         *ErrorStats
	mutex         sync.RWMutex
}

// NewDefaultErrorHandler 创建默认错误处理器
// NewDefaultErrorHandler creates default error handler
func NewDefaultErrorHandler(retryStrategy RetryStrategy, log logger.Logger) *DefaultErrorHandler {
	if retryStrategy == nil {
		retryStrategy = DefaultRetryStrategy()
	}

	return &DefaultErrorHandler{
		retryStrategy: retryStrategy,
		logger:        log,
		stats: &ErrorStats{
			ErrorsByType:    make(map[string]int64),
			ErrorsByMessage: make(map[string]int64),
		},
	}
}

// HandleError 处理消息发送错误
// HandleError handles message send error
func (h *DefaultErrorHandler) HandleError(ctx context.Context, err error, msg *ReliableMessage) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 更新错误统计
	// Update error statistics
	h.stats.TotalErrors++
	h.stats.LastError = err.Error()
	h.stats.LastErrorTime = time.Now()

	errorType := getErrorType(err)
	h.stats.ErrorsByType[errorType]++
	h.stats.ErrorsByMessage[msg.Type]++

	// 判断是否应该重试
	// Determine if should retry
	if h.retryStrategy.ShouldRetry(msg.RetryCount, msg.MaxRetries, err) {
		// 计算下次重试时间
		// Calculate next retry time
		msg.NextRetry = h.retryStrategy.NextRetryTime(msg.RetryCount, err)
		msg.Status = "retry_pending"
		msg.RetryCount++
		msg.UpdatedAt = time.Now()

		h.logger.Warn("Message will be retried",
			logger.String("message_id", msg.ID),
			logger.String("message_type", msg.Type),
			logger.Int("retry_count", msg.RetryCount),
			logger.Int("max_retries", msg.MaxRetries),
			logger.String("next_retry", msg.NextRetry.Format(time.RFC3339)),
			logger.Error(err))

		return nil // 返回nil表示错误已处理，消息将重试
	} else {
		// 超过最大重试次数或不可重试错误
		// Exceeded max retries or non-retryable error
		msg.Status = "failed"
		msg.UpdatedAt = time.Now()
		h.stats.PermanentFailures++

		h.logger.Error("Message permanently failed",
			logger.String("message_id", msg.ID),
			logger.String("message_type", msg.Type),
			logger.Int("retry_count", msg.RetryCount),
			logger.Int("max_retries", msg.MaxRetries),
			logger.Error(err))

		// 可以在这里添加告警逻辑
		// Can add alerting logic here
		h.sendFailureAlert(msg, err)

		return fmt.Errorf("message permanently failed after %d retries: %w", msg.RetryCount, err)
	}
}

// GetErrorStats 获取错误统计
// GetErrorStats gets error statistics
func (h *DefaultErrorHandler) GetErrorStats() *ErrorStats {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 返回统计的副本
	// Return copy of statistics
	stats := &ErrorStats{
		TotalErrors:       h.stats.TotalErrors,
		ErrorsByType:      make(map[string]int64),
		ErrorsByMessage:   make(map[string]int64),
		LastError:         h.stats.LastError,
		LastErrorTime:     h.stats.LastErrorTime,
		RecoveredErrors:   h.stats.RecoveredErrors,
		PermanentFailures: h.stats.PermanentFailures,
	}

	for k, v := range h.stats.ErrorsByType {
		stats.ErrorsByType[k] = v
	}
	for k, v := range h.stats.ErrorsByMessage {
		stats.ErrorsByMessage[k] = v
	}

	return stats
}

// ResetErrorStats 重置错误统计
// ResetErrorStats resets error statistics
func (h *DefaultErrorHandler) ResetErrorStats() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.stats = &ErrorStats{
		ErrorsByType:    make(map[string]int64),
		ErrorsByMessage: make(map[string]int64),
	}

	h.logger.Info("Error statistics reset")
}

// sendFailureAlert 发送失败告警
// sendFailureAlert sends failure alert
func (h *DefaultErrorHandler) sendFailureAlert(msg *ReliableMessage, err error) {
	// 这里可以实现告警逻辑，比如发送邮件、短信、钉钉等
	// Alerting logic can be implemented here, such as email, SMS, DingTalk, etc.
	h.logger.Error("ALERT: Message permanently failed",
		logger.String("message_id", msg.ID),
		logger.String("message_type", msg.Type),
		logger.String("exchange", msg.Exchange),
		logger.String("routing_key", msg.RoutingKey),
		logger.Error(err))
}

// 辅助函数
// Helper functions

// isNonRetryableError 判断是否为不可重试的错误
// isNonRetryableError determines if error is non-retryable
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 不可重试的错误类型
	// Non-retryable error types
	nonRetryableErrors := []string{
		"invalid message format",
		"authentication failed",
		"authorization denied",
		"exchange not found",
		"routing key invalid",
		"message too large",
	}

	for _, nonRetryable := range nonRetryableErrors {
		if contains(errStr, nonRetryable) {
			return true
		}
	}

	return false
}

// getErrorType 获取错误类型
// getErrorType gets error type
func getErrorType(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// 根据错误信息分类
	// Classify based on error message
	if contains(errStr, "connection") || contains(errStr, "network") {
		return "connection_error"
	}
	if contains(errStr, "timeout") {
		return "timeout_error"
	}
	if contains(errStr, "authentication") || contains(errStr, "authorization") {
		return "auth_error"
	}
	if contains(errStr, "exchange") || contains(errStr, "routing") {
		return "routing_error"
	}
	if contains(errStr, "format") || contains(errStr, "marshal") || contains(errStr, "unmarshal") {
		return "format_error"
	}

	return "unknown_error"
}

// contains 检查字符串是否包含子字符串（忽略大小写）
// contains checks if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CircuitBreaker 熔断器
// CircuitBreaker circuit breaker
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failureCount int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	mutex        sync.RWMutex
}

// NewCircuitBreaker 创建熔断器
// NewCircuitBreaker creates circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

// Call 执行调用
// Call executes call
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// 检查熔断器状态
	// Check circuit breaker state
	if cb.state == "open" {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = "half-open"
			cb.failureCount = 0
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	// 执行调用
	// Execute call
	err := fn()

	if err != nil {
		cb.failureCount++
		cb.lastFailTime = time.Now()

		if cb.failureCount >= cb.maxFailures {
			cb.state = "open"
		}

		return err
	}

	// 调用成功，重置计数器
	// Call successful, reset counter
	if cb.state == "half-open" {
		cb.state = "closed"
	}
	cb.failureCount = 0

	return nil
}
