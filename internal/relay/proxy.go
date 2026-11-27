package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"model-relay/internal/catalog"
	"model-relay/internal/httpapi"
	"model-relay/internal/names"
	"model-relay/internal/records"
	"model-relay/internal/secure"
	provider "model-relay/internal/upstream"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

func (server *Gateway) Generate(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), server.Config.TotalTimeout)
	defer cancel()
	id := secure.Random(16)
	writer.Header().Set("X-Request-ID", id)
	started := time.Now()
	record := records.Record{ID: id, Time: started.UTC().Format(time.RFC3339Nano)}
	if server.Store.Schema == 3 {
		record.Time = server.Store.Now().Format(time.RFC3339Nano)
	}
	server.State.RLock()
	keyID, ok := server.identifyKey(request)
	server.State.RUnlock()
	record.KeyID = keyID
	if server.Degraded.Load() {
		httpapi.Failure(writer, 503, "storage_unavailable", id)
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
		if state != "success" {
			record.SuccessfulUpstream = nil
		}
		elapsed := time.Since(started).Milliseconds()
		record.Duration = &elapsed
		if server.Records.Finish(record) != nil {
			server.storageFailure()
			return false
		}
		server.State.Lock()
		server.State.CompleteRequest(id, configEpoch, state == "success")
		server.State.Unlock()
		return true
	}
	reject := func(status int, code string) {
		if server.Records.Start(record) != nil {
			server.storageFailure()
			httpapi.Failure(writer, 503, "storage_unavailable", id)
			return
		}
		if !finish("rejected", code) {
			httpapi.Failure(writer, 503, "storage_unavailable", id)
			return
		}
		httpapi.Failure(writer, status, code, id)
	}
	if !ok {
		if keyID != "" {
			controller := http.NewResponseController(writer)
			_ = controller.SetReadDeadline(time.Now().Add(10 * time.Second))
			data, readErr := io.ReadAll(http.MaxBytesReader(writer, request.Body, 1<<20))
			_ = controller.SetReadDeadline(time.Time{})
			if readErr == nil {
				_, alias, stream, code := provider.ValidateRequest(data)
				if code == "" && names.Alias(alias) {
					record.Alias, record.Stream = alias, stream
					record.Model = server.Catalog.FirstRemoteModel(request.Context(), alias)
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
	payload, alias, stream, code := provider.ValidateRequest(data)
	if names.Alias(alias) {
		record.Alias = alias
	}
	record.Stream = stream
	if code != "" {
		reject(400, code)
		return
	}
	if server.Store.Schema == 3 {
		if allowed, code := server.Keys.AuthorizeModel(keyID, alias); !allowed {
			status := 403
			if code == "storage_unavailable" {
				status = 503
			}
			reject(status, code)
			return
		}
	}
	server.State.Lock()
	keyID, ok = server.appID(request)
	if !ok {
		server.State.Unlock()
		reject(401, "unauthorized")
		return
	}
	snapshot, err := server.snapshot(alias)
	server.State.Unlock()
	if err != nil || !snapshot.Model.Enabled {
		if err != nil && !errors.Is(err, catalog.ErrModelMissing) {
			reject(503, "storage_unavailable")
		} else {
			reject(404, "model_unavailable")
		}
		return
	}
	configEpoch = snapshot.Epoch
	record.ConfigRevision = snapshot.ConfigRevision
	record.ModelRevision = snapshot.Model.Revision
	if server.Records.Start(record) != nil {
		server.storageFailure()
		httpapi.Failure(writer, 503, "storage_unavailable", id)
		return
	}
	if server.Store.Schema == 3 {
		server.State.Lock()
		code, status, prior := server.Admission.Admit(record, snapshot, payload, strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "), request.Header.Get("Idempotency-Key"))
		server.State.Unlock()
		if code != "" {
			finish("rejected", code)
			if prior != "" {
				writer.Header().Set("X-Request-ID", prior)
				httpapi.Failure(writer, status, code, prior)
				return
			}
			httpapi.Failure(writer, status, code, id)
			return
		}
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

func (server *Gateway) stream(writer http.ResponseWriter, request *http.Request, ctx context.Context, response *http.Response, record *records.Record, finish func(string, string) bool, measurement *latencyMeasurement) {
	stream := provider.NewStream(response.Body)
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
			httpapi.Failure(writer, status, code, record.ID)
			return
		}
		value, _ := json.Marshal(map[string]any{"error": map[string]string{"type": "relay_error", "code": code, "message": "Stream incomplete"}, "request_id": record.ID})
		_ = write("event: error\ndata: " + string(value) + "\n\n")
	}
	var terminalErr error
	for {
		event, err := stream.Next()
		if err != nil {
			terminalErr = err
			break
		}
		if event.Kind == provider.EventComment {
			if opened && write(event.Wire) != nil {
				finish("canceled", "downstream_write_failed")
				return
			}
			continue
		}
		if event.Kind == provider.EventDone {
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
				finish("canceled", "client_canceled")
				return
			}
			if write("data: [DONE]\n\n") != nil {
				finish("canceled", "downstream_write_failed")
			} else if request.Context().Err() != nil {
				finish("canceled", "client_canceled")
			}
			return
		}
		if event.HasUsage {
			record.Usage = records.Usage{Prompt: event.Usage.Prompt, Completion: event.Usage.Completion, Total: event.Usage.Total}
		}
		if !timer.Reset(server.Config.IdleTimeout) || timedOut.Load() {
			fail(504, "timeout", "upstream_timeout")
			return
		}
		measured := server.routingNow()
		effective := event.EffectiveOutput
		if write(event.Wire) != nil {
			finish("canceled", "downstream_write_failed")
			return
		}
		if effective {
			measurement.mark(measured)
		}
	}
	err := terminalErr
	if errors.Is(err, provider.ErrProtocol) && !errors.Is(err, provider.ErrIncomplete) {
		fail(502, "stream_interrupted", "upstream_protocol_error")
		return
	}
	if timedOut.Load() {
		err = context.DeadlineExceeded
	}
	if err == nil || errors.Is(err, provider.ErrIncomplete) {
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
