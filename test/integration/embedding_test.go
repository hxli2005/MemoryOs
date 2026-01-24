package main

import (
	"context"
	"fmt"
	"log"

	einoEmbedding "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/yourusername/MemoryOs/internal/adapter"
	"github.com/yourusername/MemoryOs/internal/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	fmt.Printf("Embedding 配置:\n")
	fmt.Printf("  Provider: %s\n", cfg.Embedding.Provider)
	fmt.Printf("  Model: %s\n", cfg.Embedding.Model)
	fmt.Printf("  BaseURL: %s\n", cfg.Embedding.BaseURL)
	fmt.Printf("  APIKey: %s...%s\n", cfg.Embedding.APIKey[:10], cfg.Embedding.APIKey[len(cfg.Embedding.APIKey)-6:])
	fmt.Println()

	// 初始化 Embedder（必须设置 BaseURL）
	embedConfig := &einoEmbedding.EmbeddingConfig{
		APIKey:  cfg.Embedding.APIKey,
		Model:   cfg.Embedding.Model,
		BaseURL: cfg.Embedding.BaseURL, // ← 添加 BaseURL
	}

	einoEmbedder, err := einoEmbedding.NewEmbedder(ctx, embedConfig)
	if err != nil {
		log.Fatal("创建 Embedder 失败:", err)
	}

	// 使用降维适配器（2560 → 768）
	embedder := adapter.NewEinoEmbedderWithDim(einoEmbedder, cfg.Embedding.Dimension)

	// 测试 Embedding
	fmt.Println("测试 Embedding...")
	text := "你好，这是一个测试文本"
	embedding, err := embedder.Embed(ctx, text)
	if err != nil {
		log.Fatalf("❌ Embedding 失败: %v", err)
	}

	fmt.Printf("✅ Embedding 成功!\n")
	fmt.Printf("   向量维度: %d\n", len(embedding))
	if len(embedding) >= 5 {
		fmt.Printf("   前5个值: %v\n", embedding[:5])
	}
}
