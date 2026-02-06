package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ==================== 记忆操作指标 ====================

// MemoryRecallDuration 混合召回耗时分布
// 用途：监控召回性能，识别慢查询
// Labels: stage (session_start/topic_deepening/multi_turn)
var MemoryRecallDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "memory_recall_duration_seconds",
		Help:    "Duration of hybrid memory recall operations",
		Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
	},
	[]string{"stage"},
)

// MemoryRecallTotal 召回操作总数
// 用途：统计召回次数、成功率
// Labels: strategy (混合策略), status (success/failure)
var MemoryRecallTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "memory_recall_total",
		Help: "Total number of memory recall operations",
	},
	[]string{"strategy", "status"},
)

// MemoryRecallResultsCount 召回记忆数量分布
// 用途：分析召回质量（是否召回过多/过少）
// Labels: layer (dialogue/topic/profile)
var MemoryRecallResultsCount = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "memory_recall_results_count",
		Help:    "Number of memories returned by recall operations",
		Buckets: []float64{0, 1, 2, 3, 5, 10, 20, 50}, // 自定义桶：记忆数量范围
	},
	[]string{"layer"},
)

// MemoryCreateTotal 创建记忆总数
// 用途：监控记忆增长速率
// Labels: layer (dialogue/topic/profile), status (success/failure)
var MemoryCreateTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "memory_create_total",
		Help: "Total number of memory creation operations",
	},
	[]string{"layer", "status"},
)

// MemoryCreateDuration 创建记忆耗时（含 Embedding 生成）
// 用途：识别性能瓶颈（Embedding API 慢？数据库写入慢？）
// Labels: layer (dialogue/topic/profile)
var MemoryCreateDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "memory_create_duration_seconds",
		Help:    "Duration of memory creation operations (including embedding)",
		Buckets: []float64{0.1, 0.5, 1, 2, 3, 5, 10}, // 创建耗时通常较长（包含 Embedding）
	},
	[]string{"layer"},
)

// ==================== LLM 调用指标 ====================

// LLMRequestsTotal LLM 调用次数
// 用途：统计各 Provider 使用频率
// Labels: provider (openai/gemini), model (gpt-4/gemini-pro), operation (chat/summarize/extract)
var LLMRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "llm_requests_total",
		Help: "Total number of LLM API requests",
	},
	[]string{"provider", "model", "operation"},
)

// LLMDuration LLM 响应耗时
// 用途：监控 LLM API 性能，识别超时风险
// Labels: provider, model
var LLMDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "llm_duration_seconds",
		Help:    "Duration of LLM API calls",
		Buckets: []float64{0.5, 1, 2, 5, 10, 20, 30}, // LLM 响应通常较慢
	},
	[]string{"provider", "model"},
)

// LLMTokensUsed Token 消耗统计（成本监控核心指标）
// 用途：计算 API 成本 (Token * 单价)
// Labels: provider, model, type (input/output)
var LLMTokensUsed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "llm_tokens_used_total",
		Help: "Total number of tokens consumed by LLM API (for cost tracking)",
	},
	[]string{"provider", "model", "type"}, // type: input/output（output 更贵）
)

// LLMErrorsTotal LLM 错误计数
// 用途：监控 API 稳定性，触发告警
// Labels: provider, error_type (timeout/rate_limit/server_error)
var LLMErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "llm_errors_total",
		Help: "Total number of LLM API errors",
	},
	[]string{"provider", "error_type"},
)

// ==================== Embedding 指标 ====================

// EmbeddingRequestsTotal Embedding 调用次数
// 用途：监控向量化频率
var EmbeddingRequestsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "embedding_requests_total",
		Help: "Total number of embedding API requests",
	},
)

// EmbeddingDuration Embedding 生成耗时
// 用途：监控 Embedding API 性能
var EmbeddingDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "embedding_duration_seconds",
		Help:    "Duration of embedding generation",
		Buckets: []float64{0.1, 0.5, 1, 2, 3, 5}, // Embedding 通常较快
	},
)

// EmbeddingErrorsTotal Embedding 错误计数
// 用途：监控限流、超时等问题
// Labels: error_type (throttled/timeout/invalid_response)
var EmbeddingErrorsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "embedding_errors_total",
		Help: "Total number of embedding API errors",
	},
	[]string{"error_type"},
)

// EmbeddingThrottleWaitSeconds 限流等待时间
// 用途：分析限流影响（当前 1s 强制间隔）
var EmbeddingThrottleWaitSeconds = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "embedding_throttle_wait_seconds",
		Help:    "Time spent waiting due to rate limiting",
		Buckets: []float64{0.1, 0.3, 0.5, 1, 2}, // 等待时间通常 < 1s
	},
)

// ==================== 数据库指标 ====================

// PostgresQueryDuration PostgreSQL 查询耗时
// 用途：识别慢查询
// Labels: operation (select/insert/update)
var PostgresQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "postgres_query_duration_seconds",
		Help:    "Duration of PostgreSQL queries",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"operation"},
)

// MilvusSearchDuration Milvus 向量搜索耗时
// 用途：监控向量检索性能
var MilvusSearchDuration = promauto.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "milvus_search_duration_seconds",
		Help:    "Duration of Milvus vector search operations",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1}, // 向量搜索通常很快
	},
)

// ==================== 系统指标 ====================

// GoroutinesCount 当前 Goroutine 数量
// 用途：检测 Goroutine 泄漏
var GoroutinesCount = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "goroutines_count",
		Help: "Current number of goroutines",
	},
)

// MemoryUsageBytes 进程内存使用量
// 用途：监控内存占用，预警 OOM
var MemoryUsageBytes = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "memory_usage_bytes",
		Help: "Current memory usage in bytes",
	},
)

// ==================== Redis 队列指标 ====================

// RedisQueueLength 消息队列长度
// 用途：监控积压情况
var RedisQueueLength = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "redis_queue_length",
		Help: "Current length of Redis message queue",
	},
)

// ==================== 业务指标（可选）====================

// UserActiveCount 活跃用户数（需要应用层统计）
// 用途：业务分析
var UserActiveCount = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "user_active_count",
		Help: "Number of active users (last 24h)",
	},
)
