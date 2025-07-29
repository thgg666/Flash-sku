package config

import (
	"os"
	"strconv"
	"time"
)

// Config 应用配置结构体
// Config application configuration structure
type Config struct {
	Server   ServerConfig   `json:"server"`
	Redis    RedisConfig    `json:"redis"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq"`
	Database DatabaseConfig `json:"database"`
	Seckill  SeckillConfig  `json:"seckill"`
}

// ServerConfig 服务器配置
// ServerConfig server configuration
type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// RedisConfig Redis配置
// RedisConfig Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Password     string `json:"password"`
	Database     int    `json:"database"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

// RabbitMQConfig RabbitMQ配置
// RabbitMQConfig RabbitMQ configuration
type RabbitMQConfig struct {
	URL          string `json:"url"`
	Exchange     string `json:"exchange"`
	Queue        string `json:"queue"`
	RoutingKey   string `json:"routing_key"`
	PrefetchSize int    `json:"prefetch_size"`
}

// DatabaseConfig 数据库配置
// DatabaseConfig database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

// SeckillConfig 秒杀业务配置
// SeckillConfig seckill business configuration
type SeckillConfig struct {
	GlobalRateLimit int `json:"global_rate_limit"` // 全局限流 QPS
	IPRateLimit     int `json:"ip_rate_limit"`     // IP限流 QPS
	UserRateLimit   int `json:"user_rate_limit"`   // 用户限流 QPS
	WorkerPoolSize  int `json:"worker_pool_size"`  // 工作协程池大小
}

// LoadConfig 加载配置
// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getIntEnv("REDIS_DATABASE", 0),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 5),
		},
		RabbitMQ: RabbitMQConfig{
			URL:          getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:     getEnv("RABBITMQ_EXCHANGE", "seckill"),
			Queue:        getEnv("RABBITMQ_QUEUE", "order.create"),
			RoutingKey:   getEnv("RABBITMQ_ROUTING_KEY", "order.create"),
			PrefetchSize: getIntEnv("RABBITMQ_PREFETCH_SIZE", 10),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "pass"),
			Database: getEnv("DB_NAME", "flashsku"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Seckill: SeckillConfig{
			GlobalRateLimit: getIntEnv("SECKILL_GLOBAL_RATE_LIMIT", 1000),
			IPRateLimit:     getIntEnv("SECKILL_IP_RATE_LIMIT", 10),
			UserRateLimit:   getIntEnv("SECKILL_USER_RATE_LIMIT", 1),
			WorkerPoolSize:  getIntEnv("SECKILL_WORKER_POOL_SIZE", 100),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv 获取整数类型的环境变量
// getIntEnv gets integer environment variable
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationEnv 获取时间间隔类型的环境变量
// getDurationEnv gets duration environment variable
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
