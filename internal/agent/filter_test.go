package agent

import (
	"testing"

	"github.com/valtors/pulse/internal/connect"
)

func mkNotif(reason, typ, title, repo string) connect.GitHubNotification {
	return connect.GitHubNotification{
		Reason: reason,
		Subject: struct {
			Title string `json:"title"`
			Type  string `json:"type"`
			URL   string `json:"url"`
		}{Title: title, Type: typ},
		Repository: struct {
			FullName string `json:"full_name"`
		}{FullName: repo},
	}
}

func TestFilterGitHubPullRequestMention(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("mention", "PullRequest", "Fix bug", "valtors/relay"),
	})
	if len(items) != 1 || items[0].Priority != "urgent" {
		t.Fatalf("expected 1 urgent, got %+v", items)
	}
}

func TestFilterGitHubPullRequestReview(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("review_requested", "PullRequest", "Review me", "user/repo"),
	})
	if items[0].Priority != "urgent" {
		t.Errorf("expected urgent, got %s", items[0].Priority)
	}
}

func TestFilterGitHubPullRequestAuthor(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("author", "PullRequest", "My PR", "user/repo"),
	})
	if items[0].Priority != "urgent" {
		t.Errorf("expected urgent, got %s", items[0].Priority)
	}
}

func TestFilterGitHubIssueAssign(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("assign", "Issue", "Bug", "user/repo"),
	})
	if items[0].Priority != "urgent" {
		t.Errorf("expected urgent, got %s", items[0].Priority)
	}
}

func TestFilterGitHubCheckSuite(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("ci_activity", "CheckSuite", "CI passed", "user/repo"),
	})
	if items[0].Priority != "noise" {
		t.Errorf("expected noise, got %s", items[0].Priority)
	}
}

func TestFilterGitHubRelease(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("release", "Release", "v1.0.0", "user/repo"),
	})
	if items[0].Priority != "important" {
		t.Errorf("expected important, got %s", items[0].Priority)
	}
}

func TestFilterGitHubEmpty(t *testing.T) {
	items := filterGitHub(nil)
	if len(items) != 0 {
		t.Errorf("expected 0, got %d", len(items))
	}
}

func TestCountPriority(t *testing.T) {
	items := []FilteredItem{
		{Priority: "urgent"}, {Priority: "urgent"},
		{Priority: "important"}, {Priority: "noise"},
	}
	c := countPriority(items)
	if c["urgent"] != 2 || c["important"] != 1 || c["noise"] != 1 {
		t.Errorf("wrong counts: %+v", c)
	}
}

func TestGroupByRepo(t *testing.T) {
	items := []FilteredItem{
		{Repo: "a/b"}, {Repo: "a/b"}, {Repo: "c/d"},
	}
	g := groupByRepo(items)
	if len(g["a/b"]) != 2 || len(g["c/d"]) != 1 {
		t.Errorf("wrong groups: %+v", g)
	}
}

func TestFormatFilteredEmpty(t *testing.T) {
	if formatFiltered(nil) != "nothing to report. you're caught up." {
		t.Error("expected catch-up message")
	}
}

func TestFormatFilteredWithUrgent(t *testing.T) {
	items := []FilteredItem{
		{Source: "github", Repo: "user/repo", Type: "PullRequest", Title: "Fix", Priority: "urgent", Reason: "mentioned"},
	}
	if formatFiltered(items) == "" {
		t.Error("expected non-empty")
	}
}

func TestPriorityName(t *testing.T) {
	if priorityName(PriorityUrgent) != "urgent" {
		t.Error("urgent")
	}
	if priorityName(PriorityImportant) != "important" {
		t.Error("important")
	}
	if priorityName(PriorityNoise) != "noise" {
		t.Error("noise")
	}
}

func TestHasPriority(t *testing.T) {
	items := []FilteredItem{{Priority: "urgent"}, {Priority: "noise"}}
	if !hasPriority(items, PriorityUrgent) {
		t.Error("expected true")
	}
	if hasPriority(items, PriorityImportant) {
		t.Error("expected false")
	}
}
