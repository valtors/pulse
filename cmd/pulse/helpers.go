package main

import (
	"github.com/valtors/pulse/internal/config"
)

func loadConfig() *config.Config {
	cfg, err := config.Load("")
	if err != nil {
		return &config.Config{}
	}
	return cfg
}
