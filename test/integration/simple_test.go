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
	fmt.Println("PostgreSQL + Milvus 功能测试")
	fmt.Println("========================================\n")

	// 1. 初始化
	fmt.Println("【1】初始化应用...")
	app, err := bootstrap.Initialize("config/config.yaml")
	if err != nil {
		log.Fatal("❌ 初始化失败:", err)
	}
	fmt.Println("✅ 初始化成功\n")

	// 2. 测试数据库连接
	fmt.Println("【2】测试数据库连接...")
	if app.DB != nil {
		sqlDB, _ := app.DB.DB()
		if err := sqlDB.Ping(); err != nil {
			log.Fatal("❌ 数据库连接失败:", err)
		}
		fmt.Println("✅ PostgreSQL 连接正常")
	} else {
		fmt.Println("⚠️  使用 Mock 数据库")
	}
	fmt.Println()

	// 3. 简单创建测试
	fmt.Println("【3】创建测试记忆...")
	userID := "test_" + time.Now().Format("150405")

	mem := &model.Memory{
		UserID:  userID,
		Layer:   model.LayerDialogue,
		Type:    model.MemoryTypeUserMessage,
		Content: "测试消息",
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	fmt.Println("   开始创建...")
	err = app.MemoryManager.CreateMemory(ctx, mem)
	if err != nil {
		fmt.Printf("❌ 创建失败: %v\n", err)
		fmt.Println("\n【错误详情】")
		fmt.Printf("   错误类型: %T\n", err)
		fmt.Printf("   错误信息: %s\n", err.Error())
		return
	}

	fmt.Printf("✅ 创建成功\n")
	fmt.Printf("   ID: %s\n", mem.ID)
	fmt.Printf("   Embedding 维度: %d\n", len(mem.Embedding))
	fmt.Println()

	// 4. 查询测试
	fmt.Println("【4】查询记录...")
	if app.DB != nil {
		var count int64
		app.DB.Table("memories").Where("user_id = ?", userID).Count(&count)
		fmt.Printf("   PostgreSQL 记录数: %d\n", count)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("✅ 测试完成")
	fmt.Println("========================================")
}
