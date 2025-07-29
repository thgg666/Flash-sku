package message

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// ReliabilityManager 消息可靠性管理器
// ReliabilityManager message reliability manager
type ReliabilityManager interface {
	// 发送可靠消息
	// Send reliable message
	SendReliableMessage(ctx context.Context, msg *ReliableMessage) error
	
	// 确认消息
	// Confirm message
	ConfirmMessage(ctx context.Context, messageID string) error
	
	// 重试失败消息
	// Retry failed messages
	RetryFailedMessages(ctx context.Context) error
	
	// 获取失败消息统计
	// Get failed message statistics
	GetFailedMessageStats(ctx context.Context) (*FailedMessageStats, error)
	
	// 关闭管理器
	// Close manager
	Close() error
}

// ReliableMessage 可靠消息
// ReliableMessage reliable message
type ReliableMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // "order", "stock", "email"
	Exchange    string                 `json:"exchange"`
	RoutingKey  string                 `json:"routing_key"`
	Payload     map[string]interface{} `json:"payload"`
	
	// 可靠性相关
	// Reliability related
	Status      string    `json:"status"`       // "pending", "sent", "confirmed", "failed"
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	NextRetry   time.Time `json:"next_retry"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 错误信息
	// Error information
	LastError   string    `json:"last_error"`
	ErrorCount  int       `json:"error_count"`
}

// FailedMessageStats 失败消息统计
// FailedMessageStats failed message statistics
type FailedMessageStats struct {
	TotalFailed    int64             `json:"total_failed"`
	FailedByType   map[string]int64  `json:"failed_by_type"`
	RetryPending   int64             `json:"retry_pending"`
	DeadLetters    int64             `json:"dead_letters"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// ReliabilityConfig 可靠性配置
// ReliabilityConfig reliability configuration
type ReliabilityConfig struct {
	// Redis配置
	// Redis configuration
	RedisKeyPrefix    string        `json:"redis_key_prefix"`
	MessageTTL        time.Duration `json:"message_ttl"`
	
	// 重试配置
	// Retry configuration
	DefaultMaxRetries int           `json:"default_max_retries"`
	RetryInterval     time.Duration `json:"retry_interval"`
	RetryBackoff      float64       `json:"retry_backoff"`
	
	// 批处理配置
	// Batch processing configuration
	BatchSize         int           `json:"batch_size"`
	ProcessInterval   time.Duration `json:"process_interval"`
	
	// 死信配置
	// Dead letter configuration
	DeadLetterTTL     time.Duration `json:"dead_letter_ttl"`
}

// DefaultReliabilityConfig 默认可靠性配置
// DefaultReliabilityConfig default reliability configuration
func DefaultReliabilityConfig() *ReliabilityConfig {
	return &ReliabilityConfig{
		RedisKeyPrefix:    "message:reliability:",
		MessageTTL:        24 * time.Hour,
		DefaultMaxRetries: 3,
		RetryInterval:     5 * time.Second,
		RetryBackoff:      2.0,
		BatchSize:         100,
		ProcessInterval:   30 * time.Second,
		DeadLetterTTL:     7 * 24 * time.Hour, // 7天
	}
}

// RedisReliabilityManager Redis可靠性管理器实现
// RedisReliabilityManager Redis reliability manager implementation
type RedisReliabilityManager struct {
	producer Producer
	redis    redis.Client
	config   *ReliabilityConfig
	logger   logger.Logger
	
	// 后台处理
	// Background processing
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewRedisReliabilityManager 创建Redis可靠性管理器
// NewRedisReliabilityManager creates Redis reliability manager
func NewRedisReliabilityManager(
	producer Producer,
	redisClient redis.Client,
	config *ReliabilityConfig,
	log logger.Logger,
) *RedisReliabilityManager {
	if config == nil {
		config = DefaultReliabilityConfig()
	}
	
	manager := &RedisReliabilityManager{
		producer: producer,
		redis:    redisClient,
		config:   config,
		logger:   log,
		stopChan: make(chan struct{}),
	}
	
	// 启动后台重试处理
	// Start background retry processing
	manager.startBackgroundProcessing()
	
	return manager
}

// SendReliableMessage 发送可靠消息
// SendReliableMessage sends reliable message
func (rm *RedisReliabilityManager) SendReliableMessage(ctx context.Context, msg *ReliableMessage) error {
	// 设置默认值
	// Set default values
	if msg.ID == "" {
		msg.ID = generateMessageID()
	}
	if msg.Status == "" {
		msg.Status = "pending"
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = rm.config.DefaultMaxRetries
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	msg.UpdatedAt = time.Now()
	
	// 保存消息到Redis
	// Save message to Redis
	if err := rm.saveMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	
	// 尝试发送消息
	// Attempt to send message
	if err := rm.sendMessage(ctx, msg); err != nil {
		// 发送失败，标记为失败状态
		// Send failed, mark as failed
		msg.Status = "failed"
		msg.LastError = err.Error()
		msg.ErrorCount++
		msg.NextRetry = time.Now().Add(rm.calculateRetryDelay(msg.RetryCount))
		msg.UpdatedAt = time.Now()
		
		rm.saveMessage(ctx, msg)
		
		rm.logger.Error("Failed to send reliable message",
			logger.String("message_id", msg.ID),
			logger.String("type", msg.Type),
			logger.Error(err))
		
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	// 发送成功，更新状态
	// Send successful, update status
	msg.Status = "sent"
	msg.UpdatedAt = time.Now()
	rm.saveMessage(ctx, msg)
	
	rm.logger.Info("Reliable message sent successfully",
		logger.String("message_id", msg.ID),
		logger.String("type", msg.Type),
		logger.String("exchange", msg.Exchange),
		logger.String("routing_key", msg.RoutingKey))
	
	return nil
}

// ConfirmMessage 确认消息
// ConfirmMessage confirms message
func (rm *RedisReliabilityManager) ConfirmMessage(ctx context.Context, messageID string) error {
	key := rm.getMessageKey(messageID)
	
	// 获取消息
	// Get message
	msg, err := rm.getMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}
	
	if msg == nil {
		return fmt.Errorf("message not found: %s", messageID)
	}
	
	// 更新状态为已确认
	// Update status to confirmed
	msg.Status = "confirmed"
	msg.UpdatedAt = time.Now()
	
	// 保存更新
	// Save update
	if err := rm.saveMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	
	// 设置TTL，消息将在一段时间后自动删除
	// Set TTL, message will be automatically deleted after some time
	rm.redis.Expire(ctx, key, rm.config.MessageTTL)
	
	rm.logger.Info("Message confirmed successfully",
		logger.String("message_id", messageID))
	
	return nil
}

// RetryFailedMessages 重试失败消息
// RetryFailedMessages retries failed messages
func (rm *RedisReliabilityManager) RetryFailedMessages(ctx context.Context) error {
	// 获取需要重试的消息
	// Get messages that need retry
	messages, err := rm.getRetryableMessages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get retryable messages: %w", err)
	}
	
	retryCount := 0
	successCount := 0
	
	for _, msg := range messages {
		// 检查是否超过最大重试次数
		// Check if max retries exceeded
		if msg.RetryCount >= msg.MaxRetries {
			// 移动到死信队列
			// Move to dead letter queue
			if err := rm.moveToDeadLetter(ctx, msg); err != nil {
				rm.logger.Error("Failed to move message to dead letter queue",
					logger.String("message_id", msg.ID),
					logger.Error(err))
			}
			continue
		}
		
		// 尝试重新发送
		// Attempt to resend
		msg.RetryCount++
		msg.UpdatedAt = time.Now()
		
		if err := rm.sendMessage(ctx, msg); err != nil {
			// 重试失败
			// Retry failed
			msg.Status = "failed"
			msg.LastError = err.Error()
			msg.ErrorCount++
			msg.NextRetry = time.Now().Add(rm.calculateRetryDelay(msg.RetryCount))
			
			rm.logger.Warn("Message retry failed",
				logger.String("message_id", msg.ID),
				logger.Int("retry_count", msg.RetryCount),
				logger.Error(err))
		} else {
			// 重试成功
			// Retry successful
			msg.Status = "sent"
			msg.LastError = ""
			successCount++
			
			rm.logger.Info("Message retry successful",
				logger.String("message_id", msg.ID),
				logger.Int("retry_count", msg.RetryCount))
		}
		
		rm.saveMessage(ctx, msg)
		retryCount++
	}
	
	rm.logger.Info("Retry process completed",
		logger.Int("total_retried", retryCount),
		logger.Int("successful", successCount),
		logger.Int("failed", retryCount-successCount))
	
	return nil
}

// GetFailedMessageStats 获取失败消息统计
// GetFailedMessageStats gets failed message statistics
func (rm *RedisReliabilityManager) GetFailedMessageStats(ctx context.Context) (*FailedMessageStats, error) {
	// 这里可以实现更复杂的统计逻辑
	// More complex statistics logic can be implemented here
	stats := &FailedMessageStats{
		FailedByType: make(map[string]int64),
		LastUpdated:  time.Now(),
	}
	
	// 简单实现：扫描所有消息并统计
	// Simple implementation: scan all messages and count
	// 在生产环境中，建议使用更高效的方法，如维护单独的统计计数器
	// In production, recommend using more efficient methods like maintaining separate counters
	
	return stats, nil
}

// Close 关闭管理器
// Close closes manager
func (rm *RedisReliabilityManager) Close() error {
	close(rm.stopChan)
	rm.wg.Wait()
	rm.logger.Info("Reliability manager closed")
	return nil
}

// 私有方法
// Private methods

func (rm *RedisReliabilityManager) startBackgroundProcessing() {
	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()
		
		ticker := time.NewTicker(rm.config.ProcessInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := rm.RetryFailedMessages(ctx); err != nil {
					rm.logger.Error("Background retry failed", logger.Error(err))
				}
				cancel()
				
			case <-rm.stopChan:
				return
			}
		}
	}()
}

func (rm *RedisReliabilityManager) sendMessage(ctx context.Context, msg *ReliableMessage) error {
	switch msg.Type {
	case "order":
		var orderMsg OrderMessage
		if err := mapToStruct(msg.Payload, &orderMsg); err != nil {
			return err
		}
		return rm.producer.SendOrderMessage(ctx, &orderMsg)
		
	case "stock":
		var stockMsg StockSyncMessage
		if err := mapToStruct(msg.Payload, &stockMsg); err != nil {
			return err
		}
		return rm.producer.SendStockSyncMessage(ctx, &stockMsg)
		
	case "email":
		var emailMsg EmailMessage
		if err := mapToStruct(msg.Payload, &emailMsg); err != nil {
			return err
		}
		return rm.producer.SendEmailMessage(ctx, &emailMsg)
		
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (rm *RedisReliabilityManager) saveMessage(ctx context.Context, msg *ReliableMessage) error {
	key := rm.getMessageKey(msg.ID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	return rm.redis.Set(ctx, key, string(data), rm.config.MessageTTL)
}

func (rm *RedisReliabilityManager) getMessage(ctx context.Context, messageID string) (*ReliableMessage, error) {
	key := rm.getMessageKey(messageID)
	data, err := rm.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	
	if data == "" {
		return nil, nil
	}
	
	var msg ReliableMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return nil, err
	}
	
	return &msg, nil
}

func (rm *RedisReliabilityManager) getRetryableMessages(ctx context.Context) ([]*ReliableMessage, error) {
	// 简化实现：这里应该使用更高效的方法来查找需要重试的消息
	// Simplified implementation: should use more efficient method to find retryable messages
	// 例如使用Redis的有序集合来按时间排序
	// For example, use Redis sorted sets to sort by time
	
	return []*ReliableMessage{}, nil
}

func (rm *RedisReliabilityManager) moveToDeadLetter(ctx context.Context, msg *ReliableMessage) error {
	deadLetterKey := rm.getDeadLetterKey(msg.ID)
	msg.Status = "dead_letter"
	msg.UpdatedAt = time.Now()
	
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	// 保存到死信队列
	// Save to dead letter queue
	if err := rm.redis.Set(ctx, deadLetterKey, string(data), rm.config.DeadLetterTTL); err != nil {
		return err
	}
	
	// 删除原消息
	// Delete original message
	originalKey := rm.getMessageKey(msg.ID)
	rm.redis.Del(ctx, originalKey)
	
	return nil
}

func (rm *RedisReliabilityManager) calculateRetryDelay(retryCount int) time.Duration {
	// 指数退避算法
	// Exponential backoff algorithm
	delay := rm.config.RetryInterval
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * rm.config.RetryBackoff)
	}
	return delay
}

func (rm *RedisReliabilityManager) getMessageKey(messageID string) string {
	return rm.config.RedisKeyPrefix + "message:" + messageID
}

func (rm *RedisReliabilityManager) getDeadLetterKey(messageID string) string {
	return rm.config.RedisKeyPrefix + "dead_letter:" + messageID
}

// 辅助函数
// Helper functions

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func mapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}
