package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	reqBody, _ := json.Marshal(map[string]string{
		"name": "Test User",
		"email": "test@test.test",
		"password": "password123",
	})
	
	resp, err := http.Post("http://localhost:3000/api/auth/register", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nBody: %s\n", resp.StatusCode, string(body))
}
