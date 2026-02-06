# 监控系统实施报告 - M1 + M2 完成

## ✅ 已完成工作

### Milestone 1: 埋点基础设施

#### 1.1 添加 Prometheus 依赖
```bash
✅ prometheus/client_golang v1.23.2
✅ prometheus/promauto
✅ prometheus/promhttp
```

#### 1.2 创建 Metrics 包
**文件**: `internal/metrics/metrics.go`

**定义的指标**:

| 类别 | 指标名称 | 类型 | 说明 |
|------|---------|------|------|
| **记忆操作** | `memory_recall_duration_seconds` | Histogram | 混合召回耗时（按stage分类） |
| | `memory_recall_total` | Counter | 召回操作总数（按策略、状态分类） |
| | `memory_recall_results_count` | Histogram | 召回记忆数量（按层级分类） |
| | `memory_create_total` | Counter | 创建记忆总数（按层级、状态分类） |
| | `memory_create_duration_seconds` | Histogram | 创建耗时（按层级分类） |
| **LLM 调用** | `llm_requests_total` | Counter | LLM 调用次数 |
| | `llm_duration_seconds` | Histogram | LLM 响应耗时 |
| | `llm_tokens_used_total` | Counter | Token 消耗（预留，待完善） |
| | `llm_errors_total` | Counter | LLM 错误统计 |
| **Embedding** | `embedding_requests_total` | Counter | Embedding 调用次数 |
| | `embedding_duration_seconds` | Histogram | Embedding 耗时 |
| | `embedding_errors_total` | Counter | Embedding 错误（按类型分类） |
| | `embedding_throttle_wait_seconds` | Histogram | 限流等待时间 |
| **系统** | `goroutines_count` | Gauge | Goroutine 数量 |
| | `memory_usage_bytes` | Gauge | 内存使用量 |

#### 1.3 暴露 Metrics 端点
**修改文件**: `cmd/server/main.go`

```go
// 添加路由
router.GET("/metrics", gin.WrapH(promhttp.Handler()))

// 系统指标自动采集（每 15 秒）
go func() {
    ticker := time.NewTicker(15 * time.Second)
    for range ticker.C {
        metrics.GoroutinesCount.Set(float64(runtime.NumGoroutine()))
        metrics.MemoryUsageBytes.Set(float64(m.Alloc))
    }
}()
```

---

### Milestone 2: 核心业务埋点

#### 2.1 Manager 层埋点
**修改文件**: `internal/service/memory/manager.go`

**CreateMemory 埋点**:
```go
// 记录创建耗时
defer func() {
    metrics.MemoryCreateDuration.WithLabelValues(layer).Observe(duration)
}()

// 记录成功/失败
metrics.MemoryCreateTotal.WithLabelValues(layer, "success").Inc()
```

**HybridRecall 埋点**:
```go
// 记录召回耗时（按 stage）
defer func() {
    metrics.MemoryRecallDuration.WithLabelValues(stage).Observe(duration)
}()

// 记录召回结果数量（按 layer）
metrics.MemoryRecallResultsCount.WithLabelValues("profile").Observe(count)
metrics.MemoryRecallResultsCount.WithLabelValues("topic").Observe(count)
metrics.MemoryRecallResultsCount.WithLabelValues("dialogue").Observe(count)

// 记录成功
metrics.MemoryRecallTotal.WithLabelValues(strategy, "success").Inc()
```

#### 2.2 LLM 层埋点
**修改文件**: 
- `internal/llm/openai.go`
- `internal/llm/gemini.go`

**GenerateText 埋点**:
```go
// 记录耗时和调用次数
defer func() {
    metrics.LLMDuration.WithLabelValues(provider, model).Observe(duration)
    metrics.LLMRequestsTotal.WithLabelValues(provider, model, "chat").Inc()
}()

// 错误分类（timeout/rate_limit/server_error）
metrics.LLMErrorsTotal.WithLabelValues(provider, errorType).Inc()
```

#### 2.3 Embedding 层埋点
**修改文件**: `internal/adapter/eino.go`

**Embed 埋点**:
```go
// 记录总调用次数和耗时
defer func() {
    metrics.EmbeddingDuration.Observe(duration)
    metrics.EmbeddingRequestsTotal.Inc()
}()

// 记录限流等待时间
metrics.EmbeddingThrottleWaitSeconds.Observe(waitTime)

// 错误分类（throttled/timeout/invalid_response）
metrics.EmbeddingErrorsTotal.WithLabelValues(errorType).Inc()
```

---

## 🧪 测试验证

### 启动服务
```bash
cd d:\file\MemoryOs
.\bin\server.exe
```

### 访问 Metrics 端点
```bash
curl http://localhost:8080/metrics
```

**预期输出示例**:
```prometheus
# HELP memory_create_total Total number of memory creation operations
# TYPE memory_create_total counter
memory_create_total{layer="dialogue",status="success"} 15

# HELP memory_recall_duration_seconds Duration of hybrid memory recall operations
# TYPE memory_recall_duration_seconds histogram
memory_recall_duration_seconds_bucket{stage="multi_turn",le="0.5"} 8
memory_recall_duration_seconds_bucket{stage="multi_turn",le="1"} 12
memory_recall_duration_seconds_sum{stage="multi_turn"} 8.7
memory_recall_duration_seconds_count{stage="multi_turn"} 12

# HELP llm_requests_total Total number of LLM API requests
# TYPE llm_requests_total counter
llm_requests_total{model="gpt-4o-mini",operation="chat",provider="openai"} 5

# HELP embedding_requests_total Total number of embedding API requests
# TYPE embedding_requests_total counter
embedding_requests_total 23

# HELP goroutines_count Current number of goroutines
# TYPE goroutines_count gauge
goroutines_count 47
```

---

## 📊 下一步：Prometheus + Grafana 集成

### 待完成（M3 + M4）:

1. **Docker Compose 配置** (Milestone 3)
   - [ ] 添加 Prometheus 容器
   - [ ] 添加 Grafana 容器
   - [ ] 配置数据源连接

2. **Grafana Dashboard** (Milestone 4)
   - [ ] 创建记忆操作总览 Dashboard
   - [ ] 创建 LLM 成本监控 Dashboard
   - [ ] 创建性能指标 Dashboard

### 临时可用：命令行查询
```bash
# 查看所有指标
curl http://localhost:8080/metrics | grep -E "(memory|llm|embedding)_"

# 查看召回耗时
curl http://localhost:8080/metrics | grep "memory_recall_duration"

# 查看 LLM 调用统计
curl http://localhost:8080/metrics | grep "llm_requests_total"
```

---

## 💡 使用建议

### 监控关键指标

1. **召回性能**
   - P95 召回耗时 < 1s → 用户体验良好
   - 召回数量分布 → 检查是否过多/过少

2. **LLM 成本**
   - 每日请求数 × 平均 Token → 估算月度成本
   - 错误率 < 1% → API 稳定性良好

3. **Embedding 限流**
   - `embedding_throttle_wait_seconds` → 检查 1s 间隔是否足够
   - `embedding_errors_total{error_type="throttled"}` → 监控 403 错误

4. **系统健康**
   - `goroutines_count` 持续增长 → Goroutine 泄漏
   - `memory_usage_bytes` 超过 1GB → 可能内存泄漏

---

## 🎯 架构设计亮点

### 1. 分层埋点策略
- **业务层（Manager）**: 记忆操作的端到端性能
- **服务层（LLM/Embedding）**: API 调用的细粒度监控
- **系统层（Runtime）**: Go 进程的健康状况

### 2. 错误分类
- 不同错误类型分开统计（timeout/rate_limit/throttled）
- 便于定位问题根因

### 3. 性能桶设计
- Histogram 桶根据实际场景定制：
  - 召回：0.005s ~ 10s（覆盖快速和慢速场景）
  - LLM：0.5s ~ 30s（LLM 响应通常较慢）
  - Embedding：0.1s ~ 5s（Embedding 较快）

### 4. 资源开销小
- Prometheus 指标采集是异步的
- defer 延迟记录不阻塞主逻辑
- 系统指标每 15 秒采集一次，避免高频率

---

## 🔍 故障排查

### 如果编译失败
```bash
go mod tidy
go build -o bin/server.exe ./cmd/server
```

### 如果 /metrics 返回 404
- 检查 `router.GET("/metrics", ...)` 是否正确注册
- 确认服务器启动日志中有 `🚀 MemoryOS 服务已启动`

### 如果没有数据
- 确保调用了相关 API（如 POST /api/v1/recall/hybrid）
- 检查 `metrics.XXX.WithLabelValues(...).Inc()` 是否执行

---

**实施完成时间**: 2026-02-06  
**下一步**: 根据你的选择启动 M3 + M4（Prometheus + Grafana Docker 集成）
