# Architecture

## Overview

Pulse is a personal AI assistant that connects to your services (GitHub, Gmail, Calendar), filters signal from noise, and answers questions about your digital life. It runs as a local web server with an LLM backend, a connector system, and a persistent memory store.

```
  +----------+     +----------+     +----------+
  |  GitHub  |     |  Gmail   |     | Calendar |
  +----+-----+     +----+-----+     +----+-----+
       |                |                |
       v                v                v
  +----+----+    +------+------+   +-----v-----+
  | connector  | connector    |   | connector |
  | (fetch)    | (fetch)      |   | (fetch)   |
  +----+-------+ +----+------+   +-----+-----+
       |              |                |
       v              v                v
  +----+----+   +-----v-----+   +------v----+
  | filter  |   | unread    |   | today's   |
  | (priority) | list      |   | events    |
  +----+----+   +-----------+   +-----------+
       |              |                |
       +------+-------+------+---------+
              |              |
        +-----v-----+  +-----v-----+
        |   agent   |  |  memory   |
        | (LLM)     |  |  (sqlite) |
        +-----+-----+  +-----------+
              |
        +-----v-----+
        |  web UI   |
        | (htmx)    |
        +-----------+
```

## Design Principles

1. **Local-first.** All state lives in `~/.pulse/` (0700). The web server runs on localhost. No data leaves your machine unless you explicitly connect a service.
2. **Connector pattern.** Each service (GitHub, Gmail, Calendar) implements a `Connector` interface. Adding a new service means implementing `Name()`, `Connect(token)`, and `Test()`. The agent discovers connectors at runtime.
3. **Filter, do not flood.** The filter component classifies notifications into urgent, important, and noise. CI status checks are noise. Review requests are urgent. The digest only surfaces what matters.
4. **LLM is optional.** If no API key is configured, the LLM features degrade gracefully. Connectors and memory still work. The agent falls back to rule-based filtering.

## Components

### config (`internal/config`)

JSON config file at `~/.pulse/config.json`.

```json
{
  "llm": {
    "base_url": "https://api.openai.com/v1",
    "api_key": "sk-...",
    "model": "gpt-4o-mini"
  },
  "connectors": {
    "github": { "token": "ghp_..." },
    "gmail": { "token": "ya29..." }
  }
}
```

Defaults: OpenAI base URL, `gpt-4o-mini` model. If config file does not exist, returns empty config with defaults (no crash).

### connect (`internal/connect`)

Service connectors. Each implements the `Connector` interface:

```go
type Connector interface {
    Name() string
    Connect(token string) error
    Test() error
}
```

**GitHubConnector:**
- `Connect(token)` -- validates token via `GET /user`.
- `Notifications(limit)` -- fetches GitHub notifications with subject type, reason, and repository.
- `Test()` -- hits `GET /user` to verify the token is live.

**GmailConnector:**
- `Connect(token)` -- validates via `GET /users/me/profile`.
- `Unread(limit)` -- fetches unread messages with `is:unread` query.
- `Test()` -- profile endpoint.

**CalendarConnector:**
- `Connect(token)` -- validates via `GET /users/me/calendarList`.
- `Today()` -- fetches today's events from the primary calendar.
- `Test()` -- calendar list endpoint.

The `Service` struct registers and manages connectors by name. Tokens are stored in config and persisted to disk.

### agent (`internal/agent`)

The orchestration layer. Wires connectors, memory, and LLM together.

**Agent.Do(input):**
1. Parse input for keywords ("remember", "forget", "digest", "ask").
2. If "remember": store key-value in memory.
3. If "forget": remove key from memory.
4. If "digest": collect from all connected services, filter, format, optionally summarize with LLM.
5. If "ask": build context from memory + connector data, send to LLM, return response.

### filter (`internal/agent/filter`)

Priority classification engine for GitHub notifications.

**Priority levels:**
| Priority | Trigger | Example |
|---|---|---|
| Urgent | `review_requested`, `mention`, `author` on PR | Someone asked you to review a PR |
| Important | Any PR/Issue activity not urgent | A PR you follow was updated |
| Important | New release | A repo you watch released |
| Noise | `CheckSuite`, CI status | Build passed/failed |

**Output format:** Grouped by repository, then by priority. Noise items are counted but not listed. Only urgent and important items appear in the digest.

### llm (`internal/llm`)

OpenAI-compatible chat completions client. Works with any OpenAI-API-compatible backend (OpenAI, Ollama, LM Studio, etc.).

- `Complete(system, user)` -- sends a system+user message pair, returns the assistant response.
- Configurable via env vars `PULSE_LLM_BASE`, `PULSE_LLM_KEY`, `PULSE_LLM_MODEL` or config file.
- If no API key: returns error. Agent handles gracefully.

### memory (`internal/memory`)

SQLite key-value store with categories.

```sql
CREATE TABLE memory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accessed DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

- `Remember(key, value, category)` -- upsert (INSERT ON CONFLICT DO UPDATE). Updates `accessed` timestamp.
- `Forget(key)` -- delete by key.
- `Get(key)` / `All()` / `ByCategory(cat)` -- query operations.
- WAL mode, 5-second busy timeout.

### server (`internal/server`)

HTTP server on `:9090`. Serves the web UI and API.

| Endpoint | Method | Purpose |
|---|---|---|
| `/` | GET | Web UI (htmx frontend) |
| `/health` | GET | Health check |
| `/connect` | POST | Connect a service |
| `/ask` | POST | Ask the agent a question |
| `/memory` | GET/POST | List/store memories |
| `/status` | GET | Connected services + LLM status |
| `/llm` | POST | Configure LLM settings |

### cmd/pulse

CLI entrypoint. Routes to subcommands:
- `pulse` / `pulse serve` -- start the web server
- `pulse connect <svc> <token>` -- connect a service from CLI
- `pulse disconnect <svc>` -- disconnect
- `pulse ask "<question>"` -- ask from CLI
- `pulse digest` -- get filtered summary from CLI
- `pulse remember <key> <value>` -- store memory from CLI
- `pulse forget <key>` -- remove memory from CLI
- `pulse memory` -- list all memories
- `pulse config llm <url> <key> <model>` -- configure LLM

**Rust core integration:** `callRust(args)` shells out to `pulse-core` binary (Rust) for heavy data processing. The Go CLI passes `--data <datadir>` to the Rust binary. If `pulse-core` is not found, commands that need it fail gracefully.

## Process Model

Single Go process. HTTP server on `:9090`. Optional `pulse-core` Rust binary called as subprocess for data-heavy operations.

## File Layout

```
~/.pulse/
  config.json          # LLM + connector tokens
  pulse.db             # SQLite memory store (WAL mode)
  pulse.db-wal
  pulse.db-shm
  pulse-core           # Rust binary (if installed)
```

## Testing

123 tests, 58.1% coverage. Tests cover:
- Config: load, save, defaults, missing file
- Memory: remember, forget, get, all, by category, concurrent access
- Filter: priority classification, grouping, formatting, noise filtering
- Server: health, connect, ask, memory, status endpoints
- Agent: connector wiring, digest generation, LLM integration
- Connectors: token validation, API mocking (GitHub, Gmail, Calendar)

Untested: `callRust` (needs pulse-core binary), `cmd/pulse` main (os.Exit blocking), connector `Test()` methods (hit real APIs).

## Dependencies

- `modernc.org/sqlite` -- Pure-Go SQLite
- Go stdlib for HTTP, JSON, env, process management
- Optional: `pulse-core` Rust binary for data processing
