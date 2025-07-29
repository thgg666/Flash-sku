#!/bin/bash

# Flash Sku - 错误场景测试脚本
# 测试系统在各种异常情况下的容错能力

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost"
API_URL="$BASE_URL/api"
SECKILL_URL="$BASE_URL/seckill"
TIMEOUT=30

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

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 记录测试结果
record_test() {
    local test_name="$1"
    local result="$2"
    local details="$3"
    
    ((TOTAL_TESTS++))
    
    if [ "$result" = "PASS" ]; then
        ((PASSED_TESTS++))
        log_info "✓ $test_name - PASS"
    else
        ((FAILED_TESTS++))
        log_error "✗ $test_name - FAIL"
        if [ -n "$details" ]; then
            log_error "  详情: $details"
        fi
    fi
}

# 等待服务恢复
wait_for_service_recovery() {
    local service_name="$1"
    local health_url="$2"
    local max_wait=60
    local wait_time=0
    
    log_info "等待 $service_name 服务恢复..."
    
    while [ $wait_time -lt $max_wait ]; do
        if curl -f -s --max-time 5 "$health_url" > /dev/null 2>&1; then
            log_info "$service_name 服务已恢复"
            return 0
        fi
        
        sleep 5
        wait_time=$((wait_time + 5))
    done
    
    log_error "$service_name 服务恢复超时"
    return 1
}

# 测试数据库连接失败场景
test_database_failure() {
    log_test "测试数据库连接失败场景..."
    
    # 停止数据库服务
    log_info "停止PostgreSQL服务..."
    docker-compose stop postgres
    
    sleep 10
    
    # 测试API响应
    local response=$(curl -s -w '\n%{http_code}' --max-time 10 "$API_URL/health/" 2>/dev/null || echo -e "\n500")
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "500" ] || [ "$status_code" = "503" ]; then
        record_test "数据库失败时API响应" "PASS"
    else
        record_test "数据库失败时API响应" "FAIL" "期望5xx状态码，实际: $status_code"
    fi
    
    # 重启数据库服务
    log_info "重启PostgreSQL服务..."
    docker-compose start postgres
    
    # 等待服务恢复
    if wait_for_service_recovery "PostgreSQL" "$API_URL/health/"; then
        record_test "数据库服务恢复" "PASS"
    else
        record_test "数据库服务恢复" "FAIL"
    fi
}

# 测试Redis连接失败场景
test_redis_failure() {
    log_test "测试Redis连接失败场景..."
    
    # 停止Redis服务
    log_info "停止Redis服务..."
    docker-compose stop redis
    
    sleep 10
    
    # 测试秒杀API响应（应该降级处理）
    local response=$(curl -s -w '\n%{http_code}' --max-time 10 "$SECKILL_URL/health" 2>/dev/null || echo -e "\n500")
    local status_code=$(echo "$response" | tail -n1)
    
    # Redis失败时，秒杀服务可能返回503或者降级到数据库
    if [ "$status_code" = "200" ] || [ "$status_code" = "503" ]; then
        record_test "Redis失败时秒杀服务响应" "PASS"
    else
        record_test "Redis失败时秒杀服务响应" "FAIL" "状态码: $status_code"
    fi
    
    # 重启Redis服务
    log_info "重启Redis服务..."
    docker-compose start redis
    
    # 等待服务恢复
    if wait_for_service_recovery "Redis" "$SECKILL_URL/health"; then
        record_test "Redis服务恢复" "PASS"
    else
        record_test "Redis服务恢复" "FAIL"
    fi
}

# 测试RabbitMQ连接失败场景
test_rabbitmq_failure() {
    log_test "测试RabbitMQ连接失败场景..."
    
    # 停止RabbitMQ服务
    log_info "停止RabbitMQ服务..."
    docker-compose stop rabbitmq
    
    sleep 10
    
    # 测试秒杀请求（消息队列失败时的处理）
    local token=$(create_test_user "rabbitmq_test")
    if [ -n "$token" ]; then
        local data='{"activity_id": 1, "quantity": 1}'
        local response=$(curl -s -w '\n%{http_code}' --max-time 10 \
            -X POST \
            -H 'Content-Type: application/json' \
            -H "Authorization: Bearer $token" \
            -d "$data" \
            "$SECKILL_URL/1" 2>/dev/null || echo -e "\n500")
        
        local status_code=$(echo "$response" | tail -n1)
        
        # RabbitMQ失败时，应该有适当的错误处理
        if [ "$status_code" = "500" ] || [ "$status_code" = "503" ]; then
            record_test "RabbitMQ失败时秒杀请求处理" "PASS"
        else
            record_test "RabbitMQ失败时秒杀请求处理" "FAIL" "状态码: $status_code"
        fi
    else
        record_test "RabbitMQ失败时秒杀请求处理" "FAIL" "无法创建测试用户"
    fi
    
    # 重启RabbitMQ服务
    log_info "重启RabbitMQ服务..."
    docker-compose start rabbitmq
    
    # 等待服务恢复
    sleep 30  # RabbitMQ需要更长的启动时间
    if wait_for_service_recovery "RabbitMQ" "$API_URL/health/"; then
        record_test "RabbitMQ服务恢复" "PASS"
    else
        record_test "RabbitMQ服务恢复" "FAIL"
    fi
}

# 测试网络超时场景
test_network_timeout() {
    log_test "测试网络超时场景..."
    
    # 使用很短的超时时间测试
    local response=$(curl -s -w '\n%{http_code}' --max-time 1 "$API_URL/health/" 2>/dev/null || echo -e "\ntimeout")
    local result=$(echo "$response" | tail -n1)
    
    if [ "$result" = "timeout" ] || [ "$result" = "200" ]; then
        record_test "网络超时处理" "PASS"
    else
        record_test "网络超时处理" "FAIL" "结果: $result"
    fi
}

# 测试无效请求场景
test_invalid_requests() {
    log_test "测试无效请求场景..."
    
    # 测试无效JSON
    local response=$(curl -s -w '\n%{http_code}' --max-time 10 \
        -X POST \
        -H 'Content-Type: application/json' \
        -d 'invalid json' \
        "$API_URL/auth/login/" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "400" ]; then
        record_test "无效JSON处理" "PASS"
    else
        record_test "无效JSON处理" "FAIL" "状态码: $status_code"
    fi
    
    # 测试缺少必需字段
    local response=$(curl -s -w '\n%{http_code}' --max-time 10 \
        -X POST \
        -H 'Content-Type: application/json' \
        -d '{}' \
        "$API_URL/auth/login/" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "400" ]; then
        record_test "缺少必需字段处理" "PASS"
    else
        record_test "缺少必需字段处理" "FAIL" "状态码: $status_code"
    fi
    
    # 测试无效认证
    local response=$(curl -s -w '\n%{http_code}' --max-time 10 \
        -X POST \
        -H 'Content-Type: application/json' \
        -H 'Authorization: Bearer invalid_token' \
        -d '{"activity_id": 1, "quantity": 1}' \
        "$SECKILL_URL/1" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "401" ] || [ "$status_code" = "403" ]; then
        record_test "无效认证处理" "PASS"
    else
        record_test "无效认证处理" "FAIL" "状态码: $status_code"
    fi
}

# 测试资源耗尽场景
test_resource_exhaustion() {
    log_test "测试资源耗尽场景..."
    
    # 测试大量并发请求（模拟资源耗尽）
    local pids=()
    local results_dir="tmp/resource_test"
    mkdir -p "$results_dir"
    
    log_info "发起大量并发请求..."
    
    for i in {1..50}; do
        (
            local response=$(curl -s -w '\n%{http_code}' --max-time 5 "$API_URL/health/" 2>/dev/null || echo -e "\nerror")
            echo "$response" > "$results_dir/result_$i.txt"
        ) &
        pids+=($!)
    done
    
    # 等待所有请求完成
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    # 分析结果
    local success_count=0
    local error_count=0
    
    for result_file in "$results_dir"/result_*.txt; do
        if [ -f "$result_file" ]; then
            local status=$(tail -n1 "$result_file")
            if [ "$status" = "200" ]; then
                ((success_count++))
            else
                ((error_count++))
            fi
        fi
    done
    
    # 系统应该能处理大部分请求
    if [ $success_count -gt $((error_count * 2)) ]; then
        record_test "资源耗尽时系统稳定性" "PASS"
    else
        record_test "资源耗尽时系统稳定性" "FAIL" "成功: $success_count, 失败: $error_count"
    fi
    
    # 清理
    rm -rf "$results_dir"
}

# 测试服务重启场景
test_service_restart() {
    log_test "测试服务重启场景..."
    
    # 重启Django服务
    log_info "重启Django服务..."
    docker-compose restart django
    
    # 等待服务恢复
    if wait_for_service_recovery "Django" "$API_URL/health/"; then
        record_test "Django服务重启恢复" "PASS"
    else
        record_test "Django服务重启恢复" "FAIL"
    fi
    
    # 重启Go服务
    log_info "重启Go秒杀服务..."
    docker-compose restart gin
    
    # 等待服务恢复
    if wait_for_service_recovery "Go Seckill" "$SECKILL_URL/health"; then
        record_test "Go服务重启恢复" "PASS"
    else
        record_test "Go服务重启恢复" "FAIL"
    fi
}

# 创建测试用户
create_test_user() {
    local user_suffix="$1"
    local email="error_test_${user_suffix}@flashsku.com"
    local password="testpass123"
    
    # 注册用户
    local data="{
        \"email\": \"$email\",
        \"password\": \"$password\",
        \"username\": \"errortest$user_suffix\"
    }"
    
    curl -s --max-time 10 \
        -X POST \
        -H 'Content-Type: application/json' \
        -d "$data" \
        "$API_URL/auth/register/" > /dev/null 2>&1
    
    # 登录获取token
    local login_data="{
        \"email\": \"$email\",
        \"password\": \"$password\"
    }"
    
    local login_response=$(curl -s --max-time 10 \
        -X POST \
        -H 'Content-Type: application/json' \
        -d "$login_data" \
        "$API_URL/auth/login/" 2>/dev/null)
    
    local token=$(echo "$login_response" | grep -o '"access":"[^"]*"' | cut -d'"' -f4)
    echo "$token"
}

# 生成错误场景测试报告
generate_error_report() {
    local report_file="logs/error_scenarios_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p logs
    
    {
        echo "========================================"
        echo "Flash Sku 错误场景测试报告"
        echo "测试时间: $(date)"
        echo "========================================"
        echo ""
        echo "测试结果统计:"
        echo "总测试数: $TOTAL_TESTS"
        echo "通过测试: $PASSED_TESTS"
        echo "失败测试: $FAILED_TESTS"
        echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
        echo ""
        echo "测试场景:"
        echo "- 数据库连接失败"
        echo "- Redis连接失败"
        echo "- RabbitMQ连接失败"
        echo "- 网络超时"
        echo "- 无效请求"
        echo "- 资源耗尽"
        echo "- 服务重启"
        echo ""
        echo "系统状态:"
        docker-compose ps
    } > "$report_file"
    
    log_info "错误场景测试报告已生成: $report_file"
}

# 主测试流程
main() {
    echo "========================================"
    echo "🚨 Flash Sku 错误场景测试"
    echo "========================================"
    
    log_info "开始错误场景测试..."
    
    # 确保所有服务正常运行
    log_info "确保所有服务正常运行..."
    docker-compose up -d
    sleep 30
    
    # 执行各种错误场景测试
    test_invalid_requests
    test_network_timeout
    test_resource_exhaustion
    test_service_restart
    
    # 基础设施失败测试（这些测试会重启服务）
    test_redis_failure
    test_rabbitmq_failure
    test_database_failure  # 最后测试，因为影响最大
    
    # 生成报告
    generate_error_report
    
    # 显示结果
    echo ""
    echo "========================================"
    echo "错误场景测试完成！"
    echo "总测试数: $TOTAL_TESTS"
    echo "通过测试: $PASSED_TESTS"
    echo "失败测试: $FAILED_TESTS"
    echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo "========================================"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_info "所有错误场景测试通过！系统容错能力良好 🎉"
        exit 0
    else
        log_error "部分错误场景测试失败！需要改进系统容错能力"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Flash Sku 错误场景测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  --skip-infrastructure   跳过基础设施失败测试"
    echo ""
    echo "示例:"
    echo "  $0                      # 执行所有错误场景测试"
    echo "  $0 --skip-infrastructure # 跳过基础设施测试"
}

# 解析命令行参数
SKIP_INFRASTRUCTURE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --skip-infrastructure)
            SKIP_INFRASTRUCTURE=true
            shift
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行主函数
main
