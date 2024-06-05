package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Storage Storage `json:"storage"`
	Logger  Logger  `json:"logger"`
	Bot     Bot     `json:"bot"`
}

type Storage struct {
	Driver string `json:"driver"`
	Path   string `json:"path"`
}

type Logger struct {
	Env          string `json:"env"`
	Internal     bool   `json:"internal"`
	ExternalPath string `json:"external_path"`
}

type Bot struct {
	Token         string `json:"token"`
	DebugMode     bool   `json:"debug_mode"`
	UpdateOffset  int    `json:"update_offset"`
	UpdateTimeout int    `json:"update_timeout"`
	UpdateLimit   int    `json:"update_limit"`
}

func Get(path string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&config)
	return config, err
}
