package api

import (
	"net/http"
	"time"

	"github.com/flashsku/seckill/internal/activity"
	"github.com/flashsku/seckill/internal/cache"
	"github.com/flashsku/seckill/internal/seckill"
	"github.com/flashsku/seckill/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RouterConfig 路由配置
// RouterConfig router configuration
type RouterConfig struct {
	EnableAuth      bool `json:"enable_auth"`
	EnableRateLimit bool `json:"enable_rate_limit"`
	EnableMetrics   bool `json:"enable_metrics"`
	EnableCORS      bool `json:"enable_cors"`
}

// Router API路由器
// Router API router
type Router struct {
	engine         *gin.Engine
	config         *RouterConfig
	logger         logger.Logger
	seckillHandler *SeckillHandler
	metricsAPI     *cache.MetricsAPI
}

// NewRouter 创建路由器
// NewRouter creates router
func NewRouter(
	config *RouterConfig,
	seckillService *seckill.SeckillService,
	activityValidator *activity.ActivityValidator,
	metricsCollector *cache.MetricsCollector,
	log logger.Logger,
) *Router {
	if config == nil {
		config = DefaultRouterConfig()
	}

	// 设置Gin模式
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// 创建处理器
	// Create handlers
	seckillHandler := NewSeckillHandler(seckillService, activityValidator, metricsCollector, log)
	metricsAPI := cache.NewMetricsAPI(metricsCollector, log)

	router := &Router{
		engine:         engine,
		config:         config,
		logger:         log,
		seckillHandler: seckillHandler,
		metricsAPI:     metricsAPI,
	}

	// 设置中间件
	// Setup middlewares
	router.setupMiddlewares()

	// 设置路由
	// Setup routes
	router.setupRoutes()

	return router
}

// setupMiddlewares 设置中间件
// setupMiddlewares sets up middlewares
func (r *Router) setupMiddlewares() {
	// 恢复中间件（必须第一个）
	// Recovery middleware (must be first)
	r.engine.Use(RecoveryMiddleware(r.logger))

	// 请求ID中间件
	// Request ID middleware
	r.engine.Use(RequestIDMiddleware())

	// 日志中间件
	// Logger middleware
	r.engine.Use(LoggerMiddleware(r.logger))

	// CORS中间件
	// CORS middleware
	if r.config.EnableCORS {
		r.engine.Use(CORSMiddleware())
	}

	// 安全中间件
	// Security middleware
	r.engine.Use(SecurityMiddleware())

	// 指标中间件
	// Metrics middleware
	if r.config.EnableMetrics {
		r.engine.Use(MetricsMiddleware())
	}

	// 限流中间件
	// Rate limit middleware
	if r.config.EnableRateLimit {
		rateLimitConfig := &RateLimitConfig{
			RequestsPerSecond: 1000,
			BurstSize:         100,
			WindowSize:        time.Minute,
		}
		r.engine.Use(RateLimitMiddleware(rateLimitConfig))
	}
}

// setupRoutes 设置路由
// setupRoutes sets up routes
func (r *Router) setupRoutes() {
	// 健康检查
	// Health check
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ping", r.ping)

	// API版本组
	// API version group
	v1 := r.engine.Group("/api/v1")

	// 参数验证中间件
	// Parameter validation middleware
	v1.Use(ValidationMiddleware())

	// 公开路由（不需要认证）
	// Public routes (no authentication required)
	r.setupPublicRoutes(v1)

	// 需要认证的路由
	// Routes requiring authentication
	if r.config.EnableAuth {
		authenticated := v1.Group("")
		authenticated.Use(AuthMiddleware())
		r.setupAuthenticatedRoutes(authenticated)
	} else {
		// 如果不启用认证，直接使用v1组
		// If authentication is disabled, use v1 group directly
		r.setupAuthenticatedRoutes(v1)
	}

	// 管理路由
	// Admin routes
	admin := v1.Group("/admin")
	if r.config.EnableAuth {
		admin.Use(AuthMiddleware())
	}
	r.setupAdminRoutes(admin)
}

// setupPublicRoutes 设置公开路由
// setupPublicRoutes sets up public routes
func (r *Router) setupPublicRoutes(group *gin.RouterGroup) {
	// 库存查询（公开）
	// Stock query (public)
	group.GET("/seckill/stock/:activity_id", r.seckillHandler.GetStock)
	group.GET("/seckill/stocks", r.seckillHandler.BatchGetStocks)

	// 活动信息查询（公开）
	// Activity info query (public)
	group.GET("/seckill/activity/:activity_id/info", r.seckillHandler.GetActivityInfo)
	group.GET("/seckill/activity/:activity_id/stats", r.seckillHandler.GetActivityStats)
}

// setupAuthenticatedRoutes 设置需要认证的路由
// setupAuthenticatedRoutes sets up authenticated routes
func (r *Router) setupAuthenticatedRoutes(group *gin.RouterGroup) {
	// 注册秒杀路由
	// Register seckill routes
	r.seckillHandler.RegisterRoutes(group)
}

// setupAdminRoutes 设置管理路由
// setupAdminRoutes sets up admin routes
func (r *Router) setupAdminRoutes(group *gin.RouterGroup) {
	// 注册指标路由
	// Register metrics routes
	r.metricsAPI.RegisterRoutes(group)

	// 系统管理路由
	// System admin routes
	group.GET("/system/info", r.systemInfo)
	group.POST("/system/reload", r.reloadSystem)
}

// healthCheck 健康检查
// healthCheck health check
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "seckill-service",
	})
}

// ping Ping检查
// ping Ping check
func (r *Router) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "pong",
		"timestamp": time.Now(),
	})
}

// systemInfo 系统信息
// systemInfo system information
func (r *Router) systemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":    "seckill-service",
		"version":    "1.0.0",
		"build_time": "2024-01-01T00:00:00Z",
		"go_version": "go1.21",
		"config": gin.H{
			"enable_auth":       r.config.EnableAuth,
			"enable_rate_limit": r.config.EnableRateLimit,
			"enable_metrics":    r.config.EnableMetrics,
			"enable_cors":       r.config.EnableCORS,
		},
		"timestamp": time.Now(),
	})
}

// reloadSystem 重新加载系统
// reloadSystem reloads system
func (r *Router) reloadSystem(c *gin.Context) {
	r.logger.Info("System reload requested",
		logger.String("client_ip", c.ClientIP()),
		logger.String("request_id", c.GetString("request_id")))

	// 这里应该实现系统重新加载逻辑
	// Should implement system reload logic here

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "System reload completed",
		"timestamp": time.Now(),
	})
}

// GetEngine 获取Gin引擎
// GetEngine gets Gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Start 启动路由器
// Start starts router
func (r *Router) Start(addr string) error {
	r.logger.Info("Starting HTTP server",
		logger.String("address", addr))

	return r.engine.Run(addr)
}

// DefaultRouterConfig 默认路由配置
// DefaultRouterConfig default router configuration
func DefaultRouterConfig() *RouterConfig {
	return &RouterConfig{
		EnableAuth:      false, // 开发环境暂时关闭认证
		EnableRateLimit: true,
		EnableMetrics:   true,
		EnableCORS:      true,
	}
}

// SetupTestRouter 设置测试路由器
// SetupTestRouter sets up test router
func SetupTestRouter(
	seckillService *seckill.SeckillService,
	activityValidator *activity.ActivityValidator,
	metricsCollector *cache.MetricsCollector,
	log logger.Logger,
) *gin.Engine {
	config := &RouterConfig{
		EnableAuth:      false,
		EnableRateLimit: false,
		EnableMetrics:   false,
		EnableCORS:      true,
	}

	router := NewRouter(config, seckillService, activityValidator, metricsCollector, log)
	return router.GetEngine()
}
