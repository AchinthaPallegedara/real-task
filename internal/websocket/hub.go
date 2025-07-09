package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket connection upgrade
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin for development
		// In production, you should validate the origin
		return true
	},
}

// GetUpgrader returns the WebSocket upgrader
func GetUpgrader() *websocket.Upgrader {
	return &upgrader
}

// Client represents a WebSocket client connection
type Client struct {
	ID         string          // Unique client identifier (user ID)
	UserID     uint            // Database user ID
	ProjectIDs []uint          // Projects this user has access to
	Conn       *websocket.Conn // WebSocket connection
	Send       chan []byte     // Buffered channel for outbound messages
	Hub        *Hub            // Reference to the hub
	LastSeen   time.Time       // Last activity timestamp
}

// Message represents a real-time message structure
type Message struct {
	Type      string      `json:"type"`       // Message type: "task_created", "task_updated", "task_deleted", "user_joined", etc.
	ProjectID uint        `json:"project_id"` // Project this message relates to
	UserID    uint        `json:"user_id"`    // User who triggered the action
	Data      interface{} `json:"data"`       // Message payload
	Timestamp int64       `json:"timestamp"`  // Unix timestamp
}

// IncomingMessage represents messages received from clients
type IncomingMessage struct {
	Type      string      `json:"type"`       // Message type from client
	ProjectID uint        `json:"project_id"` // Project context
	Data      interface{} `json:"data"`       // Message data
}

// TypingIndicator represents typing status
type TypingIndicator struct {
	UserID    uint      `json:"user_id"`
	ProjectID uint      `json:"project_id"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

// UserPresence represents user online/offline status
type UserPresence struct {
	UserID    uint      `json:"user_id"`
	Status    string    `json:"status"` // "online", "offline", "away"
	LastSeen  time.Time `json:"last_seen"`
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by user ID
	clients map[uint]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Inbound messages from clients
	broadcast chan Message

	// Typing indicators by project
	typingIndicators map[uint]map[uint]*TypingIndicator // projectID -> userID -> TypingIndicator

	// User presence status
	userPresence map[uint]*UserPresence // userID -> UserPresence

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[uint]*Client),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan Message),
		typingIndicators: make(map[uint]map[uint]*TypingIndicator),
		userPresence:     make(map[uint]*UserPresence),
	}
}

// Run starts the hub and handles client registration, unregistration, and broadcasting
func (h *Hub) Run() {
	log.Println("🌐 WebSocket Hub started - Real-time collaboration enabled!")
	
	// Start cleanup goroutine for typing indicators
	go h.cleanupTypingIndicators()
	
	for {
		select {
		// Register a new client
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client.UserID] = client
			
			// Update user presence
			h.userPresence[client.UserID] = &UserPresence{
				UserID:   client.UserID,
				Status:   "online",
				LastSeen: time.Now(),
			}
			h.mutex.Unlock()
			
			log.Printf("👤 User %d connected via WebSocket", client.UserID)
			
			// Send a welcome message to the client
			welcomeMsg := Message{
				Type:      "connection_established",
				UserID:    client.UserID,
				Data:      map[string]interface{}{
					"status":  "connected", 
					"message": "Real-time updates enabled",
					"user_id": client.UserID,
					"projects": client.ProjectIDs,
				},
				Timestamp: getCurrentTimestamp(),
			}
			client.sendMessage(welcomeMsg)

			// Broadcast user presence to all projects
			h.broadcastUserPresence(client.UserID, "online")

		// Unregister a client
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
				
				// Update user presence to offline
				if presence, exists := h.userPresence[client.UserID]; exists {
					presence.Status = "offline"
					presence.LastSeen = time.Now()
				}
				
				// Clear typing indicators for this user
				h.clearUserTypingIndicators(client.UserID)
				
				log.Printf("👋 User %d disconnected from WebSocket", client.UserID)
			}
			h.mutex.Unlock()

			// Broadcast user presence to all projects
			h.broadcastUserPresence(client.UserID, "offline")

		// Broadcast a message to relevant clients
		case message := <-h.broadcast:
			h.broadcastToProject(message)
		}
	}
}

// broadcastToProject sends a message to all clients who have access to a specific project
func (h *Hub) broadcastToProject(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Error marshaling message: %v", err)
		return
	}

	sentCount := 0
	for userID, client := range h.clients {
		// Check if the client has access to this project
		if h.clientHasProjectAccess(client, message.ProjectID) {
			select {
			case client.Send <- messageBytes:
				sentCount++
			default:
				// Client's send channel is full, close it
				close(client.Send)
				delete(h.clients, userID)
				log.Printf("⚠️ Removed unresponsive client for user %d", userID)
			}
		}
	}

	log.Printf("📡 Broadcasted %s message to %d clients for project %d", message.Type, sentCount, message.ProjectID)
}

// clientHasProjectAccess checks if a client has access to a specific project
func (h *Hub) clientHasProjectAccess(client *Client, projectID uint) bool {
	for _, pid := range client.ProjectIDs {
		if pid == projectID {
			return true
		}
	}
	return false
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// BroadcastTaskUpdate sends a real-time update about task changes
func (h *Hub) BroadcastTaskUpdate(messageType string, projectID uint, userID uint, taskData interface{}) {
	message := Message{
		Type:      messageType,
		ProjectID: projectID,
		UserID:    userID,
		Data:      taskData,
		Timestamp: getCurrentTimestamp(),
	}

	// Send the message to the broadcast channel
	select {
	case h.broadcast <- message:
		// Message sent successfully
	default:
		log.Printf("⚠️ Broadcast channel is full, message dropped")
	}
}

// BroadcastProjectUpdate sends a real-time update about project changes (new collaborators, etc.)
func (h *Hub) BroadcastProjectUpdate(messageType string, projectID uint, userID uint, projectData interface{}) {
	message := Message{
		Type:      messageType,
		ProjectID: projectID,
		UserID:    userID,
		Data:      projectData,
		Timestamp: getCurrentTimestamp(),
	}

	select {
	case h.broadcast <- message:
		// Message sent successfully
	default:
		log.Printf("⚠️ Broadcast channel is full, message dropped")
	}
}

// sendMessage sends a message directly to a specific client
func (c *Client) sendMessage(message Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Error marshaling message for user %d: %v", c.UserID, err)
		return
	}

	select {
	case c.Send <- messageBytes:
		// Message sent successfully
	default:
		// Channel is full, close it
		close(c.Send)
		log.Printf("⚠️ Send channel full for user %d, closing connection", c.UserID)
	}
}

// StartReadPump starts the read pump for the client
func (c *Client) StartReadPump() {
	c.readPump()
}

// StartWritePump starts the write pump for the client
func (c *Client) StartWritePump() {
	c.writePump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler for keep-alive
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error for user %d: %v", c.UserID, err)
			}
			break
		}

		// Handle incoming messages from clients (e.g., typing indicators, presence updates)
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			log.Printf("❌ Error unmarshaling incoming message from user %d: %v", c.UserID, err)
			continue
		}

		// Update last seen timestamp
		c.LastSeen = time.Now()

		// Handle different message types
		switch incomingMsg.Type {
		case "typing_start":
			if data, ok := incomingMsg.Data.(map[string]interface{}); ok {
				if projectID, ok := data["project_id"].(float64); ok {
					c.Hub.handleTypingIndicator(c.UserID, uint(projectID), true)
				}
			}
		case "typing_stop":
			if data, ok := incomingMsg.Data.(map[string]interface{}); ok {
				if projectID, ok := data["project_id"].(float64); ok {
					c.Hub.handleTypingIndicator(c.UserID, uint(projectID), false)
				}
			}
		case "presence_update":
			if data, ok := incomingMsg.Data.(map[string]interface{}); ok {
				if status, ok := data["status"].(string); ok {
					c.Hub.handlePresenceUpdate(c.UserID, status)
				}
			}
		default:
			log.Printf("📥 Received unknown message type '%s' from user %d", incomingMsg.Type, c.UserID)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleTypingIndicator processes typing indicator messages from clients
func (h *Hub) handleTypingIndicator(userID uint, projectID uint, isTyping bool) {
	h.mutex.Lock()
	
	// Initialize project typing indicators map if needed
	if h.typingIndicators[projectID] == nil {
		h.typingIndicators[projectID] = make(map[uint]*TypingIndicator)
	}
	
	if isTyping {
		// User is typing
		h.typingIndicators[projectID][userID] = &TypingIndicator{
			UserID:    userID,
			ProjectID: projectID,
			IsTyping:  true,
			Timestamp: time.Now(),
		}
	} else {
		// User stopped typing
		delete(h.typingIndicators[projectID], userID)
	}
	
	h.mutex.Unlock()
	
	// Broadcast typing indicator to project collaborators
	typingData := map[string]interface{}{
		"user_id":    userID,
		"project_id": projectID,
		"is_typing":  isTyping,
	}
	
	h.BroadcastProjectUpdate("typing_indicator", projectID, userID, typingData)
}

// handlePresenceUpdate processes user presence updates
func (h *Hub) handlePresenceUpdate(userID uint, status string) {
	h.mutex.Lock()
	if presence, exists := h.userPresence[userID]; exists {
		presence.Status = status
		presence.LastSeen = time.Now()
	}
	h.mutex.Unlock()
	
	// Broadcast presence update to all user's projects
	h.broadcastUserPresence(userID, status)
}

// broadcastUserPresence broadcasts user presence to all their accessible projects
func (h *Hub) broadcastUserPresence(userID uint, status string) {
	h.mutex.RLock()
	client, exists := h.clients[userID]
	h.mutex.RUnlock()
	
	if !exists {
		return
	}
	
	// Broadcast to all user's projects
	for _, projectID := range client.ProjectIDs {
		presenceData := map[string]interface{}{
			"user_id": userID,
			"status":  status,
		}
		h.BroadcastProjectUpdate("user_presence", projectID, userID, presenceData)
	}
}

// GetConnectedUsers returns list of currently connected users
func (h *Hub) GetConnectedUsers() []uint {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	users := make([]uint, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// GetUserPresence returns presence information for a user
func (h *Hub) GetUserPresence(userID uint) *UserPresence {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	if presence, exists := h.userPresence[userID]; exists {
		return presence
	}
	return nil
}

// cleanupTypingIndicators periodically cleans up old typing indicators
func (h *Hub) cleanupTypingIndicators() {
	for {
		time.Sleep(1 * time.Minute)

		h.mutex.Lock()
		for projectID, indicators := range h.typingIndicators {
			for userID, indicator := range indicators {
				// Remove indicators older than 5 minutes
				if time.Since(indicator.Timestamp) > 5*time.Minute {
					delete(indicators, userID)
				}
			}

			// If no more indicators for the project, remove the project entry
			if len(indicators) == 0 {
				delete(h.typingIndicators, projectID)
			}
		}
		h.mutex.Unlock()
	}
}

// clearUserTypingIndicators removes all typing indicators for a user
func (h *Hub) clearUserTypingIndicators(userID uint) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, indicators := range h.typingIndicators {
		delete(indicators, userID)
	}
}

// getCurrentTimestamp returns the current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
