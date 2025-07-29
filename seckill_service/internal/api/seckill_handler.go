package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flashsku/seckill/internal/activity"
	"github.com/flashsku/seckill/internal/cache"
	"github.com/flashsku/seckill/internal/seckill"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/gin-gonic/gin"
)

// SeckillHandler 秒杀API处理器
// SeckillHandler seckill API handler
type SeckillHandler struct {
	seckillService    *seckill.SeckillService
	activityValidator *activity.ActivityValidator
	metricsCollector  *cache.MetricsCollector
	logger            logger.Logger
}

// SeckillRequest 秒杀请求
// SeckillRequest seckill request
type SeckillRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	PurchaseAmount int    `json:"purchase_amount" binding:"required,min=1,max=100"`
	UserLimit      int    `json:"user_limit,omitempty"`
}

// SeckillResponse 秒杀响应
// SeckillResponse seckill response
type SeckillResponse struct {
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Data      *SeckillData `json:"data,omitempty"`
	ErrorCode string       `json:"error_code,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	RequestID string       `json:"request_id,omitempty"`
}

// SeckillData 秒杀数据
// SeckillData seckill data
type SeckillData struct {
	ActivityID     string `json:"activity_id"`
	UserID         string `json:"user_id"`
	PurchaseAmount int    `json:"purchase_amount"`
	RemainingStock int    `json:"remaining_stock"`
	UserPurchased  int    `json:"user_purchased"`
	RemainingLimit int    `json:"remaining_limit"`
	OrderID        string `json:"order_id,omitempty"`
}

// StockResponse 库存响应
// StockResponse stock response
type StockResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// EnhancedStockInfo 增强的库存信息
// EnhancedStockInfo enhanced stock information
type EnhancedStockInfo struct {
	ActivityID     string    `json:"activity_id"`
	CurrentStock   int       `json:"current_stock"`
	TotalStock     int       `json:"total_stock"`
	SoldCount      int       `json:"sold_count"`
	Status         string    `json:"status"`
	LastUpdated    time.Time `json:"last_updated"`
	ActivityStatus string    `json:"activity_status"`
	ActivityName   string    `json:"activity_name"`
}

// NewSeckillHandler 创建秒杀处理器
// NewSeckillHandler creates seckill handler
func NewSeckillHandler(
	seckillService *seckill.SeckillService,
	activityValidator *activity.ActivityValidator,
	metricsCollector *cache.MetricsCollector,
	log logger.Logger,
) *SeckillHandler {
	return &SeckillHandler{
		seckillService:    seckillService,
		activityValidator: activityValidator,
		metricsCollector:  metricsCollector,
		logger:            log,
	}
}

// RegisterRoutes 注册路由
// RegisterRoutes registers routes
func (h *SeckillHandler) RegisterRoutes(router *gin.RouterGroup) {
	seckillGroup := router.Group("/seckill")
	{
		// 秒杀接口
		// Seckill endpoints
		seckillGroup.POST("/:activity_id", h.ProcessSeckill)
		seckillGroup.GET("/stock/:activity_id", h.GetStock)
		seckillGroup.GET("/stocks", h.BatchGetStocks)
		seckillGroup.POST("/rollback/:activity_id", h.RollbackStock)

		// 用户相关接口
		// User related endpoints
		seckillGroup.GET("/user/:user_id/purchases", h.GetUserPurchases)
		seckillGroup.GET("/user/:user_id/limit/:activity_id", h.GetUserLimit)

		// 活动相关接口
		// Activity related endpoints
		seckillGroup.GET("/activity/:activity_id/info", h.GetActivityInfo)
		seckillGroup.GET("/activity/:activity_id/stats", h.GetActivityStats)
	}
}

// ProcessSeckill 处理秒杀请求
// ProcessSeckill processes seckill request
func (h *SeckillHandler) ProcessSeckill(c *gin.Context) {
	startTime := time.Now()
	activityID := c.Param("activity_id")
	requestID := c.GetHeader("X-Request-ID")

	// 记录请求日志
	// Log request
	h.logger.Info("Seckill request received",
		logger.String("activity_id", activityID),
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// 解析请求体
	// Parse request body
	var req SeckillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body",
			logger.String("activity_id", activityID),
			logger.Error(err))

		c.JSON(http.StatusBadRequest, SeckillResponse{
			Success:   false,
			Message:   "Invalid request parameters",
			ErrorCode: "INVALID_PARAMS",
			Timestamp: time.Now(),
			RequestID: requestID,
		})
		return
	}

	// 构建服务请求
	// Build service request
	serviceReq := &seckill.SeckillRequest{
		ActivityID:     activityID,
		UserID:         req.UserID,
		PurchaseAmount: req.PurchaseAmount,
		UserLimit:      req.UserLimit,
	}

	// 调用秒杀服务
	// Call seckill service
	result, err := h.seckillService.ProcessSeckill(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Seckill service error",
			logger.String("activity_id", activityID),
			logger.String("user_id", req.UserID),
			logger.Error(err))

		c.JSON(http.StatusInternalServerError, SeckillResponse{
			Success:   false,
			Message:   "Internal server error",
			ErrorCode: "INTERNAL_ERROR",
			Timestamp: time.Now(),
			RequestID: requestID,
		})
		return
	}

	// 构建响应
	// Build response
	response := SeckillResponse{
		Success:   result.Success,
		Message:   result.Message,
		ErrorCode: result.ErrorCode,
		Timestamp: time.Now(),
		RequestID: requestID,
	}

	if result.Success {
		response.Data = &SeckillData{
			ActivityID:     result.ActivityID,
			UserID:         result.UserID,
			PurchaseAmount: result.PurchaseAmount,
			RemainingStock: result.RemainingStock,
			UserPurchased:  result.UserPurchased,
			RemainingLimit: result.RemainingLimit,
			OrderID:        h.generateOrderID(activityID, req.UserID),
		}
	}

	// 根据结果设置HTTP状态码
	// Set HTTP status code based on result
	statusCode := http.StatusOK
	if !result.Success {
		switch result.ErrorCode {
		case "INVALID_PARAMS":
			statusCode = http.StatusBadRequest
		case "ACTIVITY_INACTIVE":
			statusCode = http.StatusForbidden
		case "INSUFFICIENT_STOCK", "EXCEEDS_USER_LIMIT":
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusInternalServerError
		}
	}

	// 记录响应日志
	// Log response
	duration := time.Since(startTime)
	h.logger.Info("Seckill request completed",
		logger.String("activity_id", activityID),
		logger.String("user_id", req.UserID),
		logger.Bool("success", result.Success),
		logger.String("error_code", result.ErrorCode),
		logger.Duration("duration", duration))

	c.JSON(statusCode, response)
}

// GetStock 获取库存信息
// GetStock gets stock information
func (h *SeckillHandler) GetStock(c *gin.Context) {
	startTime := time.Now()
	activityID := c.Param("activity_id")
	requestID := c.GetHeader("X-Request-ID")

	// 记录API调用指标
	// Record API call metrics
	defer func() {
		duration := time.Since(startTime)
		if h.metricsCollector != nil {
			h.metricsCollector.RecordHit(duration)
		}
	}()

	// 记录请求日志
	// Log request
	h.logger.Info("Get stock request received",
		logger.String("activity_id", activityID),
		logger.String("request_id", requestID),
		logger.String("client_ip", c.ClientIP()))

	// 参数验证
	// Parameter validation
	if activityID == "" {
		h.logger.Warn("Empty activity ID",
			logger.String("request_id", requestID))

		c.JSON(http.StatusBadRequest, StockResponse{
			Success:   false,
			Message:   "Activity ID is required",
			Timestamp: time.Now(),
		})
		return
	}

	// 活动验证
	// Activity validation
	validationResult, err := h.activityValidator.ValidateActivity(c.Request.Context(), activityID)
	if err != nil {
		h.logger.Error("Failed to validate activity",
			logger.String("activity_id", activityID),
			logger.String("request_id", requestID),
			logger.Error(err))

		c.JSON(http.StatusInternalServerError, StockResponse{
			Success:   false,
			Message:   "Failed to validate activity",
			Timestamp: time.Now(),
		})
		return
	}

	// 检查活动是否有效
	// Check if activity is valid
	if !validationResult.Valid {
		h.logger.Warn("Activity validation failed",
			logger.String("activity_id", activityID),
			logger.String("reason", validationResult.Reason),
			logger.String("error_code", validationResult.ErrorCode))

		statusCode := http.StatusBadRequest
		switch validationResult.ErrorCode {
		case "ACTIVITY_NOT_FOUND":
			statusCode = http.StatusNotFound
		case "ACTIVITY_NOT_ACTIVE", "ACTIVITY_NOT_STARTED", "ACTIVITY_ENDED":
			statusCode = http.StatusForbidden
		case "OUT_OF_STOCK":
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, StockResponse{
			Success:   false,
			Message:   validationResult.Reason,
			Timestamp: time.Now(),
		})
		return
	}

	// 获取库存信息
	// Get stock information
	stockInfo, err := h.seckillService.GetStockInfo(c.Request.Context(), activityID)
	if err != nil {
		h.logger.Error("Failed to get stock info",
			logger.String("activity_id", activityID),
			logger.String("request_id", requestID),
			logger.Error(err))

		c.JSON(http.StatusInternalServerError, StockResponse{
			Success:   false,
			Message:   "Failed to get stock information",
			Timestamp: time.Now(),
		})
		return
	}

	// 构建增强的库存信息
	// Build enhanced stock information
	enhancedStockInfo := &EnhancedStockInfo{
		ActivityID:     stockInfo.ActivityID,
		CurrentStock:   stockInfo.CurrentStock,
		Status:         stockInfo.Status,
		LastUpdated:    stockInfo.LastUpdated,
		ActivityStatus: validationResult.ActivityInfo.Status,
		ActivityName:   validationResult.ActivityInfo.Name,
		TotalStock:     validationResult.ActivityInfo.TotalStock,
		SoldCount:      validationResult.ActivityInfo.SoldCount,
	}

	// 记录成功日志
	// Log success
	duration := time.Since(startTime)
	h.logger.Info("Get stock request completed",
		logger.String("activity_id", activityID),
		logger.String("request_id", requestID),
		logger.Int("current_stock", stockInfo.CurrentStock),
		logger.String("status", stockInfo.Status),
		logger.Duration("duration", duration))

	c.JSON(http.StatusOK, StockResponse{
		Success:   true,
		Message:   "Stock information retrieved successfully",
		Data:      enhancedStockInfo,
		Timestamp: time.Now(),
	})
}

// BatchGetStocks 批量获取库存信息
// BatchGetStocks batch gets stock information
func (h *SeckillHandler) BatchGetStocks(c *gin.Context) {
	activityIDsParam := c.Query("activity_ids")
	if activityIDsParam == "" {
		c.JSON(http.StatusBadRequest, StockResponse{
			Success:   false,
			Message:   "activity_ids parameter is required",
			Timestamp: time.Now(),
		})
		return
	}

	// 解析活动ID列表
	// Parse activity ID list
	activityIDs := parseActivityIDs(activityIDsParam)
	if len(activityIDs) == 0 {
		c.JSON(http.StatusBadRequest, StockResponse{
			Success:   false,
			Message:   "Invalid activity_ids format",
			Timestamp: time.Now(),
		})
		return
	}

	// 限制批量查询数量
	// Limit batch query count
	if len(activityIDs) > 50 {
		c.JSON(http.StatusBadRequest, StockResponse{
			Success:   false,
			Message:   "Too many activity IDs (max 50)",
			Timestamp: time.Now(),
		})
		return
	}

	stockInfos, err := h.seckillService.BatchCheckStock(c.Request.Context(), activityIDs)
	if err != nil {
		h.logger.Error("Failed to batch get stock info",
			logger.String("activity_ids", strings.Join(activityIDs, ",")),
			logger.Error(err))

		c.JSON(http.StatusInternalServerError, StockResponse{
			Success:   false,
			Message:   "Failed to get stock information",
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, StockResponse{
		Success:   true,
		Message:   "Stock information retrieved successfully",
		Data:      stockInfos,
		Timestamp: time.Now(),
	})
}

// RollbackStock 回滚库存
// RollbackStock rollbacks stock
func (h *SeckillHandler) RollbackStock(c *gin.Context) {
	activityID := c.Param("activity_id")

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Amount int    `json:"amount" binding:"required,min=1"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StockResponse{
			Success:   false,
			Message:   "Invalid request parameters",
			Timestamp: time.Now(),
		})
		return
	}

	err := h.seckillService.RollbackStock(c.Request.Context(), activityID, req.UserID, req.Amount)
	if err != nil {
		h.logger.Error("Failed to rollback stock",
			logger.String("activity_id", activityID),
			logger.String("user_id", req.UserID),
			logger.Int("amount", req.Amount),
			logger.Error(err))

		c.JSON(http.StatusInternalServerError, StockResponse{
			Success:   false,
			Message:   "Failed to rollback stock",
			Timestamp: time.Now(),
		})
		return
	}

	h.logger.Info("Stock rollback completed",
		logger.String("activity_id", activityID),
		logger.String("user_id", req.UserID),
		logger.Int("amount", req.Amount),
		logger.String("reason", req.Reason))

	c.JSON(http.StatusOK, StockResponse{
		Success:   true,
		Message:   "Stock rollback completed successfully",
		Timestamp: time.Now(),
	})
}

// GetUserPurchases 获取用户购买记录
// GetUserPurchases gets user purchase records
func (h *SeckillHandler) GetUserPurchases(c *gin.Context) {
	userID := c.Param("user_id")

	// 这里应该从数据库或缓存中获取用户购买记录
	// Should get user purchase records from database or cache
	h.logger.Info("Get user purchases requested",
		logger.String("user_id", userID))

	// 暂时返回空数据
	// Return empty data for now
	c.JSON(http.StatusOK, StockResponse{
		Success:   true,
		Message:   "User purchases retrieved successfully",
		Data:      []interface{}{},
		Timestamp: time.Now(),
	})
}

// GetUserLimit 获取用户限购信息
// GetUserLimit gets user purchase limit information
func (h *SeckillHandler) GetUserLimit(c *gin.Context) {
	userID := c.Param("user_id")
	activityID := c.Param("activity_id")

	// 这里应该从Redis中获取用户限购信息
	// Should get user limit information from Redis
	h.logger.Info("Get user limit requested",
		logger.String("user_id", userID),
		logger.String("activity_id", activityID))

	// 暂时返回模拟数据
	// Return mock data for now
	c.JSON(http.StatusOK, StockResponse{
		Success: true,
		Message: "User limit retrieved successfully",
		Data: map[string]interface{}{
			"user_id":     userID,
			"activity_id": activityID,
			"purchased":   0,
			"limit":       5,
			"remaining":   5,
		},
		Timestamp: time.Now(),
	})
}

// GetActivityInfo 获取活动信息
// GetActivityInfo gets activity information
func (h *SeckillHandler) GetActivityInfo(c *gin.Context) {
	activityID := c.Param("activity_id")

	// 这里应该从缓存或数据库中获取活动信息
	// Should get activity information from cache or database
	h.logger.Info("Get activity info requested",
		logger.String("activity_id", activityID))

	// 暂时返回模拟数据
	// Return mock data for now
	c.JSON(http.StatusOK, StockResponse{
		Success: true,
		Message: "Activity information retrieved successfully",
		Data: map[string]interface{}{
			"activity_id": activityID,
			"name":        "Test Activity",
			"status":      "active",
		},
		Timestamp: time.Now(),
	})
}

// GetActivityStats 获取活动统计
// GetActivityStats gets activity statistics
func (h *SeckillHandler) GetActivityStats(c *gin.Context) {
	activityID := c.Param("activity_id")

	// 这里应该从缓存中获取活动统计信息
	// Should get activity statistics from cache
	h.logger.Info("Get activity stats requested",
		logger.String("activity_id", activityID))

	// 暂时返回模拟数据
	// Return mock data for now
	c.JSON(http.StatusOK, StockResponse{
		Success: true,
		Message: "Activity statistics retrieved successfully",
		Data: map[string]interface{}{
			"activity_id":    activityID,
			"total_requests": 1000,
			"success_count":  800,
			"failure_count":  200,
			"success_rate":   0.8,
		},
		Timestamp: time.Now(),
	})
}

// generateOrderID 生成订单ID
// generateOrderID generates order ID
func (h *SeckillHandler) generateOrderID(activityID, userID string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("SK%s%s%d", activityID, userID, timestamp)
}

// parseActivityIDs 解析活动ID列表
// parseActivityIDs parses activity ID list
func parseActivityIDs(param string) []string {
	// 简单的逗号分隔解析
	// Simple comma-separated parsing
	if param == "" {
		return nil
	}

	// 按逗号分隔
	// Split by comma
	ids := strings.Split(param, ",")
	result := make([]string, 0, len(ids))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			result = append(result, id)
		}
	}

	return result
}
