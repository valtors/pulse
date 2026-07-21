package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/valtors/pulse/internal/agent"
	"github.com/valtors/pulse/internal/llm"
	"github.com/valtors/pulse/internal/memory"
)

type Server struct {
	ag   *agent.Agent
	mem  *memory.Store
	port int
	mux  *http.ServeMux
}

func New(port int, dataDir string) (*Server, error) {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".pulse")
	}
	os.MkdirAll(dataDir, 0700)

	mem, err := memory.New(filepath.Join(dataDir, "pulse.db"))
	if err != nil {
		return nil, fmt.Errorf("init memory: %w", err)
	}

	llmClient := llm.New(
		os.Getenv("PULSE_LLM_BASE"),
		os.Getenv("PULSE_LLM_KEY"),
		os.Getenv("PULSE_LLM_MODEL"),
	)

	ag := agent.New(mem, llmClient)

	s := &Server{
		ag:   ag,
		mem:  mem,
		port: port,
		mux:  http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/connect", s.handleConnect)
	s.mux.HandleFunc("/ask", s.handleAsk)
	s.mux.HandleFunc("/memory", s.handleMemory)
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/llm", s.handleLLM)
	s.mux.HandleFunc("/", s.handleUI)
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "up"})
}

type connectRequest struct {
	Service string `json:"service"`
	Token   string `json:"token"`
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	switch strings.ToLower(req.Service) {
	case "github":
		if err := s.ag.ConnectGitHub(req.Token); err != nil {
			http.Error(w, fmt.Sprintf("connect: %v", err), http.StatusUnauthorized)
			return
		}
	case "gmail":
		if err := s.ag.ConnectGmail(req.Token); err != nil {
			http.Error(w, fmt.Sprintf("connect: %v", err), http.StatusUnauthorized)
			return
		}
	case "calendar":
		if err := s.ag.ConnectCalendar(req.Token); err != nil {
			http.Error(w, fmt.Sprintf("connect: %v", err), http.StatusUnauthorized)
			return
		}
	default:
		http.Error(w, "unknown service", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": req.Service,
		"status":  "connected",
	})
}

type askRequest struct {
	Input string `json:"input"`
}

func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	task, err := s.ag.Do(req.Input)
	if err != nil {
		http.Error(w, fmt.Sprintf("agent error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		key := r.URL.Query().Get("key")
		if key != "" {
			s.mem.Forget(key)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	memories, err := s.mem.All()
	if err != nil {
		http.Error(w, fmt.Sprintf("memory error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memories)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	hasLLM := os.Getenv("PULSE_LLM_KEY") != ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected": s.ag.Connected(),
		"llm":       hasLLM,
	})
}

type llmRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

func (s *Server) handleLLM(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req llmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mem.Remember("llm_base", req.BaseURL, "config")
	s.mem.Remember("llm_model", req.Model, "config")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "configured. restart pulse to apply."})
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, uiHTML())
}

func uiHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>pulse</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #0a0c0f; color: #c8d0d8; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
.container { max-width: 680px; margin: 0 auto; padding: 40px 24px; }
h1 { font-size: 28px; font-weight: 700; color: #e8ecf0; margin-bottom: 4px; }
.subtitle { color: #6b7580; margin-bottom: 32px; font-size: 15px; }
.card { background: #11151a; border: 1px solid #1a1d23; border-radius: 8px; padding: 20px; margin-bottom: 16px; }
.card h2 { font-size: 12px; text-transform: uppercase; letter-spacing: 1.5px; color: #3d4550; margin-bottom: 12px; }
input, textarea { width: 100%; background: #0a0c0f; border: 1px solid #1a1d23; border-radius: 6px; padding: 12px; color: #c8d0d8; font-size: 14px; margin-bottom: 8px; font-family: inherit; }
input:focus, textarea:focus { outline: none; border-color: #3d4550; }
button { background: #1a1d23; border: 1px solid #2a2d33; border-radius: 6px; padding: 10px 20px; color: #e8ecf0; font-size: 14px; cursor: pointer; transition: background 0.2s; }
button:hover { background: #2a2d33; }
.row { display: flex; gap: 8px; }
.row > input { flex: 1; }
pre { background: #0a0c0f; border: 1px solid #1a1d23; border-radius: 6px; padding: 16px; overflow-x: auto; font-size: 14px; color: #c8d0d8; white-space: pre-wrap; word-wrap: break-word; line-height: 1.6; margin-top: 12px; min-height: 60px; }
.status { display: flex; gap: 8px; margin-bottom: 24px; flex-wrap: wrap; }
.badge { padding: 4px 12px; border-radius: 4px; font-size: 12px; background: #1a1d23; color: #6b7580; }
.badge.on { color: #4ade80; border: 1px solid #22c55e33; }
.thinking { color: #6b7580; font-style: italic; }
</style>
</head>
<body>
<div class="container">
  <h1>pulse</h1>
  <p class="subtitle">connect everything. your ai does the rest.</p>

  <div class="status">
    <span class="badge" id="badge-github">github: off</span>
    <span class="badge" id="badge-gmail">gmail: off</span>
    <span class="badge" id="badge-calendar">calendar: off</span>
    <span class="badge" id="badge-llm">llm: off</span>
  </div>

  <div class="card">
    <h2>connect a service</h2>
    <div class="row">
      <input id="service" placeholder="github, gmail, or calendar" />
      <input id="token" placeholder="token" type="password" />
    </div>
    <button onclick="connect()">connect</button>
  </div>

  <div class="card">
    <h2>ask</h2>
    <input id="input" placeholder="what did i miss? what do you know? remember X" onkeydown="if(event.key==='Enter')ask()" autofocus/>
    <button onclick="ask()">ask</button>
    <pre id="output">ask something.</pre>
  </div>

  <div class="card">
    <h2>memory</h2>
    <button onclick="showMemory()">show memory</button>
    <pre id="memory"></pre>
  </div>
</div>

<script>
function connect() {
  const service = document.getElementById('service').value;
  const token = document.getElementById('token').value;
  fetch('/connect', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({service,token})})
    .then(r => r.ok ? r.json() : r.text())
    .then(d => {
      if (d.status === 'connected') {
        document.getElementById('badge-' + service).className = 'badge on';
        document.getElementById('badge-' + service).textContent = service + ': on';
      } else { alert(d); }
    });
}

function ask() {
  const input = document.getElementById('input').value;
  const out = document.getElementById('output');
  out.innerHTML = '<span class="thinking">thinking...</span>';
  fetch('/ask', {method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({input})})
    .then(r => r.json())
    .then(d => {
      out.textContent = d.detail || JSON.stringify(d, null, 2);
    })
    .catch(e => { out.textContent = 'error: ' + e; });
}

function showMemory() {
  fetch('/memory').then(r => r.json()).then(d => {
    const el = document.getElementById('memory');
    if (!d || d.length === 0) { el.textContent = 'nothing remembered yet.'; return; }
    let text = '';
    d.forEach(m => { text += m.key + ': ' + m.value + '\n'; });
    el.textContent = text;
  });
}

function status() {
  fetch('/status').then(r => r.json()).then(d => {
    d.connected.forEach(s => {
      const b = document.getElementById('badge-' + s);
      if (b) { b.className = 'badge on'; b.textContent = s + ': on'; }
    });
    if (d.llm) {
      const b = document.getElementById('badge-llm');
      b.className = 'badge on'; b.textContent = 'llm: on';
    }
  });
}
status();
</script>
</body>
</html>`
}
