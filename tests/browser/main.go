package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"model-relay/internal/relay"
)

type capturedLog struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (capture *capturedLog) Write(data []byte) (int, error) {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return capture.data.Write(data)
}

func (capture *capturedLog) contains(secrets []string) bool {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	for _, secret := range secrets {
		if secret != "" && bytes.Contains(capture.data.Bytes(), []byte(secret)) {
			return true
		}
	}
	return false
}

func journey() error {
	capture := &capturedLog{}
	previousLog := log.Writer()
	log.SetOutput(capture)
	defer log.SetOutput(previousLog)
	root, err := os.MkdirTemp("", "model-relay-browser-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	password, err := relay.Init(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		return err
	}
	store, err := relay.Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		return err
	}
	defer func() { store.Close() }()
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("real browser upstream listener unavailable: %w", err)
	}
	var sensitiveFailures atomic.Int64
	var upstreamRequests atomic.Int64
	releaseStream := make(chan struct{}, 1)
	upstream := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		upstreamRequests.Add(1)
		if request.Header.Get("Authorization") != "Bearer controlled-browser-upstream" {
			http.Error(writer, "fixture authentication failed", 401)
			return
		}
		var input struct {
			Model    string
			Stream   bool
			Messages []struct{ Content string }
		}
		data, readErr := io.ReadAll(io.LimitReader(request.Body, (1<<20)+1))
		if readErr != nil || len(data) > 1<<20 || json.Unmarshal(data, &input) != nil || input.Model != "edited-controlled-model" {
			http.Error(writer, "fixture mapping failed", 400)
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "browser-error-prompt-sentinel" {
			sensitiveFailures.Add(1)
			http.Error(writer, "browser-sensitive-diagnostic-sentinel controlled-browser-upstream browser-error-prompt-sentinel controlled ordinary", 401)
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "protocol fixture" {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":`)
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "interrupt fixture" {
			writer.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled partial\"}}]}\n\n")
			writer.(http.Flusher).Flush()
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "unknown usage fixture" {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled unknown usage"}}]}`)
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "tool fixture" {
			if input.Stream {
				writer.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(writer, "data: {\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\",\"tool_calls\":[{\"index\":0,\"id\":\"call_weather\",\"type\":\"function\",\"function\":{\"name\":\"lookup_\",\"arguments\":\"\"}}]},\"finish_reason\":null}]}\n\n")
				writer.(http.Flusher).Flush()
				select {
				case <-releaseStream:
				case <-request.Context().Done():
					return
				}
				fmt.Fprint(writer, "data: {\"choices\":[{\"index\":0,\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"name\":\"weather\",\"arguments\":\"{}\"}}]},\"finish_reason\":null}]}\n\ndata: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"tool_calls\"}]}\n\ndata: [DONE]\n\n")
			} else {
				writer.Header().Set("Content-Type", "application/json")
				fmt.Fprint(writer, `{"choices":[{"index":0,"message":{"role":"assistant","content":null,"tool_calls":[{"id":"call_weather","type":"function","function":{"name":"lookup_weather","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`)
			}
			return
		}
		if input.Stream {
			writer.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled first\"}}]}\n\n")
			writer.(http.Flusher).Flush()
			if len(input.Messages) > 0 && input.Messages[0].Content == "cancel fixture" {
				<-request.Context().Done()
				return
			}
			if len(input.Messages) > 0 && input.Messages[0].Content == "stream fixture" {
				select {
				case <-releaseStream:
				case <-request.Context().Done():
					return
				}
			}
			fmt.Fprint(writer, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":0,\"completion_tokens\":2,\"total_tokens\":2}}\n\ndata: [DONE]\n\n")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled ordinary"}}],"usage":{"prompt_tokens":0,"completion_tokens":2,"total_tokens":2}}`)
	})}
	go upstream.Serve(upstreamListener)
	defer upstream.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("real browser service listener unavailable: %w", err)
	}
	server := relay.New(store, relay.Config{Version: "dev", Authority: listener.Addr().String(), Policy: relay.Policy{TestAuthority: upstreamListener.Addr().String()}})
	httpServer := &http.Server{Handler: server, ReadHeaderTimeout: 5 * time.Second}
	go httpServer.Serve(listener)
	defer func() { httpServer.Close() }()
	controlListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("isolated browser lifecycle listener unavailable: %w", err)
	}
	address := listener.Addr().String()
	control := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			writer.WriteHeader(405)
			return
		}
		switch request.URL.Path {
		case "/upstream-count":
			writer.Header().Set("X-Controlled-Upstream-Requests", fmt.Sprint(upstreamRequests.Load()))
			writer.WriteHeader(204)
		case "/release-stream":
			select {
			case releaseStream <- struct{}{}:
				writer.WriteHeader(204)
			default:
				writer.WriteHeader(409)
			}
		case "/restart":
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if httpServer.Shutdown(shutdown) != nil || store.Close() != nil {
				http.Error(writer, "controlled shutdown failed", 500)
				return
			}
			reopenedStore, openErr := relay.Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
			if openErr != nil {
				http.Error(writer, "controlled reopen failed", 500)
				return
			}
			store = reopenedStore
			restartedListener, listenErr := net.Listen("tcp", address)
			if listenErr != nil {
				http.Error(writer, "controlled restart listener failed", 500)
				return
			}
			server = relay.New(store, relay.Config{Version: "dev", Authority: address, Policy: relay.Policy{TestAuthority: upstreamListener.Addr().String()}})
			httpServer = &http.Server{Handler: server, ReadHeaderTimeout: 5 * time.Second}
			go httpServer.Serve(restartedListener)
			writer.WriteHeader(204)
		case "/audit-logs":
			var secrets []string
			if json.NewDecoder(io.LimitReader(request.Body, 64<<10)).Decode(&secrets) != nil || len(secrets) == 0 || capture.contains(secrets) {
				http.Error(writer, "controlled service log exclusion failed", 500)
				return
			}
			writer.Header().Set("X-Controlled-Sensitive-Failures", fmt.Sprint(sensitiveFailures.Load()))
			writer.WriteHeader(204)
		default:
			writer.WriteHeader(404)
		}
	})}
	go control.Serve(controlListener)
	defer control.Close()
	command := exec.Command("node", "tests/browser/journey.mjs")
	command.Env = append(os.Environ(), "RELAY_BROWSER_URL=http://"+address, "RELAY_BROWSER_PASSWORD="+password, "RELAY_BROWSER_UPSTREAM=http://"+upstreamListener.Addr().String()+"/v1", "RELAY_BROWSER_CONTROL=http://"+controlListener.Addr().String())
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
func main() {
	if err := journey(); err != nil {
		fmt.Fprintln(os.Stderr, strings.ReplaceAll(err.Error(), "\n", " "))
		os.Exit(1)
	}
}
