package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
)

// ShutdownFunc 关闭函数类型
// ShutdownFunc shutdown function type
type ShutdownFunc func(ctx context.Context) error

// Manager 优雅关闭管理器
// Manager graceful shutdown manager
type Manager struct {
	logger        logger.Logger
	shutdownFuncs []ShutdownFunc
	timeout       time.Duration
	signals       []os.Signal
	mu            sync.RWMutex
}

// Config 优雅关闭配置
// Config graceful shutdown configuration
type Config struct {
	Timeout time.Duration `json:"timeout"` // 关闭超时时间
	Signals []os.Signal   `json:"-"`       // 监听的信号
}

// NewManager 创建新的优雅关闭管理器
// NewManager creates new graceful shutdown manager
func NewManager(config *Config, logger logger.Logger) *Manager {
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if len(config.Signals) == 0 {
		config.Signals = []os.Signal{
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // 终止信号
			syscall.SIGQUIT, // 退出信号
		}
	}

	return &Manager{
		logger:        logger,
		shutdownFuncs: make([]ShutdownFunc, 0),
		timeout:       config.Timeout,
		signals:       config.Signals,
	}
}

// Register 注册关闭函数
// Register registers shutdown function
func (m *Manager) Register(fn ShutdownFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownFuncs = append(m.shutdownFuncs, fn)
}

// RegisterMultiple 注册多个关闭函数
// RegisterMultiple registers multiple shutdown functions
func (m *Manager) RegisterMultiple(fns ...ShutdownFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownFuncs = append(m.shutdownFuncs, fns...)
}

// Wait 等待关闭信号
// Wait waits for shutdown signal
func (m *Manager) Wait() {
	// 创建信号通道
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, m.signals...)

	// 等待信号
	// Wait for signal
	sig := <-sigChan
	m.logger.Info("Received shutdown signal",
		logger.String("signal", sig.String()))

	// 执行优雅关闭
	// Execute graceful shutdown
	m.shutdown()
}

// Shutdown 立即执行关闭
// Shutdown executes shutdown immediately
func (m *Manager) Shutdown() {
	m.shutdown()
}

// shutdown 内部关闭方法
// shutdown internal shutdown method
func (m *Manager) shutdown() {
	m.logger.Info("Starting graceful shutdown...",
		logger.Int("shutdown_functions", len(m.shutdownFuncs)),
		logger.Duration("timeout", m.timeout))

	// 创建带超时的上下文
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	// 并发执行所有关闭函数
	// Execute all shutdown functions concurrently
	var wg sync.WaitGroup
	errorChan := make(chan error, len(m.shutdownFuncs))

	m.mu.RLock()
	for i, fn := range m.shutdownFuncs {
		wg.Add(1)
		go func(index int, shutdownFunc ShutdownFunc) {
			defer wg.Done()
			
			m.logger.Debug("Executing shutdown function",
				logger.Int("index", index))
			
			if err := shutdownFunc(ctx); err != nil {
				m.logger.Error("Shutdown function failed",
					logger.Int("index", index),
					logger.Error(err))
				errorChan <- err
			} else {
				m.logger.Debug("Shutdown function completed",
					logger.Int("index", index))
			}
		}(i, fn)
	}
	m.mu.RUnlock()

	// 等待所有关闭函数完成或超时
	// Wait for all shutdown functions to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("All shutdown functions completed successfully")
	case <-ctx.Done():
		m.logger.Warn("Shutdown timeout reached, forcing exit",
			logger.Duration("timeout", m.timeout))
	}

	// 收集错误
	// Collect errors
	close(errorChan)
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		m.logger.Error("Some shutdown functions failed",
			logger.Int("error_count", len(errors)))
	}

	m.logger.Info("Graceful shutdown completed")
}

// WaitWithCallback 等待关闭信号并执行回调
// WaitWithCallback waits for shutdown signal and executes callback
func (m *Manager) WaitWithCallback(callback func()) {
	// 创建信号通道
	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, m.signals...)

	// 等待信号
	// Wait for signal
	sig := <-sigChan
	m.logger.Info("Received shutdown signal",
		logger.String("signal", sig.String()))

	// 执行回调
	// Execute callback
	if callback != nil {
		callback()
	}

	// 执行优雅关闭
	// Execute graceful shutdown
	m.shutdown()
}

// GetTimeout 获取超时时间
// GetTimeout gets timeout duration
func (m *Manager) GetTimeout() time.Duration {
	return m.timeout
}

// GetRegisteredCount 获取已注册的关闭函数数量
// GetRegisteredCount gets count of registered shutdown functions
func (m *Manager) GetRegisteredCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.shutdownFuncs)
}
