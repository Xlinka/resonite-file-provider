package main

import (
	"resonite-file-provider/assethost"
	"resonite-file-provider/database"
	"resonite-file-provider/authentication"
)
func main() {
	database.Connect()
	defer database.Db.Close()

	authentication.AddAuthListeners()
	assethost.StartHost()
}
