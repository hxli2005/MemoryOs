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

	// 初始化应用（使用默认配置，OpenAI Provider）
	app, err := bootstrap.Initialize("./config/config.yaml")
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	fmt.Println("========================================")
	fmt.Println("LLM 聚合功能测试")
	fmt.Println("========================================\n")

	userID := "test_aggregation_" + time.Now().Format("150405")
	sessionID := "session_weather_talk"

	// ========================================
	// 第一部分：创建对话记忆
	// ========================================
	fmt.Println("【1】创建对话记忆...")

	dialogues := []string{
		"用户：今天天气怎么样？",
		"助手：今天北京晴天，气温 15-25度，空气质量优，非常适合户外活动。",
		"用户：那我去爬香山怎么样？",
		"助手：非常好的选择！香山现在正是赏秋叶的好时节，建议早上8点出发，避开人流高峰。记得带水和防晒用品。",
		"用户：需要准备什么装备吗？",
		"助手：建议准备：1. 运动鞋（山路较陡）2. 防晒霜和帽子 3. 1-2瓶水 4. 少量零食补充能量 5. 手机充满电（导航和拍照）",
	}

	var createdIDs []string
	for _, content := range dialogues {
		mem := &model.Memory{
			UserID:  userID,
			Layer:   model.LayerDialogue,
			Type:    model.MemoryTypeUserMessage,
			Content: content,
			Metadata: map[string]interface{}{
				"session_id": sessionID,
				"timestamp":  time.Now().Unix(),
			},
		}
		if err := app.MemoryManager.CreateMemory(ctx, mem); err != nil {
			log.Printf("⚠️  创建失败: %v", err)
		} else {
			createdIDs = append(createdIDs, mem.ID)
			fmt.Printf("   ✅ %s\n", truncate(content, 40))
		}
	}
	fmt.Printf("\n   共创建 %d 条对话记忆\n\n", len(createdIDs))

	// ========================================
	// 第二部分：对话 → 话题聚合
	// ========================================
	fmt.Println("【2】执行 LLM 聚合（对话 → 话题）...")

	topicID, err := app.MemoryManager.AggregateDialogueToTopic(ctx, userID, sessionID)
	if err != nil {
		log.Fatalf("❌ 聚合失败: %v", err)
	}

	fmt.Printf("   ✅ 话题记忆已创建: %s\n", topicID)

	// 查询并显示话题记忆详情
	topic, err := app.MemoryManager.MetaStore().GetMemory(ctx, topicID)
	if err != nil {
		log.Printf("⚠️  查询话题失败: %v", err)
	} else {
		fmt.Println("\n   📋 话题详情:")
		fmt.Printf("      标题: %v\n", topic.Metadata["title"])
		fmt.Printf("      摘要: %v\n", topic.Metadata["summary"])
		fmt.Printf("      关键词: %v\n", topic.Metadata["keywords"])
		if dialogueIDs, ok := topic.Metadata["dialogue_ids"].([]interface{}); ok {
			fmt.Printf("      源对话数: %d\n", len(dialogueIDs))
		}
	}
	fmt.Println()

	// ========================================
	// 第三部分：话题 → 画像提炼
	// ========================================
	fmt.Println("【3】执行 LLM 提炼（话题 → 画像）...")

	// 再创建几个话题记忆（模拟多次对话聚合的结果）
	additionalTopics := []struct {
		content  string
		metadata map[string]interface{}
	}{
		{
			content: "用户询问了编程学习路径，对 Go 语言和云原生技术表现出浓厚兴趣",
			metadata: map[string]interface{}{
				"title":    "编程学习咨询",
				"keywords": []string{"Go语言", "云原生", "学习路径"},
			},
		},
		{
			content: "用户分享了晨跑习惯，每周3-4次，关注健康数据追踪",
			metadata: map[string]interface{}{
				"title":    "健康生活习惯",
				"keywords": []string{"晨跑", "健康", "运动"},
			},
		},
	}

	var topicIDs []string
	topicIDs = append(topicIDs, topicID) // 包含第一个话题

	for _, t := range additionalTopics {
		mem := &model.Memory{
			UserID:   userID,
			Layer:    model.LayerTopic,
			Type:     model.MemoryTypeTopicThread,
			Content:  t.content,
			Metadata: t.metadata,
		}
		if err := app.MemoryManager.CreateMemory(ctx, mem); err != nil {
			log.Printf("⚠️  创建话题失败: %v", err)
		} else {
			topicIDs = append(topicIDs, mem.ID)
			fmt.Printf("   ✅ 创建话题: %s\n", t.metadata["title"])
		}
	}

	fmt.Printf("\n   共 %d 个话题用于画像提炼\n\n", len(topicIDs))

	// 执行画像提炼
	profileID, err := app.MemoryManager.ExtractProfileFromTopics(ctx, userID, topicIDs)
	if err != nil {
		log.Fatalf("❌ 提炼失败: %v", err)
	}

	fmt.Printf("   ✅ 画像记忆已创建: %s\n", profileID)

	// 查询并显示画像详情
	profile, err := app.MemoryManager.MetaStore().GetMemory(ctx, profileID)
	if err != nil {
		log.Printf("⚠️  查询画像失败: %v", err)
	} else {
		fmt.Println("\n   👤 用户画像:")
		fmt.Printf("      描述: %s\n", profile.Content)
		if prefs, ok := profile.Metadata["preferences"].(map[string]interface{}); ok {
			fmt.Printf("      偏好: %+v\n", prefs)
		}
		if habits, ok := profile.Metadata["habits"].(map[string]interface{}); ok {
			fmt.Printf("      习惯: %+v\n", habits)
		}
		if features, ok := profile.Metadata["features"].(map[string]interface{}); ok {
			fmt.Printf("      特征: %+v\n", features)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("✅ LLM 聚合功能测试完成！")
	fmt.Println("========================================")
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
