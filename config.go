package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host        string
	HttpPort    int  `toml:"http_port"`
	HttpsPort   int  `toml:"https_port"`
	EnableHTTPS bool `toml:"enable_https"`
}

type DatabaseConfig struct {
	User     string
	Password string
	Name     string
}

func LoadConfig(filename string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return config, err
	}
	return config, nil
}

func SaveConfig(filename string, config Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}
