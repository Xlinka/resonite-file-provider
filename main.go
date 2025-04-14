package main

import (
	"fmt"
	"net/http"
	"resonite-file-provider/assethost"
	"resonite-file-provider/authentication"
	"resonite-file-provider/config"
	"resonite-file-provider/database"
	"resonite-file-provider/query"
	"resonite-file-provider/upload"
)


func main() {
	database.Connect()
	defer database.Db.Close()

	query.AddSearchListeners()
	authentication.AddAuthListeners()
	assethost.AddAssetListeners()
	upload.AddListeners()
	addr := fmt.Sprintf("%s:%d", config.GetConfig().Server.Host, config.GetConfig().Server.Port)
	go http.ListenAndServe(addr, nil)
	upload.StartWebServer()
}
