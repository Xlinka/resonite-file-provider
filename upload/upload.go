package upload

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"resonite-file-provider/config"
	"strings"
)

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("file")
	defer file.Close()
	if err != nil {
		http.Error(w, "Failed to retrieve file: ", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(header.Filename, ".resonitepackage") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}
	// This logic is temporary and will be replaced by a proper library that won't unnecessarely write to disk in the future
	dst, err := os.Create(filepath.Join(config.GetConfig().Server.AssetsPath, header.Filename))
	if err != nil {
		http.Error(w, "Internal server error: ", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = exec.Command("./ResoniteFilehost", "2", dst.Name()).Run()
	if err != nil {
		http.Error(w, "Failed to execute ResoniteFileHost: ", http.StatusInternalServerError)
		println(err.Error())
		return
	}
	os.Remove(dst.Name())
	w.Write([]byte("File uploaded successfully"))

}

func AddListeners() {
	http.HandleFunc("/upload", HandleUpload)
	http.HandleFunc("/addFolder", HandleAddFolder)
}
