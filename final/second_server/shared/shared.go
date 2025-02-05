package shared

import (
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
type Player struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name"`
	Score int                `json:"score"`
	Word  string             `bson:"word"`
}

var (
	Clients   = make(map[*websocket.Conn]bool)
	Players   = make(map[*websocket.Conn]Player)
	Mu        sync.Mutex
	Broadcast = make(chan Message)
)
