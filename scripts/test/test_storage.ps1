#!/usr/bin/env pwsh
# 测试真实存储层功能

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试 MemoryOS 真实存储层" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"

# 1. 健康检查
Write-Host "[1/5] 健康检查..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get
    Write-Host "✅ 服务状态: $($health.status)" -ForegroundColor Green
    Write-Host "   模式: $($health.mode)" -ForegroundColor Gray
    Write-Host "   数据库: $($health.db)" -ForegroundColor Gray
    Write-Host "   Redis: $($health.redis)" -ForegroundColor Gray
} catch {
    Write-Host "❌ 健康检查失败: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 2. 创建对话记忆
Write-Host "[2/5] 创建对话记忆..." -ForegroundColor Yellow
$createPayload = @{
    user_id = "test_user_$(Get-Date -Format 'HHmmss')"
    layer = "dialogue"
    type = "user_message"
    content = "你好，我正在测试真实存储层功能"
    metadata = @{
        session_id = "test_session_001"
        role = "user"
        turn_number = 1
    }
} | ConvertTo-Json

try {
    $createResult = Invoke-RestMethod -Uri "$baseUrl/api/v1/memories" -Method Post -Body $createPayload -ContentType "application/json"
    $memoryId = $createResult.id
    Write-Host "✅ 创建成功! ID: $memoryId" -ForegroundColor Green
} catch {
    Write-Host "❌ 创建失败: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 3. 获取刚创建的记忆
Write-Host "[3/5] 获取记忆..." -ForegroundColor Yellow
try {
    $memory = Invoke-RestMethod -Uri "$baseUrl/api/v1/memories/$memoryId" -Method Get
    Write-Host "✅ 获取成功!" -ForegroundColor Green
    Write-Host "   内容: $($memory.content)" -ForegroundColor Gray
    Write-Host "   层级: $($memory.layer)" -ForegroundColor Gray
    Write-Host "   类型: $($memory.type)" -ForegroundColor Gray
} catch {
    Write-Host "❌ 获取失败: $_" -ForegroundColor Red
}
Write-Host ""

# 4. 查询会话对话
Write-Host "[4/5] 查询会话对话..." -ForegroundColor Yellow
$recallPayload = @{
    user_id = $createPayload | ConvertFrom-Json | Select-Object -ExpandProperty user_id
    session_id = "test_session_001"
    limit = 10
} | ConvertTo-Json

try {
    $dialogues = Invoke-RestMethod -Uri "$baseUrl/api/v1/recall/dialogue" -Method Post -Body $recallPayload -ContentType "application/json"
    Write-Host "✅ 查询成功! 找到 $($dialogues.memories.Count) 条对话" -ForegroundColor Green
    foreach ($m in $dialogues.memories) {
        Write-Host "   - [$($m.type)] $($m.content)" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ 查询失败: $_" -ForegroundColor Red
}
Write-Host ""

# 5. 验证数据库持久化
Write-Host "[5/5] 验证数据库持久化..." -ForegroundColor Yellow
try {
    $dbCount = docker exec memoryos-postgres psql -U memoryos -d memoryos -t -c "SELECT COUNT(*) FROM dialogue_memory;"
    Write-Host "✅ 数据库中共有 $($dbCount.Trim()) 条对话记忆" -ForegroundColor Green
} catch {
    Write-Host "⚠️  无法查询数据库" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ 所有测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
