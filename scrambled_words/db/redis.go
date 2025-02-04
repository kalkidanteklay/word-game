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

var redisClusterClient *redis.ClusterClient

// Initialize Redis Cluster connection
func InitRedisCluster() {
	redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7001", // Add your Redis cluster nodes here
			"localhost:7002",
			"localhost:7003", // Example: additional node
		},
		Password: "", // Add password if set in Redis
	})

	// Test the connection
	_, err := redisClusterClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis Cluster:", err)
	}

	fmt.Println("Connected to Redis Cluster!")
}

// Save game state to Redis Cluster
func SaveGameState(gameState *models.GameState) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(gameState)
	if err != nil {
		log.Println("Failed to serialize game state:", err)
		return
	}

	err = redisClusterClient.Set(ctx, "game_state", data, 0).Err()
	if err != nil {
		log.Println("Failed to save game state in Redis Cluster:", err)
	}
}

// Load game state from Redis Cluster
func LoadGameState() (*models.GameState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := redisClusterClient.Get(ctx, "game_state").Bytes()
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
