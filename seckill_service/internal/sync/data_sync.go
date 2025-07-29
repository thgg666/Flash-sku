package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// DataSyncer 数据同步器
// DataSyncer data synchronizer
type DataSyncer struct {
	db          *sql.DB
	redisClient redis.Client
	logger      logger.Logger
	config      *SyncConfig
}

// SyncConfig 同步配置
// SyncConfig synchronization configuration
type SyncConfig struct {
	DatabaseURL     string        `json:"database_url"`
	SyncInterval    time.Duration `json:"sync_interval"`
	BatchSize       int           `json:"batch_size"`
	RetryAttempts   int           `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
	CacheExpiration time.Duration `json:"cache_expiration"`
}

// ActivityData 活动数据
// ActivityData activity data
type ActivityData struct {
	ID             string    `json:"id" db:"id"`
	ProductID      string    `json:"product_id" db:"product_id"`
	Name           string    `json:"name" db:"name"`
	StartTime      time.Time `json:"start_time" db:"start_time"`
	EndTime        time.Time `json:"end_time" db:"end_time"`
	OriginalPrice  float64   `json:"original_price" db:"original_price"`
	SeckillPrice   float64   `json:"seckill_price" db:"seckill_price"`
	TotalStock     int       `json:"total_stock" db:"total_stock"`
	AvailableStock int       `json:"available_stock" db:"available_stock"`
	MaxPerUser     int       `json:"max_per_user" db:"max_per_user"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// SyncMetrics 同步指标
// SyncMetrics synchronization metrics
type SyncMetrics struct {
	LastSyncTime     time.Time     `json:"last_sync_time"`
	TotalSynced      int64         `json:"total_synced"`
	SuccessCount     int64         `json:"success_count"`
	ErrorCount       int64         `json:"error_count"`
	AvgSyncDuration  time.Duration `json:"avg_sync_duration"`
	LastSyncDuration time.Duration `json:"last_sync_duration"`
}

// NewDataSyncer 创建数据同步器
// NewDataSyncer creates data synchronizer
func NewDataSyncer(config *SyncConfig, redisClient redis.Client, log logger.Logger) (*DataSyncer, error) {
	// 连接数据库
	// Connect to database
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 测试数据库连接
	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DataSyncer{
		db:          db,
		redisClient: redisClient,
		logger:      log,
		config:      config,
	}, nil
}

// SyncAllActivities 同步所有活动数据
// SyncAllActivities synchronizes all activity data
func (s *DataSyncer) SyncAllActivities(ctx context.Context) error {
	startTime := time.Now()
	s.logger.Info("Starting activity data synchronization")

	// 查询活动数据
	// Query activity data
	activities, err := s.fetchActivitiesFromDB(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch activities from database", logger.Error(err))
		return err
	}

	s.logger.Info("Fetched activities from database",
		logger.Int("count", len(activities)))

	// 批量同步到Redis
	// Batch sync to Redis
	successCount := 0
	errorCount := 0

	for i := 0; i < len(activities); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(activities) {
			end = len(activities)
		}

		batch := activities[i:end]
		if err := s.syncBatchToRedis(ctx, batch); err != nil {
			s.logger.Error("Failed to sync batch to Redis",
				logger.Int("batch_start", i),
				logger.Int("batch_size", len(batch)),
				logger.Error(err))
			errorCount += len(batch)
		} else {
			successCount += len(batch)
		}
	}

	duration := time.Since(startTime)
	s.logger.Info("Activity data synchronization completed",
		logger.Int("success_count", successCount),
		logger.Int("error_count", errorCount),
		logger.Duration("duration", duration))

	// 更新同步指标
	// Update sync metrics
	s.updateSyncMetrics(successCount, errorCount, duration)

	return nil
}

// fetchActivitiesFromDB 从数据库获取活动数据
// fetchActivitiesFromDB fetches activity data from database
func (s *DataSyncer) fetchActivitiesFromDB(ctx context.Context) ([]*ActivityData, error) {
	query := `
		SELECT id, product_id, name, start_time, end_time, 
		       original_price, seckill_price, total_stock, available_stock,
		       max_per_user, status, created_at, updated_at
		FROM seckill_activities 
		WHERE status IN ('active', 'upcoming')
		ORDER BY start_time ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*ActivityData
	for rows.Next() {
		activity := &ActivityData{}
		err := rows.Scan(
			&activity.ID, &activity.ProductID, &activity.Name,
			&activity.StartTime, &activity.EndTime,
			&activity.OriginalPrice, &activity.SeckillPrice,
			&activity.TotalStock, &activity.AvailableStock,
			&activity.MaxPerUser, &activity.Status,
			&activity.CreatedAt, &activity.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan activity row", logger.Error(err))
			continue
		}
		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return activities, nil
}

// syncBatchToRedis 批量同步到Redis
// syncBatchToRedis batch sync to Redis
func (s *DataSyncer) syncBatchToRedis(ctx context.Context, activities []*ActivityData) error {
	for _, activity := range activities {
		if err := s.syncActivityToRedis(ctx, activity); err != nil {
			return err
		}
	}
	return nil
}

// syncActivityToRedis 同步单个活动到Redis
// syncActivityToRedis syncs single activity to Redis
func (s *DataSyncer) syncActivityToRedis(ctx context.Context, activity *ActivityData) error {
	// 序列化活动数据
	// Serialize activity data
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}

	// 存储活动基本信息
	// Store activity basic info
	activityKey := fmt.Sprintf("seckill:activity:%s", activity.ID)
	if err := s.redisClient.Set(ctx, activityKey, activityJSON, s.config.CacheExpiration); err != nil {
		return fmt.Errorf("failed to set activity data: %w", err)
	}

	// 存储库存信息
	// Store stock info
	stockKey := fmt.Sprintf("seckill:stock:%s", activity.ID)
	if err := s.redisClient.Set(ctx, stockKey, activity.AvailableStock, s.config.CacheExpiration); err != nil {
		return fmt.Errorf("failed to set stock data: %w", err)
	}

	// 存储活动状态
	// Store activity status
	statusKey := fmt.Sprintf("seckill:status:%s", activity.ID)
	if err := s.redisClient.Set(ctx, statusKey, activity.Status, s.config.CacheExpiration); err != nil {
		return fmt.Errorf("failed to set status data: %w", err)
	}

	// 添加到活动列表
	// Add to activity list
	if activity.Status == "active" {
		listKey := "seckill:activities:active"
		// 使用Redis的SADD命令添加到集合
		// Use Redis SADD command to add to set
		// 注意：这里简化处理，实际应该使用SADD
		// Note: Simplified here, should use SADD in practice
		_ = listKey // 暂时忽略，后续实现
	}

	s.logger.Debug("Synced activity to Redis",
		logger.String("activity_id", activity.ID),
		logger.String("status", activity.Status),
		logger.Int("stock", activity.AvailableStock))

	return nil
}

// updateSyncMetrics 更新同步指标
// updateSyncMetrics updates sync metrics
func (s *DataSyncer) updateSyncMetrics(successCount, errorCount int, duration time.Duration) {
	metricsKey := "seckill:sync:metrics"

	metrics := SyncMetrics{
		LastSyncTime:     time.Now(),
		TotalSynced:      int64(successCount),
		SuccessCount:     int64(successCount),
		ErrorCount:       int64(errorCount),
		LastSyncDuration: duration,
		AvgSyncDuration:  duration, // 简化处理
	}

	metricsJSON, _ := json.Marshal(metrics)
	s.redisClient.Set(context.Background(), metricsKey, metricsJSON, 24*time.Hour)
}

// Close 关闭同步器
// Close closes synchronizer
func (s *DataSyncer) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
