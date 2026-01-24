# QQ Bot - MemoryOS 应用示例

这是一个基于 MemoryOS 的 QQ 聊天机器人示例，展示了如何通过松耦合的方式使用 MemoryOS 核心能力。

## 🎯 设计原则

- **松耦合**：应用层与核心层通过 `pkg/chatbot` 接口交互
- **可替换**：可以轻松替换为微信、Telegram 等其他平台
- **易测试**：独立的应用逻辑便于单元测试和集成测试
- **低侵入**：不修改 MemoryOS 核心代码

## 🏗️ 架构

```
QQ 消息 → QQBot (应用层) → chatbot.Adapter (接口层) → MemoryOS (核心层)
```

### 关键组件

1. **pkg/chatbot/interface.go** - 通用聊天机器人接口
2. **pkg/chatbot/adapter.go** - MemoryOS 适配器实现
3. **examples/qqbot/main.go** - QQ Bot 应用示例

## 🚀 快速开始

### 1. 配置 MemoryOS

确保 `config/config.yaml` 已正确配置 LLM API Key：

```yaml
llm:
  provider: "openai"
  api_key: "YOUR_API_KEY"
  model: "gpt-4o-mini"
```

### 2. 运行示例

```bash
# 方式一：直接运行
go run examples/qqbot/main.go

# 方式二：构建后运行
go build -o qqbot.exe examples/qqbot/main.go
./qqbot.exe
```

### 3. 测试效果

当前代码包含模拟的消息收发，你会看到：

```
🤖 QQ Bot - MemoryOS 应用示例
==================================================
🚀 启动 5 个消息处理 Worker

📨 [123456789] 收到消息: 你好呀
⏰ [Worker 0] 延迟 2s 后回复...
💬 [123456789] 机器人回复: 你好呀~ (｡•́︿•̀｡)

📨 [123456789] 收到消息: 今天天气怎么样？
⏰ [Worker 1] 延迟 5s 后回复...
💬 [123456789] 机器人回复: 让我看看...
```

## 🔌 接入真实 QQ（可选）

### 使用 go-cqhttp

1. **下载 go-cqhttp**
   ```bash
   # https://github.com/Mrs4s/go-cqhttp/releases
   ```

2. **配置 WebSocket 连接**
   修改 `examples/qqbot/main.go` 中的 `ReceiveMessage` 方法，接入 WebSocket：

   ```go
   import "github.com/gorilla/websocket"
   
   func (b *QQBot) connectToGoCQHTTP() {
       conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:6700", nil)
       // 处理 WebSocket 消息...
   }
   ```

## 📋 核心功能

### ✅ 已实现

- [x] 消息队列 + Worker Pool（并发处理）
- [x] 延迟回复（模拟打字）
- [x] 用户画像管理
- [x] 好感度系统（框架）
- [x] 松耦合架构

### 🚧 待完善（作为测试改进方向）

- [ ] LLM 回复生成（当前返回占位文本）
- [ ] 人设 Prompt 构建
- [ ] 复杂度分析（智能延迟）
- [ ] 优先级队列
- [ ] 真实 go-cqhttp 集成

## 🧪 测试方式

### 单元测试

```bash
# 测试 Chatbot 接口
go test ./pkg/chatbot/...

# 测试 QQBot 逻辑
go test ./examples/qqbot/...
```

### 集成测试

修改 `main.go` 中的模拟消息，测试不同场景：

```go
// 测试并发
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

### 问题1：消息队列满

**现象**：看到 "⚠️ 队列已满，消息被丢弃"

**解决**：增加 Worker 数量或队列容量：
```go
bot := NewQQBot(adapter, 10)  // 10 个 Worker
```

### 问题2：LLM 调用失败

**现象**：机器人回复 "啊这...我好像卡住了"

**检查**：
1. `config/config.yaml` 中的 API Key 是否正确
2. `adapter.go` 中的 `generateResponse` 方法是否实现

## 🎯 作为测试工具

这个 QQ Bot 是测试 MemoryOS 的绝佳工具：

### 测试点

1. **并发安全性** - 多个用户同时发消息
2. **记忆持久化** - 重启后是否记得用户
3. **召回准确性** - 是否能找到相关历史
4. **性能压力** - 处理大量消息的能力
5. **错误恢复** - LLM 失败时的降级策略

### 建议测试流程

```bash
# 1. 启动 QQ Bot
go run examples/qqbot/main.go

# 2. 模拟多用户并发对话
# 3. 观察日志输出
# 4. 检查 PostgreSQL 数据库中的记忆存储
# 5. 测试异常场景（断网、API 超时等）
```

## 📝 后续改进方向

1. **实现真实 LLM 调用** - 完成 `generateResponse` 方法
2. **添加配置文件** - 支持从 YAML 加载人设
3. **监控面板** - Web UI 显示消息队列、好感度等
4. **日志系统** - 结构化日志便于分析
5. **性能优化** - 缓存、批处理等

---

**这是一个独立的应用示例，不会污染 MemoryOS 核心代码！** 🎉
