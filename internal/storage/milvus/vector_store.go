package milvus

import (
	"context"
	"fmt"
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/yourusername/MemoryOs/internal/model"
)

// VectorStore Milvus 向量存储实现
type VectorStore struct {
	client         client.Client
	collectionName string
	dimension      int
}

// Config Milvus 配置
type Config struct {
	Host           string
	Port           int
	CollectionName string
	Dimension      int
}

// NewVectorStore 创建 Milvus VectorStore
func NewVectorStore(cfg Config) (*VectorStore, error) {
	ctx := context.Background()

	// 1. 连接 Milvus
	c, err := client.NewClient(ctx, client.Config{
		Address: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	})
	if err != nil {
		return nil, fmt.Errorf("连接 Milvus 失败: %w", err)
	}

	vs := &VectorStore{
		client:         c,
		collectionName: cfg.CollectionName,
		dimension:      cfg.Dimension,
	}

	// 2. 初始化 Collection
	if err := vs.initCollection(ctx); err != nil {
		return nil, fmt.Errorf("初始化 Collection 失败: %w", err)
	}

	log.Printf("✅ Milvus VectorStore 初始化成功: %s (维度: %d)", cfg.CollectionName, cfg.Dimension)
	return vs, nil
}

// initCollection 初始化 Collection（如果不存在则创建）
func (vs *VectorStore) initCollection(ctx context.Context) error {
	// 1. 检查 Collection 是否存在
	hasCollection, err := vs.client.HasCollection(ctx, vs.collectionName)
	if err != nil {
		return fmt.Errorf("检查 Collection 失败: %w", err)
	}

	if hasCollection {
		log.Printf("📦 Collection '%s' 已存在，加载中...", vs.collectionName)
		// 加载到内存（必须加载才能搜索）
		if err := vs.client.LoadCollection(ctx, vs.collectionName, false); err != nil {
			return fmt.Errorf("加载 Collection 失败: %w", err)
		}
		return nil
	}

	// 2. 创建 Schema
	schema := &entity.Schema{
		CollectionName: vs.collectionName,
		Description:    "MemoryOS 记忆向量存储",
		Fields: []*entity.Field{
			// Primary Key
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "36", // UUID 长度
				},
			},
			// 向量字段
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: fmt.Sprintf("%d", vs.dimension),
				},
			},
			// 元数据字段（用于过滤）
			{
				Name:     "user_id",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "100",
				},
			},
			{
				Name:     "layer",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "20",
				},
			},
			{
				Name:     "memory_type",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "50",
				},
			},
		},
	}

	// 3. 创建 Collection
	if err := vs.client.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("创建 Collection 失败: %w", err)
	}

	log.Printf("📦 Collection '%s' 创建成功", vs.collectionName)

	// 4. 创建索引（HNSW - 高性能近似最近邻）
	index, err := entity.NewIndexHNSW(entity.L2, 16, 200) // M=16, efConstruction=200
	if err != nil {
		return fmt.Errorf("创建索引配置失败: %w", err)
	}

	if err := vs.client.CreateIndex(ctx, vs.collectionName, "embedding", index, false); err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}

	log.Printf("🔍 索引创建成功: HNSW (M=16, ef=200)")

	// 5. 加载到内存
	if err := vs.client.LoadCollection(ctx, vs.collectionName, false); err != nil {
		return fmt.Errorf("加载 Collection 失败: %w", err)
	}

	return nil
}

// Insert 插入记忆向量
func (vs *VectorStore) Insert(ctx context.Context, memory *model.Memory) error {
	if memory.ID == "" {
		return fmt.Errorf("memory.ID 不能为空")
	}
	if len(memory.Embedding) == 0 {
		return fmt.Errorf("memory.Embedding 不能为空")
	}
	if len(memory.Embedding) != vs.dimension {
		return fmt.Errorf("embedding 维度不匹配: 期望 %d, 实际 %d", vs.dimension, len(memory.Embedding))
	}

	// 构造数据列
	idColumn := entity.NewColumnVarChar("id", []string{memory.ID})
	embeddingColumn := entity.NewColumnFloatVector("embedding", vs.dimension, [][]float32{memory.Embedding})
	userIDColumn := entity.NewColumnVarChar("user_id", []string{memory.UserID})
	layerColumn := entity.NewColumnVarChar("layer", []string{string(memory.Layer)})
	typeColumn := entity.NewColumnVarChar("memory_type", []string{string(memory.Type)})

	// 插入数据
	_, err := vs.client.Insert(ctx, vs.collectionName, "",
		idColumn, embeddingColumn, userIDColumn, layerColumn, typeColumn,
	)
	if err != nil {
		return fmt.Errorf("插入向量失败: %w", err)
	}

	// Flush（确保数据持久化）
	if err := vs.client.Flush(ctx, vs.collectionName, false); err != nil {
		return fmt.Errorf("Flush 失败: %w", err)
	}

	return nil
}

// Search 向量检索
func (vs *VectorStore) Search(ctx context.Context, embedding []float32, topK int, filters map[string]interface{}) ([]*model.Memory, error) {
	if len(embedding) != vs.dimension {
		return nil, fmt.Errorf("embedding 维度不匹配: 期望 %d, 实际 %d", vs.dimension, len(embedding))
	}

	// 构建过滤表达式
	filterExpr := vs.buildFilterExpression(filters)

	// 构造搜索向量
	searchVectors := []entity.Vector{
		entity.FloatVector(embedding),
	}

	// 搜索参数
	sp, _ := entity.NewIndexHNSWSearchParam(100) // ef=100

	// 执行搜索
	results, err := vs.client.Search(
		ctx,
		vs.collectionName,
		nil, // partitionNames
		filterExpr,
		[]string{"user_id", "layer", "memory_type"}, // 输出字段
		searchVectors,
		"embedding",
		entity.L2, // 距离度量（L2 欧氏距离）
		topK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("向量搜索失败: %w", err)
	}

	if len(results) == 0 {
		return []*model.Memory{}, nil
	}

	// 解析结果
	memories := make([]*model.Memory, 0, topK)
	for i := 0; i < results[0].ResultCount; i++ {
		// 获取 ID
		id, err := results[0].IDs.GetAsString(i)
		if err != nil {
			log.Printf("⚠️ 获取 ID 失败: %v", err)
			continue
		}

		// 获取元数据字段
		userIDCol := results[0].Fields.GetColumn("user_id")
		layerCol := results[0].Fields.GetColumn("layer")
		typeCol := results[0].Fields.GetColumn("memory_type")

		var userID, layer, memoryType string
		if userIDCol != nil {
			if vc, ok := userIDCol.(*entity.ColumnVarChar); ok {
				userID, _ = vc.ValueByIdx(i)
			}
		}
		if layerCol != nil {
			if vc, ok := layerCol.(*entity.ColumnVarChar); ok {
				layer, _ = vc.ValueByIdx(i)
			}
		}
		if typeCol != nil {
			if vc, ok := typeCol.(*entity.ColumnVarChar); ok {
				memoryType, _ = vc.ValueByIdx(i)
			}
		}

		// 获取相似度分数
		score := results[0].Scores[i]

		memories = append(memories, &model.Memory{
			ID:     id,
			UserID: userID,
			Layer:  model.MemoryLayer(layer),
			Type:   model.MemoryType(memoryType),
			Metadata: map[string]interface{}{
				"similarity_score": score,
			},
		})
	}

	return memories, nil
}

// Delete 删除向量
func (vs *VectorStore) Delete(ctx context.Context, id string) error {
	// 构造删除表达式
	expr := fmt.Sprintf("id == \"%s\"", id)

	if err := vs.client.Delete(ctx, vs.collectionName, "", expr); err != nil {
		return fmt.Errorf("删除向量失败: %w", err)
	}

	// Flush
	if err := vs.client.Flush(ctx, vs.collectionName, false); err != nil {
		return fmt.Errorf("Flush 失败: %w", err)
	}

	return nil
}

// Close 关闭连接
func (vs *VectorStore) Close() error {
	return vs.client.Close()
}

// buildFilterExpression 构建过滤表达式
func (vs *VectorStore) buildFilterExpression(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	var expressions []string
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			expressions = append(expressions, fmt.Sprintf("%s == \"%s\"", key, v))
		case int, int64:
			expressions = append(expressions, fmt.Sprintf("%s == %d", key, v))
		case []string:
			// IN 查询
			if len(v) > 0 {
				inValues := ""
				for i, val := range v {
					if i > 0 {
						inValues += ", "
					}
					inValues += fmt.Sprintf("\"%s\"", val)
				}
				expressions = append(expressions, fmt.Sprintf("%s in [%s]", key, inValues))
			}
		}
	}

	if len(expressions) == 0 {
		return ""
	}

	// 用 AND 连接
	result := expressions[0]
	for i := 1; i < len(expressions); i++ {
		result += " && " + expressions[i]
	}
	return result
}
