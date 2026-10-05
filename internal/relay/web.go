package relay

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var assets embed.FS

func serveWeb(writer http.ResponseWriter, request *http.Request) {
	name := "index.html"
	if request.URL.Path == "/admin/app.js" {
		name = "app.js"
	} else if request.URL.Path == "/admin/style.css" {
		name = "style.css"
	} else if request.URL.Path != "/" && request.URL.Path != "/admin" && request.URL.Path != "/admin/" {
		http.NotFound(writer, request)
		return
	}
	content, err := fs.ReadFile(assets, "web/"+name)
	if err != nil {
		http.Error(writer, "asset unavailable", 500)
		return
	}
	kind := "text/html; charset=utf-8"
	if name == "app.js" {
		kind = "text/javascript; charset=utf-8"
	}
	if name == "style.css" {
		kind = "text/css; charset=utf-8"
	}
	writer.Header().Set("Content-Type", kind)
	_, _ = writer.Write(content)
}
