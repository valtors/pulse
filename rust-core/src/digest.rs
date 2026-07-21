use crate::filter::{self, FilteredItem, Notification};
use crate::llm::Client;
use crate::memory::Store;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct DigestResult {
    pub urgent: Vec<FilteredItem>,
    pub important: Vec<FilteredItem>,
    pub noise_count: usize,
    pub summary: String,
    pub ai_summary: Option<String>,
}

pub fn build_digest(
    notifs: &[Notification],
    store: &Store,
    llm: &Client,
    use_ai: bool,
) -> DigestResult {
    let (urgent, important, noise_count) = filter::filter_all(notifs);

    let summary = format_summary(&urgent, &important, noise_count);

    let ai_summary = if use_ai && llm.has_key() {
        let memories = store.all().unwrap_or_default();
        let mem_str = memories
            .iter()
            .filter(|m| m.category == "user" || m.category == "config")
            .map(|m| format!("- {}: {}", m.key, m.value))
            .collect::<Vec<_>>()
            .join("\n");

        let system = format!(
            "you are pulse. a personal ai agent that lives on the user's machine.\n\n\
             live data:\n{}\n\n\
             memory:\n{}\n\n\
             rules:\n\
             - be direct. short sentences. no fluff.\n\
             - summarize what the user missed. group by service.\n\
             - highlight what needs attention. ignore noise.\n\
             - the user is not technical. speak plainly.",
            summary, mem_str
        );

        match llm.complete(&system, "what did i miss?") {
            Ok(resp) => Some(resp),
            Err(_) => None,
        }
    } else {
        None
    };

    DigestResult {
        urgent,
        important,
        noise_count,
        summary,
        ai_summary,
    }
}

fn format_summary(urgent: &[FilteredItem], important: &[FilteredItem], noise: usize) -> String {
    let mut sections = Vec::new();

    if !urgent.is_empty() {
        sections.push(format_priority(urgent, "URGENT"));
    }

    if !important.is_empty() {
        sections.push(format_priority(important, "NEEDS ATTENTION"));
    }

    if noise > 0 {
        sections.push(format!(
            "\nNOISE: {} items filtered (ci, status checks, automated noise)",
            noise
        ));
    }

    if sections.is_empty() {
        return "you're caught up. nothing to report.".to_string();
    }

    sections.join("\n")
}

fn format_priority(items: &[FilteredItem], label: &str) -> String {
    let groups = filter::group_by_repo(items);
    let mut lines = vec![format!("\n{}:", label)];

    let mut repos: Vec<_> = groups.keys().collect();
    repos.sort();

    for repo in repos {
        lines.push(format!("  {}", repo));
        for item in &groups[repo] {
            lines.push(format!(
                "    [{}] {} - {}",
                item.item_type, item.title, item.reason
            ));
        }
    }

    lines.join("\n")
}
