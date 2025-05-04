package database

import (
	"database/sql"
	"fmt"
	"resonite-file-provider/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB

func Connect() {
	cfg := config.GetConfig().Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	
	// Maximum number of connection attempts
	maxRetries := 30
	retryInterval := 2 // seconds
	
	var db *sql.DB
	var err error
	
	// Open connection to the database
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err) // Still panic if we can't even create the DB object
	}
	
	// Try to connect with retries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = db.Ping()
		if err == nil {
			break // Successfully connected
		}
		
		// Log attempt and error
		fmt.Printf("Database connection attempt %d/%d failed: %v\n", 
			attempt, maxRetries, err)
		
		// Last attempt?
		if attempt == maxRetries {
			panic(fmt.Errorf("failed to connect to database after %d attempts: %w", 
				maxRetries, err))
		}
		
		// Wait before retrying - simple now that we've imported time
		fmt.Printf("Waiting %d seconds before retry...\n", retryInterval)
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
	
	Db = db
	fmt.Println("Successfully connected to database!")
}
