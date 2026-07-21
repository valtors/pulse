package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type completionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type completionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func New(baseURL, apiKey, model string) *Client {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Client{BaseURL: baseURL, APIKey: apiKey, Model: model}
}

func (c *Client) Complete(system, user string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("no api key set")
	}

	reqBody := completionRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm error %d: %s", resp.StatusCode, string(raw))
	}

	var result completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}
