package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// createTables ensures all required tables exist
func (app *ChatApp) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sender TEXT NOT NULL,
			receiver TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			status TEXT NOT NULL DEFAULT 'away',
			last_seen DATETIME NOT NULL,
			joined_at DATETIME NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := app.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %v", err)
		}
	}
	return nil
}

func splitWords(text string) []string {
	return strings.Fields(text) // splits by whitespace
}

// registerClient adds a client connection
func (app *ChatApp) registerClient(user string, conn *websocket.Conn) {
	app.clientsMutex.Lock()
	defer app.clientsMutex.Unlock()

	if app.clients[user] == nil {
		app.clients[user] = make(map[*websocket.Conn]bool)
	}
	app.clients[user][conn] = true
	log.Printf("Client registered: %s", user)
}

// unregisterClient removes a client connection
func (app *ChatApp) unregisterClient(user string, conn *websocket.Conn) {
	app.clientsMutex.Lock()
	defer app.clientsMutex.Unlock()

	if app.clients[user] != nil {
		delete(app.clients[user], conn)
		if len(app.clients[user]) == 0 {
			delete(app.clients, user)
			// Update user status to away when no connections
			app.updateUserStatus(user, "away")
		}
	}
	conn.Close()
	log.Printf("Client unregistered: %s", user)
}

// broadcast sends a message to all WebSocket connections of a user
func (app *ChatApp) broadcast(user string, m Message) {
	app.clientsMutex.RLock()
	defer app.clientsMutex.RUnlock()

	if conns, ok := app.clients[user]; ok {
		data, _ := json.Marshal(m)
		for conn := range conns {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Broadcast error: %v", err)
			}
		}
	}
}

// updateUserStatus updates a user's online status
func (app *ChatApp) updateUserStatus(username, status string) {
	_, err := app.db.Exec(
		`UPDATE users SET status = ?, last_seen = datetime('now') WHERE username = ?`,
		status, username)
	if err != nil {
		log.Printf("Error updating user status: %v", err)
	}
}

// heartbeat sends periodic pings to keep connection alive and update user status
func (app *ChatApp) heartbeat(user string, conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
		app.updateUserStatus(user, "online")
	}
}
