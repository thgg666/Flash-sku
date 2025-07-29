package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Task 任务接口
// Task interface
type Task interface {
	Execute(ctx context.Context) error
}

// TaskFunc 任务函数类型
// TaskFunc task function type
type TaskFunc func(ctx context.Context) error

// Execute 实现Task接口
// Execute implements Task interface
func (f TaskFunc) Execute(ctx context.Context) error {
	return f(ctx)
}

// Pool 工作池结构体
// Pool worker pool structure
type Pool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	metrics     *Metrics
}

// Metrics 工作池指标
// Metrics worker pool metrics
type Metrics struct {
	mu               sync.RWMutex
	TasksSubmitted   int64         `json:"tasks_submitted"`
	TasksCompleted   int64         `json:"tasks_completed"`
	TasksFailed      int64         `json:"tasks_failed"`
	ActiveWorkers    int           `json:"active_workers"`
	QueueLength      int           `json:"queue_length"`
	AvgProcessTime   time.Duration `json:"avg_process_time"`
	totalProcessTime time.Duration
}

// Config 工作池配置
// Config worker pool configuration
type Config struct {
	WorkerCount int `json:"worker_count"`
	QueueSize   int `json:"queue_size"`
}

// NewPool 创建新的工作池
// NewPool creates new worker pool
func NewPool(config *Config) *Pool {
	if config.WorkerCount <= 0 {
		config.WorkerCount = runtime.NumCPU()
	}
	if config.QueueSize <= 0 {
		config.QueueSize = config.WorkerCount * 10
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool{
		workerCount: config.WorkerCount,
		taskQueue:   make(chan Task, config.QueueSize),
		ctx:         ctx,
		cancel:      cancel,
		metrics:     &Metrics{},
	}

	// 启动工作协程
	// Start worker goroutines
	pool.start()

	return pool
}

// start 启动工作协程
// start starts worker goroutines
func (p *Pool) start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	p.metrics.mu.Lock()
	p.metrics.ActiveWorkers = p.workerCount
	p.metrics.mu.Unlock()
}

// worker 工作协程
// worker goroutine
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case task := <-p.taskQueue:
			if task != nil {
				p.executeTask(task)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
// executeTask executes task
func (p *Pool) executeTask(task Task) {
	startTime := time.Now()

	err := task.Execute(p.ctx)

	duration := time.Since(startTime)

	p.metrics.mu.Lock()
	p.metrics.TasksCompleted++
	p.metrics.totalProcessTime += duration
	if p.metrics.TasksCompleted > 0 {
		p.metrics.AvgProcessTime = p.metrics.totalProcessTime / time.Duration(p.metrics.TasksCompleted)
	}
	if err != nil {
		p.metrics.TasksFailed++
	}
	p.metrics.QueueLength = len(p.taskQueue)
	p.metrics.mu.Unlock()
}

// Submit 提交任务
// Submit submits task
func (p *Pool) Submit(task Task) error {
	select {
	case p.taskQueue <- task:
		p.metrics.mu.Lock()
		p.metrics.TasksSubmitted++
		p.metrics.QueueLength = len(p.taskQueue)
		p.metrics.mu.Unlock()
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		return ErrQueueFull
	}
}

// SubmitFunc 提交函数任务
// SubmitFunc submits function task
func (p *Pool) SubmitFunc(fn func(ctx context.Context) error) error {
	return p.Submit(TaskFunc(fn))
}

// Stop 停止工作池
// Stop stops worker pool
func (p *Pool) Stop() {
	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()

	p.metrics.mu.Lock()
	p.metrics.ActiveWorkers = 0
	p.metrics.mu.Unlock()
}

// GetMetrics 获取指标
// GetMetrics gets metrics
func (p *Pool) GetMetrics() Metrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// 返回指标副本
	// Return copy of metrics
	return Metrics{
		TasksSubmitted:   p.metrics.TasksSubmitted,
		TasksCompleted:   p.metrics.TasksCompleted,
		TasksFailed:      p.metrics.TasksFailed,
		ActiveWorkers:    p.metrics.ActiveWorkers,
		QueueLength:      p.metrics.QueueLength,
		AvgProcessTime:   p.metrics.AvgProcessTime,
		totalProcessTime: p.metrics.totalProcessTime,
	}
}

// 错误定义
// Error definitions
var (
	ErrQueueFull = fmt.Errorf("task queue is full")
)
