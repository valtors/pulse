package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/valtors/pulse/internal/connect"
	"github.com/valtors/pulse/internal/memory"
)

type Agent struct {
	mem        *memory.Store
	github     *connect.GitHubConnector
	gmail      *connect.GmailConnector
	calendar   *connect.CalendarConnector
}

func New(mem *memory.Store) *Agent {
	return &Agent{mem: mem}
}

func (a *Agent) ConnectGitHub(token string) error {
	a.github = &connect.GitHubConnector{}
	if err := a.github.Connect(token); err != nil {
		return err
	}
	user := a.githubUser()
	a.mem.Remember("github_user", user, "config")
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

type Summary struct {
	Source    string    `json:"source"`
	Count     int       `json:"count"`
	Items     []string  `json:"items"`
	Generated time.Time `json:"generated"`
}

func (a *Agent) WhatDidIMiss() ([]Summary, error) {
	var summaries []Summary

	if a.github != nil {
		notifs, err := a.github.Notifications(50)
		if err == nil && len(notifs) > 0 {
			s := Summary{
				Source:    "github",
				Count:     len(notifs),
				Generated: time.Now(),
			}
			for _, n := range notifs {
				s.Items = append(s.Items, fmt.Sprintf("[%s] %s - %s", n.Repository.FullName, n.Subject.Type, n.Subject.Title))
			}
			summaries = append(summaries, s)
		}
	}

	if a.gmail != nil {
		msgs, err := a.gmail.Unread(20)
		if err == nil && len(msgs) > 0 {
			s := Summary{
				Source:    "gmail",
				Count:     len(msgs),
				Generated: time.Now(),
			}
			for _, m := range msgs {
				s.Items = append(s.Items, m.Snippet)
			}
			summaries = append(summaries, s)
		}
	}

	if a.calendar != nil {
		events, err := a.calendar.Today()
		if err == nil && len(events) > 0 {
			s := Summary{
				Source:    "calendar",
				Count:     len(events),
				Generated: time.Now(),
			}
			for _, e := range events {
				s.Items = append(s.Items, fmt.Sprintf("%s at %s", e.Summary, e.Start.Format("15:04")))
			}
			summaries = append(summaries, s)
		}
	}

	return summaries, nil
}

type Task struct {
	Input    string
	Source   string
	Action   string
	Detail   string
}

func (a *Agent) Do(input string) (*Task, error) {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "what did i miss") || strings.Contains(lower, "what'd i miss") || strings.Contains(lower, "summary") {
		summaries, err := a.WhatDidIMiss()
		if err != nil {
			return nil, err
		}
		out, _ := json.MarshalIndent(summaries, "", "  ")
		a.mem.Remember("last_summary", string(out), "history")
		return &Task{
			Input:  input,
			Source: "aggregator",
			Action: "summarize",
			Detail: string(out),
		}, nil
	}

	if strings.Contains(lower, "remember ") || strings.Contains(lower, "remember that ") || strings.Contains(lower, "note that ") {
		parts := strings.SplitN(input, " ", 3)
		if len(parts) >= 3 {
			key := parts[1]
			value := parts[2]
			a.mem.Remember(key, value, "user")
			return &Task{
				Input:  input,
				Source: "memory",
				Action: "remember",
				Detail: fmt.Sprintf("stored: %s = %s", key, value),
			}, nil
		}
	}

	if strings.Contains(lower, "what do you know") || strings.Contains(lower, "what do you remember") {
		memories, err := a.mem.All()
		if err != nil {
			return nil, err
		}
		out, _ := json.MarshalIndent(memories, "", "  ")
		return &Task{
			Input:  input,
			Source: "memory",
			Action: "recall",
			Detail: string(out),
		}, nil
	}

	return &Task{
		Input:  input,
		Source: "agent",
		Action: "unknown",
		Detail: "i don't know how to do that yet. connect more services or teach me.",
	}, nil
}

func (a *Agent) githubUser() string {
	return "connected"
}

func (a *Agent) Connected() []string {
	var connected []string
	if a.github != nil {
		connected = append(connected, "github")
	}
	if a.gmail != nil {
		connected = append(connected, "gmail")
	}
	if a.calendar != nil {
		connected = append(connected, "calendar")
	}
	return connected
}

func (a *Agent) Memory() *memory.Store {
	return a.mem
}
