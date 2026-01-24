package adapter

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
)

// EinoEmbedder 适配 Eino 的 Embedder 到我们的接口
type EinoEmbedder struct {
	embedder  embedding.Embedder
	targetDim int // 目标维度（0=不降维）
}

func NewEinoEmbedder(embedder embedding.Embedder) *EinoEmbedder {
	return &EinoEmbedder{
		embedder:  embedder,
		targetDim: 0,
	}
}

// NewEinoEmbedderWithDim 创建带降维的 Embedder
func NewEinoEmbedderWithDim(embedder embedding.Embedder, targetDim int) *EinoEmbedder {
	return &EinoEmbedder{
		embedder:  embedder,
		targetDim: targetDim,
	}
}

func (e *EinoEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	result, err := e.embedder.EmbedStrings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	// 转换 float64 到 float32
	vec := float64ToFloat32(result[0])

	// 降维处理（截断）
	if e.targetDim > 0 && len(vec) > e.targetDim {
		vec = vec[:e.targetDim]
	}

	return vec, nil
}

func (e *EinoEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result, err := e.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}

	converted := make([][]float32, len(result))
	for i, vec := range result {
		converted[i] = float64ToFloat32(vec)

		// 降维处理（截断）
		if e.targetDim > 0 && len(converted[i]) > e.targetDim {
			converted[i] = converted[i][:e.targetDim]
		}
	}
	return converted, nil
}

func float64ToFloat32(f64 []float64) []float32 {
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}
	return f32
}

// EinoLLM 适配 Eino 的 ChatModel
type EinoLLM struct {
	chatModel model.ChatModel
}

// 废弃：已使用新的 llm.GeminiClient 替代
// func NewEinoLLM(chatModel model.ChatModel) *EinoLLM {
// 	return &EinoLLM{chatModel: chatModel}
// }

// 废弃：已使用新的 LLM 接口替代
// func (l *EinoLLM) Chat(ctx context.Context, messages []memory.LLMMessage) (string, error) {
// 	// ... 旧实现
// }
