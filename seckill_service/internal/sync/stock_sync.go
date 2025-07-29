package sync

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// StockSyncer 库存同步器
// StockSyncer stock synchronizer
type StockSyncer struct {
	db          *sql.DB
	redisClient redis.Client
	logger      logger.Logger
	config      *StockSyncConfig
}

// StockSyncConfig 库存同步配置
// StockSyncConfig stock sync configuration
type StockSyncConfig struct {
	SyncInterval     time.Duration `json:"sync_interval"`
	BatchSize        int           `json:"batch_size"`
	StockTTL         time.Duration `json:"stock_ttl"`
	EnableRealtime   bool          `json:"enable_realtime"`
	ConflictStrategy string        `json:"conflict_strategy"` // "redis_priority", "db_priority", "merge"
}

// StockData 库存数据
// StockData stock data
type StockData struct {
	ActivityID      string    `json:"activity_id" db:"activity_id"`
	TotalStock      int       `json:"total_stock" db:"total_stock"`
	AvailableStock  int       `json:"available_stock" db:"available_stock"`
	ReservedStock   int       `json:"reserved_stock" db:"reserved_stock"`
	SoldStock       int       `json:"sold_stock" db:"sold_stock"`
	LastUpdated     time.Time `json:"last_updated" db:"updated_at"`
	Version         int       `json:"version" db:"version"` // 乐观锁版本号
}

// StockSyncResult 库存同步结果
// StockSyncResult stock sync result
type StockSyncResult struct {
	ActivityID    string    `json:"activity_id"`
	Success       bool      `json:"success"`
	OldStock      int       `json:"old_stock"`
	NewStock      int       `json:"new_stock"`
	ConflictType  string    `json:"conflict_type,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	SyncTime      time.Time `json:"sync_time"`
}

// NewStockSyncer 创建库存同步器
// NewStockSyncer creates stock synchronizer
func NewStockSyncer(db *sql.DB, redisClient redis.Client, config *StockSyncConfig, log logger.Logger) *StockSyncer {
	if config == nil {
		config = DefaultStockSyncConfig()
	}

	return &StockSyncer{
		db:          db,
		redisClient: redisClient,
		logger:      log,
		config:      config,
	}
}

// SyncAllStocks 同步所有库存
// SyncAllStocks synchronizes all stocks
func (s *StockSyncer) SyncAllStocks(ctx context.Context) ([]*StockSyncResult, error) {
	s.logger.Info("Starting stock synchronization")

	// 获取所有活动的库存数据
	// Get all active stock data
	stocks, err := s.fetchStocksFromDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stocks from database: %w", err)
	}

	s.logger.Info("Fetched stocks from database", logger.Int("count", len(stocks)))

	// 批量同步库存
	// Batch sync stocks
	var results []*StockSyncResult
	for i := 0; i < len(stocks); i += s.config.BatchSize {
		end := i + s.config.BatchSize
		if end > len(stocks) {
			end = len(stocks)
		}

		batch := stocks[i:end]
		batchResults, err := s.syncStockBatch(ctx, batch)
		if err != nil {
			s.logger.Error("Failed to sync stock batch",
				logger.Int("batch_start", i),
				logger.Int("batch_size", len(batch)),
				logger.Error(err))
		}
		results = append(results, batchResults...)
	}

	s.logger.Info("Stock synchronization completed",
		logger.Int("total_stocks", len(stocks)),
		logger.Int("results", len(results)))

	return results, nil
}

// SyncSingleStock 同步单个库存
// SyncSingleStock synchronizes single stock
func (s *StockSyncer) SyncSingleStock(ctx context.Context, activityID string) (*StockSyncResult, error) {
	// 从数据库获取库存
	// Get stock from database
	dbStock, err := s.fetchStockFromDB(ctx, activityID)
	if err != nil {
		return &StockSyncResult{
			ActivityID:   activityID,
			Success:      false,
			ErrorMessage: err.Error(),
			SyncTime:     time.Now(),
		}, err
	}

	// 从Redis获取库存
	// Get stock from Redis
	redisStock, err := s.fetchStockFromRedis(ctx, activityID)
	if err != nil {
		s.logger.Warn("Failed to get stock from Redis, using DB value",
			logger.String("activity_id", activityID),
			logger.Error(err))
		redisStock = -1 // 表示Redis中没有数据
	}

	// 处理冲突
	// Handle conflicts
	result := s.resolveStockConflict(dbStock, redisStock, activityID)

	// 更新Redis
	// Update Redis
	if err := s.updateStockInRedis(ctx, activityID, result.NewStock); err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("failed to update Redis: %v", err)
		return result, err
	}

	result.Success = true
	result.SyncTime = time.Now()

	s.logger.Debug("Stock synchronized",
		logger.String("activity_id", activityID),
		logger.Int("old_stock", result.OldStock),
		logger.Int("new_stock", result.NewStock))

	return result, nil
}

// fetchStocksFromDB 从数据库获取库存数据
// fetchStocksFromDB fetches stock data from database
func (s *StockSyncer) fetchStocksFromDB(ctx context.Context) ([]*StockData, error) {
	query := `
		SELECT activity_id, total_stock, available_stock, 
		       COALESCE(reserved_stock, 0) as reserved_stock,
		       COALESCE(sold_stock, 0) as sold_stock,
		       updated_at, COALESCE(version, 1) as version
		FROM seckill_activities 
		WHERE status IN ('active', 'upcoming')
		ORDER BY updated_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks: %w", err)
	}
	defer rows.Close()

	var stocks []*StockData
	for rows.Next() {
		stock := &StockData{}
		err := rows.Scan(
			&stock.ActivityID, &stock.TotalStock, &stock.AvailableStock,
			&stock.ReservedStock, &stock.SoldStock,
			&stock.LastUpdated, &stock.Version,
		)
		if err != nil {
			s.logger.Error("Failed to scan stock row", logger.Error(err))
			continue
		}
		stocks = append(stocks, stock)
	}

	return stocks, rows.Err()
}

// fetchStockFromDB 从数据库获取单个库存
// fetchStockFromDB fetches single stock from database
func (s *StockSyncer) fetchStockFromDB(ctx context.Context, activityID string) (int, error) {
	query := `
		SELECT available_stock 
		FROM seckill_activities 
		WHERE id = $1 AND status IN ('active', 'upcoming')
	`

	var stock int
	err := s.db.QueryRowContext(ctx, query, activityID).Scan(&stock)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch stock for activity %s: %w", activityID, err)
	}

	return stock, nil
}

// fetchStockFromRedis 从Redis获取库存
// fetchStockFromRedis fetches stock from Redis
func (s *StockSyncer) fetchStockFromRedis(ctx context.Context, activityID string) (int, error) {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	stockStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return -1, err
	}

	var stock int
	if _, err := fmt.Sscanf(stockStr, "%d", &stock); err != nil {
		return -1, fmt.Errorf("invalid stock value in Redis: %s", stockStr)
	}

	return stock, nil
}

// syncStockBatch 批量同步库存
// syncStockBatch batch sync stocks
func (s *StockSyncer) syncStockBatch(ctx context.Context, stocks []*StockData) ([]*StockSyncResult, error) {
	var results []*StockSyncResult

	for _, stock := range stocks {
		result, err := s.SyncSingleStock(ctx, stock.ActivityID)
		if err != nil {
			s.logger.Error("Failed to sync stock",
				logger.String("activity_id", stock.ActivityID),
				logger.Error(err))
		}
		results = append(results, result)
	}

	return results, nil
}

// resolveStockConflict 解决库存冲突
// resolveStockConflict resolves stock conflict
func (s *StockSyncer) resolveStockConflict(dbStock, redisStock int, activityID string) *StockSyncResult {
	result := &StockSyncResult{
		ActivityID: activityID,
		OldStock:   redisStock,
	}

	switch s.config.ConflictStrategy {
	case "redis_priority":
		// Redis优先：如果Redis有数据且合理，使用Redis的值
		// Redis priority: use Redis value if it exists and is reasonable
		if redisStock >= 0 && redisStock <= dbStock {
			result.NewStock = redisStock
			result.ConflictType = "redis_priority_kept"
		} else {
			result.NewStock = dbStock
			result.ConflictType = "redis_invalid_use_db"
		}

	case "db_priority":
		// 数据库优先：总是使用数据库的值
		// DB priority: always use database value
		result.NewStock = dbStock
		if redisStock != dbStock && redisStock >= 0 {
			result.ConflictType = "db_priority_override"
		}

	case "merge":
		// 合并策略：使用较小的值（更保守）
		// Merge strategy: use smaller value (more conservative)
		if redisStock >= 0 {
			if redisStock < dbStock {
				result.NewStock = redisStock
				result.ConflictType = "merge_use_smaller_redis"
			} else {
				result.NewStock = dbStock
				result.ConflictType = "merge_use_smaller_db"
			}
		} else {
			result.NewStock = dbStock
			result.ConflictType = "merge_redis_missing"
		}

	default:
		// 默认使用数据库值
		// Default to database value
		result.NewStock = dbStock
		result.ConflictType = "default_db"
	}

	return result
}

// updateStockInRedis 更新Redis中的库存
// updateStockInRedis updates stock in Redis
func (s *StockSyncer) updateStockInRedis(ctx context.Context, activityID string, stock int) error {
	key := fmt.Sprintf("seckill:stock:%s", activityID)
	return s.redisClient.Set(ctx, key, stock, s.config.StockTTL)
}

// DefaultStockSyncConfig 默认库存同步配置
// DefaultStockSyncConfig default stock sync configuration
func DefaultStockSyncConfig() *StockSyncConfig {
	return &StockSyncConfig{
		SyncInterval:     1 * time.Minute,
		BatchSize:        50,
		StockTTL:         1 * time.Hour,
		EnableRealtime:   true,
		ConflictStrategy: "merge",
	}
}
