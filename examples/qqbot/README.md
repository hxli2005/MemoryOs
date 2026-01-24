# QQ Bot - MemoryOS 生产级应用示例

> **基于 NapCat 的 QQ 聊天机器人，完整实现 Persona 驱动的长期记忆对话**

## ✨ 特性

- ✅ **完整 NapCat 集成**：WebSocket 消息收发、好感度系统、私聊支持
- ✅ **Persona 配置化**：YAML 定义人设，支持热切换多个角色
- ✅ **长期记忆召回**：三段式记忆（对话/主题/画像）混合检索
- ✅ **Docker 部署**：一键启动完整技术栈（PostgreSQL + Redis + Milvus）
- ✅ **优雅降级**：LLM 失败时自动回退，消息队列过载保护

## 🚀 快速启动

### 方式一：Docker 部署（推荐）

```powershell
# 1. 配置环境变量（编辑 .env 文件）
LLM_API_KEY=your-api-key-here

# 2. 启动数据库服务
docker-compose up -d postgres redis milvus

# 3. 启动 QQ Bot 容器
docker-compose -f docker-compose.qqbot.yaml up -d

# 4. 查看日志
docker logs -f memoryos-qqbot
```

### 方式二：本地开发

```bash
# 1. 安装依赖
go mod download

# 2. 配置文件
cp config/config.docker.yaml config/config.yaml
# 编辑 config.yaml，填写 LLM API Key

# 3. 启动 Bot
go run examples/qqbot/main.go
```

## 🔌 接入 NapCat

### 1. 安装 NapCat

参考官方文档：[NapCat Setup Guide](NAPCAT_SETUP.md)

```powershell
# Docker 方式（推荐）
docker run -d --name napcat \
  -p 6099:6099 -p 6700:3000 \
  -e ACCOUNT=你的QQ号 \
  mlikiowa/napcat-docker:latest
```

### 2. 配置 WebSocket URL

```yaml
# config/config.yaml 或环境变量
CQHTTP_WS_URL=ws://host.docker.internal:6700  # Docker 环境
# 或
CQHTTP_WS_URL=ws://localhost:6700             # 本地开发
```

### 3. 验证连接

启动后看到以下日志即成功：
```
✅ 成功连接到 go-cqhttp: ws://host.docker.internal:6700
🤖 QQ Bot 已启动，等待消息...
```

## 🎭 Persona 配置

### 当前可用人设

| 文件 | 人设名称 | 特点 |
|------|---------|------|
| `persona.yaml` | 陆晨 | 温柔但疏离的调酒师/摄影师 |
| `persona_xiaoai_v2.yaml` | 小艾 v2 | 活泼可爱的元气少女 |
| `persona_xiaoai.yaml` | 小艾 v1 | 初版人设（已优化） |
| `persona_amo.yaml` | Amo | 冷酷傲娇的智能助手 |

### 切换 Persona

**方式一：修改 Docker 挂载路径**
```powershell
# 编辑 docker-compose.qqbot.yaml
-v "d:\file\MemoryOs\examples\qqbot\persona_xiaoai_v2.yaml:/app/config/persona.yaml:ro"

# 重📊 数据管理

### 查看聊天记录（pgAdmin）

1. 访问 http://localhost:15432
2. 登录 PostgreSQL：
   - Host: `memoryos-postgres`（Docker）/ `localhost:15432`（本地）
   - User: `memoryos`
   - Password: `memoryos123`
   - Database: `memoryos`

3. 查询对话记忆：
```sql
-- 查看最近 10 条对话
SELECT user_id, content, role, created_at 
FROM dialogue_memory 
ORDER BY created_at DESC 
LIMIT 10;

-- 查看某用户的好感度变化
SELECT user_id, metadata->>'favorability' as favorability, created_at
FROM dialogue_memory
WHERE user_id = '你的QQ号'
ORDER BY created_at;
```

### 重置数据库

```powershell
# 清空所有记忆表
docker exec -it memoryos-postgres psql -U memoryos -d memoryos -c "TRUNCATE TABLE dialogue_memory, topic_memory, profile_memory RESTART IDENTITY CASCADE;"

# 清空 Milvus 向量库（如需）
docker-compose down
Remove-Item -Recurse -Force .\data\milvus, .\data\etcd, .\data\minio
docker-compose up -d
```/ 测试并发
go bot.ReceiveMessage("user1", "消息1")
go bot.ReceiveMessage("user2", "消息2")
go bot.ReceiveMessage("user3", "消息3")

// 测试好感度
bot.ReceiveMessage("user1", "谢谢你")  // 触发好感度+5

// 测试队列过载
for i := 0; i < 200; i++ {
    bot.ReceiveMessage("flood", "spam")
}
```

## 🎨 自定义人设

编辑 `main.go` 中的 `PersonaConfig`：

```go
persona := &chatbot.PersonaConfig{
    Name:         "阿尔法",          // 机器人名字
    Gender:       "中性",
    Age:          "未知",
    Personality:  []string{"冷静", "理性", "专业"},
    Background:   "AI 助手，专注于技术支持",
    Interests:    []string{"编程", "数学", "科幻"},
    TalkingStyle: "简洁专业，避免表情符号",
    Forbidden:    []string{"政治", "暴力", "色情"},
}
```

## 📊 性能指标

- **并发能力**：5 个 Worker 可同时处理 5 个用户
- **队列容量**：100 条消息（可调整）
- **响应延迟**：2-20 秒（模拟人类打字）
- **内存占用**：~50MB（取决于对话量）

## 🔧 故障排除

### � 故障排除

### 问题 1：无法连接 NapCat

**现象**：`❌ WebSocket 连接失败: dial tcp: connection refused`

**解决方案**：
1. 确认 NapCat 是否启动：`docker ps | grep napcat`
2. 检查端口映射：`docker port napcat`
3. 修改 `CQHTTP_WS_URL` 为正确地址

### 问题 2：VectorStore 使用 Mock

**现象**：`⚠️ VectorStore 未配置或不支持，使用 Mock`

**解决方案**：
```yaml
# config/config.docker.yaml
vector:
  provider: "milvus"  # 改为 milvus
  milvus:
    host: "memoryos-milvus"  # 确保容器名正确
    port: 19530
```
重启容器：`docker restart memoryos-qqbot`

### 问题 3：Bot 不回复消息

**排查步骤**：
```powershell
# 1. 查看容器日志
docker logs --tail 50 memoryos-qqbot

# 2. 检查数据库连接
docker exec -it memoryos-postgres psql -U memoryos -d memoryos -c "\dt"

# 3. 验证 LLM API
curl -X POST https://api.lingyaai.cn/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"model":"gemini-3-flash-preview","messages":[{"role":"user","content":"test"}]}'
```

## 📈 架构与性能

### 技术栈
- **消息接收**：WebSocket (gorilla/websocket)
- **并发处理**：Worker Pool（可配置）
- **记忆检索**：Milvus 向量检索 + PostgreSQL 元数据
- **LLM 生成**：支持 OpenAI / Gemini / 灵雅 AI

### 性能指标
- **并发处理**：5 Workers（可扩展到 20+）
- **消息队列**：100 条缓冲（可调整）
- **召回耗时**：~50ms（Milvus）/ ~200ms（pgvector）
- **端到端延迟**：1-3 秒（含 LLM 生成）

## 📚 相关文档

- [NapCat 部署指南](NAPCAT_SETUP.md)
- [Persona 改进报告](../../PERSONA_IMPROVEMENT_REPORT.md)
- [项目架构分析](../../PROJECT_STRUCTURE_ANALYSIS.md)
- [重构记录](../../REFACTORING_REPORT.md)

---

**生产级 QQ Bot 示例，完整展示 MemoryOS 长期记忆能力