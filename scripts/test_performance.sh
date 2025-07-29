#!/bin/bash

# Flash Sku - 性能压力测试脚本
# 测试系统在高并发场景下的性能表现

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

# 测试参数
CONCURRENT_USERS=100
TEST_DURATION=60
RAMP_UP_TIME=10
TEST_ACTIVITY_ID=1

# 性能指标阈值
MAX_RESPONSE_TIME=2000  # 最大响应时间(ms)
MIN_SUCCESS_RATE=95     # 最小成功率(%)
MAX_ERROR_RATE=5        # 最大错误率(%)

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

# 检查依赖工具
check_dependencies() {
    log_test "检查测试工具依赖..."
    
    # 检查Apache Bench (ab)
    if ! command -v ab &> /dev/null; then
        log_warn "Apache Bench (ab) 未安装，尝试安装..."
        if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y apache2-utils
        elif command -v yum &> /dev/null; then
            sudo yum install -y httpd-tools
        else
            log_error "无法自动安装 Apache Bench，请手动安装"
            exit 1
        fi
    fi
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装，请先安装 curl"
        exit 1
    fi
    
    log_info "测试工具依赖检查完成"
}

# 等待服务就绪
wait_for_services() {
    log_test "等待服务就绪..."
    
    local services=(
        "$BASE_URL/health"
        "$API_URL/health/"
        "$SECKILL_URL/health"
    )
    
    for url in "${services[@]}"; do
        local attempts=0
        local max_attempts=30
        
        while [ $attempts -lt $max_attempts ]; do
            if curl -f -s --max-time 5 "$url" > /dev/null 2>&1; then
                break
            fi
            ((attempts++))
            sleep 2
        done
        
        if [ $attempts -eq $max_attempts ]; then
            log_error "服务 $url 未就绪"
            return 1
        fi
    done
    
    log_info "所有服务已就绪"
}

# 基准性能测试
baseline_performance_test() {
    log_test "执行基准性能测试..."
    
    local results_dir="tmp/performance_results"
    mkdir -p "$results_dir"
    
    # 测试健康检查接口
    log_info "测试健康检查接口性能..."
    ab -n 1000 -c 10 -g "$results_dir/health_check.tsv" "$BASE_URL/health" > "$results_dir/health_check.txt" 2>&1
    
    # 测试API接口
    log_info "测试API接口性能..."
    ab -n 1000 -c 10 -g "$results_dir/api_health.tsv" "$API_URL/health/" > "$results_dir/api_health.txt" 2>&1
    
    # 测试秒杀接口
    log_info "测试秒杀接口性能..."
    ab -n 1000 -c 10 -g "$results_dir/seckill_health.tsv" "$SECKILL_URL/health" > "$results_dir/seckill_health.txt" 2>&1
    
    # 分析结果
    analyze_ab_results "$results_dir/health_check.txt" "健康检查接口"
    analyze_ab_results "$results_dir/api_health.txt" "API健康接口"
    analyze_ab_results "$results_dir/seckill_health.txt" "秒杀健康接口"
}

# 分析Apache Bench结果
analyze_ab_results() {
    local result_file="$1"
    local test_name="$2"
    
    if [ ! -f "$result_file" ]; then
        log_error "结果文件不存在: $result_file"
        return 1
    fi
    
    log_info "分析 $test_name 测试结果..."
    
    # 提取关键指标
    local requests_per_sec=$(grep "Requests per second:" "$result_file" | awk '{print $4}')
    local mean_time=$(grep "Time per request:" "$result_file" | head -n1 | awk '{print $4}')
    local failed_requests=$(grep "Failed requests:" "$result_file" | awk '{print $3}')
    local total_requests=$(grep "Complete requests:" "$result_file" | awk '{print $3}')
    
    # 计算成功率
    local success_rate=100
    if [ "$total_requests" -gt 0 ] && [ "$failed_requests" -gt 0 ]; then
        success_rate=$(echo "scale=2; (($total_requests - $failed_requests) * 100) / $total_requests" | bc)
    fi
    
    echo "  请求/秒: $requests_per_sec"
    echo "  平均响应时间: ${mean_time}ms"
    echo "  失败请求: $failed_requests"
    echo "  成功率: ${success_rate}%"
    
    # 性能评估
    if (( $(echo "$mean_time > $MAX_RESPONSE_TIME" | bc -l) )); then
        log_warn "$test_name 响应时间超过阈值 (${mean_time}ms > ${MAX_RESPONSE_TIME}ms)"
    fi
    
    if (( $(echo "$success_rate < $MIN_SUCCESS_RATE" | bc -l) )); then
        log_warn "$test_name 成功率低于阈值 (${success_rate}% < ${MIN_SUCCESS_RATE}%)"
    fi
}

# 高并发秒杀测试
concurrent_seckill_test() {
    log_test "执行高并发秒杀测试..."
    
    # 创建测试用户
    local tokens=()
    log_info "创建 $CONCURRENT_USERS 个测试用户..."
    
    for i in $(seq 1 $CONCURRENT_USERS); do
        local token=$(create_test_user $i)
        if [ -n "$token" ]; then
            tokens+=("$token")
        fi
        
        # 每创建10个用户显示一次进度
        if [ $((i % 10)) -eq 0 ]; then
            log_info "已创建 $i/$CONCURRENT_USERS 个用户"
        fi
    done
    
    log_info "成功创建 ${#tokens[@]} 个测试用户"
    
    # 记录初始库存
    local initial_stock=$(get_api_stock $TEST_ACTIVITY_ID)
    log_info "初始库存: $initial_stock"
    
    # 并发秒杀测试
    local results_dir="tmp/concurrent_seckill"
    mkdir -p "$results_dir"
    
    local start_time=$(date +%s.%N)
    local pids=()
    
    log_info "启动 ${#tokens[@]} 个并发秒杀请求..."
    
    for i in "${!tokens[@]}"; do
        local token="${tokens[$i]}"
        local result_file="$results_dir/result_$i.txt"
        
        (
            local request_start=$(date +%s.%N)
            local data="{\"activity_id\": $TEST_ACTIVITY_ID, \"quantity\": 1}"
            local response=$(curl -s -w '\n%{http_code}\n%{time_total}' --max-time 30 \
                -X POST \
                -H 'Content-Type: application/json' \
                -H "Authorization: Bearer $token" \
                -d "$data" \
                "$SECKILL_URL/$TEST_ACTIVITY_ID" 2>/dev/null)
            local request_end=$(date +%s.%N)
            
            echo "start_time:$request_start" > "$result_file"
            echo "end_time:$request_end" >> "$result_file"
            echo "$response" >> "$result_file"
        ) &
        
        pids+=($!)
        
        # 控制启动速度，避免瞬间压力过大
        if [ $((i % 10)) -eq 0 ] && [ $i -gt 0 ]; then
            sleep 0.1
        fi
    done
    
    # 等待所有请求完成
    log_info "等待所有请求完成..."
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    local end_time=$(date +%s.%N)
    local total_time=$(echo "$end_time - $start_time" | bc)
    
    log_info "所有请求完成，总耗时: ${total_time}秒"
    
    # 分析并发测试结果
    analyze_concurrent_results "$results_dir" "$total_time"
    
    # 清理
    rm -rf "$results_dir"
}

# 分析并发测试结果
analyze_concurrent_results() {
    local results_dir="$1"
    local total_time="$2"
    
    log_info "分析并发测试结果..."
    
    local success_count=0
    local fail_count=0
    local total_response_time=0
    local min_response_time=999999
    local max_response_time=0
    
    for result_file in "$results_dir"/result_*.txt; do
        if [ -f "$result_file" ]; then
            local status_code=$(grep -v "start_time\|end_time" "$result_file" | tail -n2 | head -n1)
            local response_time=$(grep -v "start_time\|end_time" "$result_file" | tail -n1)
            
            if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
                ((success_count++))
            else
                ((fail_count++))
            fi
            
            # 计算响应时间统计
            if [ -n "$response_time" ] && [ "$response_time" != "0.000000" ]; then
                total_response_time=$(echo "$total_response_time + $response_time" | bc)
                
                if (( $(echo "$response_time < $min_response_time" | bc -l) )); then
                    min_response_time=$response_time
                fi
                
                if (( $(echo "$response_time > $max_response_time" | bc -l) )); then
                    max_response_time=$response_time
                fi
            fi
        fi
    done
    
    local total_requests=$((success_count + fail_count))
    local success_rate=0
    local avg_response_time=0
    local requests_per_sec=0
    
    if [ $total_requests -gt 0 ]; then
        success_rate=$(echo "scale=2; ($success_count * 100) / $total_requests" | bc)
        requests_per_sec=$(echo "scale=2; $total_requests / $total_time" | bc)
    fi
    
    if [ $success_count -gt 0 ]; then
        avg_response_time=$(echo "scale=3; $total_response_time / $success_count" | bc)
    fi
    
    echo ""
    echo "========================================"
    echo "并发秒杀测试结果"
    echo "========================================"
    echo "总请求数: $total_requests"
    echo "成功请求: $success_count"
    echo "失败请求: $fail_count"
    echo "成功率: ${success_rate}%"
    echo "总耗时: ${total_time}秒"
    echo "请求/秒: $requests_per_sec"
    echo "平均响应时间: ${avg_response_time}秒"
    echo "最小响应时间: ${min_response_time}秒"
    echo "最大响应时间: ${max_response_time}秒"
    echo "========================================"
    
    # 性能评估
    local performance_issues=0
    
    if (( $(echo "$success_rate < $MIN_SUCCESS_RATE" | bc -l) )); then
        log_warn "成功率低于阈值: ${success_rate}% < ${MIN_SUCCESS_RATE}%"
        ((performance_issues++))
    fi
    
    local avg_response_ms=$(echo "$avg_response_time * 1000" | bc)
    if (( $(echo "$avg_response_ms > $MAX_RESPONSE_TIME" | bc -l) )); then
        log_warn "平均响应时间超过阈值: ${avg_response_ms}ms > ${MAX_RESPONSE_TIME}ms"
        ((performance_issues++))
    fi
    
    if [ $performance_issues -eq 0 ]; then
        log_info "并发性能测试通过！"
    else
        log_error "发现 $performance_issues 个性能问题"
    fi
}

# 创建测试用户
create_test_user() {
    local user_id="$1"
    local email="perf_user_${user_id}@flashsku.com"
    local password="testpass123"
    
    # 注册用户
    local data="{
        \"email\": \"$email\",
        \"password\": \"$password\",
        \"username\": \"perfuser$user_id\"
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

# 获取API库存
get_api_stock() {
    local activity_id="$1"
    
    local response=$(curl -s --max-time 10 "$SECKILL_URL/stock/$activity_id" 2>/dev/null)
    local stock=$(echo "$response" | grep -o '"stock":[0-9]*' | cut -d':' -f2)
    
    if [ -z "$stock" ]; then
        echo "0"
    else
        echo "$stock"
    fi
}

# 系统资源监控
monitor_system_resources() {
    log_test "监控系统资源使用..."
    
    local monitor_duration=60
    local monitor_interval=5
    local monitor_file="tmp/resource_monitor.log"
    
    mkdir -p tmp
    
    (
        for i in $(seq 1 $((monitor_duration / monitor_interval))); do
            echo "$(date '+%Y-%m-%d %H:%M:%S')" >> "$monitor_file"
            docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" | grep flashsku >> "$monitor_file"
            echo "---" >> "$monitor_file"
            sleep $monitor_interval
        done
    ) &
    
    local monitor_pid=$!
    
    # 执行性能测试
    concurrent_seckill_test
    
    # 停止监控
    kill $monitor_pid 2>/dev/null || true
    
    # 分析资源使用
    if [ -f "$monitor_file" ]; then
        log_info "系统资源使用分析:"
        echo "详细监控日志: $monitor_file"
    fi
}

# 生成性能测试报告
generate_performance_report() {
    local report_file="logs/performance_report_$(date +%Y%m%d_%H%M%S).txt"
    mkdir -p logs
    
    {
        echo "========================================"
        echo "Flash Sku 性能测试报告"
        echo "测试时间: $(date)"
        echo "========================================"
        echo ""
        echo "测试配置:"
        echo "并发用户数: $CONCURRENT_USERS"
        echo "测试持续时间: $TEST_DURATION 秒"
        echo "测试活动ID: $TEST_ACTIVITY_ID"
        echo ""
        echo "性能阈值:"
        echo "最大响应时间: ${MAX_RESPONSE_TIME}ms"
        echo "最小成功率: ${MIN_SUCCESS_RATE}%"
        echo "最大错误率: ${MAX_ERROR_RATE}%"
        echo ""
        echo "系统状态:"
        docker-compose ps
        echo ""
        echo "资源使用:"
        docker stats --no-stream
    } > "$report_file"
    
    log_info "性能测试报告已生成: $report_file"
}

# 主测试流程
main() {
    echo "========================================"
    echo "⚡ Flash Sku 性能压力测试"
    echo "========================================"
    
    # 检查依赖
    check_dependencies
    
    # 等待服务就绪
    if ! wait_for_services; then
        log_error "服务未就绪，测试终止"
        exit 1
    fi
    
    # 执行测试
    log_info "开始性能压力测试..."
    
    # 基准性能测试
    baseline_performance_test
    
    # 高并发测试
    monitor_system_resources
    
    # 生成报告
    generate_performance_report
    
    log_info "性能测试完成！"
}

# 显示帮助信息
show_help() {
    echo "Flash Sku 性能压力测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -c, --concurrent <NUM>  并发用户数 (默认: 100)"
    echo "  -d, --duration <SEC>    测试持续时间 (默认: 60)"
    echo "  -a, --activity <ID>     测试活动ID (默认: 1)"
    echo ""
    echo "示例:"
    echo "  $0                      # 使用默认配置"
    echo "  $0 -c 200 -d 120       # 200并发，持续120秒"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--concurrent)
            CONCURRENT_USERS="$2"
            shift 2
            ;;
        -d|--duration)
            TEST_DURATION="$2"
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
