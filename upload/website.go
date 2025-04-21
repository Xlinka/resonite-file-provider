package upload

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"resonite-file-provider/query"
	"strconv"
)

type PageData struct {
	AuthToken string
	Username  string
	FolderId  int
	Folders   []Folder
	Items     []Item
	Path      []Breadcrumb
	Error     string
}

type Folder struct {
	ID   int
	Name string
}

type Item struct {
	ID   int
	Name string
	URL  string
}

type Breadcrumb struct {
	ID   int
	Name string
}

func handleWebHome(w http.ResponseWriter, r *http.Request) {
	// Serve the home page
	http.ServeFile(w, r, filepath.Join("upload-site", "index.html"))
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Serve static files
	http.ServeFile(w, r, filepath.Join("upload-site", r.URL.Path))
}

func handleFolder(w http.ResponseWriter, r *http.Request) {
	// Get folder ID from query
	folderIdStr := r.URL.Query().Get("id")
	if folderIdStr == "" {
		http.Error(w, "Missing folder ID", http.StatusBadRequest)
		return
	}

	folderId, err := strconv.Atoi(folderIdStr)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	// Get auth token
	authToken := r.URL.Query().Get("auth")
	if authToken == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Validate token
	claims, err := authentication.ParseToken(authToken)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Check folder ownership
	if allowed, err := query.IsFolderOwner(folderId, claims.UID); !allowed || err != nil {
		http.Error(w, "You don't have access to this folder", http.StatusForbidden)
		return
	}

	// Get folder contents
	childFolders, err := getFolders(folderId, authToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := getItems(folderId, authToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get breadcrumb path
	path, err := getBreadcrumbPath(folderId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare data for template
	data := PageData{
		AuthToken: authToken,
		Username:  claims.Username,
		FolderId:  folderId,
		Folders:   childFolders,
		Items:     items,
		Path:      path,
	}

	// Parse and execute template
	tmpl, err := template.ParseFiles(filepath.Join("upload-site", "folder.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func getFolders(folderId int, authToken string) ([]Folder, error) {
	// Query database for child folders
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders WHERE parent_folder_id = ?", folderId)
	if err != nil {
		return nil, err
	}
	defer childFolders.Close()

	var folders []Folder
	for childFolders.Next() {
		var folder Folder
		if err := childFolders.Scan(&folder.ID, &folder.Name); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

func getItems(folderId int, authToken string) ([]Item, error) {
	// Query database for items in folder
	items, err := database.Db.Query("SELECT id, name, url FROM Items WHERE folder_id = ?", folderId)
	if err != nil {
		return nil, err
	}
	defer items.Close()

	var result []Item
	for items.Next() {
		var item Item
		if err := items.Scan(&item.ID, &item.Name, &item.URL); err != nil {
			return nil, err
		}
		item.URL = filepath.Join("assets", item.URL)
		result = append(result, item)
	}

	return result, nil
}

func getBreadcrumbPath(folderId int) ([]Breadcrumb, error) {
	// Query to get the folder path (parent folders)
	var path []Breadcrumb
	currentFolderId := folderId

	for currentFolderId != 0 {
		var folder Breadcrumb
		var parentID *int

		err := database.Db.QueryRow("SELECT id, name, parent_folder_id FROM Folders WHERE id = ?", currentFolderId).Scan(&folder.ID, &folder.Name, &parentID)
		if err != nil {
			return nil, err
		}

		path = append([]Breadcrumb{folder}, path...)

		if parentID == nil {
			break
		}
		currentFolderId = *parentID
	}

	// Add root
	if len(path) == 0 || path[0].ID != 1 {
		root := Breadcrumb{ID: 1, Name: "Root"}
		path = append([]Breadcrumb{root}, path...)
	}

	return path, nil
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Client-side logout - just redirect to home
	http.Redirect(w, r, "/", http.StatusFound)
}

func StartWebServer() {
	// Create the upload-site directory if it doesn't exist
	websitePath := "upload-site"
	if _, err := os.Stat(websitePath); os.IsNotExist(err) {
		os.Mkdir(websitePath, 0755)
	}

	// Set up routes
	http.HandleFunc("/", handleWebHome)
	http.HandleFunc("/folder", handleFolder)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/app.js", handleStatic)
	http.HandleFunc("/styles.css", handleStatic)

	// Start the web server
	fmt.Println("Starting web server on :8080...")
	http.ListenAndServe(":8080", nil)
}