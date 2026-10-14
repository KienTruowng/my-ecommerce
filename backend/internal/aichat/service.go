package aichat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const deepseekAPI = "https://api.deepseek.com/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
	Usage any    `json:"usage,omitempty"`
}

type deepseekReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type deepseekResp struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Service struct {
	apiKey     string
	httpClient *http.Client
}

func NewService(apiKey string) *Service {
	return &Service{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (s *Service) Chat(messages []Message) (*ChatResponse, error) {
	// Prepend system prompt for campus marketplace context
	systemMsg := Message{
		Role: "system",
		Content: `You are a helpful AI assistant for CampusTrade, a student-to-student marketplace.
You help students:
- Draft item listings (titles, descriptions)
- Suggest fair trade values
- Recommend safe meetup locations on campus
- Write polite negotiation messages
- Generate counter-offers
Keep responses concise, friendly, and in Vietnamese or English based on user's language.
Use emojis occasionally for warmth. Max 3-4 sentences per response.`,
	}

	allMessages := append([]Message{systemMsg}, messages...)

	reqBody := deepseekReq{
		Model:    "deepseek-chat",
		Messages: allMessages,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", deepseekAPI, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("DeepSeek API returned %d: %s", resp.StatusCode, string(body))
	}

	var dsResp deepseekResp
	if err := json.Unmarshal(body, &dsResp); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	if len(dsResp.Choices) == 0 {
		return &ChatResponse{Reply: "Sorry, I couldn't generate a response. Try again."}, nil
	}

	return &ChatResponse{
		Reply: dsResp.Choices[0].Message.Content,
	}, nil
}
