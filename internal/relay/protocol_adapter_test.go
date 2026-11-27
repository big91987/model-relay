package relay

import (
	"context"
	"io"

	"model-relay/internal/records"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type eofOnClose struct {
	closed chan struct{}
	once   sync.Once
}

func (body *eofOnClose) Read([]byte) (int, error) { <-body.closed; return 0, io.EOF }
func (body *eofOnClose) Close() error             { body.once.Do(func() { close(body.closed) }); return nil }

func TestIdleClosedStreamEOFRemainsTimeout(t *testing.T) {
	gateway := Gateway{Config: &Config{IdleTimeout: 10 * time.Millisecond}}
	body := &eofOnClose{closed: make(chan struct{})}
	response := &http.Response{Body: body}
	request := httptest.NewRequest(http.MethodPost, "http://relay.test/v1/chat/completions", nil)
	writer := httptest.NewRecorder()
	var state, code string
	gateway.stream(writer, request, context.Background(), response, &records.Record{ID: "controlled"}, func(nextState, nextCode string) bool { state, code = nextState, nextCode; return true }, &latencyMeasurement{})
	if state != "timeout" || code != "upstream_timeout" || writer.Code != 504 {
		t.Fatal("idle EOF lost timeout attribution", state, code, writer.Code)
	}
}
