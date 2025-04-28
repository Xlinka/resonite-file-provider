package query

import (
	"encoding/json"
	"net/http"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"strconv"
)

// JSON response structures for web API
type InventoriesResponse struct {
	Success bool                `json:"success"`
	Data    []InventoryListItem `json:"data"`
}

type InventoryListItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	RootFolderId int    `json:"rootFolderId"`
}

type InventoryRootResponse struct {
	Success     bool `json:"success"`
	RootFolderId int  `json:"rootFolderId"`
}

type FoldersResponse struct {
	Success bool                `json:"success"`
	Data    []FolderListItem    `json:"data"`
	Parent  *ParentFolderInfo   `json:"parent,omitempty"`
}

type FolderListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ParentFolderInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ItemsResponse struct {
	Success bool             `json:"success"`
	Data    []ItemListItem   `json:"data"`
}

type ItemListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type FolderContentsResponse struct {
	Success bool             `json:"success"`
	Folders []FolderListItem `json:"folders"`
	Items   []ItemListItem   `json:"items"`
	Parent  *ParentFolderInfo `json:"parent,omitempty"`
}

// Handler for JSON API endpoints for web interface

// listInventoriesJSON handles GET /api/inventories
func listInventoriesJSON(w http.ResponseWriter, r *http.Request) {
	auth := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}

	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")
	
	result, err := database.Db.Query(`
        SELECT i.name, i.id, 
            (SELECT f.id FROM Folders f WHERE f.inventory_id = i.id AND (f.parent_folder_id IS NULL OR f.parent_folder_id = -1) LIMIT 1) as root_folder_id
        FROM Inventories i 
        WHERE i.id IN (SELECT inventory_id FROM users_inventories WHERE user_id = ?)
    `, claims.UID)
	if err != nil {
		response := InventoriesResponse{
			Success: false,
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer result.Close()
	
	var inventories []InventoryListItem
	for result.Next() {
		var name string
		var id int
		var rootFolderId int
		result.Scan(&name, &id, &rootFolderId)
		inventories = append(inventories, InventoryListItem{
			ID:           id,
			Name:         name,
			RootFolderId: rootFolderId,
		})
	}
	
	response := InventoriesResponse{
		Success: true,
		Data:    inventories,
	}
	
	json.NewEncoder(w).Encode(response)
}

// listFoldersJSON handles GET /api/folders/subfolders
func listFoldersJSON(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId is either not specified or is invalid", http.StatusBadRequest)
		return
	}
	
	authKey := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(authKey)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")
	
	// Get child folders
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders WHERE parent_folder_id = ?", folderId)
	if err != nil {
		response := FoldersResponse{
			Success: false,
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer childFolders.Close()
	
	var folders []FolderListItem
	for childFolders.Next() {
		var id int
		var name string
		childFolders.Scan(&id, &name)
		folders = append(folders, FolderListItem{
			ID:   id,
			Name: name,
		})
	}
	
	// Get parent folder info
	var parentInfo *ParentFolderInfo
	var parentID *int
	var parentName string
	
	err = database.Db.QueryRow("SELECT parent_folder_id, (SELECT name FROM Folders WHERE id = f.parent_folder_id) FROM Folders f WHERE id = ?", folderId).Scan(&parentID, &parentName)
	if err == nil && parentID != nil {
		parentInfo = &ParentFolderInfo{
			ID:   *parentID,
			Name: parentName,
		}
	}
	
	response := FoldersResponse{
		Success: true,
		Data:    folders,
		Parent:  parentInfo,
	}
	
	json.NewEncoder(w).Encode(response)
}

// listItemsJSON handles GET /api/folders/items
func listItemsJSON(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId is either not specified or is invalid", http.StatusBadRequest)
		return
	}
	
	authKey := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(authKey)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")
	
	// Get items
	items, err := database.Db.Query("SELECT id, name, url FROM Items WHERE folder_id = ?", folderId)
	if err != nil {
		response := ItemsResponse{
			Success: false,
			Data:    nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer items.Close()
	
	var itemList []ItemListItem
	for items.Next() {
		var id int
		var name string
		var url string
		items.Scan(&id, &name, &url)
		itemList = append(itemList, ItemListItem{
			ID:   id,
			Name: name,
			URL:  "assets/" + url,
		})
	}
	
	response := ItemsResponse{
		Success: true,
		Data:    itemList,
	}
	
	json.NewEncoder(w).Encode(response)
}

// listFolderContentsJSON handles GET /api/folders/contents
func listFolderContentsJSON(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId is either not specified or is invalid", http.StatusBadRequest)
		return
	}
	
	authKey := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(authKey)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")
	
	// Get subfolders
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders WHERE parent_folder_id = ?", folderId)
	if err != nil {
		response := FolderContentsResponse{
			Success: false,
			Folders: nil,
			Items:   nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer childFolders.Close()
	
	var folders []FolderListItem
	for childFolders.Next() {
		var id int
		var name string
		childFolders.Scan(&id, &name)
		folders = append(folders, FolderListItem{
			ID:   id,
			Name: name,
		})
	}
	
	// Get items
	items, err := database.Db.Query("SELECT id, name, url FROM Items WHERE folder_id = ?", folderId)
	if err != nil {
		response := FolderContentsResponse{
			Success: false,
			Folders: folders,
			Items:   nil,
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	defer items.Close()
	
	var itemList []ItemListItem
	for items.Next() {
		var id int
		var name string
		var url string
		items.Scan(&id, &name, &url)
		itemList = append(itemList, ItemListItem{
			ID:   id,
			Name: name,
			URL:  "assets/" + url,
		})
	}
	
	// Get parent folder info
	var parentInfo *ParentFolderInfo
	var parentID *int
	var parentName string
	
	err = database.Db.QueryRow("SELECT parent_folder_id, (SELECT name FROM Folders WHERE id = f.parent_folder_id) FROM Folders f WHERE id = ?", folderId).Scan(&parentID, &parentName)
	if err == nil && parentID != nil {
		parentInfo = &ParentFolderInfo{
			ID:   *parentID,
			Name: parentName,
		}
	}
	
	response := FolderContentsResponse{
		Success: true,
		Folders: folders,
		Items:   itemList,
		Parent:  parentInfo,
	}
	
	json.NewEncoder(w).Encode(response)
}

// getInventoryRootFolder handles GET /api/inventory/rootFolder
func getInventoryRootFolder(w http.ResponseWriter, r *http.Request) {
    inventoryId, err := strconv.Atoi(r.URL.Query().Get("inventoryId"))
    if err != nil {
        http.Error(w, "inventoryId is either not specified or is invalid", http.StatusBadRequest)
        return
    }
    
    authKey := r.URL.Query().Get("auth")
    claims, err := authentication.ParseToken(authKey)
    if err != nil {
        http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
        return
    }
    
    // Check if user has access to this inventory
    var hasAccess bool
    err = database.Db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM users_inventories WHERE user_id = ? AND inventory_id = ?)",
        claims.UID, inventoryId,
    ).Scan(&hasAccess)
    
    if err != nil || !hasAccess {
        http.Error(w, "You don't have access to this inventory", http.StatusForbidden)
        return
    }
    
    // Set JSON content type
    w.Header().Set("Content-Type", "application/json")
    
    // Get the root folder ID
    var rootFolderId int
    err = database.Db.QueryRow(
        "SELECT id FROM Folders WHERE inventory_id = ? AND (parent_folder_id IS NULL OR parent_folder_id = -1) LIMIT 1",
        inventoryId,
    ).Scan(&rootFolderId)
    
    if err != nil {
        response := InventoryRootResponse{
            Success: false,
        }
        json.NewEncoder(w).Encode(response)
        return
    }
    
    response := InventoryRootResponse{
        Success: true,
        RootFolderId: rootFolderId,
    }
    
    json.NewEncoder(w).Encode(response)
}

// AddJSONAPIListeners registers the JSON API endpoints
func AddJSONAPIListeners() {
	http.HandleFunc("/api/inventories", listInventoriesJSON)
	http.HandleFunc("/api/folders/subfolders", listFoldersJSON)
	http.HandleFunc("/api/folders/items", listItemsJSON)
	http.HandleFunc("/api/folders/contents", listFolderContentsJSON)
	http.HandleFunc("/api/inventory/rootFolder", getInventoryRootFolder)
}
