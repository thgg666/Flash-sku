package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flashsku/seckill/internal/message"
)

// MockProducer 模拟消息生产者
// MockProducer mock message producer
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendOrderMessage(ctx context.Context, orderMsg *message.OrderMessage) error {
	args := m.Called(ctx, orderMsg)
	return args.Error(0)
}

func (m *MockProducer) SendStockSyncMessage(ctx context.Context, stockMsg *message.StockSyncMessage) error {
	args := m.Called(ctx, stockMsg)
	return args.Error(0)
}

func (m *MockProducer) SendEmailMessage(ctx context.Context, emailMsg *message.EmailMessage) error {
	args := m.Called(ctx, emailMsg)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockReliabilityManager 模拟可靠性管理器
// MockReliabilityManager mock reliability manager
type MockReliabilityManager struct {
	mock.Mock
}

func (m *MockReliabilityManager) SendReliableMessage(ctx context.Context, msg *message.ReliableMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockReliabilityManager) ConfirmMessage(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockReliabilityManager) RetryFailedMessages(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockReliabilityManager) GetFailedMessageStats(ctx context.Context) (*message.FailedMessageStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*message.FailedMessageStats), args.Error(1)
}

func (m *MockReliabilityManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestOrderMessageSending(t *testing.T) {
	// 测试订单消息发送
	// Test order message sending
	mockProducer := new(MockProducer)

	ctx := context.Background()
	orderMsg := &message.OrderMessage{
		OrderID:    "order_123",
		UserID:     "user_456",
		ActivityID: "activity_789",
		ProductID:  "product_101",
		Quantity:   2,
		Price:      99.99,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	// 设置期望
	// Set expectations
	mockProducer.On("SendOrderMessage", ctx, orderMsg).Return(nil)

	// 执行测试
	// Execute test
	err := mockProducer.SendOrderMessage(ctx, orderMsg)

	// 验证结果
	// Verify results
	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestStockSyncMessageSending(t *testing.T) {
	// 测试库存同步消息发送
	// Test stock sync message sending
	mockProducer := new(MockProducer)

	ctx := context.Background()
	stockMsg := &message.StockSyncMessage{
		ActivityID:   "activity_789",
		ProductID:    "product_101",
		StockChange:  -2,
		CurrentStock: 98,
		Operation:    "decrease",
		Timestamp:    time.Now(),
		Source:       "seckill",
	}

	// 设置期望
	// Set expectations
	mockProducer.On("SendStockSyncMessage", ctx, stockMsg).Return(nil)

	// 执行测试
	// Execute test
	err := mockProducer.SendStockSyncMessage(ctx, stockMsg)

	// 验证结果
	// Verify results
	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestEmailMessageSending(t *testing.T) {
	// 测试邮件消息发送
	// Test email message sending
	mockProducer := new(MockProducer)

	ctx := context.Background()
	emailMsg := &message.EmailMessage{
		To:       []string{"user@example.com"},
		Subject:  "秒杀成功通知",
		Template: "seckill_success",
		Data: map[string]interface{}{
			"user_name":    "张三",
			"product_name": "iPhone 15",
			"quantity":     1,
		},
		Priority:  1,
		Timestamp: time.Now(),
	}

	// 设置期望
	// Set expectations
	mockProducer.On("SendEmailMessage", ctx, emailMsg).Return(nil)

	// 执行测试
	// Execute test
	err := mockProducer.SendEmailMessage(ctx, emailMsg)

	// 验证结果
	// Verify results
	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

func TestReliableMessageSending(t *testing.T) {
	// 测试可靠消息发送
	// Test reliable message sending
	mockReliabilityManager := new(MockReliabilityManager)

	ctx := context.Background()
	reliableMsg := &message.ReliableMessage{
		ID:         "msg_123",
		Type:       "order",
		Exchange:   "seckill.exchange",
		RoutingKey: "order.created",
		Payload: map[string]interface{}{
			"order_id":    "order_123",
			"user_id":     "user_456",
			"activity_id": "activity_789",
		},
		Status:     "pending",
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 设置期望
	// Set expectations
	mockReliabilityManager.On("SendReliableMessage", ctx, reliableMsg).Return(nil)

	// 执行测试
	// Execute test
	err := mockReliabilityManager.SendReliableMessage(ctx, reliableMsg)

	// 验证结果
	// Verify results
	assert.NoError(t, err)
	mockReliabilityManager.AssertExpectations(t)
}

func TestRetryStrategy(t *testing.T) {
	// 测试重试策略
	// Test retry strategy
	strategy := message.DefaultRetryStrategy()

	// 测试重试延迟计算
	// Test retry delay calculation
	delay1 := strategy.GetRetryDelay(0)
	delay2 := strategy.GetRetryDelay(1)
	delay3 := strategy.GetRetryDelay(2)

	assert.Equal(t, 1*time.Second, delay1)
	assert.True(t, delay2 >= 2*time.Second && delay2 <= 3*time.Second) // 考虑抖动
	assert.True(t, delay3 >= 4*time.Second && delay3 <= 5*time.Second) // 考虑抖动

	// 测试是否应该重试
	// Test should retry
	assert.True(t, strategy.ShouldRetry(0, 3, assert.AnError))
	assert.True(t, strategy.ShouldRetry(2, 3, assert.AnError))
	assert.False(t, strategy.ShouldRetry(3, 3, assert.AnError))
}

func TestErrorHandler(t *testing.T) {
	// 测试错误处理器
	// Test error handler
	mockLogger := &MockLogger{}
	strategy := message.DefaultRetryStrategy()
	handler := message.NewDefaultErrorHandler(strategy, mockLogger)

	ctx := context.Background()
	msg := &message.ReliableMessage{
		ID:         "msg_123",
		Type:       "order",
		Status:     "pending",
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 测试处理可重试错误
	// Test handling retryable error
	err := handler.HandleError(ctx, assert.AnError, msg)
	assert.NoError(t, err) // 错误已处理，返回nil
	assert.Equal(t, "retry_pending", msg.Status)
	assert.Equal(t, 1, msg.RetryCount)

	// 测试超过最大重试次数
	// Test exceeding max retries
	msg.RetryCount = 3
	err = handler.HandleError(ctx, assert.AnError, msg)
	assert.Error(t, err) // 永久失败，返回错误
	assert.Equal(t, "failed", msg.Status)

	// 测试错误统计
	// Test error statistics
	stats := handler.GetErrorStats()
	assert.Equal(t, int64(2), stats.TotalErrors)
	assert.Equal(t, int64(1), stats.PermanentFailures)
}

func TestCircuitBreaker(t *testing.T) {
	// 测试熔断器
	// Test circuit breaker
	cb := message.NewCircuitBreaker(3, 5*time.Second)

	// 测试正常调用
	// Test normal call
	err := cb.Call(func() error {
		return nil
	})
	assert.NoError(t, err)

	// 测试失败调用
	// Test failed calls
	for i := 0; i < 3; i++ {
		err = cb.Call(func() error {
			return assert.AnError
		})
		assert.Error(t, err)
	}

	// 熔断器应该打开
	// Circuit breaker should be open
	err = cb.Call(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestProducerConfig(t *testing.T) {
	// 测试生产者配置
	// Test producer configuration
	config := message.DefaultProducerConfig()

	assert.NotNil(t, config)
	assert.NotEmpty(t, config.OrderExchange)
	assert.NotEmpty(t, config.StockExchange)
	assert.NotEmpty(t, config.EmailExchange)
	assert.NotEmpty(t, config.OrderRoutingKey)
	assert.NotEmpty(t, config.StockRoutingKey)
	assert.NotEmpty(t, config.EmailRoutingKey)
	assert.True(t, config.DefaultMaxRetries > 0)
	assert.True(t, config.RetryInterval > 0)
	assert.True(t, config.PublishTimeout > 0)
}

func TestReliabilityConfig(t *testing.T) {
	// 测试可靠性配置
	// Test reliability configuration
	config := message.DefaultReliabilityConfig()

	assert.NotNil(t, config)
	assert.NotEmpty(t, config.RedisKeyPrefix)
	assert.True(t, config.MessageTTL > 0)
	assert.True(t, config.DefaultMaxRetries > 0)
	assert.True(t, config.RetryInterval > 0)
	assert.True(t, config.RetryBackoff > 1.0)
	assert.True(t, config.BatchSize > 0)
	assert.True(t, config.ProcessInterval > 0)
	assert.True(t, config.DeadLetterTTL > 0)
}
