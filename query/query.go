package query

import (
	"net/http"
	"path/filepath"
	"resonite-file-provider/animxmaker"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"strconv"
)

func getChildFoldersTracks(folderId int, nodeName string) (animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, error) {
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders where parent_folder_id = ?", folderId);
	if err != nil {
		return nil, nil, nil, err
	}
	var parentFolderId int
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
	parentFolderTrack := animxmaker.ListTrack([]int32{int32(parentFolderId)}, nodeName, "parentFolder")
	return &idsTrack, &namesTrack, &parentFolderTrack, nil
}

func getChildItemsTracks(folderId int, nodeName string) (animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, animxmaker.AnimationTrackWrapper, error) {
	items, err := database.Db.Query("SELECT id, name, url FROM Items where folder_id = ?", folderId);
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

func IsFolderOwner(folderId int, userId int) (bool, error) {
	rows, err := database.Db.Query("SELECT id from Users WHERE id = (SELECT user_id from users_inventories where inventory_id = (SELECT inventory_id FROM Folders WHERE id = ?))", folderId)
	if err != nil {
		return false, err
	}
	for rows.Next(){
		var currectUserId int
		if err := rows.Scan(&currectUserId); err != nil{
			return false, err
		}
		if currectUserId == userId {
			return true, nil
		}
	}
	return false, nil
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
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
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
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
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
	}
	result, err := database.Db.Query("SELECT name, id FROM `Inventories` WHERE id in (SELECT id FROM users_inventories WHERE user_id = ?)", claims.UID)
	if err != nil {
		http.Error(w, "Failed to query the database", http.StatusInternalServerError)
	}
	var inventoryIds []int
	var inventoryNames []string
	for result.Next() {
		var name string
		var id int
		result.Scan(&name, &id)
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
	if allowed, err := IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
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
