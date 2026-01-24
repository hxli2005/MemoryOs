// QQ Bot - 接入 NapCat/go-cqhttp
// 支持 Docker 独立部署
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/yourusername/MemoryOs/examples/qqbot/cqhttp"
	"github.com/yourusername/MemoryOs/internal/bootstrap"
	"github.com/yourusername/MemoryOs/pkg/chatbot"
)

// RealQQBot 真实 QQ 机器人
type RealQQBot struct {
	chatbot      chatbot.Chatbot
	cqClient     *cqhttp.Client
	messageQueue chan MessageTask
	workerPool   int
	wg           sync.WaitGroup
}

// MessageTask 消息任务
type MessageTask struct {
	UserID     int64
	Message    string
	ReceivedAt time.Time
	Nickname   string
}

func main() {
	fmt.Println("🤖 真实 QQ Bot - MemoryOS")
	fmt.Println(strings.Repeat("=", 50))

	// ========== 配置区域（支持环境变量）==========
	cqhttpURL := getEnv("CQHTTP_WS_URL", "ws://127.0.0.1:6700")
	configPath := getEnv("CONFIG_PATH", "config/config.yaml")
	workerCount := getEnvInt("WORKER_COUNT", 5)
	// =============================================

	log.Printf("📡 NapCat/go-cqhttp 地址: %s", cqhttpURL)
	log.Printf("📄 配置文件: %s", configPath)

	// 1. 初始化 MemoryOS 核心
	app, err := bootstrap.Initialize(configPath)
	if err != nil {
		log.Fatalf("❌ 初始化 MemoryOS 失败: %v", err)
	}
	defer app.Shutdown()

	// 2. 加载人设配置
	persona := loadPersona()

	// 3. 创建 Chatbot 适配器
	adapter := chatbot.NewMemoryOSAdapter(app.MemoryManager, persona)

	// 4. 创建 go-cqhttp 客户端
	cqClient := cqhttp.NewClient(cqhttpURL)

	// 5. 创建并启动 Bot
	bot := &RealQQBot{
		chatbot:      adapter,
		cqClient:     cqClient,
		messageQueue: make(chan MessageTask, 100),
		workerPool:   workerCount,
	}

	// 6. 设置消息回调
	cqClient.OnPrivateMessage(func(msg *cqhttp.PrivateMessage) {
		bot.onMessage(msg)
	})

	// 7. 连接 go-cqhttp
	if err := cqClient.Connect(); err != nil {
		log.Fatalf("❌ 连接 go-cqhttp 失败: %v", err)
	}

	// 8. 启动 Worker
	bot.startWorkers()

	fmt.Println("✅ QQ Bot 已启动，等待私聊消息...")
	fmt.Println("💡 按 Ctrl+C 停止")

	// 9. 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 10. 优雅关闭
	fmt.Println("\n🛑 正在关闭...")
	close(bot.messageQueue)
	bot.wg.Wait()
	cqClient.Close()
	fmt.Println("👋 QQ Bot 已停止")
}

// onMessage 收到私聊消息
func (b *RealQQBot) onMessage(msg *cqhttp.PrivateMessage) {
	// 过滤空消息
	if strings.TrimSpace(msg.Message) == "" {
		return
	}

	task := MessageTask{
		UserID:     msg.UserID,
		Message:    msg.Message,
		ReceivedAt: time.Now(),
		Nickname:   msg.Nickname,
	}

	select {
	case b.messageQueue <- task:
		log.Printf("📨 [%s(%d)] %s", msg.Nickname, msg.UserID, msg.Message)
	default:
		log.Printf("⚠️  队列已满，丢弃消息")
	}
}

// startWorkers 启动 Worker 协程
func (b *RealQQBot) startWorkers() {
	log.Printf("🚀 启动 %d 个消息处理 Worker", b.workerPool)

	for i := 0; i < b.workerPool; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}
}

// worker 消息处理协程
func (b *RealQQBot) worker(id int) {
	defer b.wg.Done()

	for task := range b.messageQueue {
		b.handleMessage(id, task)
	}
}

// handleMessage 处理单条消息
func (b *RealQQBot) handleMessage(workerID int, task MessageTask) {
	ctx := context.Background()
	userID := "qq_" + strconv.FormatInt(task.UserID, 10)

	// 1. 构造消息
	msg := chatbot.Message{
		UserID:  userID,
		Content: task.Message,
		Metadata: map[string]interface{}{
			"platform":  "qq",
			"nickname":  task.Nickname,
			"qq_number": task.UserID,
		},
	}

	// 2. 调用 Chatbot 处理
	response, err := b.chatbot.Chat(ctx, msg)
	if err != nil {
		log.Printf("❌ [Worker %d] 处理失败: %v", workerID, err)
		b.cqClient.SendPrivateMessage(task.UserID, "啊这...我好像卡住了 (╯°□°）╯︵ ┻━┻")
		return
	}

	// 3. 模拟打字延迟
	if response.Delay > 0 {
		time.Sleep(response.Delay)
	}

	// 4. 发送回复
	if err := b.cqClient.SendPrivateMessage(task.UserID, response.Content); err != nil {
		log.Printf("❌ 发送失败: %v", err)
		return
	}

	log.Printf("💬 [%s(%d)] 回复: %s", task.Nickname, task.UserID, truncate(response.Content, 50))

	// 5. 更新好感度（示例）
	if containsKeyword(task.Message, []string{"谢谢", "感谢", "爱你"}) {
		b.chatbot.UpdateFavorability(ctx, userID, 5, "用户表示感谢")
	}
}

// loadPersona 从配置文件加载人设
func loadPersona() *chatbot.PersonaConfig {
	personaPath := getEnv("PERSONA_PATH", "examples/qqbot/persona.yaml")
	log.Printf("📝 加载人设配置: %s", personaPath)

	data, err := os.ReadFile(personaPath)
	if err != nil {
		log.Printf("⚠️  人设配置加载失败，使用默认: %v", err)
		return defaultPersona()
	}

	var raw PersonaYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Printf("⚠️  人设配置解析失败，使用默认: %v", err)
		return defaultPersona()
	}

	log.Printf("✅ 已加载人设: %s (%s)", raw.Name, raw.Occupation)

	return &chatbot.PersonaConfig{
		Name:          raw.Name,
		Nickname:      raw.Nickname,
		Gender:        raw.Gender,
		Age:           raw.Age,
		Occupation:    raw.Occupation,
		Location:      raw.Location,
		Personality:   raw.Personality,
		Strengths:     raw.Strengths,
		Weaknesses:    raw.Weaknesses,
		Quirks:        raw.Quirks,
		Background:    raw.Background,
		DailyLife:     raw.DailyLife,
		Dreams:        raw.Dreams,
		Worries:       raw.Worries,
		Interests:     raw.Interests,
		Favorites:     raw.Favorites,
		Dislikes:      raw.Dislikes,
		TalkingStyle:  raw.TalkingStyle,
		Catchphrases:  raw.Catchphrases,
		Emojis:        raw.Emojis,
		Tone:          raw.Tone,
		Greeting:      raw.Greeting,
		Farewell:      raw.Farewell,
		IntimacyStyle: raw.IntimacyStyle,
		Forbidden:     raw.Forbidden,
		Boundaries:    raw.Boundaries,
	}
}

// PersonaYAML 人设 YAML 结构
type PersonaYAML struct {
	Name          string            `yaml:"name"`
	Nickname      string            `yaml:"nickname"`
	Gender        string            `yaml:"gender"`
	Age           string            `yaml:"age"`
	Occupation    string            `yaml:"occupation"`
	Location      string            `yaml:"location"`
	Personality   []string          `yaml:"personality"`
	Strengths     []string          `yaml:"strengths"`
	Weaknesses    []string          `yaml:"weaknesses"`
	Quirks        []string          `yaml:"quirks"`
	Background    string            `yaml:"background"`
	DailyLife     string            `yaml:"daily_life"`
	Dreams        []string          `yaml:"dreams"`
	Worries       []string          `yaml:"worries"`
	Interests     []string          `yaml:"interests"`
	Favorites     map[string]string `yaml:"favorites"`
	Dislikes      []string          `yaml:"dislikes"`
	TalkingStyle  string            `yaml:"talking_style"`
	Catchphrases  []string          `yaml:"catchphrases"`
	Emojis        []string          `yaml:"emojis"`
	Tone          string            `yaml:"tone"`
	Greeting      string            `yaml:"greeting"`
	Farewell      string            `yaml:"farewell"`
	IntimacyStyle map[string]string `yaml:"intimacy_style"`
	Forbidden     []string          `yaml:"forbidden"`
	Boundaries    []string          `yaml:"boundaries"`
}

// defaultPersona 默认人设（备用）
func defaultPersona() *chatbot.PersonaConfig {
	return &chatbot.PersonaConfig{
		Name:         "小助手",
		Nickname:     "助手",
		Gender:       "中性",
		Age:          "未知",
		Occupation:   "AI 助手",
		Personality:  []string{"友好", "乐于助人"},
		TalkingStyle: "简洁明了",
		Greeting:     "你好~",
		Farewell:     "再见~",
	}
}

// 辅助函数
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func containsKeyword(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
