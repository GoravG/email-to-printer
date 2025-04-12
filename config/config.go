package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ImapServer   string   `yaml:"imap_server"`
	Email        string   `yaml:"email"`
	Password     string   `yaml:"password"`
	PrinterName  string   `yaml:"printer_name"`
	Debug        bool     `yaml:"debug"`
	AllowedTypes []string `yaml:"allowed_file_types"`
	LogLevel     string   `yaml:"log_level"`
}

func Load() *Config {
	f, err := os.ReadFile("config.yaml")
	if err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		panic("Failed to parse config file: " + err.Error())
	}

	return &cfg
}
