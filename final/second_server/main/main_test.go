package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8081"

func sendPostRequest(endpoint string, requestBody map[string]string) (*http.Response, error) {
	jsonValue, _ := json.Marshal(requestBody)
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewBuffer(jsonValue))
	return resp, err
}

func TestCheckMenu_ValidPlayerID(t *testing.T) {
	resp, err := sendPostRequest("/menu", map[string]string{
		"player_id": "6794d69bc1b5b71a3a2f1e1a",
		"type":      "new",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCheckMenu_InvalidPlayerID(t *testing.T) {
	resp, err := sendPostRequest("/menu", map[string]string{
		"player_id": "invalid_id",
		"type":      "new",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestStartGame_ValidPlayerID(t *testing.T) {
	resp, err := sendPostRequest("/start", map[string]string{
		"player_id": "6794d69bc1b5b71a3a2f1e1a",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStartGame_InvalidPlayerID(t *testing.T) {
	resp, err := sendPostRequest("/start", map[string]string{
		"player_id": "invalid_id",
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
