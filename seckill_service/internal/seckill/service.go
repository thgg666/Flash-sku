package seckill

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flashsku/seckill/internal/lua"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// SeckillService 秒杀服务
// SeckillService seckill service
type SeckillService struct {
	redisClient        redis.Client
	scriptManager      *lua.ScriptManager
	messageProducer    MessageProducer
	reliabilityManager ReliabilityManager
	logger             logger.Logger
	config             *SeckillConfig
}

// MessageProducer 消息生产者接口
// MessageProducer message producer interface
type MessageProducer interface {
	SendOrderMessage(ctx context.Context, orderMsg *OrderMessage) error
	SendStockSyncMessage(ctx context.Context, stockMsg *StockSyncMessage) error
	SendEmailMessage(ctx context.Context, emailMsg *EmailMessage) error
}

// ReliabilityManager 可靠性管理器接口
// ReliabilityManager reliability manager interface
type ReliabilityManager interface {
	SendReliableMessage(ctx context.Context, msg *ReliableMessage) error
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
}

// StockSyncMessage 库存同步消息
// StockSyncMessage stock sync message
type StockSyncMessage struct {
	ActivityID   string    `json:"activity_id"`
	ProductID    string    `json:"product_id"`
	StockChange  int       `json:"stock_change"`
	CurrentStock int       `json:"current_stock"`
	Operation    string    `json:"operation"`
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
}

// EmailMessage 邮件消息
// EmailMessage email message
type EmailMessage struct {
	To        []string               `json:"to"`
	Subject   string                 `json:"subject"`
	Template  string                 `json:"template"`
	Data      map[string]interface{} `json:"data"`
	Priority  int                    `json:"priority"`
	Timestamp time.Time              `json:"timestamp"`
}

// ReliableMessage 可靠消息
// ReliableMessage reliable message
type ReliableMessage struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Exchange   string                 `json:"exchange"`
	RoutingKey string                 `json:"routing_key"`
	Payload    map[string]interface{} `json:"payload"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	NextRetry  time.Time              `json:"next_retry"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SeckillConfig 秒杀配置
// SeckillConfig seckill configuration
type SeckillConfig struct {
	DefaultUserLimit  int           `json:"default_user_limit"`
	LimitTTL          time.Duration `json:"limit_ttl"`
	StockMinThreshold int           `json:"stock_min_threshold"`
	EnableRollback    bool          `json:"enable_rollback"`
}

// SeckillRequest 秒杀请求
// SeckillRequest seckill request
type SeckillRequest struct {
	ActivityID     string `json:"activity_id"`
	UserID         string `json:"user_id"`
	PurchaseAmount int    `json:"purchase_amount"`
	UserLimit      int    `json:"user_limit,omitempty"`
}

// SeckillResult 秒杀结果
// SeckillResult seckill result
type SeckillResult struct {
	Success        bool      `json:"success"`
	Message        string    `json:"message"`
	ActivityID     string    `json:"activity_id"`
	UserID         string    `json:"user_id"`
	PurchaseAmount int       `json:"purchase_amount"`
	RemainingStock int       `json:"remaining_stock"`
	UserPurchased  int       `json:"user_purchased"`
	RemainingLimit int       `json:"remaining_limit"`
	Timestamp      time.Time `json:"timestamp"`
	ErrorCode      string    `json:"error_code,omitempty"`
}

// StockInfo 库存信息
// StockInfo stock information
type StockInfo struct {
	ActivityID   string    `json:"activity_id"`
	CurrentStock int       `json:"current_stock"`
	Status       string    `json:"status"`
	LastUpdated  time.Time `json:"last_updated"`
}

// NewSeckillService 创建秒杀服务
// NewSeckillService creates seckill service
func NewSeckillService(
	redisClient redis.Client,
	config *SeckillConfig,
	messageProducer MessageProducer,
	reliabilityManager ReliabilityManager,
	log logger.Logger,
) *SeckillService {
	if config == nil {
		config = DefaultSeckillConfig()
	}

	scriptManager := lua.NewScriptManager(redisClient, log)

	return &SeckillService{
		redisClient:        redisClient,
		scriptManager:      scriptManager,
		messageProducer:    messageProducer,
		reliabilityManager: reliabilityManager,
		logger:             log,
		config:             config,
	}
}

// Initialize 初始化服务
// Initialize initializes service
func (s *SeckillService) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing seckill service")

	// 加载所有Lua脚本
	// Load all Lua scripts
	if err := s.scriptManager.LoadAllScripts(ctx); err != nil {
		return fmt.Errorf("failed to load scripts: %w", err)
	}

	s.logger.Info("Seckill service initialized successfully")
	return nil
}

// ProcessSeckill 处理秒杀请求
// ProcessSeckill processes seckill request
func (s *SeckillService) ProcessSeckill(ctx context.Context, req *SeckillRequest) (*SeckillResult, error) {
	startTime := time.Now()

	result := &SeckillResult{
		ActivityID:     req.ActivityID,
		UserID:         req.UserID,
		PurchaseAmount: req.PurchaseAmount,
		Timestamp:      startTime,
	}

	// 参数验证
	// Parameter validation
	if err := s.validateRequest(req); err != nil {
		result.Success = false
		result.Message = err.Error()
		result.ErrorCode = "INVALID_PARAMS"
		return result, err
	}

	// 构建Redis键
	// Build Redis keys
	keys := s.buildRedisKeys(req.ActivityID, req.UserID)

	// 设置用户限购数量
	// Set user purchase limit
	userLimit := req.UserLimit
	if userLimit <= 0 {
		userLimit = s.config.DefaultUserLimit
	}

	// 构建脚本参数
	// Build script arguments
	args := []interface{}{
		req.PurchaseAmount,
		userLimit,
		startTime.Unix(),
		int(s.config.LimitTTL.Seconds()),
	}

	// 执行秒杀脚本
	// Execute seckill script
	scriptResult, err := s.scriptManager.ExecuteScript(ctx, "seckill_process", keys, args)
	if err != nil {
		s.logger.Error("Failed to execute seckill script",
			logger.String("activity_id", req.ActivityID),
			logger.String("user_id", req.UserID),
			logger.Error(err))

		result.Success = false
		result.Message = "Internal error"
		result.ErrorCode = "SCRIPT_ERROR"
		return result, err
	}

	// 解析脚本结果
	// Parse script result
	return s.parseScriptResult(result, scriptResult)
}

// GetStockInfo 获取库存信息
// GetStockInfo gets stock information
func (s *SeckillService) GetStockInfo(ctx context.Context, activityID string) (*StockInfo, error) {
	stockKey := fmt.Sprintf("seckill:stock:%s", activityID)

	stockStr, err := s.redisClient.Get(ctx, stockKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	stock, err := strconv.Atoi(stockStr)
	if err != nil {
		return nil, fmt.Errorf("invalid stock value: %w", err)
	}

	status := "normal"
	if stock == 0 {
		status = "out_of_stock"
	} else if stock <= s.config.StockMinThreshold {
		status = "low_stock"
	}

	return &StockInfo{
		ActivityID:   activityID,
		CurrentStock: stock,
		Status:       status,
		LastUpdated:  time.Now(),
	}, nil
}

// BatchCheckStock 批量检查库存
// BatchCheckStock batch checks stock
func (s *SeckillService) BatchCheckStock(ctx context.Context, activityIDs []string) (map[string]*StockInfo, error) {
	if len(activityIDs) == 0 {
		return make(map[string]*StockInfo), nil
	}

	// 构建库存键
	// Build stock keys
	keys := make([]string, len(activityIDs))
	for i, activityID := range activityIDs {
		keys[i] = fmt.Sprintf("seckill:stock:%s", activityID)
	}

	// 执行批量检查脚本
	// Execute batch check script
	args := []interface{}{s.config.StockMinThreshold}
	scriptResult, err := s.scriptManager.ExecuteScript(ctx, "batch_stock_check", keys, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute batch check script: %w", err)
	}

	// 解析结果
	// Parse results
	return s.parseBatchStockResult(activityIDs, scriptResult)
}

// RollbackStock 回滚库存
// RollbackStock rollbacks stock
func (s *SeckillService) RollbackStock(ctx context.Context, activityID, userID string, amount int) error {
	if !s.config.EnableRollback {
		return fmt.Errorf("rollback is disabled")
	}

	stockKey := fmt.Sprintf("seckill:stock:%s", activityID)
	userLimitKey := fmt.Sprintf("seckill:user_limit:%s:%s", userID, activityID)

	keys := []string{stockKey, userLimitKey}
	args := []interface{}{amount}

	scriptResult, err := s.scriptManager.ExecuteScript(ctx, "stock_rollback", keys, args)
	if err != nil {
		return fmt.Errorf("failed to execute rollback script: %w", err)
	}

	if !scriptResult.Success {
		return fmt.Errorf("rollback failed: %s", scriptResult.Error)
	}

	s.logger.Info("Stock rollback completed",
		logger.String("activity_id", activityID),
		logger.String("user_id", userID),
		logger.Int("amount", amount))

	return nil
}

// validateRequest 验证请求
// validateRequest validates request
func (s *SeckillService) validateRequest(req *SeckillRequest) error {
	if req.ActivityID == "" {
		return fmt.Errorf("activity_id is required")
	}
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.PurchaseAmount <= 0 {
		return fmt.Errorf("purchase_amount must be positive")
	}
	if req.PurchaseAmount > 100 { // 防止恶意大量购买
		return fmt.Errorf("purchase_amount too large")
	}
	return nil
}

// buildRedisKeys 构建Redis键
// buildRedisKeys builds Redis keys
func (s *SeckillService) buildRedisKeys(activityID, userID string) []string {
	return []string{
		fmt.Sprintf("seckill:activity:%s", activityID),              // 活动信息键
		fmt.Sprintf("seckill:stock:%s", activityID),                 // 库存键
		fmt.Sprintf("seckill:user_limit:%s:%s", userID, activityID), // 用户限购键
		fmt.Sprintf("seckill:status:%s", activityID),                // 活动状态键
	}
}

// parseScriptResult 解析脚本结果
// parseScriptResult parses script result
func (s *SeckillService) parseScriptResult(result *SeckillResult, scriptResult *lua.ScriptResult) (*SeckillResult, error) {
	if !scriptResult.Success {
		result.Success = false
		result.Message = scriptResult.Error
		result.ErrorCode = "SECKILL_FAILED"
		return result, nil
	}

	// 解析Lua脚本返回的结果
	// Parse Lua script return result
	luaResult, ok := scriptResult.Result.([]interface{})
	if !ok || len(luaResult) < 2 {
		result.Success = false
		result.Message = "Invalid script result"
		result.ErrorCode = "PARSE_ERROR"
		return result, nil
	}

	success, _ := luaResult[0].(int64)
	message, _ := luaResult[1].(string)

	if success == 1 {
		result.Success = true
		result.Message = message

		// 解析详细结果
		// Parse detailed result
		if len(luaResult) > 2 {
			if details, ok := luaResult[2].(map[string]interface{}); ok {
				if newStock, exists := details["new_stock"]; exists {
					if stock, ok := newStock.(int64); ok {
						result.RemainingStock = int(stock)
					}
				}
				if userPurchased, exists := details["user_purchased"]; exists {
					if purchased, ok := userPurchased.(int64); ok {
						result.UserPurchased = int(purchased)
					}
				}
				if remainingLimit, exists := details["remaining_limit"]; exists {
					if limit, ok := remainingLimit.(int64); ok {
						result.RemainingLimit = int(limit)
					}
				}
			}
		}

		// 秒杀成功，发送订单消息
		// Seckill successful, send order message
		go s.sendOrderMessage(result)

		// 发送库存同步消息
		// Send stock sync message
		go s.sendStockSyncMessage(result)
	} else {
		result.Success = false
		result.Message = message

		// 设置具体的错误码
		// Set specific error code
		switch message {
		case "activity not active":
			result.ErrorCode = "ACTIVITY_INACTIVE"
		case "insufficient stock":
			result.ErrorCode = "INSUFFICIENT_STOCK"
		case "exceeds user limit":
			result.ErrorCode = "EXCEEDS_USER_LIMIT"
		default:
			result.ErrorCode = "SECKILL_FAILED"
		}

		// 解析额外信息
		// Parse additional information
		if len(luaResult) > 2 {
			if currentStock, ok := luaResult[2].(int64); ok {
				result.RemainingStock = int(currentStock)
			}
		}
		if len(luaResult) > 3 {
			if userLimit, ok := luaResult[3].(int64); ok {
				result.RemainingLimit = int(userLimit)
			}
		}
	}

	return result, nil
}

// parseBatchStockResult 解析批量库存结果
// parseBatchStockResult parses batch stock result
func (s *SeckillService) parseBatchStockResult(activityIDs []string, scriptResult *lua.ScriptResult) (map[string]*StockInfo, error) {
	result := make(map[string]*StockInfo)

	if !scriptResult.Success {
		return result, fmt.Errorf("batch check failed: %s", scriptResult.Error)
	}

	// 解析Lua脚本返回的结果
	// Parse Lua script return result
	luaResult, ok := scriptResult.Result.([]interface{})
	if !ok || len(luaResult) < 3 {
		return result, fmt.Errorf("invalid batch check result")
	}

	results, ok := luaResult[2].([]interface{})
	if !ok {
		return result, fmt.Errorf("invalid batch check results format")
	}

	for i, item := range results {
		if i >= len(activityIDs) {
			break
		}

		activityID := activityIDs[i]
		if itemMap, ok := item.(map[string]interface{}); ok {
			stock := -1
			status := "unknown"

			if stockVal, exists := itemMap["stock"]; exists {
				if stockInt, ok := stockVal.(int64); ok {
					stock = int(stockInt)
				}
			}

			if statusVal, exists := itemMap["status"]; exists {
				if statusStr, ok := statusVal.(string); ok {
					status = statusStr
				}
			}

			result[activityID] = &StockInfo{
				ActivityID:   activityID,
				CurrentStock: stock,
				Status:       status,
				LastUpdated:  time.Now(),
			}
		}
	}

	return result, nil
}

// sendOrderMessage 发送订单消息
// sendOrderMessage sends order message
func (s *SeckillService) sendOrderMessage(result *SeckillResult) {
	if s.messageProducer == nil {
		s.logger.Warn("Message producer not configured, skipping order message")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 生成订单ID
	// Generate order ID
	orderID := fmt.Sprintf("order_%s_%s_%d", result.ActivityID, result.UserID, time.Now().UnixNano())

	orderMsg := &OrderMessage{
		OrderID:    orderID,
		UserID:     result.UserID,
		ActivityID: result.ActivityID,
		ProductID:  result.ActivityID, // 简化处理，使用ActivityID作为ProductID
		Quantity:   result.PurchaseAmount,
		Price:      0.0, // 这里应该从活动信息中获取价格
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	if err := s.messageProducer.SendOrderMessage(ctx, orderMsg); err != nil {
		s.logger.Error("Failed to send order message",
			logger.String("order_id", orderID),
			logger.String("user_id", result.UserID),
			logger.String("activity_id", result.ActivityID),
			logger.Error(err))

		// 如果有可靠性管理器，尝试可靠发送
		// If reliability manager available, try reliable send
		if s.reliabilityManager != nil {
			s.sendReliableOrderMessage(ctx, orderMsg)
		}
	} else {
		s.logger.Info("Order message sent successfully",
			logger.String("order_id", orderID),
			logger.String("user_id", result.UserID),
			logger.String("activity_id", result.ActivityID))
	}
}

// sendStockSyncMessage 发送库存同步消息
// sendStockSyncMessage sends stock sync message
func (s *SeckillService) sendStockSyncMessage(result *SeckillResult) {
	if s.messageProducer == nil {
		s.logger.Warn("Message producer not configured, skipping stock sync message")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stockMsg := &StockSyncMessage{
		ActivityID:   result.ActivityID,
		ProductID:    result.ActivityID,      // 简化处理
		StockChange:  -result.PurchaseAmount, // 负数表示减少
		CurrentStock: result.RemainingStock,
		Operation:    "decrease",
		Timestamp:    time.Now(),
		Source:       "seckill",
	}

	if err := s.messageProducer.SendStockSyncMessage(ctx, stockMsg); err != nil {
		s.logger.Error("Failed to send stock sync message",
			logger.String("activity_id", result.ActivityID),
			logger.Int("stock_change", stockMsg.StockChange),
			logger.Int("current_stock", stockMsg.CurrentStock),
			logger.Error(err))

		// 如果有可靠性管理器，尝试可靠发送
		// If reliability manager available, try reliable send
		if s.reliabilityManager != nil {
			s.sendReliableStockSyncMessage(ctx, stockMsg)
		}
	} else {
		s.logger.Info("Stock sync message sent successfully",
			logger.String("activity_id", result.ActivityID),
			logger.Int("stock_change", stockMsg.StockChange),
			logger.Int("current_stock", stockMsg.CurrentStock))
	}
}

// sendReliableOrderMessage 发送可靠订单消息
// sendReliableOrderMessage sends reliable order message
func (s *SeckillService) sendReliableOrderMessage(ctx context.Context, orderMsg *OrderMessage) {
	reliableMsg := &ReliableMessage{
		Type:       "order",
		Exchange:   "seckill.exchange",
		RoutingKey: "order.created",
		Payload: map[string]interface{}{
			"order_id":    orderMsg.OrderID,
			"user_id":     orderMsg.UserID,
			"activity_id": orderMsg.ActivityID,
			"product_id":  orderMsg.ProductID,
			"quantity":    orderMsg.Quantity,
			"price":       orderMsg.Price,
			"status":      orderMsg.Status,
			"created_at":  orderMsg.CreatedAt,
		},
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	if err := s.reliabilityManager.SendReliableMessage(ctx, reliableMsg); err != nil {
		s.logger.Error("Failed to send reliable order message",
			logger.String("order_id", orderMsg.OrderID),
			logger.Error(err))
	}
}

// sendReliableStockSyncMessage 发送可靠库存同步消息
// sendReliableStockSyncMessage sends reliable stock sync message
func (s *SeckillService) sendReliableStockSyncMessage(ctx context.Context, stockMsg *StockSyncMessage) {
	reliableMsg := &ReliableMessage{
		Type:       "stock",
		Exchange:   "seckill.exchange",
		RoutingKey: "stock.sync",
		Payload: map[string]interface{}{
			"activity_id":   stockMsg.ActivityID,
			"product_id":    stockMsg.ProductID,
			"stock_change":  stockMsg.StockChange,
			"current_stock": stockMsg.CurrentStock,
			"operation":     stockMsg.Operation,
			"timestamp":     stockMsg.Timestamp,
			"source":        stockMsg.Source,
		},
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	if err := s.reliabilityManager.SendReliableMessage(ctx, reliableMsg); err != nil {
		s.logger.Error("Failed to send reliable stock sync message",
			logger.String("activity_id", stockMsg.ActivityID),
			logger.Error(err))
	}
}

// DefaultSeckillConfig 默认秒杀配置
// DefaultSeckillConfig default seckill configuration
func DefaultSeckillConfig() *SeckillConfig {
	return &SeckillConfig{
		DefaultUserLimit:  5,
		LimitTTL:          24 * time.Hour,
		StockMinThreshold: 10,
		EnableRollback:    true,
	}
}
