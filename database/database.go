package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	User     string
	Password string
	Host     string
	Port     int
     	Name     string
}

var Db *sql.DB
func Connect(cfg Config) {
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
