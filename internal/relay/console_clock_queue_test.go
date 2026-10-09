package relay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConsoleClockQueuedDeadlineInProcess(test *testing.T) {
	test.Parallel()
	clockQueuedDeadline(test, false)
}
func TestConsoleClockQueuedDeadlineActualHTTP(test *testing.T) {
	test.Parallel()
	clockQueuedDeadline(test, true)
}

func clockQueuedDeadline(test *testing.T, actual bool) {
	for _, boundary := range []string{"models", "invitation"} {
		test.Run(boundary, func(test *testing.T) {
			fixture := singleResourceInstallation(test, actual)
			start := time.Unix(0, fixture.clock.Load())
			status, _ := fixture.admin.call("POST", "tenants/"+fixture.tenant+"/keys/"+fixture.keyID+"/rotations", map[string]any{"expected_revision": 1, "transition_hours": 1}, "queue-rotation")
			if status != 200 {
				test.Fatal("queue rotation", status)
			}
			status, invitation := fixture.admin.call("POST", "tenants/"+fixture.tenant+"/invitations", map[string]any{"account": "queued@example.com", "role": "developer", "hours": 1}, "")
			if status != 201 {
				test.Fatal("queue invitation", status)
			}
			request := httptest.NewRequest("GET", fixture.origin+"/v1/models", nil)
			request.Header.Set("Authorization", "Bearer "+fixture.key)
			want := 401
			if boundary == "invitation" {
				body, err := json.Marshal(map[string]any{"token": invitation["acceptance_token"], "password": "controlled-queue-password"})
				if err != nil {
					test.Fatal(err)
				}
				request = httptest.NewRequest("POST", fixture.origin+"/admin/api/invitations/accept", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set("Origin", fixture.origin)
				want = 400
			}
			observed := make(chan struct{}, 1)
			fixture.store.clock = func() time.Time {
				now := time.Unix(0, fixture.clock.Load())
				select {
				case observed <- struct{}{}:
				default:
				}
				return now
			}
			fixture.server.mu.Lock()
			locked := true
			defer func() {
				if locked {
					fixture.server.mu.Unlock()
				}
			}()
			finished := make(chan resourceReceipt, 1)
			go func() {
				var response *http.Response
				if actual {
					request.RequestURI = ""
					var err error
					response, err = http.DefaultClient.Do(request)
					if err != nil {
						finished <- resourceReceipt{err: err}
						return
					}
				} else {
					writer := httptest.NewRecorder()
					fixture.server.ServeHTTP(writer, request)
					response = writer.Result()
				}
				_, err := io.Copy(io.Discard, response.Body)
				response.Body.Close()
				finished <- resourceReceipt{status: response.StatusCode, err: err}
			}()
			resourceWait(test, observed)
			fixture.clock.Store(start.Add(time.Hour).UnixNano())
			fixture.server.mu.Unlock()
			locked = false
			resourceFinish(test, finished, want)
			if fixture.hits.Load() != 0 || resourceUsage(test, fixture, "tenants/"+fixture.tenant).Calls != 0 {
				test.Fatal("queued expiry dispatched/debited")
			}
		})
	}
}
