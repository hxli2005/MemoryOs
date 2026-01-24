package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/MemoryOs/internal/model"
	"github.com/yourusername/MemoryOs/internal/service/memory"
)

// MemoryHandler 记忆处理器
type MemoryHandler struct {
	manager *memory.Manager
}

// NewMemoryHandler 创建记忆处理器
func NewMemoryHandler(manager *memory.Manager) *MemoryHandler {
	return &MemoryHandler{
		manager: manager,
	}
}

// ========== API 请求/响应结构 ==========

// CreateMemoryRequest 创建记忆请求
type CreateMemoryRequest struct {
	UserID   string                 `json:"user_id" binding:"required"`
	Layer    model.MemoryLayer      `json:"layer" binding:"required"`
	Type     model.MemoryType       `json:"type" binding:"required"`
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchMemoryRequest 搜索记忆请求
type SearchMemoryRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Query  string `json:"query" binding:"required"`
	TopK   int    `json:"top_k"`
}

// RecallDialogueRequest 召回对话请求
type RecallDialogueRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	SessionID   string `json:"session_id" binding:"required"`
	RecentTurns int    `json:"recent_turns"`
}

// RecallTopicRequest 召回话题请求
type RecallTopicRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Query  string `json:"query" binding:"required"`
	TopK   int    `json:"top_k"`
}

// RecallProfileRequest 召回画像请求
type RecallProfileRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Category string `json:"category,omitempty"`
}

// HybridRecallRequest HTTP 请求结构
type HybridRecallRequestHTTP struct {
	UserID      string `json:"user_id" binding:"required"`
	SessionID   string `json:"session_id" binding:"required"`
	Query       string `json:"query" binding:"required"`
	DialogStage string `json:"dialog_stage"` // session_start/topic_deepening/multi_turn
	MaxTokens   int    `json:"max_tokens"`
}

// ========== API 处理方法 ==========

// CreateMemory 创建记忆
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mem := &model.Memory{
		UserID:   req.UserID,
		Layer:    req.Layer,
		Type:     req.Type,
		Content:  req.Content,
		Metadata: req.Metadata,
	}

	if err := h.manager.CreateMemory(c.Request.Context(), mem); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "记忆创建成功",
		"id":      mem.ID,
	})
}

// SearchMemory 搜索记忆
func (h *MemoryHandler) SearchMemory(c *gin.Context) {
	var req SearchMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 10
	}

	memories, err := h.manager.SearchMemory(c.Request.Context(), req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(memories),
		"memories": memories,
	})
}

// RecallDialogue 召回对话上下文
func (h *MemoryHandler) RecallDialogue(c *gin.Context) {
	var req RecallDialogueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.RecentTurns == 0 {
		req.RecentTurns = 10
	}

	memories, err := h.manager.RecallDialogueContext(
		c.Request.Context(),
		req.UserID,
		req.SessionID,
		req.RecentTurns,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": req.SessionID,
		"count":      len(memories),
		"dialogue":   memories,
	})
}

// RecallTopic 召回话题线索
func (h *MemoryHandler) RecallTopic(c *gin.Context) {
	var req RecallTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TopK == 0 {
		req.TopK = 5
	}

	memories, err := h.manager.RecallTopicThread(
		c.Request.Context(),
		req.UserID,
		req.Query,
		req.TopK,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(memories),
		"topics": memories,
	})
}

// RecallProfile 召回用户画像
func (h *MemoryHandler) RecallProfile(c *gin.Context) {
	var req RecallProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	memories, err := h.manager.RecallUserProfile(
		c.Request.Context(),
		req.UserID,
		req.Category,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  req.UserID,
		"category": req.Category,
		"count":    len(memories),
		"profile":  memories,
	})
}

// HybridRecall 混合召回
func (h *MemoryHandler) HybridRecall(c *gin.Context) {
	var req HybridRecallRequestHTTP
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.DialogStage == "" {
		req.DialogStage = "multi_turn"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4000
	}

	managerReq := memory.ChatbotRecallRequest{
		UserID:      req.UserID,
		SessionID:   req.SessionID,
		Query:       req.Query,
		DialogStage: req.DialogStage,
		MaxTokens:   req.MaxTokens,
	}

	result, err := h.manager.HybridRecall(c.Request.Context(), managerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dialogue_count": len(result.DialogueMemories),
		"topic_count":    len(result.TopicMemories),
		"profile_count":  len(result.ProfileMemories),
		"strategy":       result.Strategy,
		"tokens_used":    result.TokensUsed,
		"dialogue":       result.DialogueMemories,
		"topics":         result.TopicMemories,
		"profile":        result.ProfileMemories,
	})
}

// GetMemory 获取单个记忆（通过 ID）
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "记忆 ID 不能为空"})
		return
	}

	// TODO: 需要在 Manager 中添加 GetByID 方法
	c.JSON(http.StatusNotImplemented, gin.H{"error": "功能暂未实现"})
}

// ListMemories 列出用户的记忆（分页）
func (h *MemoryHandler) ListMemories(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id 参数必填"})
		return
	}

	layer := c.Query("layer")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// TODO: 需要在 Manager 中添加 List 方法
	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"layer":    layer,
		"limit":    limit,
		"memories": []model.Memory{},
		"message":  "Mock 模式：暂无数据",
	})
}
