use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FilteredItem {
    pub source: String,
    pub repo: String,
    pub item_type: String,
    pub title: String,
    pub priority: Priority,
    pub reason: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Priority {
    Urgent,
    Important,
    Noise,
}

impl Priority {
    pub fn as_str(&self) -> &'static str {
        match self {
            Priority::Urgent => "urgent",
            Priority::Important => "important",
            Priority::Noise => "noise",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Notification {
    pub source: String,
    pub repo: String,
    pub title: String,
    pub ntype: String,
    pub reason: String,
}

pub fn classify(notif: &Notification) -> Priority {
    match notif.ntype.as_str() {
        "PullRequest" => match notif.reason.as_str() {
            "review_requested" | "mention" | "author" => Priority::Urgent,
            _ => Priority::Important,
        },
        "Issue" => match notif.reason.as_str() {
            "assign" | "mention" => Priority::Urgent,
            "author" => Priority::Important,
            _ => Priority::Important,
        },
        "Discussion" => match notif.reason.as_str() {
            "mention" => Priority::Urgent,
            _ => Priority::Important,
        },
        "CheckSuite" | "CheckRun" | "WorkflowRun" => Priority::Noise,
        "Release" => Priority::Important,
        _ => Priority::Noise,
    }
}

pub fn filter_all(notifs: &[Notification]) -> (Vec<FilteredItem>, Vec<FilteredItem>, usize) {
    let mut urgent = Vec::new();
    let mut important = Vec::new();
    let mut noise_count = 0;

    for n in notifs {
        let priority = classify(n);
        let item = FilteredItem {
            source: n.source.clone(),
            repo: n.repo.clone(),
            item_type: n.ntype.clone(),
            title: n.title.clone(),
            priority,
            reason: reason_text(&n.reason, &n.ntype),
        };

        match priority {
            Priority::Urgent => urgent.push(item),
            Priority::Important => important.push(item),
            Priority::Noise => noise_count += 1,
        }
    }

    (urgent, important, noise_count)
}

fn reason_text(reason: &str, ntype: &str) -> String {
    match ntype {
        "PullRequest" => match reason {
            "review_requested" | "mention" => "you were mentioned or asked to review".to_string(),
            "author" => "your PR has activity".to_string(),
            _ => "pr activity".to_string(),
        },
        "Issue" => match reason {
            "assign" | "mention" => "you were assigned or mentioned".to_string(),
            "author" => "your issue has activity".to_string(),
            _ => "issue activity".to_string(),
        },
        "Discussion" => match reason {
            "mention" => "you were mentioned".to_string(),
            _ => "discussion activity".to_string(),
        },
        "CheckSuite" | "CheckRun" | "WorkflowRun" => "ci status".to_string(),
        "Release" => "new release".to_string(),
        _ => reason.to_string(),
    }
}

pub fn group_by_repo(items: &[FilteredItem]) -> HashMap<String, Vec<&FilteredItem>> {
    let mut groups: HashMap<String, Vec<&FilteredItem>> = HashMap::new();
    for item in items {
        groups.entry(item.repo.clone()).or_default().push(item);
    }
    groups
}
