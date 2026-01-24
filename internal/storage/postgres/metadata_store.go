package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/MemoryOs/internal/model"
	"gorm.io/gorm"
)

// MetadataStore PostgreSQL 元数据存储实现
type MetadataStore struct {
	db *gorm.DB
}

// NewMetadataStore 创建 PostgreSQL 元数据存储
func NewMetadataStore(db *gorm.DB) *MetadataStore {
	return &MetadataStore{db: db}
}

// ========== 基础 CRUD ==========

// Insert 插入记忆到数据库
// 根据 memory.Layer 路由到不同的表
func (s *MetadataStore) Insert(ctx context.Context, memory *model.Memory) error {
	if memory == nil {
		return fmt.Errorf("memory cannot be nil")
	}

	// 根据层级选择不同的转换器和表
	switch memory.Layer {
	case model.LayerDialogue:
		po, err := MemoryToDialoguePO(memory)
		if err != nil {
			return fmt.Errorf("convert to dialogue PO failed: %w", err)
		}
		return s.db.WithContext(ctx).Create(po).Error

	case model.LayerTopic:
		po, err := MemoryToTopicPO(memory)
		if err != nil {
			return fmt.Errorf("convert to topic PO failed: %w", err)
		}
		return s.db.WithContext(ctx).Create(po).Error

	case model.LayerProfile:
		po, err := MemoryToProfilePO(memory)
		if err != nil {
			return fmt.Errorf("convert to profile PO failed: %w", err)
		}
		return s.db.WithContext(ctx).Create(po).Error

	default:
		return fmt.Errorf("unknown memory layer: %s", memory.Layer)
	}
}

// Get 根据 ID 获取记忆
// 策略：依次查询三张表（后续可优化为按 ID 前缀路由）
func (s *MetadataStore) Get(ctx context.Context, id string) (*model.Memory, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// 尝试从 dialogue_memory 查询
	var dialoguePO DialogueMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&dialoguePO).Error; err == nil {
		return DialoguePOToMemory(&dialoguePO)
	}

	// 尝试从 topic_memory 查询
	var topicPO TopicMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&topicPO).Error; err == nil {
		return TopicPOToMemory(&topicPO)
	}

	// 尝试从 profile_memory 查询
	var profilePO ProfileMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&profilePO).Error; err == nil {
		return ProfilePOToMemory(&profilePO)
	}

	return nil, fmt.Errorf("memory not found: %s", id)
}

// Update 更新记忆
// 根据 memory.Layer 路由到不同的表
func (s *MetadataStore) Update(ctx context.Context, memory *model.Memory) error {
	if memory == nil {
		return fmt.Errorf("memory cannot be nil")
	}
	if memory.ID == "" {
		return fmt.Errorf("memory ID cannot be empty")
	}

	uid, err := uuid.Parse(memory.ID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	switch memory.Layer {
	case model.LayerDialogue:
		po, err := MemoryToDialoguePO(memory)
		if err != nil {
			return fmt.Errorf("convert to dialogue PO failed: %w", err)
		}
		// 使用 Save 进行全量更新
		return s.db.WithContext(ctx).Model(&DialogueMemoryPO{}).Where("id = ?", uid).Updates(po).Error

	case model.LayerTopic:
		po, err := MemoryToTopicPO(memory)
		if err != nil {
			return fmt.Errorf("convert to topic PO failed: %w", err)
		}
		return s.db.WithContext(ctx).Model(&TopicMemoryPO{}).Where("id = ?", uid).Updates(po).Error

	case model.LayerProfile:
		po, err := MemoryToProfilePO(memory)
		if err != nil {
			return fmt.Errorf("convert to profile PO failed: %w", err)
		}
		return s.db.WithContext(ctx).Model(&ProfileMemoryPO{}).Where("id = ?", uid).Updates(po).Error

	default:
		return fmt.Errorf("unknown memory layer: %s", memory.Layer)
	}
}

// Delete 删除记忆
// 策略：依次尝试删除三张表中的记录
func (s *MetadataStore) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// 尝试从三张表删除（只有一张表会成功）
	tx := s.db.WithContext(ctx)

	result := tx.Where("id = ?", uid).Delete(&DialogueMemoryPO{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	result = tx.Where("id = ?", uid).Delete(&TopicMemoryPO{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	result = tx.Where("id = ?", uid).Delete(&ProfileMemoryPO{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("memory not found: %s", id)
}

// ========== 记忆管理相关 ==========

// CountMemories 统计用户的记忆总数（跨三层）
func (s *MetadataStore) CountMemories(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, fmt.Errorf("userID cannot be empty")
	}

	var dialogueCount, topicCount, profileCount int64

	if err := s.db.WithContext(ctx).Model(&DialogueMemoryPO{}).Where("user_id = ?", userID).Count(&dialogueCount).Error; err != nil {
		return 0, err
	}

	if err := s.db.WithContext(ctx).Model(&TopicMemoryPO{}).Where("user_id = ?", userID).Count(&topicCount).Error; err != nil {
		return 0, err
	}

	if err := s.db.WithContext(ctx).Model(&ProfileMemoryPO{}).Where("user_id = ?", userID).Count(&profileCount).Error; err != nil {
		return 0, err
	}

	return int(dialogueCount + topicCount + profileCount), nil
}

// GetOldMemories 获取指定时间之前的旧记忆（用于衰减/清理）
// 按创建时间排序，返回最旧的记忆
func (s *MetadataStore) GetOldMemories(ctx context.Context, userID string, before time.Time, limit int) ([]*model.Memory, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if limit <= 0 {
		limit = 100 // 默认限制
	}

	var memories []*model.Memory

	// 从 dialogue_memory 查询
	var dialoguePOs []DialogueMemoryPO
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND created_at < ?", userID, before).
		Order("created_at ASC").
		Limit(limit).
		Find(&dialoguePOs).Error; err != nil {
		return nil, err
	}
	for _, po := range dialoguePOs {
		m, err := DialoguePOToMemory(&po)
		if err != nil {
			continue // 跳过转换失败的记录
		}
		memories = append(memories, m)
	}

	// 从 topic_memory 查询
	var topicPOs []TopicMemoryPO
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND created_at < ?", userID, before).
		Order("created_at ASC").
		Limit(limit - len(memories)).
		Find(&topicPOs).Error; err != nil {
		return nil, err
	}
	for _, po := range topicPOs {
		m, err := TopicPOToMemory(&po)
		if err != nil {
			continue
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// UpdateAccessInfo 更新记忆的访问信息（访问次数 +1，更新访问时间）
func (s *MetadataStore) UpdateAccessInfo(ctx context.Context, id string, accessTime time.Time) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	// 更新访问信息的 SQL
	updateData := map[string]interface{}{
		"access_count":  gorm.Expr("access_count + ?", 1),
		"last_accessed": accessTime,
	}

	// 尝试更新三张表（只有一张会成功）
	tx := s.db.WithContext(ctx)

	result := tx.Model(&DialogueMemoryPO{}).Where("id = ?", uid).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	result = tx.Model(&TopicMemoryPO{}).Where("id = ?", uid).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	result = tx.Model(&ProfileMemoryPO{}).Where("id = ?", uid).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("memory not found: %s", id)
}

// ========== Chatbot Intent Memory 专用查询 ==========

// GetDialoguesBySession 获取会话中的对话记忆
// 按创建时间正序排列（对话顺序）
func (s *MetadataStore) GetDialoguesBySession(ctx context.Context, userID string, sessionID string, limit int) ([]*model.Memory, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}
	if limit <= 0 {
		limit = 50 // 默认返回最近 50 轮对话
	}

	var dialoguePOs []DialogueMemoryPO
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND session_id = ?", userID, sessionID).
		Order("created_at ASC"). // 对话按时间正序
		Limit(limit).
		Find(&dialoguePOs).Error; err != nil {
		return nil, err
	}

	memories := make([]*model.Memory, 0, len(dialoguePOs))
	for _, po := range dialoguePOs {
		m, err := DialoguePOToMemory(&po)
		if err != nil {
			continue // 跳过转换失败的记录
		}
		memories = append(memories, m)
	}

	return memories, nil
}

// GetMemoriesByLayer 获取指定层级的记忆
// 按创建时间倒序排列（最新的在前）
func (s *MetadataStore) GetMemoriesByLayer(ctx context.Context, userID string, layer model.MemoryLayer, limit int) ([]*model.Memory, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if limit <= 0 {
		limit = 100
	}

	var memories []*model.Memory

	switch layer {
	case model.LayerDialogue:
		var dialoguePOs []DialogueMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(limit).
			Find(&dialoguePOs).Error; err != nil {
			return nil, err
		}
		for _, po := range dialoguePOs {
			m, err := DialoguePOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}

	case model.LayerTopic:
		var topicPOs []TopicMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(limit).
			Find(&topicPOs).Error; err != nil {
			return nil, err
		}
		for _, po := range topicPOs {
			m, err := TopicPOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}

	case model.LayerProfile:
		var profilePOs []ProfileMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Limit(limit).
			Find(&profilePOs).Error; err != nil {
			return nil, err
		}
		for _, po := range profilePOs {
			m, err := ProfilePOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}

	default:
		return nil, fmt.Errorf("unknown memory layer: %s", layer)
	}

	return memories, nil
}

// GetMemoriesByType 获取指定类型的记忆
// 按创建时间倒序排列
func (s *MetadataStore) GetMemoriesByType(ctx context.Context, userID string, memoryType model.MemoryType, limit int) ([]*model.Memory, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	if limit <= 0 {
		limit = 100
	}

	var memories []*model.Memory

	// 根据 MemoryType 判断应该查询哪个表
	// 对话层类型
	if memoryType == model.MemoryTypeUserMessage ||
		memoryType == model.MemoryTypeAssistantMessage ||
		memoryType == model.MemoryTypeDialogueContext {
		var dialoguePOs []DialogueMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND memory_type = ?", userID, string(memoryType)).
			Order("created_at DESC").
			Limit(limit).
			Find(&dialoguePOs).Error; err != nil {
			return nil, err
		}
		for _, po := range dialoguePOs {
			m, err := DialoguePOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}
		return memories, nil
	}

	// 话题层类型
	if memoryType == model.MemoryTypeTopicThread ||
		memoryType == model.MemoryTypeIntent ||
		memoryType == model.MemoryTypeConversationFlow {
		var topicPOs []TopicMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND memory_type = ?", userID, string(memoryType)).
			Order("created_at DESC").
			Limit(limit).
			Find(&topicPOs).Error; err != nil {
			return nil, err
		}
		for _, po := range topicPOs {
			m, err := TopicPOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}
		return memories, nil
	}

	// 画像层类型
	if memoryType == model.MemoryTypeUserIdentity ||
		memoryType == model.MemoryTypeCommunicationStyle ||
		memoryType == model.MemoryTypePersonality ||
		memoryType == model.MemoryTypePreference {
		var profilePOs []ProfileMemoryPO
		if err := s.db.WithContext(ctx).
			Where("user_id = ? AND memory_type = ?", userID, string(memoryType)).
			Order("created_at DESC").
			Limit(limit).
			Find(&profilePOs).Error; err != nil {
			return nil, err
		}
		for _, po := range profilePOs {
			m, err := ProfilePOToMemory(&po)
			if err != nil {
				continue
			}
			memories = append(memories, m)
		}
		return memories, nil
	}

	return nil, fmt.Errorf("unknown memory type: %s", memoryType)
}

// ========== LLM 聚合专用查询 ==========

// GetBySessionID 查询指定 session 的所有对话记忆（用于聚合）
func (s *MetadataStore) GetBySessionID(ctx context.Context, userID string, sessionID string) ([]*model.Memory, error) {
	var dialogues []DialogueMemoryPO

	// 查询条件：user_id + metadata->>'session_id'
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("metadata->>'session_id' = ?", sessionID).
		Order("created_at ASC").
		Find(&dialogues).Error

	if err != nil {
		return nil, fmt.Errorf("query by session_id failed: %w", err)
	}

	// 转换为 Memory 对象
	var memories []*model.Memory
	for _, po := range dialogues {
		mem, err := DialoguePOToMemory(&po)
		if err != nil {
			continue
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

// GetMemoriesByUserAndLayer 查询用户在指定层级的所有记忆（用于提炼）
func (s *MetadataStore) GetMemoriesByUserAndLayer(ctx context.Context, userID string, layer model.MemoryLayer) ([]*model.Memory, error) {
	var memories []*model.Memory

	switch layer {
	case model.LayerDialogue:
		var dialogues []DialogueMemoryPO
		err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Find(&dialogues).Error
		if err != nil {
			return nil, err
		}
		for _, po := range dialogues {
			if mem, err := DialoguePOToMemory(&po); err == nil {
				memories = append(memories, mem)
			}
		}

	case model.LayerTopic:
		var topics []TopicMemoryPO
		err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Find(&topics).Error
		if err != nil {
			return nil, err
		}
		for _, po := range topics {
			if mem, err := TopicPOToMemory(&po); err == nil {
				memories = append(memories, mem)
			}
		}

	case model.LayerProfile:
		var profiles []ProfileMemoryPO
		err := s.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Order("created_at DESC").
			Find(&profiles).Error
		if err != nil {
			return nil, err
		}
		for _, po := range profiles {
			if mem, err := ProfilePOToMemory(&po); err == nil {
				memories = append(memories, mem)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported layer: %s", layer)
	}

	return memories, nil
}

// GetMemory 根据 ID 查询单个记忆（通用查询，自动识别层级）
func (s *MetadataStore) GetMemory(ctx context.Context, id string) (*model.Memory, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// 尝试从三个表中查询（按概率从对话层开始）
	// 1. 对话层
	var dialogue DialogueMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&dialogue).Error; err == nil {
		return DialoguePOToMemory(&dialogue)
	}

	// 2. 话题层
	var topic TopicMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&topic).Error; err == nil {
		return TopicPOToMemory(&topic)
	}

	// 3. 画像层
	var profile ProfileMemoryPO
	if err := s.db.WithContext(ctx).Where("id = ?", uid).First(&profile).Error; err == nil {
		return ProfilePOToMemory(&profile)
	}

	return nil, fmt.Errorf("memory not found: %s", id)
}
