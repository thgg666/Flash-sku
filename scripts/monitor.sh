#!/bin/bash

# Flash Sku - 系统监控脚本
# 用于监控整个微服务系统的健康状态

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="Flash Sku"
MONITOR_INTERVAL=30
LOG_FILE="logs/monitor.log"

# 日志函数
log_info() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $1"
    echo -e "${GREEN}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

log_warn() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [WARN] $1"
    echo -e "${YELLOW}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

log_error() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1"
    echo -e "${RED}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

log_step() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [STEP] $1"
    echo -e "${BLUE}$msg${NC}"
    echo "$msg" >> "$LOG_FILE"
}

# 初始化监控
init_monitor() {
    mkdir -p logs
    log_info "开始监控 $PROJECT_NAME 系统..."
}

# 检查服务状态
check_service_status() {
    local service=$1
    local name=$2
    
    if docker-compose ps "$service" | grep -q "Up"; then
        if docker-compose ps "$service" | grep -q "healthy"; then
            log_info "$name 服务健康"
            return 0
        else
            log_warn "$name 服务运行但不健康"
            return 1
        fi
    else
        log_error "$name 服务未运行"
        return 2
    fi
}

# 检查服务响应
check_service_response() {
    local url=$1
    local name=$2
    local timeout=${3:-10}
    
    if curl -f -s --max-time "$timeout" "$url" > /dev/null 2>&1; then
        log_info "$name 响应正常"
        return 0
    else
        log_error "$name 响应异常"
        return 1
    fi
}

# 检查资源使用情况
check_resource_usage() {
    log_step "检查资源使用情况..."
    
    # 检查CPU和内存使用
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" | while read line; do
        if [[ "$line" == *"flashsku"* ]]; then
            echo "$line"
            
            # 提取内存使用百分比
            mem_percent=$(echo "$line" | awk '{print $4}' | sed 's/%//')
            if (( $(echo "$mem_percent > 80" | bc -l) )); then
                log_warn "容器内存使用率过高: $line"
            fi
            
            # 提取CPU使用百分比
            cpu_percent=$(echo "$line" | awk '{print $2}' | sed 's/%//')
            if (( $(echo "$cpu_percent > 80" | bc -l) )); then
                log_warn "容器CPU使用率过高: $line"
            fi
        fi
    done
}

# 检查磁盘空间
check_disk_space() {
    log_step "检查磁盘空间..."
    
    local usage=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$usage" -gt 80 ]; then
        log_warn "磁盘空间使用率过高: ${usage}%"
    else
        log_info "磁盘空间使用率: ${usage}%"
    fi
}

# 检查日志错误
check_logs_for_errors() {
    log_step "检查服务日志错误..."
    
    local services=("django" "gin" "nginx" "celery")
    
    for service in "${services[@]}"; do
        local error_count=$(docker-compose logs --tail=100 "$service" 2>/dev/null | grep -i "error\|exception\|failed" | wc -l)
        
        if [ "$error_count" -gt 0 ]; then
            log_warn "$service 服务发现 $error_count 个错误日志"
        else
            log_info "$service 服务日志正常"
        fi
    done
}

# 检查数据库连接
check_database_connection() {
    log_step "检查数据库连接..."
    
    if docker-compose exec -T postgres pg_isready -U flashsku_user -d flashsku_db > /dev/null 2>&1; then
        log_info "PostgreSQL数据库连接正常"
    else
        log_error "PostgreSQL数据库连接失败"
    fi
}

# 检查Redis连接
check_redis_connection() {
    log_step "检查Redis连接..."
    
    if docker-compose exec -T redis redis-cli ping | grep -q "PONG"; then
        log_info "Redis连接正常"
    else
        log_error "Redis连接失败"
    fi
}

# 检查RabbitMQ连接
check_rabbitmq_connection() {
    log_step "检查RabbitMQ连接..."
    
    if docker-compose exec -T rabbitmq rabbitmq-diagnostics ping > /dev/null 2>&1; then
        log_info "RabbitMQ连接正常"
    else
        log_error "RabbitMQ连接失败"
    fi
}

# 性能测试
performance_test() {
    log_step "执行性能测试..."
    
    # 测试API响应时间
    local api_response_time=$(curl -o /dev/null -s -w "%{time_total}" http://localhost/api/health/ 2>/dev/null || echo "999")
    
    if (( $(echo "$api_response_time < 1.0" | bc -l) )); then
        log_info "API响应时间正常: ${api_response_time}s"
    else
        log_warn "API响应时间过长: ${api_response_time}s"
    fi
    
    # 测试秒杀API响应时间
    local seckill_response_time=$(curl -o /dev/null -s -w "%{time_total}" http://localhost/seckill/health 2>/dev/null || echo "999")
    
    if (( $(echo "$seckill_response_time < 0.5" | bc -l) )); then
        log_info "秒杀API响应时间正常: ${seckill_response_time}s"
    else
        log_warn "秒杀API响应时间过长: ${seckill_response_time}s"
    fi
}

# 生成监控报告
generate_report() {
    local report_file="logs/monitor_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "========================================"
        echo "$PROJECT_NAME 系统监控报告"
        echo "生成时间: $(date)"
        echo "========================================"
        echo ""
        
        echo "=== 服务状态 ==="
        docker-compose ps
        echo ""
        
        echo "=== 资源使用 ==="
        docker stats --no-stream
        echo ""
        
        echo "=== 磁盘使用 ==="
        df -h
        echo ""
        
        echo "=== 网络连接 ==="
        netstat -tlnp | grep -E ":80|:8000|:8080|:3000|:5432|:6379|:5672"
        echo ""
        
        echo "=== 最近错误日志 ==="
        tail -n 50 "$LOG_FILE" | grep "ERROR"
        
    } > "$report_file"
    
    log_info "监控报告已生成: $report_file"
}

# 主监控循环
monitor_loop() {
    log_info "开始持续监控，间隔: ${MONITOR_INTERVAL}秒"
    
    while true; do
        echo ""
        log_step "执行监控检查..."
        
        # 检查所有服务
        check_service_status "nginx" "Nginx"
        check_service_status "django" "Django"
        check_service_status "gin" "Go Seckill"
        check_service_status "frontend" "Frontend"
        check_service_status "postgres" "PostgreSQL"
        check_service_status "redis" "Redis"
        check_service_status "rabbitmq" "RabbitMQ"
        check_service_status "celery" "Celery"
        
        # 检查服务响应
        check_service_response "http://localhost/health" "系统健康检查"
        check_service_response "http://localhost/api/health/" "Django API"
        check_service_response "http://localhost/seckill/health" "Go秒杀API"
        
        # 检查资源使用
        check_resource_usage
        check_disk_space
        
        # 检查数据库连接
        check_database_connection
        check_redis_connection
        check_rabbitmq_connection
        
        # 检查日志错误
        check_logs_for_errors
        
        # 性能测试
        performance_test
        
        log_info "监控检查完成，等待下次检查..."
        sleep "$MONITOR_INTERVAL"
    done
}

# 一次性检查
single_check() {
    log_step "执行一次性系统检查..."
    
    # 执行所有检查
    check_service_status "nginx" "Nginx"
    check_service_status "django" "Django"
    check_service_status "gin" "Go Seckill"
    check_service_status "frontend" "Frontend"
    check_service_status "postgres" "PostgreSQL"
    check_service_status "redis" "Redis"
    check_service_status "rabbitmq" "RabbitMQ"
    check_service_status "celery" "Celery"
    
    check_resource_usage
    check_disk_space
    check_database_connection
    check_redis_connection
    check_rabbitmq_connection
    performance_test
    
    generate_report
    
    log_info "系统检查完成"
}

# 主函数
main() {
    echo "========================================"
    echo "📊 $PROJECT_NAME 系统监控"
    echo "========================================"
    
    init_monitor
    
    case "$1" in
        --once|-o)
            single_check
            ;;
        --report|-r)
            generate_report
            ;;
        --interval|-i)
            MONITOR_INTERVAL="$2"
            monitor_loop
            ;;
        *)
            monitor_loop
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo "$PROJECT_NAME 系统监控脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -o, --once              执行一次性检查"
    echo "  -r, --report            生成监控报告"
    echo "  -i, --interval <秒>     设置监控间隔（默认30秒）"
    echo ""
    echo "示例:"
    echo "  $0                      # 持续监控"
    echo "  $0 --once              # 一次性检查"
    echo "  $0 --interval 60       # 每60秒监控一次"
    echo "  $0 --report            # 生成报告"
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
