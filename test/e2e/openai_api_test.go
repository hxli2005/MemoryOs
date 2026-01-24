package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 测试灵娅AI的 OpenAI 兼容接口
	config := &openai.ChatModelConfig{
		APIKey:  "sk-pMbjWcgMus0HAFZ8uKBPYqzoStHK9letE7dssEXK8IN8ZYX6",
		Model:   "gemini-3-flash-preview",
		BaseURL: "https://api.lingyaai.cn/v1",
	}

	fmt.Println("创建 OpenAI ChatModel...")
	chatModel, err := openai.NewChatModel(ctx, config)
	if err != nil {
		log.Fatalf("创建失败: %v", err)
	}

	fmt.Println("发送测试消息...")
	resp, err := chatModel.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: "你好，请用一句话介绍你自己",
		},
	})

	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	fmt.Printf("✅ 响应成功:\n%s\n", resp.Content)
}
