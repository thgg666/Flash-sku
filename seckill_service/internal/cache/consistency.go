package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// ConsistencyManager 一致性管理器
// ConsistencyManager consistency manager
type ConsistencyManager struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *ConsistencyConfig
	validators  map[string]DataValidator
	mu          sync.RWMutex
}

// ConsistencyConfig 一致性配置
// ConsistencyConfig consistency configuration
type ConsistencyConfig struct {
	CheckInterval    time.Duration `json:"check_interval"`
	RepairEnabled    bool          `json:"repair_enabled"`
	MaxRepairRetries int           `json:"max_repair_retries"`
	RepairDelay      time.Duration `json:"repair_delay"`
	AlertThreshold   float64       `json:"alert_threshold"` // 不一致率告警阈值
}

// DataValidator 数据验证器
// DataValidator data validator
type DataValidator interface {
	Validate(ctx context.Context, key string, cacheValue, sourceValue interface{}) (*ValidationResult, error)
	LoadFromSource(ctx context.Context, key string) (interface{}, error)
}

// ValidationResult 验证结果
// ValidationResult validation result
type ValidationResult struct {
	Key           string    `json:"key"`
	IsConsistent  bool      `json:"is_consistent"`
	CacheValue    string    `json:"cache_value,omitempty"`
	SourceValue   string    `json:"source_value,omitempty"`
	Difference    string    `json:"difference,omitempty"`
	RepairAction  string    `json:"repair_action,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// ConsistencyReport 一致性报告
// ConsistencyReport consistency report
type ConsistencyReport struct {
	TotalChecked     int                  `json:"total_checked"`
	ConsistentCount  int                  `json:"consistent_count"`
	InconsistentKeys []string             `json:"inconsistent_keys"`
	ValidationResults []*ValidationResult `json:"validation_results"`
	ConsistencyRate  float64              `json:"consistency_rate"`
	CheckTime        time.Time            `json:"check_time"`
	Duration         time.Duration        `json:"duration"`
}

// NewConsistencyManager 创建一致性管理器
// NewConsistencyManager creates consistency manager
func NewConsistencyManager(redisClient redis.Client, config *ConsistencyConfig, log logger.Logger) *ConsistencyManager {
	if config == nil {
		config = DefaultConsistencyConfig()
	}

	return &ConsistencyManager{
		redisClient: redisClient,
		logger:      log,
		config:      config,
		validators:  make(map[string]DataValidator),
	}
}

// RegisterValidator 注册数据验证器
// RegisterValidator registers data validator
func (m *ConsistencyManager) RegisterValidator(pattern string, validator DataValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validators[pattern] = validator
}

// CheckConsistency 检查一致性
// CheckConsistency checks consistency
func (m *ConsistencyManager) CheckConsistency(ctx context.Context, keys []string) (*ConsistencyReport, error) {
	startTime := time.Now()
	report := &ConsistencyReport{
		CheckTime:         startTime,
		ValidationResults: make([]*ValidationResult, 0, len(keys)),
		InconsistentKeys:  make([]string, 0),
	}

	for _, key := range keys {
		result, err := m.validateKey(ctx, key)
		if err != nil {
			m.logger.Error("Failed to validate key",
				logger.String("key", key),
				logger.Error(err))
			continue
		}

		report.ValidationResults = append(report.ValidationResults, result)
		report.TotalChecked++

		if result.IsConsistent {
			report.ConsistentCount++
		} else {
			report.InconsistentKeys = append(report.InconsistentKeys, key)
			
			// 如果启用修复，尝试修复
			// If repair is enabled, try to repair
			if m.config.RepairEnabled {
				go m.repairInconsistency(ctx, key, result)
			}
		}
	}

	// 计算一致性率
	// Calculate consistency rate
	if report.TotalChecked > 0 {
		report.ConsistencyRate = float64(report.ConsistentCount) / float64(report.TotalChecked)
	}

	report.Duration = time.Since(startTime)

	// 检查是否需要告警
	// Check if alert is needed
	if report.ConsistencyRate < m.config.AlertThreshold {
		m.logger.Warn("Cache consistency rate below threshold",
			logger.Float64("consistency_rate", report.ConsistencyRate),
			logger.Float64("threshold", m.config.AlertThreshold),
			logger.Int("inconsistent_count", len(report.InconsistentKeys)))
	}

	return report, nil
}

// validateKey 验证单个键
// validateKey validates single key
func (m *ConsistencyManager) validateKey(ctx context.Context, key string) (*ValidationResult, error) {
	result := &ValidationResult{
		Key:       key,
		Timestamp: time.Now(),
	}

	// 获取缓存值
	// Get cache value
	cacheValue, err := m.redisClient.Get(ctx, key)
	if err != nil {
		result.CacheValue = "ERROR: " + err.Error()
	} else {
		result.CacheValue = cacheValue
	}

	// 查找合适的验证器
	// Find appropriate validator
	validator := m.findValidator(key)
	if validator == nil {
		result.IsConsistent = true // 没有验证器时假设一致
		result.RepairAction = "NO_VALIDATOR"
		return result, nil
	}

	// 从数据源加载值
	// Load value from data source
	sourceValue, err := validator.LoadFromSource(ctx, key)
	if err != nil {
		result.SourceValue = "ERROR: " + err.Error()
		result.IsConsistent = false
		result.Difference = "Failed to load from source"
		return result, err
	}

	// 验证一致性
	// Validate consistency
	validationResult, err := validator.Validate(ctx, key, cacheValue, sourceValue)
	if err != nil {
		return result, err
	}

	return validationResult, nil
}

// findValidator 查找验证器
// findValidator finds validator
func (m *ConsistencyManager) findValidator(key string) DataValidator {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 简单的模式匹配，实际应该支持更复杂的匹配
	// Simple pattern matching, should support more complex matching in practice
	for pattern, validator := range m.validators {
		if m.matchPattern(key, pattern) {
			return validator
		}
	}

	return nil
}

// matchPattern 模式匹配
// matchPattern pattern matching
func (m *ConsistencyManager) matchPattern(key, pattern string) bool {
	// 简化的模式匹配，支持前缀匹配
	// Simplified pattern matching, supports prefix matching
	if len(pattern) == 0 {
		return false
	}

	if pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return key == pattern
}

// repairInconsistency 修复不一致
// repairInconsistency repairs inconsistency
func (m *ConsistencyManager) repairInconsistency(ctx context.Context, key string, result *ValidationResult) {
	m.logger.Info("Starting consistency repair",
		logger.String("key", key),
		logger.String("repair_action", result.RepairAction))

	validator := m.findValidator(key)
	if validator == nil {
		m.logger.Warn("No validator found for repair", logger.String("key", key))
		return
	}

	for attempt := 1; attempt <= m.config.MaxRepairRetries; attempt++ {
		// 从数据源重新加载
		// Reload from data source
		sourceValue, err := validator.LoadFromSource(ctx, key)
		if err != nil {
			m.logger.Error("Failed to load from source for repair",
				logger.String("key", key),
				logger.Int("attempt", attempt),
				logger.Error(err))
			
			if attempt < m.config.MaxRepairRetries {
				time.Sleep(m.config.RepairDelay)
				continue
			}
			return
		}

		// 更新缓存
		// Update cache
		err = m.updateCache(ctx, key, sourceValue)
		if err != nil {
			m.logger.Error("Failed to update cache for repair",
				logger.String("key", key),
				logger.Int("attempt", attempt),
				logger.Error(err))
			
			if attempt < m.config.MaxRepairRetries {
				time.Sleep(m.config.RepairDelay)
				continue
			}
			return
		}

		m.logger.Info("Consistency repair completed",
			logger.String("key", key),
			logger.Int("attempt", attempt))
		return
	}
}

// updateCache 更新缓存
// updateCache updates cache
func (m *ConsistencyManager) updateCache(ctx context.Context, key string, value interface{}) error {
	// 这里应该根据键的类型选择合适的更新方式
	// Should choose appropriate update method based on key type
	return m.redisClient.Set(ctx, key, fmt.Sprintf("%v", value), 1*time.Hour)
}

// StartPeriodicCheck 启动定期检查
// StartPeriodicCheck starts periodic check
func (m *ConsistencyManager) StartPeriodicCheck(ctx context.Context, keys []string) {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			report, err := m.CheckConsistency(ctx, keys)
			if err != nil {
				m.logger.Error("Periodic consistency check failed", logger.Error(err))
			} else {
				m.logger.Info("Periodic consistency check completed",
					logger.Int("total_checked", report.TotalChecked),
					logger.Float64("consistency_rate", report.ConsistencyRate),
					logger.Duration("duration", report.Duration))
			}

		case <-ctx.Done():
			return
		}
	}
}

// DefaultConsistencyConfig 默认一致性配置
// DefaultConsistencyConfig default consistency configuration
func DefaultConsistencyConfig() *ConsistencyConfig {
	return &ConsistencyConfig{
		CheckInterval:    5 * time.Minute,
		RepairEnabled:    true,
		MaxRepairRetries: 3,
		RepairDelay:      1 * time.Second,
		AlertThreshold:   0.95, // 95%一致性率
	}
}
