package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/yourusername/MemoryOs/internal/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	fmt.Printf("Embedding 配置:\n")
	fmt.Printf("  BaseURL: %s\n", cfg.Embedding.BaseURL)
	fmt.Printf("  Model: %s\n", cfg.Embedding.Model)
	fmt.Printf("  APIKey: %s...%s\n\n", cfg.Embedding.APIKey[:10], cfg.Embedding.APIKey[len(cfg.Embedding.APIKey)-6:])

	// 手动构造请求
	url := cfg.Embedding.BaseURL + "/v1/embeddings"
	requestBody := map[string]interface{}{
		"model": cfg.Embedding.Model,
		"input": "你好，这是一个测试文本",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	fmt.Printf("请求URL: %s\n", url)
	fmt.Printf("请求Body: %s\n\n", string(bodyBytes))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatal("创建请求失败:", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Embedding.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("发送请求失败:", err)
	}
	defer resp.Body.Close()

	fmt.Printf("响应状态: %s\n", resp.Status)
	fmt.Printf("响应Header:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %v\n", k, v)
	}

	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取响应失败:", err)
	}

	fmt.Printf("\n原始响应:\n%s\n\n", string(bodyBytes))

	if resp.StatusCode != 200 {
		log.Fatalf("请求失败: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Fatal("解析JSON失败:", err)
	}

	data := result["data"].([]interface{})
	embedding := data[0].(map[string]interface{})["embedding"].([]interface{})

	fmt.Printf("✅ Embedding 成功!\n")
	fmt.Printf("   向量维度: %d\n", len(embedding))
}
