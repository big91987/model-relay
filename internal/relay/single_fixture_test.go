package relay

import "strings"

func singleFixtureInput(store *Store, method, path string, body any) (string, string, any) {
	input, ok := body.(map[string]any)
	if path == "upstream" {
		if method == "GET" {
			return method, "upstreams/1", body
		}
		if method == "PUT" && ok {
			value, err := store.upstreamByID(1)
			copy := map[string]any{"name": "isolated upstream", "url": input["url"], "token": input["token"], "enabled": input["enabled"]}
			if copy["token"] == nil {
				copy["token"] = ""
			}
			if err != nil {
				return "POST", "upstreams", copy
			}
			copy["expected_revision"] = value.Revision
			return "PATCH", "upstreams/1", copy
		}
	}
	if ok && (path == "models" || strings.HasPrefix(path, "models/")) && input["model"] != nil {
		copy := map[string]any{"enabled": input["enabled"], "candidates": []map[string]any{{"upstream_id": int64(1), "remote_model": input["model"], "enabled": true, "priority": 0, "weight": 1}}}
		if method == "POST" {
			copy["alias"] = input["alias"]
		} else {
			value, _ := store.routeModel(strings.TrimPrefix(path, "models/"))
			copy["expected_revision"] = value.Revision
			if input["alias"] != nil {
				copy["alias"] = input["alias"]
			}
		}
		return method, path, copy
	}
	return method, path, body
}
