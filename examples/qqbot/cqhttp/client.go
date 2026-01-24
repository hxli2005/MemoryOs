// Package cqhttp 提供 go-cqhttp WebSocket 客户端
// 仅支持私聊，最小化实现
package cqhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client go-cqhttp WebSocket 客户端
type Client struct {
	wsURL     string
	conn      *websocket.Conn
	mu        sync.Mutex
	onMessage func(msg *PrivateMessage)
	reconnect bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// PrivateMessage 私聊消息
type PrivateMessage struct {
	UserID    int64  `json:"user_id"`
	Message   string `json:"message"`
	MessageID int64  `json:"message_id"`
	Nickname  string `json:"sender_nickname"`
	Time      int64  `json:"time"`
}

// Event go-cqhttp 上报事件
type Event struct {
	PostType    string          `json:"post_type"`
	MessageType string          `json:"message_type"`
	SubType     string          `json:"sub_type"`
	UserID      int64           `json:"user_id"`
	MessageID   int64           `json:"message_id"`
	Message     json.RawMessage `json:"message"` // 兼容数组和字符串格式
	RawMessage  string          `json:"raw_message"`
	Time        int64           `json:"time"`
	Sender      Sender          `json:"sender"`
}

// Sender 发送者信息
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
}

// APIResponse API 响应
type APIResponse struct {
	Status  string      `json:"status"`
	RetCode int         `json:"retcode"`
	Data    interface{} `json:"data"`
	Echo    string      `json:"echo"`
}

// NewClient 创建客户端
// wsURL: WebSocket 地址，如 "ws://127.0.0.1:6700"
func NewClient(wsURL string) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		wsURL:     wsURL,
		reconnect: true,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// OnPrivateMessage 设置私聊消息回调
func (c *Client) OnPrivateMessage(handler func(msg *PrivateMessage)) {
	c.onMessage = handler
}

// Connect 连接到 go-cqhttp
func (c *Client) Connect() error {
	log.Printf("🔌 正在连接 go-cqhttp: %s", c.wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(c.wsURL, nil)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	c.conn = conn
	log.Println("✅ 已连接到 go-cqhttp")

	// 启动消息接收循环
	go c.receiveLoop()

	return nil
}

// receiveLoop 消息接收循环
func (c *Client) receiveLoop() {
	defer func() {
		if c.reconnect {
			log.Println("⚠️  连接断开，5秒后重连...")
			time.Sleep(5 * time.Second)
			if err := c.Connect(); err != nil {
				log.Printf("❌ 重连失败: %v", err)
			}
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("❌ 读取消息失败: %v", err)
			return
		}

		// 调试：打印原始消息
		log.Printf("🔍 收到原始数据: %s", string(message))

		// 解析事件
		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("⚠️  JSON解析失败: %v", err)
			continue
		}

		// 调试：打印解析结果
		log.Printf("🔍 PostType=%s, MessageType=%s, Message=%s", event.PostType, event.MessageType, event.RawMessage)

		// 只处理私聊消息
		if event.PostType == "message" && event.MessageType == "private" {
			if c.onMessage != nil {
				c.onMessage(&PrivateMessage{
					UserID:    event.UserID,
					Message:   event.RawMessage,
					MessageID: event.MessageID,
					Nickname:  event.Sender.Nickname,
					Time:      event.Time,
				})
			}
		}
	}
}

// SendPrivateMessage 发送私聊消息
func (c *Client) SendPrivateMessage(userID int64, message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("未连接")
	}

	// 构造 API 调用
	request := map[string]interface{}{
		"action": "send_private_msg",
		"params": map[string]interface{}{
			"user_id": userID,
			"message": message,
		},
		"echo": fmt.Sprintf("send_%d_%d", userID, time.Now().UnixNano()),
	}

	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Close 关闭连接
func (c *Client) Close() {
	c.reconnect = false
	c.cancel()
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("👋 go-cqhttp 连接已关闭")
}
