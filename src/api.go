package gocan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// API请求和响应数据结构

// ConnectRequest 建立CAN连接请求
type ConnectRequest struct {
	Channel string `json:"channel" binding:"required"`
}

// ConnectResponse 建立CAN连接响应
type ConnectResponse struct {
	Status  string `json:"status"`
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
}

// DisconnectRequest 断开连接请求
type DisconnectRequest struct {
	ID string `json:"id" binding:"required"`
}

// DisconnectResponse 断开连接响应
type DisconnectResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// SendRequest 发送CAN帧请求
type SendRequest struct {
	SessionID string `json:"id" binding:"required"`
	Frame     Frame  `json:"frame" binding:"required"`
}

// SendResponse 发送CAN帧响应
type SendResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// SubscribeRequest 订阅CAN消息请求
type SubscribeRequest struct {
	SessionID string `json:"id" binding:"required"`
	CanID     uint32 `json:"canId" binding:"required"`
}

// CANSession CAN通信会话
type CANSession struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // active, inactive, error
	Bus       *CANBus
}

// APIServer API服务器
type APIServer struct {
	sessions map[string]*CANSession
	mutex    sync.RWMutex
	router   *gin.Engine
}

// NewAPIServer 创建新的API服务器
func NewAPIServer() *APIServer {
	server := &APIServer{
		sessions: make(map[string]*CANSession),
		router:   gin.Default(),
	}
	server.setupRoutes()
	return server
}

// setupRoutes 设置API路由
func (s *APIServer) setupRoutes() {
	api := s.router.Group("/api/can")
	{
		api.POST("/connect", s.handleConnect)
		api.POST("/disconnect", s.handleDisconnect)
		api.POST("/send", s.handleSend)
		api.GET("/subscribe", s.handleSubscribe)
	}
}

// handleConnect 处理连接请求
func (s *APIServer) handleConnect(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建模拟驱动和总线
	mockDriver := &MockDriver{
		frames: []Frame{
			{ID: 291, Data: []byte{9, 10, 11, 12}},
			{ID: 292, Data: []byte{13, 14, 15, 16}},
		},
	}
	bus := NewCANBus(mockDriver)

	// 创建会话
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &CANSession{
		ID:        sessionID,
		Channel:   req.Channel,
		CreatedAt: time.Now(),
		Status:    "active",
		Bus:       bus,
	}

	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.mutex.Unlock()

	c.JSON(http.StatusOK, ConnectResponse{
		Status:  "connected",
		ID:      sessionID,
		Message: fmt.Sprintf("Connected to channel %s", req.Channel),
	})
}

// handleDisconnect 处理断开连接请求
func (s *APIServer) handleDisconnect(c *gin.Context) {
	var req DisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mutex.Lock()
	if session, exists := s.sessions[req.ID]; exists {
		session.Status = "inactive"
		delete(s.sessions, req.ID)
	}
	s.mutex.Unlock()

	c.JSON(http.StatusOK, DisconnectResponse{
		Status:  "disconnected",
		Message: fmt.Sprintf("Session %s disconnected", req.ID),
	})
}

// handleSend 处理发送CAN帧请求
func (s *APIServer) handleSend(c *gin.Context) {
	var req SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mutex.RLock()
	session, exists := s.sessions[req.SessionID]
	s.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 发送CAN帧
	if err := session.Bus.Send(req.Frame); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SendResponse{
		Status:    "sent",
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Frame sent: ID=0x%X", req.Frame.ID),
	})
}

// handleSubscribe 处理订阅请求
func (s *APIServer) handleSubscribe(c *gin.Context) {
	// GET请求使用查询参数
	sessionID := c.Query("id")
	canIDStr := c.Query("canId")

	fmt.Printf("DEBUG: Subscribe request - id: %s, canId: %s\n", sessionID, canIDStr)

	if sessionID == "" || canIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters: id and canId"})
		return
	}

	var canID uint32
	n, err := fmt.Sscanf(canIDStr, "%d", &canID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid canId format: %s (parsed %d items)", err.Error(), n)})
		return
	}

	s.mutex.RLock()
	_, exists := s.sessions[sessionID]
	s.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 设置SSE响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 模拟接收CAN消息
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// 模拟接收CAN帧
			frame := Frame{
				ID:   canID,
				Data: []byte{byte(time.Now().Second()), byte(time.Now().Minute())},
			}

			// 发送SSE事件
			eventData, _ := json.Marshal(map[string]interface{}{
				"session_id": sessionID,
				"can_id":     canID,
				"frame":      frame,
				"timestamp":  time.Now(),
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", eventData)
			c.Writer.Flush()
		}
	}
}

// Run 启动API服务器
func (s *APIServer) Run(addr string) error {
	return s.router.Run(addr)
}

// LoadMockData 加载mock数据到JSON文件
func LoadMockData() error {
	// 创建连接mock数据
	connectData := []ConnectRequest{
		{Channel: "PCAN_USBBUS1"},
		{Channel: "PCAN_USBBUS2"},
		{Channel: "can0"},
	}

	connectFile, err := os.Create("data/connect_mock.json")
	if err != nil {
		return err
	}
	defer connectFile.Close()

	encoder := json.NewEncoder(connectFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(connectData); err != nil {
		return err
	}

	// 创建发送mock数据
	sendData := []SendRequest{
		{
			SessionID: "session_123",
			Frame: Frame{
				ID:   291,
				Data: []byte{1, 2, 3, 4},
			},
		},
		{
			SessionID: "session_123",
			Frame: Frame{
				ID:   292,
				Data: []byte{5, 6, 7, 8},
			},
		},
	}

	sendFile, err := os.Create("data/send_mock.json")
	if err != nil {
		return err
	}
	defer sendFile.Close()

	encoder = json.NewEncoder(sendFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sendData); err != nil {
		return err
	}

	// 创建订阅mock数据
	subscribeData := []SubscribeRequest{
		{SessionID: "session_123", CanID: 291},
		{SessionID: "session_123", CanID: 292},
	}

	subscribeFile, err := os.Create("data/subscribe_mock.json")
	if err != nil {
		return err
	}
	defer subscribeFile.Close()

	encoder = json.NewEncoder(subscribeFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(subscribeData); err != nil {
		return err
	}

	return nil
}
