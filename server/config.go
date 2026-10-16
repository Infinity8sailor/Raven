package server

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	Port    int    `json:"port"`
	DataDir string `json:"data_dir"`
	WebDir  string `json:"web_dir"`
}

func LoadConfig(path string) (*Config, error) {
	c := &Config{
		Port:    9471,
		DataDir: "data",
		WebDir:  "web",
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(b, c); err != nil {
			return nil, err
		}
	}

	// Environment variable overrides (RAVEN_PORT or PORT)
	if envPort := os.Getenv("RAVEN_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			c.Port = p
		}
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			c.Port = p
		}
	}

	return c, nil
}
