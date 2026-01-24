package mock

import (
	"context"
	"time"

	"github.com/yourusername/MemoryOs/internal/model"
)

// MockVectorStore 临时的向量存储实现
type MockVectorStore struct{}

func NewMockVectorStore() *MockVectorStore {
	return &MockVectorStore{}
}

func (m *MockVectorStore) Insert(ctx context.Context, memory *model.Memory) error {
	// TODO: 后续替换为真实的 Milvus 实现
	return nil
}

func (m *MockVectorStore) Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]*model.Memory, error) {
	// TODO: 后续替换为真实实现
	return []*model.Memory{}, nil
}

func (m *MockVectorStore) Delete(ctx context.Context, id string) error {
	return nil
}

// MockMetadataStore 临时的元数据存储实现
type MockMetadataStore struct{}

func NewMockMetadataStore() *MockMetadataStore {
	return &MockMetadataStore{}
}

func (m *MockMetadataStore) Insert(ctx context.Context, memory *model.Memory) error {
	return nil
}

func (m *MockMetadataStore) Get(ctx context.Context, id string) (*model.Memory, error) {
	return nil, nil
}

func (m *MockMetadataStore) Update(ctx context.Context, memory *model.Memory) error {
	return nil
}

func (m *MockMetadataStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockMetadataStore) CountMemories(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (m *MockMetadataStore) GetOldMemories(ctx context.Context, userID string, before time.Time, limit int) ([]*model.Memory, error) {
	return []*model.Memory{}, nil
}

// LLM 聚合专用方法
func (m *MockMetadataStore) GetBySessionID(ctx context.Context, userID string, sessionID string) ([]*model.Memory, error) {
	return []*model.Memory{}, nil
}

func (m *MockMetadataStore) GetMemoriesByUserAndLayer(ctx context.Context, userID string, layer model.MemoryLayer) ([]*model.Memory, error) {
	return []*model.Memory{}, nil
}

func (m *MockMetadataStore) GetMemory(ctx context.Context, id string) (*model.Memory, error) {
	return nil, nil
}
func (m *MockMetadataStore) UpdateAccessInfo(ctx context.Context, id string, accessTime time.Time) error {
	return nil
}

// Chatbot Intent Memory 专用方法
func (m *MockMetadataStore) GetDialoguesBySession(ctx context.Context, userID string, sessionID string, limit int) ([]*model.Memory, error) {
	// Mock 实现：返回空列表
	// 真实实现应该查询 PostgreSQL: WHERE user_id = ? AND layer = 'dialogue' AND metadata->>'session_id' = ?
	// ORDER BY metadata->>'turn_number' ASC LIMIT ?
	return []*model.Memory{}, nil
}

func (m *MockMetadataStore) GetMemoriesByLayer(ctx context.Context, userID string, layer model.MemoryLayer, limit int) ([]*model.Memory, error) {
	// Mock 实现：返回空列表
	return []*model.Memory{}, nil
}

func (m *MockMetadataStore) GetMemoriesByType(ctx context.Context, userID string, memoryType model.MemoryType, limit int) ([]*model.Memory, error) {
	// Mock 实现：返回空列表
	return []*model.Memory{}, nil
}
