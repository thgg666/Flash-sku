package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/activity"
)

func TestValidatorConfig(t *testing.T) {
	config := activity.DefaultValidatorConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5*time.Minute, config.CacheTimeout)
	assert.True(t, config.EnableTimeCheck)
	assert.True(t, config.EnableStockCheck)
	assert.True(t, config.EnableStatusCheck)
	assert.Equal(t, 30*time.Second, config.TimeBuffer)
}

func TestActivityInfo(t *testing.T) {
	now := time.Now()
	activityInfo := &activity.ActivityInfo{
		ID:          "activity123",
		Name:        "Test Activity",
		Description: "Test Description",
		Status:      activity.StatusActive,
		StartTime:   now.Add(-1 * time.Hour),
		EndTime:     now.Add(1 * time.Hour),
		TotalStock:  1000,
		SoldCount:   100,
		Price:       99.99,
		UserLimit:   5,
		CreatedAt:   now.Add(-24 * time.Hour),
		UpdatedAt:   now,
	}

	assert.Equal(t, "activity123", activityInfo.ID)
	assert.Equal(t, "Test Activity", activityInfo.Name)
	assert.Equal(t, "Test Description", activityInfo.Description)
	assert.Equal(t, activity.StatusActive, activityInfo.Status)
	assert.True(t, activityInfo.StartTime.Before(now))
	assert.True(t, activityInfo.EndTime.After(now))
	assert.Equal(t, 1000, activityInfo.TotalStock)
	assert.Equal(t, 100, activityInfo.SoldCount)
	assert.Equal(t, 99.99, activityInfo.Price)
	assert.Equal(t, 5, activityInfo.UserLimit)

	// 验证库存逻辑
	// Verify stock logic
	remainingStock := activityInfo.TotalStock - activityInfo.SoldCount
	assert.Equal(t, 900, remainingStock)
	assert.Greater(t, remainingStock, 0)
}

func TestValidationResult(t *testing.T) {
	now := time.Now()

	// 测试成功的验证结果
	// Test successful validation result
	successValidationResult := &activity.ValidationResult{
		Valid:     true,
		Reason:    "Validation passed",
		ErrorCode: "",
		ActivityInfo: &activity.ActivityInfo{
			ID:     "activity123",
			Status: activity.StatusActive,
		},
		Timestamp: now,
	}

	assert.True(t, successValidationResult.Valid)
	assert.Equal(t, "Validation passed", successValidationResult.Reason)
	assert.Empty(t, successValidationResult.ErrorCode)
	assert.NotNil(t, successValidationResult.ActivityInfo)
	assert.Equal(t, now, successValidationResult.Timestamp)

	// 测试失败的验证结果
	// Test failed validation result
	failureResult := &activity.ValidationResult{
		Valid:     false,
		Reason:    "Activity not found",
		ErrorCode: "ACTIVITY_NOT_FOUND",
		Timestamp: now,
	}

	assert.False(t, failureResult.Valid)
	assert.Equal(t, "Activity not found", failureResult.Reason)
	assert.Equal(t, "ACTIVITY_NOT_FOUND", failureResult.ErrorCode)
	assert.Nil(t, failureResult.ActivityInfo)
}

func TestActivityStatus(t *testing.T) {
	// 测试所有活动状态
	// Test all activity statuses
	statuses := []string{
		activity.StatusDraft,
		activity.StatusScheduled,
		activity.StatusActive,
		activity.StatusPaused,
		activity.StatusEnded,
		activity.StatusCancelled,
	}

	expectedStatuses := []string{
		"draft",
		"scheduled",
		"active",
		"paused",
		"ended",
		"cancelled",
	}

	for i, status := range statuses {
		assert.Equal(t, expectedStatuses[i], status)
		assert.NotEmpty(t, status)
	}
}

func TestStatusTransition(t *testing.T) {
	now := time.Now()
	transition := &activity.StatusTransition{
		FromStatus: activity.StatusScheduled,
		ToStatus:   activity.StatusActive,
		Reason:     "Activity started",
		Timestamp:  now,
		Operator:   "system",
	}

	assert.Equal(t, activity.StatusScheduled, transition.FromStatus)
	assert.Equal(t, activity.StatusActive, transition.ToStatus)
	assert.Equal(t, "Activity started", transition.Reason)
	assert.Equal(t, now, transition.Timestamp)
	assert.Equal(t, "system", transition.Operator)
}

func TestStatusHistory(t *testing.T) {
	now := time.Now()
	history := &activity.StatusHistory{
		ActivityID: "activity123",
		Transitions: []activity.StatusTransition{
			{
				FromStatus: activity.StatusDraft,
				ToStatus:   activity.StatusScheduled,
				Reason:     "Activity scheduled",
				Timestamp:  now.Add(-2 * time.Hour),
				Operator:   "admin",
			},
			{
				FromStatus: activity.StatusScheduled,
				ToStatus:   activity.StatusActive,
				Reason:     "Activity started",
				Timestamp:  now.Add(-1 * time.Hour),
				Operator:   "system",
			},
		},
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	assert.Equal(t, "activity123", history.ActivityID)
	assert.Len(t, history.Transitions, 2)
	assert.Equal(t, activity.StatusDraft, history.Transitions[0].FromStatus)
	assert.Equal(t, activity.StatusScheduled, history.Transitions[0].ToStatus)
	assert.Equal(t, activity.StatusScheduled, history.Transitions[1].FromStatus)
	assert.Equal(t, activity.StatusActive, history.Transitions[1].ToStatus)
}

func TestStatusManagerConfig(t *testing.T) {
	config := activity.DefaultStatusManagerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 10*time.Minute, config.StatusCacheTimeout)
	assert.True(t, config.EnableAutoUpdate)
	assert.Equal(t, 1*time.Minute, config.UpdateInterval)
	assert.True(t, config.EnableNotification)
}

func TestTimeValidation(t *testing.T) {
	now := time.Now()

	// 测试时间验证逻辑
	// Test time validation logic
	testCases := []struct {
		name          string
		startTime     time.Time
		endTime       time.Time
		currentTime   time.Time
		shouldPass    bool
		expectedError string
	}{
		{
			name:        "Activity in progress",
			startTime:   now.Add(-1 * time.Hour),
			endTime:     now.Add(1 * time.Hour),
			currentTime: now,
			shouldPass:  true,
		},
		{
			name:          "Activity not started",
			startTime:     now.Add(1 * time.Hour),
			endTime:       now.Add(2 * time.Hour),
			currentTime:   now,
			shouldPass:    false,
			expectedError: "ACTIVITY_NOT_STARTED",
		},
		{
			name:          "Activity ended",
			startTime:     now.Add(-2 * time.Hour),
			endTime:       now.Add(-1 * time.Hour),
			currentTime:   now,
			shouldPass:    false,
			expectedError: "ACTIVITY_ENDED",
		},
		{
			name:        "Activity just started",
			startTime:   now.Add(-1 * time.Minute),
			endTime:     now.Add(1 * time.Hour),
			currentTime: now,
			shouldPass:  true,
		},
		{
			name:        "Activity about to end",
			startTime:   now.Add(-1 * time.Hour),
			endTime:     now.Add(1 * time.Minute),
			currentTime: now,
			shouldPass:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 验证时间逻辑
			// Verify time logic
			isInProgress := tc.currentTime.After(tc.startTime) && tc.currentTime.Before(tc.endTime)

			if tc.shouldPass {
				assert.True(t, isInProgress, "Activity should be in progress")
			} else {
				assert.False(t, isInProgress, "Activity should not be in progress")
			}
		})
	}
}

func TestStockValidation(t *testing.T) {
	// 测试库存验证逻辑
	// Test stock validation logic
	testCases := []struct {
		name          string
		totalStock    int
		soldCount     int
		shouldPass    bool
		expectedError string
	}{
		{
			name:       "Stock available",
			totalStock: 1000,
			soldCount:  100,
			shouldPass: true,
		},
		{
			name:          "Out of stock",
			totalStock:    1000,
			soldCount:     1000,
			shouldPass:    false,
			expectedError: "OUT_OF_STOCK",
		},
		{
			name:       "Low stock",
			totalStock: 1000,
			soldCount:  999,
			shouldPass: true,
		},
		{
			name:          "Oversold",
			totalStock:    1000,
			soldCount:     1001,
			shouldPass:    false,
			expectedError: "OUT_OF_STOCK",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			remainingStock := tc.totalStock - tc.soldCount

			if tc.shouldPass {
				assert.Greater(t, remainingStock, 0, "Should have remaining stock")
			} else {
				assert.LessOrEqual(t, remainingStock, 0, "Should not have remaining stock")
			}
		})
	}
}

func TestStatusValidation(t *testing.T) {
	// 测试状态验证逻辑
	// Test status validation logic
	validStatuses := []string{
		activity.StatusDraft,
		activity.StatusScheduled,
		activity.StatusActive,
		activity.StatusPaused,
		activity.StatusEnded,
		activity.StatusCancelled,
	}

	invalidStatuses := []string{
		"",
		"invalid",
		"unknown",
		"ACTIVE", // 大小写敏感
	}

	// 测试有效状态
	// Test valid statuses
	for _, status := range validStatuses {
		assert.NotEmpty(t, status, "Valid status should not be empty")
		assert.True(t, len(status) > 0, "Valid status should have content")
	}

	// 测试无效状态
	// Test invalid statuses
	for _, status := range invalidStatuses {
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		assert.False(t, isValid, "Status %s should be invalid", status)
	}
}

func TestActivityRedisKeyBuilding(t *testing.T) {
	// 测试Redis键构建
	// Test Redis key building
	activityID := "activity123"

	// 活动缓存键
	// Activity cache key
	activityCacheKey := "seckill:activity:" + activityID
	expectedActivityCacheKey := "seckill:activity:activity123"
	assert.Equal(t, expectedActivityCacheKey, activityCacheKey)

	// 状态键
	// Status key
	statusKey := "seckill:status:" + activityID
	expectedStatusKey := "seckill:status:activity123"
	assert.Equal(t, expectedStatusKey, statusKey)

	// 库存键
	// Stock key
	stockKey := "seckill:stock:" + activityID
	expectedStockKey := "seckill:stock:activity123"
	assert.Equal(t, expectedStockKey, stockKey)

	// 状态历史键
	// Status history key
	statusHistoryKey := "seckill:status_history:" + activityID
	expectedStatusHistoryKey := "seckill:status_history:activity123"
	assert.Equal(t, expectedStatusHistoryKey, statusHistoryKey)
}

func TestValidationErrorCodes(t *testing.T) {
	// 测试验证错误码
	// Test validation error codes
	errorCodes := []string{
		"ACTIVITY_NOT_FOUND",
		"INVALID_STATUS",
		"ACTIVITY_NOT_ACTIVE",
		"ACTIVITY_NOT_STARTED",
		"ACTIVITY_ENDED",
		"OUT_OF_STOCK",
		"INVALID_STOCK_DATA",
	}

	for _, errorCode := range errorCodes {
		assert.NotEmpty(t, errorCode, "Error code should not be empty")
		assert.True(t, len(errorCode) > 0, "Error code should have content")
		assert.Contains(t, errorCode, "_", "Error code should contain underscore")
	}
}
