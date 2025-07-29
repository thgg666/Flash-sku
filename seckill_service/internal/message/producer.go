package message

import (
	"context"
	"fmt"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/rabbitmq"
)

// Producer 消息生产者接口
// Producer message producer interface
type Producer interface {
	// 发送订单消息
	// Send order message
	SendOrderMessage(ctx context.Context, orderMsg *OrderMessage) error

	// 发送库存同步消息
	// Send stock sync message
	SendStockSyncMessage(ctx context.Context, stockMsg *StockSyncMessage) error

	// 发送邮件通知消息
	// Send email notification message
	SendEmailMessage(ctx context.Context, emailMsg *EmailMessage) error

	// 关闭生产者
	// Close producer
	Close() error
}

// OrderMessage 订单消息
// OrderMessage order message
type OrderMessage struct {
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	ActivityID string    `json:"activity_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Price      float64   `json:"price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`

	// 重试相关
	// Retry related
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	NextRetry  time.Time `json:"next_retry"`
}

// StockSyncMessage 库存同步消息
// StockSyncMessage stock sync message
type StockSyncMessage struct {
	ActivityID   string    `json:"activity_id"`
	ProductID    string    `json:"product_id"`
	StockChange  int       `json:"stock_change"`
	CurrentStock int       `json:"current_stock"`
	Operation    string    `json:"operation"` // "decrease", "increase", "sync"
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"` // "seckill", "admin", "sync"
}

// EmailMessage 邮件消息
// EmailMessage email message
type EmailMessage struct {
	To        []string               `json:"to"`
	Subject   string                 `json:"subject"`
	Template  string                 `json:"template"`
	Data      map[string]interface{} `json:"data"`
	Priority  int                    `json:"priority"` // 1-高, 2-中, 3-低
	Timestamp time.Time              `json:"timestamp"`
}

// ProducerConfig 生产者配置
// ProducerConfig producer configuration
type ProducerConfig struct {
	// RabbitMQ配置
	// RabbitMQ configuration
	RabbitMQConfig *rabbitmq.Config `json:"rabbitmq_config"`

	// 交换机和路由配置
	// Exchange and routing configuration
	OrderExchange string `json:"order_exchange"`
	StockExchange string `json:"stock_exchange"`
	EmailExchange string `json:"email_exchange"`

	OrderRoutingKey string `json:"order_routing_key"`
	StockRoutingKey string `json:"stock_routing_key"`
	EmailRoutingKey string `json:"email_routing_key"`

	// 重试配置
	// Retry configuration
	DefaultMaxRetries int           `json:"default_max_retries"`
	RetryInterval     time.Duration `json:"retry_interval"`

	// 性能配置
	// Performance configuration
	EnableConfirm  bool          `json:"enable_confirm"`
	PublishTimeout time.Duration `json:"publish_timeout"`
}

// DefaultProducerConfig 默认生产者配置
// DefaultProducerConfig default producer configuration
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		RabbitMQConfig: &rabbitmq.Config{
			URL:          "amqp://guest:guest@localhost:5672/",
			Exchange:     "seckill.exchange",
			ExchangeType: "topic",
			Durable:      true,
			AutoDelete:   false,
		},
		OrderExchange:     "seckill.exchange",
		StockExchange:     "seckill.exchange",
		EmailExchange:     "seckill.exchange",
		OrderRoutingKey:   "order.created",
		StockRoutingKey:   "stock.sync",
		EmailRoutingKey:   "email.send",
		DefaultMaxRetries: 3,
		RetryInterval:     5 * time.Second,
		EnableConfirm:     true,
		PublishTimeout:    10 * time.Second,
	}
}

// RabbitMQProducer RabbitMQ消息生产者实现
// RabbitMQProducer RabbitMQ message producer implementation
type RabbitMQProducer struct {
	client rabbitmq.Client
	config *ProducerConfig
	logger logger.Logger
}

// NewRabbitMQProducer 创建新的RabbitMQ生产者
// NewRabbitMQProducer creates new RabbitMQ producer
func NewRabbitMQProducer(config *ProducerConfig, log logger.Logger) (*RabbitMQProducer, error) {
	if config == nil {
		config = DefaultProducerConfig()
	}

	// 创建RabbitMQ客户端
	// Create RabbitMQ client
	client, err := rabbitmq.NewClientImpl(config.RabbitMQConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ client: %w", err)
	}

	producer := &RabbitMQProducer{
		client: client,
		config: config,
		logger: log,
	}

	// 初始化交换机和队列
	// Initialize exchanges and queues
	if err := producer.setupInfrastructure(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to setup infrastructure: %w", err)
	}

	log.Info("RabbitMQ producer created successfully",
		logger.String("order_exchange", config.OrderExchange),
		logger.String("stock_exchange", config.StockExchange),
		logger.String("email_exchange", config.EmailExchange))

	return producer, nil
}

// SendOrderMessage 发送订单消息
// SendOrderMessage sends order message
func (p *RabbitMQProducer) SendOrderMessage(ctx context.Context, orderMsg *OrderMessage) error {
	// 设置默认值
	// Set default values
	if orderMsg.CreatedAt.IsZero() {
		orderMsg.CreatedAt = time.Now()
	}
	if orderMsg.MaxRetries == 0 {
		orderMsg.MaxRetries = p.config.DefaultMaxRetries
	}

	// 创建带超时的上下文
	// Create context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, p.config.PublishTimeout)
	defer cancel()

	// 发送消息
	// Send message
	var err error
	if p.config.EnableConfirm {
		err = p.client.PublishWithConfirm(publishCtx, p.config.OrderExchange, p.config.OrderRoutingKey, orderMsg)
	} else {
		err = p.client.Publish(publishCtx, p.config.OrderExchange, p.config.OrderRoutingKey, orderMsg)
	}

	if err != nil {
		p.logger.Error("Failed to send order message",
			logger.String("order_id", orderMsg.OrderID),
			logger.String("user_id", orderMsg.UserID),
			logger.String("activity_id", orderMsg.ActivityID),
			logger.Error(err))
		return fmt.Errorf("failed to send order message: %w", err)
	}

	p.logger.Info("Order message sent successfully",
		logger.String("order_id", orderMsg.OrderID),
		logger.String("user_id", orderMsg.UserID),
		logger.String("activity_id", orderMsg.ActivityID),
		logger.String("routing_key", p.config.OrderRoutingKey))

	return nil
}

// SendStockSyncMessage 发送库存同步消息
// SendStockSyncMessage sends stock sync message
func (p *RabbitMQProducer) SendStockSyncMessage(ctx context.Context, stockMsg *StockSyncMessage) error {
	// 设置默认值
	// Set default values
	if stockMsg.Timestamp.IsZero() {
		stockMsg.Timestamp = time.Now()
	}
	if stockMsg.Source == "" {
		stockMsg.Source = "seckill"
	}

	// 创建带超时的上下文
	// Create context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, p.config.PublishTimeout)
	defer cancel()

	// 发送消息
	// Send message
	var err error
	if p.config.EnableConfirm {
		err = p.client.PublishWithConfirm(publishCtx, p.config.StockExchange, p.config.StockRoutingKey, stockMsg)
	} else {
		err = p.client.Publish(publishCtx, p.config.StockExchange, p.config.StockRoutingKey, stockMsg)
	}

	if err != nil {
		p.logger.Error("Failed to send stock sync message",
			logger.String("activity_id", stockMsg.ActivityID),
			logger.String("product_id", stockMsg.ProductID),
			logger.Int("stock_change", stockMsg.StockChange),
			logger.Error(err))
		return fmt.Errorf("failed to send stock sync message: %w", err)
	}

	p.logger.Info("Stock sync message sent successfully",
		logger.String("activity_id", stockMsg.ActivityID),
		logger.String("product_id", stockMsg.ProductID),
		logger.Int("stock_change", stockMsg.StockChange),
		logger.String("operation", stockMsg.Operation))

	return nil
}

// SendEmailMessage 发送邮件消息
// SendEmailMessage sends email message
func (p *RabbitMQProducer) SendEmailMessage(ctx context.Context, emailMsg *EmailMessage) error {
	// 设置默认值
	// Set default values
	if emailMsg.Timestamp.IsZero() {
		emailMsg.Timestamp = time.Now()
	}
	if emailMsg.Priority == 0 {
		emailMsg.Priority = 2 // 默认中等优先级
	}

	// 创建带超时的上下文
	// Create context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, p.config.PublishTimeout)
	defer cancel()

	// 发送消息
	// Send message
	var err error
	if p.config.EnableConfirm {
		err = p.client.PublishWithConfirm(publishCtx, p.config.EmailExchange, p.config.EmailRoutingKey, emailMsg)
	} else {
		err = p.client.Publish(publishCtx, p.config.EmailExchange, p.config.EmailRoutingKey, emailMsg)
	}

	if err != nil {
		p.logger.Error("Failed to send email message",
			logger.String("to", fmt.Sprintf("%v", emailMsg.To)),
			logger.String("subject", emailMsg.Subject),
			logger.String("template", emailMsg.Template),
			logger.Error(err))
		return fmt.Errorf("failed to send email message: %w", err)
	}

	p.logger.Info("Email message sent successfully",
		logger.String("to", fmt.Sprintf("%v", emailMsg.To)),
		logger.String("subject", emailMsg.Subject),
		logger.String("template", emailMsg.Template),
		logger.Int("priority", emailMsg.Priority))

	return nil
}

// Close 关闭生产者
// Close closes producer
func (p *RabbitMQProducer) Close() error {
	if p.client != nil {
		err := p.client.Close()
		p.logger.Info("RabbitMQ producer closed")
		return err
	}
	return nil
}

// setupInfrastructure 设置基础设施
// setupInfrastructure sets up infrastructure
func (p *RabbitMQProducer) setupInfrastructure() error {
	// 注意：这里假设交换机和队列已经在RabbitMQ服务器上配置好了
	// 在生产环境中，通常由运维或部署脚本来创建这些基础设施
	// Note: This assumes exchanges and queues are already configured on RabbitMQ server
	// In production, these are usually created by ops or deployment scripts

	p.logger.Info("RabbitMQ infrastructure setup completed",
		logger.String("order_exchange", p.config.OrderExchange),
		logger.String("stock_exchange", p.config.StockExchange),
		logger.String("email_exchange", p.config.EmailExchange))

	return nil
}
