package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// EnhancedLogger 增强的日志实现
// EnhancedLogger enhanced logger implementation
type EnhancedLogger struct {
	level  Level
	format string
	writer io.Writer
}

// LogEntry 日志条目结构
// LogEntry log entry structure
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Fields    map[string]any    `json:"fields,omitempty"`
	Service   string            `json:"service"`
}

// NewEnhancedLogger 创建增强的日志器
// NewEnhancedLogger creates enhanced logger
func NewEnhancedLogger(config *Config) Logger {
	var writer io.Writer

	// 根据输出配置选择写入器
	// Choose writer based on output configuration
	switch config.Output {
	case "stderr":
		writer = os.Stderr
	case "stdout", "":
		writer = os.Stdout
	default:
		// 文件输出 - 在实际项目中可以集成日志轮转库
		// File output - can integrate log rotation library in real project
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Failed to open log file %s: %v, falling back to stdout\n", config.Output, err)
			writer = os.Stdout
		} else {
			writer = file
		}
	}

	return &EnhancedLogger{
		level:  config.Level,
		format: config.Format,
		writer: writer,
	}
}

// Debug 调试日志
// Debug debug log
func (l *EnhancedLogger) Debug(msg string, fields ...Field) {
	if l.level <= DEBUG {
		l.log("DEBUG", msg, fields...)
	}
}

// Info 信息日志
// Info info log
func (l *EnhancedLogger) Info(msg string, fields ...Field) {
	if l.level <= INFO {
		l.log("INFO", msg, fields...)
	}
}

// Warn 警告日志
// Warn warning log
func (l *EnhancedLogger) Warn(msg string, fields ...Field) {
	if l.level <= WARN {
		l.log("WARN", msg, fields...)
	}
}

// Error 错误日志
// Error error log
func (l *EnhancedLogger) Error(msg string, fields ...Field) {
	if l.level <= ERROR {
		l.log("ERROR", msg, fields...)
	}
}

// Fatal 致命错误日志
// Fatal fatal log
func (l *EnhancedLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields...)
	os.Exit(1)
}

// log 内部日志方法
// log internal log method
func (l *EnhancedLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if l.format == "json" {
		l.logJSON(timestamp, level, msg, fields...)
	} else {
		l.logText(timestamp, level, msg, fields...)
	}
}

// logJSON JSON格式日志
// logJSON JSON format log
func (l *EnhancedLogger) logJSON(timestamp, level, msg string, fields ...Field) {
	entry := LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   msg,
		Service:   "seckill",
	}

	if len(fields) > 0 {
		entry.Fields = make(map[string]any)
		for _, field := range fields {
			entry.Fields[field.Key] = field.Value
		}
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// 如果JSON序列化失败，回退到文本格式
		// If JSON serialization fails, fallback to text format
		l.logText(timestamp, level, msg, fields...)
		return
	}

	fmt.Fprintln(l.writer, string(jsonData))
}

// logText 文本格式日志
// logText text format log
func (l *EnhancedLogger) logText(timestamp, level, msg string, fields ...Field) {
	logMsg := fmt.Sprintf("[%s] %s %s", timestamp, level, msg)

	// 添加字段
	// Add fields
	for _, field := range fields {
		logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
	}

	fmt.Fprintln(l.writer, logMsg)
}

// LevelFromString 从字符串解析日志级别
// LevelFromString parses log level from string
func LevelFromString(levelStr string) Level {
	switch levelStr {
	case "DEBUG", "debug":
		return DEBUG
	case "INFO", "info":
		return INFO
	case "WARN", "warn", "WARNING", "warning":
		return WARN
	case "ERROR", "error":
		return ERROR
	case "FATAL", "fatal":
		return FATAL
	default:
		return INFO
	}
}

// LevelToString 将日志级别转换为字符串
// LevelToString converts log level to string
func LevelToString(level Level) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "INFO"
	}
}
