package assethost

import (
	"net/http"
	"resonite-file-provider/authentication"
	"resonite-file-provider/config"
	"resonite-file-provider/database"
	"strings"
)
func isOwnedBy(owner int, url string) bool {
	var exists bool
	url = strings.TrimSuffix(url, ".brson")
	database.Db.QueryRow(`
	SELECT EXISTS (
	  SELECT 1 
	  FROM Users 
	  WHERE id = ?
	    AND id IN (
	      SELECT user_id 
	      FROM users_inventories 
	      WHERE inventory_id IN (
	        SELECT inventory_id 
	        FROM Folders 
	        WHERE id IN (
	          SELECT folder_id 
	          FROM Items 
	          WHERE url = ?
	        )
	      )
	    )
	)
	`, owner, url).Scan(&exists)
	return exists
}
func handleRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/assets/")
		if !strings.HasSuffix(r.URL.Path, ".brson") {
			next.ServeHTTP(w, r)
			return
		}
		authToken := r.URL.Query().Get("auth")
		claims, err := authentication.ParseToken(authToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		uId := claims.UID
		if !isOwnedBy(uId, r.URL.Path) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AddAssetListeners(){
	http.Handle("/assets/", handleRequest(http.FileServer(http.Dir(config.GetConfig().Server.AssetsPath))))
}
