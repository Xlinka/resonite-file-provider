package upload

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"resonite-file-provider/authentication"
	"resonite-file-provider/config"
	"resonite-file-provider/database"
	"resonite-file-provider/query"
	"strconv"
)

func HandleAddFolder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[FOLDER] AddFolder request received:", r.Method, r.URL.String())
	fmt.Println("[FOLDER] Request headers:", r.Header)
	
	if r.Method != http.MethodPost {
		fmt.Println("[FOLDER] Invalid method:", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Invalid request method",
		})
		return
	}
	
	// Log cookies
	cookies := r.Cookies()
	fmt.Println("[FOLDER] Request cookies:", cookies)
	
	// Try to get auth token from multiple sources
	var auth string
	
	// First try cookie (preferred)
	authCookie, err := r.Cookie("auth_token")
	if err == nil {
		auth = authCookie.Value
		fmt.Println("[FOLDER] Found auth_token cookie:", auth[:10]+"...")
	} else {
		// Fallback to query parameter
		auth = r.URL.Query().Get("auth")
		if auth != "" {
			fmt.Println("[FOLDER] Found auth in query param:", auth[:10]+"...")
		}
	}
	
	if auth == "" {
		// Log debug information
		fmt.Println("[FOLDER] No auth token found in cookie or query param")
		
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token missing",
		})
		return
	}
	
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		fmt.Println("[FOLDER] Auth token invalid:", err.Error())
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token invalid: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[FOLDER] Auth successful for user ID:", claims.UID, "Username:", claims.Username)
	
	// Get folder ID from query parameters
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		fmt.Println("[FOLDER] Invalid folder ID:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "folderId missing or invalid",
		})
		return
	}
	
	// Check if user has editor access to this folder
	if allowed, err := query.CheckFolderAccess(folderId, claims.UID, "editor"); err != nil || !allowed {
		fmt.Println("[FOLDER] Access denied to folder ID:", folderId, "for user:", claims.Username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "You don't have permission to create folders here",
		})
		return
	}
	
	// Get folder name from query parameters
	folderName := r.URL.Query().Get("folderName")
	if folderName == "" {
		fmt.Println("[FOLDER] Missing folder name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "folderName parameter is missing",
		})
		return
	}
	
	fmt.Println("[FOLDER] Creating folder:", folderName, "in parent folder ID:", folderId)
	
	// Insert the new folder
	result, err := database.Db.Exec(`
		INSERT INTO Folders (name, parent_folder_id, inventory_id)
		SELECT ?, ?, inventory_id
		FROM Folders
		WHERE id = ?
		`,
		folderName, folderId, folderId,
	)
	
	if err != nil {
		fmt.Println("[FOLDER] Database error creating folder:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to add folder: " + err.Error(),
		})
		return
	}
	
	// Get the new folder ID
	newFolderId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("[FOLDER] Error getting new folder ID:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to retrieve new folder ID: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[FOLDER] Successfully created folder with ID:", newFolderId)
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"folderId": newFolderId,
		"name": folderName,
		"parentId": folderId,
	})
}

func HandleAddInventory(w http.ResponseWriter, r* http.Request){
	fmt.Println("[INVENTORY] AddInventory request received:", r.Method, r.URL.String())
	fmt.Println("[INVENTORY] Request headers:", r.Header)
	
	if r.Method != http.MethodPost {
		fmt.Println("[INVENTORY] Invalid method:", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Invalid request method",
		})
		return
	}
	
	// Log cookies
	cookies := r.Cookies()
	fmt.Println("[INVENTORY] Request cookies:", cookies)
	
	// Try to get auth token from multiple sources
	var auth string
	
	// First try cookie (preferred)
	authCookie, err := r.Cookie("auth_token")
	if err == nil {
		auth = authCookie.Value
		fmt.Println("[INVENTORY] Found auth_token cookie:", auth[:10]+"...")
	} else {
		// Fallback to query parameter
		auth = r.URL.Query().Get("auth")
		if auth != "" {
			fmt.Println("[INVENTORY] Found auth in query param:", auth[:10]+"...")
		}
	}
	
	if auth == "" {
		// Log debug information
		fmt.Println("[INVENTORY] No auth token found in cookie or query param")
		
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token missing",
		})
		return
	}
	
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		fmt.Println("[INVENTORY] Auth token invalid:", err.Error())
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token invalid: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[INVENTORY] Auth successful for user ID:", claims.UID, "Username:", claims.Username)
	inventoryName := r.URL.Query().Get("inventoryName")
	if inventoryName == "" {
		fmt.Println("[INVENTORY] inventoryName missing in request")
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "inventoryName parameter is missing",
		})
		return
	}
	
	// Debug output
	fmt.Println("[INVENTORY] Creating inventory:", inventoryName, "for user ID:", claims.UID)
	
	// Use a transaction to ensure data consistency
	tx, err := database.Db.Begin()
	if err != nil {
		fmt.Println("[INVENTORY] Failed to start transaction:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to start transaction: " + err.Error(),
		})
		return
	}
	defer tx.Rollback()
	
	res, err := tx.Exec(`INSERT INTO Inventories (name) VALUES (?)`, inventoryName)
	if err != nil {
		fmt.Println("[INVENTORY] Failed to add inventory:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to add inventory: " + err.Error(),
		})
		return
	}
	
	invID, err := res.LastInsertId()
	if err != nil {
		fmt.Println("[INVENTORY] Failed to get inventory ID:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to get inventory ID: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[INVENTORY] Created inventory with ID:", invID)
	
	// Use the updated schema with access_level field
	_, err = tx.Exec(`INSERT INTO users_inventories (user_id, inventory_id, access_level) VALUES (?, ?, 'owner')`, claims.UID, invID)
	if err != nil {
		fmt.Println("[INVENTORY] Failed to add inventory association:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to add inventory association: " + err.Error(),
		})
		return
	}
	
	// Create root folder with NULL parent_folder_id
	fmt.Println("[INVENTORY] Creating root folder for inventory:", invID)
	res, err = tx.Exec(`INSERT INTO Folders (name, parent_folder_id, inventory_id) VALUES (?, NULL, ?)`, "Root", invID)
	if err != nil {
		fmt.Println("[INVENTORY] Failed to create root folder:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to create root folder: " + err.Error(),
		})
		return
	}
	
	folderID, err := res.LastInsertId()
	if err != nil {
		fmt.Println("[INVENTORY] Failed to get new folder ID:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to get new folder ID: " + err.Error(),
		})
		return
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Println("[INVENTORY] Failed to commit transaction:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to commit transaction: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[INVENTORY] Created root folder with ID:", folderID)
	
	// Return inventory and root folder IDs
	w.Header().Set("Content-Type", "application/json")
	
	// Create response object
	response := map[string]interface{}{
		"success": true,
		"inventoryId": invID,
		"rootFolderId": folderID,
	}
	
	// Debug output
	fmt.Println("[INVENTORY] Sending JSON response:", response)
	
	// Encode as JSON
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("[INVENTORY] Error encoding JSON response:", err.Error())
		// At this point, we've already started writing the response, so we can't change the status code
		// Just log the error
	}
	
	fmt.Println("[INVENTORY] Successfully completed inventory creation")
}

func removeItem(itemId int) error {
	// Use a transaction to ensure consistency
	tx, err := database.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Get affected asset IDs
	var affectedAssetIds []int
	rows, err := tx.Query("SELECT asset_id FROM `hash-usage` WHERE item_id = ?", itemId)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var assetId int
		if err := rows.Scan(&assetId); err != nil {
			return err
		}
		affectedAssetIds = append(affectedAssetIds, assetId)
	}
	
	// Delete hash-usage entries for this item
	_, err = tx.Exec("DELETE FROM `hash-usage` WHERE item_id = ?", itemId)
	if err != nil {
		return err
	}
	
	// Delete the item
	_, err = tx.Exec("DELETE FROM Items WHERE id = ?", itemId)
	if err != nil {
		return err
	}
	
	// Check each affected asset to see if it's still used
	for _, affectedId := range affectedAssetIds {
		var assetHash string
		err := tx.QueryRow("SELECT hash FROM Assets WHERE id = ?", affectedId).Scan(&assetHash)
		if err != nil {
			return err
		}
		
		var count int
		err = tx.QueryRow("SELECT COUNT(*) FROM `hash-usage` WHERE asset_id = ?", affectedId).Scan(&count)
		if err != nil {
			return err
		}
		
		// If asset is no longer used, delete it
		if count == 0 {
			_, err := tx.Exec("DELETE FROM Assets WHERE id = ?", affectedId)
			if err != nil {
				return err
			}
			
			// Delete the physical files
			os.Remove(filepath.Join(config.GetConfig().Server.AssetsPath, assetHash))
			os.Remove(filepath.Join(config.GetConfig().Server.AssetsPath, assetHash) + ".brson")
		}
	}
	
	// Commit the transaction
	return tx.Commit()
}

func HandleRemoveItem(w http.ResponseWriter, r *http.Request){
	fmt.Println("[ITEM] RemoveItem request received:", r.Method, r.URL.String())
	fmt.Println("[ITEM] Request headers:", r.Header)
	
	if r.Method != http.MethodPost {
		fmt.Println("[ITEM] Invalid method:", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Invalid request method",
		})
		return
	}
	
	// Log cookies
	cookies := r.Cookies()
	fmt.Println("[ITEM] Request cookies:", cookies)
	
	// Try to get auth token from multiple sources
	var auth string
	
	// First try cookie (preferred)
	authCookie, err := r.Cookie("auth_token")
	if err == nil {
		auth = authCookie.Value
		fmt.Println("[ITEM] Found auth_token cookie:", auth[:10]+"...")
	} else {
		// Fallback to query parameter
		auth = r.URL.Query().Get("auth")
		if auth != "" {
			fmt.Println("[ITEM] Found auth in query param:", auth[:10]+"...")
		}
	}
	
	if auth == "" {
		// Log debug information
		fmt.Println("[ITEM] No auth token found in cookie or query param")
		
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token missing",
		})
		return
	}
	
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		fmt.Println("[ITEM] Auth token invalid:", err.Error())
		// Return JSON error instead of HTML error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Auth token invalid: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[ITEM] Auth successful for user ID:", claims.UID, "Username:", claims.Username)
	
	// Get item ID from query parameters
	itemId, err := strconv.Atoi(r.URL.Query().Get("itemId"))
	if err != nil {
		fmt.Println("[ITEM] Invalid item ID:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "itemId missing or invalid",
		})
		return
	}
	
	// Get the folder ID for this item
	var folderId int
	err = database.Db.QueryRow("SELECT folder_id FROM Items WHERE id = ?", itemId).Scan(&folderId)
	if err != nil {
		fmt.Println("[ITEM] Error finding item in database:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Item not found: " + err.Error(),
		})
		return
	}
	
	// Check if user has editor access to the folder
	if allowed, err := query.CheckFolderAccess(folderId, claims.UID, "editor"); err != nil || !allowed {
		fmt.Println("[ITEM] Access denied to folder ID:", folderId, "for user:", claims.Username)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "You don't have permission to delete items in this folder",
		})
		return
	}
	
	// Remove the item
	err = removeItem(itemId)
	if err != nil {
		fmt.Println("[ITEM] Error removing item:", err.Error())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to remove item: " + err.Error(),
		})
		return
	}
	
	fmt.Println("[ITEM] Successfully removed item ID:", itemId)
	
	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"itemId": itemId,
		"message": "Item successfully removed",
	})
}