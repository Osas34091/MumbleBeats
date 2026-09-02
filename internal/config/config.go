package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerAddress string   `json:"server_address"`
	ServerPort    string   `json:"server_port"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	Channel       string   `json:"channel"`
	Insecure      bool     `json:"insecure"`
	Admins        []string `json:"admins"`
}

var DefaultConfig = Config{
	ServerAddress: "localhost",
	ServerPort:    "64738",
	Username:      "MumbleBeats",
	Password:      "",
	Channel:       "Root",
	Insecure:      true,
	Admins:        []string{},
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return createDefaultConfig(filename)
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func createDefaultConfig(filename string) (*Config, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&DefaultConfig); err != nil {
		return nil, err
	}
	return &DefaultConfig, nil
}
