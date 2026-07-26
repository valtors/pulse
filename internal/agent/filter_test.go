package agent

import (
	"strings"
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

func TestFilterGitHubMultipleItems(t *testing.T) {
	notifs := []connect.GitHubNotification{
		mkNotif("mention", "PullRequest", "PR1", "a/b"),
		mkNotif("author", "Issue", "Issue1", "c/d"),
		mkNotif("ci_activity", "CheckSuite", "CI", "e/f"),
		mkNotif("release", "Release", "v1.0", "g/h"),
	}
	items := filterGitHub(notifs)
	if len(items) != 4 {
		t.Fatalf("expected 4, got %d", len(items))
	}
	if items[0].Priority != "urgent" {
		t.Errorf("expected urgent, got %s", items[0].Priority)
	}
	if items[1].Priority != "important" {
		t.Errorf("expected important for author issue, got %s", items[1].Priority)
	}
	if items[2].Priority != "noise" {
		t.Errorf("expected noise, got %s", items[2].Priority)
	}
	if items[3].Priority != "important" {
		t.Errorf("expected important, got %s", items[3].Priority)
	}
}

func TestFormatFilteredOnlyNoise(t *testing.T) {
	items := []FilteredItem{
		{Source: "github", Repo: "x/y", Type: "CheckSuite", Title: "CI", Priority: "noise", Reason: "ci status"},
	}
	result := formatFiltered(items)
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestFormatFilteredImportantOnly(t *testing.T) {
	items := []FilteredItem{
		{Source: "github", Repo: "x/y", Type: "Release", Title: "v1.0", Priority: "important", Reason: "new release"},
	}
	result := formatFiltered(items)
	if result == "" {
		t.Error("expected non-empty")
	}
	if result == "nothing to report. you're caught up." {
		t.Error("should not be caught up")
	}
}

func TestCountPriorityAllZero(t *testing.T) {
	c := countPriority(nil)
	if c["urgent"] != 0 || c["important"] != 0 || c["noise"] != 0 {
		t.Error("expected all zeros")
	}
}

func TestGroupByRepoEmpty(t *testing.T) {
	g := groupByRepo(nil)
	if len(g) != 0 {
		t.Errorf("expected empty, got %d", len(g))
	}
}

func TestHasPriorityEmpty(t *testing.T) {
	if hasPriority(nil, PriorityUrgent) {
		t.Error("expected false for empty")
	}
}

func TestFormatPriorityEmpty(t *testing.T) {
	items := []FilteredItem{}
	result := formatPriority(items, PriorityUrgent, "URGENT")
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestFilterGitHubDiscussionMention(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("mention", "Discussion", "roadmap talk", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "urgent" {
		t.Errorf("expected urgent, got %s", items[0].Priority)
	}
}

func TestFilterGitHubDiscussionOther(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("subscribed", "Discussion", "random", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "important" {
		t.Errorf("expected important, got %s", items[0].Priority)
	}
}

func TestFilterGitHubIssueAuthor(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("author", "Issue", "bug", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "important" {
		t.Errorf("expected important, got %s", items[0].Priority)
	}
}

func TestFilterGitHubIssueDefault(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("subscribed", "Issue", "bug", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "important" {
		t.Errorf("expected important, got %s", items[0].Priority)
	}
}

func TestFilterGitHubPullRequestDefault(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("subscribed", "PullRequest", "pr", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "important" {
		t.Errorf("expected important, got %s", items[0].Priority)
	}
}

func TestFilterGitHubUnknownType(t *testing.T) {
	items := filterGitHub([]connect.GitHubNotification{
		mkNotif("whatever", "Unknown", "???", "valtors/relay"),
	})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Priority != "noise" {
		t.Errorf("expected noise, got %s", items[0].Priority)
	}
	if items[0].Reason != "whatever" {
		t.Errorf("expected reason 'whatever', got %s", items[0].Reason)
	}
}

func TestFormatFilteredAllPriorities(t *testing.T) {
	items := []FilteredItem{
		{Source: "github", Repo: "a/b", Title: "urgent", Priority: "urgent"},
		{Source: "github", Repo: "a/b", Title: "important", Priority: "important"},
		{Source: "github", Repo: "a/b", Title: "noise1", Priority: "noise"},
		{Source: "github", Repo: "a/b", Title: "noise2", Priority: "noise"},
	}
	out := formatFiltered(items)
	if !strings.Contains(out, "URGENT") {
		t.Error("expected URGENT section")
	}
	if !strings.Contains(out, "NEEDS ATTENTION") {
		t.Error("expected NEEDS ATTENTION section")
	}
	if !strings.Contains(out, "NOISE: 2") {
		t.Error("expected NOISE: 2")
	}
}
