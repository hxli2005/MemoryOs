package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"google.golang.org/genai"

	memoryModel "github.com/yourusername/MemoryOs/internal/model"
	"github.com/yourusername/MemoryOs/internal/service/memory"
)

// GeminiClient Gemini LLM 客户端实现
type GeminiClient struct {
	chatModel model.ChatModel
	config    GeminiConfig
}

// GeminiConfig Gemini 配置
type GeminiConfig struct {
	APIKey  string
	Model   string // 默认: gemini-2.0-flash-exp
	BaseURL string // 可选，自定义 API 端点
}

// NewGeminiClient 创建 Gemini 客户端
func NewGeminiClient(cfg GeminiConfig) (*GeminiClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "gemini-2.0-flash-exp" // 默认使用免费的 Flash 模型
	}

	// 1. 创建 genai.Client
	genaiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: cfg.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// 2. 创建 Gemini ChatModel
	chatModel, err := gemini.NewChatModel(context.Background(), &gemini.Config{
		Client: genaiClient,
		Model:  cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini chat model: %w", err)
	}

	return &GeminiClient{
		chatModel: chatModel,
		config:    cfg,
	}, nil
}

// GenerateText 通用文本生成（用于 Chatbot 回复）
func (c *GeminiClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("gemini generate failed: %w", err)
	}

	if resp == nil || resp.Content == "" {
		return "", fmt.Errorf("empty response from gemini")
	}

	return resp.Content, nil
}

// SummarizeDialogues 实现对话聚合
func (c *GeminiClient) SummarizeDialogues(ctx context.Context, dialogues []*memoryModel.Memory) (*memory.TopicSummary, error) {
	if len(dialogues) == 0 {
		return nil, fmt.Errorf("empty dialogues")
	}

	// 构建对话历史文本
	var conversationText strings.Builder
	var dialogueIDs []string
	conversationText.WriteString("以下是一段完整的对话记录：\n\n")

	for i, mem := range dialogues {
		conversationText.WriteString(fmt.Sprintf("%d. %s\n", i+1, mem.Content))
		dialogueIDs = append(dialogueIDs, mem.ID)
	}

	// 构建 Prompt
	prompt := fmt.Sprintf(`%s

请分析这段对话，提取以下信息：
1. 话题标题（5-10字，精炼概括核心主题）
2. 话题摘要（50-200字，详细描述对话内容和要点）
3. 关键词（3-5个，用逗号分隔）

请严格按照以下 JSON 格式返回，不要添加任何其他内容：
{
  "title": "话题标题",
  "summary": "详细摘要内容",
  "keywords": ["关键词1", "关键词2", "关键词3"]
}`, conversationText.String())

	// 调用 Gemini
	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: prompt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}

	// 解析响应
	content := resp.Content

	// 清理 Markdown 代码块标记（Gemini 可能返回 ```json ... ```）
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 解析 JSON
	var result struct {
		Title    string   `json:"title"`
		Summary  string   `json:"summary"`
		Keywords []string `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w\nResponse: %s", err, content)
	}

	return &memory.TopicSummary{
		Title:       result.Title,
		Summary:     result.Summary,
		Keywords:    result.Keywords,
		DialogueIDs: dialogueIDs,
	}, nil
}

// ExtractProfile 实现画像提炼
func (c *GeminiClient) ExtractProfile(ctx context.Context, topics []*memoryModel.Memory) (*memory.UserProfile, error) {
	if len(topics) == 0 {
		return nil, fmt.Errorf("empty topics")
	}

	// 构建话题列表文本
	var topicsText strings.Builder
	var topicIDs []string
	topicsText.WriteString("以下是用户的历史话题记录：\n\n")

	for i, mem := range topics {
		topicsText.WriteString(fmt.Sprintf("%d. %s\n", i+1, mem.Content))
		topicIDs = append(topicIDs, mem.ID)
	}

	// 构建 Prompt
	prompt := fmt.Sprintf(`%s

请深度分析这些话题，提炼用户画像，包括：

1. **偏好特征 (Preferences)**：用户的兴趣爱好、喜欢的话题类型、沟通风格等
2. **行为习惯 (Habits)**：对话频率、活跃时段、提问方式等
3. **认知特征 (Features)**：知识水平、学习能力、思维方式等

请严格按照以下 JSON 格式返回，不要添加任何其他内容：
{
  "preferences": {
    "interests": ["兴趣1", "兴趣2"],
    "communication_style": "风格描述"
  },
  "habits": {
    "active_time": "时段描述",
    "question_pattern": "提问方式"
  },
  "features": {
    "knowledge_level": "水平描述",
    "learning_ability": "能力描述"
  }
}`, topicsText.String())

	// 调用 Gemini
	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: prompt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("gemini api call failed: %w", err)
	}

	// 解析响应
	content := resp.Content

	// 清理 Markdown 代码块标记
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// 解析 JSON
	var result struct {
		Preferences map[string]interface{} `json:"preferences"`
		Habits      map[string]interface{} `json:"habits"`
		Features    map[string]interface{} `json:"features"`
	}
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w\nResponse: %s", err, content)
	}

	return &memory.UserProfile{
		Preferences: result.Preferences,
		Habits:      result.Habits,
		Features:    result.Features,
		TopicIDs:    topicIDs,
	}, nil
}

// AnalyzeIntent 实现意图分析
func (c *GeminiClient) AnalyzeIntent(ctx context.Context, userMessage string) (string, error) {
	prompt := fmt.Sprintf(`分析以下用户消息的意图，从以下类型中选择一个：
- question: 用户在提问
- chat: 用户在闲聊
- task: 用户在请求执行任务
- feedback: 用户在提供反馈

用户消息："%s"

只返回意图类型（question/chat/task/feedback），不要添加任何其他内容。`, userMessage)

	resp, err := c.chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: prompt,
		},
	})
	if err != nil {
		return "", fmt.Errorf("gemini api call failed: %w", err)
	}

	intent := strings.TrimSpace(strings.ToLower(resp.Content))

	// 验证返回值
	validIntents := map[string]bool{
		"question": true,
		"chat":     true,
		"task":     true,
		"feedback": true,
	}

	if !validIntents[intent] {
		return "chat", nil // 默认为闲聊
	}

	return intent, nil
}
