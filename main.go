package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Verify database schema
	if err := database.InitializeSchema(); err != nil {
		log.Fatalf("Schema verification failed: %v", err)
	}

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
	
	// Check if TLS certificates exist
	if _, err := os.Stat("certs/fullchain.pem"); err == nil {
		if _, err := os.Stat("certs/privkey.pem"); err == nil {
			log.Println("Using TLS certificates")
			log.Fatal(server.ListenAndServeTLS("certs/fullchain.pem", "certs/privkey.pem"))
		}
	}
	
	// Fallback to HTTP if certificates don't exist
	log.Println("TLS certificates not found, starting with HTTP")
	log.Fatal(server.ListenAndServe())
}