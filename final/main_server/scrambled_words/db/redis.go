package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"scrambled_words/models"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	fmt.Println(" Connected to Redis!")
}

func SaveGameState(gameState *models.GameState) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(gameState)
	if err != nil {
		log.Println(" Failed to serialize game state:", err)
		return
	}

	err = redisClient.Set(ctx, "game_state", data, 0).Err()
	if err != nil {
		log.Println(" Failed to save game state in Redis:", err)
	}
}

func LoadGameState() (*models.GameState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := redisClient.Get(ctx, "game_state").Bytes()
	if err != nil {
		return nil, err
	}

	var gameState models.GameState
	err = json.Unmarshal(data, &gameState)
	if err != nil {
		return nil, err
	}

	return &gameState, nil
}
