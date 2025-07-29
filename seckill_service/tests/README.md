# 测试目录 / Tests Directory

这个目录包含Go秒杀服务的所有测试文件。
This directory contains all test files for the Go seckill service.

## 目录结构 / Directory Structure

```
tests/
├── unit/           # 单元测试 / Unit tests
├── integration/    # 集成测试 / Integration tests
├── performance/    # 性能测试 / Performance tests
├── fixtures/       # 测试数据 / Test fixtures
└── README.md       # 本文件 / This file
```

## 测试规范 / Testing Guidelines

### 单元测试 / Unit Tests
- 测试单个函数或方法
- 使用mock对象隔离依赖
- 覆盖正常和异常情况
- 文件命名: `*_test.go`

### 集成测试 / Integration Tests
- 测试多个组件的交互
- 使用真实的外部依赖（Redis、RabbitMQ等）
- 测试完整的业务流程
- 文件命名: `*_integration_test.go`

### 性能测试 / Performance Tests
- 测试系统性能指标
- 压力测试和负载测试
- QPS、延迟、吞吐量测试
- 文件命名: `*_benchmark_test.go`

## 运行测试 / Running Tests

```bash
# 运行所有测试
go test ./...

# 运行单元测试
go test ./tests/unit/...

# 运行集成测试
go test ./tests/integration/...

# 运行性能测试
go test -bench=. ./tests/performance/...

# 生成测试覆盖率报告
go test -cover ./...
```

## 测试数据 / Test Data

测试数据存放在 `fixtures/` 目录中，包括：
Test data is stored in the `fixtures/` directory, including:

- 模拟的活动数据 / Mock activity data
- 测试用户数据 / Test user data
- Redis测试数据 / Redis test data
- RabbitMQ测试消息 / RabbitMQ test messages

## 注意事项 / Notes

1. 测试应该是独立的，不依赖于执行顺序
2. 使用适当的测试数据清理机制
3. 集成测试需要外部服务支持
4. 性能测试应在稳定环境中运行
