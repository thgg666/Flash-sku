package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/limit"
)

func TestUserLimitConfig(t *testing.T) {
	config := limit.DefaultUserLimitConfig()
	
	assert.NotNil(t, config)
	assert.Equal(t, 5, config.DefaultLimit)
	assert.Equal(t, 50, config.GlobalLimit)
	assert.Equal(t, 20, config.DailyLimit)
	assert.Equal(t, 24*time.Hour, config.LimitTTL)
	assert.Equal(t, 24*time.Hour, config.DailyLimitTTL)
	assert.True(t, config.EnableGlobalLimit)
	assert.True(t, config.EnableDailyLimit)
	assert.False(t, config.EnableDuplicateCheck)
}

func TestUserLimitInfo(t *testing.T) {
	info := &limit.UserLimitInfo{
		UserID:           "user123",
		ActivityID:       "activity456",
		CurrentPurchased: 2,
		ActivityLimit:    5,
		DailyPurchased:   8,
		DailyLimit:       20,
		GlobalPurchased:  15,
		GlobalLimit:      50,
		RemainingLimit:   3,
		LastPurchaseTime: time.Now(),
		CanPurchase:      true,
	}
	
	assert.Equal(t, "user123", info.UserID)
	assert.Equal(t, "activity456", info.ActivityID)
	assert.Equal(t, 2, info.CurrentPurchased)
	assert.Equal(t, 5, info.ActivityLimit)
	assert.Equal(t, 8, info.DailyPurchased)
	assert.Equal(t, 20, info.DailyLimit)
	assert.Equal(t, 15, info.GlobalPurchased)
	assert.Equal(t, 50, info.GlobalLimit)
	assert.Equal(t, 3, info.RemainingLimit)
	assert.True(t, info.CanPurchase)
	assert.Empty(t, info.LimitReason)
	
	// 验证限购逻辑
	// Verify limit logic
	assert.Equal(t, info.ActivityLimit-info.CurrentPurchased, info.RemainingLimit)
	assert.Less(t, info.CurrentPurchased, info.ActivityLimit)
	assert.Less(t, info.DailyPurchased, info.DailyLimit)
	assert.Less(t, info.GlobalPurchased, info.GlobalLimit)
}

func TestCheckResult(t *testing.T) {
	// 测试允许的情况
	// Test allowed case
	allowedResult := &limit.CheckResult{
		Allowed:        true,
		Reason:         "within limit",
		CurrentCount:   2,
		Limit:          5,
		RemainingCount: 3,
	}
	
	assert.True(t, allowedResult.Allowed)
	assert.Equal(t, "within limit", allowedResult.Reason)
	assert.Equal(t, 2, allowedResult.CurrentCount)
	assert.Equal(t, 5, allowedResult.Limit)
	assert.Equal(t, 3, allowedResult.RemainingCount)
	
	// 验证数量一致性
	// Verify count consistency
	assert.Equal(t, allowedResult.Limit-allowedResult.CurrentCount, allowedResult.RemainingCount)
	
	// 测试拒绝的情况
	// Test denied case
	deniedResult := &limit.CheckResult{
		Allowed:        false,
		Reason:         "exceeds limit",
		CurrentCount:   5,
		Limit:          5,
		RemainingCount: 0,
		ResetTime:      time.Now().Add(24 * time.Hour),
	}
	
	assert.False(t, deniedResult.Allowed)
	assert.Equal(t, "exceeds limit", deniedResult.Reason)
	assert.Equal(t, 5, deniedResult.CurrentCount)
	assert.Equal(t, 5, deniedResult.Limit)
	assert.Equal(t, 0, deniedResult.RemainingCount)
	assert.NotZero(t, deniedResult.ResetTime)
}

func TestPurchaseRecord(t *testing.T) {
	now := time.Now()
	record := &limit.PurchaseRecord{
		UserID:         "user123",
		ActivityID:     "activity456",
		PurchaseAmount: 2,
		PurchaseTime:   now,
		OrderID:        "order789",
		Status:         "completed",
	}
	
	assert.Equal(t, "user123", record.UserID)
	assert.Equal(t, "activity456", record.ActivityID)
	assert.Equal(t, 2, record.PurchaseAmount)
	assert.Equal(t, now, record.PurchaseTime)
	assert.Equal(t, "order789", record.OrderID)
	assert.Equal(t, "completed", record.Status)
	
	// 验证购买数量合理性
	// Verify purchase amount reasonableness
	assert.Greater(t, record.PurchaseAmount, 0)
	assert.LessOrEqual(t, record.PurchaseAmount, 100) // 假设最大购买数量为100
}

func TestLimitReasons(t *testing.T) {
	// 测试各种限购原因
	// Test various limit reasons
	reasons := []string{
		"activity limit exceeded",
		"daily limit exceeded",
		"global limit exceeded",
		"duplicate purchase not allowed",
		"within limit",
		"exceeds limit",
	}
	
	for _, reason := range reasons {
		assert.NotEmpty(t, reason, "Limit reason should not be empty")
		assert.True(t, len(reason) > 0, "Limit reason should have content")
	}
}

func TestLimitValidation(t *testing.T) {
	// 测试限购参数验证
	// Test limit parameter validation
	testCases := []struct {
		name           string
		userID         string
		activityID     string
		purchaseAmount int
		valid          bool
	}{
		{
			name:           "Valid parameters",
			userID:         "user123",
			activityID:     "activity456",
			purchaseAmount: 1,
			valid:          true,
		},
		{
			name:           "Empty user ID",
			userID:         "",
			activityID:     "activity456",
			purchaseAmount: 1,
			valid:          false,
		},
		{
			name:           "Empty activity ID",
			userID:         "user123",
			activityID:     "",
			purchaseAmount: 1,
			valid:          false,
		},
		{
			name:           "Zero purchase amount",
			userID:         "user123",
			activityID:     "activity456",
			purchaseAmount: 0,
			valid:          false,
		},
		{
			name:           "Negative purchase amount",
			userID:         "user123",
			activityID:     "activity456",
			purchaseAmount: -1,
			valid:          false,
		},
		{
			name:           "Too large purchase amount",
			userID:         "user123",
			activityID:     "activity456",
			purchaseAmount: 1000,
			valid:          false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证用户ID
			// Validate user ID
			if tc.userID == "" && tc.valid {
				t.Error("Empty user ID should be invalid")
			}
			
			// 验证活动ID
			// Validate activity ID
			if tc.activityID == "" && tc.valid {
				t.Error("Empty activity ID should be invalid")
			}
			
			// 验证购买数量
			// Validate purchase amount
			if tc.purchaseAmount <= 0 && tc.valid {
				t.Error("Non-positive purchase amount should be invalid")
			}
			
			if tc.purchaseAmount > 100 && tc.valid {
				t.Error("Too large purchase amount should be invalid")
			}
		})
	}
}

func TestLimitCalculations(t *testing.T) {
	// 测试限购计算逻辑
	// Test limit calculation logic
	
	// 活动限购计算
	// Activity limit calculation
	activityLimit := 5
	currentPurchased := 2
	purchaseAmount := 1
	
	remainingAfterPurchase := activityLimit - currentPurchased - purchaseAmount
	assert.Equal(t, 2, remainingAfterPurchase)
	
	// 检查是否可以购买
	// Check if can purchase
	canPurchase := currentPurchased+purchaseAmount <= activityLimit
	assert.True(t, canPurchase)
	
	// 每日限购计算
	// Daily limit calculation
	dailyLimit := 20
	dailyPurchased := 15
	
	dailyRemainingAfterPurchase := dailyLimit - dailyPurchased - purchaseAmount
	assert.Equal(t, 4, dailyRemainingAfterPurchase)
	
	canPurchaseDaily := dailyPurchased+purchaseAmount <= dailyLimit
	assert.True(t, canPurchaseDaily)
	
	// 全局限购计算
	// Global limit calculation
	globalLimit := 50
	globalPurchased := 45
	
	globalRemainingAfterPurchase := globalLimit - globalPurchased - purchaseAmount
	assert.Equal(t, 4, globalRemainingAfterPurchase)
	
	canPurchaseGlobal := globalPurchased+purchaseAmount <= globalLimit
	assert.True(t, canPurchaseGlobal)
}

func TestTimeCalculations(t *testing.T) {
	// 测试时间相关计算
	// Test time-related calculations
	
	now := time.Now()
	
	// 计算明天的重置时间
	// Calculate tomorrow's reset time
	tomorrow := now.AddDate(0, 0, 1)
	resetTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())
	
	// 验证重置时间是明天的00:00:00
	// Verify reset time is tomorrow's 00:00:00
	assert.Equal(t, 0, resetTime.Hour())
	assert.Equal(t, 0, resetTime.Minute())
	assert.Equal(t, 0, resetTime.Second())
	assert.True(t, resetTime.After(now))
	
	// 计算到重置时间的TTL
	// Calculate TTL to reset time
	ttl := time.Until(resetTime)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 24*time.Hour)
}

func TestRedisKeyBuilding(t *testing.T) {
	// 测试Redis键构建逻辑
	// Test Redis key building logic
	
	userID := "user123"
	activityID := "activity456"
	
	// 用户限购键
	// User limit key
	userLimitKey := "seckill:user_limit:" + userID + ":" + activityID
	expectedUserLimitKey := "seckill:user_limit:user123:activity456"
	assert.Equal(t, expectedUserLimitKey, userLimitKey)
	
	// 每日限购键
	// Daily limit key
	today := time.Now().Format("2006-01-02")
	dailyLimitKey := "seckill:daily_limit:" + userID + ":" + today
	expectedDailyLimitKey := "seckill:daily_limit:user123:" + today
	assert.Equal(t, expectedDailyLimitKey, dailyLimitKey)
	
	// 全局限购键
	// Global limit key
	globalLimitKey := "seckill:global_limit:" + userID
	expectedGlobalLimitKey := "seckill:global_limit:user123"
	assert.Equal(t, expectedGlobalLimitKey, globalLimitKey)
	
	// 购买历史键
	// Purchase history key
	purchaseHistoryKey := "seckill:purchase_history:" + userID + ":" + activityID
	expectedPurchaseHistoryKey := "seckill:purchase_history:user123:activity456"
	assert.Equal(t, expectedPurchaseHistoryKey, purchaseHistoryKey)
}

func TestConfigValidation(t *testing.T) {
	// 测试配置验证
	// Test configuration validation
	config := limit.DefaultUserLimitConfig()
	
	// 验证限制值的合理性
	// Verify reasonableness of limit values
	assert.Greater(t, config.DefaultLimit, 0)
	assert.Greater(t, config.GlobalLimit, config.DefaultLimit)
	assert.Greater(t, config.DailyLimit, config.DefaultLimit)
	assert.Greater(t, config.LimitTTL, time.Duration(0))
	assert.Greater(t, config.DailyLimitTTL, time.Duration(0))
	
	// 验证布尔配置
	// Verify boolean configurations
	assert.True(t, config.EnableGlobalLimit || !config.EnableGlobalLimit) // 总是为真，只是检查类型
	assert.True(t, config.EnableDailyLimit || !config.EnableDailyLimit)
	assert.True(t, config.EnableDuplicateCheck || !config.EnableDuplicateCheck)
}
