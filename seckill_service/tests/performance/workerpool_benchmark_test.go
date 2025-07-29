package performance

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/flashsku/seckill/pkg/workerpool"
	"github.com/stretchr/testify/assert"
)

// TestTask 测试任务
// TestTask test task
type TestTask struct {
	duration time.Duration
	result   chan error
}

func (t *TestTask) Execute(ctx context.Context) error {
	// 模拟工作负载
	// Simulate workload
	time.Sleep(t.duration)
	
	select {
	case t.result <- nil:
	default:
	}
	
	return nil
}

func BenchmarkBasicPool(b *testing.B) {
	config := &workerpool.Config{
		WorkerCount: 10,
		QueueSize:   1000,
	}
	
	pool := workerpool.NewPool(config)
	defer pool.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := &TestTask{
				duration: time.Microsecond * 100,
				result:   make(chan error, 1),
			}
			
			err := pool.SubmitFunc(func(ctx context.Context) error {
				return task.Execute(ctx)
			})
			
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkEnhancedPool(b *testing.B) {
	config := workerpool.HighPerformanceConfig()
	config.WorkerCount = 10
	config.QueueSize = 1000
	
	pool, err := workerpool.NewEnhancedPool(config)
	if err != nil {
		b.Fatal(err)
	}
	defer pool.Stop()
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := &TestTask{
				duration: time.Microsecond * 100,
				result:   make(chan error, 1),
			}
			
			err := pool.Submit(task)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func TestEnhancedPoolAutoScaling(t *testing.T) {
	config := workerpool.DefaultPoolConfig()
	config.EnableAutoScale = true
	config.MinWorkers = 2
	config.MaxWorkers = 10
	config.ScaleThreshold = 0.7
	config.MetricsInterval = 100 * time.Millisecond
	
	pool, err := workerpool.NewEnhancedPool(config)
	assert.NoError(t, err)
	defer pool.Stop()
	
	// 提交大量任务触发扩容
	// Submit many tasks to trigger scaling up
	var wg sync.WaitGroup
	taskCount := 50
	
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		task := workerpool.TaskFunc(func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		
		err := pool.Submit(task)
		assert.NoError(t, err)
	}
	
	// 等待一段时间让自动扩容生效
	// Wait for auto scaling to take effect
	time.Sleep(500 * time.Millisecond)
	
	metrics := pool.GetMetrics()
	t.Logf("Metrics after load: Workers=%d, Queue=%d, Submitted=%d, Completed=%d",
		metrics.ActiveWorkers, metrics.QueueLength, metrics.TasksSubmitted, metrics.TasksCompleted)
	
	// 验证扩容效果
	// Verify scaling effect
	assert.Greater(t, int(metrics.ActiveWorkers), config.MinWorkers)
	assert.Greater(t, metrics.TasksSubmitted, int64(0))
	
	// 等待所有任务完成
	// Wait for all tasks to complete
	wg.Wait()
	
	// 等待缩容
	// Wait for scaling down
	time.Sleep(1 * time.Second)
	
	finalMetrics := pool.GetMetrics()
	t.Logf("Final metrics: Workers=%d, Completed=%d, ScaleUp=%d, ScaleDown=%d",
		finalMetrics.ActiveWorkers, finalMetrics.TasksCompleted, 
		finalMetrics.ScaleUpCount, finalMetrics.ScaleDownCount)
}

func TestEnhancedPoolMetrics(t *testing.T) {
	config := workerpool.DefaultPoolConfig()
	config.EnableMetrics = true
	config.MetricsInterval = 50 * time.Millisecond
	
	pool, err := workerpool.NewEnhancedPool(config)
	assert.NoError(t, err)
	defer pool.Stop()
	
	// 提交一些任务
	// Submit some tasks
	taskCount := 10
	var wg sync.WaitGroup
	
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		task := workerpool.TaskFunc(func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		
		err := pool.Submit(task)
		assert.NoError(t, err)
	}
	
	// 等待任务完成
	// Wait for tasks to complete
	wg.Wait()
	
	// 等待指标收集
	// Wait for metrics collection
	time.Sleep(200 * time.Millisecond)
	
	metrics := pool.GetMetrics()
	
	// 验证指标
	// Verify metrics
	assert.Equal(t, int64(taskCount), metrics.TasksSubmitted)
	assert.Equal(t, int64(taskCount), metrics.TasksCompleted)
	assert.Equal(t, int64(0), metrics.TasksFailed)
	assert.Greater(t, metrics.AvgProcessTime, time.Duration(0))
	assert.Greater(t, metrics.MaxProcessTime, time.Duration(0))
	assert.Greater(t, metrics.MinProcessTime, time.Duration(0))
	
	t.Logf("Metrics: Submitted=%d, Completed=%d, Failed=%d, AvgTime=%v",
		metrics.TasksSubmitted, metrics.TasksCompleted, metrics.TasksFailed, metrics.AvgProcessTime)
}

func TestPoolConfigValidation(t *testing.T) {
	// 测试无效配置
	// Test invalid configurations
	invalidConfigs := []*workerpool.PoolConfig{
		{WorkerCount: 0, QueueSize: 10},
		{WorkerCount: 10, QueueSize: 0},
		{WorkerCount: 10, QueueSize: 10, EnableAutoScale: true, MinWorkers: 0},
		{WorkerCount: 10, QueueSize: 10, EnableAutoScale: true, MinWorkers: 5, MaxWorkers: 3},
		{WorkerCount: 10, QueueSize: 10, TaskTimeout: 0},
	}
	
	for i, config := range invalidConfigs {
		err := config.Validate()
		assert.Error(t, err, "Config %d should be invalid", i)
	}
	
	// 测试有效配置
	// Test valid configuration
	validConfig := workerpool.DefaultPoolConfig()
	err := validConfig.Validate()
	assert.NoError(t, err)
}

func TestPoolConfigOptimization(t *testing.T) {
	config := &workerpool.PoolConfig{}
	config.OptimizeForCPU()
	
	assert.Greater(t, config.WorkerCount, 0)
	assert.Greater(t, config.QueueSize, 0)
	assert.Equal(t, config.QueueSize, config.WorkerCount*10)
}
