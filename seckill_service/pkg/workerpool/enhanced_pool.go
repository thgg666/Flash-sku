package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// EnhancedPool 增强的工作池
// EnhancedPool enhanced worker pool
type EnhancedPool struct {
	config      *PoolConfig
	taskQueue   chan Task
	workers     []*Worker
	workerCount int32
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	metrics     *EnhancedMetrics
	mu          sync.RWMutex
	running     bool
}

// Worker 工作协程
// Worker goroutine
type Worker struct {
	id       int
	pool     *EnhancedPool
	taskChan chan Task
	quit     chan bool
	lastUsed time.Time
}

// EnhancedMetrics 增强的指标
// EnhancedMetrics enhanced metrics
type EnhancedMetrics struct {
	mu               sync.RWMutex
	TasksSubmitted   int64         `json:"tasks_submitted"`
	TasksCompleted   int64         `json:"tasks_completed"`
	TasksFailed      int64         `json:"tasks_failed"`
	TasksTimeout     int64         `json:"tasks_timeout"`
	ActiveWorkers    int32         `json:"active_workers"`
	IdleWorkers      int32         `json:"idle_workers"`
	QueueLength      int           `json:"queue_length"`
	AvgProcessTime   time.Duration `json:"avg_process_time"`
	MaxProcessTime   time.Duration `json:"max_process_time"`
	MinProcessTime   time.Duration `json:"min_process_time"`
	TotalProcessTime time.Duration `json:"total_process_time"`
	LastScaleTime    time.Time     `json:"last_scale_time"`
	ScaleUpCount     int64         `json:"scale_up_count"`
	ScaleDownCount   int64         `json:"scale_down_count"`
}

// NewEnhancedPool 创建增强的工作池
// NewEnhancedPool creates enhanced worker pool
func NewEnhancedPool(config *PoolConfig) (*EnhancedPool, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	config.OptimizeForCPU()

	ctx, cancel := context.WithCancel(context.Background())

	pool := &EnhancedPool{
		config:    config,
		taskQueue: make(chan Task, config.QueueSize),
		workers:   make([]*Worker, 0, config.MaxWorkers),
		ctx:       ctx,
		cancel:    cancel,
		metrics:   &EnhancedMetrics{},
		running:   true,
	}

	// 启动初始工作协程
	// Start initial workers
	for i := 0; i < config.WorkerCount; i++ {
		pool.addWorker()
	}

	// 启动监控协程
	// Start monitoring goroutines
	if config.EnableMetrics {
		go pool.metricsCollector()
	}

	if config.EnableAutoScale {
		go pool.autoScaler()
	}

	go pool.healthChecker()

	return pool, nil
}

// addWorker 添加工作协程
// addWorker adds worker goroutine
func (p *EnhancedPool) addWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.workers) >= p.config.MaxWorkers {
		return
	}

	worker := &Worker{
		id:       len(p.workers),
		pool:     p,
		taskChan: make(chan Task, 1),
		quit:     make(chan bool),
		lastUsed: time.Now(),
	}

	p.workers = append(p.workers, worker)
	atomic.AddInt32(&p.workerCount, 1)
	atomic.AddInt32(&p.metrics.ActiveWorkers, 1)

	p.wg.Add(1)
	go worker.start()
}

// removeWorker 移除工作协程
// removeWorker removes worker goroutine
func (p *EnhancedPool) removeWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.workers) <= p.config.MinWorkers {
		return
	}

	// 找到最久未使用的工作协程
	// Find the least recently used worker
	oldestIndex := 0
	oldestTime := p.workers[0].lastUsed

	for i, worker := range p.workers {
		if worker.lastUsed.Before(oldestTime) {
			oldestIndex = i
			oldestTime = worker.lastUsed
		}
	}

	// 停止工作协程
	// Stop worker
	worker := p.workers[oldestIndex]
	close(worker.quit)

	// 从切片中移除
	// Remove from slice
	p.workers = append(p.workers[:oldestIndex], p.workers[oldestIndex+1:]...)
	atomic.AddInt32(&p.workerCount, -1)
	atomic.AddInt32(&p.metrics.ActiveWorkers, -1)
}

// start 启动工作协程
// start starts worker goroutine
func (w *Worker) start() {
	defer w.pool.wg.Done()

	for {
		select {
		case task := <-w.taskChan:
			w.executeTask(task)
		case <-w.quit:
			return
		case <-w.pool.ctx.Done():
			return
		}
	}
}

// executeTask 执行任务
// executeTask executes task
func (w *Worker) executeTask(task Task) {
	startTime := time.Now()
	w.lastUsed = startTime

	// 创建带超时的上下文
	// Create context with timeout
	ctx, cancel := context.WithTimeout(w.pool.ctx, w.pool.config.TaskTimeout)
	defer cancel()

	// 执行任务
	// Execute task
	err := task.Execute(ctx)
	duration := time.Since(startTime)

	// 更新指标
	// Update metrics
	w.pool.updateMetrics(duration, err)
}

// updateMetrics 更新指标
// updateMetrics updates metrics
func (p *EnhancedPool) updateMetrics(duration time.Duration, err error) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.TasksCompleted++
	p.metrics.TotalProcessTime += duration

	if err != nil {
		p.metrics.TasksFailed++
	}

	// 更新处理时间统计
	// Update processing time statistics
	if p.metrics.TasksCompleted == 1 {
		p.metrics.MinProcessTime = duration
		p.metrics.MaxProcessTime = duration
	} else {
		if duration < p.metrics.MinProcessTime {
			p.metrics.MinProcessTime = duration
		}
		if duration > p.metrics.MaxProcessTime {
			p.metrics.MaxProcessTime = duration
		}
	}

	p.metrics.AvgProcessTime = p.metrics.TotalProcessTime / time.Duration(p.metrics.TasksCompleted)
	p.metrics.QueueLength = len(p.taskQueue)
}

// Submit 提交任务
// Submit submits task
func (p *EnhancedPool) Submit(task Task) error {
	if !p.running {
		return ErrPoolClosed
	}

	select {
	case p.taskQueue <- task:
		atomic.AddInt64(&p.metrics.TasksSubmitted, 1)

		// 分发任务给空闲的工作协程
		// Distribute task to idle worker
		p.distributeTask(task)
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
		return ErrQueueFull
	}
}

// distributeTask 分发任务
// distributeTask distributes task
func (p *EnhancedPool) distributeTask(task Task) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// 找到空闲的工作协程
	// Find idle worker
	for _, worker := range p.workers {
		select {
		case worker.taskChan <- task:
			return
		default:
			continue
		}
	}
}

// GetMetrics 获取指标
// GetMetrics gets metrics
func (p *EnhancedPool) GetMetrics() EnhancedMetrics {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// 创建副本避免复制锁
	// Create copy to avoid copying lock
	metrics := EnhancedMetrics{
		TasksSubmitted:   p.metrics.TasksSubmitted,
		TasksCompleted:   p.metrics.TasksCompleted,
		TasksFailed:      p.metrics.TasksFailed,
		TasksTimeout:     p.metrics.TasksTimeout,
		ActiveWorkers:    atomic.LoadInt32(&p.workerCount),
		IdleWorkers:      p.metrics.IdleWorkers,
		QueueLength:      len(p.taskQueue),
		AvgProcessTime:   p.metrics.AvgProcessTime,
		MaxProcessTime:   p.metrics.MaxProcessTime,
		MinProcessTime:   p.metrics.MinProcessTime,
		TotalProcessTime: p.metrics.TotalProcessTime,
		LastScaleTime:    p.metrics.LastScaleTime,
		ScaleUpCount:     p.metrics.ScaleUpCount,
		ScaleDownCount:   p.metrics.ScaleDownCount,
	}

	return metrics
}

// metricsCollector 指标收集器
// metricsCollector metrics collector
func (p *EnhancedPool) metricsCollector() {
	ticker := time.NewTicker(p.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 收集指标
			// Collect metrics
			p.collectMetrics()
		case <-p.ctx.Done():
			return
		}
	}
}

// autoScaler 自动扩缩容器
// autoScaler auto scaler
func (p *EnhancedPool) autoScaler() {
	ticker := time.NewTicker(p.config.MetricsInterval * 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkAndScale()
		case <-p.ctx.Done():
			return
		}
	}
}

// healthChecker 健康检查器
// healthChecker health checker
func (p *EnhancedPool) healthChecker() {
	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.performHealthCheck()
		case <-p.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集指标
// collectMetrics collects metrics
func (p *EnhancedPool) collectMetrics() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	// 更新队列长度
	// Update queue length
	p.metrics.QueueLength = len(p.taskQueue)

	// 计算空闲工作协程数
	// Calculate idle workers
	activeWorkers := atomic.LoadInt32(&p.workerCount)
	p.metrics.ActiveWorkers = activeWorkers

	// 简单估算空闲工作协程数
	// Simple estimation of idle workers
	if p.metrics.QueueLength == 0 {
		p.metrics.IdleWorkers = activeWorkers
	} else {
		p.metrics.IdleWorkers = 0
	}
}

// checkAndScale 检查并扩缩容
// checkAndScale checks and scales
func (p *EnhancedPool) checkAndScale() {
	metrics := p.GetMetrics()

	// 计算负载率
	// Calculate load ratio
	loadRatio := float64(metrics.QueueLength) / float64(p.config.QueueSize)

	currentWorkers := int(atomic.LoadInt32(&p.workerCount))

	// 扩容条件
	// Scale up condition
	if loadRatio > p.config.ScaleThreshold && currentWorkers < p.config.MaxWorkers {
		p.addWorker()
		p.metrics.mu.Lock()
		p.metrics.ScaleUpCount++
		p.metrics.LastScaleTime = time.Now()
		p.metrics.mu.Unlock()
	}

	// 缩容条件
	// Scale down condition
	if loadRatio < p.config.ScaleThreshold/2 && currentWorkers > p.config.MinWorkers {
		p.removeWorker()
		p.metrics.mu.Lock()
		p.metrics.ScaleDownCount++
		p.metrics.LastScaleTime = time.Now()
		p.metrics.mu.Unlock()
	}
}

// performHealthCheck 执行健康检查
// performHealthCheck performs health check
func (p *EnhancedPool) performHealthCheck() {
	// 检查工作协程是否正常
	// Check if workers are healthy
	p.mu.RLock()
	workerCount := len(p.workers)
	p.mu.RUnlock()

	expectedWorkers := int(atomic.LoadInt32(&p.workerCount))

	// 如果工作协程数量不匹配，尝试修复
	// If worker count doesn't match, try to fix
	if workerCount != expectedWorkers {
		// 简单的修复策略：重新同步计数
		// Simple fix strategy: resync count
		atomic.StoreInt32(&p.workerCount, int32(workerCount))
	}
}

// Stop 停止工作池
// Stop stops worker pool
func (p *EnhancedPool) Stop() {
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()

	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
}

// 错误定义
// Error definitions
var (
	ErrPoolClosed = fmt.Errorf("worker pool is closed")
)
