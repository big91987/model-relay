package relay

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

func (server *Server) generate(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), server.Config.TotalTimeout)
	defer cancel()
	id := randomSecret(16)
	writer.Header().Set("X-Request-ID", id)
	started := time.Now()
	record := Record{ID: id, Time: started.UTC().Format(time.RFC3339Nano)}
	server.mu.RLock()
	keyID, ok := server.identifyKey(request)
	server.mu.RUnlock()
	record.KeyID = keyID
	if server.degraded.Load() {
		failure(writer, 503, "storage_unavailable", id)
		return
	}
	var configEpoch uint64
	finish := func(state, code string) bool {
		if request.Context().Err() != nil {
			state = "canceled"
			code = "client_canceled"
		}
		record.State = state
		record.Code = code
		elapsed := time.Since(started).Milliseconds()
		record.Duration = &elapsed
		if server.Store.Finish(record) != nil {
			server.storageFailure()
			return false
		}
		server.mu.Lock()
		if state == "success" && configEpoch == server.configEpoch {
			server.successID = id
		} else if server.successID == id {
			server.successID = ""
		}
		server.mu.Unlock()
		return true
	}
	reject := func(status int, code string) {
		if server.Store.Start(record) != nil {
			server.storageFailure()
			failure(writer, 503, "storage_unavailable", id)
			return
		}
		if !finish("rejected", code) {
			failure(writer, 503, "storage_unavailable", id)
			return
		}
		failure(writer, status, code, id)
	}
	if !ok {
		if keyID != "" {
			controller := http.NewResponseController(writer)
			_ = controller.SetReadDeadline(time.Now().Add(10 * time.Second))
			data, readErr := io.ReadAll(http.MaxBytesReader(writer, request.Body, 1<<20))
			_ = controller.SetReadDeadline(time.Time{})
			if readErr == nil {
				_, alias, stream, code := validateRequest(data)
				if code == "" && validAlias(alias) {
					record.Alias, record.Stream = alias, stream
					_ = server.Store.DB.QueryRow("SELECT remote_model FROM route_candidates WHERE alias=? ORDER BY upstream_id LIMIT 1", alias).Scan(&record.Model)
				}
			}
		}
		reject(401, "unauthorized")
		return
	}
	kind, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err != nil || kind != "application/json" {
		reject(400, "invalid_request")
		return
	}
	controller := http.NewResponseController(writer)
	_ = controller.SetReadDeadline(time.Now().Add(10 * time.Second))
	data, err := io.ReadAll(http.MaxBytesReader(writer, request.Body, 1<<20))
	_ = controller.SetReadDeadline(time.Time{})
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			reject(413, "request_too_large")
		} else {
			reject(400, "invalid_request")
		}
		return
	}
	payload, alias, stream, code := validateRequest(data)
	if validAlias(alias) {
		record.Alias = alias
	}
	record.Stream = stream
	if code != "" {
		reject(400, code)
		return
	}
	server.mu.Lock()
	keyID, ok = server.appID(request)
	if !ok {
		server.mu.Unlock()
		reject(401, "unauthorized")
		return
	}
	snapshot, err := server.snapshot(alias)
	server.mu.Unlock()
	if err != nil || !snapshot.Model.Enabled {
		if err != nil && err != sql.ErrNoRows {
			reject(503, "storage_unavailable")
		} else {
			reject(404, "model_unavailable")
		}
		return
	}
	configEpoch = snapshot.Epoch
	record.ConfigRevision = snapshot.ConfigRevision
	record.ModelRevision = snapshot.Model.Revision
	if server.Store.Start(record) != nil {
		server.storageFailure()
		failure(writer, 503, "storage_unavailable", id)
		return
	}
	server.routedGenerate(writer, request, ctx, &record, payload, snapshot, finish)
}
func classify(ctx context.Context, err error, stream bool) (int, string, string) {
	if errors.Is(ctx.Err(), context.Canceled) {
		return 499, "canceled", "client_canceled"
	}
	var network net.Error
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) || errors.As(err, &network) && network.Timeout() {
		return 504, "timeout", "upstream_timeout"
	}
	if stream {
		return 502, "stream_interrupted", "stream_interrupted"
	}
	return 502, "upstream_failed", "upstream_connection_failed"
}
func (server *Server) stream(writer http.ResponseWriter, request *http.Request, ctx context.Context, response *http.Response, record *Record, finish func(string, string) bool) {
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 4096), 256*1024)
	var timedOut atomic.Bool
	timer := time.AfterFunc(server.Config.IdleTimeout, func() { timedOut.Store(true); response.Body.Close() })
	defer timer.Stop()
	opened := false
	controller := http.NewResponseController(writer)
	write := func(data string) error {
		setWriteDeadline(writer, ctx)
		if !opened {
			writer.Header().Set("Content-Type", "text/event-stream")
			opened = true
		}
		if _, err := io.WriteString(writer, data); err != nil {
			return err
		}
		return controller.Flush()
	}
	fail := func(status int, state, code string) {
		if !finish(state, code) {
			status = 503
			code = "storage_unavailable"
		}
		if !opened {
			failure(writer, status, code, record.ID)
			return
		}
		value, _ := json.Marshal(map[string]any{"error": map[string]string{"type": "relay_error", "code": code, "message": "Stream incomplete"}, "request_id": record.ID})
		_ = write("event: error\ndata: " + string(value) + "\n\n")
	}
	var lines []string
	eventBytes := 0
	total := 0
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		eventBytes += len(line) + 1
		total += len(line) + 1
		if eventBytes > 256*1024 || total > 32<<20 || !utf8.ValidString(line) {
			fail(502, "stream_interrupted", "upstream_protocol_error")
			return
		}
		if line != "" {
			if !strings.HasPrefix(line, ":") && line != "data" && !strings.HasPrefix(line, "data:") {
				fail(502, "stream_interrupted", "upstream_protocol_error")
				return
			}
			lines = append(lines, line)
			continue
		}
		var dataLines []string
		var comments []string
		for _, entry := range lines {
			if strings.HasPrefix(entry, ":") {
				comments = append(comments, entry)
			} else {
				value := strings.TrimPrefix(entry, "data")
				value = strings.TrimPrefix(value, ":")
				value = strings.TrimPrefix(value, " ")
				dataLines = append(dataLines, value)
			}
		}
		lines = nil
		eventBytes = 0
		if len(dataLines) == 0 {
			if opened && len(comments) > 0 {
				if write(strings.Join(comments, "\n")+"\n\n") != nil {
					finish("canceled", "downstream_write_failed")
					return
				}
			}
			continue
		}
		payload := strings.Join(dataLines, "\n")
		if payload == "[DONE]" {
			if !opened {
				fail(502, "stream_interrupted", "upstream_protocol_error")
				return
			}
			if ctx.Err() != nil || timedOut.Load() {
				status, state, code := classify(ctx, context.DeadlineExceeded, true)
				fail(status, state, code)
				return
			}
			if !finish("success", "") {
				value := `event: error` + "\ndata: {\"error\":{\"code\":\"storage_unavailable\"}}\n\n"
				_ = write(value)
				return
			}
			if request.Context().Err() != nil {
				return
			}
			if write("data: [DONE]\n\n") != nil {
				finish("canceled", "downstream_write_failed")
			}
			return
		}
		value, ok := validResponse([]byte(payload), true)
		if !ok {
			fail(502, "stream_interrupted", "upstream_protocol_error")
			return
		}
		if _, exists := object(value["usage"]); exists {
			record.Usage = usageFrom(value)
		}
		if !timer.Reset(server.Config.IdleTimeout) || timedOut.Load() {
			fail(504, "timeout", "upstream_timeout")
			return
		}
		if write("data: "+strings.Join(dataLines, "\ndata: ")+"\n\n") != nil {
			finish("canceled", "downstream_write_failed")
			return
		}
	}
	err := scanner.Err()
	if timedOut.Load() {
		err = context.DeadlineExceeded
	}
	if err == nil {
		fail(502, "stream_interrupted", "upstream_protocol_error")
		return
	}
	status, state, code := classify(ctx, err, true)
	if timedOut.Load() {
		status = 504
		state = "timeout"
		code = "upstream_timeout"
	}
	fail(status, state, code)
}
