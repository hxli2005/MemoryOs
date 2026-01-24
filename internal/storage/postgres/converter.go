package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"github.com/yourusername/MemoryOs/internal/model"
	"gorm.io/datatypes"
)

// ========== Memory → PO 转换器 ==========

// MemoryToDialoguePO 将 Memory 转换为 DialogueMemoryPO
func MemoryToDialoguePO(m *model.Memory) (*DialogueMemoryPO, error) {
	if m.Layer != model.LayerDialogue {
		return nil, fmt.Errorf("memory layer must be dialogue, got: %s", m.Layer)
	}

	po := &DialogueMemoryPO{
		UserID:      m.UserID,
		Content:     m.Content,
		MemoryType:  string(m.Type),
		Importance:  m.Importance,
		AccessCount: m.AccessCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	// 解析 UUID
	if m.ID != "" {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}
		po.ID = id
	}

	// 转换 Embedding
	if len(m.Embedding) > 0 {
		po.Embedding = pgvector.NewVector(m.Embedding)
	}

	// 转换 Metadata
	if m.Metadata != nil {
		metaJSON, err := json.Marshal(m.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		po.Metadata = datatypes.JSON(metaJSON)

		// 提取特定字段
		if sessionID, ok := m.Metadata["session_id"].(string); ok && sessionID != "" {
			po.SessionID = &sessionID
		}
		if role, ok := m.Metadata["role"].(string); ok && role != "" {
			po.Role = role
		}
	}

	// LastAccessed
	if !m.LastAccessed.IsZero() {
		po.LastAccessed = &m.LastAccessed
	}

	return po, nil
}

// MemoryToTopicPO 将 Memory 转换为 TopicMemoryPO
func MemoryToTopicPO(m *model.Memory) (*TopicMemoryPO, error) {
	if m.Layer != model.LayerTopic {
		return nil, fmt.Errorf("memory layer must be topic, got: %s", m.Layer)
	}

	po := &TopicMemoryPO{
		UserID:      m.UserID,
		Content:     m.Content,
		MemoryType:  string(m.Type),
		Importance:  m.Importance,
		AccessCount: m.AccessCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	// 解析 UUID
	if m.ID != "" {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}
		po.ID = id
	}

	// 转换 Embedding
	if len(m.Embedding) > 0 {
		po.Embedding = pgvector.NewVector(m.Embedding)
	}

	// 转换 Metadata
	if m.Metadata != nil {
		metaJSON, err := json.Marshal(m.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		po.Metadata = datatypes.JSON(metaJSON)

		// 提取特定字段
		if title, ok := m.Metadata["title"].(string); ok && title != "" {
			po.Title = title
		} else {
			// 默认使用 content 前50字符(runes)作为标题，避免 UTF-8 截断
			runes := []rune(m.Content)
			if len(runes) > 50 {
				po.Title = string(runes[:50]) + "..."
			} else {
				po.Title = m.Content
			}
		}

		if summary, ok := m.Metadata["summary"].(string); ok && summary != "" {
			po.Summary = &summary
		}

		if keywords, ok := m.Metadata["keywords"].([]interface{}); ok {
			keywordsJSON, _ := json.Marshal(keywords)
			po.Keywords = datatypes.JSON(keywordsJSON)
		}

		if dialogueIDs, ok := m.Metadata["dialogue_ids"].([]interface{}); ok {
			dialogueIDsJSON, _ := json.Marshal(dialogueIDs)
			po.DialogueIDs = datatypes.JSON(dialogueIDsJSON)
		}
	} else {
		// 无 metadata 时使用默认标题
		runes := []rune(m.Content)
		if len(runes) > 50 {
			po.Title = string(runes[:50]) + "..."
		} else {
			po.Title = m.Content
		}
	}

	// LastAccessed
	if !m.LastAccessed.IsZero() {
		po.LastAccessed = &m.LastAccessed
	}

	return po, nil
}

// MemoryToProfilePO 将 Memory 转换为 ProfileMemoryPO
func MemoryToProfilePO(m *model.Memory) (*ProfileMemoryPO, error) {
	if m.Layer != model.LayerProfile {
		return nil, fmt.Errorf("memory layer must be profile, got: %s", m.Layer)
	}

	po := &ProfileMemoryPO{
		UserID:      m.UserID,
		Content:     m.Content, // 添加 Content 字段映射
		MemoryType:  string(m.Type),
		Importance:  m.Importance,
		AccessCount: m.AccessCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}

	// 解析 UUID
	if m.ID != "" {
		id, err := uuid.Parse(m.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID: %w", err)
		}
		po.ID = id
	}

	// 转换 Embedding
	if len(m.Embedding) > 0 {
		po.Embedding = pgvector.NewVector(m.Embedding)
	}

	// 转换 Metadata
	if m.Metadata != nil {
		metaJSON, err := json.Marshal(m.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		po.Metadata = datatypes.JSON(metaJSON)

		// 提取特定字段
		if preferences, ok := m.Metadata["preferences"].(map[string]interface{}); ok {
			prefJSON, _ := json.Marshal(preferences)
			po.Preferences = datatypes.JSON(prefJSON)
		}

		if habits, ok := m.Metadata["habits"].(map[string]interface{}); ok {
			habitsJSON, _ := json.Marshal(habits)
			po.Habits = datatypes.JSON(habitsJSON)
		}

		if features, ok := m.Metadata["features"].(map[string]interface{}); ok {
			featuresJSON, _ := json.Marshal(features)
			po.Features = datatypes.JSON(featuresJSON)
		}

		if topicIDs, ok := m.Metadata["topic_ids"].([]interface{}); ok {
			topicIDsJSON, _ := json.Marshal(topicIDs)
			po.TopicIDs = datatypes.JSON(topicIDsJSON)
		}
	}

	// LastAccessed
	if !m.LastAccessed.IsZero() {
		po.LastAccessed = &m.LastAccessed
	}

	return po, nil
}

// ========== PO → Memory 转换器 ==========

// DialoguePOToMemory 将 DialogueMemoryPO 转换为 Memory
func DialoguePOToMemory(po *DialogueMemoryPO) (*model.Memory, error) {
	m := &model.Memory{
		ID:          po.ID.String(),
		UserID:      po.UserID,
		Layer:       model.LayerDialogue,
		Type:        model.MemoryType(po.MemoryType),
		Content:     po.Content,
		Importance:  po.Importance,
		AccessCount: po.AccessCount,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}

	// 转换 Embedding
	embSlice := po.Embedding.Slice()
	if len(embSlice) > 0 {
		m.Embedding = embSlice
	}

	// 转换 Metadata
	if len(po.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(po.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		m.Metadata = metadata
	} else {
		m.Metadata = make(map[string]interface{})
	}

	// 补充字段到 Metadata
	if po.SessionID != nil {
		m.Metadata["session_id"] = *po.SessionID
	}
	if po.Role != "" {
		m.Metadata["role"] = po.Role
	}

	// LastAccessed
	if po.LastAccessed != nil {
		m.LastAccessed = *po.LastAccessed
	}

	return m, nil
}

// TopicPOToMemory 将 TopicMemoryPO 转换为 Memory
func TopicPOToMemory(po *TopicMemoryPO) (*model.Memory, error) {
	m := &model.Memory{
		ID:          po.ID.String(),
		UserID:      po.UserID,
		Layer:       model.LayerTopic,
		Type:        model.MemoryType(po.MemoryType),
		Content:     po.Content,
		Importance:  po.Importance,
		AccessCount: po.AccessCount,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}

	// 转换 Embedding
	embSlice := po.Embedding.Slice()
	if len(embSlice) > 0 {
		m.Embedding = embSlice
	}

	// 转换 Metadata
	if len(po.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(po.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		m.Metadata = metadata
	} else {
		m.Metadata = make(map[string]interface{})
	}

	// 补充字段到 Metadata
	m.Metadata["title"] = po.Title
	if po.Summary != nil {
		m.Metadata["summary"] = *po.Summary
	}
	if len(po.Keywords) > 0 {
		var keywords []interface{}
		json.Unmarshal(po.Keywords, &keywords)
		m.Metadata["keywords"] = keywords
	}
	if len(po.DialogueIDs) > 0 {
		var dialogueIDs []interface{}
		json.Unmarshal(po.DialogueIDs, &dialogueIDs)
		m.Metadata["dialogue_ids"] = dialogueIDs
	}

	// LastAccessed
	if po.LastAccessed != nil {
		m.LastAccessed = *po.LastAccessed
	}

	return m, nil
}

// ProfilePOToMemory 将 ProfileMemoryPO 转换为 Memory
func ProfilePOToMemory(po *ProfileMemoryPO) (*model.Memory, error) {
	m := &model.Memory{
		ID:          po.ID.String(),
		UserID:      po.UserID,
		Layer:       model.LayerProfile,
		Type:        model.MemoryType(po.MemoryType),
		Content:     "", // Profile 层通常没有单独的 content 字段
		Importance:  po.Importance,
		AccessCount: po.AccessCount,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}

	// 转换 Embedding
	embSlice := po.Embedding.Slice()
	if len(embSlice) > 0 {
		m.Embedding = embSlice
	}

	// 转换 Metadata
	if len(po.Metadata) > 0 {
		var metadata map[string]interface{}
		if err := json.Unmarshal(po.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		m.Metadata = metadata
	} else {
		m.Metadata = make(map[string]interface{})
	}

	// 补充结构化字段到 Metadata
	if len(po.Preferences) > 0 {
		var preferences map[string]interface{}
		json.Unmarshal(po.Preferences, &preferences)
		m.Metadata["preferences"] = preferences
	}
	if len(po.Habits) > 0 {
		var habits map[string]interface{}
		json.Unmarshal(po.Habits, &habits)
		m.Metadata["habits"] = habits
	}
	if len(po.Features) > 0 {
		var features map[string]interface{}
		json.Unmarshal(po.Features, &features)
		m.Metadata["features"] = features
	}
	if len(po.TopicIDs) > 0 {
		var topicIDs []interface{}
		json.Unmarshal(po.TopicIDs, &topicIDs)
		m.Metadata["topic_ids"] = topicIDs
	}

	// Profile 层的 Content 通常是偏好/习惯的文本描述
	if len(po.Preferences) > 0 || len(po.Habits) > 0 || len(po.Features) > 0 {
		contentJSON, _ := json.Marshal(m.Metadata)
		m.Content = string(contentJSON)
	}

	// LastAccessed
	if po.LastAccessed != nil {
		m.LastAccessed = *po.LastAccessed
	}

	return m, nil
}
