package gocan

import (
	"encoding/json"
	"os"
)

type Config struct {
	Interface string `json:"interface"`
	Bitrate   int    `json:"bitrate"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}