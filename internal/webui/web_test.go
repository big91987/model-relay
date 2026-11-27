package webui_test

import (
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"

	"model-relay/internal/webui"
)

func TestBothWorkspacesServeTheirCompleteNativeModuleGraph(t *testing.T) {
	imports := regexp.MustCompile(`from ['"](\.[^'"]+)['"]`)
	for _, schema := range []int{2, 3} {
		t.Run(map[int]string{2: "legacy", 3: "console"}[schema], func(t *testing.T) {
			entry := "/admin/app.js"
			if schema == 3 {
				entry = "/admin/workspace.js"
			}
			queue := []string{entry}
			seen := map[string]bool{}
			for len(queue) > 0 {
				route := queue[0]
				queue = queue[1:]
				if seen[route] {
					continue
				}
				seen[route] = true
				response := httptest.NewRecorder()
				webui.Serve(response, httptest.NewRequest("GET", route, nil), schema)
				if response.Code != 200 || response.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
					t.Fatalf("unserved module %s: %d %s", route, response.Code, response.Body.String())
				}
				for _, match := range imports.FindAllStringSubmatch(response.Body.String(), -1) {
					queue = append(queue, path.Join(path.Dir(route), match[1]))
				}
			}
			feature, shared := "/admin/features/legacy-records/index.mjs", "/admin/shared/legacy-data.mjs"
			if schema == 3 {
				feature, shared = "/admin/features/reports/index.mjs", "/admin/shared/lifetime.mjs"
			}
			if !seen[feature] || !seen[shared] {
				t.Fatalf("feature dependencies were not traversed: %v", seen)
			}
			if schema == 3 && !seen["/admin/features/limits/presentation.mjs"] {
				t.Fatal("quota presentation module is absent from the served dependency graph")
			}
		})
	}
}
func TestEmbeddedRoutesKeepTheirWorkspaceScopeAndRejectTraversal(t *testing.T) {
	for _, test := range []struct {
		schema int
		route  string
		status int
		body   string
	}{
		{2, "/admin", 200, "/admin/app.js"}, {3, "/admin", 200, "/admin/workspace.js"}, {3, "/admin/accept", 200, "邀请"},
		{2, "/admin/accept", 404, ""}, {2, "/admin/workspace.js", 404, ""}, {3, "/admin/app.js", 404, ""},
		{3, "/admin/shared/../app.js", 404, ""}, {3, "/admin/features//models/index.mjs", 404, ""}, {3, "/admin/features/models/index.mjs/", 404, ""}, {3, "/admin/../../web.go", 404, ""}, {3, "/admin/index.html", 404, ""}, {3, "/admin/shared/missing.mjs", 404, ""},
	} {
		t.Run(test.route, func(t *testing.T) {
			response := httptest.NewRecorder()
			webui.Serve(response, httptest.NewRequest("GET", test.route, nil), test.schema)
			if response.Code != test.status || test.body != "" && !strings.Contains(response.Body.String(), test.body) {
				t.Fatalf("route %s: %d %s", test.route, response.Code, response.Body.String())
			}
		})
	}
}
