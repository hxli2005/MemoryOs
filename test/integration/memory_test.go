package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/MemoryOs/internal/bootstrap"
	"github.com/yourusername/MemoryOs/internal/model"
	"github.com/yourusername/MemoryOs/internal/service/memory"
)

func main() {
	ctx := context.Background()

	fmt.Println("========================================")
	fmt.Println("MemoryOS 集成测试")
	fmt.Println("========================================\n")

	// 1. 初始化应用
	fmt.Println("【步骤 1】初始化应用...")
	app, err := bootstrap.Initialize("config/config.yaml")
	if err != nil {
		log.Fatal("❌ 初始化失败:", err)
	}
	fmt.Println("✅ 应用初始化成功\n")

	// 2. 创建对话记忆
	fmt.Println("【步骤 2】创建对话记忆...")
	userID := "test_user_" + time.Now().Format("150405")

	mem := &model.Memory{
		UserID:  userID,
		Layer:   model.LayerDialogue,
		Type:    model.MemoryTypeUserMessage,
		Content: "用户问：今天天气怎么样？",
		Metadata: map[string]interface{}{
			"dialogue_turn": 1,
		},
	}

	err = app.MemoryManager.CreateMemory(ctx, mem)
	if err != nil {
		log.Fatal("❌ 创建记忆失败:", err)
	}
	fmt.Printf("✅ 记忆创建成功\n")
	fmt.Printf("   Memory ID: %s\n", mem.ID)
	fmt.Printf("   Embedding 维度: %d\n", len(mem.Embedding))
	fmt.Println()

	// 3. 向量搜索
	fmt.Println("【步骤 3】测试向量搜索...")
	query := "天气如何"
	results, err := app.MemoryManager.SearchMemory(ctx, query, 5)
	if err != nil {
		log.Fatal("❌ 向量搜索失败:", err)
	}
	fmt.Printf("✅ 搜索成功，找到 %d 条记忆\n", len(results))
	for i, r := range results {
		fmt.Printf("   [%d] %s\n", i+1, truncate(r.Content, 50))
	}
	fmt.Println()

	// 4. 混合召回
	fmt.Println("【步骤 4】测试混合召回...")
	recallReq := memory.ChatbotRecallRequest{
		UserID:    userID,
		Query:     "今天天气怎么样？",
		SessionID: "test_" + time.Now().Format("150405"),
		MaxTokens: 2000,
	}
	recallResult, err := app.MemoryManager.HybridRecall(ctx, recallReq)
	if err != nil {
		log.Fatal("❌ 混合召回失败:", err)
	}
	fmt.Printf("✅ 混合召回成功\n")
	fmt.Printf("   对话层: %d 条\n", len(recallResult.DialogueMemories))
	fmt.Printf("   话题层: %d 条\n", len(recallResult.TopicMemories))
	fmt.Printf("   画像层: %d 条\n", len(recallResult.ProfileMemories))
	fmt.Println()

	// 最终验证
	fmt.Println("========================================")
	fmt.Println("✅ 所有测试通过！")
	fmt.Println("========================================")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
