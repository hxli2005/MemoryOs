# Gemini Pro 省钱策略 💰

## 一、免费额度对比

| 模型                    | 免费额度（每分钟） | 免费额度（每天） | 成本（超出后）     |
|------------------------|------------------|-----------------|-------------------|
| gemini-2.0-flash-exp   | ♾️ **无限制**    | ♾️ **无限制**   | **完全免费**      |
| gemini-1.5-flash       | 15 RPM           | 1500 RPD        | $0.075/$0.30 /1M  |
| gemini-1.5-pro         | 2 RPM            | 50 RPD          | $1.25/$5.00 /1M   |

> **RPM** = Requests Per Minute（每分钟请求数）  
> **RPD** = Requests Per Day（每天请求数）

## 二、最划算策略 🎯

### ✅ 推荐配置（日常使用 Flash，需要时切换 Pro）

```yaml
# config/config.yaml

llm:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"
  model: "gemini-2.0-flash-exp"  # 默认用免费的 Flash
  base_url: ""

embedding:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"
  model: "text-embedding-004"    # Embedding 一直用免费的
  dimension: 768
```

### 📊 什么时候用 Flash，什么时候用 Pro？

#### 用 Flash 的场景（99% 情况）：
- ✅ **日常对话**：闲聊、简单问答
- ✅ **记忆召回**：检索用户画像、话题线
- ✅ **短文本生成**：回复用户消息
- ✅ **开发测试**：调试代码、验证功能

**原因**：gemini-2.0-flash-exp 完全免费且性能接近 Pro

#### 用 Pro 的场景（<1% 情况）：
- 💎 **复杂推理**：多步逻辑推导、数学问题
- 💎 **长文本分析**：处理超长对话历史（>100轮）
- 💎 **精确任务**：法律文档分析、医疗咨询
- 💎 **创意生成**：写小说、诗歌创作

**原因**：Pro 在复杂任务上准确率更高

### 💡 切换模型的方法

#### 方法一：修改配置文件（推荐）

```bash
# 需要 Pro 时，编辑 config/config.yaml
vim config/config.yaml

# 将 model 改为：
model: "gemini-1.5-pro"

# 重启程序
.\start_chatbot.bat
```

#### 方法二：代码中动态切换（高级）

在 `examples/chatbot/main.go` 中添加检测逻辑：

```go
// 根据用户输入复杂度选择模型
func (c *Chatbot) selectModel(userInput string) string {
    // 简单启发式规则
    if len(userInput) > 500 || strings.Contains(userInput, "复杂") {
        return "gemini-1.5-pro"  // 长文本或显式要求用 Pro
    }
    return "gemini-2.0-flash-exp"  // 默认用 Flash
}
```

## 三、Embedding 策略

### ✅ 永远用免费的 `text-embedding-004`

**原因**：
1. **完全免费**（无额度限制）
2. **性能优秀**（768 维足够用）
3. **与 Pro 无关**（Embedding 不需要 Pro 级别能力）

```yaml
embedding:
  provider: "gemini"
  model: "text-embedding-004"  # 👈 不要改
  dimension: 768
```

## 四、省钱小技巧 💸

### 1️⃣ 控制 Token 数量

```yaml
llm:
  max_tokens: 1000  # 限制每次回复的 Token 数（默认 2000 太多）
```

### 2️⃣ 优化 Prompt

❌ **不好的 Prompt**（浪费 Token）：
```
你是一个智能助手，请根据以下所有用户历史记忆、详细画像、完整对话历史... [1000+ tokens]
```

✅ **好的 Prompt**（精简有效）：
```
用户画像：软件工程师，喜欢简洁回答
上次聊到：Go 并发模型
用户问：channel 和 goroutine 的关系？
```

### 3️⃣ 批量处理记忆

```go
// ❌ 每条记忆都调用 LLM（浪费）
for _, memory := range memories {
    llm.Summarize(memory)  // 调用 N 次
}

// ✅ 批量处理（省钱）
batch := strings.Join(memories, "\n")
llm.Summarize(batch)  // 只调用 1 次
```

### 4️⃣ 缓存常见问题

```go
// 缓存常见问题的回复
cache := map[string]string{
    "你好": "你好！有什么可以帮你的吗？",
    "再见": "再见！祝你愉快！",
}

if reply, ok := cache[userInput]; ok {
    return reply  // 不调用 LLM
}
```

## 五、监控成本 📊

### 查看 API 使用情况

访问 [Google AI Studio Dashboard](https://aistudio.google.com/app/apikey)：
- 查看每日请求数
- 查看 Token 消耗
- 设置使用上限（可选）

### 添加日志记录

在 `internal/adapter/eino.go` 中：

```go
func (e *EinoLLM) Chat(ctx context.Context, messages []LLMMessage) (string, error) {
    start := time.Now()
    
    resp, err := e.model.Generate(ctx, einoMsgs)
    
    // 记录使用情况
    log.Printf("[LLM] Model: %s, Time: %v, Tokens: ~%d", 
        e.model.Name(), 
        time.Since(start), 
        len(resp.Content)/4)  // 粗略估算 Token 数
    
    return resp.Content, nil
}
```

## 六、实际成本估算

假设你每天使用 Chatbot **100 次对话**：

| 场景          | 模型            | 每次 Token | 每天 Token | 每月成本   |
|--------------|----------------|-----------|-----------|-----------|
| **日常使用** | Flash (免费)   | ~500      | 50K       | **$0**    |
| **混合使用** | 90% Flash + 10% Pro | ~550 | 55K  | **~$0.15** |
| **全用 Pro** | Pro (付费)     | ~500      | 50K       | **~$1.50** |

> 结论：**混合使用策略每月只需 $0.15**，节省 90% 成本！

## 七、快速配置指南

### Step 1: 填入 API Key

```yaml
# config/config.yaml
llm:
  api_key: "AIzaSy..."  # 👈 你的真实 Gemini API Key
  
embedding:
  api_key: "AIzaSy..."  # 👈 同上
```

### Step 2: 选择模型

**日常使用（推荐）**：
```yaml
llm:
  model: "gemini-2.0-flash-exp"  # 免费 + 快速
```

**需要高质量时**：
```yaml
llm:
  model: "gemini-1.5-pro"  # 付费但准确率更高
```

### Step 3: 启动测试

```bash
.\start_chatbot.bat
```

## 八、常见问题

### ❓ Flash 和 Pro 性能差距大吗？

**日常对话**：几乎无差异（Flash 已经很强）  
**复杂推理**：Pro 准确率高约 10-15%

### ❓ 如何知道我用的是哪个模型？

查看启动日志：
```
[INFO] 使用模型: gemini-2.0-flash-exp
```

或在对话中问 Chatbot：
```
You: 你现在用的是什么模型？
Bot: 我正在使用 gemini-2.0-flash-exp 模型
```

### ❓ 超过免费额度会怎样？

**Flash 2.0 Exp**：无额度限制，永久免费  
**Flash 1.5/Pro**：超过后自动按量计费（可在 Google Cloud Console 设置上限）

---

## 🎯 最终建议

**最划算的配置**：

```yaml
# config/config.yaml

llm:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"
  model: "gemini-2.0-flash-exp"  # 👈 99% 场景用这个
  max_tokens: 1000               # 👈 限制 Token 数

embedding:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"
  model: "text-embedding-004"    # 👈 永远用免费的
  dimension: 768
```

**需要 Pro 时**：手动改为 `model: "gemini-1.5-pro"`，用完再改回来。

**预计每月成本**：**$0** （纯 Flash）到 **$0.15** （偶尔用 Pro）

---

**省钱就是赚钱！** 💰✨
