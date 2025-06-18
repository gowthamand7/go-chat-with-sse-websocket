package main

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message represents a chat message
type Message struct {
	ID        int       `json:"id"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a user with their status
type User struct {
	Username string    `json:"username"`
	Status   string    `json:"status"` // "online" or "away"
	LastSeen time.Time `json:"last_seen"`
	JoinedAt time.Time `json:"joined_at"`
}

// ChatApp holds the application state
type ChatApp struct {
	db           *sql.DB
	upgrader     websocket.Upgrader
	clients      map[string]map[*websocket.Conn]bool
	clientsMutex sync.RWMutex
}
