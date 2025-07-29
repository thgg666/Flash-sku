package limit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flashsku/seckill/internal/lua"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// UserLimitService 用户限购服务
// UserLimitService user purchase limit service
type UserLimitService struct {
	redisClient   redis.Client
	scriptManager *lua.ScriptManager
	logger        logger.Logger
	config        *UserLimitConfig
}

// UserLimitConfig 用户限购配置
// UserLimitConfig user purchase limit configuration
type UserLimitConfig struct {
	DefaultLimit         int           `json:"default_limit"`
	GlobalLimit          int           `json:"global_limit"`
	DailyLimit           int           `json:"daily_limit"`
	LimitTTL             time.Duration `json:"limit_ttl"`
	DailyLimitTTL        time.Duration `json:"daily_limit_ttl"`
	EnableGlobalLimit    bool          `json:"enable_global_limit"`
	EnableDailyLimit     bool          `json:"enable_daily_limit"`
	EnableDuplicateCheck bool          `json:"enable_duplicate_check"`
}

// UserLimitInfo 用户限购信息
// UserLimitInfo user purchase limit information
type UserLimitInfo struct {
	UserID           string    `json:"user_id"`
	ActivityID       string    `json:"activity_id"`
	CurrentPurchased int       `json:"current_purchased"`
	ActivityLimit    int       `json:"activity_limit"`
	DailyPurchased   int       `json:"daily_purchased"`
	DailyLimit       int       `json:"daily_limit"`
	GlobalPurchased  int       `json:"global_purchased"`
	GlobalLimit      int       `json:"global_limit"`
	RemainingLimit   int       `json:"remaining_limit"`
	LastPurchaseTime time.Time `json:"last_purchase_time"`
	CanPurchase      bool      `json:"can_purchase"`
	LimitReason      string    `json:"limit_reason,omitempty"`
}

// CheckResult 检查结果
// CheckResult check result
type CheckResult struct {
	Allowed        bool      `json:"allowed"`
	Reason         string    `json:"reason"`
	CurrentCount   int       `json:"current_count"`
	Limit          int       `json:"limit"`
	RemainingCount int       `json:"remaining_count"`
	ResetTime      time.Time `json:"reset_time,omitempty"`
}

// PurchaseRecord 购买记录
// PurchaseRecord purchase record
type PurchaseRecord struct {
	UserID         string    `json:"user_id"`
	ActivityID     string    `json:"activity_id"`
	PurchaseAmount int       `json:"purchase_amount"`
	PurchaseTime   time.Time `json:"purchase_time"`
	OrderID        string    `json:"order_id"`
	Status         string    `json:"status"`
}

// NewUserLimitService 创建用户限购服务
// NewUserLimitService creates user limit service
func NewUserLimitService(redisClient redis.Client, config *UserLimitConfig, log logger.Logger) *UserLimitService {
	if config == nil {
		config = DefaultUserLimitConfig()
	}

	scriptManager := lua.NewScriptManager(redisClient, log)

	return &UserLimitService{
		redisClient:   redisClient,
		scriptManager: scriptManager,
		logger:        log,
		config:        config,
	}
}

// Initialize 初始化服务
// Initialize initializes service
func (s *UserLimitService) Initialize(ctx context.Context) error {
	s.logger.Info("Initializing user limit service")

	// 加载Lua脚本
	// Load Lua scripts
	if err := s.scriptManager.LoadAllScripts(ctx); err != nil {
		return fmt.Errorf("failed to load scripts: %w", err)
	}

	s.logger.Info("User limit service initialized successfully")
	return nil
}

// CheckUserLimit 检查用户限购
// CheckUserLimit checks user purchase limit
func (s *UserLimitService) CheckUserLimit(ctx context.Context, userID, activityID string, purchaseAmount int) (*CheckResult, error) {
	// 获取活动限购配置
	// Get activity limit configuration
	activityLimit, err := s.getActivityLimit(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity limit: %w", err)
	}

	// 构建Redis键
	// Build Redis keys
	limitKey := s.buildUserLimitKey(userID, activityID)

	// 使用Lua脚本检查限购
	// Use Lua script to check limit
	keys := []string{limitKey}
	args := []interface{}{
		purchaseAmount,
		activityLimit,
		int(s.config.LimitTTL.Seconds()),
	}

	scriptResult, err := s.scriptManager.ExecuteScript(ctx, "user_limit_check", keys, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute limit check script: %w", err)
	}

	return s.parseCheckResult(scriptResult, activityLimit)
}

// CheckDailyLimit 检查每日限购
// CheckDailyLimit checks daily purchase limit
func (s *UserLimitService) CheckDailyLimit(ctx context.Context, userID string, purchaseAmount int) (*CheckResult, error) {
	if !s.config.EnableDailyLimit {
		return &CheckResult{
			Allowed:        true,
			Reason:         "daily limit disabled",
			CurrentCount:   0,
			Limit:          s.config.DailyLimit,
			RemainingCount: s.config.DailyLimit,
		}, nil
	}

	// 获取当前每日购买数量
	// Get current daily purchase count
	currentCount, err := s.getDailyPurchaseCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily purchase count: %w", err)
	}

	// 检查是否超过每日限购
	// Check if exceeds daily limit
	if currentCount+purchaseAmount > s.config.DailyLimit {
		return &CheckResult{
			Allowed:        false,
			Reason:         "exceeds daily limit",
			CurrentCount:   currentCount,
			Limit:          s.config.DailyLimit,
			RemainingCount: s.config.DailyLimit - currentCount,
			ResetTime:      s.getNextDayResetTime(),
		}, nil
	}

	return &CheckResult{
		Allowed:        true,
		Reason:         "within daily limit",
		CurrentCount:   currentCount,
		Limit:          s.config.DailyLimit,
		RemainingCount: s.config.DailyLimit - currentCount - purchaseAmount,
	}, nil
}

// CheckGlobalLimit 检查全局限购
// CheckGlobalLimit checks global purchase limit
func (s *UserLimitService) CheckGlobalLimit(ctx context.Context, userID string, purchaseAmount int) (*CheckResult, error) {
	if !s.config.EnableGlobalLimit {
		return &CheckResult{
			Allowed:        true,
			Reason:         "global limit disabled",
			CurrentCount:   0,
			Limit:          s.config.GlobalLimit,
			RemainingCount: s.config.GlobalLimit,
		}, nil
	}

	// 获取用户全局购买数量
	// Get user global purchase count
	globalCount, err := s.getGlobalPurchaseCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get global purchase count: %w", err)
	}

	// 检查是否超过全局限购
	// Check if exceeds global limit
	if globalCount+purchaseAmount > s.config.GlobalLimit {
		return &CheckResult{
			Allowed:        false,
			Reason:         "exceeds global limit",
			CurrentCount:   globalCount,
			Limit:          s.config.GlobalLimit,
			RemainingCount: s.config.GlobalLimit - globalCount,
		}, nil
	}

	return &CheckResult{
		Allowed:        true,
		Reason:         "within global limit",
		CurrentCount:   globalCount,
		Limit:          s.config.GlobalLimit,
		RemainingCount: s.config.GlobalLimit - globalCount - purchaseAmount,
	}, nil
}

// CheckDuplicatePurchase 检查重复购买
// CheckDuplicatePurchase checks duplicate purchase
func (s *UserLimitService) CheckDuplicatePurchase(ctx context.Context, userID, activityID string) (*CheckResult, error) {
	if !s.config.EnableDuplicateCheck {
		return &CheckResult{
			Allowed: true,
			Reason:  "duplicate check disabled",
		}, nil
	}

	// 检查用户是否已经购买过该活动
	// Check if user has already purchased this activity
	hasPurchased, err := s.hasUserPurchased(ctx, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate purchase: %w", err)
	}

	if hasPurchased {
		return &CheckResult{
			Allowed: false,
			Reason:  "duplicate purchase not allowed",
		}, nil
	}

	return &CheckResult{
		Allowed: true,
		Reason:  "no duplicate purchase",
	}, nil
}

// GetUserLimitInfo 获取用户限购信息
// GetUserLimitInfo gets user purchase limit information
func (s *UserLimitService) GetUserLimitInfo(ctx context.Context, userID, activityID string) (*UserLimitInfo, error) {
	info := &UserLimitInfo{
		UserID:     userID,
		ActivityID: activityID,
	}

	// 获取活动限购
	// Get activity limit
	activityLimit, err := s.getActivityLimit(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity limit: %w", err)
	}
	info.ActivityLimit = activityLimit

	// 获取当前购买数量
	// Get current purchase count
	currentPurchased, err := s.getCurrentPurchaseCount(ctx, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current purchase count: %w", err)
	}
	info.CurrentPurchased = currentPurchased

	// 获取每日购买数量
	// Get daily purchase count
	if s.config.EnableDailyLimit {
		dailyPurchased, err := s.getDailyPurchaseCount(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get daily purchase count: %w", err)
		}
		info.DailyPurchased = dailyPurchased
		info.DailyLimit = s.config.DailyLimit
	}

	// 获取全局购买数量
	// Get global purchase count
	if s.config.EnableGlobalLimit {
		globalPurchased, err := s.getGlobalPurchaseCount(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get global purchase count: %w", err)
		}
		info.GlobalPurchased = globalPurchased
		info.GlobalLimit = s.config.GlobalLimit
	}

	// 计算剩余限购数量
	// Calculate remaining limit
	info.RemainingLimit = activityLimit - currentPurchased
	if info.RemainingLimit < 0 {
		info.RemainingLimit = 0
	}

	// 判断是否可以购买
	// Determine if can purchase
	info.CanPurchase = info.RemainingLimit > 0

	// 检查限购原因
	// Check limit reason
	if !info.CanPurchase {
		info.LimitReason = "activity limit exceeded"
	} else if s.config.EnableDailyLimit && info.DailyPurchased >= info.DailyLimit {
		info.CanPurchase = false
		info.LimitReason = "daily limit exceeded"
	} else if s.config.EnableGlobalLimit && info.GlobalPurchased >= info.GlobalLimit {
		info.CanPurchase = false
		info.LimitReason = "global limit exceeded"
	}

	return info, nil
}

// RecordPurchase 记录购买
// RecordPurchase records purchase
func (s *UserLimitService) RecordPurchase(ctx context.Context, record *PurchaseRecord) error {
	// 更新活动限购计数
	// Update activity limit count
	if err := s.incrementActivityCount(ctx, record.UserID, record.ActivityID, record.PurchaseAmount); err != nil {
		return fmt.Errorf("failed to increment activity count: %w", err)
	}

	// 更新每日限购计数
	// Update daily limit count
	if s.config.EnableDailyLimit {
		if err := s.incrementDailyCount(ctx, record.UserID, record.PurchaseAmount); err != nil {
			return fmt.Errorf("failed to increment daily count: %w", err)
		}
	}

	// 更新全局限购计数
	// Update global limit count
	if s.config.EnableGlobalLimit {
		if err := s.incrementGlobalCount(ctx, record.UserID, record.PurchaseAmount); err != nil {
			return fmt.Errorf("failed to increment global count: %w", err)
		}
	}

	// 记录购买历史
	// Record purchase history
	if s.config.EnableDuplicateCheck {
		if err := s.recordPurchaseHistory(ctx, record); err != nil {
			return fmt.Errorf("failed to record purchase history: %w", err)
		}
	}

	s.logger.Info("Purchase recorded",
		logger.String("user_id", record.UserID),
		logger.String("activity_id", record.ActivityID),
		logger.Int("amount", record.PurchaseAmount),
		logger.String("order_id", record.OrderID))

	return nil
}

// parseCheckResult 解析检查结果
// parseCheckResult parses check result
func (s *UserLimitService) parseCheckResult(scriptResult *lua.ScriptResult, limit int) (*CheckResult, error) {
	if !scriptResult.Success {
		return &CheckResult{
			Allowed: false,
			Reason:  scriptResult.Error,
		}, nil
	}

	// 解析Lua脚本返回结果
	// Parse Lua script return result
	luaResult, ok := scriptResult.Result.([]interface{})
	if !ok || len(luaResult) < 2 {
		return nil, fmt.Errorf("invalid script result format")
	}

	success, _ := luaResult[0].(int64)
	message, _ := luaResult[1].(string)

	if success == 1 {
		// 成功情况，解析详细信息
		// Success case, parse detailed information
		currentCount := 0
		remainingCount := limit

		if len(luaResult) > 2 {
			if count, ok := luaResult[2].(int64); ok {
				currentCount = int(count)
			}
		}
		if len(luaResult) > 3 {
			if remaining, ok := luaResult[3].(int64); ok {
				remainingCount = int(remaining)
			}
		}

		return &CheckResult{
			Allowed:        true,
			Reason:         message,
			CurrentCount:   currentCount,
			Limit:          limit,
			RemainingCount: remainingCount,
		}, nil
	} else {
		// 失败情况
		// Failure case
		currentCount := 0
		if len(luaResult) > 2 {
			if count, ok := luaResult[2].(int64); ok {
				currentCount = int(count)
			}
		}

		return &CheckResult{
			Allowed:        false,
			Reason:         message,
			CurrentCount:   currentCount,
			Limit:          limit,
			RemainingCount: limit - currentCount,
		}, nil
	}
}

// getActivityLimit 获取活动限购数量
// getActivityLimit gets activity purchase limit
func (s *UserLimitService) getActivityLimit(ctx context.Context, activityID string) (int, error) {
	// 从Redis获取活动限购配置
	// Get activity limit configuration from Redis
	limitKey := fmt.Sprintf("activity:limit:%s", activityID)
	limitStr, err := s.redisClient.Get(ctx, limitKey)
	if err != nil {
		// 如果没有配置，使用默认值
		// If no configuration, use default value
		s.logger.Debug("Activity limit not found, using default",
			logger.String("activity_id", activityID),
			logger.Int("default_limit", s.config.DefaultLimit))
		return s.config.DefaultLimit, nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		s.logger.Warn("Invalid activity limit value, using default",
			logger.String("activity_id", activityID),
			logger.String("limit_value", limitStr),
			logger.Int("default_limit", s.config.DefaultLimit))
		return s.config.DefaultLimit, nil
	}

	return limit, nil
}

// buildUserLimitKey 构建用户限购键
// buildUserLimitKey builds user limit key
func (s *UserLimitService) buildUserLimitKey(userID, activityID string) string {
	return fmt.Sprintf("seckill:user_limit:%s:%s", userID, activityID)
}

// buildDailyLimitKey 构建每日限购键
// buildDailyLimitKey builds daily limit key
func (s *UserLimitService) buildDailyLimitKey(userID string) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("seckill:daily_limit:%s:%s", userID, today)
}

// buildGlobalLimitKey 构建全局限购键
// buildGlobalLimitKey builds global limit key
func (s *UserLimitService) buildGlobalLimitKey(userID string) string {
	return fmt.Sprintf("seckill:global_limit:%s", userID)
}

// buildPurchaseHistoryKey 构建购买历史键
// buildPurchaseHistoryKey builds purchase history key
func (s *UserLimitService) buildPurchaseHistoryKey(userID, activityID string) string {
	return fmt.Sprintf("seckill:purchase_history:%s:%s", userID, activityID)
}

// getCurrentPurchaseCount 获取当前购买数量
// getCurrentPurchaseCount gets current purchase count
func (s *UserLimitService) getCurrentPurchaseCount(ctx context.Context, userID, activityID string) (int, error) {
	key := s.buildUserLimitKey(userID, activityID)
	countStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, nil // 如果键不存在，返回0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid count value: %w", err)
	}

	return count, nil
}

// getDailyPurchaseCount 获取每日购买数量
// getDailyPurchaseCount gets daily purchase count
func (s *UserLimitService) getDailyPurchaseCount(ctx context.Context, userID string) (int, error) {
	key := s.buildDailyLimitKey(userID)
	countStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, nil // 如果键不存在，返回0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid daily count value: %w", err)
	}

	return count, nil
}

// getGlobalPurchaseCount 获取全局购买数量
// getGlobalPurchaseCount gets global purchase count
func (s *UserLimitService) getGlobalPurchaseCount(ctx context.Context, userID string) (int, error) {
	key := s.buildGlobalLimitKey(userID)
	countStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, nil // 如果键不存在，返回0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid global count value: %w", err)
	}

	return count, nil
}

// hasUserPurchased 检查用户是否已购买
// hasUserPurchased checks if user has purchased
func (s *UserLimitService) hasUserPurchased(ctx context.Context, userID, activityID string) (bool, error) {
	key := s.buildPurchaseHistoryKey(userID, activityID)
	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check purchase history: %w", err)
	}

	return exists > 0, nil
}

// incrementActivityCount 增加活动购买计数
// incrementActivityCount increments activity purchase count
func (s *UserLimitService) incrementActivityCount(ctx context.Context, userID, activityID string, amount int) error {
	key := s.buildUserLimitKey(userID, activityID)
	_, err := s.redisClient.IncrBy(ctx, key, int64(amount))
	if err != nil {
		return err
	}

	// 设置过期时间
	// Set expiration time
	return s.redisClient.Expire(ctx, key, s.config.LimitTTL)
}

// incrementDailyCount 增加每日购买计数
// incrementDailyCount increments daily purchase count
func (s *UserLimitService) incrementDailyCount(ctx context.Context, userID string, amount int) error {
	key := s.buildDailyLimitKey(userID)
	_, err := s.redisClient.IncrBy(ctx, key, int64(amount))
	if err != nil {
		return err
	}

	// 设置过期时间到明天
	// Set expiration time to tomorrow
	tomorrow := s.getNextDayResetTime()
	ttl := time.Until(tomorrow)
	return s.redisClient.Expire(ctx, key, ttl)
}

// incrementGlobalCount 增加全局购买计数
// incrementGlobalCount increments global purchase count
func (s *UserLimitService) incrementGlobalCount(ctx context.Context, userID string, amount int) error {
	key := s.buildGlobalLimitKey(userID)
	_, err := s.redisClient.IncrBy(ctx, key, int64(amount))
	return err
}

// recordPurchaseHistory 记录购买历史
// recordPurchaseHistory records purchase history
func (s *UserLimitService) recordPurchaseHistory(ctx context.Context, record *PurchaseRecord) error {
	key := s.buildPurchaseHistoryKey(record.UserID, record.ActivityID)

	// 设置购买标记
	// Set purchase flag
	err := s.redisClient.Set(ctx, key, "1", s.config.LimitTTL)
	if err != nil {
		return err
	}

	// 可以在这里添加更详细的购买记录存储
	// Can add more detailed purchase record storage here
	return nil
}

// getNextDayResetTime 获取下一天重置时间
// getNextDayResetTime gets next day reset time
func (s *UserLimitService) getNextDayResetTime() time.Time {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
}

// DefaultUserLimitConfig 默认用户限购配置
// DefaultUserLimitConfig default user limit configuration
func DefaultUserLimitConfig() *UserLimitConfig {
	return &UserLimitConfig{
		DefaultLimit:         5,
		GlobalLimit:          50,
		DailyLimit:           20,
		LimitTTL:             24 * time.Hour,
		DailyLimitTTL:        24 * time.Hour,
		EnableGlobalLimit:    true,
		EnableDailyLimit:     true,
		EnableDuplicateCheck: false, // 根据业务需求决定
	}
}
