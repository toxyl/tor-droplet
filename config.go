package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIToken string `yaml:"api_token"`
	Droplet  struct {
		Size   string        `yaml:"size"`
		Region string        `yaml:"region"`
		Image  string        `yaml:"image"`
		TTL    time.Duration `yaml:"ttl"`
	} `yaml:"droplet"`
	TorOptions []string `yaml:"tor_options"`
	Ports      struct {
		Local  int `yaml:"local"`
		Remote int `yaml:"remote"`
	} `yaml:"ports"`
	DNS struct {
		Primary   string `yaml:"primary"`
		Secondary string `yaml:"secondary"`
	} `yaml:"dns"`
}

func loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
