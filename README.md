# Flash Sku - 高并发秒杀系统

## 项目概述
基于 Django + Gin 微服务架构的高性能秒杀电商平台，展示现代Web开发的核心技术栈和工程实践能力。

## 技术栈
- **前端**: Vue 3 + TypeScript + Pinia + Element Plus
- **业务后端**: Django + DRF + Celery + PostgreSQL
- **秒杀服务**: Go + Gin + Redis + RabbitMQ
- **API网关**: Nginx
- **容器化**: Docker + Docker Compose

## 项目结构
```
flash-sku/
├── docs/                    # 项目文档
│   ├── 系统设计文档.md       # 系统架构设计
│   ├── 开发路线图.md         # 开发计划
│   ├── 项目开发规范.md       # 开发规范
│   └── 项目结构说明.md       # 项目结构
├── backend/                 # Django 业务服务
├── seckill/                # Go 秒杀服务
├── frontend/               # Vue 前端应用
├── nginx/                  # Nginx 配置
├── scripts/                # 自动化脚本
├── docker-compose.yml      # 容器编排
├── Makefile               # 项目管理脚本
└── README.md              # 项目说明
```

## 快速开始

### 环境要求
- Docker & Docker Compose
- Python 3.11+
- Go 1.21+
- Node.js 18+

### 一键启动 (Docker)
```bash
# 克隆项目
git clone <repository-url>
cd flash-sku

# 方式1: 使用Makefile (推荐)
make setup    # 初始化项目
make up       # 启动所有服务

# 方式2: 使用Docker Compose
docker-compose up --build
```

### 本地开发环境
```bash
# 1. 设置环境变量
cp .env.example .env
vim .env  # 编辑配置

# 2. Django 后端开发
cd backend
python3 -m venv venv                    # 创建虚拟环境
source venv/bin/activate                # 激活虚拟环境 (Linux/Mac)
# 或 venv\Scripts\activate             # Windows
pip install -r requirements.txt        # 安装依赖
python manage.py migrate               # 数据库迁移
python manage.py runserver             # 启动开发服务器

# 3. Go 秒杀服务开发
cd ../seckill
go mod tidy                            # 安装依赖
go run cmd/server/main.go              # 启动服务

# 4. Vue 前端开发
cd ../frontend
npm install                            # 安装依赖
npm run dev                            # 启动开发服务器
```

### 访问地址
- 🌐 **前端应用**: http://localhost
- 🔧 **Django管理**: http://localhost/admin/
- 📊 **RabbitMQ管理**: http://localhost:15672/
- ⚡ **秒杀API**: http://localhost/seckill/
- 📡 **Django API**: http://localhost/api/

### 默认账户
- **RabbitMQ**: guest/guest
- **Django超级用户**: 需要手动创建

### 创建Django超级用户
```bash
docker-compose exec django python manage.py createsuperuser
```

## 开发指南

### Sprint 开发流程
项目采用敏捷开发模式，分为6个Sprint：
1. **Sprint 1**: 核心数据模型与管理后台
2. **Sprint 2**: 用户认证系统
3. **Sprint 3**: 高性能秒杀服务 (Go)
4. **Sprint 4**: 前端应用开发
5. **Sprint 5**: 系统整合与订单处理
6. **Sprint 6**: 监控、测试与部署

### 常用命令
```bash
# 项目管理
./scripts/start.sh        # 启动系统
./scripts/stop.sh         # 停止系统
./scripts/monitor.sh      # 系统监控

# 测试
./scripts/test_all.sh              # 运行所有测试
./scripts/test_e2e.sh              # 端到端测试
./scripts/test_performance.sh     # 性能测试
./scripts/test_data_consistency.sh # 数据一致性测试
./scripts/test_error_scenarios.sh # 错误场景测试

# Docker Compose
docker-compose up -d      # 启动所有服务
docker-compose down       # 停止所有服务
docker-compose logs -f    # 查看日志
```

## 🧪 测试

项目包含全面的测试覆盖：

- **单元测试**: 各服务的功能测试
- **集成测试**: 服务间通信测试
- **端到端测试**: 完整用户流程测试
- **性能测试**: 高并发压力测试
- **错误场景测试**: 系统容错能力测试

### 运行测试
```bash
# 运行所有测试
./scripts/test_all.sh

# 运行特定测试
./scripts/test_e2e.sh              # 端到端测试
./scripts/test_performance.sh     # 性能测试
./scripts/test_data_consistency.sh # 数据一致性测试
./scripts/test_error_scenarios.sh # 错误场景测试
```

## 📊 系统监控

### 启动监控
```bash
# 持续监控
./scripts/monitor.sh

# 一次性检查
./scripts/monitor.sh --once

# 生成监控报告
./scripts/monitor.sh --report
```

### 性能指标
- **秒杀接口响应时间**: < 100ms
- **并发支持**: 1000+ 用户
- **秒杀TPS**: 500+
- **系统可用性**: 99.9%

### 代码规范
请严格遵循 `docs/项目开发规范.md` 中的所有规则。

### 项目文档
- 📋 **系统设计**: [docs/02-系统设计/系统设计文档.md](docs/02-系统设计/系统设计文档.md)
- 📅 **开发计划**: [docs/01-项目规划/开发路线图.md](docs/01-项目规划/开发路线图.md)
- 📏 **开发规范**: [docs/03-开发规范/项目开发规范.md](docs/03-开发规范/项目开发规范.md)
- 📁 **项目结构**: [docs/03-开发规范/项目结构说明.md](docs/03-开发规范/项目结构说明.md)
- 📝 **项目总结**: [docs/项目完成总结.md](docs/项目完成总结.md)

## ✨ 核心功能

### 用户系统
- ✅ 用户注册/登录
- ✅ JWT认证
- ✅ 权限管理

### 商品管理
- ✅ 商品CRUD操作
- ✅ 商品分类管理
- ✅ 库存管理

### 秒杀系统
- ✅ 高性能秒杀接口
- ✅ 库存预扣和回滚
- ✅ 防重复提交
- ✅ 限流和熔断

### 订单系统
- ✅ 异步订单创建
- ✅ 订单状态管理
- ✅ 支付超时处理
- ✅ 库存回滚机制

### 前端界面
- ✅ 响应式设计
- ✅ 实时库存显示
- ✅ 秒杀按钮优化
- ✅ 用户体验增强

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 开发规范
- 遵循 [Conventional Commits](https://conventionalcommits.org/) 提交规范
- 代码需要通过所有测试
- 新功能需要添加相应的测试
- 更新相关文档

## 📄 许可证

本项目采用 MIT 许可证。

---

**Flash Sku** - 让秒杀更简单、更可靠！ 🎯
