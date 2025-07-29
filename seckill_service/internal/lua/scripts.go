package lua

// StockDeductScript 库存扣减脚本
// StockDeductScript stock deduction script
const StockDeductScript = `
-- 库存扣减脚本
-- Stock deduction script
-- KEYS[1]: 库存键 stock key
-- ARGV[1]: 扣减数量 deduction amount
-- ARGV[2]: 最小库存 minimum stock (可选，默认0)

local stock_key = KEYS[1]
local deduct_amount = tonumber(ARGV[1])
local min_stock = tonumber(ARGV[2]) or 0

-- 检查参数
-- Check parameters
if not stock_key or not deduct_amount or deduct_amount <= 0 then
    return {0, "invalid parameters"}
end

-- 获取当前库存
-- Get current stock
local current_stock = redis.call('GET', stock_key)
if not current_stock then
    return {0, "stock not found"}
end

current_stock = tonumber(current_stock)
if not current_stock then
    return {0, "invalid stock value"}
end

-- 检查库存是否足够
-- Check if stock is sufficient
if current_stock < deduct_amount then
    return {0, "insufficient stock", current_stock}
end

-- 检查扣减后是否低于最小库存
-- Check if stock after deduction is below minimum
local new_stock = current_stock - deduct_amount
if new_stock < min_stock then
    return {0, "below minimum stock", current_stock}
end

-- 原子性扣减库存
-- Atomically deduct stock
redis.call('DECRBY', stock_key, deduct_amount)

-- 返回成功结果
-- Return success result
return {1, "success", new_stock, current_stock}
`

// UserLimitCheckScript 用户限购检查脚本
// UserLimitCheckScript user purchase limit check script
const UserLimitCheckScript = `
-- 用户限购检查脚本
-- User purchase limit check script
-- KEYS[1]: 用户限购键 user limit key
-- ARGV[1]: 购买数量 purchase amount
-- ARGV[2]: 最大限购数量 max purchase limit
-- ARGV[3]: TTL过期时间(秒) TTL expiration time in seconds

local limit_key = KEYS[1]
local purchase_amount = tonumber(ARGV[1])
local max_limit = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3]) or 86400

-- 检查参数
-- Check parameters
if not limit_key or not purchase_amount or not max_limit or 
   purchase_amount <= 0 or max_limit <= 0 then
    return {0, "invalid parameters"}
end

-- 获取当前已购买数量
-- Get current purchased amount
local current_purchased = redis.call('GET', limit_key)
current_purchased = tonumber(current_purchased) or 0

-- 检查是否超过限购
-- Check if exceeds purchase limit
local total_after_purchase = current_purchased + purchase_amount
if total_after_purchase > max_limit then
    return {0, "exceeds purchase limit", current_purchased, max_limit}
end

-- 增加购买计数
-- Increment purchase count
local new_count = redis.call('INCRBY', limit_key, purchase_amount)

-- 设置过期时间（如果是新键）
-- Set expiration time (if new key)
if current_purchased == 0 then
    redis.call('EXPIRE', limit_key, ttl)
end

-- 返回成功结果
-- Return success result
return {1, "success", new_count, max_limit - new_count}
`

// ActivityCheckScript 活动状态检查脚本
// ActivityCheckScript activity status check script
const ActivityCheckScript = `
-- 活动状态检查脚本
-- Activity status check script
-- KEYS[1]: 活动信息键 activity info key
-- KEYS[2]: 活动状态键 activity status key
-- ARGV[1]: 当前时间戳 current timestamp

local activity_key = KEYS[1]
local status_key = KEYS[2]
local current_time = tonumber(ARGV[1])

-- 检查参数
-- Check parameters
if not activity_key or not status_key or not current_time then
    return {0, "invalid parameters"}
end

-- 获取活动信息
-- Get activity information
local activity_info = redis.call('GET', activity_key)
if not activity_info then
    return {0, "activity not found"}
end

-- 解析活动信息（假设是JSON格式）
-- Parse activity info (assuming JSON format)
-- 这里简化处理，实际应该解析JSON
-- Simplified here, should parse JSON in practice
local activity_status = redis.call('GET', status_key)
if not activity_status then
    return {0, "activity status not found"}
end

-- 检查活动状态
-- Check activity status
if activity_status ~= "active" then
    return {0, "activity not active", activity_status}
end

-- 返回成功结果
-- Return success result
return {1, "activity is active", activity_status}
`

// SeckillProcessScript 秒杀完整流程脚本
// SeckillProcessScript complete seckill process script
const SeckillProcessScript = `
-- 秒杀完整流程脚本
-- Complete seckill process script
-- KEYS[1]: 活动信息键 activity info key
-- KEYS[2]: 库存键 stock key
-- KEYS[3]: 用户限购键 user limit key
-- KEYS[4]: 活动状态键 activity status key
-- ARGV[1]: 购买数量 purchase amount
-- ARGV[2]: 用户最大限购 user max limit
-- ARGV[3]: 当前时间戳 current timestamp
-- ARGV[4]: 限购TTL user limit TTL

local activity_key = KEYS[1]
local stock_key = KEYS[2]
local user_limit_key = KEYS[3]
local status_key = KEYS[4]

local purchase_amount = tonumber(ARGV[1])
local max_user_limit = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])
local limit_ttl = tonumber(ARGV[4]) or 86400

-- 检查参数
-- Check parameters
if not purchase_amount or purchase_amount <= 0 or not max_user_limit then
    return {0, "invalid parameters"}
end

-- 1. 检查活动状态
-- 1. Check activity status
local activity_status = redis.call('GET', status_key)
if not activity_status or activity_status ~= "active" then
    return {0, "activity not active", activity_status or "unknown"}
end

-- 2. 检查用户限购
-- 2. Check user purchase limit
local current_purchased = redis.call('GET', user_limit_key)
current_purchased = tonumber(current_purchased) or 0

if current_purchased + purchase_amount > max_user_limit then
    return {0, "exceeds user limit", current_purchased, max_user_limit}
end

-- 3. 检查库存并扣减
-- 3. Check stock and deduct
local current_stock = redis.call('GET', stock_key)
if not current_stock then
    return {0, "stock not found"}
end

current_stock = tonumber(current_stock)
if current_stock < purchase_amount then
    return {0, "insufficient stock", current_stock}
end

-- 4. 原子性执行所有操作
-- 4. Atomically execute all operations

-- 扣减库存
-- Deduct stock
local new_stock = redis.call('DECRBY', stock_key, purchase_amount)

-- 增加用户购买计数
-- Increment user purchase count
local new_user_count = redis.call('INCRBY', user_limit_key, purchase_amount)

-- 设置用户限购过期时间（如果是新键）
-- Set user limit expiration (if new key)
if current_purchased == 0 then
    redis.call('EXPIRE', user_limit_key, limit_ttl)
end

-- 5. 返回成功结果
-- 5. Return success result
return {1, "seckill success", {
    new_stock = new_stock,
    user_purchased = new_user_count,
    remaining_limit = max_user_limit - new_user_count,
    timestamp = current_time
}}
`

// StockRollbackScript 库存回滚脚本
// StockRollbackScript stock rollback script
const StockRollbackScript = `
-- 库存回滚脚本
-- Stock rollback script
-- KEYS[1]: 库存键 stock key
-- KEYS[2]: 用户限购键 user limit key (可选)
-- ARGV[1]: 回滚数量 rollback amount
-- ARGV[2]: 最大库存限制 max stock limit (可选)

local stock_key = KEYS[1]
local user_limit_key = KEYS[2]
local rollback_amount = tonumber(ARGV[1])
local max_stock = tonumber(ARGV[2])

-- 检查参数
-- Check parameters
if not stock_key or not rollback_amount or rollback_amount <= 0 then
    return {0, "invalid parameters"}
end

-- 获取当前库存
-- Get current stock
local current_stock = redis.call('GET', stock_key)
if not current_stock then
    return {0, "stock not found"}
end

current_stock = tonumber(current_stock)

-- 检查是否超过最大库存限制
-- Check if exceeds max stock limit
if max_stock and (current_stock + rollback_amount) > max_stock then
    return {0, "exceeds max stock limit", current_stock, max_stock}
end

-- 回滚库存
-- Rollback stock
local new_stock = redis.call('INCRBY', stock_key, rollback_amount)

-- 如果提供了用户限购键，也回滚用户购买计数
-- If user limit key is provided, also rollback user purchase count
local new_user_count = nil
if user_limit_key and user_limit_key ~= "" then
    local current_user_count = redis.call('GET', user_limit_key)
    if current_user_count and tonumber(current_user_count) >= rollback_amount then
        new_user_count = redis.call('DECRBY', user_limit_key, rollback_amount)
    end
end

-- 返回成功结果
-- Return success result
return {1, "rollback success", {
    new_stock = new_stock,
    rollback_amount = rollback_amount,
    user_count = new_user_count
}}
`

// BatchStockCheckScript 批量库存检查脚本
// BatchStockCheckScript batch stock check script
const BatchStockCheckScript = `
-- 批量库存检查脚本
-- Batch stock check script
-- KEYS: 多个库存键 multiple stock keys
-- ARGV[1]: 检查的最小库存阈值 minimum stock threshold

local min_threshold = tonumber(ARGV[1]) or 0

-- 检查参数
-- Check parameters
if #KEYS == 0 then
    return {0, "invalid parameters: no keys provided"}
end

local results = {}

-- 遍历所有库存键
-- Iterate through all stock keys
for i = 1, #KEYS do
    local stock_key = KEYS[i]

    -- 检查键是否有效
    -- Check if key is valid
    if not stock_key or stock_key == "" then
        return {0, "invalid parameters: empty key"}
    end

    local current_stock = redis.call('GET', stock_key)

    if current_stock then
        current_stock = tonumber(current_stock)
        local status = "normal"

        if current_stock == 0 then
            status = "out_of_stock"
        elseif current_stock <= min_threshold then
            status = "low_stock"
        end

        results[i] = {
            key = stock_key,
            stock = current_stock,
            status = status
        }
    else
        results[i] = {
            key = stock_key,
            stock = -1,
            status = "not_found"
        }
    end
end

-- 返回批量检查结果
-- Return batch check results
return {1, "batch check completed", results}
`
