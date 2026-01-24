# MemoryOS API 接口测试脚本

$baseUrl = "http://localhost:8080"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "MemoryOS API 测试" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# 生成测试 ID
$timestamp = Get-Date -Format "HHmmss"
$userId = "test_user_$timestamp"
$sessionId = "test_session_$timestamp"

# 测试 1: 创建记忆
Write-Host "[测试 1] 创建对话记忆..." -ForegroundColor Yellow
$createBody = @{
    user_id = $userId
    session_id = $sessionId
    layer = "dialogue"
    type = "interaction"
    content = "用户问：什么是 RAG？助手答：RAG 是检索增强生成技术。"
    metadata = @{
        dialogue_turn = 1
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/memory" -Method Post -Headers $headers -Body $createBody
    Write-Host "✅ 创建成功" -ForegroundColor Green
    Write-Host "   Memory ID: $($response.data.id)" -ForegroundColor Gray
    $memoryId = $response.data.id
} catch {
    Write-Host "❌ 创建失败: $_" -ForegroundColor Red
    exit 1
}

Start-Sleep -Seconds 1

# 测试 2: 召回对话记忆
Write-Host "`n[测试 2] 召回对话记忆..." -ForegroundColor Yellow
$recallBody = @{
    user_id = $userId
    session_id = $sessionId
    query = "什么是 RAG"
    limit = 5
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/recall/dialogue" -Method Post -Headers $headers -Body $recallBody
    Write-Host "✅ 召回成功，找到 $($response.data.Count) 条记忆" -ForegroundColor Green
    foreach ($item in $response.data) {
        Write-Host "   相似度: $($item.score.ToString('0.0000')) | $($item.content.Substring(0, [Math]::Min(40, $item.content.Length)))..." -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ 召回失败: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试 3: 混合召回
Write-Host "`n[测试 3] 混合召回..." -ForegroundColor Yellow
$hybridBody = @{
    user_id = $userId
    query = "RAG 技术"
    limit = 5
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/recall/hybrid" -Method Post -Headers $headers -Body $hybridBody
    Write-Host "✅ 混合召回成功" -ForegroundColor Green
    Write-Host "   对话层: $($response.data.dialogues.Count) 条" -ForegroundColor Gray
    Write-Host "   话题层: $($response.data.topics.Count) 条" -ForegroundColor Gray
    Write-Host "   画像层: $($response.data.profiles.Count) 条" -ForegroundColor Gray
} catch {
    Write-Host "❌ 混合召回失败: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试 4: 更新记忆
Write-Host "`n[测试 4] 更新记忆..." -ForegroundColor Yellow
$updateBody = @{
    metadata = @{
        dialogue_turn = 1
        updated = $true
        update_time = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/memory/$memoryId" -Method Put -Headers $headers -Body $updateBody
    Write-Host "✅ 更新成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 更新失败: $_" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试 5: 删除记忆
Write-Host "`n[测试 5] 删除记忆..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/memory/$memoryId" -Method Delete -Headers $headers
    Write-Host "✅ 删除成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 删除失败: $_" -ForegroundColor Red
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "✅ API 测试完成" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
