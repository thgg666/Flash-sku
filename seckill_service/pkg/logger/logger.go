package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Level 日志级别
// Level log level
type Level int

const (
	// DEBUG 调试级别
	DEBUG Level = iota
	// INFO 信息级别
	INFO
	// WARN 警告级别
	WARN
	// ERROR 错误级别
	ERROR
	// FATAL 致命错误级别
	FATAL
)

// Logger 日志接口
// Logger interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field 日志字段
// Field log field
type Field struct {
	Key   string
	Value any
}

// SimpleLogger 简单日志实现
// SimpleLogger simple logger implementation
type SimpleLogger struct {
	level  Level
	logger *log.Logger
}

// Config 日志配置
// Config logger configuration
type Config struct {
	Level      Level  `json:"level"`
	Format     string `json:"format"`      // "json" or "text"
	Output     string `json:"output"`      // "stdout", "stderr", or file path
	MaxSize    int    `json:"max_size"`    // 日志文件最大大小(MB)
	MaxBackups int    `json:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `json:"max_age"`     // 保留日志文件的最大天数
	Compress   bool   `json:"compress"`    // 是否压缩旧日志文件
}

// NewLogger 创建新的日志器
// NewLogger creates new logger
func NewLogger(config *Config) Logger {
	logger := log.New(os.Stdout, "", 0)

	return &SimpleLogger{
		level:  config.Level,
		logger: logger,
	}
}

// Debug 调试日志
// Debug debug log
func (l *SimpleLogger) Debug(msg string, fields ...Field) {
	if l.level <= DEBUG {
		l.log("DEBUG", msg, fields...)
	}
}

// Info 信息日志
// Info info log
func (l *SimpleLogger) Info(msg string, fields ...Field) {
	if l.level <= INFO {
		l.log("INFO", msg, fields...)
	}
}

// Warn 警告日志
// Warn warning log
func (l *SimpleLogger) Warn(msg string, fields ...Field) {
	if l.level <= WARN {
		l.log("WARN", msg, fields...)
	}
}

// Error 错误日志
// Error error log
func (l *SimpleLogger) Error(msg string, fields ...Field) {
	if l.level <= ERROR {
		l.log("ERROR", msg, fields...)
	}
}

// Fatal 致命错误日志
// Fatal fatal log
func (l *SimpleLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields...)
	os.Exit(1)
}

// log 内部日志方法
// log internal log method
func (l *SimpleLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level, msg)

	// 添加字段
	// Add fields
	for _, field := range fields {
		logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
	}

	l.logger.Println(logMsg)
}

// String 字符串字段
// String string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int 整数字段
// Int integer field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 64位整数字段
// Int64 64-bit integer field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 浮点数字段
// Float64 float field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool 布尔字段
// Bool boolean field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error 错误字段
// Error error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Duration 时间间隔字段
// Duration duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}
