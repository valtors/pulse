use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
struct Message {
    role: String,
    content: String,
}

#[derive(Debug, Serialize)]
struct CompletionRequest {
    model: String,
    messages: Vec<Message>,
}

#[derive(Debug, Deserialize)]
struct CompletionResponse {
    choices: Vec<Choice>,
}

#[derive(Debug, Deserialize)]
struct Choice {
    message: MessageContent,
}

#[derive(Debug, Deserialize)]
struct MessageContent {
    content: String,
}

pub struct Client {
    base_url: String,
    api_key: String,
    model: String,
}

impl Client {
    pub fn new(base_url: &str, api_key: &str, model: &str) -> Self {
        Client {
            base_url: if base_url.is_empty() {
                "https://api.openai.com/v1".to_string()
            } else {
                base_url.to_string()
            },
            api_key: api_key.to_string(),
            model: if model.is_empty() {
                "gpt-4o-mini".to_string()
            } else {
                model.to_string()
            },
        }
    }

    pub fn has_key(&self) -> bool {
        !self.api_key.is_empty()
    }

    pub fn complete(&self, system: &str, user: &str) -> Result<String, String> {
        if self.api_key.is_empty() {
            return Err("no api key set".to_string());
        }

        let req = CompletionRequest {
            model: self.model.clone(),
            messages: vec![
                Message {
                    role: "system".to_string(),
                    content: system.to_string(),
                },
                Message {
                    role: "user".to_string(),
                    content: user.to_string(),
                },
            ],
        };

        let body = serde_json::to_string(&req).map_err(|e| e.to_string())?;
        let url = format!("{}/chat/completions", self.base_url);

        let response = reqwest::blocking::Client::builder()
            .timeout(std::time::Duration::from_secs(60))
            .build()
            .map_err(|e| e.to_string())?
            .post(&url)
            .header("Content-Type", "application/json")
            .header("Authorization", format!("Bearer {}", self.api_key))
            .body(body)
            .send()
            .map_err(|e| format!("request: {}", e))?;

        if !response.status().is_success() {
            let status = response.status();
            let text = response.text().unwrap_or_default();
            return Err(format!("llm error {}: {}", status, text));
        }

        let result: CompletionResponse = response.json().map_err(|e| format!("decode: {}", e))?;

        if result.choices.is_empty() {
            return Err("no choices returned".to_string());
        }

        Ok(result.choices[0].message.content.clone())
    }
}
