package upload

import (
	"net/http"
	"resonite-file-provider/authentication"
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
	res, err = database.Db.Exec(`INSERT INTO Folders (name, parent_folder_id, inventory_id) VALUES (?, ?, ?)`, "root", -1, invID)
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
