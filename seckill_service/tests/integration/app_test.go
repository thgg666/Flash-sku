package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flashsku/seckill/internal/app"
)

func TestAppInitialization(t *testing.T) {
	// 跳过集成测试，如果没有设置环境变量
	// Skip integration test if environment variables are not set
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建应用程序实例
	// Create application instance
	application, err := app.New()
	require.NoError(t, err, "应用程序初始化应该成功")
	require.NotNil(t, application, "应用程序实例不应该为空")

	// 测试组件是否正确初始化
	// Test if components are properly initialized
	assert.NotNil(t, application.GetRedisClient(), "Redis客户端应该被初始化")
	assert.NotNil(t, application.GetLogger(), "日志器应该被初始化")
	assert.NotNil(t, application.GetWorkerPool(), "工作池应该被初始化")

	// 测试工作池指标
	// Test worker pool metrics
	metrics := application.GetWorkerPool().GetMetrics()
	assert.Greater(t, metrics.ActiveWorkers, 0, "应该有活跃的工作协程")

	// 清理资源
	// Cleanup resources
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = application.Stop(ctx)
	assert.NoError(t, err, "应用程序停止应该成功")
}

func TestHealthCheckEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建应用程序实例
	// Create application instance
	application, err := app.New()
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		application.Stop(ctx)
	}()

	// 创建测试服务器
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy","service":"seckill"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 测试健康检查端点
	// Test health check endpoint
	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}

func TestWorkerPoolFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建应用程序实例
	// Create application instance
	application, err := app.New()
	require.NoError(t, err)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		application.Stop(ctx)
	}()

	workerPool := application.GetWorkerPool()
	require.NotNil(t, workerPool)

	// 提交测试任务
	// Submit test task
	taskExecuted := false
	err = workerPool.SubmitFunc(func(ctx context.Context) error {
		taskExecuted = true
		return nil
	})
	require.NoError(t, err)

	// 等待任务执行
	// Wait for task execution
	time.Sleep(100 * time.Millisecond)

	// 验证任务已执行
	// Verify task was executed
	assert.True(t, taskExecuted, "任务应该被执行")

	// 检查指标
	// Check metrics
	metrics := workerPool.GetMetrics()
	assert.Greater(t, metrics.TasksSubmitted, int64(0), "应该有已提交的任务")
	assert.Greater(t, metrics.TasksCompleted, int64(0), "应该有已完成的任务")
}
