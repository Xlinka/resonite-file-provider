package query

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"resonite-file-provider/animxmaker"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"strconv"
)

func getChildFoldersTracks(folderId int, nodeName string) (animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, error) {
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders where parent_folder_id = ?", folderId)
	if err != nil {
		return nil, nil, nil, err
	}
	var parentFolderId sql.NullInt64
	if err := database.Db.QueryRow("SELECT parent_folder_id FROM Folders WHERE id = ?", folderId).Scan(&parentFolderId); err != nil {
		return nil, nil, nil, err
	}
	var childFoldersIds []int32
	var childFoldersNames []string
	defer childFolders.Close()
	
	for childFolders.Next() {
		var id int32
		var name string
		if err := childFolders.Scan(&id, &name); err != nil {
			return nil, nil, nil, err
		}
		childFoldersIds = append(childFoldersIds, id)
		childFoldersNames = append(childFoldersNames, name)
	}

	idsTrack := animxmaker.ListTrack(childFoldersIds, nodeName, "id")
	namesTrack := animxmaker.ListTrack(childFoldersNames, nodeName, "name")
	
	// Handle NULL parent folder ID (which indicates root folder)
	var parentFolderIdValue int32
	if parentFolderId.Valid {
		parentFolderIdValue = int32(parentFolderId.Int64)
	} else {
		parentFolderIdValue = -1 // Use a sentinel value for NULL parent
	}
	
	parentFolderTrack := animxmaker.ListTrack([]int32{parentFolderIdValue}, nodeName, "parentFolder")
	return &idsTrack, &namesTrack, &parentFolderTrack, nil
}

func getChildItemsTracks(folderId int, nodeName string) (animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, error) {
	items, err := database.Db.Query("SELECT id, name, url FROM Items where folder_id = ?", folderId)
	if err != nil {
		return nil, nil, nil, err
	}

	var itemsIds []int32
	var itemsNames []string
	var itemsUrls []string
	defer items.Close()

	for items.Next() {
		var id int32
		var name string
		var url string
		if err := items.Scan(&id, &name, &url); err != nil {
			return nil, nil, nil, err
		}
		itemsIds = append(itemsIds, id)
		itemsNames = append(itemsNames, name)
		itemsUrls = append(itemsUrls, filepath.Join("assets", url))
	}
	idsTrack := animxmaker.ListTrack(itemsIds, nodeName, "id")
	namesTrack := animxmaker.ListTrack(itemsNames, nodeName, "name")
	urlsTrack := animxmaker.ListTrack(itemsUrls, nodeName, "url")
	return &idsTrack, &namesTrack, &urlsTrack, nil
}

// CheckFolderAccess verifies if a user has access to a folder
func CheckFolderAccess(folderId int, userId int, requiredLevel string) (bool, error) {
	// Get the inventory ID for this folder
	var inventoryId int
	err := database.Db.QueryRow("SELECT inventory_id FROM Folders WHERE id = ?", folderId).Scan(&inventoryId)
	if err != nil {
		return false, err
	}
	
	// Check user's access level for this inventory
	return database.CheckUserInventoryAccess(userId, inventoryId, requiredLevel)
}

// Updated function to use the new access control
func IsFolderOwner(folderId int, userId int) (bool, error) {
	return CheckFolderAccess(folderId, userId, "owner")
}

func listFolders(w http.ResponseWriter, r *http.Request) {
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
	
	// Check if user has at least viewer access
	if allowed, err := CheckFolderAccess(folderId, claims.UID, "viewer"); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	idsTrack, namesTrack, parentFoldertrack, err := getChildFoldersTracks(folderId, "results")
	response := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			idsTrack,
			namesTrack,
			parentFoldertrack,
		},
	}
	encodedResponse, err := response.EncodeAnimation("response")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encodedResponse)
}

func listItems(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId is either not specified or is invalid", http.StatusBadRequest)
	}
	authKey := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(authKey)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	// Check if user has at least viewer access
	if allowed, err := CheckFolderAccess(folderId, claims.UID, "viewer"); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	idsTrack, namesTrack, urlsTrack, err := getChildItemsTracks(folderId, "results")
	response := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			idsTrack,
			namesTrack,
			urlsTrack,
		},
	}
	encodedResponse, err := response.EncodeAnimation("response")
	if err != nil {
		http.Error(w, "Error while encoding animx", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(encodedResponse)
}

func listInventories(w http.ResponseWriter, r *http.Request){
	auth := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	// Updated query to use the new table structure
	result, err := database.Db.Query(`
		SELECT i.id, i.name 
		FROM Inventories i
		INNER JOIN users_inventories ui ON i.id = ui.inventory_id
		WHERE ui.user_id = ?
	`, claims.UID)
	
	if err != nil {
		http.Error(w, "Failed to query the database", http.StatusInternalServerError)
		return
	}
	
	var inventoryIds []int
	var inventoryNames []string
	for result.Next() {
		var name string
		var id int
		result.Scan(&id, &name)
		inventoryIds = append(inventoryIds, id)
		inventoryNames = append(inventoryNames, name)
	}
	idsTrack := animxmaker.ListTrack(inventoryIds, "results", "id")
	namesTrack := animxmaker.ListTrack(inventoryNames, "results", "name")
	response := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			animxmaker.AnimationTrackWrapper(&idsTrack),
			animxmaker.AnimationTrackWrapper(&namesTrack),
		},
	}
	encodedResponse, err := response.EncodeAnimation("response")
	if err != nil {
		http.Error(w, "Error while encoding animx", http.StatusInternalServerError)
	}
	w.Write(encodedResponse)
	w.WriteHeader(http.StatusOK)
}

func listFolderContents(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId is either not specified or is invalid", http.StatusBadRequest)
	}
	authKey := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(authKey)
	if err != nil {
		http.Error(w, "Auth token invalid or missing", http.StatusUnauthorized)
		return
	}
	
	// Check if user has at least viewer access
	if allowed, err := CheckFolderAccess(folderId, claims.UID, "viewer"); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}
	
	itemIdsTrack, itemNamesTrack, itemUrlsTrack, err := getChildItemsTracks(folderId, "items")
	if err != nil {
		http.Error(w, "Error while getting items", http.StatusInternalServerError)
		return
	}
	folderIdsTrack, folderNamesTrack, parentFolderTrack, err := getChildFoldersTracks(folderId, "folders")
	if err != nil {
		http.Error(w, "Error while getting folders", http.StatusInternalServerError)
		return
	}
	response := animxmaker.Animation{
		Tracks: []animxmaker.AnimationTrackWrapper{
			itemIdsTrack,
			itemNamesTrack,
			itemUrlsTrack,
			folderIdsTrack,
			folderNamesTrack,
			parentFolderTrack,
		},
	}
	encodedResponse, err := response.EncodeAnimation("response")
	if err != nil {
		http.Error(w, "Error while encoding animx", http.StatusInternalServerError)
	}
	w.Write(encodedResponse)
}

func AddSearchListeners() {
	http.HandleFunc("/query/childFolders", listFolders)
	http.HandleFunc("/query/childItems", listItems)
	http.HandleFunc("/query/folderContent", listFolderContents)
	http.HandleFunc("/query/inventories", listInventories)
}