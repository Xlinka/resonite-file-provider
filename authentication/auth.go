package authentication

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"resonite-file-provider/database"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func readBody(r *http.Request) (string, string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", "", err
	}
	bodyString := string(body)
	// Non standard way to read the body for ease of use in Resonite
	creds := strings.Split(bodyString, "\n")
	if len(creds) < 2 {
		return "", "", fmt.Errorf("invalid credentials format")
	}
	username := creds[0]
	password := creds[1]
	return username, password, nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	username, password, err := readBody(r)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Read error:", err)
		return
	}
	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}
	
	var exists bool
	err = database.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM Users WHERE username = ?)", username).Scan(&exists)
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
	
	// Start a transaction for creating user with inventory
	tx, err := database.Db.Begin()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Transaction error:", err)
		return
	}
	defer tx.Rollback()
	
	// Create user without specifying id (let it auto-increment)
	result, err := tx.Exec("INSERT INTO Users (username, auth) VALUES (?, ?)", username, hashedPassword)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Insert user error:", err)
		return
	}
	
	userId, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Get user ID error:", err)
		return
	}
	
	// Create inventory for the user
	inventoryName := fmt.Sprintf("%s's Inventory", username)
	inventoryResult, err := tx.Exec("INSERT INTO Inventories (name) VALUES (?)", inventoryName)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Insert inventory error:", err)
		return
	}
	
	inventoryId, err := inventoryResult.LastInsertId()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Get inventory ID error:", err)
		return
	}
	
	// Associate user with inventory
	_, err = tx.Exec("INSERT INTO users_inventories (user_id, inventory_id, access_level) VALUES (?, ?, 'owner')", userId, inventoryId)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Insert user_inventory error:", err)
		return
	}
	
	// Create root folder
	_, err = tx.Exec("INSERT INTO Folders (name, parent_folder_id, inventory_id) VALUES ('Root', NULL, ?)", inventoryId)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Insert folder error:", err)
		return
	}
	
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("Commit error:", err)
		return
	}
	
	w.Write([]byte("User registered successfully"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[AUTH] Login request received", r.Method)
	username, password, err := readBody(r)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("[AUTH] Read error:", err)
		return
	}
	
	fmt.Printf("[AUTH] Login attempt for user: %s\n", username)
	
	var storedHash string
	var uId int
	err = database.Db.QueryRow("SELECT auth, id FROM Users WHERE username = ?", username).Scan(&storedHash, &uId)
	if err == sql.ErrNoRows {
		fmt.Printf("[AUTH] User not found: %s\n", username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("[AUTH] Query error:", err)
		return
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		fmt.Printf("[AUTH] Invalid password for user: %s\n", username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	
	token, err := GenerateToken(username, uId)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		fmt.Println("[AUTH] Token generation error:", err)
		return
	}
	
	// Set the auth token as a cookie with detailed settings for troubleshooting
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: false,  // Allow JavaScript access for debugging
		SameSite: http.SameSiteLaxMode,
		Secure:   false,  // Since we're in development
	}
	
	http.SetCookie(w, cookie)
	
	// Log the cookie details
	fmt.Printf("[AUTH] Setting cookie: %s=%s; Path=%s; MaxAge=%d; HttpOnly=%t; SameSite=%v\n", 
		cookie.Name, cookie.Value[:10]+"...", cookie.Path, cookie.MaxAge, cookie.HttpOnly, cookie.SameSite)
	
	fmt.Printf("[AUTH] Login successful for user: %s\n", username)
	
	// Also return the token in the response body for non-browser clients
	w.Write([]byte(token))
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