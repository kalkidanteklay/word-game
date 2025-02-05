package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"third_server/models"

	"github.com/redis/go-redis/v9"
)

var redisClusterClient *redis.ClusterClient

func InitRedisCluster() {
	redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7001",
			"localhost:7002",
			"localhost:7003",
		},
		Password: "",
	})

	_, err := redisClusterClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis Cluster:", err)
	}

	fmt.Println("Connected to Redis Cluster!")
}

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
