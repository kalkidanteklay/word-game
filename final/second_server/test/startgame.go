package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"second_server/controllers"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCheckMenu(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()

	router.POST("/menu", controllers.CheckMenu)

	playerID := primitive.NewObjectID().Hex()
	payload := map[string]string{
		"player_id": playerID,
		"type":      "new",
	}
	jsonValue, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/check_menu", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
}

func TestStartGame(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.POST("/start", controllers.StartGame)

	playerID := primitive.NewObjectID().Hex()
	payload := map[string]string{
		"player_id": playerID,
	}
	jsonValue, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/start_game", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, true, response["success"])
	assert.NotEmpty(t, response["word"])
}
