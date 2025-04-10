package main

import (
	"resonite-file-provider/assethost"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"resonite-file-provider/query"
)

func main() {
	database.Connect()
	defer database.Db.Close()

	query.AddSearchListeners()
	authentication.AddAuthListeners()
	assethost.StartHost()
}
