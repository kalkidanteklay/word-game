package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"scrambled_words/controllers"
	"scrambled_words/db"
	"scrambled_words/routes"
	"scrambled_words/shared"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)

func CheckServerHealth(url string) bool {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url + "/health")
	if err != nil {
		log.Printf("Server %s is down: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

var gameServers = []string{
	"http://localhost:8081",
	"http://localhost:8082",
}

var mu sync.Mutex

func getHealthyServer() string {
	for _, server := range gameServers {
		resp, err := http.Get(server + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			return server
		}
	}
	return ""
}

func WebSocketHandler(c *gin.Context) {

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer clientConn.Close()

	mu.Lock()
	clients[clientConn] = true
	mu.Unlock()

	defer func() {
		mu.Lock()
		delete(clients, clientConn)
		mu.Unlock()
	}()

	for {

		targetServer := getHealthyServer()
		if targetServer == "" {
			log.Println("No available game servers for WebSocket")
			clientConn.WriteMessage(websocket.TextMessage, []byte("Error: No game servers available. Retrying..."))
			time.Sleep(5 * time.Second)
			continue
		}

		targetWS := fmt.Sprintf("ws://%s/ws", targetServer[7:])

		serverConn, _, err := websocket.DefaultDialer.Dial(targetWS, nil)
		if err != nil {
			log.Println("Failed to connect to game server WebSocket:", err)
			clientConn.WriteMessage(websocket.TextMessage, []byte("Error: Unable to connect to game server. Retrying..."))
			time.Sleep(5 * time.Second)
			continue
		}
		defer serverConn.Close()

		clientConn.WriteMessage(websocket.TextMessage, []byte("Connected to game server: "+targetServer))

		go func() {
			for {
				messageType, msg, err := clientConn.ReadMessage()
				if err != nil {
					log.Println("Client WebSocket disconnected:", err)
					return
				}
				if err := serverConn.WriteMessage(messageType, msg); err != nil {
					log.Println("Failed to forward message to server:", err)
					return
				}
			}
		}()

		for {
			messageType, msg, err := serverConn.ReadMessage()
			if err != nil {
				log.Println("Game server WebSocket disconnected:", err)
				clientConn.WriteMessage(websocket.TextMessage, []byte("Game server disconnected. Reconnecting..."))
				break
			}
			if err := clientConn.WriteMessage(messageType, msg); err != nil {
				log.Println("Failed to forward message to client:", err)
				return
			}
		}

		serverConn.Close()
	}
}

func ForwardRequest(c *gin.Context) {

	var targetServer string
	for _, server := range gameServers {
		resp, err := http.Get(server + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			targetServer = server
			break
		}
	}

	if targetServer == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No game servers available"})
		return
	}

	url := fmt.Sprintf("%s%s", targetServer, c.Request.URL.Path)
	req, err := http.NewRequest(c.Request.Method, url, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach game server"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

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

	r := gin.Default()
	r.GET("/ws", WebSocketHandler)

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")

		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	routes.RegisterRoutes(r)
	gameEndpoints := []string{"/start", "/submit", "/menu"}
	for _, endpoint := range gameEndpoints {
		r.Any(endpoint, ForwardRequest)
	}

	db.InitRedis()
	controllers.LoadGameState()

	log.Println("Main server is running on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
