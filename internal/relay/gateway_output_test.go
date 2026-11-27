package relay

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAttemptOutputStopsTimerOnFirstCommit(t *testing.T) {
	response := httptest.NewRecorder()
	output := newAttemptOutput(response, context.Background(), 20*time.Millisecond)
	defer output.stop()
	output.Header().Set("Content-Type", "text/event-stream")
	if _, err := io.WriteString(output, "data: first\n\n"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	if output.timeout() || output.child.Err() != nil || !output.isCommitted() {
		t.Fatal("attempt timer canceled an already committed stream")
	}
}

func TestUnexpectedEOFClassification(t *testing.T) {
	if _, _, code := classify(context.Background(), io.ErrUnexpectedEOF, true); code != "stream_interrupted" {
		t.Fatal("unexpected EOF classification")
	}
}
