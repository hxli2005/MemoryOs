package adapter

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/model"
	"github.com/yourusername/MemoryOs/internal/metrics"
)

var (
	// 全局互斥锁,确保 Embedding 请求串行执行,避免触发服务商的并发限制
	embeddingMutex sync.Mutex
	// 最小请求间隔,防止瞬时连续请求被 WAF 拦截 (调整为 1s 以更好地规避限流)
	minRequestInterval = 1 * time.Second
	lastRequestTime    time.Time
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
	start := time.Now()

	// 延迟记录指标
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.EmbeddingDuration.Observe(duration)
		metrics.EmbeddingRequestsTotal.Inc()
	}()

	// 🔒 全局锁:确保同一时间只有一个 Embedding 请求,防止并发触发 403
	embeddingMutex.Lock()
	defer embeddingMutex.Unlock()

	// ⏱️ 请求节流:确保两次请求之间有最小间隔
	if !lastRequestTime.IsZero() {
		elapsed := time.Since(lastRequestTime)
		if elapsed < minRequestInterval {
			waitTime := minRequestInterval - elapsed
			time.Sleep(waitTime)
			// 记录实际等待时间
			metrics.EmbeddingThrottleWaitSeconds.Observe(waitTime.Seconds())
		}
	}

	result, err := e.embedder.EmbedStrings(ctx, []string{text})
	lastRequestTime = time.Now() // 记录请求时间

	if err != nil {
		// 记录错误类型
		errorType := "unknown"
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Forbidden") {
			errorType = "throttled"
		} else if strings.Contains(errMsg, "timeout") {
			errorType = "timeout"
		} else if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "parse") {
			errorType = "invalid_response"
		}
		metrics.EmbeddingErrorsTotal.WithLabelValues(errorType).Inc()
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
