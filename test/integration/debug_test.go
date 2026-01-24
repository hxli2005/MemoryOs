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
	fmt.Println("MemoryOS 功能检查与 Bug 诊断")
	fmt.Println("========================================\n")

	// 1. 初始化应用
	fmt.Println("【步骤 1】初始化应用...")
	app, err := bootstrap.Initialize("config/config.yaml")
	if err != nil {
		log.Fatal("❌ 初始化失败:", err)
	}
	fmt.Println("✅ 应用初始化成功")
	fmt.Printf("   数据库: %v\n", app.DB != nil)
	fmt.Printf("   Redis: %v\n", app.Redis != nil)
	fmt.Println()

	// 2. 测试 Embedding 生成
	fmt.Println("【步骤 2】测试 Embedding 生成...")
	testText := "今天天气怎么样？"

	// 使用 panic recovery 捕获崩溃
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("❌ Embedding 生成时 panic: %v\n", r)
			}
		}()

		embedding, err := app.MemoryManager.LLM().(*mockLLM).embedder.Embed(ctx, testText)
		if err != nil {
			fmt.Printf("❌ Embedding 生成失败: %v\n", err)
		} else {
			fmt.Printf("✅ Embedding 生成成功\n")
			fmt.Printf("   文本: %s\n", testText)
			fmt.Printf("   维度: %d\n", len(embedding))
		}
	}()
	fmt.Println()

	// 3. 测试记忆创建（不包含 Embedding）
	fmt.Println("【步骤 3】测试记忆创建...")
	userID := "debug_user_" + time.Now().Format("150405")

	mem := &model.Memory{
		UserID:  userID,
		Layer:   model.LayerDialogue,
		Type:    model.MemoryTypeUserMessage,
		Content: testText,
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	// 使用 panic recovery
	err = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		return app.MemoryManager.CreateMemory(ctx, mem)
	}()

	if err != nil {
		fmt.Printf("❌ 创建记忆失败: %v\n", err)

		// 尝试诊断问题
		fmt.Println("\n【诊断信息】")
		if app.DB == nil {
			fmt.Println("   ⚠️  数据库连接为空")
		} else {
			sqlDB, _ := app.DB.DB()
			if err := sqlDB.Ping(); err != nil {
				fmt.Printf("   ⚠️  数据库连接异常: %v\n", err)
			} else {
				fmt.Println("   ✅ 数据库连接正常")
			}
		}

		if app.Redis == nil {
			fmt.Println("   ⚠️  Redis 连接为空")
		} else {
			if err := app.Redis.Ping(ctx).Err(); err != nil {
				fmt.Printf("   ⚠️  Redis 连接异常: %v\n", err)
			} else {
				fmt.Println("   ✅ Redis 连接正常")
			}
		}
	} else {
		fmt.Printf("✅ 记忆创建成功\n")
		fmt.Printf("   Memory ID: %s\n", mem.ID)
		fmt.Printf("   Embedding 维度: %d\n", len(mem.Embedding))
	}
	fmt.Println()

	// 4. 检查数据库数据
	if err == nil && app.DB != nil {
		fmt.Println("【步骤 4】检查数据库存储...")
		var count int64
		app.DB.Table("memories").Where("user_id = ?", userID).Count(&count)
		fmt.Printf("   数据库中的记录数: %d\n", count)
	}

	fmt.Println("\n========================================")
	fmt.Println("检查完成")
	fmt.Println("========================================")
}

// mockLLM 临时类型转换辅助
type mockLLM struct {
	embedder interface {
		Embed(context.Context, string) ([]float32, error)
	}
}
