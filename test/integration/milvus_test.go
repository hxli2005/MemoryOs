package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/MemoryOs/internal/bootstrap"
	"github.com/yourusername/MemoryOs/internal/model"
)

func main() {
	ctx := context.Background()

	fmt.Println("========================================")
	fmt.Println("Milvus VectorStore 集成测试")
	fmt.Println("========================================\n")

	// 1. 初始化应用（会尝试连接 Milvus）
	fmt.Println("【步骤 1】初始化应用...")
	app, err := bootstrap.Initialize("config/config.yaml")
	if err != nil {
		log.Fatal("❌ 初始化失败:", err)
	}
	fmt.Println("✅ 应用初始化成功\n")

	// 2. 创建测试记忆
	fmt.Println("【步骤 2】创建测试记忆...")
	userID := "test_user_" + time.Now().Format("150405")

	memories := []*model.Memory{
		{
			UserID:  userID,
			Layer:   model.LayerDialogue,
			Type:    model.MemoryTypeUserMessage,
			Content: "今天天气怎么样？",
			Metadata: map[string]interface{}{
				"dialogue_turn": 1,
			},
		},
		{
			UserID:  userID,
			Layer:   model.LayerDialogue,
			Type:    model.MemoryTypeAssistantMessage,
			Content: "今天天气晴朗，温度适宜，是个出门的好日子！",
			Metadata: map[string]interface{}{
				"dialogue_turn": 2,
			},
		},
		{
			UserID:  userID,
			Layer:   model.LayerTopic,
			Type:    model.MemoryTypeTopicThread,
			Content: "用户关心天气信息，可能计划外出活动",
			Metadata: map[string]interface{}{
				"topic": "weather",
			},
		},
	}

	for i, mem := range memories {
		if err := app.MemoryManager.CreateMemory(ctx, mem); err != nil {
			log.Fatalf("❌ 创建记忆 %d 失败: %v", i+1, err)
		}
		fmt.Printf("✅ 记忆 %d 创建成功\n", i+1)
		fmt.Printf("   ID: %s\n", mem.ID)
		fmt.Printf("   Layer: %s\n", mem.Layer)
		fmt.Printf("   Embedding 维度: %d\n", len(mem.Embedding))
	}
	fmt.Println()

	// 3. 向量搜索测试
	fmt.Println("【步骤 3】测试向量搜索...")
	query := "天气情况"
	results, err := app.MemoryManager.SearchMemory(ctx, query, 5)
	if err != nil {
		log.Fatal("❌ 向量搜索失败:", err)
	}
	fmt.Printf("✅ 搜索成功，找到 %d 条记忆\n", len(results))
	for i, r := range results {
		fmt.Printf("   [%d] %s (Layer: %s)\n", i+1, truncate(r.Content, 40), r.Layer)
		if score, ok := r.Metadata["similarity_score"].(float32); ok {
			fmt.Printf("       相似度: %.4f\n", score)
		}
	}
	fmt.Println()

	// 4. 按层级过滤搜索
	fmt.Println("【步骤 4】测试按层级过滤...")
	dialogueResults, err := app.MemoryManager.SearchMemory(ctx, query, 10)
	if err != nil {
		log.Fatal("❌ 搜索失败:", err)
	}

	// 手动过滤对话层
	dialogueCount := 0
	for _, r := range dialogueResults {
		if r.Layer == model.LayerDialogue {
			dialogueCount++
		}
	}
	fmt.Printf("✅ 对话层记忆: %d 条\n", dialogueCount)
	fmt.Println()

	// 最终验证
	fmt.Println("========================================")
	if len(results) > 0 {
		fmt.Println("✅ Milvus VectorStore 集成成功！")
		fmt.Println("   向量存储: 正常")
		fmt.Println("   向量检索: 正常")
		fmt.Println("   相似度计算: 正常")
	} else {
		fmt.Println("⚠️  VectorStore 可能使用 Mock 模式")
		fmt.Println("   请检查 Milvus 是否正常运行")
	}
	fmt.Println("========================================")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
