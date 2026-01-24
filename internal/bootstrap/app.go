package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/redis/go-redis/v9"
	"google.golang.org/genai"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	geminiEmbedding "github.com/cloudwego/eino-ext/components/embedding/gemini"
	einoEmbedding "github.com/cloudwego/eino-ext/components/embedding/openai"

	"github.com/yourusername/MemoryOs/internal/adapter"
	"github.com/yourusername/MemoryOs/internal/config"
	"github.com/yourusername/MemoryOs/internal/llm"
	"github.com/yourusername/MemoryOs/internal/mock"
	"github.com/yourusername/MemoryOs/internal/service/memory"
	milvusStore "github.com/yourusername/MemoryOs/internal/storage/milvus"
	postgresStore "github.com/yourusername/MemoryOs/internal/storage/postgres"
)

// App 应用容器
type App struct {
	Config        *config.Config
	DB            *gorm.DB
	Redis         *redis.Client
	MemoryManager *memory.Manager
}

// Initialize 初始化应用
func Initialize(configPath string) (*App, error) {
	ctx := context.Background()

	// 1. 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 环境变量覆盖配置（Docker Compose 支持）
	overrideConfigFromEnv(cfg)

	// 检查是否使用 Mock 模式（开发环境）
	// 只有配置为空时才使用 Mock，localhost 也会尝试连接
	useMockMode := cfg.Database.Postgres.Host == ""

	var db *gorm.DB
	var rdb *redis.Client

	if !useMockMode {
		// 2. 初始化 PostgreSQL
		db, err = initDB(cfg.Database.Postgres)
		if err != nil {
			log.Printf("⚠️  数据库连接失败，切换到 Mock 模式: %v", err)
			useMockMode = true
		}

		// 3. 初始化 Redis
		if !useMockMode {
			rdb = initRedis(cfg.Database.Redis)
			if err := rdb.Ping(ctx).Err(); err != nil {
				log.Printf("⚠️  Redis 连接失败，切换到 Mock 模式: %v", err)
				useMockMode = true
			}
		}
	}

	if useMockMode {
		log.Println("🔧 使用 Mock 模式运行（postgres.host 未配置）")
	} else {
		log.Printf("🔌 连接到数据库: %s@%s:%d/%s", cfg.Database.Postgres.User, cfg.Database.Postgres.Host, cfg.Database.Postgres.Port, cfg.Database.Postgres.DBName)
	}

	// 4. 初始化 Eino 组件
	embedder, err := initEmbedding(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化 Embedding 失败: %w", err)
	}

	llmClient, err := initLLM(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化 LLM 失败: %w", err)
	}

	// 5. 初始化存储层
	var vectorStore memory.VectorStore
	var metaStore memory.MetadataStore

	if useMockMode {
		// Mock 模式
		log.Println("📦 使用 Mock 存储")
		vectorStore = mock.NewMockVectorStore()
		metaStore = mock.NewMockMetadataStore()
	} else {
		// 真实存储模式
		log.Println("🗄️  使用 PostgreSQL 存储")
		metaStore = postgresStore.NewMetadataStore(db)

		// 初始化 Milvus VectorStore
		if cfg.Vector.Provider == "milvus" {
			milvusVS, err := milvusStore.NewVectorStore(milvusStore.Config{
				Host:           cfg.Vector.Milvus.Host,
				Port:           cfg.Vector.Milvus.Port,
				CollectionName: "memories",
				Dimension:      cfg.Embedding.Dimension,
			})
			if err != nil {
				log.Printf("⚠️  Milvus 初始化失败，切换到 Mock: %v", err)
				vectorStore = mock.NewMockVectorStore()
			} else {
				vectorStore = milvusVS
			}
		} else {
			log.Println("⚠️  VectorStore 未配置或不支持，使用 Mock")
			vectorStore = mock.NewMockVectorStore()
		}
	}

	// 6. 初始化记忆管理器
	memoryManager := memory.NewManager(
		vectorStore,
		metaStore,
		embedder,
		llmClient,
		memory.Config{
			MaxWorkingMemory:     cfg.Memory.MaxWorkingMemory,
			CompressionThreshold: cfg.Memory.CompressionThreshold,
			DecayDays:            cfg.Memory.DecayDays,
		},
	)

	log.Println("✅ 应用初始化成功")

	return &App{
		Config:        cfg,
		DB:            db,
		Redis:         rdb,
		MemoryManager: memoryManager,
	}, nil
}

func initDB(cfg config.PostgresConfig) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
}

func initRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// Shutdown 优雅关闭
func (app *App) Shutdown() error {
	log.Println("⏳ 正在关闭应用...")

	// 关闭数据库连接（如果存在）
	if app.DB != nil {
		if sqlDB, err := app.DB.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// 关闭 Redis 连接（如果存在）
	if app.Redis != nil {
		app.Redis.Close()
	}

	log.Println("👋 应用已关闭")
	return nil
}

func initEmbedding(ctx context.Context, cfg *config.Config) (memory.Embedder, error) {
	provider := cfg.Embedding.Provider
	if provider == "" {
		provider = cfg.LLM.Provider // 如果未指定，使用 LLM 的 Provider
	}

	switch provider {
	case "gemini":
		clientConfig := &genai.ClientConfig{
			APIKey: cfg.Embedding.APIKey,
		}
		// 配置代理（如果需要）
		if httpClient := getProxyHTTPClient(); httpClient != nil {
			clientConfig.HTTPClient = httpClient
		}

		client, err := genai.NewClient(ctx, clientConfig)
		if err != nil {
			return nil, fmt.Errorf("创建 Gemini Client 失败: %w", err)
		}

		embedder, err := geminiEmbedding.NewEmbedder(ctx, &geminiEmbedding.EmbeddingConfig{
			Client: client,
			Model:  cfg.Embedding.Model,
		})
		if err != nil {
			return nil, err
		}

		return adapter.NewEinoEmbedderWithDim(embedder, cfg.Embedding.Dimension), nil

	case "openai":
		embedConfig := &einoEmbedding.EmbeddingConfig{
			APIKey: cfg.Embedding.APIKey,
			Model:  cfg.Embedding.Model,
		}
		// 支持中转接口：优先使用 Embedding 自己的 BaseURL，否则使用 LLM 的
		if cfg.Embedding.BaseURL != "" {
			embedConfig.BaseURL = cfg.Embedding.BaseURL
		} else if cfg.LLM.BaseURL != "" {
			embedConfig.BaseURL = cfg.LLM.BaseURL
		}

		embedder, err := einoEmbedding.NewEmbedder(ctx, embedConfig)
		if err != nil {
			return nil, err
		}

		return adapter.NewEinoEmbedderWithDim(embedder, cfg.Embedding.Dimension), nil

	default:
		return nil, fmt.Errorf("不支持的 Embedding Provider: %s (支持: gemini, openai)", provider)
	}
}

func initLLM(ctx context.Context, cfg *config.Config) (memory.LLMClient, error) {
	switch cfg.LLM.Provider {
	case "gemini":
		// 使用 GeminiClient
		llmClient, err := llm.NewGeminiClient(llm.GeminiConfig{
			APIKey:  cfg.LLM.APIKey,
			Model:   cfg.LLM.Model,
			BaseURL: cfg.LLM.BaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("创建 Gemini LLM Client 失败: %w", err)
		}
		return llmClient, nil

	case "openai":
		// 使用 OpenAIClient（支持标准 OpenAI API 和兼容接口）
		llmClient, err := llm.NewOpenAIClient(llm.OpenAIConfig{
			APIKey:  cfg.LLM.APIKey,
			Model:   cfg.LLM.Model,
			BaseURL: cfg.LLM.BaseURL,
		})
		if err != nil {
			return nil, fmt.Errorf("创建 OpenAI LLM Client 失败: %w", err)
		}
		return llmClient, nil

	default:
		return nil, fmt.Errorf("不支持的 LLM Provider: %s (支持: gemini, openai)", cfg.LLM.Provider)
	}
}

// getProxyHTTPClient 获取配置了代理的 HTTP 客户端
// 优先级：环境变量 > 硬编码默认值（127.0.0.1:7890）
func getProxyHTTPClient() *http.Client {
	// 1. 从环境变量读取代理
	proxyURL := os.Getenv("HTTPS_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTP_PROXY")
	}

	// 2. 如果没有环境变量，使用默认代理（Clash 默认端口）
	if proxyURL == "" {
		proxyURL = "http://127.0.0.1:7890"
		log.Printf("ℹ️  未检测到代理环境变量，使用默认代理: %s", proxyURL)
		log.Printf("   如需修改，请设置环境变量: $env:HTTPS_PROXY=\"http://127.0.0.1:端口\"")
	} else {
		log.Printf("✅ 使用代理: %s", proxyURL)
	}

	// 3. 解析并配置代理
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		log.Printf("⚠️  代理 URL 解析失败: %v，将不使用代理", err)
		return nil
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
}

// overrideConfigFromEnv 从环境变量覆盖配置（支持 Docker Compose）
func overrideConfigFromEnv(cfg *config.Config) {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Postgres.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Postgres.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Postgres.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Postgres.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		cfg.Database.Postgres.DBName = dbname
	}

	// Redis 配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Database.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Redis.Port)
	}

	// Milvus 配置
	if host := os.Getenv("MILVUS_HOST"); host != "" {
		cfg.Vector.Milvus.Host = host
	}
	if port := os.Getenv("MILVUS_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Vector.Milvus.Port)
	}

	// LLM 配置
	if provider := os.Getenv("LLM_PROVIDER"); provider != "" {
		cfg.LLM.Provider = provider
	}
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.LLM.Model = model
	}
	if baseURL := os.Getenv("LLM_BASE_URL"); baseURL != "" {
		cfg.LLM.BaseURL = baseURL
	}

	// Embedding 配置
	if provider := os.Getenv("EMBEDDING_PROVIDER"); provider != "" {
		cfg.Embedding.Provider = provider
	}
	if apiKey := os.Getenv("EMBEDDING_API_KEY"); apiKey != "" {
		cfg.Embedding.APIKey = apiKey
	}
	if model := os.Getenv("EMBEDDING_MODEL"); model != "" {
		cfg.Embedding.Model = model
	}
	if dimension := os.Getenv("EMBEDDING_DIMENSION"); dimension != "" {
		fmt.Sscanf(dimension, "%d", &cfg.Embedding.Dimension)
	}

	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
}
