package database

import (
	"database/sql"
	"fmt"
	"resonite-file-provider/config"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB
func Connect() {
	cfg := config.GetConfig().Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
	    cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}
	Db = db
	println("Connected to db")
}
