package main

import (
	"database/sql"
	"resonite-file-provider/assethost"
	"resonite-file-provider/database"
)
var db *sql.DB
func main() {
	db = database.Connect()
	assethost.StartHost()
	defer db.Close()
}
