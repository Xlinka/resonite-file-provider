package query

import (
	"net/http"
	"os"
	"resonite-file-provider/animxmaker"
	"resonite-file-provider/database"
)

func listFolders(w http.ResponseWriter, r *http.Request) {
	println("listFolders")
	folderId := r.URL.Query().Get("folderId")
	if folderId == "" {
		folderId = "-1"
	}
	childFolders, err := database.Db.Query("SELECT id, name FROM Folders where parent_folder_id = ?", folderId);
	var childFoldersIds []float32
	var childFoldersNames []string
	childFolders.Scan(&childFoldersIds)
	childFolders.Next()
	childFolders.Scan(&childFoldersNames)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	idsTrack := animxmaker.ListTrack(childFoldersIds, "ids")
	namesTrack := animxmaker.ListTrack(childFoldersNames, "names")
	response := animxmaker.Animx{
		TrackCount: 2,
		Tracks: []animxmaker.AnimationTrack{
			idsTrack,
			namesTrack,
		},
	}
	encodedResponse, err := response.EncodeBinary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	os.WriteFile("anim.animx", encodedResponse, 0644)
	w.Write(encodedResponse)
}
func AddSearchListeners() {
	http.HandleFunc("/query/list/", listFolders)
}
