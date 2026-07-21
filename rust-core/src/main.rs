use pulse_core::{digest, llm, memory, Notification};
use std::env;

fn main() {
    let args: Vec<String> = env::args().collect();
    let mut data_dir = format!("{}/.pulse", env::var("HOME").unwrap_or_default());

    let mut i = 1;
    while i < args.len() {
        if args[i] == "--data" && i + 1 < args.len() {
            data_dir = args[i + 1].clone();
            i += 2;
        } else {
            break;
        }
    }

    let cmd = if i < args.len() { args[i].as_str() } else { "help" };
    let cmd_args = &args[i + 1..];

    match cmd {
        "connect" => cmd_connect(&data_dir, cmd_args),
        "ask" => cmd_ask(&data_dir, cmd_args),
        "digest" => cmd_digest(&data_dir, cmd_args),
        "remember" => cmd_remember(&data_dir, cmd_args),
        "forget" => cmd_forget(&data_dir, cmd_args),
        "memory" => cmd_memory(&data_dir, cmd_args),
        _ => {
            eprintln!("pulse-core: unknown command {}", cmd);
            std::process::exit(1);
        }
    }
}

fn open_store(data_dir: &str) -> memory::Store {
    let db_path = format!("{}/pulse.db", data_dir);
    memory::Store::open(&db_path).unwrap_or_else(|e| {
        eprintln!("memory error: {}", e);
        std::process::exit(1);
    })
}

fn load_config(data_dir: &str) -> Config {
    let path = format!("{}/config.json", data_dir);
    Config::load(&path).unwrap_or_default()
}

fn make_llm(cfg: &Config) -> llm::Client {
    llm::Client::new(&cfg.llm.base_url, &cfg.llm.api_key, &cfg.llm.model)
}

fn cmd_connect(data_dir: &str, args: &[String]) {
    if args.len() < 2 {
        eprintln!("usage: pulse-core connect <service> <token>");
        std::process::exit(1);
    }
    let service = &args[0];
    let token = &args[1];

    match service.as_str() {
        "github" => {
            let resp = reqwest::blocking::Client::builder()
                .timeout(std::time::Duration::from_secs(10))
                .build()
                .unwrap()
                .get("https://api.github.com/user")
                .header("Authorization", format!("token {}", token))
                .header("Accept", "application/vnd.github+json")
                .send();
            match resp {
                Ok(r) if r.status().is_success() => {
                    println!("{{\"status\":\"connected\"}}");
                }
                Ok(r) => {
                    eprintln!("github auth failed: {}", r.status());
                    std::process::exit(1);
                }
                Err(e) => {
                    eprintln!("connect error: {}", e);
                    std::process::exit(1);
                }
            }
        }
        _ => {
            eprintln!("unknown service: {}", service);
            std::process::exit(1);
        }
    }

    let store = open_store(data_dir);
    let _ = store.remember(&format!("{}_connected", service), "connected", "config");
}

fn cmd_digest(data_dir: &str, _args: &[String]) {
    let cfg = load_config(data_dir);
    let store = open_store(data_dir);
    let llm = make_llm(&cfg);

    let token = cfg.get_connector("github");
    if token.is_empty() {
        println!(
            "{}",
            r#"{"summary":"no services connected. run: pulse connect github <token>"}"#
        );
        return;
    }

    let notifs = match fetch_github(&token) {
        Ok(n) => n,
        Err(e) => {
            println!(r#"{{"summary":"github error: {}"}}"#, e);
            return;
        }
    };

    let result = digest::build_digest(&notifs, &store, &llm, llm.has_key());
    let json = serde_json::to_string(&result)
        .unwrap_or_else(|e| format!(r#"{{"summary":"serialize error: {}"}}"#, e));
    println!("{}", json);
}

fn cmd_ask(data_dir: &str, args: &[String]) {
    if args.is_empty() {
        eprintln!("usage: pulse-core ask <question>");
        std::process::exit(1);
    }

    let question = args.join(" ");
    let cfg = load_config(data_dir);
    let store = open_store(data_dir);
    let llm = make_llm(&cfg);
    let lower = question.to_lowercase();

    if lower.starts_with("remember ") {
        let parts: Vec<&str> = question.splitn(3, ' ').collect();
        if parts.len() >= 3 {
            let _ = store.remember(parts[1], parts[2], "user");
            println!(
                r#"{{"action":"remember","detail":"stored: {} = {}"}}"#,
                parts[1], parts[2]
            );
            return;
        }
    }

    if lower.starts_with("forget ") {
        let key = lower.trim_start_matches("forget ").trim();
        let _ = store.forget(key);
        println!(r#"{{"action":"forget","detail":"forgot: {}"}}"#, key);
        return;
    }

    if lower.contains("what did i miss") || lower.contains("summary") || lower.contains("digest") {
        let token = cfg.get_connector("github");
        if token.is_empty() {
            println!(r#"{{"detail":"no services connected. run: pulse connect github <token>"}}"#);
            return;
        }
        let notifs = match fetch_github(&token) {
            Ok(n) => n,
            Err(e) => {
                println!(r#"{{"detail":"github error: {}"}}"#, e);
                return;
            }
        };
        let result = digest::build_digest(&notifs, &store, &llm, llm.has_key());
        let json = serde_json::to_string(&result).unwrap_or_default();
        println!("{}", json);
        return;
    }

    if lower.contains("what do you know") || lower.contains("what do you remember") {
        let memories = store.all().unwrap_or_default();
        let json = serde_json::to_string(&memories).unwrap_or_else(|_| "[]".to_string());
        println!("{}", json);
        return;
    }

    if llm.has_key() {
        let token = cfg.get_connector("github");
        let mut context = String::new();
        if !token.is_empty() {
            if let Ok(notifs) = fetch_github(&token) {
                let result = digest::build_digest(&notifs, &store, &llm, false);
                context = result.summary;
            }
        }
        let memories = store.all().unwrap_or_default();
        let mem_str = memories
            .iter()
            .filter(|m| m.category == "user" || m.category == "config")
            .map(|m| format!("- {}: {}", m.key, m.value))
            .collect::<Vec<_>>()
            .join("\n");
        let system = format!(
            "you are pulse. a personal ai agent on the user's machine.\n\n\
             live data:\n{}\n\n\
             memory:\n{}\n\n\
             rules:\n- be direct. short sentences. no fluff.\n\
             - the user is not technical. speak plainly.",
            context, mem_str
        );
        match llm.complete(&system, &question) {
            Ok(resp) => {
                let _ = store.remember("last_interaction", &question, "history");
                println!(
                    r#"{{"action":"respond","detail":"{}"}}"#,
                    resp.replace('"', "'")
                );
            }
            Err(e) => {
                println!(r#"{{"action":"error","detail":"{}"}}"#, e);
            }
        }
        return;
    }

    println!(
        r#"{{"detail":"connect an llm to enable full ai. or try: what did i miss, remember X, what do you know"}}"#
    );
}

fn cmd_remember(data_dir: &str, args: &[String]) {
    if args.len() < 2 {
        eprintln!("usage: pulse-core remember <key> <value>");
        std::process::exit(1);
    }
    let store = open_store(data_dir);
    let category = if args.len() > 2 { args[2].as_str() } else { "user" };
    let _ = store.remember(&args[0], &args[1], category);
    println!(r#"{{"status":"stored", "key":"{}"}}"#, args[0]);
}

fn cmd_forget(data_dir: &str, args: &[String]) {
    if args.is_empty() {
        eprintln!("usage: pulse-core forget <key>");
        std::process::exit(1);
    }
    let store = open_store(data_dir);
    let _ = store.forget(&args[0]);
    println!(r#"{{"status":"forgot", "key":"{}"}}"#, args[0]);
}

fn cmd_memory(data_dir: &str, _args: &[String]) {
    let store = open_store(data_dir);
    let memories = store.all().unwrap_or_default();
    let json = serde_json::to_string(&memories).unwrap_or_else(|_| "[]".to_string());
    println!("{}", json);
}

fn fetch_github(token: &str) -> Result<Vec<Notification>, String> {
    let resp = reqwest::blocking::Client::builder()
        .timeout(std::time::Duration::from_secs(15))
        .build()
        .map_err(|e| e.to_string())?
        .get("https://api.github.com/notifications?per_page=50")
        .header("Authorization", format!("token {}", token))
        .header("Accept", "application/vnd.github+json")
        .send()
        .map_err(|e| e.to_string())?;

    if !resp.status().is_success() {
        return Err(format!("github status: {}", resp.status()));
    }

    #[derive(serde::Deserialize)]
    struct GhNotif {
        reason: String,
        subject: GhSubject,
        repository: GhRepo,
    }
    #[derive(serde::Deserialize)]
    struct GhSubject {
        title: String,
        #[serde(rename = "type")]
        ntype: String,
    }
    #[derive(serde::Deserialize)]
    struct GhRepo {
        full_name: String,
    }

    let notifs: Vec<GhNotif> = resp.json().map_err(|e| e.to_string())?;
    Ok(notifs
        .into_iter()
        .map(|n| Notification {
            source: "github".to_string(),
            repo: n.repository.full_name,
            title: n.subject.title,
            ntype: n.subject.ntype,
            reason: n.reason,
        })
        .collect())
}

#[derive(serde::Serialize, serde::Deserialize, Default)]
struct Config {
    llm: ConfigLLM,
    connectors: std::collections::HashMap<String, ConfigConnector>,
}

#[derive(serde::Serialize, serde::Deserialize, Default)]
struct ConfigLLM {
    base_url: String,
    api_key: String,
    model: String,
}

#[derive(serde::Serialize, serde::Deserialize, Default)]
struct ConfigConnector {
    token: String,
}

impl Config {
    fn load(path: &str) -> Result<Self, String> {
        let data = std::fs::read_to_string(path).map_err(|e| e.to_string())?;
        let mut cfg: Config = serde_json::from_str(&data).map_err(|e| e.to_string())?;
        if cfg.llm.base_url.is_empty() {
            cfg.llm.base_url = "https://api.openai.com/v1".to_string();
        }
        if cfg.llm.model.is_empty() {
            cfg.llm.model = "gpt-4o-mini".to_string();
        }
        Ok(cfg)
    }

    fn get_connector(&self, name: &str) -> String {
        self.connectors
            .get(name)
            .map(|c| c.token.clone())
            .unwrap_or_default()
    }
}
