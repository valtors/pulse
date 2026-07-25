package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	s, err := New(0, t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestHandleHealth(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.handleHealth(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "up" {
		t.Errorf("expected up, got %s", resp["status"])
	}
}

func TestHandleUI(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("expected non-empty body")
	}
	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("expected text/html, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHandleUI_NotFound(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleConnect_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/connect", nil)
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleConnect_BadJSON(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("POST", "/connect", strings.NewReader("bad json"))
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleConnect_UnknownService(t *testing.T) {
	s := newTestServer(t)
	body := `{"service":"unknown","token":"x"}`
	req := httptest.NewRequest("POST", "/connect", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleConnect_GitHubEmptyToken(t *testing.T) {
	s := newTestServer(t)
	body := `{"service":"github","token":""}`
	req := httptest.NewRequest("POST", "/connect", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandleConnect_GmailEmptyToken(t *testing.T) {
	s := newTestServer(t)
	body := `{"service":"gmail","token":""}`
	req := httptest.NewRequest("POST", "/connect", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandleConnect_CalendarEmptyToken(t *testing.T) {
	s := newTestServer(t)
	body := `{"service":"calendar","token":""}`
	req := httptest.NewRequest("POST", "/connect", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleConnect(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandleAsk_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/ask", nil)
	w := httptest.NewRecorder()
	s.handleAsk(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAsk_BadJSON(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("POST", "/ask", strings.NewReader("bad"))
	w := httptest.NewRecorder()
	s.handleAsk(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleMemory_GET(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/memory", nil)
	w := httptest.NewRecorder()
	s.handleMemory(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleMemory_DELETE(t *testing.T) {
	s := newTestServer(t)
	s.mem.Remember("test_key", "test_val", "test")
	req := httptest.NewRequest("DELETE", "/memory?key=test_key", nil)
	w := httptest.NewRecorder()
	s.handleMemory(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleMemory_DELETE_NoKey(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("DELETE", "/memory", nil)
	w := httptest.NewRecorder()
	s.handleMemory(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleStatus(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	s.handleStatus(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["connected"]; !ok {
		t.Error("expected connected field")
	}
	if _, ok := resp["llm"]; !ok {
		t.Error("expected llm field")
	}
}

func TestHandleLLM_MethodNotAllowed(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("GET", "/llm", nil)
	w := httptest.NewRecorder()
	s.handleLLM(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleLLM_BadJSON(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest("POST", "/llm", strings.NewReader("bad"))
	w := httptest.NewRecorder()
	s.handleLLM(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleLLM_Success(t *testing.T) {
	s := newTestServer(t)
	body := `{"base_url":"http://localhost:8080","api_key":"key","model":"gpt-4o"}`
	req := httptest.NewRequest("POST", "/llm", strings.NewReader(body))
	w := httptest.NewRecorder()
	s.handleLLM(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestNewServer(t *testing.T) {
	s := newTestServer(t)
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	if s.mux == nil {
		t.Error("expected non-nil mux")
	}
}

func TestNewServer_EmptyDataDir(t *testing.T) {
	s, err := New(0, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestRoutesRegistered(t *testing.T) {
	s := newTestServer(t)
	paths := []string{"/health", "/connect", "/ask", "/memory", "/status", "/llm", "/"}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
		if w.Code == http.StatusNotImplemented {
			t.Errorf("route %s not registered", p)
		}
	}
}
