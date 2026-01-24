package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/MemoryOs/internal/bootstrap"
	"github.com/yourusername/MemoryOs/internal/model"
)

func main() {
	ctx := context.Background()

	// 初始化应用
	app, err := bootstrap.InitializeApp(ctx, "config/config.yaml", false)
	if err != nil {
		log.Fatal("初始化失败:", err)
	}

	// 测试创建记忆
	memory := &model.Memory{
		UserID:    "test_user",
		SessionID: "test_session",
		Layer:     model.LayerDialogue,
		Type:      model.TypeInteraction,
		Content:   "测试记忆内容：你好世界",
	}

	fmt.Println("开始创建记忆...")
	createdMemory, err := app.MemoryManager.CreateMemory(ctx, memory)
	if err != nil {
		log.Fatalf("❌ 创建记忆失败: %v", err)
	}

	fmt.Printf("✅ 创建成功!\n")
	fmt.Printf("   Memory ID: %s\n", createdMemory.ID)
	fmt.Printf("   Embedding 维度: %d\n", len(createdMemory.Embedding))
	fmt.Printf("   预期维度: 768\n")

	if len(createdMemory.Embedding) != 768 {
		log.Fatalf("❌ 维度不匹配! 期望 768，实际 %d", len(createdMemory.Embedding))
	}

	fmt.Println("✅ 所有检查通过!")
}
