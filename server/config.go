package server

import (
	"encoding/json"
	"os"
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
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}
