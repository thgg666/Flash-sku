package cache

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/gin-gonic/gin"
)

// MetricsAPI 指标API
// MetricsAPI metrics API
type MetricsAPI struct {
	collector *MetricsCollector
	logger    logger.Logger
}

// NewMetricsAPI 创建指标API
// NewMetricsAPI creates metrics API
func NewMetricsAPI(collector *MetricsCollector, log logger.Logger) *MetricsAPI {
	return &MetricsAPI{
		collector: collector,
		logger:    log,
	}
}

// RegisterRoutes 注册路由
// RegisterRoutes registers routes
func (api *MetricsAPI) RegisterRoutes(router *gin.RouterGroup) {
	metrics := router.Group("/metrics")
	{
		metrics.GET("/snapshot", api.GetSnapshot)
		metrics.GET("/current", api.GetCurrentMetrics)
		metrics.GET("/export", api.ExportMetrics)
		metrics.GET("/health", api.GetHealth)
		metrics.GET("/stock", api.GetStockMetrics)
		metrics.GET("/activity", api.GetActivityMetrics)
		metrics.POST("/reset", api.ResetMetrics)
	}
}

// GetSnapshot 获取指标快照
// GetSnapshot gets metrics snapshot
func (api *MetricsAPI) GetSnapshot(c *gin.Context) {
	snapshot := api.collector.GenerateSnapshot()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    snapshot,
	})
}

// GetCurrentMetrics 获取当前指标
// GetCurrentMetrics gets current metrics
func (api *MetricsAPI) GetCurrentMetrics(c *gin.Context) {
	metrics := api.collector.GetCurrentMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// ExportMetrics 导出指标
// ExportMetrics exports metrics
func (api *MetricsAPI) ExportMetrics(c *gin.Context) {
	format := c.DefaultQuery("format", "json")

	switch format {
	case "json":
		data, err := api.collector.ExportMetrics()
		if err != nil {
			api.logger.Error("Failed to export metrics", logger.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to export metrics",
			})
			return
		}

		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=cache_metrics.json")
		c.Data(http.StatusOK, "application/json", data)

	case "prometheus":
		// 生成Prometheus格式的指标
		// Generate Prometheus format metrics
		prometheusData := api.generatePrometheusMetrics()
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusOK, prometheusData)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Unsupported format. Use 'json' or 'prometheus'",
		})
	}
}

// GetHealth 获取健康状态
// GetHealth gets health status
func (api *MetricsAPI) GetHealth(c *gin.Context) {
	snapshot := api.collector.GenerateSnapshot()

	// 计算健康分数
	// Calculate health score
	healthScore := api.calculateHealthScore(snapshot)

	status := "healthy"
	if healthScore < 0.7 {
		status = "unhealthy"
	} else if healthScore < 0.9 {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":       status,
			"health_score": healthScore,
			"hit_rate":     snapshot.HitRate,
			"error_rate":   snapshot.ErrorRate,
			"alerts_count": len(snapshot.Alerts),
			"timestamp":    time.Now(),
		},
	})
}

// GetStockMetrics 获取库存指标
// GetStockMetrics gets stock metrics
func (api *MetricsAPI) GetStockMetrics(c *gin.Context) {
	activityID := c.Query("activity_id")

	metrics := api.collector.GetCurrentMetrics()

	if activityID != "" {
		// 返回特定活动的库存指标
		// Return specific activity stock metrics
		if stockMetric, exists := metrics.StockMetrics[activityID]; exists {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    stockMetric,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Activity not found",
			})
		}
	} else {
		// 返回所有库存指标
		// Return all stock metrics
		snapshot := api.collector.GenerateSnapshot()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"summary":       snapshot.StockSummary,
				"stock_metrics": metrics.StockMetrics,
			},
		})
	}
}

// GetActivityMetrics 获取活动指标
// GetActivityMetrics gets activity metrics
func (api *MetricsAPI) GetActivityMetrics(c *gin.Context) {
	activityID := c.Query("activity_id")

	metrics := api.collector.GetCurrentMetrics()

	if activityID != "" {
		// 返回特定活动的指标
		// Return specific activity metrics
		if activityMetric, exists := metrics.ActivityMetrics[activityID]; exists {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    activityMetric,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Activity not found",
			})
		}
	} else {
		// 返回所有活动指标
		// Return all activity metrics
		snapshot := api.collector.GenerateSnapshot()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"summary":          snapshot.ActivitySummary,
				"activity_metrics": metrics.ActivityMetrics,
			},
		})
	}
}

// ResetMetrics 重置指标
// ResetMetrics resets metrics
func (api *MetricsAPI) ResetMetrics(c *gin.Context) {
	// 检查权限（这里简化处理）
	// Check permissions (simplified here)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authorization required",
		})
		return
	}

	// 重置指标
	// Reset metrics
	api.collector.metrics.mu.Lock()
	api.collector.metrics.HitCount = 0
	api.collector.metrics.MissCount = 0
	api.collector.metrics.SetCount = 0
	api.collector.metrics.DeleteCount = 0
	api.collector.metrics.ErrorCount = 0
	api.collector.metrics.OperationCount = 0
	api.collector.metrics.TotalLatency = 0
	api.collector.metrics.AvgLatency = 0
	api.collector.metrics.MaxLatency = 0
	api.collector.metrics.MinLatency = 0
	api.collector.metrics.StartTime = time.Now()
	api.collector.metrics.LastUpdated = time.Now()
	api.collector.metrics.mu.Unlock()

	api.logger.Info("Cache metrics reset via API")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Metrics reset successfully",
	})
}

// calculateHealthScore 计算健康分数
// calculateHealthScore calculates health score
func (api *MetricsAPI) calculateHealthScore(snapshot *MetricsSnapshot) float64 {
	score := 1.0

	// 命中率影响 (40%)
	// Hit rate impact (40%)
	hitRateScore := snapshot.HitRate * 0.4

	// 错误率影响 (30%)
	// Error rate impact (30%)
	errorRateScore := (1.0 - snapshot.ErrorRate) * 0.3

	// 告警影响 (20%)
	// Alert impact (20%)
	alertScore := 0.2
	if len(snapshot.Alerts) > 0 {
		criticalCount := 0
		errorCount := 0
		warningCount := 0

		for _, alert := range snapshot.Alerts {
			switch alert.Level {
			case "critical":
				criticalCount++
			case "error":
				errorCount++
			case "warning":
				warningCount++
			}
		}

		// 严重告警大幅降低分数
		// Critical alerts significantly reduce score
		alertScore -= float64(criticalCount) * 0.1
		alertScore -= float64(errorCount) * 0.05
		alertScore -= float64(warningCount) * 0.02

		if alertScore < 0 {
			alertScore = 0
		}
	}

	// 延迟影响 (10%)
	// Latency impact (10%)
	latencyScore := 0.1
	if snapshot.AvgLatency > 100*time.Millisecond {
		latencyScore *= 0.5
	} else if snapshot.AvgLatency > 50*time.Millisecond {
		latencyScore *= 0.8
	}

	score = hitRateScore + errorRateScore + alertScore + latencyScore

	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	return score
}

// generatePrometheusMetrics 生成Prometheus格式指标
// generatePrometheusMetrics generates Prometheus format metrics
func (api *MetricsAPI) generatePrometheusMetrics() string {
	snapshot := api.collector.GenerateSnapshot()
	metrics := api.collector.GetCurrentMetrics()

	var result string

	// 基础指标
	// Basic metrics
	result += "# HELP cache_hit_total Total number of cache hits\n"
	result += "# TYPE cache_hit_total counter\n"
	result += "cache_hit_total " + strconv.FormatInt(metrics.HitCount, 10) + "\n\n"

	result += "# HELP cache_miss_total Total number of cache misses\n"
	result += "# TYPE cache_miss_total counter\n"
	result += "cache_miss_total " + strconv.FormatInt(metrics.MissCount, 10) + "\n\n"

	result += "# HELP cache_hit_rate Cache hit rate\n"
	result += "# TYPE cache_hit_rate gauge\n"
	result += "cache_hit_rate " + strconv.FormatFloat(snapshot.HitRate, 'f', 4, 64) + "\n\n"

	result += "# HELP cache_error_rate Cache error rate\n"
	result += "# TYPE cache_error_rate gauge\n"
	result += "cache_error_rate " + strconv.FormatFloat(snapshot.ErrorRate, 'f', 4, 64) + "\n\n"

	result += "# HELP cache_avg_latency_ms Average cache operation latency in milliseconds\n"
	result += "# TYPE cache_avg_latency_ms gauge\n"
	result += "cache_avg_latency_ms " + strconv.FormatFloat(float64(snapshot.AvgLatency.Milliseconds()), 'f', 2, 64) + "\n\n"

	// 库存指标
	// Stock metrics
	result += "# HELP stock_current_level Current stock level by activity\n"
	result += "# TYPE stock_current_level gauge\n"
	for activityID, stockMetric := range metrics.StockMetrics {
		result += "stock_current_level{activity_id=\"" + activityID + "\"} " + strconv.Itoa(stockMetric.CurrentStock) + "\n"
	}
	result += "\n"

	return result
}
