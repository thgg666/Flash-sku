package workerpool

import (
	"fmt"
	"runtime"
	"time"
)

// PoolConfig 工作池配置
// PoolConfig worker pool configuration
type PoolConfig struct {
	// 基础配置 / Basic configuration
	WorkerCount int `json:"worker_count"` // 工作协程数量
	QueueSize   int `json:"queue_size"`   // 任务队列大小

	// 性能配置 / Performance configuration
	MaxIdleTime     time.Duration `json:"max_idle_time"`    // 最大空闲时间
	TaskTimeout     time.Duration `json:"task_timeout"`     // 任务超时时间
	ShutdownTimeout time.Duration `json:"shutdown_timeout"` // 关闭超时时间

	// 监控配置 / Monitoring configuration
	EnableMetrics       bool          `json:"enable_metrics"`        // 启用指标收集
	MetricsInterval     time.Duration `json:"metrics_interval"`      // 指标收集间隔
	HealthCheckInterval time.Duration `json:"health_check_interval"` // 健康检查间隔

	// 自适应配置 / Adaptive configuration
	EnableAutoScale bool    `json:"enable_auto_scale"` // 启用自动扩缩容
	MinWorkers      int     `json:"min_workers"`       // 最小工作协程数
	MaxWorkers      int     `json:"max_workers"`       // 最大工作协程数
	ScaleThreshold  float64 `json:"scale_threshold"`   // 扩缩容阈值
}

// DefaultPoolConfig 默认工作池配置
// DefaultPoolConfig default worker pool configuration
func DefaultPoolConfig() *PoolConfig {
	cpuCount := runtime.NumCPU()

	return &PoolConfig{
		// 基础配置
		WorkerCount: cpuCount * 2,
		QueueSize:   cpuCount * 20,

		// 性能配置
		MaxIdleTime:     30 * time.Second,
		TaskTimeout:     10 * time.Second,
		ShutdownTimeout: 30 * time.Second,

		// 监控配置
		EnableMetrics:       true,
		MetricsInterval:     5 * time.Second,
		HealthCheckInterval: 10 * time.Second,

		// 自适应配置
		EnableAutoScale: false,
		MinWorkers:      cpuCount,
		MaxWorkers:      cpuCount * 4,
		ScaleThreshold:  0.8,
	}
}

// HighPerformanceConfig 高性能配置
// HighPerformanceConfig high performance configuration
func HighPerformanceConfig() *PoolConfig {
	cpuCount := runtime.NumCPU()

	return &PoolConfig{
		// 基础配置 - 更多工作协程
		WorkerCount: cpuCount * 4,
		QueueSize:   cpuCount * 50,

		// 性能配置 - 更短的超时时间
		MaxIdleTime:     10 * time.Second,
		TaskTimeout:     5 * time.Second,
		ShutdownTimeout: 15 * time.Second,

		// 监控配置 - 更频繁的监控
		EnableMetrics:       true,
		MetricsInterval:     1 * time.Second,
		HealthCheckInterval: 5 * time.Second,

		// 自适应配置 - 启用自动扩缩容
		EnableAutoScale: true,
		MinWorkers:      cpuCount * 2,
		MaxWorkers:      cpuCount * 8,
		ScaleThreshold:  0.7,
	}
}

// LowResourceConfig 低资源配置
// LowResourceConfig low resource configuration
func LowResourceConfig() *PoolConfig {
	cpuCount := runtime.NumCPU()

	return &PoolConfig{
		// 基础配置 - 较少的工作协程
		WorkerCount: cpuCount,
		QueueSize:   cpuCount * 10,

		// 性能配置 - 较长的超时时间
		MaxIdleTime:     60 * time.Second,
		TaskTimeout:     30 * time.Second,
		ShutdownTimeout: 60 * time.Second,

		// 监控配置 - 较少的监控
		EnableMetrics:       true,
		MetricsInterval:     10 * time.Second,
		HealthCheckInterval: 30 * time.Second,

		// 自适应配置 - 禁用自动扩缩容
		EnableAutoScale: false,
		MinWorkers:      1,
		MaxWorkers:      cpuCount * 2,
		ScaleThreshold:  0.9,
	}
}

// Validate 验证配置
// Validate validates configuration
func (c *PoolConfig) Validate() error {
	if c.WorkerCount <= 0 {
		return fmt.Errorf("worker count must be positive, got %d", c.WorkerCount)
	}

	if c.QueueSize <= 0 {
		return fmt.Errorf("queue size must be positive, got %d", c.QueueSize)
	}

	if c.EnableAutoScale {
		if c.MinWorkers <= 0 {
			return fmt.Errorf("min workers must be positive when auto scale enabled, got %d", c.MinWorkers)
		}

		if c.MaxWorkers <= c.MinWorkers {
			return fmt.Errorf("max workers (%d) must be greater than min workers (%d)", c.MaxWorkers, c.MinWorkers)
		}

		if c.ScaleThreshold <= 0 || c.ScaleThreshold >= 1 {
			return fmt.Errorf("scale threshold must be between 0 and 1, got %f", c.ScaleThreshold)
		}
	}

	if c.TaskTimeout <= 0 {
		return fmt.Errorf("task timeout must be positive, got %v", c.TaskTimeout)
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive, got %v", c.ShutdownTimeout)
	}

	return nil
}

// OptimizeForCPU 根据CPU数量优化配置
// OptimizeForCPU optimizes configuration based on CPU count
func (c *PoolConfig) OptimizeForCPU() {
	cpuCount := runtime.NumCPU()

	// 根据CPU数量调整工作协程数
	// Adjust worker count based on CPU count
	if c.WorkerCount == 0 {
		c.WorkerCount = cpuCount * 2
	}

	// 根据工作协程数调整队列大小
	// Adjust queue size based on worker count
	if c.QueueSize == 0 {
		c.QueueSize = c.WorkerCount * 10
	}

	// 调整自动扩缩容参数
	// Adjust auto scaling parameters
	if c.EnableAutoScale {
		if c.MinWorkers == 0 {
			c.MinWorkers = cpuCount
		}
		if c.MaxWorkers == 0 {
			c.MaxWorkers = cpuCount * 4
		}
	}
}

// Clone 克隆配置
// Clone clones configuration
func (c *PoolConfig) Clone() *PoolConfig {
	return &PoolConfig{
		WorkerCount:         c.WorkerCount,
		QueueSize:           c.QueueSize,
		MaxIdleTime:         c.MaxIdleTime,
		TaskTimeout:         c.TaskTimeout,
		ShutdownTimeout:     c.ShutdownTimeout,
		EnableMetrics:       c.EnableMetrics,
		MetricsInterval:     c.MetricsInterval,
		HealthCheckInterval: c.HealthCheckInterval,
		EnableAutoScale:     c.EnableAutoScale,
		MinWorkers:          c.MinWorkers,
		MaxWorkers:          c.MaxWorkers,
		ScaleThreshold:      c.ScaleThreshold,
	}
}
