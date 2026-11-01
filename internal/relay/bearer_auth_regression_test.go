package relay

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModelsRequireExactBearerAuthorization(t *testing.T) {
	fixture := governedInstallation(t, false)

	requests := []struct {
		name          string
		authorization func(string) string
		wantStatus    int
	}{
		{name: "missing", authorization: func(string) string { return "" }, wantStatus: http.StatusUnauthorized},
		{name: "credential without scheme", authorization: func(key string) string { return key }, wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", authorization: func(key string) string { return "Basic " + key }, wantStatus: http.StatusUnauthorized},
		{name: "lowercase scheme", authorization: func(key string) string { return "bearer " + key }, wantStatus: http.StatusUnauthorized},
		{name: "leading whitespace", authorization: func(key string) string { return " Bearer " + key }, wantStatus: http.StatusUnauthorized},
		{name: "extra separator whitespace", authorization: func(key string) string { return "Bearer  " + key }, wantStatus: http.StatusUnauthorized},
		{name: "trailing whitespace", authorization: func(key string) string { return "Bearer " + key + " " }, wantStatus: http.StatusUnauthorized},
		{name: "appended content", authorization: func(key string) string { return "Bearer " + key + ",extra" }, wantStatus: http.StatusUnauthorized},
		{name: "empty credential", authorization: func(string) string { return "Bearer " }, wantStatus: http.StatusUnauthorized},
		{name: "exact bearer", authorization: func(key string) string { return "Bearer " + key }, wantStatus: http.StatusOK},
	}

	for _, test := range requests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:5544/v1/models", nil)
			request.Header.Set("Authorization", test.authorization(fixture.key))
			writer := httptest.NewRecorder()

			fixture.server.ServeHTTP(writer, request)

			if writer.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", writer.Code, test.wantStatus)
			}
			if test.wantStatus == http.StatusOK {
				var response struct {
					Data []struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				if err := json.Unmarshal(writer.Body.Bytes(), &response); err != nil {
					t.Fatal("decode model response", err)
				}
				if len(response.Data) != 1 || response.Data[0].ID != "governed" {
					t.Fatalf("authorized model scope changed: %+v", response.Data)
				}
			}
		})
	}
}

func TestGenerationRejectsCredentialWithoutBearerScheme(t *testing.T) {
	fixture := governedInstallation(t, false)
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:5544/v1/chat/completions", nil)
	request.Header.Set("Authorization", fixture.key)
	writer := httptest.NewRecorder()

	fixture.server.ServeHTTP(writer, request)

	if writer.Code != http.StatusUnauthorized || fixture.hits.Load() != 0 {
		t.Fatalf("generation boundary status = %d, upstream visits = %d", writer.Code, fixture.hits.Load())
	}
}
