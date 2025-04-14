package upload

import "net/http"

func StartWebServer() {
	fs := http.FileServer(http.Dir("/home/space/Projects/package-uploader-site"))
	http.ListenAndServe(":8080", fs)
}
