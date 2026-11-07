package relay

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

type attemptOutput struct {
	writer      http.ResponseWriter
	parent      context.Context
	child       context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	timer       *time.Timer
	timedOut    bool
	committed   bool
	committedAt string
	header      http.Header
}

func newAttemptOutput(writer http.ResponseWriter, parent context.Context, duration time.Duration) *attemptOutput {
	child, cancel := context.WithCancel(parent)
	output := &attemptOutput{writer: writer, parent: parent, child: child, cancel: cancel, header: make(http.Header)}
	output.timer = time.AfterFunc(duration, func() {
		output.mu.Lock()
		defer output.mu.Unlock()
		if !output.committed {
			output.timedOut = true
			cancel()
		}
	})
	return output
}
func (output *attemptOutput) Header() http.Header    { return output.header }
func (output *attemptOutput) WriteHeader(status int) {}
func (output *attemptOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.header.Get("Content-Type") != "text/event-stream" {
		return len(data), nil
	}
	if output.parent.Err() != nil {
		return 0, output.parent.Err()
	}
	if output.timedOut {
		return 0, context.DeadlineExceeded
	}
	if !output.committed {
		if !output.timer.Stop() {
			return 0, context.DeadlineExceeded
		}
		output.committed = true
		output.committedAt = time.Now().UTC().Format(time.RFC3339Nano)
		output.writer.Header().Set("Content-Type", "text/event-stream")
	}
	return output.writer.Write(data)
}
func (output *attemptOutput) Flush() {
	if flusher, ok := output.writer.(http.Flusher); ok && output.isCommitted() {
		flusher.Flush()
	}
}
func (output *attemptOutput) FlushError() error {
	if output.isCommitted() {
		return http.NewResponseController(output.writer).Flush()
	}
	return nil
}
func (output *attemptOutput) Unwrap() http.ResponseWriter { return output.writer }
func (output *attemptOutput) isCommitted() bool {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.committed
}
func (output *attemptOutput) timeout() bool {
	output.mu.Lock()
	defer output.mu.Unlock()
	return output.timedOut
}
func (output *attemptOutput) stop() {
	output.mu.Lock()
	output.timer.Stop()
	output.mu.Unlock()
	output.cancel()
}
func (output *attemptOutput) commitNormal() error {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.parent.Err() != nil {
		return output.parent.Err()
	}
	if output.timedOut || !output.timer.Stop() {
		return context.DeadlineExceeded
	}
	output.committed = true
	output.committedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func transportTemporary(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, errDestinationForbidden)
}
