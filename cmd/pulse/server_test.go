package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	handleStatus(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["version"] != version {
		t.Errorf("expected %s, got %v", version, resp["version"])
	}
}

func TestHandleConnect_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/connect", nil)
	w := httptest.NewRecorder()
	handleConnect(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleConnect_BadRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/connect", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	handleConnect(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAsk_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/ask", nil)
	w := httptest.NewRecorder()
	handleAsk(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAsk_BadRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/ask", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	handleAsk(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleMemory_Delete(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/api/memory?key=test", nil)
	w := httptest.NewRecorder()
	handleMemory(w, req)
	if w.Code != 204 {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleMemory_Get(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/memory", nil)
	w := httptest.NewRecorder()
	handleMemory(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleUI_Root(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handleUI(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}
}

func TestHandleUI_NotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	handleUI(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
