package authentication

import (
	"database/sql"
	"fmt"
	"net/http"
	"resonite-file-provider/database"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	println("Registering user:", username)
	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	var exists bool
	err := database.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE username = ?)", username).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Query error:", err)
		return
	}
	if exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}
	hashedPassword := hashPassword(password)
	_, err = database.Db.Exec("INSERT INTO `Users`(`username`, `auth`) VALUES (?, ?); ", username, hashedPassword)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Insert error:", err)
		return
	}
	println("User registered successfully")
	w.Write([]byte("User registered successfully"))
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	var storedHash string
	err := database.Db.QueryRow("SELECT auth FROM Users WHERE username = ?", username).Scan(&storedHash)
    	if err == sql.ErrNoRows {
    	    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
    	    return
    	} else if err != nil {
    	    http.Error(w, "Server error", http.StatusInternalServerError)
    	    fmt.Println("Query error:", err)
    	    return
    	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
        	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        	return
    	}
	w.Write([]byte("Authenticated"))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Handle logout logic here
	w.Write([]byte("Logout handler"))
}
// Call this before starting the server
func AddAuthListeners() {
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/register", registerHandler)

}
