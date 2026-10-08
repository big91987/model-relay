package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"
)

func (server *Server) routedGenerate(writer http.ResponseWriter, request *http.Request, ctx context.Context, record *Record, payload map[string]json.RawMessage, snapshot routeSnapshot, finish func(string, string) bool) {
	tried := map[int64]bool{}
	respond := func(status int, state, code string) {
		if request.Context().Err() != nil {
			finish("canceled", "client_canceled")
			return
		}
		if ctx.Err() != nil {
			status = 504
			state = "timeout"
			code = "request_timeout"
		}
		if !finish(state, code) {
			status = 503
			code = "storage_unavailable"
		}
		failure(writer, status, code, record.ID)
	}
	for round := 1; round <= 4; round++ {
		if ctx.Err() != nil {
			respond(504, "timeout", "request_timeout")
			return
		}
		candidate, attempt, code := server.beginAttempt(ctx, request, record, snapshot, tried, round)
		if code != "" {
			status := 503
			state := "rejected"
			if code == "storage_unavailable" {
				server.storageFailure()
			}
			if code == "unauthorized" {
				status = 401
			}
			if code == "tenant_paused" || code == "model_forbidden" {
				status = 403
			}
			if code == "request_timeout" {
				status = 504
				state = "timeout"
			}
			respond(status, state, code)
			return
		}
		if candidate == nil {
			if len(tried) == 0 {
				respond(503, "rejected", "upstream_unavailable")
			} else {
				respond(502, "upstream_failed", "upstream_attempts_exhausted")
			}
			return
		}
		if len(tried) == 3 {
			respond(502, "upstream_failed", "upstream_attempts_exhausted")
			return
		}
		tried[candidate.UpstreamID] = true
		upstream := snapshot.Upstreams[candidate.UpstreamID]
		record.Model = candidate.RemoteModel
		record.Revision = upstream.Revision
		record.Status = nil
		record.Usage = Usage{}
		record.SuccessfulUpstream = nil
		token, err := unseal(server.Store.Key, upstream.Encrypted, upstreamAAD(upstream.ID, upstream.URL))
		output := newAttemptOutput(writer, ctx, server.Config.AttemptTimeout)
		started := time.Now()
		temporary := false
		status := 502
		state := "upstream_failed"
		errorCode := "upstream_connection_failed"
		persist := func(resultState, resultCode string) bool {
			attempt.State = resultState
			attempt.Code = resultCode
			attempt.Status = record.Status
			attempt.Usage = record.Usage
			ended := time.Now().UTC().Format(time.RFC3339Nano)
			elapsed := time.Since(started).Milliseconds()
			attempt.Finished = &ended
			attempt.Duration = &elapsed
			if server.Store.finishAttempt(record.ID, attempt) != nil {
				server.storageFailure()
				return false
			}
			return true
		}
		if err != nil {
			persist("rejected", "storage_unavailable")
			output.stop()
			server.storageFailure()
			respond(503, "rejected", "storage_unavailable")
			return
		}
		payload["model"], _ = json.Marshal(candidate.RemoteModel)
		body, err := json.Marshal(payload)
		if err != nil {
			persist("rejected", "invalid_request")
			output.stop()
			respond(400, "rejected", "invalid_request")
			return
		}
		outgoing, err := http.NewRequestWithContext(output.child, http.MethodPost, upstream.URL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			persist("rejected", "upstream_connection_failed")
			output.stop()
			respond(502, "rejected", "upstream_connection_failed")
			return
		}
		outgoing.GetBody = nil
		outgoing.Header.Set("Content-Type", "application/json")
		outgoing.Header.Set("Authorization", "Bearer "+token)
		outgoing.Header.Set("X-Request-ID", record.ID)
		outgoing.Header.Set("Accept", "application/json")
		if record.Stream {
			outgoing.Header.Set("Accept", "text/event-stream")
		}
		if server.Store.Schema == 3 {
			if _, err := server.Store.DB.Exec("UPDATE attempt_dispatches SET possible=1 WHERE request_id=? AND ordinal=?", record.ID, attempt.Ordinal); err != nil {
				persist("rejected", "storage_unavailable")
				output.stop()
				server.storageFailure()
				respond(503, "rejected", "storage_unavailable")
				return
			}
		}
		response, err := server.client.Do(outgoing)
		if err != nil {
			status, state, errorCode = classify(output.child, err, false)
			temporary = transportTemporary(err) || output.timeout()
		} else {
			record.Status = &response.StatusCode
			if response.StatusCode != 200 {
				if server.Store.Schema == 3 {
					data, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
					if err == nil && len(data) <= 1<<20 && uniqueJSON(data) {
						if value, ok := object(data); ok {
							record.Usage = usageFrom(value)
						}
					}
				}
				temporary = temporaryStatus(response.StatusCode)
				errorCode = "upstream_http_error"
				if response.StatusCode >= 300 && response.StatusCode < 400 {
					errorCode = "upstream_redirect"
				} else if response.StatusCode == 429 {
					status = 429
					state = "upstream_rejected"
					errorCode = "upstream_rate_limited"
				} else if response.StatusCode >= 400 && response.StatusCode < 500 {
					state = "upstream_rejected"
					errorCode = "upstream_rejected"
				}
			} else {
				kind, _, parseErr := mime.ParseMediaType(response.Header.Get("Content-Type"))
				expected := "application/json"
				if record.Stream {
					expected = "text/event-stream"
				}
				if parseErr != nil || kind != expected {
					errorCode = "upstream_protocol_error"
				} else if record.Stream {
					var saved bool
					streamFinish := func(resultState, resultCode string) bool {
						state = resultState
						errorCode = resultCode
						if output.timeout() {
							state = "timeout"
							errorCode = "upstream_timeout"
						}
						if request.Context().Err() != nil {
							state = "canceled"
							errorCode = "client_canceled"
						}
						if request.Context().Err() == nil && ctx.Err() != nil {
							state = "timeout"
							errorCode = "request_timeout"
						}
						saved = persist(state, errorCode)
						if !saved {
							return false
						}
						if output.isCommitted() {
							record.CommittedAt = &output.committedAt
						}
						if state == "success" {
							record.SuccessfulUpstream = &attempt.UpstreamID
						} else {
							record.SuccessfulUpstream = nil
						}
						if output.isCommitted() || state == "success" {
							return finish(state, errorCode)
						}
						return true
					}
					server.stream(output, request, output.child, response, record, streamFinish)
					temporary = state == "timeout" || errorCode == "stream_interrupted"
					if output.isCommitted() {
						response.Body.Close()
						output.stop()
						if saved {
							server.observe(snapshot, *candidate, *attempt, temporary && ctx.Err() == nil)
						}
						return
					}
					if !saved {
						response.Body.Close()
						output.stop()
						respond(503, "upstream_failed", "storage_unavailable")
						return
					}
					status = 502
					if state == "timeout" {
						status = 504
					}
				} else {
					result, readErr := io.ReadAll(io.LimitReader(response.Body, (8<<20)+1))
					if readErr != nil {
						status, state, errorCode = classify(output.child, readErr, false)
						temporary = true
					} else {
						value, valid := validResponse(result, false)
						if !valid || len(result) > 8<<20 {
							errorCode = "upstream_protocol_error"
						} else {
							record.Usage = usageFrom(value)
							if err = output.commitNormal(); err == nil {
								record.CommittedAt = &output.committedAt
								record.SuccessfulUpstream = &attempt.UpstreamID
								if !persist("success", "") || !finish("success", "") {
									response.Body.Close()
									output.stop()
									failure(writer, 503, "storage_unavailable", record.ID)
									return
								}
								server.observe(snapshot, *candidate, *attempt, false)
								if request.Context().Err() == nil {
									writer.Header().Set("Content-Type", "application/json")
									setWriteDeadline(writer, ctx)
									if _, err = writer.Write(result); err != nil {
										record.SuccessfulUpstream = nil
										persist("canceled", "downstream_write_failed")
										finish("canceled", "downstream_write_failed")
									}
								}
								response.Body.Close()
								output.stop()
								return
							}
							status = 504
							state = "timeout"
							errorCode = "upstream_timeout"
							temporary = true
						}
					}
				}
			}
			response.Body.Close()
		}
		if output.timeout() {
			status = 504
			state = "timeout"
			errorCode = "upstream_timeout"
			temporary = true
		}
		output.stop()
		if request.Context().Err() != nil {
			state = "canceled"
			errorCode = "client_canceled"
			temporary = false
		} else if ctx.Err() != nil {
			state = "timeout"
			errorCode = "request_timeout"
			temporary = false
			status = 504
		}
		if !persist(state, errorCode) {
			respond(503, "upstream_failed", "storage_unavailable")
			return
		}
		if request.Context().Err() == nil && ctx.Err() == nil {
			server.observe(snapshot, *candidate, *attempt, temporary)
		}
		if !temporary {
			respond(status, state, errorCode)
			return
		}
		if len(tried) == 3 {
			respond(502, "upstream_failed", "upstream_attempts_exhausted")
			return
		}
	}
}

func setWriteDeadline(writer http.ResponseWriter, ctx context.Context) {
	deadline := time.Now().Add(10 * time.Second)
	if total, ok := ctx.Deadline(); ok && total.Before(deadline) {
		deadline = total
	}
	_ = http.NewResponseController(writer).SetWriteDeadline(deadline)
}
