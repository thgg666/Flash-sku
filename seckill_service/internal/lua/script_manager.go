package lua

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sync"

	"github.com/flashsku/seckill/pkg/logger"
	"github.com/flashsku/seckill/pkg/redis"
)

// ScriptManager Lua脚本管理器
// ScriptManager Lua script manager
type ScriptManager struct {
	redisClient redis.Client
	logger      logger.Logger
	scripts     map[string]*Script
	mu          sync.RWMutex
}

// Script Lua脚本
// Script Lua script
type Script struct {
	Name    string
	Content string
	SHA1    string
	Loaded  bool
}

// ScriptResult 脚本执行结果
// ScriptResult script execution result
type ScriptResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
}

// NewScriptManager 创建脚本管理器
// NewScriptManager creates script manager
func NewScriptManager(redisClient redis.Client, log logger.Logger) *ScriptManager {
	manager := &ScriptManager{
		redisClient: redisClient,
		logger:      log,
		scripts:     make(map[string]*Script),
	}

	// 注册内置脚本
	// Register built-in scripts
	manager.registerBuiltinScripts()

	return manager
}

// registerBuiltinScripts 注册内置脚本
// registerBuiltinScripts registers built-in scripts
func (sm *ScriptManager) registerBuiltinScripts() {
	// 库存扣减脚本
	// Stock deduction script
	sm.RegisterScript("stock_deduct", StockDeductScript)

	// 用户限购检查脚本
	// User purchase limit check script
	sm.RegisterScript("user_limit_check", UserLimitCheckScript)

	// 活动状态检查脚本
	// Activity status check script
	sm.RegisterScript("activity_check", ActivityCheckScript)

	// 秒杀完整流程脚本
	// Complete seckill process script
	sm.RegisterScript("seckill_process", SeckillProcessScript)

	// 库存回滚脚本
	// Stock rollback script
	sm.RegisterScript("stock_rollback", StockRollbackScript)

	// 批量库存检查脚本
	// Batch stock check script
	sm.RegisterScript("batch_stock_check", BatchStockCheckScript)
}

// RegisterScript 注册脚本
// RegisterScript registers script
func (sm *ScriptManager) RegisterScript(name, content string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 计算SHA1
	// Calculate SHA1
	hash := sha1.Sum([]byte(content))
	sha1Str := fmt.Sprintf("%x", hash)

	script := &Script{
		Name:    name,
		Content: content,
		SHA1:    sha1Str,
		Loaded:  false,
	}

	sm.scripts[name] = script
	sm.logger.Debug("Script registered",
		logger.String("name", name),
		logger.String("sha1", sha1Str))
}

// LoadScript 加载脚本到Redis
// LoadScript loads script to Redis
func (sm *ScriptManager) LoadScript(ctx context.Context, name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	script, exists := sm.scripts[name]
	if !exists {
		return fmt.Errorf("script %s not found", name)
	}

	if script.Loaded {
		return nil
	}

	// 使用SCRIPT LOAD命令加载脚本
	// Use SCRIPT LOAD command to load script
	sha1, err := sm.redisClient.ScriptLoad(ctx, script.Content)
	if err != nil {
		return fmt.Errorf("failed to load script %s: %w", name, err)
	}

	// 验证SHA1
	// Verify SHA1
	if sha1 != script.SHA1 {
		sm.logger.Warn("Script SHA1 mismatch",
			logger.String("name", name),
			logger.String("expected", script.SHA1),
			logger.String("actual", sha1))
		script.SHA1 = sha1
	}

	script.Loaded = true
	sm.logger.Info("Script loaded successfully",
		logger.String("name", name),
		logger.String("sha1", sha1))

	return nil
}

// LoadAllScripts 加载所有脚本
// LoadAllScripts loads all scripts
func (sm *ScriptManager) LoadAllScripts(ctx context.Context) error {
	sm.mu.RLock()
	scriptNames := make([]string, 0, len(sm.scripts))
	for name := range sm.scripts {
		scriptNames = append(scriptNames, name)
	}
	sm.mu.RUnlock()

	for _, name := range scriptNames {
		if err := sm.LoadScript(ctx, name); err != nil {
			sm.logger.Error("Failed to load script",
				logger.String("name", name),
				logger.Error(err))
			return err
		}
	}

	sm.logger.Info("All scripts loaded successfully",
		logger.Int("count", len(scriptNames)))
	return nil
}

// ExecuteScript 执行脚本
// ExecuteScript executes script
func (sm *ScriptManager) ExecuteScript(ctx context.Context, name string, keys []string, args []interface{}) (*ScriptResult, error) {
	sm.mu.RLock()
	script, exists := sm.scripts[name]
	sm.mu.RUnlock()

	if !exists {
		return &ScriptResult{
			Success: false,
			Error:   fmt.Sprintf("script %s not found", name),
		}, fmt.Errorf("script %s not found", name)
	}

	// 确保脚本已加载
	// Ensure script is loaded
	if !script.Loaded {
		if err := sm.LoadScript(ctx, name); err != nil {
			return &ScriptResult{
				Success: false,
				Error:   fmt.Sprintf("failed to load script: %v", err),
			}, err
		}
	}

	// 执行脚本
	// Execute script
	result, err := sm.redisClient.EvalSHA(ctx, script.SHA1, keys, args...)
	if err != nil {
		// 如果脚本不存在，尝试重新加载
		// If script doesn't exist, try to reload
		if isScriptNotFoundError(err) {
			sm.logger.Warn("Script not found in Redis, reloading",
				logger.String("name", name))

			script.Loaded = false
			if loadErr := sm.LoadScript(ctx, name); loadErr != nil {
				return &ScriptResult{
					Success: false,
					Error:   fmt.Sprintf("failed to reload script: %v", loadErr),
				}, loadErr
			}

			// 重新执行
			// Re-execute
			result, err = sm.redisClient.EvalSHA(ctx, script.SHA1, keys, args...)
		}

		if err != nil {
			return &ScriptResult{
				Success: false,
				Error:   err.Error(),
			}, err
		}
	}

	return &ScriptResult{
		Success: true,
		Result:  result,
	}, nil
}

// GetScriptInfo 获取脚本信息
// GetScriptInfo gets script information
func (sm *ScriptManager) GetScriptInfo(name string) (*Script, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	script, exists := sm.scripts[name]
	if !exists {
		return nil, fmt.Errorf("script %s not found", name)
	}

	// 返回副本
	// Return copy
	return &Script{
		Name:    script.Name,
		Content: script.Content,
		SHA1:    script.SHA1,
		Loaded:  script.Loaded,
	}, nil
}

// ListScripts 列出所有脚本
// ListScripts lists all scripts
func (sm *ScriptManager) ListScripts() map[string]*Script {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*Script)
	for name, script := range sm.scripts {
		result[name] = &Script{
			Name:    script.Name,
			Content: script.Content,
			SHA1:    script.SHA1,
			Loaded:  script.Loaded,
		}
	}

	return result
}

// ReloadScript 重新加载脚本
// ReloadScript reloads script
func (sm *ScriptManager) ReloadScript(ctx context.Context, name string) error {
	sm.mu.Lock()
	script, exists := sm.scripts[name]
	if !exists {
		sm.mu.Unlock()
		return fmt.Errorf("script %s not found", name)
	}

	script.Loaded = false
	sm.mu.Unlock()

	return sm.LoadScript(ctx, name)
}

// CheckScriptExists 检查脚本是否存在于Redis中
// CheckScriptExists checks if script exists in Redis
func (sm *ScriptManager) CheckScriptExists(ctx context.Context, name string) (bool, error) {
	sm.mu.RLock()
	script, exists := sm.scripts[name]
	sm.mu.RUnlock()

	if !exists {
		return false, fmt.Errorf("script %s not found", name)
	}

	// 使用SCRIPT EXISTS命令检查
	// Use SCRIPT EXISTS command to check
	result, err := sm.redisClient.ScriptExists(ctx, script.SHA1)
	if err != nil {
		return false, err
	}

	return result, nil
}

// isScriptNotFoundError 检查是否是脚本未找到错误
// isScriptNotFoundError checks if it's a script not found error
func isScriptNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	return contains(errStr, "NOSCRIPT") || contains(errStr, "script not found")
}

// contains 检查字符串是否包含子字符串
// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      indexOf(s, substr) >= 0)))
}

// indexOf 查找子字符串位置
// indexOf finds substring position
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
