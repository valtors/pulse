package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	c := New("", "", "")
	if c.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default base url, got %s", c.BaseURL)
	}
	if c.Model != "gpt-4o-mini" {
		t.Errorf("expected default model, got %s", c.Model)
	}
}

func TestNew_Custom(t *testing.T) {
	c := New("http://localhost:8080", "key123", "custom-model")
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("expected custom base url, got %s", c.BaseURL)
	}
	if c.APIKey != "key123" {
		t.Errorf("expected key123, got %s", c.APIKey)
	}
	if c.Model != "custom-model" {
		t.Errorf("expected custom-model, got %s", c.Model)
	}
}

func TestComplete_NoAPIKey(t *testing.T) {
	c := New("", "", "")
	_, err := c.Complete("system", "user")
	if err == nil {
		t.Error("expected error for no api key")
	}
}

func TestComplete_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer testkey" {
			t.Errorf("expected Bearer testkey, got %s", r.Header.Get("Authorization"))
		}
		var req completionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode: %v", err)
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("expected system role, got %s", req.Messages[0].Role)
		}
		resp := completionResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "hello there"}}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(srv.URL, "testkey", "test-model")
	result, err := c.Complete("system prompt", "user prompt")
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if result != "hello there" {
		t.Errorf("expected 'hello there', got %s", result)
	}
}

func TestComplete_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	c := New(srv.URL, "key", "model")
	_, err := c.Complete("s", "u")
	if err == nil {
		t.Error("expected error for 500")
	}
}

func TestComplete_NoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(completionResponse{})
	}))
	defer srv.Close()

	c := New(srv.URL, "key", "model")
	_, err := c.Complete("s", "u")
	if err == nil {
		t.Error("expected error for no choices")
	}
}
