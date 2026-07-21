package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		serve()
		return
	}

	switch os.Args[1] {
	case "version", "-v", "--version":
		fmt.Printf("pulse v%s\n", version)
	case "serve":
		serve()
	case "connect":
		if len(os.Args) < 4 {
			fmt.Println("usage: pulse connect <service> <token>")
			fmt.Println("services: github, gmail, calendar")
			os.Exit(1)
		}
		connectCmd(os.Args[2], os.Args[3])
	case "disconnect":
		if len(os.Args) < 3 {
			fmt.Println("usage: pulse disconnect <service>")
			os.Exit(1)
		}
		disconnectCmd(os.Args[2])
	case "ask":
		if len(os.Args) < 3 {
			fmt.Println("usage: pulse ask <question>")
			fmt.Println("try: pulse ask \"what did i miss\"")
			os.Exit(1)
		}
		askCmd(os.Args[2])
	case "digest":
		digestCmd()
	case "remember":
		if len(os.Args) < 4 {
			fmt.Println("usage: pulse remember <key> <value>")
			os.Exit(1)
		}
		rememberCmd(os.Args[2], os.Args[3])
	case "forget":
		if len(os.Args) < 3 {
			fmt.Println("usage: pulse forget <key>")
			os.Exit(1)
		}
		forgetCmd(os.Args[2])
	case "memory":
		memoryCmd()
	case "config":
		if len(os.Args) < 3 {
			fmt.Println("usage: pulse config llm <base_url> <api_key> <model>")
			os.Exit(1)
		}
		if os.Args[2] == "llm" {
			if len(os.Args) < 6 {
				fmt.Println("usage: pulse config llm <base_url> <api_key> <model>")
				os.Exit(1)
			}
			configLLMCmd(os.Args[3], os.Args[4], os.Args[5])
		}
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`pulse v%s - connect everything. your ai does the rest.

commands:
  pulse                    start the web server
  pulse serve              same thing
  pulse connect <svc> <token>   connect a service (github, gmail, calendar)
  pulse disconnect <svc>        disconnect a service
  pulse ask <question>          ask pulse something
  pulse digest                  get your filtered summary right now
  pulse remember <key> <val>    store something in pulse memory
  pulse forget <key>            remove from memory
  pulse memory                  show everything pulse remembers
  pulse config llm <url> <key> <model>   configure ai
  pulse version                 show version
  pulse help                    this message

examples:
  pulse connect github ghp_xxxx
  pulse digest
  pulse ask "what did i miss"
  pulse remember focus "ship pulse v1"

web: http://localhost:9090
`, version)
}

func dataDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".pulse")
	os.MkdirAll(dir, 0700)
	return dir
}

func rustCore() (string, error) {
	home, _ := os.UserHomeDir()
	bin := filepath.Join(home, ".pulse", "pulse-core")
	if _, err := os.Stat(bin); err == nil {
		return bin, nil
	}
	candidate := filepath.Join("rust-core", "target", "release", "pulse-core")
	if abs, err := filepath.Abs(candidate); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}
	return "", fmt.Errorf("pulse-core not found. build with: cd rust-core && cargo build --release")
}

func callRust(args ...string) ([]byte, error) {
	bin, err := rustCore()
	if err != nil {
		return nil, err
	}
	dd := dataDir()
	full := append([]string{"--data", dd}, args...)
	cmd := exec.Command(bin, full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}

func callRustJSON(args ...string) (map[string]interface{}, error) {
	out, err := callRust(args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, string(out))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse: %s", string(out))
	}
	return result, nil
}
