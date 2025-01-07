package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type ProtocolType int

const (
	None      ProtocolType = iota // 0
	HTTP                          // 1
	HTTPS                         // 2
	HTTPHTTPS                     // 3
)

type Config struct {
	Admin    AdminConfig
	Tokens   TokenConfig
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host      string `toml:"host"`
	HttpPort  int    `toml:"http_port"`
	HttpsPort int    `toml:"https_port"`
	Protocols int    `toml:"protocols"`
}

type TokenConfig struct {
	Discord  string `toml:"ds"`
	Telegram string `toml:"tg"`
}

type CanvasConfig struct {
	Side      int `toml:"side"`
	Timeout   int `toml:"timeout"`
	FreeSpace int `toml:"free_space"`
}

type AdminConfig struct {
	Login    string `toml:"login"`
	Password string `toml:"password"`
}

type DatabaseConfig struct {
	Name     string `toml:"name"`
	User     string `toml:"user"`
	Password string `toml:"password"`
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
