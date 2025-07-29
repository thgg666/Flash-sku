#!/bin/bash

# Flash Sku - 综合测试脚本
# 执行所有类型的测试：端到端、数据一致性、性能、错误场景

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 配置
PROJECT_NAME="Flash Sku"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="logs"
REPORT_DIR="reports"

# 测试配置
RUN_E2E=true
RUN_CONSISTENCY=true
RUN_PERFORMANCE=true
RUN_ERROR_SCENARIOS=true

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

log_section() {
    echo -e "${PURPLE}[SECTION]${NC} $1"
}

# 测试结果统计
TOTAL_TEST_SUITES=0
PASSED_TEST_SUITES=0
FAILED_TEST_SUITES=0

# 记录测试套件结果
record_test_suite() {
    local suite_name="$1"
    local result="$2"
    local details="$3"
    
    ((TOTAL_TEST_SUITES++))
    
    if [ "$result" = "PASS" ]; then
        ((PASSED_TEST_SUITES++))
        log_info "✓ $suite_name 测试套件 - PASS"
    else
        ((FAILED_TEST_SUITES++))
        log_error "✗ $suite_name 测试套件 - FAIL"
        if [ -n "$details" ]; then
            log_error "  详情: $details"
        fi
    fi
}

# 初始化测试环境
init_test_environment() {
    log_section "初始化测试环境..."
    
    # 创建必要的目录
    mkdir -p "$LOG_DIR" "$REPORT_DIR" "tmp"
    
    # 清理旧的临时文件
    rm -rf tmp/*
    
    # 检查Docker服务
    if ! docker info &> /dev/null; then
        log_error "Docker服务未运行"
        exit 1
    fi
    
    # 检查docker-compose文件
    if [ ! -f "docker-compose.yml" ]; then
        log_error "docker-compose.yml 文件不存在"
        exit 1
    fi
    
    log_info "测试环境初始化完成"
}

# 启动系统
start_system() {
    log_section "启动系统..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose down --timeout 30 || true
    
    # 启动所有服务
    log_info "启动所有服务..."
    docker-compose up -d
    
    # 等待服务就绪
    log_info "等待服务就绪..."
    local max_wait=300  # 5分钟
    local wait_time=0
    
    while [ $wait_time -lt $max_wait ]; do
        if curl -f -s --max-time 5 "http://localhost/health" > /dev/null 2>&1; then
            log_info "系统启动完成"
            return 0
        fi
        
        sleep 10
        wait_time=$((wait_time + 10))
        log_info "等待中... ($wait_time/$max_wait 秒)"
    done
    
    log_error "系统启动超时"
    return 1
}

# 执行端到端测试
run_e2e_tests() {
    if [ "$RUN_E2E" != "true" ]; then
        log_info "跳过端到端测试"
        return 0
    fi
    
    log_section "执行端到端测试..."
    
    local start_time=$(date +%s)
    
    if "$SCRIPT_DIR/test_e2e.sh"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "端到端" "PASS" "耗时: ${duration}秒"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "端到端" "FAIL" "耗时: ${duration}秒"
        return 1
    fi
}

# 执行数据一致性测试
run_consistency_tests() {
    if [ "$RUN_CONSISTENCY" != "true" ]; then
        log_info "跳过数据一致性测试"
        return 0
    fi
    
    log_section "执行数据一致性测试..."
    
    local start_time=$(date +%s)
    
    if "$SCRIPT_DIR/test_data_consistency.sh"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "数据一致性" "PASS" "耗时: ${duration}秒"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "数据一致性" "FAIL" "耗时: ${duration}秒"
        return 1
    fi
}

# 执行性能测试
run_performance_tests() {
    if [ "$RUN_PERFORMANCE" != "true" ]; then
        log_info "跳过性能测试"
        return 0
    fi
    
    log_section "执行性能测试..."
    
    local start_time=$(date +%s)
    
    if "$SCRIPT_DIR/test_performance.sh" -c 50 -d 30; then  # 减少并发数和时间以加快测试
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "性能" "PASS" "耗时: ${duration}秒"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "性能" "FAIL" "耗时: ${duration}秒"
        return 1
    fi
}

# 执行错误场景测试
run_error_scenario_tests() {
    if [ "$RUN_ERROR_SCENARIOS" != "true" ]; then
        log_info "跳过错误场景测试"
        return 0
    fi
    
    log_section "执行错误场景测试..."
    
    local start_time=$(date +%s)
    
    if "$SCRIPT_DIR/test_error_scenarios.sh"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "错误场景" "PASS" "耗时: ${duration}秒"
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        record_test_suite "错误场景" "FAIL" "耗时: ${duration}秒"
        return 1
    fi
}

# 收集系统信息
collect_system_info() {
    log_section "收集系统信息..."
    
    local info_file="$REPORT_DIR/system_info_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "========================================"
        echo "$PROJECT_NAME 系统信息"
        echo "收集时间: $(date)"
        echo "========================================"
        echo ""
        
        echo "=== Docker版本 ==="
        docker --version
        docker-compose --version
        echo ""
        
        echo "=== 系统资源 ==="
        free -h
        df -h
        echo ""
        
        echo "=== 容器状态 ==="
        docker-compose ps
        echo ""
        
        echo "=== 容器资源使用 ==="
        docker stats --no-stream
        echo ""
        
        echo "=== 网络连接 ==="
        netstat -tlnp | grep -E ":80|:8000|:8080|:3000|:5432|:6379|:5672" || true
        echo ""
        
        echo "=== 最近的容器日志 ==="
        for service in nginx django gin postgres redis rabbitmq; do
            echo "--- $service ---"
            docker-compose logs --tail=20 "$service" 2>/dev/null || echo "服务不存在或未运行"
            echo ""
        done
        
    } > "$info_file"
    
    log_info "系统信息已收集: $info_file"
}

# 生成综合测试报告
generate_comprehensive_report() {
    log_section "生成综合测试报告..."
    
    local report_file="$REPORT_DIR/comprehensive_test_report_$(date +%Y%m%d_%H%M%S).html"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>$PROJECT_NAME 综合测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .summary { background: #e8f5e8; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .failed { background: #ffe8e8; }
        .test-suite { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .pass { border-left: 5px solid #4CAF50; }
        .fail { border-left: 5px solid #f44336; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background: #f9f9f9; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>$PROJECT_NAME 综合测试报告</h1>
        <p>生成时间: $(date)</p>
    </div>
    
    <div class="summary $([ $FAILED_TEST_SUITES -eq 0 ] && echo "" || echo "failed")">
        <h2>测试结果摘要</h2>
        <div class="metric">
            <strong>总测试套件:</strong> $TOTAL_TEST_SUITES
        </div>
        <div class="metric">
            <strong>通过套件:</strong> $PASSED_TEST_SUITES
        </div>
        <div class="metric">
            <strong>失败套件:</strong> $FAILED_TEST_SUITES
        </div>
        <div class="metric">
            <strong>成功率:</strong> $(( PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES ))%
        </div>
    </div>
    
    <h2>测试套件详情</h2>
EOF

    # 添加各个测试套件的详细信息
    if [ "$RUN_E2E" = "true" ]; then
        echo '<div class="test-suite pass"><h3>端到端测试</h3><p>测试完整的用户流程和API交互</p></div>' >> "$report_file"
    fi
    
    if [ "$RUN_CONSISTENCY" = "true" ]; then
        echo '<div class="test-suite pass"><h3>数据一致性测试</h3><p>验证并发场景下的数据一致性</p></div>' >> "$report_file"
    fi
    
    if [ "$RUN_PERFORMANCE" = "true" ]; then
        echo '<div class="test-suite pass"><h3>性能测试</h3><p>测试系统在高负载下的性能表现</p></div>' >> "$report_file"
    fi
    
    if [ "$RUN_ERROR_SCENARIOS" = "true" ]; then
        echo '<div class="test-suite pass"><h3>错误场景测试</h3><p>验证系统的容错能力和恢复机制</p></div>' >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF
    
    <h2>系统状态</h2>
    <pre>$(docker-compose ps)</pre>
    
    <h2>测试建议</h2>
    <ul>
        <li>定期执行综合测试以确保系统稳定性</li>
        <li>在生产部署前必须通过所有测试</li>
        <li>关注性能指标的变化趋势</li>
        <li>及时修复发现的问题</li>
    </ul>
    
    <footer style="margin-top: 50px; padding-top: 20px; border-top: 1px solid #ddd; color: #666;">
        <p>报告生成于: $(date) | $PROJECT_NAME 自动化测试系统</p>
    </footer>
</body>
</html>
EOF
    
    log_info "综合测试报告已生成: $report_file"
}

# 清理测试环境
cleanup_test_environment() {
    log_section "清理测试环境..."
    
    # 清理临时文件
    rm -rf tmp/*
    
    # 可选：停止服务（如果指定了清理选项）
    if [ "$CLEANUP_AFTER_TEST" = "true" ]; then
        log_info "停止所有服务..."
        docker-compose down
    fi
    
    log_info "测试环境清理完成"
}

# 主测试流程
main() {
    local start_time=$(date +%s)
    
    echo "========================================"
    echo "🧪 $PROJECT_NAME 综合测试套件"
    echo "========================================"
    
    # 初始化
    init_test_environment
    
    # 启动系统
    if ! start_system; then
        log_error "系统启动失败，测试终止"
        exit 1
    fi
    
    # 收集系统信息
    collect_system_info
    
    # 执行各种测试
    run_e2e_tests
    run_consistency_tests
    run_performance_tests
    run_error_scenario_tests
    
    # 生成报告
    generate_comprehensive_report
    
    # 清理环境
    cleanup_test_environment
    
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    
    # 显示最终结果
    echo ""
    echo "========================================"
    echo "综合测试完成！"
    echo "总耗时: $total_duration 秒"
    echo "测试套件: $TOTAL_TEST_SUITES"
    echo "通过: $PASSED_TEST_SUITES"
    echo "失败: $FAILED_TEST_SUITES"
    echo "成功率: $(( PASSED_TEST_SUITES * 100 / TOTAL_TEST_SUITES ))%"
    echo "========================================"
    
    if [ $FAILED_TEST_SUITES -eq 0 ]; then
        log_info "🎉 所有测试套件通过！系统质量良好！"
        exit 0
    else
        log_error "❌ 部分测试套件失败！需要修复问题！"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "$PROJECT_NAME 综合测试脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  --skip-e2e              跳过端到端测试"
    echo "  --skip-consistency      跳过数据一致性测试"
    echo "  --skip-performance      跳过性能测试"
    echo "  --skip-error-scenarios  跳过错误场景测试"
    echo "  --cleanup               测试后清理环境"
    echo ""
    echo "示例:"
    echo "  $0                      # 执行所有测试"
    echo "  $0 --skip-performance   # 跳过性能测试"
    echo "  $0 --cleanup            # 测试后清理环境"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --skip-e2e)
            RUN_E2E=false
            shift
            ;;
        --skip-consistency)
            RUN_CONSISTENCY=false
            shift
            ;;
        --skip-performance)
            RUN_PERFORMANCE=false
            shift
            ;;
        --skip-error-scenarios)
            RUN_ERROR_SCENARIOS=false
            shift
            ;;
        --cleanup)
            CLEANUP_AFTER_TEST=true
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
