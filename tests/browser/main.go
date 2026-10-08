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
	var routePrimary, routeSecondary, routeChecks atomic.Int64
	var routeCheckMode, routeCheckCanceled atomic.Int64
	var routeFailure atomic.Bool
	releaseStream := make(chan struct{}, 1)
	upstream := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/routing-") {
			if request.Header.Get("Authorization") != "Bearer controlled-browser-routing" {
				writer.WriteHeader(401)
				return
			}
			if request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/models") {
				routeChecks.Add(1)
				writer.Header().Set("Content-Type", "application/json")
				switch routeCheckMode.Load() {
				case 1:
					fmt.Fprint(writer, `{"data":[]}`)
				case 2:
					writer.WriteHeader(401)
				case 3:
					writer.WriteHeader(429)
				case 4:
					<-request.Context().Done()
					routeCheckCanceled.Add(1)
				default:
					fmt.Fprint(writer, `{"object":"list","data":[{"id":"controlled-listing"}]}`)
				}
				return
			}
			var input struct {
				Model    string
				Stream   bool
				Messages []struct{ Content string }
			}
			if json.NewDecoder(request.Body).Decode(&input) != nil {
				writer.WriteHeader(400)
				return
			}
			primary := strings.HasPrefix(request.URL.Path, "/routing-a/")
			expected := "routing-b-model"
			if primary {
				routePrimary.Add(1)
				expected = "routing-a-model"
			} else {
				routeSecondary.Add(1)
			}
			if input.Model != expected {
				writer.WriteHeader(400)
				return
			}
			if primary && routeFailure.Load() {
				writer.WriteHeader(503)
				return
			}
			if input.Stream {
				writer.Header().Set("Content-Type", "text/event-stream")
				fmt.Fprint(writer, "data: {\"choices\":[{\"delta\":{\"content\":\"controlled routed first\"}}]}\n\n")
				writer.(http.Flusher).Flush()
				if len(input.Messages) > 0 && input.Messages[0].Content == "route cancel" {
					<-request.Context().Done()
					return
				}
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled routed response"}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`)
			return
		}
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
		case "/route/reset":
			routePrimary.Store(0)
			routeSecondary.Store(0)
			routeChecks.Store(0)
			writer.WriteHeader(204)
		case "/route/counts":
			writer.Header().Set("X-Check-Canceled", fmt.Sprint(routeCheckCanceled.Load()))
			writer.Header().Set("X-Primary", fmt.Sprint(routePrimary.Load()))
			writer.Header().Set("X-Secondary", fmt.Sprint(routeSecondary.Load()))
			writer.Header().Set("X-Checks", fmt.Sprint(routeChecks.Load()))
			writer.WriteHeader(204)
		case "/route/fail":
			routeFailure.Store(true)
			writer.WriteHeader(204)
		case "/route/recover":
			routeFailure.Store(false)
		case "/route/check-valid":
			routeCheckMode.Store(0)
		case "/route/check-invalid":
			routeCheckMode.Store(1)
		case "/route/check-auth":
			routeCheckMode.Store(2)
		case "/route/check-rate":
			routeCheckMode.Store(3)
		case "/route/check-delay":
			routeCheckMode.Store(4)
			writer.WriteHeader(204)
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
	if err = browserPhase(os.Stdout, "legacy-ui", command.Run); err != nil {
		return err
	}
	routes := exec.Command("node", "tests/browser/routing.mjs")
	routes.Env = command.Env
	routes.Stdout = os.Stdout
	routes.Stderr = os.Stderr
	return browserPhase(os.Stdout, "routing", routes.Run)
}

func browserPhase(output io.Writer, name string, action func() error) error {
	started := time.Now()
	fmt.Fprintf(output, "BROWSER_PHASE name=%s state=start utc=%s\n", name, started.UTC().Format(time.RFC3339Nano))
	err := action()
	fmt.Fprintf(output, "BROWSER_PHASE name=%s state=end utc=%s elapsed_ms=%d ok=%t\n", name, time.Now().UTC().Format(time.RFC3339Nano), time.Since(started).Milliseconds(), err == nil)
	return err
}

func main() {
	if err := browserPhase(os.Stdout, "migration", func() error { return migrationJourney() }); err != nil {
		fmt.Fprintln(os.Stderr, "migration browser verification failed:", err)
		os.Exit(1)
	}
	if err := journey(); err != nil {
		fmt.Fprintln(os.Stderr, strings.ReplaceAll(err.Error(), "\n", " "))
		os.Exit(1)
	}
	if err := browserPhase(os.Stdout, "console", func() error { return consoleJourney() }); err != nil {
		fmt.Fprintln(os.Stderr, "console browser verification failed:", err)
		os.Exit(1)
	}
	if err := browserPhase(os.Stdout, "full-console", func() error { return fullConsoleJourneys() }); err != nil {
		fmt.Fprintln(os.Stderr, "continuous console browser verification failed:", err)
		os.Exit(1)
	}
}
