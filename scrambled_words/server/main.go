package main

import (
	"flag"
	"log"
	"net/http"
	"scrambled_words/controllers"
	"scrambled_words/db"
	"scrambled_words/routes"
	"scrambled_words/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Track connected clients

func broadcastMessages() {
	for {
		msg := <-shared.Broadcast
		for client := range shared.Clients {
			err := client.WriteJSON(msg)
			log.Printf("Broadcasting message: %+v\n", msg)
			log.Printf("Connected clients: %d\n", len(clients))
			if err != nil {
				log.Println("WebSocket write error:", err)
				client.Close()
				shared.Mu.Lock()
				delete(shared.Clients, client)
				shared.Mu.Unlock()
			}
		}
	}
}

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	go broadcastMessages()
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5502")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	routes.RegisterRoutes(r)
	db.InitRedis()
	controllers.LoadGameState()

	// Start the server
	log.Printf("Server is running on http://localhost:%s\n", *port)
	if err := r.Run(":" + *port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
