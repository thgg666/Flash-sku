package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQClientImpl RabbitMQ客户端实现
// RabbitMQClientImpl RabbitMQ client implementation
type RabbitMQClientImpl struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *Config
}

// NewClientImpl 创建新的RabbitMQ客户端实现
// NewClientImpl creates new RabbitMQ client implementation
func NewClientImpl(config *Config) (Client, error) {
	// 连接到RabbitMQ
	// Connect to RabbitMQ
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// 创建通道
	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明交换机
	// Declare exchange
	if config.Exchange != "" {
		err = channel.ExchangeDeclare(
			config.Exchange,     // name
			config.ExchangeType, // type
			config.Durable,      // durable
			config.AutoDelete,   // auto-deleted
			false,               // internal
			false,               // no-wait
			nil,                 // arguments
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return nil, fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	client := &RabbitMQClientImpl{
		conn:    conn,
		channel: channel,
		config:  config,
	}

	return client, nil
}

// Publish 发布消息
// Publish publishes message
func (c *RabbitMQClientImpl) Publish(ctx context.Context, exchange, routingKey string, message any) error {
	// 序列化消息
	// Serialize message
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 发布消息
	// Publish message
	err = c.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishWithConfirm 发布消息并确认
// PublishWithConfirm publishes message with confirmation
func (c *RabbitMQClientImpl) PublishWithConfirm(ctx context.Context, exchange, routingKey string, message any) error {
	// 启用发布确认模式
	// Enable publisher confirms
	if err := c.channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable confirm mode: %w", err)
	}

	// 发布消息
	// Publish message
	if err := c.Publish(ctx, exchange, routingKey, message); err != nil {
		return err
	}

	// 等待确认
	// Wait for confirmation
	select {
	case confirm := <-c.channel.NotifyPublish(make(chan amqp.Confirmation, 1)):
		if !confirm.Ack {
			return fmt.Errorf("message was not acknowledged")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for confirmation")
	}
}

// Close 关闭连接
// Close closes connection
func (c *RabbitMQClientImpl) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected 检查连接状态
// IsConnected checks connection status
func (c *RabbitMQClientImpl) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
