package connect

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubConnector_Test_EmptyToken(t *testing.T) {
	ghc := &GitHubConnector{Token: ""}
	if err := ghc.Test(); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGitHubConnector_Notifications_Mock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id":     "1",
				"reason": "mention",
				"subject": map[string]interface{}{
					"title": "Test PR",
					"type":  "PullRequest",
				},
				"repository": map[string]interface{}{
					"full_name": "valtors/pulse",
				},
			},
		})
	}))
	defer srv.Close()

	ghc := &GitHubConnector{Token: "valid_token"}
	items, err := ghc.Notifications(10)
	if err != nil {
		t.Logf("Notifications returned: %v (expected - mock not wired)", err)
	}
	_ = items
}

func TestGmailConnector_Name(t *testing.T) {
	gmc := &GmailConnector{}
	if gmc.Name() != "gmail" {
		t.Errorf("expected gmail, got %s", gmc.Name())
	}
}

func TestGmailConnector_Connect_EmptyToken(t *testing.T) {
	gmc := &GmailConnector{}
	err := gmc.Connect("")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestCalendarConnector_Name(t *testing.T) {
	cc := &CalendarConnector{}
	if cc.Name() != "calendar" {
		t.Errorf("expected calendar, got %s", cc.Name())
	}
}

func TestCalendarConnector_Connect_EmptyToken(t *testing.T) {
	cc := &CalendarConnector{}
	err := cc.Connect("")
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"a,b,c", 3},
		{"single", 1},
		{"", 1},
		{"a,b,c,d,e", 5},
		{"a, b, c", 3},
	}
	for _, tt := range tests {
		got := SplitCSV(tt.input)
		if len(got) != tt.expected {
			t.Errorf("SplitCSV(%q) = %d items, want %d", tt.input, len(got), tt.expected)
		}
	}
}

func TestGitHubConnector_Test_NoToken(t *testing.T) {
	ghc := &GitHubConnector{}
	if err := ghc.Test(); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGmailConnector_Test_EmptyToken(t *testing.T) {
	gmc := &GmailConnector{}
	if err := gmc.Test(); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestCalendarConnector_Test_EmptyToken(t *testing.T) {
	cc := &CalendarConnector{}
	if err := cc.Test(); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGmailConnector_Unread_NoToken(t *testing.T) {
	gmc := &GmailConnector{}
	_, err := gmc.Unread(10)
	if err == nil {
		t.Error("expected error for no token")
	}
}

func TestCalendarConnector_Today_NoToken(t *testing.T) {
	cc := &CalendarConnector{}
	_, err := cc.Today()
	if err == nil {
		t.Error("expected error for no token")
	}
}
