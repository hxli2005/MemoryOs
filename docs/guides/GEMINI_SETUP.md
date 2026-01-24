# Google Gemini 配置指南

## 为什么选择 Google Gemini？

✅ **免费额度更高**：Gemini 2.0 Flash Exp 目前完全免费  
✅ **性能优秀**：gemini-2.0-flash-exp 性能接近 GPT-4  
✅ **多模态支持**：原生支持图片、音频、视频理解  
✅ **无需信用卡**：直接使用 Google 账号即可获取 API Key

## 快速开始

### 1️⃣ 获取 Gemini API Key

1. 访问 [Google AI Studio](https://aistudio.google.com/app/apikey)
2. 使用 Google 账号登录
3. 点击 **"Get API Key"** 按钮
4. 点击 **"Create API Key"** → 选择项目或创建新项目
5. 复制生成的 API Key（格式类似：`AIzaSy...`）

> 💡 提示：Gemini API Key 没有使用期限，可以长期使用

### 2️⃣ 配置 MemoryOS

编辑 `config/config.yaml`：

```yaml
llm:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"  # 👈 粘贴你的 API Key
  model: "gemini-2.0-flash-exp"   # 推荐模型（免费且强大）
  base_url: ""                     # Gemini 不需要 base_url

embedding:
  provider: "gemini"
  api_key: "YOUR_GEMINI_API_KEY"  # 👈 同上
  model: "text-embedding-004"      # Gemini 推荐的 embedding 模型
  dimension: 768                   # text-embedding-004 输出 768 维向量
```

### 3️⃣ 启动 Chatbot

```bash
# Windows
.\start_chatbot.bat

# Linux/Mac
cd examples/chatbot
go run main.go
```

## Gemini 模型选择

### LLM 模型对比

| 模型                     | 特点                                | 推荐场景           | 价格      |
|-------------------------|-------------------------------------|-------------------|----------|
| `gemini-2.0-flash-exp`  | 🌟 最新实验版，性能最强，速度快      | **推荐首选**       | **免费** |
| `gemini-1.5-flash`      | 稳定版，速度快                       | 生产环境           | 低成本   |
| `gemini-1.5-pro`        | 性能强大，上下文窗口大（2M tokens） | 复杂任务           | 中等成本 |
| `gemini-2.0-flash-thinking-exp` | 支持推理过程可视化           | 需要思考链的场景   | **免费** |

### Embedding 模型对比

| 模型                     | 维度 | 特点                    | 推荐场景       |
|-------------------------|------|-------------------------|---------------|
| `text-embedding-004`    | 768  | 🌟 最新版，性能优秀      | **推荐首选**   |
| `embedding-001`         | 768  | 早期版本                | 兼容性需求     |

## 配置示例

### 完整配置（使用 Gemini）

```yaml
# config/config.yaml

server:
  host: "0.0.0.0"
  port: 8080

# 数据库配置（留空使用 Mock 模式）
database:
  postgres:
    host: ""
  redis:
    host: ""

# LLM 配置
llm:
  provider: "gemini"
  api_key: "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXX"
  model: "gemini-2.0-flash-exp"
  base_url: ""

# Embedding 配置
embedding:
  provider: "gemini"
  api_key: "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXX"
  model: "text-embedding-004"
  dimension: 768

vector:
  provider: "milvus"
  milvus:
    host: "localhost"
    port: 19530

memory:
  max_working_memory: 10
  compression_threshold: 100
  decay_days: 30
```

### 混合配置（Gemini + OpenAI）

你也可以混合使用不同的提供商：

```yaml
llm:
  provider: "gemini"           # LLM 使用 Gemini
  api_key: "YOUR_GEMINI_API_KEY"
  model: "gemini-2.0-flash-exp"

embedding:
  provider: "openai"           # Embedding 使用 OpenAI
  api_key: "YOUR_OPENAI_API_KEY"
  model: "text-embedding-3-small"
  dimension: 1536
```

## 功能特性

### ✅ 已支持的功能

- [x] 文本生成（Chat）
- [x] 流式输出（Streaming）
- [x] Embedding 生成
- [x] 多轮对话
- [x] 系统提示词
- [x] 工具调用（Tool Calling）
- [x] 混合召回（三层记忆架构）

### 🚧 Gemini 特有功能（待集成）

- [ ] 多模态输入（图片、音频、视频）
- [ ] 思考链可视化（gemini-2.0-flash-thinking-exp）
- [ ] 代码执行（Code Execution）
- [ ] Google 搜索集成

## 常见问题

### ❓ 如何切换回 OpenAI？

编辑 `config/config.yaml`，将 provider 改为 `openai`：

```yaml
llm:
  provider: "openai"
  api_key: "sk-YOUR_OPENAI_API_KEY"
  model: "gpt-4o-mini"
  base_url: "https://api.openai.com/v1"

embedding:
  provider: "openai"
  api_key: "sk-YOUR_OPENAI_API_KEY"
  model: "text-embedding-3-small"
  dimension: 1536
```

### ❓ Gemini 免费额度是多少？

当前 `gemini-2.0-flash-exp` 模型完全免费（实验阶段）。

正式版 Gemini 1.5 Flash 免费额度：
- **15 RPM** (每分钟请求数)
- **100万 TPM** (每分钟 Token 数)
- **1500 RPD** (每天请求数)

详见：[Gemini API 定价](https://ai.google.dev/pricing)

### ❓ Gemini API Key 安全吗？

⚠️ **重要**：API Key 是敏感信息，请：
- ✅ 不要提交到 Git 仓库
- ✅ 使用环境变量存储（可选）
- ✅ 定期轮换 API Key
- ✅ 为不同项目创建不同的 Key

### ❓ 遇到 "API key not valid" 错误？

1. 检查 API Key 是否正确复制（无多余空格）
2. 确认 API Key 已启用（访问 [API Dashboard](https://aistudio.google.com/app/apikey)）
3. 检查网络连接（Gemini API 需要访问 Google 服务）

### ❓ 为什么 dimension 从 1536 改为 768？

不同的 Embedding 模型输出维度不同：
- OpenAI `text-embedding-3-small`：**1536** 维
- Gemini `text-embedding-004`：**768** 维

如果遇到维度不匹配错误，请确保配置文件中的 `dimension` 与模型一致。

## 性能对比

| 指标         | Gemini 2.0 Flash Exp | GPT-4o-mini   | 备注                |
|-------------|---------------------|---------------|---------------------|
| 速度         | ⚡⚡⚡⚡⚡             | ⚡⚡⚡⚡        | Gemini 略快         |
| 成本         | **免费**            | 付费          | Gemini 优势明显     |
| 多语言支持   | ✅                  | ✅            | 两者都很好          |
| 上下文窗口   | 1M tokens           | 128K tokens   | Gemini 优势         |
| 多模态       | ✅ 图片/音频/视频   | ✅ 图片       | Gemini 更全面       |

## 下一步

- [Chatbot 使用指南](CHATBOT_USAGE.md)
- [API 文档](API_GUIDE.md)
- [Gemini 官方文档](https://ai.google.dev/docs)

---

**享受免费的 AI 能力吧！** 🎉
