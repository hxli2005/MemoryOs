# MemoryOS API 使用指南

## 📡 完整 API 列表

### 基础接口

#### 1. 健康检查
```bash
curl http://localhost:8080/health
```

**响应示例**：
```json
{
  "status": "healthy",
  "service": "MemoryOS",
  "version": "0.1.0",
  "mode": "Mock",
  "db": false,
  "redis": false
}
```

---

### 记忆管理 API

#### 2. 创建记忆

**创建对话记忆**：
```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "layer": "dialogue",
    "type": "user_message",
    "content": "我想学习 Go 的并发编程",
    "metadata": {
      "session_id": "session_001",
      "turn_number": 1,
      "role": "user"
    }
  }'
```

**创建用户画像**：
```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "layer": "profile",
    "type": "user_identity",
    "content": "Alice 是一名后端工程师，主要使用 Python，正在学习 Go",
    "metadata": {
      "category": "identity",
      "tags": ["后端工程师", "Python", "Go学习者"],
      "confidence_level": 0.9,
      "is_pinned": true
    }
  }'
```

**响应示例**：
```json
{
  "message": "记忆创建成功",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

#### 3. 搜索记忆

```bash
curl -X POST http://localhost:8080/api/v1/memories/search \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "query": "Go 并发编程",
    "top_k": 5
  }'
```

**响应示例**：
```json
{
  "count": 2,
  "memories": [
    {
      "id": "xxx",
      "user_id": "user_alice",
      "layer": "dialogue",
      "type": "user_message",
      "content": "我想学习 Go 的并发编程",
      "importance": 0.6,
      "access_count": 3,
      "created_at": "2026-01-21T15:30:00Z"
    }
  ]
}
```

---

### 召回接口

#### 4. 召回对话上下文

获取最近 N 轮对话（用于构建 LLM context window）：

```bash
curl -X POST http://localhost:8080/api/v1/recall/dialogue \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "session_id": "session_001",
    "recent_turns": 10
  }'
```

**响应示例**：
```json
{
  "session_id": "session_001",
  "count": 4,
  "dialogue": [
    {
      "id": "xxx",
      "content": "我想学习 Go 的并发编程",
      "metadata": {
        "turn_number": 1,
        "role": "user"
      }
    },
    {
      "id": "yyy",
      "content": "很好！Go 的并发模型是其最强大的特性...",
      "metadata": {
        "turn_number": 2,
        "role": "assistant"
      }
    }
  ]
}
```

---

#### 5. 召回话题线索

根据当前 query 召回相关话题：

```bash
curl -X POST http://localhost:8080/api/v1/recall/topic \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "query": "继续刚才关于 goroutine 的讨论",
    "top_k": 5
  }'
```

**使用场景**：
- 用户说"继续刚才的话题"
- 跨会话话题延续
- 话题切换提示

---

#### 6. 召回用户画像

快速加载用户认知：

```bash
curl -X POST http://localhost:8080/api/v1/recall/profile \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "category": "identity"
  }'
```

**category 可选值**：
- `identity` - 用户身份
- `style` - 沟通风格
- `personality` - 人格特质
- `preference` - 偏好记录
- 留空 - 返回所有画像

**响应示例**：
```json
{
  "user_id": "user_alice",
  "category": "identity",
  "count": 1,
  "profile": [
    {
      "content": "Alice 是一名后端工程师...",
      "metadata": {
        "tags": ["后端工程师", "Python", "Go学习者"],
        "confidence_level": 0.9,
        "is_pinned": true
      }
    }
  ]
}
```

---

#### 7. 混合召回（核心创新）

**自适应召回策略** - 根据对话阶段动态调整三层比例：

```bash
curl -X POST http://localhost:8080/api/v1/recall/hybrid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "session_id": "session_001",
    "query": "goroutine 的性能优化技巧",
    "dialog_stage": "multi_turn",
    "max_tokens": 4000
  }'
```

**dialog_stage 参数**：
- `session_start` - 新会话开始（80% Profile + 15% Topic + 5% Dialogue）
- `topic_deepening` - 话题深入中（30% Profile + 50% Topic + 20% Dialogue）
- `multi_turn` - 多轮对话（10% Profile + 20% Topic + 70% Dialogue）

**响应示例**：
```json
{
  "dialogue_count": 7,
  "topic_count": 2,
  "profile_count": 1,
  "strategy": "multi_turn",
  "tokens_used": 3850,
  "dialogue": [...],
  "topics": [...],
  "profile": [...]
}
```

---

## 🎯 典型使用场景

### 场景 1：Chatbot 对话流程

```javascript
// 1. 用户发送消息
const userMessage = {
  user_id: "user_alice",
  layer: "dialogue",
  type: "user_message",
  content: userInput,
  metadata: {
    session_id: currentSessionId,
    turn_number: currentTurn,
    role: "user"
  }
};

await fetch('/api/v1/memories', {
  method: 'POST',
  body: JSON.stringify(userMessage)
});

// 2. 混合召回相关记忆
const recallResult = await fetch('/api/v1/recall/hybrid', {
  method: 'POST',
  body: JSON.stringify({
    user_id: "user_alice",
    session_id: currentSessionId,
    query: userInput,
    dialog_stage: "multi_turn"
  })
}).then(r => r.json());

// 3. 构建 LLM Prompt
const context = [
  ...recallResult.profile,    // 用户画像
  ...recallResult.topics,     // 相关话题
  ...recallResult.dialogue    // 对话历史
];

const llmResponse = await callLLM(context, userInput);

// 4. 存储助手回复
await fetch('/api/v1/memories', {
  method: 'POST',
  body: JSON.stringify({
    user_id: "user_alice",
    layer: "dialogue",
    type: "assistant_message",
    content: llmResponse,
    metadata: {
      session_id: currentSessionId,
      turn_number: currentTurn + 1,
      role: "assistant"
    }
  })
});
```

---

### 场景 2：会话开始时加载用户画像

```bash
curl -X POST http://localhost:8080/api/v1/recall/profile \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice"
  }'
```

用于：
- 个性化问候
- 调整回复风格
- 意图预判

---

### 场景 3：话题延续

用户："继续刚才的话题"

```bash
curl -X POST http://localhost:8080/api/v1/recall/topic \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_alice",
    "query": "继续刚才的话题",
    "top_k": 3
  }'
```

---

## 📋 记忆层级与类型

### Dialogue 层（短期，快速衰减）
- `user_message` - 用户消息
- `assistant_message` - 助手回复
- `dialogue_context` - 上下文快照

### Topic 层（中期，中速衰减）
- `topic_thread` - 话题线索
- `intent` - 意图识别
- `conversation_flow` - 对话流转

### Profile 层（长期，几乎不衰减）
- `user_identity` - 用户身份
- `communication_style` - 沟通风格
- `personality` - 人格特质
- `preference` - 偏好记录

---

## ⚠️ 当前限制（Mock 模式）

1. ✅ API 接口完整可用
2. ⚠️ 数据不持久化（重启后清空）
3. ⚠️ 向量检索返回空结果
4. ⚠️ LLM 相关功能需要配置 API Key

## 🔧 下一步

配置真实数据库后将支持：
- ✅ 数据持久化
- ✅ 向量相似度检索
- ✅ 按时间/类型过滤
- ✅ 分页查询
