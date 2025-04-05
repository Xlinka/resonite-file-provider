package assethost

import (
	"io"
	"net/http"
	"resonite-file-provider/authentication"
	"resonite-file-provider/database"
	"strings"
)
func isOwnedBy(owner string, url string) bool {
	var exists bool
	database.Db.QueryRow("SELECT EXISTS (SELECT 1 from Users where username = ? AND id = (SELECT user_id FROM users_inventories WHERE inventory_id = (SELECT inventory_id FROM Folders WHERE id = (SELECT folder_id FROM `Items` WHERE url = ?))))", owner, url).Scan(&exists)
	return exists
}
func handleRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/assets/")
		next.ServeHTTP(w, r)
		return
		//TODO
		if !strings.HasSuffix(r.URL.Path, ".brson") {
			next.ServeHTTP(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		claims, err := authentication.ParseToken(strings.Split(string(body), "\n")[0])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		username := claims.Username
		if !isOwnedBy(username, r.URL.Path) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func StartHost(){
	http.Handle("/assets/", handleRequest(http.FileServer(http.Dir("./assets"))))
	http.ListenAndServe(":8080", nil)
}
