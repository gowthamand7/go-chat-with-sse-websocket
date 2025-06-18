package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

// NewChatApp creates a new chat application instance
func NewChatApp() *ChatApp {
	return &ChatApp{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

func main() {
	app := NewChatApp()

	var err error
	app.db, err = sql.Open("sqlite", "./chat.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer app.db.Close()

	if err := app.createTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Use(static.Serve("/assets", static.LocalFile("./assets", false)))

	// Routes
	r.GET("/", app.indexHandler)
	r.POST("/create-user", app.createUserHandler)
	r.POST("/join", app.joinHandler)
	r.GET("/users", app.getUsersHandler)
	r.GET("/events", app.sseHistoryHandler)
	r.GET("/ws", app.websocketHandler)

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
