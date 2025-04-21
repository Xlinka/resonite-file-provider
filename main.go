package main

import (
	"crypto/tls"
	"fmt"
	"log"
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

	query.AddSearchListeners()    // AnimX API endpoints for VR client
	query.AddJSONAPIListeners()   // JSON API endpoints for web interface
	authentication.AddAuthListeners()
	assethost.AddAssetListeners()
	upload.AddListeners()

	addr := fmt.Sprintf(":%d", config.GetConfig().Server.Port)

	server := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	go upload.StartWebServer()
	log.Fatal(server.ListenAndServeTLS("certs/fullchain.pem", "certs/privkey.pem"))
}