# Go 秒杀服务 / Go Seckill Service

高性能秒杀服务，基于Go + Gin + Redis + RabbitMQ实现。
High-performance seckill service built with Go + Gin + Redis + RabbitMQ.

## 🚀 功能特性 / Features

- ⚡ **高并发处理**: 支持1000+ QPS的秒杀请求
- 🔒 **防超卖机制**: Redis Lua脚本原子操作 + 数据库约束双重保障
- 🛡️ **多级限流**: 全局/IP/用户三级限流防刷
- 📊 **性能监控**: 实时性能指标和健康检查
- 🔄 **异步处理**: RabbitMQ异步订单创建
- 🐳 **容器化**: Docker容器化部署

## 📁 项目结构 / Project Structure

```
seckill/
├── cmd/                    # 应用入口 / Application entry
│   └── server/            # 服务器启动 / Server startup
├── internal/              # 内部包 / Internal packages
│   ├── config/           # 配置管理 / Configuration
│   ├── handler/          # HTTP处理器 / HTTP handlers
│   ├── service/          # 业务逻辑层 / Business logic
│   ├── repository/       # 数据访问层 / Data access
│   ├── middleware/       # 中间件 / Middleware
│   └── model/            # 数据模型 / Data models
├── pkg/                   # 公共包 / Public packages
│   ├── redis/            # Redis客户端 / Redis client
│   ├── rabbitmq/         # RabbitMQ客户端 / RabbitMQ client
│   ├── logger/           # 日志组件 / Logger
│   └── utils/            # 工具函数 / Utilities
├── tests/                 # 测试文件 / Test files
├── Dockerfile            # Docker构建文件 / Docker build file
├── go.mod                # Go模块定义 / Go module definition
└── README.md             # 项目说明 / Project documentation
```

## 🔧 技术栈 / Tech Stack

- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **缓存**: Redis 7+
- **消息队列**: RabbitMQ 3+
- **数据库**: PostgreSQL 15+
- **容器**: Docker & Docker Compose

## 📡 API接口 / API Endpoints

### 秒杀相关 / Seckill APIs

```http
# 秒杀请求 / Seckill request
POST /seckill/{activity_id}

# 获取实时库存 / Get real-time stock
GET /seckill/stock/{activity_id}
```

### 管理接口 / Management APIs

```http
# 健康检查 / Health check
GET /seckill/health

# 性能指标 / Performance metrics
GET /seckill/metrics
```

## 🚀 快速开始 / Quick Start

### 环境要求 / Prerequisites

- Go 1.21+
- Redis 7+
- RabbitMQ 3+
- PostgreSQL 15+

### 本地开发 / Local Development

1. **克隆项目 / Clone repository**
```bash
git clone <repository-url>
cd seckill
```

2. **安装依赖 / Install dependencies**
```bash
go mod download
```

3. **设置环境变量 / Set environment variables**
```bash
export REDIS_HOST=localhost
export REDIS_PORT=6379
export RABBITMQ_URL=amqp://guest:guest@localhost:5672/
export DB_HOST=localhost
export DB_PORT=5432
```

4. **运行服务 / Run service**
```bash
go run cmd/server/main.go
```

### Docker部署 / Docker Deployment

```bash
# 构建镜像 / Build image
docker build -t seckill-service .

# 运行容器 / Run container
docker run -p 8080:8080 seckill-service
```

## 🧪 测试 / Testing

```bash
# 运行所有测试 / Run all tests
go test ./...

# 运行单元测试 / Run unit tests
go test ./tests/unit/...

# 运行性能测试 / Run benchmark tests
go test -bench=. ./tests/performance/...

# 生成覆盖率报告 / Generate coverage report
go test -cover ./...
```

## 📊 性能指标 / Performance Metrics

- **QPS**: 1000+ (目标)
- **延迟**: < 100ms (P99)
- **并发**: 支持高并发请求
- **可用性**: 99.9%+

## 🔧 配置说明 / Configuration

所有配置通过环境变量设置：
All configurations are set via environment variables:

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| SERVER_PORT | 8080 | 服务端口 |
| REDIS_HOST | localhost | Redis主机 |
| REDIS_PORT | 6379 | Redis端口 |
| RABBITMQ_URL | amqp://guest:guest@localhost:5672/ | RabbitMQ连接URL |
| SECKILL_GLOBAL_RATE_LIMIT | 1000 | 全局限流QPS |
| SECKILL_IP_RATE_LIMIT | 10 | IP限流QPS |
| SECKILL_USER_RATE_LIMIT | 1 | 用户限流QPS |

## 📝 开发状态 / Development Status

- [x] 项目结构创建 / Project structure created
- [ ] 依赖管理配置 / Dependency management
- [ ] 基础组件集成 / Basic components integration
- [ ] 缓存预热机制 / Cache warming mechanism
- [ ] 秒杀核心逻辑 / Seckill core logic
- [ ] 限流机制 / Rate limiting
- [ ] 异步消息处理 / Async message processing
- [ ] 性能测试 / Performance testing

## 🤝 贡献指南 / Contributing

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证 / License

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
