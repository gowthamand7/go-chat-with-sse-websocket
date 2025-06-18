package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// indexHandler serves the main page
func (app *ChatApp) indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// createUserHandler handles creating new users
func (app *ChatApp) createUserHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	username := req.Username
	if len(username) < 2 || len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 2 and 50 characters"})
		return
	}

	now := time.Now()

	// Check if user already exists
	var existingUser string
	err := app.db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Create new user with 'away' status
	_, err = app.db.Exec(
		`INSERT INTO users (username, status, last_seen, joined_at) 
		 VALUES (?, 'away', ?, ?)`,
		username, now, now)

	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User created successfully",
		"username": username,
	})
}

// joinHandler handles user joining the chat
func (app *ChatApp) joinHandler(c *gin.Context) {
	var req struct {
		Username  string `json:"username" binding:"required"`
		Recipient string `json:"recipient" binding:"required"` // Added recipient to payload
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and Recipient are required"})
		return
	}

	username := req.Username
	recipient := req.Recipient

	if username == recipient {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot chat with yourself"})
		return
	}

	if len(username) < 2 || len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 2 and 50 characters"})
		return
	}

	// Validate recipient exists
	var existingRecipient string
	err := app.db.QueryRow("SELECT username FROM users WHERE username = ?", recipient).Scan(&existingRecipient)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	now := time.Now()

	// Handle user creation or status update
	var existingUser string
	err = app.db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
	if err != nil {
		_, err = app.db.Exec(
			`INSERT INTO users (username, status, last_seen, joined_at) 
			 VALUES (?, 'online', ?, ?)`,
			username, now, now)
		if err != nil {
			log.Printf("Error creating user during join: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join chat"})
			return
		}
	} else {
		_, err = app.db.Exec(
			`UPDATE users SET status = 'online', last_seen = ? WHERE username = ?`,
			now, username)
		if err != nil {
			log.Printf("Error updating user status during join: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join chat"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined successfully"})
}

// getUsersHandler returns list of users with their status
func (app *ChatApp) getUsersHandler(c *gin.Context) {
	currentUser := c.Query("current")

	// Update offline users (users not seen in last 30 seconds)
	_, err := app.db.Exec(
		`UPDATE users SET status = 'away' 
		 WHERE last_seen < datetime('now', '-30 seconds') AND status = 'online'`)
	if err != nil {
		log.Printf("Error updating user statuses: %v", err)
	}

	rows, err := app.db.Query(
		`SELECT username, status, last_seen, joined_at FROM users 
		 WHERE username != ? ORDER BY status DESC, username ASC`, currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username, &user.Status, &user.LastSeen, &user.JoinedAt); err != nil {
			continue
		}
		users = append(users, user)
	}

	c.JSON(http.StatusOK, users)
}

// sseHistoryHandler streams past messages to the client via SSE
func (app *ChatApp) sseHistoryHandler(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		c.String(http.StatusBadRequest, "User parameter required")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rows, err := app.db.Query(
		`SELECT id, sender, receiver, content, created_at FROM messages
		 WHERE receiver = ? OR sender = ?
		 ORDER BY created_at ASC`, user, user)
	if err != nil {
		log.Printf("Error fetching message history: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Sender, &m.Receiver, &m.Content, &m.CreatedAt); err != nil {
			continue
		}

		// Simulate live typing by flushing one word at a time
		words := splitWords(m.Content)
		for _, word := range words {
			stream := struct {
				ID        int       `json:"id"`
				Sender    string    `json:"sender"`
				Receiver  string    `json:"receiver"`
				Content   string    `json:"content"`
				CreatedAt time.Time `json:"created_at"`
			}{
				ID:        m.ID,
				Sender:    m.Sender,
				Receiver:  m.Receiver,
				Content:   word,
				CreatedAt: m.CreatedAt,
			}
			data, _ := json.Marshal(stream)
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			flusher.Flush()
			time.Sleep(70 * time.Millisecond)
		}
	}

	// Signal completion
	c.Writer.Write([]byte("event: done\ndata: {}\n\n"))
	flusher.Flush()
}

// websocketHandler handles real-time messaging
func (app *ChatApp) websocketHandler(c *gin.Context) {
	user := c.Query("user")
	if user == "" {
		c.String(http.StatusBadRequest, "User parameter required")
		return
	}

	conn, err := app.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	app.registerClient(user, conn)
	defer app.unregisterClient(user, conn)

	// Update user as online
	app.updateUserStatus(user, "online")

	// Send periodic heartbeat to keep connection alive
	go app.heartbeat(user, conn)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		var m Message
		if err := json.Unmarshal(msgBytes, &m); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		// Validate message
		if m.Sender == "" || m.Receiver == "" || m.Content == "" {
			continue
		}

		m.CreatedAt = time.Now()

		// Save to database
		_, err = app.db.Exec(
			`INSERT INTO messages (sender, receiver, content, created_at)
			 VALUES (?, ?, ?, ?)`,
			m.Sender, m.Receiver, m.Content, m.CreatedAt)
		if err != nil {
			log.Printf("Database insert error: %v", err)
			continue
		}

		// Update sender's last seen
		app.updateUserStatus(m.Sender, "online")

		// Broadcast to receiver
		app.broadcast(m.Receiver, m)
	}
}
