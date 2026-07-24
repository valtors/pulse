package agent

import (
	"fmt"
	"strings"

	"github.com/valtors/pulse/internal/connect"
)

type Priority int

const (
	PriorityUrgent Priority = iota
	PriorityImportant
	PriorityNoise
)

type FilteredItem struct {
	Source   string `json:"source"`
	Repo     string `json:"repo"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Priority string `json:"priority"`
	Reason   string `json:"reason"`
}

func priorityName(p Priority) string {
	switch p {
	case PriorityUrgent:
		return "urgent"
	case PriorityImportant:
		return "important"
	default:
		return "noise"
	}
}

func filterGitHub(notifs []connect.GitHubNotification) []FilteredItem {
	var items []FilteredItem
	for _, n := range notifs {
		item := FilteredItem{
			Source: "github",
			Repo:   n.Repository.FullName,
			Type:   n.Subject.Type,
			Title:  n.Subject.Title,
		}

		switch n.Subject.Type {
		case "PullRequest":
			if n.Reason == "review_requested" || n.Reason == "mention" {
				item.Priority = priorityName(PriorityUrgent)
				item.Reason = "you were mentioned or asked to review"
			} else if n.Reason == "author" {
				item.Priority = priorityName(PriorityUrgent)
				item.Reason = "your PR has activity"
			} else {
				item.Priority = priorityName(PriorityImportant)
				item.Reason = "pr activity"
			}
		case "Issue":
			if n.Reason == "assign" || n.Reason == "mention" {
				item.Priority = priorityName(PriorityUrgent)
				item.Reason = "you were assigned or mentioned"
			} else if n.Reason == "author" {
				item.Priority = priorityName(PriorityImportant)
				item.Reason = "your issue has activity"
			} else {
				item.Priority = priorityName(PriorityImportant)
				item.Reason = "issue activity"
			}
		case "Discussion":
			if n.Reason == "mention" {
				item.Priority = priorityName(PriorityUrgent)
				item.Reason = "you were mentioned"
			} else {
				item.Priority = priorityName(PriorityImportant)
				item.Reason = "discussion activity"
			}
		case "CheckSuite":
			item.Priority = priorityName(PriorityNoise)
			item.Reason = "ci status"
		case "Release":
			item.Priority = priorityName(PriorityImportant)
			item.Reason = "new release"
		default:
			item.Priority = priorityName(PriorityNoise)
			item.Reason = n.Reason
		}

		items = append(items, item)
	}
	return items
}

func groupByRepo(items []FilteredItem) map[string][]FilteredItem {
	groups := make(map[string][]FilteredItem)
	for _, item := range items {
		groups[item.Repo] = append(groups[item.Repo], item)
	}
	return groups
}

func countPriority(items []FilteredItem) map[string]int {
	counts := map[string]int{"urgent": 0, "important": 0, "noise": 0}
	for _, item := range items {
		counts[item.Priority]++
	}
	return counts
}

func formatFiltered(items []FilteredItem) string {
	if len(items) == 0 {
		return "nothing to report. you're caught up."
	}

	counts := countPriority(items)
	var sections []string

	if counts["urgent"] > 0 {
		sections = append(sections, formatPriority(items, PriorityUrgent, "URGENT"))
	}
	if counts["important"] > 0 {
		sections = append(sections, formatPriority(items, PriorityImportant, "NEEDS ATTENTION"))
	}

	noiseCount := counts["noise"]
	if noiseCount > 0 {
		sections = append(sections, fmt.Sprintf("\nNOISE: %d items filtered (ci failures, status checks, automated noise)", noiseCount))
	}

	return strings.Join(sections, "\n")
}

func formatPriority(items []FilteredItem, p Priority, label string) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("\n%s:", label))
	groups := groupByRepo(items)
	for repo, repoItems := range groups {
		if !hasPriority(repoItems, p) {
			continue
		}
		lines = append(lines, fmt.Sprintf("  %s", repo))
		for _, item := range repoItems {
			if item.Priority != priorityName(p) {
				continue
			}
			lines = append(lines, fmt.Sprintf("    [%s] %s - %s", item.Type, item.Title, item.Reason))
		}
	}
	return strings.Join(lines, "\n")
}

func hasPriority(items []FilteredItem, p Priority) bool {
	for _, item := range items {
		if item.Priority == priorityName(p) {
			return true
		}
	}
	return false
}
