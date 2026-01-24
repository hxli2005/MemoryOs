package model

import "time"

// MemoryLayer 记忆层级（Chatbot Intent Memory 三层架构）
type MemoryLayer string

const (
	LayerDialogue MemoryLayer = "dialogue" // 对话层：原始对话轮次（短期，快速衰减）
	LayerTopic    MemoryLayer = "topic"    // 话题层：话题线索与上下文（中期，中速衰减）
	LayerProfile  MemoryLayer = "profile"  // 画像层：用户认知与风格（长期，几乎不衰减）
)

// MemoryType 记忆类型（Chatbot Intent Memory 专用）
type MemoryType string

const (
	// 对话层 (Dialogue Layer) - 短期记忆
	MemoryTypeUserMessage      MemoryType = "user_message"      // 用户消息原文
	MemoryTypeAssistantMessage MemoryType = "assistant_message" // 助手回复原文
	MemoryTypeDialogueContext  MemoryType = "dialogue_context"  // 对话上下文快照

	// 话题层 (Topic Layer) - 中期记忆（对抗话题连续性熵增）
	MemoryTypeTopicThread      MemoryType = "topic_thread"      // 话题线索（聚合多轮对话）
	MemoryTypeIntent           MemoryType = "intent"            // 用户意图识别结果
	MemoryTypeConversationFlow MemoryType = "conversation_flow" // 对话流转记录

	// 画像层 (Profile Layer) - 长期记忆（对抗人格熵增）
	MemoryTypeUserIdentity       MemoryType = "user_identity"       // 用户是谁（职业、背景、兴趣）
	MemoryTypeCommunicationStyle MemoryType = "communication_style" // 沟通风格（语气、表达偏好）
	MemoryTypePersonality        MemoryType = "personality"         // 人格特质（价值观、态度）
	MemoryTypePreference         MemoryType = "preference"          // 偏好记录（喜好、禁忌）
)

// Memory 记忆实体（Chatbot Intent Memory 架构）
type Memory struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Layer     MemoryLayer `json:"layer"`   // 记忆层级（dialogue/topic/profile）
	Type      MemoryType  `json:"type"`    // 记忆类型（层级内细分）
	Content   string      `json:"content"` // 记忆内容
	Embedding []float32   `json:"-"`       // 向量表示，不序列化到 JSON

	// 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// 记忆强度相关
	Importance   float64   `json:"importance"`   // 重要性评分 0-1
	AccessCount  int       `json:"access_count"` // 访问次数
	LastAccessed time.Time `json:"last_accessed"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchRequest 记忆搜索请求
type SearchRequest struct {
	Query   string                 `json:"query"`
	UserID  string                 `json:"user_id"`
	Type    MemoryType             `json:"type,omitempty"`
	TopK    int                    `json:"top_k"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Memory *Memory `json:"memory"`
	Score  float64 `json:"score"` // 相似度分数
}

// ========== Chatbot Intent Memory 元数据定义 ==========

// DialogueMetadata 对话层元数据
// 对抗：短期遗忘，保留对话细节
type DialogueMetadata struct {
	SessionID    string    `json:"session_id"`              // 会话 ID
	TurnNumber   int       `json:"turn_number"`             // 对话轮次（从1开始）
	Role         string    `json:"role"`                    // user/assistant/system
	Timestamp    time.Time `json:"timestamp"`               // 精确时间戳
	PreviousTurn string    `json:"previous_turn,omitempty"` // 上一轮 ID（建立对话链）
	NextTurn     string    `json:"next_turn,omitempty"`     // 下一轮 ID
}

// TopicMetadata 话题层元数据
// 对抗：话题连续性熵增，维持话题线索
type TopicMetadata struct {
	TopicID         string    `json:"topic_id"`                   // 话题唯一标识
	ParentTopicID   string    `json:"parent_topic_id,omitempty"`  // 父话题（话题树）
	TopicKeywords   []string  `json:"topic_keywords,omitempty"`   // 话题关键词
	IntentCategory  string    `json:"intent_category,omitempty"`  // 意图分类：ask/inform/command/chitchat
	Sentiment       string    `json:"sentiment,omitempty"`        // 情感：positive/neutral/negative
	AggregatedTurns []int     `json:"aggregated_turns,omitempty"` // 聚合自哪些对话轮次
	TopicStatus     string    `json:"topic_status,omitempty"`     // active/paused/completed
	StartedAt       time.Time `json:"started_at"`                 // 话题开始时间
	LastActiveAt    time.Time `json:"last_active_at"`             // 最后活跃时间
}

// ProfileMetadata 画像层元数据
// 对抗：人格熵增，稳定用户认知
type ProfileMetadata struct {
	Category        string          `json:"category"`                 // 画像类别：identity/style/personality/preference
	Tags            []string        `json:"tags,omitempty"`           // 高密度标签
	ConfidenceLevel float64         `json:"confidence_level"`         // 置信度（0-1）
	Evidence        []string        `json:"evidence,omitempty"`       // 证据来源（DialogueID/TopicID）
	LastConfirmed   *time.Time      `json:"last_confirmed,omitempty"` // 最后确认时间
	ConflictsWith   []string        `json:"conflicts_with,omitempty"` // 冲突的画像 ID
	IsPinned        bool            `json:"is_pinned"`                // 是否钉住（核心画像不衰减）
	UpdateHistory   []ProfileUpdate `json:"update_history,omitempty"` // 画像更新历史
}

// ProfileUpdate 画像更新记录
type ProfileUpdate struct {
	Timestamp time.Time `json:"timestamp"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	UpdatedBy string    `json:"updated_by"` // dialogue_id 或 topic_id
}
