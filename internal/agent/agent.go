package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/valtors/pulse/internal/connect"
	"github.com/valtors/pulse/internal/llm"
	"github.com/valtors/pulse/internal/memory"
)

type Agent struct {
	mem      *memory.Store
	llm      *llm.Client
	github   *connect.GitHubConnector
	gmail    *connect.GmailConnector
	calendar *connect.CalendarConnector
}

func New(mem *memory.Store, llmClient *llm.Client) *Agent {
	return &Agent{mem: mem, llm: llmClient}
}

func (a *Agent) ConnectGitHub(token string) error {
	a.github = &connect.GitHubConnector{}
	if err := a.github.Connect(token); err != nil {
		return err
	}
	a.mem.Remember("github_connected", time.Now().Format(time.RFC3339), "config")
	return nil
}

func (a *Agent) ConnectGmail(token string) error {
	a.gmail = &connect.GmailConnector{}
	if err := a.gmail.Connect(token); err != nil {
		return err
	}
	a.mem.Remember("gmail_connected", time.Now().Format(time.RFC3339), "config")
	return nil
}

func (a *Agent) ConnectCalendar(token string) error {
	a.calendar = &connect.CalendarConnector{}
	if err := a.calendar.Connect(token); err != nil {
		return err
	}
	a.mem.Remember("calendar_connected", time.Now().Format(time.RFC3339), "config")
	return nil
}

func (a *Agent) Connected() []string {
	var c []string
	if a.github != nil {
		c = append(c, "github")
	}
	if a.gmail != nil {
		c = append(c, "gmail")
	}
	if a.calendar != nil {
		c = append(c, "calendar")
	}
	return c
}

func (a *Agent) Memory() *memory.Store {
	return a.mem
}

type TaskResult struct {
	Input    string `json:"input"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
}

func (a *Agent) Do(input string) (*TaskResult, error) {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "remember ") {
		parts := strings.SplitN(input, " ", 3)
		if len(parts) >= 3 {
			key := parts[1]
			value := parts[2]
			a.mem.Remember(key, value, "user")
			return &TaskResult{
				Input:  input,
				Action: "remember",
				Detail: fmt.Sprintf("stored: %s = %s", key, value),
			}, nil
		}
	}

	if strings.Contains(lower, "forget ") {
		key := strings.TrimSpace(input[strings.Index(input, " ")+1:])
		a.mem.Forget(key)
		return &TaskResult{
			Input:  input,
			Action: "forget",
			Detail: fmt.Sprintf("forgot: %s", key),
		}, nil
	}

	if a.llm == nil || a.llm.APIKey == "" {
		return a.doWithoutLLM(input)
	}

	return a.doWithLLM(input)
}

func (a *Agent) doWithoutLLM(input string) (*TaskResult, error) {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "what did i miss") || strings.Contains(lower, "summary") {
		data, err := a.gatherContext()
		if err != nil {
			return nil, err
		}
		if data == "" {
			return &TaskResult{
				Input:  input,
				Action: "summarize",
				Detail: "nothing to report. connect a service first.",
			}, nil
		}
		return &TaskResult{
			Input:  input,
			Action: "summarize",
			Detail: data,
		}, nil
	}

	if strings.Contains(lower, "what do you know") || strings.Contains(lower, "what do you remember") {
		mems, _ := a.mem.All()
		out, _ := json.MarshalIndent(mems, "", "  ")
		return &TaskResult{
			Input:  input,
			Action: "recall",
			Detail: string(out),
		}, nil
	}

	return &TaskResult{
		Input:  input,
		Action: "unknown",
		Detail: "connect an llm api key to enable full agent capabilities. or try: what did i miss, remember X, what do you know",
	}, nil
}

func (a *Agent) doWithLLM(input string) (*TaskResult, error) {
	context := a.gatherContextRaw()
	memories, _ := a.mem.All()
	memStr := ""
	if len(memories) > 0 {
		var memLines []string
		for _, m := range memories {
			if m.Category == "user" || m.Category == "config" {
				memLines = append(memLines, fmt.Sprintf("- %s: %s", m.Key, m.Value))
			}
		}
		memStr = strings.Join(memLines, "\n")
	}

	system := fmt.Sprintf(`you are pulse. a personal ai agent that lives on the user's machine.

the user has connected these services: %s

memory about the user:
%s

live data from connected services:
%s

rules:
- be direct. short sentences. no fluff.
- if the user asks what they missed, summarize the live data above. group by service. highlight what needs attention.
- if the user asks you to do something, explain what you would do based on the data available.
- if you don't have the data, say so. don't make things up.
- the user is not technical. speak plainly.`, strings.Join(a.Connected(), ", "), memStr, context)

	resp, err := a.llm.Complete(system, input)
	if err != nil {
		return &TaskResult{
			Input:  input,
			Action: "error",
			Detail: fmt.Sprintf("llm error: %v", err),
		}, nil
	}

	a.mem.Remember("last_interaction", input, "history")

	return &TaskResult{
		Input:  input,
		Action: "respond",
		Detail: resp,
	}, nil
}

func (a *Agent) gatherContextRaw() string {
	var sections []string

	if a.github != nil {
		notifs, err := a.github.Notifications(50)
		if err == nil && len(notifs) > 0 {
			filtered := filterGitHub(notifs)
			summary := formatFiltered(filtered)
			sections = append(sections, "GITHUB:")
			sections = append(sections, summary)
		}
	}

	if a.gmail != nil {
		msgs, err := a.gmail.Unread(20)
		if err == nil && len(msgs) > 0 {
			sections = append(sections, fmt.Sprintf("GMAIL (%d unread):", len(msgs)))
			for _, m := range msgs {
				sections = append(sections, fmt.Sprintf("  %s", m.Snippet))
			}
		}
	}

	if a.calendar != nil {
		events, err := a.calendar.Today()
		if err == nil && len(events) > 0 {
			sections = append(sections, fmt.Sprintf("CALENDAR (%d events today):", len(events)))
			for _, e := range events {
				sections = append(sections, fmt.Sprintf("  %s at %s", e.Summary, e.Start.Format("15:04")))
			}
		}
	}

	if len(sections) == 0 {
		return "no services connected."
	}

	return strings.Join(sections, "\n")
}

func (a *Agent) gatherContext() (string, error) {
	return a.gatherContextRaw(), nil
}
