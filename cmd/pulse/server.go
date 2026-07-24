package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func serve() {
	port := "9090"
	if p := os.Getenv("PULSE_PORT"); p != "" {
		port = p
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "up"})
	})
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/connect", handleConnect)
	mux.HandleFunc("/api/ask", handleAsk)
	mux.HandleFunc("/api/digest", handleDigest)
	mux.HandleFunc("/api/memory", handleMemory)
	mux.HandleFunc("/", handleUI)

	fmt.Printf("pulse on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("server error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	cfg := loadConfig()
	connected := make([]string, 0)
	for name := range cfg.Connectors {
		connected = append(connected, name)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": connected,
		"llm":       cfg.LLM.APIKey != "",
		"version":   version,
	})
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Service string `json:"service"`
		Token   string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	cfg := loadConfig()
	cfg.SetConnector(req.Service, req.Token)
	cfg.Save("")
	out, err := callRust("connect", req.Service, req.Token)
	if err != nil {
		http.Error(w, string(out), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"service":"%s","status":"connected"}`, req.Service)
}

func handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Input string `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	out, err := callRust("ask", req.Input)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"detail":"%s"}`, strings.ReplaceAll(string(out), `"`, `\"`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func handleDigest(w http.ResponseWriter, r *http.Request) {
	out, err := callRust("digest")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"summary":"error: %s"}`, strings.ReplaceAll(string(out), `"`, `\"`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func handleMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		key := r.URL.Query().Get("key")
		if key != "" {
			callRust("forget", key)
		}
		w.WriteHeader(204)
		return
	}
	out, err := callRust("memory")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(uiHTML))
}
