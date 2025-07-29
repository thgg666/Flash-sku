package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/flashsku/seckill/internal/config"
	"github.com/flashsku/seckill/internal/handler"
	"github.com/flashsku/seckill/internal/middleware"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/rabbitmq"
	"github.com/flashsku/seckill/pkg/redis"
	"github.com/flashsku/seckill/pkg/workerpool"
)

// App 应用程序结构体
// App application structure
type App struct {
	config       *config.Config
	server       *http.Server
	redisClient  redis.Client
	mqClient     rabbitmq.Client
	logger       logger.Logger
	workerPool   *workerpool.Pool
	enhancedPool *workerpool.EnhancedPool
}

// New 创建新的应用程序实例
// New creates new application instance
func New() (*App, error) {
	// 加载配置
	// Load configuration
	cfg := config.LoadConfig()

	// 初始化日志器
	// Initialize logger
	loggerConfig := &logger.Config{
		Level:  logger.INFO,
		Format: "text",
		Output: "stdout",
	}
	appLogger := logger.NewLogger(loggerConfig)

	// 初始化Redis客户端
	// Initialize Redis client
	redisConfig := &redis.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		Database:     cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	redisClient, err := redis.NewClient(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// 初始化RabbitMQ客户端
	// Initialize RabbitMQ client
	mqConfig := &rabbitmq.Config{
		URL:          cfg.RabbitMQ.URL,
		Exchange:     cfg.RabbitMQ.Exchange,
		ExchangeType: "direct",
		Queue:        cfg.RabbitMQ.Queue,
		RoutingKey:   cfg.RabbitMQ.RoutingKey,
		Durable:      true,
		AutoDelete:   false,
		PrefetchSize: cfg.RabbitMQ.PrefetchSize,
	}

	mqClient, err := rabbitmq.NewClientImpl(mqConfig)
	if err != nil {
		appLogger.Warn("Failed to initialize RabbitMQ client, continuing without it",
			logger.Error(err))
		mqClient = nil // 允许在没有RabbitMQ的情况下运行
	}

	// 初始化工作池
	// Initialize worker pool
	workerPoolConfig := &workerpool.Config{
		WorkerCount: cfg.Seckill.WorkerPoolSize,
		QueueSize:   cfg.Seckill.WorkerPoolSize * 10,
	}
	workerPool := workerpool.NewPool(workerPoolConfig)

	// 初始化增强工作池
	// Initialize enhanced worker pool
	enhancedPoolConfig := workerpool.HighPerformanceConfig()
	enhancedPoolConfig.WorkerCount = cfg.Seckill.WorkerPoolSize
	enhancedPool, err := workerpool.NewEnhancedPool(enhancedPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced worker pool: %w", err)
	}

	app := &App{
		config:       cfg,
		redisClient:  redisClient,
		mqClient:     mqClient,
		logger:       appLogger,
		workerPool:   workerPool,
		enhancedPool: enhancedPool,
	}

	// 初始化HTTP服务器
	// Initialize HTTP server
	if err := app.initServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	return app, nil
}

// initServer 初始化HTTP服务器
// initServer initializes HTTP server
func (a *App) initServer() error {
	// 设置Gin模式
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由器
	// Create Gin router
	router := gin.New()

	// 添加中间件
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// 初始化处理器
	// Initialize handlers
	seckillHandler := handler.NewSeckillHandler()

	// 注册路由
	// Register routes
	a.registerRoutes(router, seckillHandler)

	// 创建HTTP服务器
	// Create HTTP server
	a.server = &http.Server{
		Addr:         ":" + a.config.Server.Port,
		Handler:      router,
		ReadTimeout:  a.config.Server.ReadTimeout,
		WriteTimeout: a.config.Server.WriteTimeout,
		IdleTimeout:  a.config.Server.IdleTimeout,
	}

	a.logger.Info("HTTP server initialized",
		logger.String("port", a.config.Server.Port))

	return nil
}

// registerRoutes 注册路由
// registerRoutes registers routes
func (a *App) registerRoutes(router *gin.Engine, seckillHandler *handler.SeckillHandler) {
	// 健康检查
	// Health check
	router.GET("/health", seckillHandler.HealthCheck)

	// 秒杀相关路由
	// Seckill related routes
	seckillGroup := router.Group("/seckill")
	{
		seckillGroup.POST("/:activity_id", seckillHandler.HandleSeckill)
		seckillGroup.GET("/stock/:activity_id", seckillHandler.GetStock)
		seckillGroup.GET("/metrics", seckillHandler.GetMetrics)
	}

	a.logger.Info("Routes registered successfully")
}

// Start 启动应用程序
// Start starts the application
func (a *App) Start() error {
	a.logger.Info("Starting seckill service",
		logger.String("port", a.config.Server.Port))

	// 启动HTTP服务器
	// Start HTTP server
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop 停止应用程序
// Stop stops the application
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("Stopping seckill service...")

	// 关闭HTTP服务器
	// Shutdown HTTP server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to shutdown server gracefully", logger.Error(err))
		return err
	}

	// 关闭Redis连接
	// Close Redis connection
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("Failed to close Redis client", logger.Error(err))
		}
	}

	// 关闭RabbitMQ连接
	// Close RabbitMQ connection
	if a.mqClient != nil {
		if err := a.mqClient.Close(); err != nil {
			a.logger.Error("Failed to close RabbitMQ client", logger.Error(err))
		}
	}

	// 关闭工作池
	// Close worker pool
	if a.workerPool != nil {
		a.workerPool.Stop()
		a.logger.Info("Worker pool stopped")
	}

	// 关闭增强工作池
	// Close enhanced worker pool
	if a.enhancedPool != nil {
		a.enhancedPool.Stop()
		a.logger.Info("Enhanced worker pool stopped")
	}

	a.logger.Info("Seckill service stopped successfully")
	return nil
}

// GetRedisClient 获取Redis客户端
// GetRedisClient gets Redis client
func (a *App) GetRedisClient() redis.Client {
	return a.redisClient
}

// GetMQClient 获取RabbitMQ客户端
// GetMQClient gets RabbitMQ client
func (a *App) GetMQClient() rabbitmq.Client {
	return a.mqClient
}

// GetLogger 获取日志器
// GetLogger gets logger
func (a *App) GetLogger() logger.Logger {
	return a.logger
}

// GetWorkerPool 获取工作池
// GetWorkerPool gets worker pool
func (a *App) GetWorkerPool() *workerpool.Pool {
	return a.workerPool
}

// GetEnhancedPool 获取增强工作池
// GetEnhancedPool gets enhanced worker pool
func (a *App) GetEnhancedPool() *workerpool.EnhancedPool {
	return a.enhancedPool
}
