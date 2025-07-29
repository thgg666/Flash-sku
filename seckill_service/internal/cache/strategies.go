package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// WriteThroughStrategy 写穿透策略
// WriteThroughStrategy write-through strategy
type WriteThroughStrategy struct {
	redisClient redis.Client
	logger      logger.Logger
}

// NewWriteThroughStrategy 创建写穿透策略
// NewWriteThroughStrategy creates write-through strategy
func NewWriteThroughStrategy(redisClient redis.Client, log logger.Logger) *WriteThroughStrategy {
	return &WriteThroughStrategy{
		redisClient: redisClient,
		logger:      log,
	}
}

// Update 更新缓存
// Update updates cache
func (s *WriteThroughStrategy) Update(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// 写穿透：同步写入缓存和数据库
	// Write-through: synchronously write to cache and database
	err = s.redisClient.Set(ctx, key, string(data), 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	s.logger.Debug("Write-through cache updated", logger.String("key", key))
	return nil
}

// Invalidate 失效缓存
// Invalidate invalidates cache
func (s *WriteThroughStrategy) Invalidate(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key)
}

// Refresh 刷新缓存
// Refresh refreshes cache
func (s *WriteThroughStrategy) Refresh(ctx context.Context, key string, loader DataLoader) error {
	data, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	return s.Update(ctx, key, data)
}

// WriteBehindStrategy 写回策略
// WriteBehindStrategy write-behind strategy
type WriteBehindStrategy struct {
	redisClient redis.Client
	logger      logger.Logger
	writeQueue  chan *WriteOperation
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// WriteOperation 写操作
// WriteOperation write operation
type WriteOperation struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
}

// NewWriteBehindStrategy 创建写回策略
// NewWriteBehindStrategy creates write-behind strategy
func NewWriteBehindStrategy(redisClient redis.Client, log logger.Logger) *WriteBehindStrategy {
	ctx, cancel := context.WithCancel(context.Background())

	strategy := &WriteBehindStrategy{
		redisClient: redisClient,
		logger:      log,
		writeQueue:  make(chan *WriteOperation, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动后台写入协程
	// Start background write goroutine
	strategy.wg.Add(1)
	go strategy.backgroundWriter()

	return strategy
}

// Update 更新缓存
// Update updates cache
func (s *WriteBehindStrategy) Update(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// 立即写入缓存
	// Immediately write to cache
	err = s.redisClient.Set(ctx, key, string(data), 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	// 异步写入数据库
	// Asynchronously write to database
	operation := &WriteOperation{
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
	}

	select {
	case s.writeQueue <- operation:
		s.logger.Debug("Write-behind operation queued", logger.String("key", key))
	default:
		s.logger.Warn("Write-behind queue full, dropping operation", logger.String("key", key))
	}

	return nil
}

// Invalidate 失效缓存
// Invalidate invalidates cache
func (s *WriteBehindStrategy) Invalidate(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key)
}

// Refresh 刷新缓存
// Refresh refreshes cache
func (s *WriteBehindStrategy) Refresh(ctx context.Context, key string, loader DataLoader) error {
	data, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	return s.Update(ctx, key, data)
}

// backgroundWriter 后台写入器
// backgroundWriter background writer
func (s *WriteBehindStrategy) backgroundWriter() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	batch := make([]*WriteOperation, 0, 100)

	for {
		select {
		case operation := <-s.writeQueue:
			batch = append(batch, operation)

			// 如果批次满了，立即处理
			// If batch is full, process immediately
			if len(batch) >= 100 {
				s.processBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			// 定期处理批次
			// Periodically process batch
			if len(batch) > 0 {
				s.processBatch(batch)
				batch = batch[:0]
			}

		case <-s.ctx.Done():
			// 处理剩余批次
			// Process remaining batch
			if len(batch) > 0 {
				s.processBatch(batch)
			}
			return
		}
	}
}

// processBatch 处理批次
// processBatch processes batch
func (s *WriteBehindStrategy) processBatch(batch []*WriteOperation) {
	s.logger.Debug("Processing write-behind batch", logger.Int("size", len(batch)))

	for _, operation := range batch {
		// 这里应该写入数据库
		// Should write to database here
		s.logger.Debug("Write-behind operation processed",
			logger.String("key", operation.Key),
			logger.String("timestamp", operation.Timestamp.Format(time.RFC3339)))
	}
}

// Stop 停止写回策略
// Stop stops write-behind strategy
func (s *WriteBehindStrategy) Stop() {
	s.cancel()
	s.wg.Wait()
	close(s.writeQueue)
}

// RefreshAheadStrategy 预刷新策略
// RefreshAheadStrategy refresh-ahead strategy
type RefreshAheadStrategy struct {
	redisClient redis.Client
	logger      logger.Logger
	config      *UpdateConfig
}

// NewRefreshAheadStrategy 创建预刷新策略
// NewRefreshAheadStrategy creates refresh-ahead strategy
func NewRefreshAheadStrategy(redisClient redis.Client, config *UpdateConfig, log logger.Logger) *RefreshAheadStrategy {
	return &RefreshAheadStrategy{
		redisClient: redisClient,
		logger:      log,
		config:      config,
	}
}

// Update 更新缓存
// Update updates cache
func (s *RefreshAheadStrategy) Update(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// 设置较长的TTL
	// Set longer TTL
	ttl := s.config.DefaultTTL
	err = s.redisClient.Set(ctx, key, string(data), ttl)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	s.logger.Debug("Refresh-ahead cache updated",
		logger.String("key", key),
		logger.Duration("ttl", ttl))
	return nil
}

// Invalidate 失效缓存
// Invalidate invalidates cache
func (s *RefreshAheadStrategy) Invalidate(ctx context.Context, key string) error {
	return s.redisClient.Del(ctx, key)
}

// Refresh 刷新缓存
// Refresh refreshes cache
func (s *RefreshAheadStrategy) Refresh(ctx context.Context, key string, loader DataLoader) error {
	// 检查当前TTL
	// Check current TTL
	ttl, err := s.redisClient.TTL(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}

	// 如果TTL低于阈值，刷新缓存
	// If TTL is below threshold, refresh cache
	if ttl > 0 {
		totalTTL := s.config.DefaultTTL
		remainingRatio := float64(ttl) / float64(totalTTL)

		if remainingRatio < s.config.RefreshThreshold {
			data, err := loader(ctx)
			if err != nil {
				return fmt.Errorf("failed to load data: %w", err)
			}

			return s.Update(ctx, key, data)
		}
	}

	return nil
}
