#!/bin/bash

# Flash Sku - 端到端测试脚本
# 测试完整的秒杀流程

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

# 测试数据
TEST_USER_EMAIL="test@flashsku.com"
TEST_USER_PASSWORD="testpass123"
TEST_ACTIVITY_ID=1

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
    
    ((TOTAL_TESTS++))
    
    if [ "$result" = "PASS" ]; then
        ((PASSED_TESTS++))
        log_info "✓ $test_name - PASS"
    else
        ((FAILED_TESTS++))
        log_error "✗ $test_name - FAIL"
    fi
}

# HTTP请求函数
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"
    
    local curl_cmd="curl -s -w '\n%{http_code}' --max-time $TIMEOUT"
    
    if [ -n "$headers" ]; then
        curl_cmd="$curl_cmd $headers"
    fi
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        curl_cmd="$curl_cmd -X POST -H 'Content-Type: application/json' -d '$data'"
    elif [ "$method" = "GET" ]; then
        curl_cmd="$curl_cmd -X GET"
    fi
    
    curl_cmd="$curl_cmd '$url'"
    
    eval "$curl_cmd"
}

# 等待服务就绪
wait_for_services() {
    log_test "等待服务启动..."
    
    local services=(
        "$BASE_URL/health:系统健康检查"
        "$API_URL/health/:Django API"
        "$SECKILL_URL/health:Go秒杀API"
    )
    
    for service_info in "${services[@]}"; do
        local url="${service_info%:*}"
        local name="${service_info#*:}"
        local attempts=0
        local max_attempts=30
        
        log_info "等待 $name 服务..."
        
        while [ $attempts -lt $max_attempts ]; do
            if curl -f -s --max-time 5 "$url" > /dev/null 2>&1; then
                log_info "$name 服务就绪"
                break
            fi
            
            ((attempts++))
            sleep 2
        done
        
        if [ $attempts -eq $max_attempts ]; then
            log_error "$name 服务启动超时"
            return 1
        fi
    done
    
    log_info "所有服务已就绪"
}

# 测试用户注册
test_user_registration() {
    log_test "测试用户注册..."
    
    local data="{
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\",
        \"username\": \"testuser\"
    }"
    
    local response=$(make_request "POST" "$API_URL/auth/register/" "$data")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "201" ] || [ "$status_code" = "400" ]; then
        record_test "用户注册" "PASS"
        return 0
    else
        record_test "用户注册" "FAIL"
        log_error "注册失败: $body"
        return 1
    fi
}

# 测试用户登录
test_user_login() {
    log_test "测试用户登录..."
    
    local data="{
        \"email\": \"$TEST_USER_EMAIL\",
        \"password\": \"$TEST_USER_PASSWORD\"
    }"
    
    local response=$(make_request "POST" "$API_URL/auth/login/" "$data")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ]; then
        # 提取token
        ACCESS_TOKEN=$(echo "$body" | grep -o '"access":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$ACCESS_TOKEN" ]; then
            record_test "用户登录" "PASS"
            log_info "获取到访问令牌"
            return 0
        fi
    fi
    
    record_test "用户登录" "FAIL"
    log_error "登录失败: $body"
    return 1
}

# 测试获取秒杀活动列表
test_get_activities() {
    log_test "测试获取秒杀活动列表..."
    
    local headers="-H 'Authorization: Bearer $ACCESS_TOKEN'"
    local response=$(make_request "GET" "$API_URL/activities/" "" "$headers")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ]; then
        record_test "获取活动列表" "PASS"
        log_info "成功获取活动列表"
        return 0
    else
        record_test "获取活动列表" "FAIL"
        log_error "获取活动列表失败: $body"
        return 1
    fi
}

# 测试秒杀参与
test_seckill_participation() {
    log_test "测试秒杀参与..."
    
    local data="{
        \"activity_id\": $TEST_ACTIVITY_ID,
        \"quantity\": 1
    }"
    
    local headers="-H 'Authorization: Bearer $ACCESS_TOKEN'"
    local response=$(make_request "POST" "$SECKILL_URL/$TEST_ACTIVITY_ID" "$data" "$headers")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
        # 提取订单ID或任务ID
        ORDER_ID=$(echo "$body" | grep -o '"order_id":"[^"]*"' | cut -d'"' -f4)
        TASK_ID=$(echo "$body" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
        
        record_test "秒杀参与" "PASS"
        log_info "秒杀请求成功"
        
        if [ -n "$ORDER_ID" ]; then
            log_info "获取到订单ID: $ORDER_ID"
        fi
        
        if [ -n "$TASK_ID" ]; then
            log_info "获取到任务ID: $TASK_ID"
        fi
        
        return 0
    else
        record_test "秒杀参与" "FAIL"
        log_error "秒杀参与失败: $body"
        return 1
    fi
}

# 测试订单状态查询
test_order_status() {
    log_test "测试订单状态查询..."
    
    if [ -z "$ORDER_ID" ] && [ -z "$TASK_ID" ]; then
        record_test "订单状态查询" "SKIP"
        log_warn "没有订单ID或任务ID，跳过测试"
        return 0
    fi
    
    local headers="-H 'Authorization: Bearer $ACCESS_TOKEN'"
    local url
    
    if [ -n "$TASK_ID" ]; then
        url="$API_URL/orders/status/$TASK_ID/"
    else
        url="$API_URL/orders/$ORDER_ID/"
    fi
    
    # 轮询订单状态
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        local response=$(make_request "GET" "$url" "" "$headers")
        local status_code=$(echo "$response" | tail -n1)
        local body=$(echo "$response" | head -n -1)
        
        if [ "$status_code" = "200" ]; then
            local order_status=$(echo "$body" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
            log_info "订单状态: $order_status"
            
            if [ "$order_status" = "completed" ] || [ "$order_status" = "pending_payment" ]; then
                record_test "订单状态查询" "PASS"
                return 0
            fi
        fi
        
        ((attempts++))
        sleep 3
    done
    
    record_test "订单状态查询" "FAIL"
    log_error "订单状态查询超时"
    return 1
}

# 测试库存查询
test_stock_query() {
    log_test "测试库存查询..."
    
    local response=$(make_request "GET" "$SECKILL_URL/stock/$TEST_ACTIVITY_ID")
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ]; then
        local stock=$(echo "$body" | grep -o '"stock":[0-9]*' | cut -d':' -f2)
        record_test "库存查询" "PASS"
        log_info "当前库存: $stock"
        return 0
    else
        record_test "库存查询" "FAIL"
        log_error "库存查询失败: $body"
        return 1
    fi
}

# 测试系统健康检查
test_health_checks() {
    log_test "测试系统健康检查..."
    
    local endpoints=(
        "$BASE_URL/health:系统总体健康"
        "$API_URL/health/:Django健康检查"
        "$SECKILL_URL/health:Go服务健康检查"
        "$BASE_URL/nginx_status:Nginx状态"
    )
    
    local health_passed=0
    local health_total=${#endpoints[@]}
    
    for endpoint_info in "${endpoints[@]}"; do
        local url="${endpoint_info%:*}"
        local name="${endpoint_info#*:}"
        
        local response=$(make_request "GET" "$url")
        local status_code=$(echo "$response" | tail -n1)
        
        if [ "$status_code" = "200" ]; then
            log_info "$name - 健康"
            ((health_passed++))
        else
            log_error "$name - 异常"
        fi
    done
    
    if [ $health_passed -eq $health_total ]; then
        record_test "系统健康检查" "PASS"
    else
        record_test "系统健康检查" "FAIL"
    fi
}

# 生成测试报告
generate_report() {
    local report_file="logs/e2e_test_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p logs
    
    {
        echo "========================================"
        echo "Flash Sku 端到端测试报告"
        echo "测试时间: $(date)"
        echo "========================================"
        echo ""
        echo "测试结果统计:"
        echo "总测试数: $TOTAL_TESTS"
        echo "通过测试: $PASSED_TESTS"
        echo "失败测试: $FAILED_TESTS"
        echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
        echo ""
        echo "测试环境:"
        echo "基础URL: $BASE_URL"
        echo "API URL: $API_URL"
        echo "秒杀URL: $SECKILL_URL"
        echo ""
        echo "测试用户:"
        echo "邮箱: $TEST_USER_EMAIL"
        echo "活动ID: $TEST_ACTIVITY_ID"
        echo ""
        if [ -n "$ACCESS_TOKEN" ]; then
            echo "访问令牌: ${ACCESS_TOKEN:0:20}..."
        fi
        if [ -n "$ORDER_ID" ]; then
            echo "订单ID: $ORDER_ID"
        fi
        if [ -n "$TASK_ID" ]; then
            echo "任务ID: $TASK_ID"
        fi
    } > "$report_file"
    
    log_info "测试报告已生成: $report_file"
}

# 主测试流程
main() {
    echo "========================================"
    echo "🧪 Flash Sku 端到端测试"
    echo "========================================"
    
    # 等待服务就绪
    if ! wait_for_services; then
        log_error "服务未就绪，测试终止"
        exit 1
    fi
    
    # 执行测试
    log_info "开始执行端到端测试..."
    
    test_health_checks
    test_user_registration
    test_user_login
    test_get_activities
    test_stock_query
    test_seckill_participation
    test_order_status
    
    # 生成报告
    generate_report
    
    # 显示结果
    echo ""
    echo "========================================"
    echo "测试完成！"
    echo "总测试数: $TOTAL_TESTS"
    echo "通过测试: $PASSED_TESTS"
    echo "失败测试: $FAILED_TESTS"
    echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo "========================================"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_info "所有测试通过！ 🎉"
        exit 0
    else
        log_error "部分测试失败！"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Flash Sku 端到端测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -u, --url <URL>         设置基础URL (默认: http://localhost)"
    echo "  -e, --email <EMAIL>     设置测试用户邮箱"
    echo "  -a, --activity <ID>     设置测试活动ID"
    echo ""
    echo "示例:"
    echo "  $0                      # 使用默认配置测试"
    echo "  $0 -u http://test.com   # 指定测试URL"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--url)
            BASE_URL="$2"
            API_URL="$BASE_URL/api"
            SECKILL_URL="$BASE_URL/seckill"
            shift 2
            ;;
        -e|--email)
            TEST_USER_EMAIL="$2"
            shift 2
            ;;
        -a|--activity)
            TEST_ACTIVITY_ID="$2"
            shift 2
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
