package activity

import (
	"context"
	"fmt"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// StatusManager 活动状态管理器
// StatusManager activity status manager
type StatusManager struct {
	redisClient redis.Client
	validator   *ActivityValidator
	logger      logger.Logger
	config      *StatusManagerConfig
}

// StatusManagerConfig 状态管理器配置
// StatusManagerConfig status manager configuration
type StatusManagerConfig struct {
	StatusCacheTimeout time.Duration `json:"status_cache_timeout"`
	EnableAutoUpdate   bool          `json:"enable_auto_update"`
	UpdateInterval     time.Duration `json:"update_interval"`
	EnableNotification bool          `json:"enable_notification"`
}

// StatusTransition 状态转换
// StatusTransition status transition
type StatusTransition struct {
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
	Operator   string    `json:"operator"`
}

// StatusHistory 状态历史
// StatusHistory status history
type StatusHistory struct {
	ActivityID  string             `json:"activity_id"`
	Transitions []StatusTransition `json:"transitions"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// NewStatusManager 创建状态管理器
// NewStatusManager creates status manager
func NewStatusManager(redisClient redis.Client, validator *ActivityValidator, config *StatusManagerConfig, log logger.Logger) *StatusManager {
	if config == nil {
		config = DefaultStatusManagerConfig()
	}

	return &StatusManager{
		redisClient: redisClient,
		validator:   validator,
		logger:      log,
		config:      config,
	}
}

// UpdateActivityStatus 更新活动状态
// UpdateActivityStatus updates activity status
func (sm *StatusManager) UpdateActivityStatus(ctx context.Context, activityID, newStatus, reason, operator string) error {
	// 获取当前状态
	// Get current status
	currentStatus, err := sm.GetActivityStatus(ctx, activityID)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	// 验证状态转换
	// Validate status transition
	if !sm.isValidTransition(currentStatus, newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
	}

	// 更新Redis中的状态
	// Update status in Redis
	statusKey := sm.buildStatusKey(activityID)
	if err := sm.redisClient.Set(ctx, statusKey, newStatus, sm.config.StatusCacheTimeout); err != nil {
		return fmt.Errorf("failed to update status in Redis: %w", err)
	}

	// 记录状态转换历史
	// Record status transition history
	transition := StatusTransition{
		FromStatus: currentStatus,
		ToStatus:   newStatus,
		Reason:     reason,
		Timestamp:  time.Now(),
		Operator:   operator,
	}

	if err := sm.recordStatusTransition(ctx, activityID, transition); err != nil {
		sm.logger.Warn("Failed to record status transition",
			logger.String("activity_id", activityID),
			logger.Error(err))
	}

	// 发送状态变更通知
	// Send status change notification
	if sm.config.EnableNotification {
		sm.notifyStatusChange(ctx, activityID, transition)
	}

	sm.logger.Info("Activity status updated",
		logger.String("activity_id", activityID),
		logger.String("from_status", currentStatus),
		logger.String("to_status", newStatus),
		logger.String("reason", reason),
		logger.String("operator", operator))

	return nil
}

// GetActivityStatus 获取活动状态
// GetActivityStatus gets activity status
func (sm *StatusManager) GetActivityStatus(ctx context.Context, activityID string) (string, error) {
	// 先从Redis缓存获取
	// First get from Redis cache
	statusKey := sm.buildStatusKey(activityID)
	status, err := sm.redisClient.Get(ctx, statusKey)
	if err == nil && status != "" {
		return status, nil
	}

	// 缓存未命中，从活动信息获取
	// Cache miss, get from activity info
	activityInfo, err := sm.validator.GetActivityInfo(ctx, activityID)
	if err != nil {
		return "", fmt.Errorf("failed to get activity info: %w", err)
	}

	// 缓存状态
	// Cache status
	if err := sm.redisClient.Set(ctx, statusKey, activityInfo.Status, sm.config.StatusCacheTimeout); err != nil {
		sm.logger.Warn("Failed to cache activity status",
			logger.String("activity_id", activityID),
			logger.Error(err))
	}

	return activityInfo.Status, nil
}

// GetStatusHistory 获取状态历史
// GetStatusHistory gets status history
func (sm *StatusManager) GetStatusHistory(ctx context.Context, activityID string) (*StatusHistory, error) {
	// 这里应该从数据库或Redis获取历史记录
	// Should get history from database or Redis here
	// 暂时返回空历史
	// Return empty history for now
	return &StatusHistory{
		ActivityID:  activityID,
		Transitions: []StatusTransition{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// AutoUpdateStatuses 自动更新状态
// AutoUpdateStatuses automatically updates statuses
func (sm *StatusManager) AutoUpdateStatuses(ctx context.Context) error {
	if !sm.config.EnableAutoUpdate {
		return nil
	}

	// 获取需要更新的活动列表
	// Get list of activities that need updating
	activities, err := sm.getActivitiesForUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get activities for update: %w", err)
	}

	for _, activityID := range activities {
		if err := sm.updateActivityStatusAuto(ctx, activityID); err != nil {
			sm.logger.Error("Failed to auto update activity status",
				logger.String("activity_id", activityID),
				logger.Error(err))
		}
	}

	return nil
}

// StartAutoUpdate 启动自动更新
// StartAutoUpdate starts auto update
func (sm *StatusManager) StartAutoUpdate(ctx context.Context) {
	if !sm.config.EnableAutoUpdate {
		return
	}

	ticker := time.NewTicker(sm.config.UpdateInterval)
	defer ticker.Stop()

	sm.logger.Info("Starting activity status auto update",
		logger.Duration("interval", sm.config.UpdateInterval))

	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("Activity status auto update stopped")
			return
		case <-ticker.C:
			if err := sm.AutoUpdateStatuses(ctx); err != nil {
				sm.logger.Error("Auto update statuses failed", logger.Error(err))
			}
		}
	}
}

// isValidTransition 检查状态转换是否有效
// isValidTransition checks if status transition is valid
func (sm *StatusManager) isValidTransition(fromStatus, toStatus string) bool {
	// 定义有效的状态转换规则
	// Define valid status transition rules
	validTransitions := map[string][]string{
		StatusDraft:     {StatusScheduled, StatusCancelled},
		StatusScheduled: {StatusActive, StatusCancelled},
		StatusActive:    {StatusPaused, StatusEnded, StatusCancelled},
		StatusPaused:    {StatusActive, StatusEnded, StatusCancelled},
		StatusEnded:     {}, // 已结束状态不能转换到其他状态
		StatusCancelled: {}, // 已取消状态不能转换到其他状态
	}

	allowedStatuses, exists := validTransitions[fromStatus]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if toStatus == allowedStatus {
			return true
		}
	}

	return false
}

// updateActivityStatusAuto 自动更新活动状态
// updateActivityStatusAuto automatically updates activity status
func (sm *StatusManager) updateActivityStatusAuto(ctx context.Context, activityID string) error {
	// 获取活动信息
	// Get activity information
	activityInfo, err := sm.validator.GetActivityInfo(ctx, activityID)
	if err != nil {
		return err
	}

	currentTime := time.Now()
	newStatus := activityInfo.Status

	// 根据时间自动更新状态
	// Auto update status based on time
	switch activityInfo.Status {
	case StatusScheduled:
		if currentTime.After(activityInfo.StartTime) {
			newStatus = StatusActive
		}
	case StatusActive:
		if currentTime.After(activityInfo.EndTime) {
			newStatus = StatusEnded
		}
	}

	// 如果状态需要更新
	// If status needs to be updated
	if newStatus != activityInfo.Status {
		reason := fmt.Sprintf("Auto updated based on time: %s", currentTime.Format(time.RFC3339))
		return sm.UpdateActivityStatus(ctx, activityID, newStatus, reason, "system")
	}

	return nil
}

// getActivitiesForUpdate 获取需要更新的活动列表
// getActivitiesForUpdate gets list of activities that need updating
func (sm *StatusManager) getActivitiesForUpdate(ctx context.Context) ([]string, error) {
	// 这里应该从数据库查询需要更新的活动
	// Should query database for activities that need updating
	// 暂时返回空列表
	// Return empty list for now
	return []string{}, nil
}

// recordStatusTransition 记录状态转换
// recordStatusTransition records status transition
func (sm *StatusManager) recordStatusTransition(ctx context.Context, activityID string, transition StatusTransition) error {
	// 这里应该将状态转换记录到数据库
	// Should record status transition to database
	sm.logger.Debug("Status transition recorded",
		logger.String("activity_id", activityID),
		logger.String("from_status", transition.FromStatus),
		logger.String("to_status", transition.ToStatus),
		logger.String("reason", transition.Reason))

	return nil
}

// notifyStatusChange 发送状态变更通知
// notifyStatusChange sends status change notification
func (sm *StatusManager) notifyStatusChange(ctx context.Context, activityID string, transition StatusTransition) {
	// 这里应该发送状态变更通知
	// Should send status change notification
	sm.logger.Info("Activity status change notification",
		logger.String("activity_id", activityID),
		logger.String("from_status", transition.FromStatus),
		logger.String("to_status", transition.ToStatus))
}

// buildStatusKey 构建状态键
// buildStatusKey builds status key
func (sm *StatusManager) buildStatusKey(activityID string) string {
	return fmt.Sprintf("seckill:status:%s", activityID)
}

// buildStatusHistoryKey 构建状态历史键
// buildStatusHistoryKey builds status history key
func (sm *StatusManager) buildStatusHistoryKey(activityID string) string {
	return fmt.Sprintf("seckill:status_history:%s", activityID)
}

// DefaultStatusManagerConfig 默认状态管理器配置
// DefaultStatusManagerConfig default status manager configuration
func DefaultStatusManagerConfig() *StatusManagerConfig {
	return &StatusManagerConfig{
		StatusCacheTimeout: 10 * time.Minute,
		EnableAutoUpdate:   true,
		UpdateInterval:     1 * time.Minute,
		EnableNotification: true,
	}
}
