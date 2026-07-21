package connect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Connector interface {
	Name() string
	Connect(token string) error
	Test() error
}

type GitHubConnector struct {
	Token string
}

type Service struct {
	connectors map[string]Connector
}

func NewService() *Service {
	return &Service{
		connectors: make(map[string]Connector),
	}
}

func (s *Service) Register(c Connector) {
	s.connectors[c.Name()] = c
}

func (s *Service) Get(name string) (Connector, bool) {
	c, ok := s.connectors[name]
	return c, ok
}

func (s *Service) List() []string {
	var names []string
	for name := range s.connectors {
		names = append(names, name)
	}
	return names
}

func (g *GitHubConnector) Name() string { return "github" }

func (g *GitHubConnector) Connect(token string) error {
	g.Token = token
	return g.Test()
}

func (g *GitHubConnector) Test() error {
	if g.Token == "" {
		return fmt.Errorf("no token")
	}
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("github auth failed: %d", resp.StatusCode)
	}
	return nil
}

type GitHubNotification struct {
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Subject   struct {
		Title string `json:"title"`
		Type  string `json:"type"`
		URL   string `json:"url"`
	} `json:"subject"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (g *GitHubConnector) Notifications(limit int) ([]GitHubNotification, error) {
	if g.Token == "" {
		return nil, fmt.Errorf("not connected")
	}
	req, _ := http.NewRequest("GET", "https://api.github.com/notifications?per_page="+fmt.Sprintf("%d", limit), nil)
	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github notifications failed: %d", resp.StatusCode)
	}
	var notifications []GitHubNotification
	if err := json.NewDecoder(resp.Body).Decode(&notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

type GmailConnector struct {
	Token   string
	Refresh string
}

func (g *GmailConnector) Name() string { return "gmail" }

func (g *GmailConnector) Connect(token string) error {
	g.Token = token
	return g.Test()
}

func (g *GmailConnector) Test() error {
	if g.Token == "" {
		return fmt.Errorf("no token")
	}
	u := "https://gmail.googleapis.com/gmail/v1/users/me/profile"
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+g.Token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("gmail auth failed: %d", resp.StatusCode)
	}
	return nil
}

type GmailMessage struct {
	ID     string `json:"id"`
	Snippet string `json:"snippet"`
}

func (g *GmailConnector) Unread(limit int) ([]GmailMessage, error) {
	if g.Token == "" {
		return nil, fmt.Errorf("not connected")
	}
	u := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages?q=is:unread&maxResults=%d", limit)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+g.Token)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gmail unread failed: %d", resp.StatusCode)
	}
	var result struct {
		Messages []GmailMessage `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Messages, nil
}

type CalendarConnector struct {
	Token string
}

func (c *CalendarConnector) Name() string { return "calendar" }

func (c *CalendarConnector) Connect(token string) error {
	c.Token = token
	return c.Test()
}

func (c *CalendarConnector) Test() error {
	if c.Token == "" {
		return fmt.Errorf("no token")
	}
	u := "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("calendar auth failed: %d", resp.StatusCode)
	}
	return nil
}

type CalendarEvent struct {
	Summary     string    `json:"summary"`
	Start       time.Time `json:"start"`
	Location    string    `json:"location"`
}

func (c *CalendarConnector) Today() ([]CalendarEvent, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("not connected")
	}
	now := time.Now().UTC()
	start := now.Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	u := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events?timeMin=%s&timeMax=%s&maxResults=20&orderBy=startTime&singleEvents=true",
		url.QueryEscape(start.Format(time.RFC3339)),
		url.QueryEscape(end.Format(time.RFC3339)))
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+c.Token)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("calendar today failed: %d", resp.StatusCode)
	}
	var result struct {
		Items []CalendarEvent `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func SplitCSV(s string) []string {
	return strings.Split(s, ",")
}
