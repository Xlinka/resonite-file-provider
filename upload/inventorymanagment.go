package upload

import (
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
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	auth := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		http.Error(w, "Auth token missing or invalid", http.StatusUnauthorized)
	}
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId missing or invalid", http.StatusBadRequest)
		return
	}
	if allowed, err := query.IsFolderOwner(folderId, claims.UID); err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	folderName := r.URL.Query().Get("folderName")
	if folderName == "" {
		http.Error(w, "folderName missing", http.StatusBadRequest)
		return
	}
	result, err := database.Db.Exec(`
		INSERT INTO Folders (name, parent_folder_id, inventory_id)
		SELECT ?, ?, t.inventory_id
		FROM (SELECT inventory_id FROM Folders WHERE id = ?) AS t;
		`,
		folderName, folderId, folderId,
	)
	if err != nil {
		http.Error(w, "Failed to add folder", http.StatusInternalServerError)
		return
	}
	newFolderId, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve new folder ID but folder has been created", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(strconv.FormatInt(newFolderId, 10)))
}

func HandleAddInventory(w http.ResponseWriter, r* http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
	auth := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		http.Error(w, "Auth token missing or invalid", http.StatusUnauthorized)
	}
	inventoryName := r.URL.Query().Get("inventoryName")
	if inventoryName == "" {
		http.Error(w, "inventoryName missing", http.StatusBadRequest)
		return
	}
	res, err := database.Db.Exec(`INSERT INTO Inventories (name) VALUES (?)`, inventoryName)
	if err != nil {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
	}
	invID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
	}
	_, err = database.Db.Exec(`INSERT INTO users_inventories (user_id, inventory_id) VALUES (?, ?)`, claims.UID, invID)
	if err != nil {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
	}
	res, err = database.Db.Exec(`INSERT INTO Folders (name, parent_folder_id, inventory_id) VALUES (?, NULL, ?)`, "root", invID)
	if err != nil {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
	}
	folderID, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to add inventory", http.StatusInternalServerError)
	}
	w.Write(
		[]byte(
			strconv.FormatInt(invID, 10) + "\n" + strconv.FormatInt(folderID, 10),
		),
	)
}

func removeItem(itemId int) error {
	var affectedAssetIds []int
	rows, err := database.Db.Query("SELECT asset_id FROM `hash-usage` WHERE item_id = ?", itemId)
	if err != nil{
		return err
	}
	for rows.Next() {
		var assetId int
		rows.Scan(&assetId)
		affectedAssetIds = append(affectedAssetIds, assetId)
	}
	_, err = database.Db.Exec("DELETE FROM `hash-usage` WHERE item_id = ?", itemId)
	_, err = database.Db.Exec("DELETE FROM Items WHERE id = ?", itemId)
	for _, affectedId := range affectedAssetIds {
		var assetHash string
		err := database.Db.QueryRow("SELECT hash FROM `Assets` WHERE ID = ?", affectedId).Scan(&assetHash)
		if err != nil {
			return err
		}
		var deleteAsset bool
		err = database.Db.QueryRow("SELECT NOT EXISTS(SELECT 1 FROM `hash-usage` WHERE `asset_id` = ?)", affectedId).Scan(&deleteAsset)
		if deleteAsset{
			_, err := database.Db.Exec("DELETE FROM `Assets` WHERE id = ?", affectedId)
			if err != nil {
				return err
			}
			os.Remove(filepath.Join(config.GetConfig().Server.AssetsPath, assetHash))
			os.Remove(filepath.Join(config.GetConfig().Server.AssetsPath, assetHash) + ".brson")
		}
	}
	return nil
}

//func removeFolder(folderId int) error {
//	
//}

func HandleRemoveItem(w http.ResponseWriter, r *http.Request){
	//if r.Method != http.MethodPost {
	//	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	//	return
	//}
	auth := r.URL.Query().Get("auth")
	claims, err := authentication.ParseToken(auth)
	if err != nil {
		http.Error(w, "Auth token missing or invalid", http.StatusUnauthorized)
	}
	itemId, err := strconv.Atoi(r.URL.Query().Get("itemId"))
	if err != nil {
		http.Error(w, "itemId missing or invalid", http.StatusBadRequest)
		return
	}
	var folderId int
	database.Db.QueryRow("SELECT folder_id FROM Items WHERE id = ?", itemId).Scan(&folderId)
	if allowed, err := query.IsFolderOwner(folderId, claims.UID); err != nil || !allowed {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	err = removeItem(itemId)
	if err != nil {
		http.Error(w, "Failed to remove item", http.StatusInternalServerError)
	}
}
