package httpconsole

import (
	"model-relay/internal/reporting"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReportingSnapshotDoesNotWaitForConfigurationWriter(t *testing.T) {
	controller, db := identityController(t)
	controller.Reports = reporting.New(db, time.Now)
	controller.State.Lock()
	finished := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		finished <- facadeRequest(controller, "GET", "/admin/api/audit?tenant=tenant", "", "admin", "", "")
	}()
	select {
	case response := <-finished:
		controller.State.Unlock()
		if response.Code != 200 {
			t.Fatalf("authorized report: %d %s", response.Code, response.Body.String())
		}
	case <-time.After(time.Second):
		controller.State.Unlock()
		<-finished
		t.Fatal("report read waited for unrelated configuration writer")
	}
}
