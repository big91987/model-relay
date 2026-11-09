package relay

import (
	"context"
	"encoding/json"
	"errors"
	"model-relay/internal/catalog"
	"model-relay/internal/httpapi"
	"model-relay/internal/records"
	"model-relay/internal/secure"
	provider "model-relay/internal/upstream"
	"net/http"
	"time"
)

func (server *Gateway) routedGenerate(writer http.ResponseWriter, request *http.Request, ctx context.Context, record *records.Record, payload map[string]json.RawMessage, snapshot catalog.Snapshot, finish func(string, string) bool) {
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
		httpapi.Failure(writer, status, code, record.ID)
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
		record.Usage = records.Usage{}
		record.SuccessfulUpstream = nil
		token, err := secure.Unseal(server.Store.Key, upstream.Encrypted, secure.UpstreamAAD(upstream.ID, upstream.URL))
		output := newAttemptOutput(writer, ctx, server.Config.AttemptTimeout)
		started := time.Now()
		measurement := &latencyMeasurement{}
		defer func(attempt *records.Attempt, candidate catalog.Candidate, measurement *latencyMeasurement) {
			server.acceptLatency(ctx, record.ID, snapshot, candidate, *attempt, record.Stream, *measurement)
		}(attempt, *candidate, measurement)
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
			if server.Records.FinishAttempt(record.ID, attempt) != nil {
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
		prepared, prepareErr := provider.PrepareChat(output.child, provider.ChatInput{Endpoint: provider.Endpoint{URL: upstream.URL, Token: token}, Payload: payload, Model: candidate.RemoteModel, RequestID: record.ID, Stream: record.Stream})
		if prepareErr != nil {
			code, preparationStatus := "upstream_connection_failed", 502
			var failure *provider.PreparationError
			if errors.As(prepareErr, &failure) && failure.Code == "invalid_request" {
				code, preparationStatus = "invalid_request", 400
			}
			persist("rejected", code)
			output.stop()
			respond(preparationStatus, "rejected", code)
			return
		}
		if err := server.Records.MarkDispatch(record.ID, attempt.Ordinal); err != nil {
			persist("rejected", "storage_unavailable")
			output.stop()
			server.storageFailure()
			respond(503, "rejected", "storage_unavailable")
			return
		}
		measurement.Started = server.routingNow()
		response, err := server.Upstream.Send(prepared)
		if response != nil && !record.Stream {
			response.Body = &latencyReader{ReadCloser: response.Body, Measurement: measurement, Now: server.routingNow}
		}
		if err != nil {
			status, state, errorCode = classify(output.child, err, false)
			temporary = transportTemporary(err) || output.timeout()
		} else {
			record.Status = &response.StatusCode
			if response.StatusCode != 200 {
				if server.Store.Schema == 3 {
					usage := provider.ReadErrorUsage(response.Body)
					record.Usage = records.Usage{Prompt: usage.Prompt, Completion: usage.Completion, Total: usage.Total}
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
				if provider.ValidateChatResponse(response, record.Stream) != nil {
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
					server.stream(output, request, output.child, response, record, streamFinish, measurement)
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
					completion, readErr := provider.ReadCompletion(response)
					result := completion.Data
					if readErr != nil {
						if errors.Is(readErr, provider.ErrProtocol) {
							errorCode = "upstream_protocol_error"
						} else {
							status, state, errorCode = classify(output.child, readErr, false)
							temporary = true
						}
					} else {
						{
							record.Usage = records.Usage{Prompt: completion.Usage.Prompt, Completion: completion.Usage.Completion, Total: completion.Usage.Total}
							if err = output.commitNormal(); err == nil {
								record.CommittedAt = &output.committedAt
								record.SuccessfulUpstream = &attempt.UpstreamID
								if !persist("success", "") || !finish("success", "") {
									response.Body.Close()
									output.stop()
									httpapi.Failure(writer, 503, "storage_unavailable", record.ID)
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
