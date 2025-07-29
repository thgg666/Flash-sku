#!/bin/bash

# Flash Sku - 系统启动脚本
# 用于启动整个微服务系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="Flash Sku"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_step "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查Docker服务状态
    if ! docker info &> /dev/null; then
        log_error "Docker服务未运行，请启动Docker服务"
        exit 1
    fi
    
    log_info "系统依赖检查通过"
}

# 检查环境文件
check_env_file() {
    log_step "检查环境配置..."
    
    if [ ! -f "$ENV_FILE" ]; then
        log_warn "环境文件 $ENV_FILE 不存在，创建默认配置..."
        create_default_env
    else
        log_info "环境文件已存在"
    fi
}

# 创建默认环境文件
create_default_env() {
    cat > "$ENV_FILE" << EOF
# Flash Sku 环境配置

# 数据库配置
POSTGRES_DB=flashsku_db
POSTGRES_USER=flashsku_user
POSTGRES_PASSWORD=flashsku_pass

# Redis配置
REDIS_PASSWORD=

# RabbitMQ配置
RABBITMQ_DEFAULT_USER=guest
RABBITMQ_DEFAULT_PASS=guest

# Django配置
SECRET_KEY=your-secret-key-here-$(openssl rand -hex 32)
DEBUG=True
ALLOWED_HOSTS=localhost,127.0.0.1,nginx

# Go服务配置
JWT_SECRET=your-jwt-secret-here-$(openssl rand -hex 32)
GIN_MODE=debug

# 前端配置
NODE_ENV=development
VITE_API_BASE_URL=http://localhost/api
VITE_SECKILL_BASE_URL=http://localhost/seckill
EOF
    
    log_info "默认环境文件已创建: $ENV_FILE"
    log_warn "请根据需要修改环境配置"
}

# 构建镜像
build_images() {
    log_step "构建Docker镜像..."
    
    # 构建所有服务镜像
    docker-compose build --no-cache
    
    log_info "Docker镜像构建完成"
}

# 启动基础服务
start_infrastructure() {
    log_step "启动基础设施服务..."
    
    # 启动数据库、缓存、消息队列
    docker-compose up -d postgres redis rabbitmq
    
    # 等待服务就绪
    log_info "等待基础服务启动..."
    sleep 30
    
    # 检查服务状态
    check_service_health "postgres" "PostgreSQL"
    check_service_health "redis" "Redis"
    check_service_health "rabbitmq" "RabbitMQ"
    
    log_info "基础设施服务启动完成"
}

# 启动应用服务
start_applications() {
    log_step "启动应用服务..."
    
    # 启动Django服务
    log_info "启动Django服务..."
    docker-compose up -d django
    sleep 20
    check_service_health "django" "Django"
    
    # 启动Celery服务
    log_info "启动Celery服务..."
    docker-compose up -d celery celery-beat order-consumer
    sleep 10
    
    # 启动Go服务
    log_info "启动Go秒杀服务..."
    docker-compose up -d gin
    sleep 15
    check_service_health "gin" "Go Seckill"
    
    # 启动前端服务
    log_info "启动前端服务..."
    docker-compose up -d frontend
    sleep 20
    check_service_health "frontend" "Frontend"
    
    log_info "应用服务启动完成"
}

# 启动网关
start_gateway() {
    log_step "启动API网关..."
    
    docker-compose up -d nginx
    sleep 10
    check_service_health "nginx" "Nginx Gateway"
    
    log_info "API网关启动完成"
}

# 检查服务健康状态
check_service_health() {
    local service=$1
    local name=$2
    local max_attempts=30
    local attempt=1
    
    log_info "检查 $name 服务健康状态..."
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose ps $service | grep -q "healthy\|Up"; then
            log_info "$name 服务健康"
            return 0
        fi
        
        log_warn "$name 服务未就绪，等待中... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "$name 服务启动失败或超时"
    return 1
}

# 显示系统状态
show_status() {
    log_step "系统状态概览..."
    
    echo ""
    echo "=== 服务状态 ==="
    docker-compose ps
    
    echo ""
    echo "=== 访问地址 ==="
    echo "🌐 前端应用: http://localhost"
    echo "🔧 Django管理: http://localhost/admin/"
    echo "📊 RabbitMQ管理: http://localhost:15672/"
    echo "⚡ Go秒杀API: http://localhost/seckill/"
    echo "📡 Django API: http://localhost/api/"
    
    echo ""
    echo "=== 默认账户 ==="
    echo "RabbitMQ: guest/guest"
    echo "Django超级用户: 需要手动创建"
    
    echo ""
    log_info "系统启动完成！"
}

# 清理函数
cleanup() {
    log_step "清理资源..."
    docker-compose down
}

# 主函数
main() {
    echo "========================================"
    echo "🚀 $PROJECT_NAME 系统启动脚本"
    echo "========================================"
    
    # 检查依赖
    check_dependencies
    
    # 检查环境配置
    check_env_file
    
    # 构建镜像
    if [ "$1" = "--build" ] || [ "$1" = "-b" ]; then
        build_images
    fi
    
    # 启动服务
    start_infrastructure
    start_applications
    start_gateway
    
    # 显示状态
    show_status
    
    echo ""
    log_info "启动完成！按 Ctrl+C 停止系统"
    
    # 设置信号处理
    trap cleanup EXIT
    
    # 保持脚本运行
    if [ "$1" = "--detach" ] || [ "$1" = "-d" ]; then
        log_info "后台运行模式"
        exit 0
    else
        # 跟踪日志
        log_info "跟踪系统日志..."
        docker-compose logs -f
    fi
}

# 显示帮助信息
show_help() {
    echo "$PROJECT_NAME 系统启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示帮助信息"
    echo "  -b, --build    重新构建镜像"
    echo "  -d, --detach   后台运行模式"
    echo ""
    echo "示例:"
    echo "  $0             # 启动系统并跟踪日志"
    echo "  $0 --build     # 重新构建并启动"
    echo "  $0 --detach    # 后台启动"
}

# 解析命令行参数
case "$1" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
