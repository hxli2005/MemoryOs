package memory

import (
	"context" // 用于错误包装
	"fmt"
	"strings"
	"time"

	// 生成唯一 ID
	"github.com/google/uuid"
	"github.com/yourusername/MemoryOs/internal/model"
)

// Manager 记忆管理器 - 你的核心业务逻辑
type Manager struct {
	vectorStore VectorStore   // 向量存储
	metaStore   MetadataStore // 元数据存储（PostgreSQL）
	embedder    Embedder      // Embedding 客户端
	llm         LLMClient     // LLM 客户端（压缩/反思用）
	config      Config        // 配置
}

// Config 记忆管理配置
type Config struct {
	MaxWorkingMemory     int // 工作记忆最大条数
	CompressionThreshold int // 触发压缩的阈值
	DecayDays            int // 记忆衰减周期（天）
}

func NewManager(
	vectorStore VectorStore,
	metaStore MetadataStore,
	embedder Embedder,
	llm LLMClient,
	cfg Config,
) *Manager {
	return &Manager{
		vectorStore: vectorStore,
		metaStore:   metaStore,
		embedder:    embedder,
		llm:         llm,
		config:      cfg,
	}
}

// LLM 返回 LLM 客户端（供外部调用）
func (m *Manager) LLM() LLMClient {
	return m.llm
}

// MetaStore 返回元数据存储（供外部调用）
func (m *Manager) MetaStore() MetadataStore {
	return m.metaStore
}

// CreateMemory 创建记忆（支持三层架构）
// 根据 Layer 和 Type 自动处理不同存储逻辑
func (m *Manager) CreateMemory(ctx context.Context, memory *model.Memory) error {
	// [第1部分] 输入验证
	if memory == nil {
		return fmt.Errorf("memory 不能为 nil")
	}
	if memory.Content == "" {
		return fmt.Errorf("memory.Content 不能为空")
	}
	if memory.UserID == "" {
		return fmt.Errorf("memory.UserID 不能为空")
	}
	if memory.Layer == "" {
		return fmt.Errorf("memory.Layer 不能为空（dialogue/topic/profile）")
	}
	if memory.Type == "" {
		return fmt.Errorf("memory.Type 不能为空")
	}

	// [第2部分] 生成 ID 和时间戳
	memory.ID = uuid.New().String()

	now := time.Now()
	memory.CreatedAt = now
	memory.LastAccessed = now

	// [第2.5部分] 设置默认值（根据层级调整初始重要性）
	if memory.Importance == 0 {
		switch memory.Layer {
		case model.LayerDialogue:
			memory.Importance = 0.6 // 对话层：中等重要性
		case model.LayerTopic:
			memory.Importance = 0.8 // 话题层：较高重要性
		case model.LayerProfile:
			memory.Importance = 1.0 // 画像层：最高重要性
		default:
			memory.Importance = 1.0
		}
	}
	memory.AccessCount = 0

	// [第3部分] 生成 Embedding（带重试机制）
	embedding, err := m.embedWithRetry(ctx, memory.Content, 3)
	if err != nil {
		// 降级处理：Embedding 失败时只存储到元数据库，跳过向量库
		fmt.Printf("⚠️  [降级] 向量生成失败，仅存储元数据: %v\n", err)

		// 设置 Embedding 为 nil（PostgreSQL 允许 NULL）
		memory.Embedding = nil

		// 直接存储到元数据库（无向量）
		if err := m.metaStore.Insert(ctx, memory); err != nil {
			return fmt.Errorf("存储到元数据库失败: %w", err)
		}
		// 成功但无向量，不影响对话流程
		return nil
	}
	memory.Embedding = embedding

	// [第4部分] 存储到向量库
	if err := m.vectorStore.Insert(ctx, memory); err != nil {
		// 向量库失败时也降级：至少保证元数据存储成功
		fmt.Printf("⚠️  [降级] 向量库存储失败，仅存储元数据: %v\n", err)
		if err := m.metaStore.Insert(ctx, memory); err != nil {
			return fmt.Errorf("存储到元数据库失败: %w", err)
		}
		return nil
	}

	// [第5部分] 存储到元数据库
	if err := m.metaStore.Insert(ctx, memory); err != nil {
		return fmt.Errorf("存储到元数据库失败: %w", err)
	}

	// [第6部分] 对话层特殊处理：触发聚合检查
	if memory.Layer == model.LayerDialogue {
		// TODO: 检查是否需要聚合到话题层
		// 获取该 session 的对话轮次数
		// if turnCount >= m.config.AggregationThreshold {
		//     go m.AggregateDialogueToTopic(context.Background(), memory.UserID, sessionID)
		// }
	}

	return nil
}

// SearchMemory 搜索记忆（通用向量检索）
// 跨层级搜索，返回最相关的记忆
func (m *Manager) SearchMemory(ctx context.Context, query string, topK int) ([]*model.Memory, error) {
	// 输入验证
	if query == "" {
		return nil, fmt.Errorf("query 不能为空")
	}
	if topK <= 0 {
		return nil, fmt.Errorf("topK 必须大于 0")
	}
	// 限制 topK 上限
	if topK > 100 {
		topK = 100
	}

	// 生成查询向量
	queryEmbedding, err := m.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %w", err)
	}

	// 向量检索（不使用过滤条件）
	memories, err := m.vectorStore.Search(ctx, queryEmbedding, topK, nil)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	// 异步更新访问信息
	now := time.Now()
	for _, mem := range memories {
		go func(id string) {
			_ = m.metaStore.UpdateAccessInfo(context.Background(), id, now)
		}(mem.ID)
	}

	return memories, nil
}

// ========== Chatbot Intent Memory 召回策略 ==========

// RecallDialogueContext 对话层召回：获取最近 N 轮对话原文
// 对抗：短期遗忘，保持对话连续性
// 使用场景：构建 LLM 的 context window
func (m *Manager) RecallDialogueContext(ctx context.Context, userID string, sessionID string, recentTurns int) ([]*model.Memory, error) {
	// 输入验证
	if userID == "" {
		return nil, fmt.Errorf("userID 不能为空")
	}
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID 不能为空")
	}
	if recentTurns <= 0 {
		return nil, fmt.Errorf("recentTurns 必须大于 0")
	}

	// 调用 MetadataStore 的专用方法获取对话
	// GetDialoguesBySession 应该返回按 turn_number 排序的结果（从旧到新）
	memories, err := m.metaStore.GetDialoguesBySession(ctx, userID, sessionID, recentTurns)
	if err != nil {
		return nil, fmt.Errorf("查询对话记忆失败: %w", err)
	}

	// 更新访问信息（异步）
	now := time.Now()
	for _, mem := range memories {
		go func(id string) {
			// 忽略错误，访问信息更新失败不影响召回
			_ = m.metaStore.UpdateAccessInfo(context.Background(), id, now)
		}(mem.ID)
	}

	return memories, nil
}

// RecallTopicThread 话题层召回：根据当前话题召回相关线索
// 对抗：话题连续性熵增，唤醒历史话题
// 使用场景：用户说"继续刚才的话题"、跨会话话题延续
func (m *Manager) RecallTopicThread(ctx context.Context, userID string, currentQuery string, topK int) ([]*model.Memory, error) {
	// 输入验证
	if userID == "" {
		return nil, fmt.Errorf("userID 不能为空")
	}
	if currentQuery == "" {
		return nil, fmt.Errorf("currentQuery 不能为空")
	}
	if topK <= 0 {
		return nil, fmt.Errorf("topK 必须大于 0")
	}

	// 生成查询向量
	queryEmbedding, err := m.embedder.Embed(ctx, currentQuery)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %w", err)
	}

	// 构造过滤条件：只搜索话题层记忆
	filters := map[string]interface{}{
		"user_id": userID,
		"layer":   model.LayerTopic,
	}

	// 向量检索话题记忆
	memories, err := m.vectorStore.Search(ctx, queryEmbedding, topK, filters)
	if err != nil {
		return nil, fmt.Errorf("话题检索失败: %w", err)
	}

	// 异步更新访问信息
	now := time.Now()
	for _, mem := range memories {
		go func(id string) {
			_ = m.metaStore.UpdateAccessInfo(context.Background(), id, now)
		}(mem.ID)
	}

	// TODO: 可选优化 - 激活父话题和子话题
	// 从 metadata 中提取 parent_topic_id，递归查询

	return memories, nil
}

// RecallUserProfile 画像层召回：快速获取用户画像
// 对抗：人格熵增，稳定 AI 对用户的认知
// 使用场景：对话开始时加载用户画像、意图识别、个性化回复
func (m *Manager) RecallUserProfile(ctx context.Context, userID string, category string) ([]*model.Memory, error) {
	// 输入验证
	if userID == "" {
		return nil, fmt.Errorf("userID 不能为空")
	}

	// 使用 MetadataStore 查询画像层记忆
	// 注意：画像层不需要向量检索，直接按类型查询即可
	var memories []*model.Memory
	var err error

	if category != "" {
		// 按具体类别查询（identity/style/personality/preference）
		var targetType model.MemoryType
		switch category {
		case "identity":
			targetType = model.MemoryTypeUserIdentity
		case "style":
			targetType = model.MemoryTypeCommunicationStyle
		case "personality":
			targetType = model.MemoryTypePersonality
		case "preference":
			targetType = model.MemoryTypePreference
		default:
			return nil, fmt.Errorf("未知的 category: %s（支持：identity/style/personality/preference）", category)
		}
		memories, err = m.metaStore.GetMemoriesByType(ctx, userID, targetType, 50)
	} else {
		// 查询所有画像层记忆
		memories, err = m.metaStore.GetMemoriesByLayer(ctx, userID, model.LayerProfile, 100)
	}

	if err != nil {
		return nil, fmt.Errorf("查询用户画像失败: %w", err)
	}

	// 过滤低置信度画像（confidence < 0.7）
	filteredMemories := make([]*model.Memory, 0)
	for _, mem := range memories {
		if metadata, ok := mem.Metadata["confidence_level"].(float64); ok {
			if metadata >= 0.7 {
				filteredMemories = append(filteredMemories, mem)
			}
		} else {
			// 没有置信度字段的默认保留
			filteredMemories = append(filteredMemories, mem)
		}
	}

	// 异步更新访问信息
	now := time.Now()
	for _, mem := range filteredMemories {
		go func(id string) {
			_ = m.metaStore.UpdateAccessInfo(context.Background(), id, now)
		}(mem.ID)
	}

	return filteredMemories, nil
}

// embedWithRetry Embedding 重试机制
// 用于应对 API 频率限制（Rate Limit）和临时错误
func (m *Manager) embedWithRetry(ctx context.Context, text string, maxRetries int) ([]float32, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		embedding, err := m.embedder.Embed(ctx, text)
		if err == nil {
			return embedding, nil
		}

		lastErr = err

		// 检查错误类型
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "429") {
			// 频率限制或权限错误，采用指数退避策略
			// 基础延迟: 1s, 2s, 4s (更长的等待时间避免持续触发限制)
			baseWait := time.Duration(1<<uint(attempt-1)) * time.Second
			waitTime := baseWait
			if waitTime > 5*time.Second {
				waitTime = 5 * time.Second // 最大等待 5 秒
			}
			fmt.Printf("⚠️  [重试 %d/%d] Embedding API 错误，等待 %v 后重试: %v\n",
				attempt, maxRetries, waitTime, err)
			time.Sleep(waitTime)
			continue
		}

		// 其他错误直接返回，不重试
		return nil, fmt.Errorf("embedding 失败（不可重试）: %w", err)
	}

	return nil, fmt.Errorf("embedding 重试 %d 次后仍失败: %w", maxRetries, lastErr)
}

// HybridRecall 混合召回：根据对话阶段自适应组合三层记忆
// 核心创新：动态熵减策略
// 使用场景：每次对话前的记忆加载
func (m *Manager) HybridRecall(ctx context.Context, req ChatbotRecallRequest) (*ChatbotRecallResult, error) {
	// 输入验证
	if req.UserID == "" {
		return nil, fmt.Errorf("userID 不能为空")
	}
	if req.Query == "" {
		return nil, fmt.Errorf("query 不能为空")
	}

	result := &ChatbotRecallResult{}

	// 根据对话阶段确定召回策略和数量
	var profileLimit, topicLimit, dialogueLimit int
	switch req.DialogStage {
	case "session_start":
		// 新会话开始：重点加载用户画像
		profileLimit = 10
		topicLimit = 3
		dialogueLimit = 2
		result.Strategy = "session_start: 重画像，轻对话"

	case "topic_deepening":
		// 话题深入：重点加载相关话题
		profileLimit = 5
		topicLimit = 8
		dialogueLimit = 5
		result.Strategy = "topic_deepening: 重话题，中画像"

	case "multi_turn":
		// 多轮对话：重点保持对话连续性
		profileLimit = 2
		topicLimit = 3
		dialogueLimit = 10
		result.Strategy = "multi_turn: 重对话，轻画像"

	default:
		// 默认均衡策略
		profileLimit = 5
		topicLimit = 5
		dialogueLimit = 5
		result.Strategy = "default: 均衡召回"
	}

	// 控制总 token 数
	if req.MaxTokens > 0 {
		// 简化：假设每条记忆平均 100 tokens
		maxMemories := req.MaxTokens / 100
		total := profileLimit + topicLimit + dialogueLimit
		if total > maxMemories {
			// 按比例缩减
			scale := float64(maxMemories) / float64(total)
			profileLimit = int(float64(profileLimit) * scale)
			topicLimit = int(float64(topicLimit) * scale)
			dialogueLimit = int(float64(dialogueLimit) * scale)
		}
	}

	// 并发召回三层记忆
	type recallResult struct {
		memories []*model.Memory
		err      error
		layer    string
	}
	resultChan := make(chan recallResult, 3)

	// 1. 画像层召回（不需要 query，直接加载）
	go func() {
		memories, err := m.RecallUserProfile(ctx, req.UserID, "")
		if err != nil {
			resultChan <- recallResult{nil, err, "profile"}
			return
		}
		// 限制数量
		if len(memories) > profileLimit {
			memories = memories[:profileLimit]
		}
		resultChan <- recallResult{memories, nil, "profile"}
	}()

	// 2. 话题层召回（基于 query 的向量检索）
	go func() {
		memories, err := m.RecallTopicThread(ctx, req.UserID, req.Query, topicLimit)
		resultChan <- recallResult{memories, err, "topic"}
	}()

	// 3. 对话层召回（需要 sessionID）
	go func() {
		if req.SessionID == "" {
			resultChan <- recallResult{[]*model.Memory{}, nil, "dialogue"}
			return
		}
		memories, err := m.RecallDialogueContext(ctx, req.UserID, req.SessionID, dialogueLimit)
		resultChan <- recallResult{memories, err, "dialogue"}
	}()

	// 收集结果
	for i := 0; i < 3; i++ {
		res := <-resultChan
		if res.err != nil {
			// 某一层召回失败不中断整体流程，记录错误继续
			// 生产环境应使用 logger
			fmt.Printf("警告：%s 层召回失败: %v\n", res.layer, res.err)
			continue
		}

		switch res.layer {
		case "profile":
			result.ProfileMemories = res.memories
		case "topic":
			result.TopicMemories = res.memories
		case "dialogue":
			result.DialogueMemories = res.memories
		}
	}

	// 计算实际使用的 token 数（简化估算）
	totalMemories := len(result.ProfileMemories) + len(result.TopicMemories) + len(result.DialogueMemories)
	result.TokensUsed = totalMemories * 100 // 假设每条 100 tokens

	return result, nil
}

// ChatbotRecallRequest Chatbot 召回请求
type ChatbotRecallRequest struct {
	UserID      string
	SessionID   string
	Query       string // 当前用户输入
	DialogStage string // "session_start"/"topic_deepening"/"multi_turn"
	MaxTokens   int    // 最大 token 数限制（用于控制召回量）
}

// ChatbotRecallResult Chatbot 召回结果
type ChatbotRecallResult struct {
	DialogueMemories []*model.Memory // 对话层记忆（原文）
	TopicMemories    []*model.Memory // 话题层记忆（聚合摘要）
	ProfileMemories  []*model.Memory // 画像层记忆（用户认知）
	Strategy         string          // 使用的召回策略
	TokensUsed       int             // 实际使用的 token 数
}

// ========== 三层架构维护逻辑 ==========

// ========== 三层架构维护逻辑 ==========
// （已废弃，新实现见文件末尾的 LLM 聚合功能部分）

// CompressMemories 层级迁移压缩 ⭐核心创新点
// Dialogue → Topic → Profile 的渐进式抽象
// 对抗熵增：将高频低价值的对话聚合为结构化记忆
func (m *Manager) CompressMemories(ctx context.Context, userID string) error {
	// TODO: 实现层级迁移逻辑
	// 1. Dialogue → Topic: 聚合超过阈值的对话轮次
	// 2. Topic → Profile: 提取高置信度的用户认知
	// 3. 删除冗余的 Dialogue（保留关键对话）
	panic("not implemented")
}

// DecayMemories 记忆衰减 ⭐核心创新点
// 分层衰减策略：Dialogue 快速衰减，Profile 几乎不衰减
// 公式: new_importance = old_importance * layer_decay_rate * access_boost
func (m *Manager) DecayMemories(ctx context.Context) error {
	// TODO: 实现分层衰减算法
	// Dialogue 层: decay_rate = 0.85 (快速遗忘，对话过期快)
	// Topic 层: decay_rate = 0.95 (中等遗忘，话题有生命周期)
	// Profile 层: decay_rate = 0.99 (长期保留，核心认知稳定)
	// 特殊规则：IsPinned=true 的 Profile 不衰减
	panic("not implemented")
}

// ReflectMemories 记忆反思（可选高级功能）
func (m *Manager) ReflectMemories(ctx context.Context, userID string, timeRange time.Duration) ([]string, error) {
	// TODO: 可选实现
	panic("not implemented")
}

// ========== 依赖接口定义 ==========

// Embedder 将文本转换为向量
type Embedder interface {
	// Embed 单条文本转向量
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch 批量转换
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// VectorStore 向量存储接口
type VectorStore interface {
	// Insert 插入记忆向量
	// 思考：为什么传入完整的 Memory 而不只是 embedding？
	// 答：向量库也需要存储元数据（用于过滤）
	Insert(ctx context.Context, memory *model.Memory) error

	// Search 向量检索
	// 思考：filters 参数有什么用？
	// 答：可以过滤特定用户、特定类型的记忆
	Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]*model.Memory, error)

	// Delete 删除向量
	Delete(ctx context.Context, id string) error
}

// MetadataStore 元数据存储接口（PostgreSQL）
type MetadataStore interface {
	// 基础 CRUD
	Insert(ctx context.Context, memory *model.Memory) error
	Get(ctx context.Context, id string) (*model.Memory, error)
	Update(ctx context.Context, memory *model.Memory) error
	Delete(ctx context.Context, id string) error

	// 记忆管理相关
	CountMemories(ctx context.Context, userID string) (int, error)
	GetOldMemories(ctx context.Context, userID string, before time.Time, limit int) ([]*model.Memory, error)
	UpdateAccessInfo(ctx context.Context, id string, accessTime time.Time) error

	// Chatbot Intent Memory 专用查询
	GetDialoguesBySession(ctx context.Context, userID string, sessionID string, limit int) ([]*model.Memory, error)
	GetMemoriesByLayer(ctx context.Context, userID string, layer model.MemoryLayer, limit int) ([]*model.Memory, error)
	GetMemoriesByType(ctx context.Context, userID string, memoryType model.MemoryType, limit int) ([]*model.Memory, error)

	// LLM 聚合专用查询
	GetBySessionID(ctx context.Context, userID string, sessionID string) ([]*model.Memory, error)
	GetMemoriesByUserAndLayer(ctx context.Context, userID string, layer model.MemoryLayer) ([]*model.Memory, error)
	GetMemory(ctx context.Context, id string) (*model.Memory, error)
}

// LLMClient LLM 客户端接口（用于压缩、反思）
type LLMClient interface {
	// SummarizeDialogues 对话聚合：从多轮对话中提炼话题摘要
	SummarizeDialogues(ctx context.Context, dialogues []*model.Memory) (*TopicSummary, error)

	// ExtractProfile 画像提炼：从多个话题中分析用户特征
	ExtractProfile(ctx context.Context, topics []*model.Memory) (*UserProfile, error)

	// AnalyzeIntent 意图分析：判断用户当前对话意图
	AnalyzeIntent(ctx context.Context, userMessage string) (string, error)
}

// TopicSummary 话题摘要结构
type TopicSummary struct {
	Title       string
	Summary     string
	Keywords    []string
	DialogueIDs []string
}

// UserProfile 用户画像结构
type UserProfile struct {
	Preferences map[string]interface{}
	Habits      map[string]interface{}
	Features    map[string]interface{}
	TopicIDs    []string
}

// ========================================
// 🔄 LLM 聚合功能
// ========================================

// AggregateDialogueToTopic 对话聚合：将一个 session 的对话记忆聚合为话题记忆
// 参数：
//   - ctx: 上下文
//   - userID: 用户 ID
//   - sessionID: 会话 ID（通过 metadata.session_id 筛选）
//
// 返回：创建的话题记忆 ID
func (m *Manager) AggregateDialogueToTopic(ctx context.Context, userID, sessionID string) (string, error) {
	if m.llm == nil {
		return "", fmt.Errorf("LLM 客户端未初始化，无法执行聚合")
	}

	// 1. 查询该 session 的所有对话记忆
	dialogues, err := m.metaStore.GetBySessionID(ctx, userID, sessionID)
	if err != nil {
		return "", fmt.Errorf("查询对话记忆失败: %w", err)
	}

	if len(dialogues) == 0 {
		return "", fmt.Errorf("session %s 没有对话记忆", sessionID)
	}

	// 2. 调用 LLM 聚合
	summary, err := m.llm.SummarizeDialogues(ctx, dialogues)
	if err != nil {
		return "", fmt.Errorf("LLM 聚合失败: %w", err)
	}

	// 3. 创建话题记忆
	topicMemory := &model.Memory{
		UserID:  userID,
		Layer:   model.LayerTopic,
		Type:    model.MemoryTypeTopicThread,
		Content: summary.Summary,
		Metadata: map[string]interface{}{
			"title":        summary.Title,
			"summary":      summary.Summary,
			"keywords":     summary.Keywords,
			"dialogue_ids": summary.DialogueIDs,
			"session_id":   sessionID,
			"source":       "llm_aggregation",
		},
	}

	// 4. 存储话题记忆
	if err := m.CreateMemory(ctx, topicMemory); err != nil {
		return "", fmt.Errorf("创建话题记忆失败: %w", err)
	}

	return topicMemory.ID, nil
}

// ExtractProfileFromTopics 画像提炼：从用户的多个话题中提炼用户画像
// 参数：
//   - ctx: 上下文
//   - userID: 用户 ID
//   - topicIDs: 指定的话题 ID 列表（可选，为空则使用所有话题）
//
// 返回：创建的画像记忆 ID
func (m *Manager) ExtractProfileFromTopics(ctx context.Context, userID string, topicIDs []string) (string, error) {
	if m.llm == nil {
		return "", fmt.Errorf("LLM 客户端未初始化，无法执行提炼")
	}

	var topics []*model.Memory
	var err error

	// 1. 查询话题记忆
	if len(topicIDs) > 0 {
		// 按指定 ID 查询
		for _, id := range topicIDs {
			topic, err := m.metaStore.GetMemory(ctx, id)
			if err != nil {
				continue // 跳过不存在的 ID
			}
			if topic.Layer == model.LayerTopic {
				topics = append(topics, topic)
			}
		}
	} else {
		// 查询该用户的所有话题记忆
		allMemories, err := m.metaStore.GetMemoriesByUserAndLayer(ctx, userID, model.LayerTopic)
		if err != nil {
			return "", fmt.Errorf("查询话题记忆失败: %w", err)
		}
		topics = allMemories
	}

	if len(topics) == 0 {
		return "", fmt.Errorf("用户 %s 没有话题记忆", userID)
	}

	// 2. 调用 LLM 提炼
	profile, err := m.llm.ExtractProfile(ctx, topics)
	if err != nil {
		return "", fmt.Errorf("LLM 提炼失败: %w", err)
	}

	// 3. 生成画像描述文本（用于 Content 字段）
	contentText := fmt.Sprintf("用户画像：%d 个话题分析结果", len(topics))
	if prefs, ok := profile.Preferences["interests"].([]interface{}); ok && len(prefs) > 0 {
		contentText = fmt.Sprintf("用户兴趣：%v", prefs)
	}

	// 4. 创建画像记忆
	profileMemory := &model.Memory{
		UserID:  userID,
		Layer:   model.LayerProfile,
		Type:    model.MemoryTypeUserIdentity,
		Content: contentText,
		Metadata: map[string]interface{}{
			"preferences": profile.Preferences,
			"habits":      profile.Habits,
			"features":    profile.Features,
			"topic_ids":   profile.TopicIDs,
			"source":      "llm_extraction",
		},
	}

	// 5. 存储画像记忆
	if err := m.CreateMemory(ctx, profileMemory); err != nil {
		return "", fmt.Errorf("创建画像记忆失败: %w", err)
	}

	return profileMemory.ID, nil
}

// GetBySessionID 查询指定 session 的对话记忆（需要在 MetadataStore 接口中添加）
// 这个方法应该在 metadata_store.go 中实现，这里先定义接口
