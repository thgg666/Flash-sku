package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// ActivityValidator 活动验证器
// ActivityValidator activity validator
type ActivityValidator struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *ValidatorConfig
}

// ValidatorConfig 验证器配置
// ValidatorConfig validator configuration
type ValidatorConfig struct {
	CacheTimeout    time.Duration `json:"cache_timeout"`
	EnableTimeCheck bool          `json:"enable_time_check"`
	EnableStockCheck bool         `json:"enable_stock_check"`
	EnableStatusCheck bool        `json:"enable_status_check"`
	TimeBuffer       time.Duration `json:"time_buffer"`
}

// ActivityInfo 活动信息
// ActivityInfo activity information
type ActivityInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TotalStock  int       `json:"total_stock"`
	SoldCount   int       `json:"sold_count"`
	Price       float64   `json:"price"`
	UserLimit   int       `json:"user_limit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ValidationResult 验证结果
// ValidationResult validation result
type ValidationResult struct {
	Valid       bool      `json:"valid"`
	Reason      string    `json:"reason"`
	ErrorCode   string    `json:"error_code"`
	ActivityInfo *ActivityInfo `json:"activity_info,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ActivityStatus 活动状态枚举
// ActivityStatus activity status enum
const (
	StatusDraft     = "draft"     // 草稿
	StatusScheduled = "scheduled" // 已安排
	StatusActive    = "active"    // 进行中
	StatusPaused    = "paused"    // 暂停
	StatusEnded     = "ended"     // 已结束
	StatusCancelled = "cancelled" // 已取消
)

// NewActivityValidator 创建活动验证器
// NewActivityValidator creates activity validator
func NewActivityValidator(redisClient redis.Client, config *ValidatorConfig, log logger.Logger) *ActivityValidator {
	if config == nil {
		config = DefaultValidatorConfig()
	}

	return &ActivityValidator{
		redisClient: redisClient,
		logger:      log,
		config:      config,
	}
}

// ValidateActivity 验证活动
// ValidateActivity validates activity
func (v *ActivityValidator) ValidateActivity(ctx context.Context, activityID string) (*ValidationResult, error) {
	startTime := time.Now()
	
	result := &ValidationResult{
		Timestamp: startTime,
	}

	// 获取活动信息
	// Get activity information
	activityInfo, err := v.getActivityInfo(ctx, activityID)
	if err != nil {
		v.logger.Error("Failed to get activity info",
			logger.String("activity_id", activityID),
			logger.Error(err))
		
		result.Valid = false
		result.Reason = "Activity not found"
		result.ErrorCode = "ACTIVITY_NOT_FOUND"
		return result, nil
	}

	result.ActivityInfo = activityInfo

	// 验证活动状态
	// Validate activity status
	if v.config.EnableStatusCheck {
		if !v.isValidStatus(activityInfo.Status) {
			result.Valid = false
			result.Reason = fmt.Sprintf("Invalid activity status: %s", activityInfo.Status)
			result.ErrorCode = "INVALID_STATUS"
			return result, nil
		}

		if activityInfo.Status != StatusActive {
			result.Valid = false
			result.Reason = fmt.Sprintf("Activity is not active, current status: %s", activityInfo.Status)
			result.ErrorCode = "ACTIVITY_NOT_ACTIVE"
			return result, nil
		}
	}

	// 验证活动时间
	// Validate activity time
	if v.config.EnableTimeCheck {
		timeValidation := v.validateTime(activityInfo, startTime)
		if !timeValidation.Valid {
			result.Valid = false
			result.Reason = timeValidation.Reason
			result.ErrorCode = timeValidation.ErrorCode
			return result, nil
		}
	}

	// 验证库存
	// Validate stock
	if v.config.EnableStockCheck {
		stockValidation := v.validateStock(ctx, activityInfo)
		if !stockValidation.Valid {
			result.Valid = false
			result.Reason = stockValidation.Reason
			result.ErrorCode = stockValidation.ErrorCode
			return result, nil
		}
	}

	// 验证通过
	// Validation passed
	result.Valid = true
	result.Reason = "Activity validation passed"
	result.ErrorCode = ""

	v.logger.Debug("Activity validation completed",
		logger.String("activity_id", activityID),
		logger.Bool("valid", result.Valid),
		logger.String("reason", result.Reason),
		logger.Duration("duration", time.Since(startTime)))

	return result, nil
}

// GetActivityInfo 获取活动信息
// GetActivityInfo gets activity information
func (v *ActivityValidator) GetActivityInfo(ctx context.Context, activityID string) (*ActivityInfo, error) {
	return v.getActivityInfo(ctx, activityID)
}

// IsActivityActive 检查活动是否激活
// IsActivityActive checks if activity is active
func (v *ActivityValidator) IsActivityActive(ctx context.Context, activityID string) (bool, error) {
	result, err := v.ValidateActivity(ctx, activityID)
	if err != nil {
		return false, err
	}
	
	return result.Valid, nil
}

// GetActivityStatus 获取活动状态
// GetActivityStatus gets activity status
func (v *ActivityValidator) GetActivityStatus(ctx context.Context, activityID string) (string, error) {
	activityInfo, err := v.getActivityInfo(ctx, activityID)
	if err != nil {
		return "", err
	}
	
	return activityInfo.Status, nil
}

// getActivityInfo 获取活动信息
// getActivityInfo gets activity information
func (v *ActivityValidator) getActivityInfo(ctx context.Context, activityID string) (*ActivityInfo, error) {
	// 从Redis缓存获取
	// Get from Redis cache
	cacheKey := v.buildActivityCacheKey(activityID)
	cachedData, err := v.redisClient.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var activityInfo ActivityInfo
		if err := json.Unmarshal([]byte(cachedData), &activityInfo); err == nil {
			v.logger.Debug("Activity info loaded from cache",
				logger.String("activity_id", activityID))
			return &activityInfo, nil
		}
	}

	// 缓存未命中，从数据库获取（这里模拟）
	// Cache miss, get from database (simulated here)
	activityInfo, err := v.loadActivityFromDatabase(ctx, activityID)
	if err != nil {
		return nil, err
	}

	// 缓存活动信息
	// Cache activity information
	if err := v.cacheActivityInfo(ctx, activityInfo); err != nil {
		v.logger.Warn("Failed to cache activity info",
			logger.String("activity_id", activityID),
			logger.Error(err))
	}

	return activityInfo, nil
}

// loadActivityFromDatabase 从数据库加载活动信息
// loadActivityFromDatabase loads activity information from database
func (v *ActivityValidator) loadActivityFromDatabase(ctx context.Context, activityID string) (*ActivityInfo, error) {
	// 这里应该从实际数据库加载，现在返回模拟数据
	// Should load from actual database, returning mock data for now
	
	// 检查活动是否存在
	// Check if activity exists
	if activityID == "" {
		return nil, fmt.Errorf("activity ID is empty")
	}

	// 模拟数据库查询
	// Simulate database query
	now := time.Now()
	activityInfo := &ActivityInfo{
		ID:          activityID,
		Name:        fmt.Sprintf("Activity %s", activityID),
		Description: fmt.Sprintf("Description for activity %s", activityID),
		Status:      StatusActive,
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now.Add(1 * time.Hour),
		TotalStock:  1000,
		SoldCount:   0,
		Price:       99.99,
		UserLimit:   5,
		CreatedAt:   now.Add(-24 * time.Hour),
		UpdatedAt:   now,
	}

	v.logger.Debug("Activity info loaded from database",
		logger.String("activity_id", activityID))

	return activityInfo, nil
}

// cacheActivityInfo 缓存活动信息
// cacheActivityInfo caches activity information
func (v *ActivityValidator) cacheActivityInfo(ctx context.Context, activityInfo *ActivityInfo) error {
	cacheKey := v.buildActivityCacheKey(activityInfo.ID)
	
	data, err := json.Marshal(activityInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal activity info: %w", err)
	}

	return v.redisClient.Set(ctx, cacheKey, string(data), v.config.CacheTimeout)
}

// validateTime 验证活动时间
// validateTime validates activity time
func (v *ActivityValidator) validateTime(activityInfo *ActivityInfo, currentTime time.Time) *ValidationResult {
	// 添加时间缓冲
	// Add time buffer
	bufferedCurrentTime := currentTime.Add(-v.config.TimeBuffer)
	bufferedEndTime := activityInfo.EndTime.Add(v.config.TimeBuffer)

	// 检查活动是否已开始
	// Check if activity has started
	if bufferedCurrentTime.Before(activityInfo.StartTime) {
		return &ValidationResult{
			Valid:     false,
			Reason:    "Activity has not started yet",
			ErrorCode: "ACTIVITY_NOT_STARTED",
			Timestamp: currentTime,
		}
	}

	// 检查活动是否已结束
	// Check if activity has ended
	if currentTime.After(bufferedEndTime) {
		return &ValidationResult{
			Valid:     false,
			Reason:    "Activity has ended",
			ErrorCode: "ACTIVITY_ENDED",
			Timestamp: currentTime,
		}
	}

	return &ValidationResult{
		Valid:     true,
		Reason:    "Activity time is valid",
		Timestamp: currentTime,
	}
}

// validateStock 验证库存
// validateStock validates stock
func (v *ActivityValidator) validateStock(ctx context.Context, activityInfo *ActivityInfo) *ValidationResult {
	// 从Redis获取当前库存
	// Get current stock from Redis
	stockKey := v.buildStockKey(activityInfo.ID)
	stockStr, err := v.redisClient.Get(ctx, stockKey)
	if err != nil {
		// 如果Redis中没有库存信息，使用活动信息中的库存
		// If no stock info in Redis, use stock from activity info
		currentStock := activityInfo.TotalStock - activityInfo.SoldCount
		if currentStock <= 0 {
			return &ValidationResult{
				Valid:     false,
				Reason:    "No stock available",
				ErrorCode: "OUT_OF_STOCK",
				Timestamp: time.Now(),
			}
		}
		return &ValidationResult{
			Valid:     true,
			Reason:    "Stock is available",
			Timestamp: time.Now(),
		}
	}

	currentStock, err := strconv.Atoi(stockStr)
	if err != nil {
		return &ValidationResult{
			Valid:     false,
			Reason:    "Invalid stock data",
			ErrorCode: "INVALID_STOCK_DATA",
			Timestamp: time.Now(),
		}
	}

	if currentStock <= 0 {
		return &ValidationResult{
			Valid:     false,
			Reason:    "No stock available",
			ErrorCode: "OUT_OF_STOCK",
			Timestamp: time.Now(),
		}
	}

	return &ValidationResult{
		Valid:     true,
		Reason:    "Stock is available",
		Timestamp: time.Now(),
	}
}

// isValidStatus 检查状态是否有效
// isValidStatus checks if status is valid
func (v *ActivityValidator) isValidStatus(status string) bool {
	validStatuses := []string{
		StatusDraft,
		StatusScheduled,
		StatusActive,
		StatusPaused,
		StatusEnded,
		StatusCancelled,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}

	return false
}

// buildActivityCacheKey 构建活动缓存键
// buildActivityCacheKey builds activity cache key
func (v *ActivityValidator) buildActivityCacheKey(activityID string) string {
	return fmt.Sprintf("seckill:activity:%s", activityID)
}

// buildStockKey 构建库存键
// buildStockKey builds stock key
func (v *ActivityValidator) buildStockKey(activityID string) string {
	return fmt.Sprintf("seckill:stock:%s", activityID)
}

// DefaultValidatorConfig 默认验证器配置
// DefaultValidatorConfig default validator configuration
func DefaultValidatorConfig() *ValidatorConfig {
	return &ValidatorConfig{
		CacheTimeout:     5 * time.Minute,
		EnableTimeCheck:  true,
		EnableStockCheck: true,
		EnableStatusCheck: true,
		TimeBuffer:       30 * time.Second,
	}
}
