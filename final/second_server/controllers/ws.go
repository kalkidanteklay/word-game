package controllers

import (
	"context"
	"log"
	"net/http"
	"second_server/db"
	"second_server/models"
	"second_server/shared"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	shared.Mu.Lock()
	shared.Clients[conn] = true
	shared.Mu.Unlock()

	log.Printf("New WebSocket connection. Total clients: %d\n", len(shared.Clients))

	for {
		var msg shared.Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		if msg.Type == "register" {
			shared.Mu.Lock()
			username := msg.Payload.(map[string]interface{})["username"].(string)
			var user models.User
			collection := db.GetCollection("scrambled_words", "users")

			err = collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
			if err != nil {
				log.Println("User not found:", err)
				shared.Mu.Unlock()
				return
			}

			shared.Players[conn] = shared.Player{ID: user.ID, Name: username, Score: user.Score}
			player := models.Player{
				Name:  shared.Players[conn].Name,
				Score: shared.Players[conn].Score,
			}

			gameState.Players = append(gameState.Players, player)
			shared.Mu.Unlock()

			broadcastPlayerList()
		}

		shared.Broadcast <- msg
	}

	shared.Mu.Lock()
	delete(shared.Clients, conn)
	delete(shared.Players, conn)
	shared.Mu.Unlock()
	log.Printf("WebSocket disconnected. Total clients: %d\n", len(shared.Clients))

	broadcastPlayerList()
}

func broadcastPlayerList() {
	log.Println("Acquiring lock in broadcastPlayerList()")
	shared.Mu.Lock()
	defer func() {
		log.Println("Releasing lock in broadcastPlayerList()")
		shared.Mu.Unlock()
	}()

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in broadcastPlayerList:", r)
		}
	}()

	playerList := []gin.H{}
	for _, player := range shared.Players {
		playerList = append(playerList, gin.H{
			"name":  player.Name,
			"score": player.Score,
		})
	}

	message := shared.Message{
		Type:    "player_list",
		Payload: gin.H{"players": playerList},
	}

	for conn := range shared.Players {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Println("Error broadcasting to client:", err)
			conn.Close()
			delete(shared.Players, conn)
		}
	}
}
