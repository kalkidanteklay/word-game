package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"second_server/models"

	"github.com/redis/go-redis/v9"
)

var redisClusterClient *redis.ClusterClient

// Initialize Redis Cluster connection
func InitRedisCluster() {
	log.Println("Initializing Redis Cluster connection...")
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
	log.Println("Attempting to save game state to Redis Cluster...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(gameState)
	if err != nil {
		log.Println("Failed to serialize game state:", err)
		return
	}
	log.Println("Game state serialized successfully.")

	err = redisClusterClient.Set(ctx, "game_state", data, 0).Err()
	if err != nil {
		log.Println("Failed to save game state in Redis Cluster:", err)
	}
	log.Println("Game state saved successfully to Redis Cluster!")
}

// Load game state from Redis Cluster
func LoadGameState() (*models.GameState, error) {
	log.Println("Attempting to load game state from Redis Cluster...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := redisClusterClient.Get(ctx, "game_state").Bytes()
	if err != nil {
		return nil, err
	}

	log.Println("Game state retrieved successfully from Redis Cluster.")

	var gameState models.GameState
	err = json.Unmarshal(data, &gameState)
	if err != nil {
		log.Printf("Failed to unmarshal game state: %v", err)
		return nil, err
	}
	log.Println("Game state loaded and deserialized successfully.")
	return &gameState, nil
}
