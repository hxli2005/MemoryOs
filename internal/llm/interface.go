package llm

import (
	"context"

	"github.com/yourusername/MemoryOs/internal/model"
	"github.com/yourusername/MemoryOs/internal/service/memory"
)

// LLMClient 定义 LLM 服务的接口
// 用于记忆聚合、提炼、分析等高级功能
type LLMClient interface {
	// SummarizeDialogues 对话聚合：从多轮对话中提炼话题摘要
	// 输入：一组对话记忆（同一 session）
	// 输出：话题标题、摘要、关键词
	SummarizeDialogues(ctx context.Context, dialogues []*model.Memory) (*memory.TopicSummary, error)

	// ExtractProfile 画像提炼：从多个话题中分析用户特征
	// 输入：一组话题记忆（同一 user）
	// 输出：用户偏好、习惯、特征
	ExtractProfile(ctx context.Context, topics []*model.Memory) (*memory.UserProfile, error)

	// AnalyzeIntent 意图分析：判断用户当前对话意图
	// 输入：当前对话内容
	// 输出：意图类型（提问/闲聊/任务请求等）
	AnalyzeIntent(ctx context.Context, userMessage string) (string, error)
}
