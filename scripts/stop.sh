#!/bin/bash

# Flash Sku - 系统停止脚本
# 用于安全停止整个微服务系统

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

# 优雅停止服务
graceful_stop() {
    log_step "优雅停止系统服务..."
    
    # 首先停止网关，阻止新请求
    log_info "停止API网关..."
    docker-compose stop nginx || true
    
    # 停止前端服务
    log_info "停止前端服务..."
    docker-compose stop frontend || true
    
    # 停止应用服务
    log_info "停止应用服务..."
    docker-compose stop gin django || true
    
    # 停止后台任务服务
    log_info "停止后台任务服务..."
    docker-compose stop celery celery-beat order-consumer || true
    
    # 最后停止基础设施服务
    log_info "停止基础设施服务..."
    docker-compose stop postgres redis rabbitmq || true
    
    log_info "所有服务已停止"
}

# 强制停止服务
force_stop() {
    log_step "强制停止所有服务..."
    
    docker-compose down --timeout 30
    
    log_info "所有服务已强制停止"
}

# 清理资源
cleanup_resources() {
    log_step "清理系统资源..."
    
    # 清理停止的容器
    log_info "清理停止的容器..."
    docker container prune -f || true
    
    # 清理未使用的网络
    log_info "清理未使用的网络..."
    docker network prune -f || true
    
    # 清理未使用的镜像（可选）
    if [ "$1" = "--clean-images" ]; then
        log_info "清理未使用的镜像..."
        docker image prune -f || true
    fi
    
    # 清理未使用的数据卷（可选）
    if [ "$1" = "--clean-volumes" ]; then
        log_warn "清理数据卷（这将删除所有数据）..."
        read -p "确定要删除所有数据卷吗？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            docker-compose down -v
            docker volume prune -f || true
            log_warn "数据卷已清理"
        else
            log_info "跳过数据卷清理"
        fi
    fi
    
    log_info "资源清理完成"
}

# 显示系统状态
show_status() {
    log_step "检查系统状态..."
    
    echo ""
    echo "=== 容器状态 ==="
    docker-compose ps || echo "没有运行的容器"
    
    echo ""
    echo "=== 系统资源 ==="
    echo "Docker镜像:"
    docker images | grep flashsku || echo "没有Flash Sku镜像"
    
    echo ""
    echo "Docker数据卷:"
    docker volume ls | grep flashsku || echo "没有Flash Sku数据卷"
    
    echo ""
    echo "Docker网络:"
    docker network ls | grep flashsku || echo "没有Flash Sku网络"
}

# 备份数据
backup_data() {
    log_step "备份系统数据..."
    
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份数据库
    if docker-compose ps postgres | grep -q "Up"; then
        log_info "备份PostgreSQL数据库..."
        docker-compose exec -T postgres pg_dump -U flashsku_user flashsku_db > "$backup_dir/database.sql"
        log_info "数据库备份完成: $backup_dir/database.sql"
    else
        log_warn "PostgreSQL服务未运行，跳过数据库备份"
    fi
    
    # 备份媒体文件
    if [ -d "backend/media" ]; then
        log_info "备份媒体文件..."
        cp -r backend/media "$backup_dir/"
        log_info "媒体文件备份完成: $backup_dir/media"
    fi
    
    # 备份配置文件
    log_info "备份配置文件..."
    cp docker-compose.yml "$backup_dir/"
    cp .env "$backup_dir/" 2>/dev/null || true
    
    log_info "数据备份完成: $backup_dir"
}

# 主函数
main() {
    echo "========================================"
    echo "🛑 $PROJECT_NAME 系统停止脚本"
    echo "========================================"
    
    local action="$1"
    
    case "$action" in
        --force|-f)
            force_stop
            ;;
        --backup|-b)
            backup_data
            graceful_stop
            ;;
        --clean)
            graceful_stop
            cleanup_resources
            ;;
        --clean-all)
            graceful_stop
            cleanup_resources "--clean-images"
            ;;
        --clean-volumes)
            graceful_stop
            cleanup_resources "--clean-volumes"
            ;;
        --status|-s)
            show_status
            exit 0
            ;;
        *)
            graceful_stop
            ;;
    esac
    
    # 显示最终状态
    show_status
    
    echo ""
    log_info "系统停止完成！"
}

# 显示帮助信息
show_help() {
    echo "$PROJECT_NAME 系统停止脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示帮助信息"
    echo "  -f, --force         强制停止所有服务"
    echo "  -b, --backup        备份数据后停止"
    echo "  -s, --status        显示系统状态"
    echo "      --clean         停止并清理资源"
    echo "      --clean-all     停止并清理所有资源（包括镜像）"
    echo "      --clean-volumes 停止并清理数据卷（危险操作）"
    echo ""
    echo "示例:"
    echo "  $0                  # 优雅停止系统"
    echo "  $0 --force          # 强制停止"
    echo "  $0 --backup         # 备份后停止"
    echo "  $0 --clean          # 停止并清理"
    echo "  $0 --status         # 查看状态"
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
