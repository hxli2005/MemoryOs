# MemoryOS 监控系统部署指南

## 架构概览

```
MemoryOS API (localhost:8080)
      ↓ /metrics
Prometheus (localhost:9090) → 数据采集 + 告警评估
      ↓ PromQL 查询
Grafana (localhost:3000) → 可视化 Dashboard
```

## 快速启动

### 1. 前置条件
- Docker Desktop 已安装并运行
- MemoryOS API 服务运行在 `localhost:8080`

### 2. 启动监控栈

```powershell
# 启动 Prometheus + Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# 查看日志
docker-compose -f docker-compose.monitoring.yml logs -f

# 检查服务状态
docker-compose -f docker-compose.monitoring.yml ps
```

### 3. 访问界面

| 服务 | URL | 凭据 |
|------|-----|------|
| Grafana | http://localhost:3000 | admin / memoryos123 |
| Prometheus | http://localhost:9090 | 无需认证 |
| MemoryOS Metrics | http://localhost:8080/metrics | 无需认证 |

### 4. 验证数据采集

1. 访问 Prometheus UI: http://localhost:9090/targets
   - 确认 `memoryos-api` 目标状态为 **UP**
   
2. 执行测试查询:
   ```promql
   # 召回延迟 P95
   histogram_quantile(0.95, rate(memory_recall_duration_seconds_bucket[5m]))
   
   # 记忆创建成功率
   sum(rate(memory_create_total{status="success"}[5m])) / sum(rate(memory_create_total[5m]))
   ```

3. 登录 Grafana 查看 Dashboard:
   - 导入已配置的 "MemoryOS 核心业务监控" 面板
   - 验证所有图表显示数据

## 告警配置

告警规则位于 `monitoring/alerts.yml`，覆盖：

### P0 告警 (立即处理)
- **HighRecallLatency**: 召回延迟 P95 > 2s
- **HighMemoryCreationFailureRate**: 创建失败率 > 5%
- **MemoryOSDown**: 服务不可用

### P1 告警 (性能降级)
- **HighLLMLatency**: LLM 延迟 P95 > 5s
- **HighEmbeddingErrorRate**: Embedding 错误率 > 10%
- **HighEmbeddingThrottleTime**: 限流等待过长

### P2 告警 (系统健康)
- **GoroutineLeakSuspected**: Goroutine 数 > 100
- **HighMemoryUsage**: 内存占用 > 500 MB

查看告警状态: http://localhost:9090/alerts

## Dashboard 说明

### 核心面板 (memoryos-overview.json)

1. **混合召回延迟分布**: P50/P95/P99 延迟趋势，目标 < 2s
2. **记忆创建成功率**: 实时成功率，目标 > 99%
3. **Embedding 生成 QPS**: 每秒请求数，监控限流情况
4. **召回策略分布**: 按对话阶段统计使用频率
5. **LLM 调用次数**: 按 Provider (OpenAI/Gemini) 分类
6. **系统资源**: Goroutines 和内存占用趋势
7. **Embedding 限流等待**: 监控限流等待时间
8. **错误率趋势**: Embedding & LLM 错误分类统计
9. **召回结果数量分布**: 热力图展示返回记忆数量分布

## 配置调整

### 修改 Prometheus 抓取间隔

编辑 `monitoring/prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'memoryos-api'
    scrape_interval: 5s  # 改为 5 秒
```

重载配置:
```powershell
curl -X POST http://localhost:9090/-/reload
```

### 修改数据保留时长

编辑 `deploy/compose/docker-compose.monitoring.yml`:

```yaml
command:
  - '--storage.tsdb.retention.time=30d'  # 保留 30 天
```

### 配置告警通知 (可选)

取消注释 AlertManager 服务，创建 `monitoring/alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m

route:
  receiver: 'webhook'

receivers:
  - name: 'webhook'
    webhook_configs:
      - url: 'http://your-webhook-url'
```

## 故障排查

### Prometheus 无法抓取 MemoryOS 指标

1. 检查 MemoryOS 服务是否运行:
   ```powershell
   curl http://localhost:8080/health
   ```

2. 检查 Prometheus 配置中的 target 地址:
   - Windows/Mac: `host.docker.internal:8080`
   - Linux: `172.17.0.1:8080` 或宿主机 IP

3. 查看 Prometheus 日志:
   ```powershell
   docker logs memoryos-prometheus
   ```

### Grafana Dashboard 无数据

1. 检查 Datasource 配置: Settings → Data Sources → Prometheus
2. 测试连接: "Save & Test" 应显示绿色成功提示
3. 检查 Prometheus 是否有数据:
   ```promql
   up{job="memoryos-api"}
   ```

### 告警未触发

1. 检查告警规则语法: http://localhost:9090/rules
2. 查看告警评估状态: http://localhost:9090/alerts
3. 确认 `evaluation_interval` 配置合理 (默认 15s)

## 停止和清理

```powershell
# 停止服务 (保留数据)
docker-compose -f deploy/compose/docker-compose.monitoring.yml stop

# 停止并删除容器 (保留数据卷)
docker-compose -f deploy/compose/docker-compose.monitoring.yml down

# 完全清理 (删除数据卷)
docker-compose -f deploy/compose/docker-compose.monitoring.yml down -v
```

## 性能优化建议

1. **Prometheus 数据量过大**:
   - 增加抓取间隔 (`scrape_interval`)
   - 减少保留时长 (`retention.time`)
   - 启用远程存储 (Thanos/Cortex)

2. **Grafana 查询缓慢**:
   - 减少 Dashboard 刷新频率 (默认 10s)
   - 优化 PromQL 查询 (使用 recording rules)
   - 限制时间范围 (避免查询超过 24h 数据)

3. **告警风暴**:
   - 调整告警阈值和 `for` 持续时间
   - 使用 AlertManager 的 grouping 和 inhibition
   - 设置告警静默期 (Silences)
