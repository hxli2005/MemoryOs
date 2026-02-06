package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourusername/MemoryOs/internal/bootstrap"
	"github.com/yourusername/MemoryOs/internal/handler"
	"github.com/yourusername/MemoryOs/internal/metrics"
)

func main() {
	// 初始化应用
	app, err := bootstrap.Initialize("config/config.yaml")
	if err != nil {
		log.Fatalf("❌ 初始化失败: %v", err)
	}
	defer app.Shutdown()

	// 设置 Gin 模式
	if app.Config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.Default()

	// 静态文件服务（测试页面）
	router.StaticFile("/test", "test/index.html")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		dbOK := false
		redisOK := false

		if app.DB != nil {
			if sqlDB, err := app.DB.DB(); err == nil {
				dbOK = sqlDB.Ping() == nil
			}
		}
		if app.Redis != nil {
			redisOK = app.Redis.Ping(context.Background()).Err() == nil
		}

		status := "healthy"
		if app.DB != nil && !dbOK {
			status = "degraded"
		}
		if app.Redis != nil && !redisOK {
			status = "degraded"
		}

		c.JSON(200, gin.H{
			"status":  status,
			"service": "MemoryOS",
			"version": "0.1.0",
			"mode":    "Mock",
			"db":      dbOK,
			"redis":   redisOK,
		})
	})

	// Prometheus Metrics 端点
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 系统指标采集（定时更新）
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			metrics.GoroutinesCount.Set(float64(runtime.NumGoroutine()))
			metrics.MemoryUsageBytes.Set(float64(m.Alloc))
		}
	}()

	// 注册业务路由
	memoryHandler := handler.NewMemoryHandler(app.MemoryManager)

	api := router.Group("/api/v1")
	{
		// 记忆管理
		memories := api.Group("/memories")
		{
			memories.POST("", memoryHandler.CreateMemory)        // 创建记忆
			memories.POST("/search", memoryHandler.SearchMemory) // 搜索记忆
			memories.GET("/:id", memoryHandler.GetMemory)        // 获取单个记忆
			memories.GET("", memoryHandler.ListMemories)         // 列出记忆
		}

		// 召回接口
		recall := api.Group("/recall")
		{
			recall.POST("/dialogue", memoryHandler.RecallDialogue) // 召回对话上下文
			recall.POST("/topic", memoryHandler.RecallTopic)       // 召回话题线索
			recall.POST("/profile", memoryHandler.RecallProfile)   // 召回用户画像
			recall.POST("/hybrid", memoryHandler.HybridRecall)     // 混合召回
		}
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 MemoryOS 服务已启动: http://%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ 正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ 服务器关闭失败:", err)
	}

	log.Println("👋 服务器已退出")
}
