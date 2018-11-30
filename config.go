package main

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type config struct {
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
	Target          string   `yaml:"target"`
	IgnoredGuilds   []string `yaml:"ignoredGuilds"`
	IgnoredChannels []string `yaml:"ignoredChannels"`
}

func loadConfig(filename string) (*config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	cfg := config{}
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML: %v", err)
	}

	return &cfg, nil
}
