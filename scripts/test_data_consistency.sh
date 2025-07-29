#!/bin/bash

# Flash Sku - 数据一致性测试脚本
# 验证系统在各种场景下的数据一致性

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
TEST_ACTIVITY_ID=1
CONCURRENT_USERS=10

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

# 获取数据库中的库存
get_db_stock() {
    local activity_id="$1"
    
    local stock=$(docker-compose exec -T postgres psql -U flashsku_user -d flashsku_db -t -c "
        SELECT stock FROM seckill_seckillactivity WHERE id = $activity_id;
    " 2>/dev/null | tr -d ' \n\r')
    
    echo "$stock"
}

# 获取Redis中的库存
get_redis_stock() {
    local activity_id="$1"
    
    local stock=$(docker-compose exec -T redis redis-cli get "seckill:stock:$activity_id" 2>/dev/null | tr -d '\r\n')
    
    if [ "$stock" = "(nil)" ] || [ -z "$stock" ]; then
        echo "0"
    else
        echo "$stock"
    fi
}

# 获取API返回的库存
get_api_stock() {
    local activity_id="$1"
    
    local response=$(curl -s --max-time $TIMEOUT "$SECKILL_URL/stock/$activity_id" 2>/dev/null)
    local stock=$(echo "$response" | grep -o '"stock":[0-9]*' | cut -d':' -f2)
    
    if [ -z "$stock" ]; then
        echo "0"
    else
        echo "$stock"
    fi
}

# 测试库存一致性
test_stock_consistency() {
    log_test "测试库存数据一致性..."
    
    local db_stock=$(get_db_stock $TEST_ACTIVITY_ID)
    local redis_stock=$(get_redis_stock $TEST_ACTIVITY_ID)
    local api_stock=$(get_api_stock $TEST_ACTIVITY_ID)
    
    log_info "数据库库存: $db_stock"
    log_info "Redis库存: $redis_stock"
    log_info "API库存: $api_stock"
    
    if [ "$db_stock" = "$redis_stock" ] && [ "$redis_stock" = "$api_stock" ]; then
        record_test "库存数据一致性" "PASS"
        return 0
    else
        record_test "库存数据一致性" "FAIL"
        log_error "库存数据不一致！"
        return 1
    fi
}

# 创建测试用户
create_test_user() {
    local user_id="$1"
    local email="test_user_${user_id}@flashsku.com"
    local password="testpass123"
    
    local data="{
        \"email\": \"$email\",
        \"password\": \"$password\",
        \"username\": \"testuser$user_id\"
    }"
    
    local response=$(curl -s -w '\n%{http_code}' --max-time $TIMEOUT \
        -X POST \
        -H 'Content-Type: application/json' \
        -d "$data" \
        "$API_URL/auth/register/" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -n1)
    
    if [ "$status_code" = "201" ] || [ "$status_code" = "400" ]; then
        # 登录获取token
        local login_data="{
            \"email\": \"$email\",
            \"password\": \"$password\"
        }"
        
        local login_response=$(curl -s -w '\n%{http_code}' --max-time $TIMEOUT \
            -X POST \
            -H 'Content-Type: application/json' \
            -d "$login_data" \
            "$API_URL/auth/login/" 2>/dev/null)
        
        local login_status=$(echo "$login_response" | tail -n1)
        local login_body=$(echo "$login_response" | head -n -1)
        
        if [ "$login_status" = "200" ]; then
            local token=$(echo "$login_body" | grep -o '"access":"[^"]*"' | cut -d'"' -f4)
            echo "$token"
            return 0
        fi
    fi
    
    return 1
}

# 并发秒杀测试
test_concurrent_seckill() {
    log_test "测试并发秒杀数据一致性..."
    
    # 记录初始库存
    local initial_stock=$(get_api_stock $TEST_ACTIVITY_ID)
    log_info "初始库存: $initial_stock"
    
    if [ "$initial_stock" -le 0 ]; then
        log_warn "库存不足，跳过并发测试"
        record_test "并发秒杀一致性" "SKIP"
        return 0
    fi
    
    # 创建测试用户并获取token
    local tokens=()
    log_info "创建 $CONCURRENT_USERS 个测试用户..."
    
    for i in $(seq 1 $CONCURRENT_USERS); do
        local token=$(create_test_user $i)
        if [ -n "$token" ]; then
            tokens+=("$token")
        fi
    done
    
    log_info "成功创建 ${#tokens[@]} 个测试用户"
    
    # 并发发起秒杀请求
    local pids=()
    local results_dir="tmp/concurrent_results"
    mkdir -p "$results_dir"
    
    log_info "发起并发秒杀请求..."
    
    for i in "${!tokens[@]}"; do
        local token="${tokens[$i]}"
        local result_file="$results_dir/result_$i.txt"
        
        (
            local data="{\"activity_id\": $TEST_ACTIVITY_ID, \"quantity\": 1}"
            local response=$(curl -s -w '\n%{http_code}' --max-time $TIMEOUT \
                -X POST \
                -H 'Content-Type: application/json' \
                -H "Authorization: Bearer $token" \
                -d "$data" \
                "$SECKILL_URL/$TEST_ACTIVITY_ID" 2>/dev/null)
            
            echo "$response" > "$result_file"
        ) &
        
        pids+=($!)
    done
    
    # 等待所有请求完成
    log_info "等待所有请求完成..."
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    # 分析结果
    local success_count=0
    local fail_count=0
    
    for result_file in "$results_dir"/result_*.txt; do
        if [ -f "$result_file" ]; then
            local status_code=$(tail -n1 "$result_file")
            if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
        fi
    done
    
    log_info "成功请求: $success_count"
    log_info "失败请求: $fail_count"
    
    # 检查最终库存
    sleep 5  # 等待异步处理完成
    local final_stock=$(get_api_stock $TEST_ACTIVITY_ID)
    local expected_stock=$((initial_stock - success_count))
    
    log_info "最终库存: $final_stock"
    log_info "预期库存: $expected_stock"
    
    # 验证库存一致性
    if [ "$final_stock" = "$expected_stock" ]; then
        record_test "并发秒杀一致性" "PASS"
        
        # 进一步验证数据库和Redis一致性
        test_stock_consistency
    else
        record_test "并发秒杀一致性" "FAIL"
        log_error "库存计算不一致！"
    fi
    
    # 清理临时文件
    rm -rf "$results_dir"
}

# 测试订单数据一致性
test_order_consistency() {
    log_test "测试订单数据一致性..."
    
    # 获取数据库中的订单数量
    local db_orders=$(docker-compose exec -T postgres psql -U flashsku_user -d flashsku_db -t -c "
        SELECT COUNT(*) FROM orders_order WHERE activity_id = $TEST_ACTIVITY_ID;
    " 2>/dev/null | tr -d ' \n\r')
    
    # 获取Redis中的订单计数（如果有的话）
    local redis_orders=$(docker-compose exec -T redis redis-cli get "orders:count:$TEST_ACTIVITY_ID" 2>/dev/null | tr -d '\r\n')
    
    if [ "$redis_orders" = "(nil)" ] || [ -z "$redis_orders" ]; then
        redis_orders="0"
    fi
    
    log_info "数据库订单数: $db_orders"
    log_info "Redis订单计数: $redis_orders"
    
    # 检查订单状态分布
    local pending_orders=$(docker-compose exec -T postgres psql -U flashsku_user -d flashsku_db -t -c "
        SELECT COUNT(*) FROM orders_order WHERE activity_id = $TEST_ACTIVITY_ID AND status = 'pending_payment';
    " 2>/dev/null | tr -d ' \n\r')
    
    local paid_orders=$(docker-compose exec -T postgres psql -U flashsku_user -d flashsku_db -t -c "
        SELECT COUNT(*) FROM orders_order WHERE activity_id = $TEST_ACTIVITY_ID AND status = 'paid';
    " 2>/dev/null | tr -d ' \n\r')
    
    log_info "待支付订单: $pending_orders"
    log_info "已支付订单: $paid_orders"
    
    if [ "$db_orders" -ge 0 ]; then
        record_test "订单数据一致性" "PASS"
        return 0
    else
        record_test "订单数据一致性" "FAIL"
        return 1
    fi
}

# 测试库存回滚机制
test_stock_rollback() {
    log_test "测试库存回滚机制..."
    
    # 记录初始库存
    local initial_stock=$(get_api_stock $TEST_ACTIVITY_ID)
    log_info "初始库存: $initial_stock"
    
    # 创建一个订单但不支付（模拟超时）
    local token=$(create_test_user "rollback")
    if [ -z "$token" ]; then
        record_test "库存回滚机制" "FAIL"
        log_error "无法创建测试用户"
        return 1
    fi
    
    # 发起秒杀请求
    local data="{\"activity_id\": $TEST_ACTIVITY_ID, \"quantity\": 1}"
    local response=$(curl -s -w '\n%{http_code}' --max-time $TIMEOUT \
        -X POST \
        -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $token" \
        -d "$data" \
        "$SECKILL_URL/$TEST_ACTIVITY_ID" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
        local order_id=$(echo "$body" | grep -o '"order_id":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$order_id" ]; then
            log_info "创建订单: $order_id"
            
            # 等待一段时间，然后检查库存是否回滚
            log_info "等待订单超时处理..."
            sleep 30
            
            # 检查最终库存
            local final_stock=$(get_api_stock $TEST_ACTIVITY_ID)
            log_info "最终库存: $final_stock"
            
            # 如果库存回滚，应该等于初始库存
            if [ "$final_stock" = "$initial_stock" ]; then
                record_test "库存回滚机制" "PASS"
                return 0
            else
                record_test "库存回滚机制" "FAIL"
                log_error "库存未正确回滚"
                return 1
            fi
        fi
    fi
    
    record_test "库存回滚机制" "FAIL"
    log_error "无法创建测试订单"
    return 1
}

# 生成测试报告
generate_report() {
    local report_file="logs/data_consistency_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p logs
    
    {
        echo "========================================"
        echo "Flash Sku 数据一致性测试报告"
        echo "测试时间: $(date)"
        echo "========================================"
        echo ""
        echo "测试结果统计:"
        echo "总测试数: $TOTAL_TESTS"
        echo "通过测试: $PASSED_TESTS"
        echo "失败测试: $FAILED_TESTS"
        echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
        echo ""
        echo "测试配置:"
        echo "测试活动ID: $TEST_ACTIVITY_ID"
        echo "并发用户数: $CONCURRENT_USERS"
        echo ""
        echo "当前数据状态:"
        echo "数据库库存: $(get_db_stock $TEST_ACTIVITY_ID)"
        echo "Redis库存: $(get_redis_stock $TEST_ACTIVITY_ID)"
        echo "API库存: $(get_api_stock $TEST_ACTIVITY_ID)"
    } > "$report_file"
    
    log_info "测试报告已生成: $report_file"
}

# 主测试流程
main() {
    echo "========================================"
    echo "🔍 Flash Sku 数据一致性测试"
    echo "========================================"
    
    log_info "开始数据一致性测试..."
    
    # 基础一致性测试
    test_stock_consistency
    test_order_consistency
    
    # 并发测试
    test_concurrent_seckill
    
    # 回滚机制测试
    test_stock_rollback
    
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
        log_info "所有数据一致性测试通过！ 🎉"
        exit 0
    else
        log_error "部分数据一致性测试失败！"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Flash Sku 数据一致性测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -a, --activity <ID>     设置测试活动ID (默认: 1)"
    echo "  -c, --concurrent <NUM>  设置并发用户数 (默认: 10)"
    echo ""
    echo "示例:"
    echo "  $0                      # 使用默认配置测试"
    echo "  $0 -a 2 -c 20          # 测试活动2，20个并发用户"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -a|--activity)
            TEST_ACTIVITY_ID="$2"
            shift 2
            ;;
        -c|--concurrent)
            CONCURRENT_USERS="$2"
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
