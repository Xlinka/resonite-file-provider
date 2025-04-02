package assethost

import "net/http"

func StartHost(){
	http.Handle("/", http.FileServer(http.Dir("./assets")))
	http.ListenAndServe(":8080", nil)
}
