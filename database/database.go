package database
import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
)
type Config struct {
    	Database struct {
        	User     string
        	Password string
        	Host     string
        	Port     int
   	     	Name     string
    	}
}
func Connect() *sql.DB {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
	    panic(err)
	}
	cfg := config.Database
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
	println("Connected to db")
	return db
}
