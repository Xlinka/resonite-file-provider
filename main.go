package main

import (
	"fmt"
	"net/http"
	"resonite-file-provider/assethost"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"resonite-file-provider/query"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Database database.Config
	Server   ServerConfig
}

type ServerConfig struct {
	Host       string
	Port       int
	ItemsPath  string
	AssetsPath string
}

func main() {
	var config Config
	toml.DecodeFile("config.toml", &config)

	database.Connect(config.Database)
	defer database.Db.Close()

	query.AddSearchListeners()
	authentication.AddAuthListeners()
	assethost.AddAssetListeners(config.Server.AssetsPath)
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	http.ListenAndServe(addr, nil)
}
