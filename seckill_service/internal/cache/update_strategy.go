package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// UpdateStrategy 缓存更新策略
// UpdateStrategy cache update strategy
type UpdateStrategy interface {
	Update(ctx context.Context, key string, value interface{}) error
	Invalidate(ctx context.Context, key string) error
	Refresh(ctx context.Context, key string, loader DataLoader) error
}

// DataLoader 数据加载器
// DataLoader data loader
type DataLoader func(ctx context.Context) (interface{}, error)

// CacheUpdateManager 缓存更新管理器
// CacheUpdateManager cache update manager
type CacheUpdateManager struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *UpdateConfig
	strategies  map[string]UpdateStrategy
	mu          sync.RWMutex
}

// UpdateConfig 更新配置
// UpdateConfig update configuration
type UpdateConfig struct {
	DefaultTTL       time.Duration `json:"default_ttl"`
	RefreshThreshold float64       `json:"refresh_threshold"` // 剩余TTL比例阈值
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	EnableAsync      bool          `json:"enable_async"`
	BatchSize        int           `json:"batch_size"`
}

// UpdateResult 更新结果
// UpdateResult update result
type UpdateResult struct {
	Key       string        `json:"key"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Strategy  string        `json:"strategy"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewCacheUpdateManager 创建缓存更新管理器
// NewCacheUpdateManager creates cache update manager
func NewCacheUpdateManager(redisClient redis.Client, config *UpdateConfig, log logger.Logger) *CacheUpdateManager {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	manager := &CacheUpdateManager{
		redisClient: redisClient,
		logger:      log,
		config:      config,
		strategies:  make(map[string]UpdateStrategy),
	}

	// 注册默认策略
	// Register default strategies
	manager.RegisterStrategy("write_through", NewWriteThroughStrategy(redisClient, log))
	manager.RegisterStrategy("write_behind", NewWriteBehindStrategy(redisClient, log))
	manager.RegisterStrategy("refresh_ahead", NewRefreshAheadStrategy(redisClient, config, log))

	return manager
}

// RegisterStrategy 注册更新策略
// RegisterStrategy registers update strategy
func (m *CacheUpdateManager) RegisterStrategy(name string, strategy UpdateStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategies[name] = strategy
}

// UpdateWithStrategy 使用指定策略更新缓存
// UpdateWithStrategy updates cache with specified strategy
func (m *CacheUpdateManager) UpdateWithStrategy(ctx context.Context, strategyName, key string, value interface{}) (*UpdateResult, error) {
	startTime := time.Now()
	result := &UpdateResult{
		Key:       key,
		Strategy:  strategyName,
		Timestamp: startTime,
	}

	m.mu.RLock()
	strategy, exists := m.strategies[strategyName]
	m.mu.RUnlock()

	if !exists {
		result.Error = fmt.Sprintf("strategy %s not found", strategyName)
		result.Duration = time.Since(startTime)
		return result, fmt.Errorf("strategy %s not found", strategyName)
	}

	err := strategy.Update(ctx, key, value)
	result.Duration = time.Since(startTime)
	result.Success = err == nil

	if err != nil {
		result.Error = err.Error()
		m.logger.Error("Cache update failed",
			logger.String("strategy", strategyName),
			logger.String("key", key),
			logger.Error(err))
	} else {
		m.logger.Debug("Cache updated successfully",
			logger.String("strategy", strategyName),
			logger.String("key", key),
			logger.Duration("duration", result.Duration))
	}

	return result, err
}

// BatchUpdate 批量更新缓存
// BatchUpdate batch updates cache
func (m *CacheUpdateManager) BatchUpdate(ctx context.Context, strategyName string, updates map[string]interface{}) ([]*UpdateResult, error) {
	var results []*UpdateResult
	var errors []error

	// 分批处理
	// Process in batches
	keys := make([]string, 0, len(updates))
	for key := range updates {
		keys = append(keys, key)
	}

	for i := 0; i < len(keys); i += m.config.BatchSize {
		end := i + m.config.BatchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		batchResults, batchErrors := m.processBatch(ctx, strategyName, batch, updates)
		results = append(results, batchResults...)
		errors = append(errors, batchErrors...)
	}

	var combinedError error
	if len(errors) > 0 {
		combinedError = fmt.Errorf("batch update had %d errors", len(errors))
	}

	return results, combinedError
}

// processBatch 处理批次
// processBatch processes batch
func (m *CacheUpdateManager) processBatch(ctx context.Context, strategyName string, keys []string, updates map[string]interface{}) ([]*UpdateResult, []error) {
	var results []*UpdateResult
	var errors []error

	for _, key := range keys {
		value := updates[key]
		result, err := m.UpdateWithStrategy(ctx, strategyName, key, value)
		results = append(results, result)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return results, errors
}

// InvalidatePattern 按模式失效缓存
// InvalidatePattern invalidates cache by pattern
func (m *CacheUpdateManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// 注意：这里简化实现，实际应该使用SCAN命令
	// Note: Simplified implementation, should use SCAN command in practice
	m.logger.Info("Invalidating cache pattern", logger.String("pattern", pattern))
	
	// 实际实现应该：
	// 1. 使用SCAN命令遍历匹配的键
	// 2. 批量删除匹配的键
	// 3. 记录删除的键数量
	
	return nil
}

// RefreshExpiring 刷新即将过期的缓存
// RefreshExpiring refreshes expiring cache
func (m *CacheUpdateManager) RefreshExpiring(ctx context.Context, keys []string, loaders map[string]DataLoader) error {
	for _, key := range keys {
		// 检查TTL
		// Check TTL
		ttl, err := m.redisClient.TTL(ctx, key)
		if err != nil {
			m.logger.Warn("Failed to get TTL", logger.String("key", key), logger.Error(err))
			continue
		}

		// 计算剩余时间比例
		// Calculate remaining time ratio
		if ttl > 0 {
			totalTTL := m.config.DefaultTTL
			remainingRatio := float64(ttl) / float64(totalTTL)

			// 如果剩余时间低于阈值，触发刷新
			// If remaining time is below threshold, trigger refresh
			if remainingRatio < m.config.RefreshThreshold {
				if loader, exists := loaders[key]; exists {
					go m.asyncRefresh(ctx, key, loader)
				}
			}
		}
	}

	return nil
}

// asyncRefresh 异步刷新
// asyncRefresh asynchronous refresh
func (m *CacheUpdateManager) asyncRefresh(ctx context.Context, key string, loader DataLoader) {
	m.logger.Debug("Starting async refresh", logger.String("key", key))

	data, err := loader(ctx)
	if err != nil {
		m.logger.Error("Failed to load data for refresh",
			logger.String("key", key),
			logger.Error(err))
		return
	}

	// 使用refresh_ahead策略更新
	// Update using refresh_ahead strategy
	_, err = m.UpdateWithStrategy(ctx, "refresh_ahead", key, data)
	if err != nil {
		m.logger.Error("Failed to refresh cache",
			logger.String("key", key),
			logger.Error(err))
	} else {
		m.logger.Debug("Cache refreshed successfully", logger.String("key", key))
	}
}

// GetUpdateStats 获取更新统计
// GetUpdateStats gets update statistics
func (m *CacheUpdateManager) GetUpdateStats() map[string]interface{} {
	return map[string]interface{}{
		"strategies_count": len(m.strategies),
		"config":          m.config,
		"timestamp":       time.Now(),
	}
}

// DefaultUpdateConfig 默认更新配置
// DefaultUpdateConfig default update configuration
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		DefaultTTL:       1 * time.Hour,
		RefreshThreshold: 0.2, // 剩余20%时间时刷新
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
		EnableAsync:      true,
		BatchSize:        50,
	}
}
