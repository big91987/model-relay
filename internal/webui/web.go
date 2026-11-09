// Package webui serves the embedded management workspaces and their modules.
package webui

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web
var assets embed.FS

// Serve selects the management workspace for the installation's storage schema.
// Native module imports stay under the same /admin/features and /admin/shared
// roots in both workspaces. No filesystem paths or directory listings are served.
func Serve(writer http.ResponseWriter, request *http.Request, schema int) {
	name := ""
	if schema == 3 {
		switch request.URL.Path {
		case "/", "/admin", "/admin/", "/admin/accept":
			name = "workspace.html"
		case "/admin/workspace.js":
			name = "workspace.js"
		case "/admin/workspace.css":
			name = "workspace.css"
		}
	} else {
		switch request.URL.Path {
		case "/", "/admin", "/admin/":
			name = "index.html"
		case "/admin/app.js":
			name = "app.js"
		case "/admin/style.css":
			name = "style.css"
		}
	}
	if name == "" && strings.HasPrefix(request.URL.Path, "/admin/") {
		candidate := strings.TrimPrefix(request.URL.Path, "/admin/")
		if fs.ValidPath(candidate) && !strings.Contains(candidate, "\\") && strings.HasSuffix(candidate, ".mjs") && (strings.HasPrefix(candidate, "shared/") || strings.HasPrefix(candidate, "features/")) {
			name = candidate
		}
	}
	if name == "" {
		http.NotFound(writer, request)
		return
	}
	content, err := fs.ReadFile(assets, "web/"+name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(writer, request)
		} else {
			http.Error(writer, "asset unavailable", 500)
		}
		return
	}
	kind := "text/html; charset=utf-8"
	if strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".mjs") {
		kind = "text/javascript; charset=utf-8"
	}
	if strings.HasSuffix(name, ".css") {
		kind = "text/css; charset=utf-8"
	}
	writer.Header().Set("Content-Type", kind)
	_, _ = writer.Write(content)
}
