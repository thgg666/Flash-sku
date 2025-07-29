package unit

import (
	"os"
	"testing"
	"time"

	"github.com/flashsku/seckill/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// 测试默认配置加载
	// Test default config loading
	cfg := config.LoadConfig()
	
	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, "6379", cfg.Redis.Port)
	assert.Equal(t, 1000, cfg.Seckill.GlobalRateLimit)
}

func TestLoadConfigWithEnv(t *testing.T) {
	// 设置环境变量
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("REDIS_HOST", "redis-server")
	os.Setenv("SECKILL_GLOBAL_RATE_LIMIT", "2000")
	
	defer func() {
		// 清理环境变量
		// Clean up environment variables
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("SECKILL_GLOBAL_RATE_LIMIT")
	}()
	
	cfg := config.LoadConfig()
	
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "redis-server", cfg.Redis.Host)
	assert.Equal(t, 2000, cfg.Seckill.GlobalRateLimit)
}

func TestConfigTimeouts(t *testing.T) {
	cfg := config.LoadConfig()
	
	// 测试默认超时设置
	// Test default timeout settings
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.IdleTimeout)
}
