package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	LLM struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
		Model   string `json:"model"`
	} `json:"llm"`
	Connectors map[string]struct {
		Token string `json:"token"`
	} `json:"connectors"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pulse", "config.json")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			c := &Config{}
			c.Connectors = make(map[string]struct {
				Token string `json:"token"`
			})
			c.LLM.BaseURL = "https://api.openai.com/v1"
			c.LLM.Model = "gpt-4o-mini"
			return c, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Connectors == nil {
		c.Connectors = make(map[string]struct {
			Token string `json:"token"`
		})
	}
	if c.LLM.BaseURL == "" {
		c.LLM.BaseURL = "https://api.openai.com/v1"
	}
	if c.LLM.Model == "" {
		c.LLM.Model = "gpt-4o-mini"
	}
	return &c, nil
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = DefaultPath()
	}
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0700)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) SetConnector(name, token string) {
	c.Connectors[name] = struct {
		Token string `json:"token"`
	}{Token: token}
}

func (c *Config) GetConnector(name string) string {
	if conn, ok := c.Connectors[name]; ok {
		return conn.Token
	}
	return ""
}

func (c *Config) RemoveConnector(name string) {
	delete(c.Connectors, name)
}
