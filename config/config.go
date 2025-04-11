package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

type ServerConfig struct {
	Host       string
	Port       int
	ItemsPath  string
	AssetsPath string
}

type DatabaseConfig struct {
	User     string
	Password string
	Host     string
	Port     int
     	Name     string
}

var config Config

func GetConfig() Config {
	if config == (Config{}) {
		toml.DecodeFile("config.toml", &config)
	}
	return config
}

