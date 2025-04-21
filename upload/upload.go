package upload

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"resonite-file-provider/config"
	"resonite-file-provider/database"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)
var brsonHeader = []byte{70, 114, 68, 84, 0, 0, 0, 0, 3}

func mapRecursiveReplace(data interface{}, old string, new string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			v[key] = mapRecursiveReplace(value, old, new)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = mapRecursiveReplace(item, old, new)
		}
		return v
	case primitive.A:
		for i, item := range v {
			v[i] = mapRecursiveReplace(item, old, new)
		}
		return v
	case string:
		return strings.ReplaceAll(v, old, new)
	default:
		return v
	}
}

func writeBrson(doc map[string]interface{}) ([]byte, error) {
	bsonData, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal BSON: %w", err)
	}

	var compressedBuf bytes.Buffer
	writer := brotli.NewWriter(&compressedBuf)
	if _, err := writer.Write(bsonData); err != nil {
		return nil, fmt.Errorf("brotli compression failed: %w", err)
	}
	writer.Close()

	final := append(brsonHeader, compressedBuf.Bytes()...)
	return final, nil
}

func readBrson(data []byte) (map[string]any, error) {
	if !bytes.Equal(data[:9], brsonHeader) {
		return nil, fmt.Errorf("invalid BRSON header")
	}
	// BRSON header is skipped
	compressed := data[9:]

	br := brotli.NewReader(bytes.NewReader(compressed))
	decompressed, err := io.ReadAll(br)
	if err != nil {
		return nil, fmt.Errorf("brotli decompression failed: %w", err)
	}

	var doc map[string]any
	if err := bson.Unmarshal(decompressed, &doc); err != nil {
		return nil, fmt.Errorf("bson unmarshal failed: %w", err)
	}

	return doc, nil
}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	folderId, err := strconv.Atoi(r.URL.Query().Get("folderId"))
	if err != nil {
		http.Error(w, "folderId missing or invalid", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	if !strings.HasSuffix(header.Filename, ".resonitepackage") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}
<<<<<<< Updated upstream
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		http.Error(w, "Failed to unzip file", http.StatusInternalServerError)
		return
	}
	var assetFilename string
	var itemName string
	// first read asset record
	for _, f := range zipReader.File {
		file, err := f.Open()
		if err != nil {
			http.Error(w, "Failed to read file: ", http.StatusInternalServerError)
			return
		}
		if filepath.Base(f.Name) == "R-Main.record" {
			data, err := io.ReadAll(file)
			if err != nil {
				http.Error(w, "Failed to read file main record", http.StatusInternalServerError)
				return
			}
			var recordData map[string]any
			if err := json.Unmarshal(data, &recordData); err != nil {
				http.Error(w, "Failed to read file, invalid main record ", http.StatusBadRequest)
				return
			}
			assetFilename = strings.TrimPrefix(recordData["assetUri"].(string), "packdb:///")
			itemName = recordData["name"].(string)
			if assetFilename == "" || itemName == "" {
				http.Error(w, "Failed to read file, invalid main record ", http.StatusBadRequest)
				return
			}
			break
		}
	}
	itemInsertResult, err := database.Db.Exec("INSERT INTO `Items` (`name`, `folder_id`, `url`) VALUES (?, ?, ?)", itemName, folderId, assetFilename)
	if err != nil {
		http.Error(w, "Failed to insert item into database", http.StatusInternalServerError)
		return
	}
	itemId, err := itemInsertResult.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get last insert id", http.StatusInternalServerError)
		return
	}
	for _, f := range zipReader.File {
		file, err := f.Open()
		if err != nil {
			http.Error(w, "Failed to read file: ", http.StatusInternalServerError)
			return
		}
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file: ", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		if filepath.Dir(f.Name) == "Assets"  {
			filedir := filepath.Join(config.GetConfig().Server.AssetsPath, filepath.Base(f.Name))
			if filepath.Base(f.Name) != assetFilename {
				err = os.WriteFile(filedir, data, 0644)
			} else {
				err = os.WriteFile(filedir + ".brson", data, 0644)
			}
			if err != nil {
				http.Error(w, "Failed to write file: ", http.StatusInternalServerError)
				return
			}
			assetInsertResult, err := database.Db.Exec("INSERT INTO `Assets` (`hash`) VALUES (?)", filepath.Base(f.Name))
			if err == nil {
				assetId, err := assetInsertResult.LastInsertId()
				if err != nil {
					http.Error(w, "Failed to get last insert id", http.StatusInternalServerError)
					return
				}
				database.Db.Exec("INSERT INTO `hash-usage` (`asset_id`, `item_id`) VALUES (?, ?)", assetId, itemId)
			}
		}
		
	}
	if assetFilename == "" || itemName == "" {
		http.Error(w, "Failed to read file, invalid main record ", http.StatusBadRequest)
		return
	}
	brson, err := os.ReadFile(filepath.Join(config.GetConfig().Server.AssetsPath, assetFilename + ".brson"))
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	brsonData, err := readBrson(brson)
	if err != nil {
		http.Error(w, "Failed to read file: ", http.StatusInternalServerError)
		return
	}
	assetUrl := "https://" + filepath.Join(config.GetConfig().Server.Host + ":" + strconv.Itoa(config.GetConfig().Server.Port), "assets")
	newBrsonData := mapRecursiveReplace(brsonData, "packdb://", assetUrl)
	newBrson, err := writeBrson(newBrsonData.(map[string]interface{}))
	os.WriteFile(filepath.Join(config.GetConfig().Server.AssetsPath, assetFilename) + ".brson", newBrson, 0644)
=======
	
	// This logic is temporary and will be replaced by a proper library that won't unnecessarely write to disk in the future
	dstPath := filepath.Join(config.GetConfig().Server.AssetsPath, header.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Close file before executing command that will process it
	dst.Close()
	
	err = exec.Command("./ResoniteFilehost", "2", dstPath).Run()
	if err != nil {
		http.Error(w, "Failed to execute ResoniteFileHost: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	os.Remove(dstPath)
>>>>>>> Stashed changes
	w.Write([]byte("File uploaded successfully"))
}

func AddListeners() {
	http.HandleFunc("/upload", HandleUpload)
	http.HandleFunc("/addFolder", HandleAddFolder)
<<<<<<< Updated upstream
	http.HandleFunc("/removeItem/", HandleRemoveItem)
}
=======
}
>>>>>>> Stashed changes
