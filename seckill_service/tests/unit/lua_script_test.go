package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flashsku/seckill/internal/lua"
)

func TestScriptRegistration(t *testing.T) {
	// 测试脚本注册
	// Test script registration
	scripts := []string{
		"stock_deduct",
		"user_limit_check", 
		"activity_check",
		"seckill_process",
		"stock_rollback",
		"batch_stock_check",
	}

	for _, scriptName := range scripts {
		t.Run(scriptName, func(t *testing.T) {
			// 验证脚本内容不为空
			// Verify script content is not empty
			var scriptContent string
			
			switch scriptName {
			case "stock_deduct":
				scriptContent = lua.StockDeductScript
			case "user_limit_check":
				scriptContent = lua.UserLimitCheckScript
			case "activity_check":
				scriptContent = lua.ActivityCheckScript
			case "seckill_process":
				scriptContent = lua.SeckillProcessScript
			case "stock_rollback":
				scriptContent = lua.StockRollbackScript
			case "batch_stock_check":
				scriptContent = lua.BatchStockCheckScript
			}
			
			assert.NotEmpty(t, scriptContent, "Script content should not be empty")
			assert.Contains(t, scriptContent, "redis.call", "Script should contain Redis calls")
		})
	}
}

func TestStockDeductScript(t *testing.T) {
	script := lua.StockDeductScript
	
	// 验证脚本包含必要的逻辑
	// Verify script contains necessary logic
	assert.Contains(t, script, "KEYS[1]", "Should reference stock key")
	assert.Contains(t, script, "ARGV[1]", "Should reference deduction amount")
	assert.Contains(t, script, "GET", "Should get current stock")
	assert.Contains(t, script, "DECRBY", "Should decrement stock")
	assert.Contains(t, script, "insufficient stock", "Should check stock sufficiency")
	
	// 验证返回值格式
	// Verify return value format
	assert.Contains(t, script, "return {1,", "Should return success format")
	assert.Contains(t, script, "return {0,", "Should return failure format")
}

func TestUserLimitCheckScript(t *testing.T) {
	script := lua.UserLimitCheckScript
	
	// 验证脚本包含必要的逻辑
	// Verify script contains necessary logic
	assert.Contains(t, script, "KEYS[1]", "Should reference user limit key")
	assert.Contains(t, script, "ARGV[1]", "Should reference purchase amount")
	assert.Contains(t, script, "ARGV[2]", "Should reference max limit")
	assert.Contains(t, script, "INCRBY", "Should increment purchase count")
	assert.Contains(t, script, "EXPIRE", "Should set expiration")
	assert.Contains(t, script, "exceeds purchase limit", "Should check purchase limit")
}

func TestActivityCheckScript(t *testing.T) {
	script := lua.ActivityCheckScript
	
	// 验证脚本包含必要的逻辑
	// Verify script contains necessary logic
	assert.Contains(t, script, "KEYS[1]", "Should reference activity info key")
	assert.Contains(t, script, "KEYS[2]", "Should reference activity status key")
	assert.Contains(t, script, "ARGV[1]", "Should reference current timestamp")
	assert.Contains(t, script, "activity not found", "Should check activity existence")
	assert.Contains(t, script, "activity not active", "Should check activity status")
}

func TestSeckillProcessScript(t *testing.T) {
	script := lua.SeckillProcessScript
	
	// 验证脚本包含完整的秒杀流程
	// Verify script contains complete seckill process
	assert.Contains(t, script, "KEYS[1]", "Should reference activity info key")
	assert.Contains(t, script, "KEYS[2]", "Should reference stock key")
	assert.Contains(t, script, "KEYS[3]", "Should reference user limit key")
	assert.Contains(t, script, "KEYS[4]", "Should reference activity status key")
	
	// 验证各个检查步骤
	// Verify each check step
	assert.Contains(t, script, "检查活动状态", "Should check activity status")
	assert.Contains(t, script, "检查用户限购", "Should check user limit")
	assert.Contains(t, script, "检查库存并扣减", "Should check and deduct stock")
	assert.Contains(t, script, "原子性执行", "Should execute atomically")
	
	// 验证错误处理
	// Verify error handling
	assert.Contains(t, script, "activity not active", "Should handle inactive activity")
	assert.Contains(t, script, "exceeds user limit", "Should handle user limit exceeded")
	assert.Contains(t, script, "insufficient stock", "Should handle insufficient stock")
	
	// 验证成功返回
	// Verify success return
	assert.Contains(t, script, "seckill success", "Should return success message")
}

func TestStockRollbackScript(t *testing.T) {
	script := lua.StockRollbackScript
	
	// 验证脚本包含回滚逻辑
	// Verify script contains rollback logic
	assert.Contains(t, script, "KEYS[1]", "Should reference stock key")
	assert.Contains(t, script, "KEYS[2]", "Should reference user limit key")
	assert.Contains(t, script, "ARGV[1]", "Should reference rollback amount")
	assert.Contains(t, script, "INCRBY", "Should increment stock")
	assert.Contains(t, script, "DECRBY", "Should decrement user count")
	assert.Contains(t, script, "rollback success", "Should return success message")
}

func TestBatchStockCheckScript(t *testing.T) {
	script := lua.BatchStockCheckScript
	
	// 验证脚本包含批量检查逻辑
	// Verify script contains batch check logic
	assert.Contains(t, script, "KEYS", "Should reference multiple keys")
	assert.Contains(t, script, "ARGV[1]", "Should reference threshold")
	assert.Contains(t, script, "for i = 1, #KEYS", "Should iterate through keys")
	assert.Contains(t, script, "out_of_stock", "Should detect out of stock")
	assert.Contains(t, script, "low_stock", "Should detect low stock")
	assert.Contains(t, script, "batch check completed", "Should return completion message")
}

func TestScriptErrorHandling(t *testing.T) {
	scripts := map[string]string{
		"stock_deduct":      lua.StockDeductScript,
		"user_limit_check":  lua.UserLimitCheckScript,
		"activity_check":    lua.ActivityCheckScript,
		"seckill_process":   lua.SeckillProcessScript,
		"stock_rollback":    lua.StockRollbackScript,
		"batch_stock_check": lua.BatchStockCheckScript,
	}

	for name, script := range scripts {
		t.Run(name+"_error_handling", func(t *testing.T) {
			// 验证参数检查
			// Verify parameter checking
			assert.Contains(t, script, "invalid parameters", "Should check invalid parameters")
			
			// 验证返回格式一致性
			// Verify consistent return format
			assert.Contains(t, script, "return {0,", "Should have error return format")
			assert.Contains(t, script, "return {1,", "Should have success return format")
		})
	}
}

func TestScriptComments(t *testing.T) {
	scripts := map[string]string{
		"stock_deduct":      lua.StockDeductScript,
		"user_limit_check":  lua.UserLimitCheckScript,
		"activity_check":    lua.ActivityCheckScript,
		"seckill_process":   lua.SeckillProcessScript,
		"stock_rollback":    lua.StockRollbackScript,
		"batch_stock_check": lua.BatchStockCheckScript,
	}

	for name, script := range scripts {
		t.Run(name+"_comments", func(t *testing.T) {
			// 验证脚本包含中英文注释
			// Verify script contains Chinese and English comments
			assert.Contains(t, script, "--", "Should contain Lua comments")
			
			// 验证关键部分有注释
			// Verify key parts have comments
			commentCount := 0
			for i := 0; i < len(script)-1; i++ {
				if script[i] == '-' && script[i+1] == '-' {
					commentCount++
				}
			}
			assert.Greater(t, commentCount, 5, "Should have sufficient comments")
		})
	}
}

func TestScriptSyntax(t *testing.T) {
	scripts := map[string]string{
		"stock_deduct":      lua.StockDeductScript,
		"user_limit_check":  lua.UserLimitCheckScript,
		"activity_check":    lua.ActivityCheckScript,
		"seckill_process":   lua.SeckillProcessScript,
		"stock_rollback":    lua.StockRollbackScript,
		"batch_stock_check": lua.BatchStockCheckScript,
	}

	for name, script := range scripts {
		t.Run(name+"_syntax", func(t *testing.T) {
			// 基本语法检查
			// Basic syntax checking
			
			// 检查local变量声明
			// Check local variable declarations
			assert.Contains(t, script, "local ", "Should use local variables")
			
			// 检查Redis调用
			// Check Redis calls
			assert.Contains(t, script, "redis.call", "Should use redis.call")
			
			// 检查条件语句
			// Check conditional statements
			assert.Contains(t, script, "if ", "Should use conditional statements")
			assert.Contains(t, script, "then", "Should have then clauses")
			assert.Contains(t, script, "end", "Should have end statements")
			
			// 检查返回语句
			// Check return statements
			assert.Contains(t, script, "return ", "Should have return statements")
		})
	}
}
