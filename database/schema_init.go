package database

import (
	"database/sql"
	"fmt"
)

// InitializeSchema verifies that the database schema is properly set up
// This should be called after establishing the database connection
func InitializeSchema() error {
	// First, let's verify tables exist with correct structure
	tables := []string{"Users", "Inventories", "users_inventories", "Folders", "Items", "Assets", "hash-usage", "asset_tags", "Tags", "item_tags"}
	
	for _, table := range tables {
		var exists bool
		err := Db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)", table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		
		if !exists {
			return fmt.Errorf("table %s does not exist", table)
		}
	}
	
	// Verify foreign key constraints are in place
	if err := verifyForeignKeys(); err != nil {
		return fmt.Errorf("foreign key verification failed: %w", err)
	}
	
	return nil
}

func verifyForeignKeys() error {
	// Check foreign key constraints exist
	constraints := []struct {
		table      string
		constraint string
	}{
		{"asset_tags", "asset_tags_ibfk_1"},
		{"asset_tags", "asset_tags_ibfk_2"},
		{"Folders", "Folders_ibfk_1"},
		{"hash-usage", "hash-usage_ibfk_1"},
		{"hash-usage", "hash-usage_ibfk_2"},
		{"Items", "Items_ibfk_1"},
		{"item_tags", "item_tags_ibfk_1"},
		{"item_tags", "item_tags_ibfk_2"},
		{"users_inventories", "users_inventories_ibfk_1"},
		{"users_inventories", "users_inventories_ibfk_2"},
	}
	
	for _, c := range constraints {
		var count int
		err := Db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.TABLE_CONSTRAINTS 
			WHERE CONSTRAINT_SCHEMA = DATABASE() 
			AND TABLE_NAME = ? 
			AND CONSTRAINT_NAME = ?
		`, c.table, c.constraint).Scan(&count)
		
		if err != nil {
			return fmt.Errorf("failed to check constraint %s on table %s: %w", c.constraint, c.table, err)
		}
		
		if count == 0 {
			return fmt.Errorf("missing constraint %s on table %s", c.constraint, c.table)
		}
	}
	
	return nil
}

// CreateUserWithInventory creates a new user with a default inventory
// This replaces the stored procedure approach and ensures consistency
func CreateUserWithInventory(username, authHash string) error {
	tx, err := Db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Create user
	result, err := tx.Exec("INSERT INTO Users (username, auth) VALUES (?, ?)", username, authHash)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	userID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	
	// Create personal inventory
	inventoryName := fmt.Sprintf("%s's Inventory", username)
	result, err = tx.Exec("INSERT INTO Inventories (name) VALUES (?)", inventoryName)
	if err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}
	
	inventoryID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inventory ID: %w", err)
	}
	
	// Associate user with inventory using the correct field name 'access_level'
	_, err = tx.Exec(`
		INSERT INTO users_inventories (user_id, inventory_id, access_level) 
		VALUES (?, ?, 'owner')
	`, userID, inventoryID)
	if err != nil {
		return fmt.Errorf("failed to associate user with inventory: %w", err)
	}
	
	// Create root folder
	_, err = tx.Exec(`
		INSERT INTO Folders (name, parent_folder_id, inventory_id) 
		VALUES ('Root', NULL, ?)
	`, inventoryID)
	if err != nil {
		return fmt.Errorf("failed to create root folder: %w", err)
	}
	
	return tx.Commit()
}

// GetUserAccessLevel returns the access level for a user on an inventory
func GetUserAccessLevel(userID, inventoryID int) (string, error) {
	var accessLevel string
	err := Db.QueryRow(`
		SELECT access_level 
		FROM users_inventories 
		WHERE user_id = ? AND inventory_id = ?
	`, userID, inventoryID).Scan(&accessLevel)
	
	if err == sql.ErrNoRows {
		return "", nil // No access
	}
	if err != nil {
		return "", err
	}
	
	return accessLevel, nil
}

// CheckUserInventoryAccess checks if a user has sufficient access to an inventory
func CheckUserInventoryAccess(userID, inventoryID int, requiredLevel string) (bool, error) {
	accessLevel, err := GetUserAccessLevel(userID, inventoryID)
	if err != nil {
		return false, err
	}
	
	if accessLevel == "" {
		return false, nil
	}
	
	// Access level hierarchy: owner > editor > viewer
	switch requiredLevel {
	case "viewer":
		return true, nil // Any access is sufficient
	case "editor":
		return accessLevel == "owner" || accessLevel == "editor", nil
	case "owner":
		return accessLevel == "owner", nil
	default:
		return false, fmt.Errorf("invalid access level: %s", requiredLevel)
	}
}