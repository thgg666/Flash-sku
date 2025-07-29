package unit

import (
	"os"
	"testing"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestSimpleLogger(t *testing.T) {
	// 测试简单日志器
	// Test simple logger
	config := &logger.Config{
		Level:  logger.INFO,
		Format: "text",
		Output: "stdout",
	}

	log := logger.NewLogger(config)
	assert.NotNil(t, log)

	// 测试不同级别的日志
	// Test different log levels
	log.Debug("Debug message") // 不应该输出，因为级别是INFO
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")
}

func TestEnhancedLogger(t *testing.T) {
	// 测试增强日志器
	// Test enhanced logger
	config := &logger.Config{
		Level:  logger.DEBUG,
		Format: "json",
		Output: "stdout",
	}

	log := logger.NewEnhancedLogger(config)
	assert.NotNil(t, log)

	// 测试带字段的日志
	// Test logging with fields
	log.Info("Test message with fields",
		logger.String("key1", "value1"),
		logger.Int("key2", 42),
		logger.Bool("key3", true))
}

func TestLoggerLevels(t *testing.T) {
	// 测试日志级别转换
	// Test log level conversion
	assert.Equal(t, logger.DEBUG, logger.LevelFromString("debug"))
	assert.Equal(t, logger.INFO, logger.LevelFromString("info"))
	assert.Equal(t, logger.WARN, logger.LevelFromString("warn"))
	assert.Equal(t, logger.ERROR, logger.LevelFromString("error"))
	assert.Equal(t, logger.FATAL, logger.LevelFromString("fatal"))
	assert.Equal(t, logger.INFO, logger.LevelFromString("unknown"))

	assert.Equal(t, "DEBUG", logger.LevelToString(logger.DEBUG))
	assert.Equal(t, "INFO", logger.LevelToString(logger.INFO))
	assert.Equal(t, "WARN", logger.LevelToString(logger.WARN))
	assert.Equal(t, "ERROR", logger.LevelToString(logger.ERROR))
	assert.Equal(t, "FATAL", logger.LevelToString(logger.FATAL))
}

func TestLoggerWithFile(t *testing.T) {
	// 测试文件输出
	// Test file output
	tmpFile := "/tmp/test_seckill.log"
	defer os.Remove(tmpFile)

	config := &logger.Config{
		Level:  logger.INFO,
		Format: "text",
		Output: tmpFile,
	}

	log := logger.NewEnhancedLogger(config)
	log.Info("Test file logging")

	// 检查文件是否存在
	// Check if file exists
	_, err := os.Stat(tmpFile)
	assert.NoError(t, err)
}
