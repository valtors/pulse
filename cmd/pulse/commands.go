package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func connectCmd(service, token string) {
	cfg := loadConfig()
	cfg.SetConnector(service, token)
	if err := cfg.Save(""); err != nil {
		fmt.Printf("error saving config: %v\n", err)
		os.Exit(1)
	}
	out, err := callRust("connect", service, token)
	if err != nil {
		fmt.Printf("connect failed: %s\n", string(out))
		os.Exit(1)
	}
	var result map[string]interface{}
	json.Unmarshal(out, &result)
	if status, ok := result["status"].(string); ok && status == "connected" {
		fmt.Printf("connected %s\n", service)
	} else {
		fmt.Printf("failed: %s\n", string(out))
		os.Exit(1)
	}
}

func disconnectCmd(service string) {
	cfg := loadConfig()
	cfg.RemoveConnector(service)
	cfg.Save("")
	fmt.Printf("disconnected %s\n", service)
}

func askCmd(question string) {
	out, err := callRust("ask", question)
	if err != nil {
		fmt.Printf("%s\n", string(out))
		os.Exit(1)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		fmt.Printf("%s\n", string(out))
		return
	}
	if detail, ok := result["detail"].(string); ok {
		fmt.Println(detail)
		return
	}
	if summary, ok := result["summary"].(string); ok {
		fmt.Println(summary)
		return
	}
	fmt.Printf("%s\n", string(out))
}

func digestCmd() {
	out, err := callRust("digest")
	if err != nil {
		fmt.Printf("%s\n", string(out))
		os.Exit(1)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		fmt.Printf("%s\n", string(out))
		return
	}
	if ai, ok := result["ai_summary"].(string); ok && ai != "" {
		fmt.Println(ai)
		fmt.Println()
		fmt.Println("---")
	}
	if summary, ok := result["summary"].(string); ok {
		fmt.Println(summary)
		return
	}
	fmt.Printf("%s\n", string(out))
}

func rememberCmd(key, value string) {
	out, err := callRust("remember", key, value, "user")
	if err != nil {
		fmt.Printf("%s\n", string(out))
		os.Exit(1)
	}
	fmt.Printf("remembered: %s = %s\n", key, value)
}

func forgetCmd(key string) {
	_, err := callRust("forget", key)
	if err != nil {
		fmt.Printf("forget failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("forgot: %s\n", key)
}

func memoryCmd() {
	out, err := callRust("memory")
	if err != nil {
		fmt.Printf("%s\n", string(out))
		os.Exit(1)
	}
	var memories []map[string]interface{}
	if err := json.Unmarshal(out, &memories); err != nil {
		var single map[string]interface{}
		if err := json.Unmarshal(out, &single); err != nil {
			fmt.Printf("%s\n", string(out))
			return
		}
		memories = []map[string]interface{}{single}
	}
	if len(memories) == 0 {
		fmt.Println("nothing remembered yet.")
		return
	}
	for _, m := range memories {
		key, _ := m["key"].(string)
		val, _ := m["value"].(string)
		cat, _ := m["category"].(string)
		fmt.Printf("  %s: %s (%s)\n", key, val, cat)
	}
}

func configLLMCmd(baseURL, apiKey, model string) {
	cfg := loadConfig()
	cfg.LLM.BaseURL = baseURL
	cfg.LLM.APIKey = apiKey
	cfg.LLM.Model = model
	if err := cfg.Save(""); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("llm configured. base=%s model=%s\n", baseURL, model)
}
