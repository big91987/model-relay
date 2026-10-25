package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"model-relay/internal/relay"
)

func fullConsoleJourneys() error {
	for _, width := range []int{1280, 1440, 390} {
		if err := fullConsoleJourney(width); err != nil {
			return fmt.Errorf("continuous console journey %d: %w", width, err)
		}
		if err := fullConsoleJourneyMode(width, "self_hosted"); err != nil {
			return fmt.Errorf("first self-hosted console journey %d: %w", width, err)
		}
	}
	return nil
}

func fullConsoleJourney(width int) error {
	return fullConsoleJourneyMode(width, "provider")
}

func fullConsoleJourneyMode(width int, mode string) error {
	root, err := os.MkdirTemp("", "relay-continuous-ui-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	password, err := relay.InitConsole(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		return err
	}
	store, err := relay.Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
	if err != nil {
		return err
	}
	defer func() { store.Close() }()
	var clock, dispatched atomic.Int64
	var restartGateway atomic.Value
	current := time.Now().UTC()
	clock.Store(time.Date(current.Year(), current.Month(), 1, 12, 0, 0, 0, time.UTC).UnixNano())
	capability := make([]byte, 32)
	if _, err := rand.Read(capability); err != nil {
		return err
	}
	controlToken := hex.EncodeToString(capability)
	controlListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	control := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Test-Capability") != controlToken {
			writer.WriteHeader(403)
			return
		}
		if request.Method == "POST" && request.URL.Path == "/restart" {
			action, ok := restartGateway.Load().(func() (string, error))
			if !ok {
				writer.WriteHeader(503)
				return
			}
			origin, err := action()
			if err != nil {
				writer.WriteHeader(500)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(map[string]string{"origin": origin})
			return
		}
		if request.Method == "POST" && request.URL.Path == "/advance" {
			var input struct {
				Seconds int64 `json:"seconds"`
			}
			if json.NewDecoder(http.MaxBytesReader(writer, request.Body, 1024)).Decode(&input) != nil || input.Seconds < 0 || input.Seconds > 35*24*3600 {
				writer.WriteHeader(400)
				return
			}
			clock.Add(input.Seconds * int64(time.Second))
		} else if request.Method != "GET" || request.URL.Path != "/facts" {
			writer.WriteHeader(404)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]any{"now": time.Unix(0, clock.Load()).UTC().Format(time.RFC3339Nano), "dispatches": dispatched.Load()})
	})}
	go control.Serve(controlListener)
	defer control.Close()
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	upstreamHandler := continuousUpstream(&dispatched)
	basePath := "/v1"
	if mode == "self_hosted" {
		upstreamHandler = selfHostedUpstream(&dispatched)
		basePath = selfHostedPath
	}
	upstream := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: upstreamHandler}
	go upstream.Serve(upstreamListener)
	defer upstream.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: relay.New(store, relay.Config{
		Authority: listener.Addr().String(), Version: "continuous-controlled-ui",
		Clock:  func() time.Time { return time.Unix(0, clock.Load()).UTC() },
		Policy: relay.Policy{TestAuthority: upstreamListener.Addr().String()},
	})}
	go server.Serve(listener)
	defer func() { server.Close() }()
	authority := listener.Addr().String()
	restartGateway.Store(func() (string, error) {
		server.Close()
		store.Close()
		var restartErr error
		store, restartErr = relay.Open(filepath.Join(root, "data"), filepath.Join(root, "key"))
		if restartErr != nil {
			return "", restartErr
		}
		replacement, restartErr := net.Listen("tcp", authority)
		if restartErr != nil {
			return "", restartErr
		}
		server = &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: relay.New(store, relay.Config{Authority: authority, Version: "continuous-controlled-ui", Clock: func() time.Time { return time.Unix(0, clock.Load()).UTC() }, Policy: relay.Policy{TestAuthority: upstreamListener.Addr().String()}})}
		go server.Serve(replacement)
		return "http://" + authority, nil
	})
	command := exec.Command("node", "tests/browser/journeys.mjs")
	command.Env = append(os.Environ(), "RELAY_JOURNEY_MODE="+mode, "RELAY_JOURNEY_WIDTH="+fmt.Sprint(width), "RELAY_JOURNEY_ORIGIN=http://"+listener.Addr().String(), "RELAY_JOURNEY_PASSWORD="+password, "RELAY_JOURNEY_UPSTREAM=http://"+upstreamListener.Addr().String()+basePath, "RELAY_JOURNEY_CONTROL=http://"+controlListener.Addr().String(), "RELAY_JOURNEY_CAPABILITY="+controlToken)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func continuousUpstream(dispatched *atomic.Int64) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		for _, prefix := range []string{"/latency-a", "/latency-b"} {
			if strings.HasPrefix(request.URL.Path, prefix+"/") {
				request.URL.Path = strings.TrimPrefix(request.URL.Path, prefix)
				break
			}
		}
		if request.Header.Get("Authorization") != "Bearer controlled-journey-upstream" {
			writer.WriteHeader(401)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		if request.Method == "GET" && request.URL.Path == "/v1/models" {
			fmt.Fprint(writer, `{"object":"list","data":[{"id":"controlled","object":"model"}]}`)
			return
		}
		if request.Method != "POST" || request.URL.Path != "/v1/chat/completions" {
			writer.WriteHeader(404)
			return
		}
		var input struct {
			Stream   bool `json:"stream"`
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if json.NewDecoder(request.Body).Decode(&input) != nil {
			writer.WriteHeader(400)
			return
		}
		dispatched.Add(1)
		if input.Stream {
			writer.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(writer, "data: {\"id\":\"controlled\",\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"controlled stream\"},\"finish_reason\":null}]}\n\n")
			writer.(http.Flusher).Flush()
			fmt.Fprint(writer, "data: {\"id\":\"controlled\",\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":4,\"total_tokens\":7}}\n\ndata: [DONE]\n\n")
			return
		}
		if len(input.Messages) > 0 && input.Messages[0].Content == "controlled unknown usage" {
			fmt.Fprint(writer, `{"choices":[{"message":{"role":"assistant","content":"controlled unknown"}}]}`)
			return
		}
		fmt.Fprint(writer, `{"id":"controlled","choices":[{"message":{"role":"assistant","content":"controlled response"}}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
	})
}

const selfHostedPath = "/isolated-controlled-self-hosted-long-address-readable-without-truncation/v1"

func selfHostedUpstream(dispatched *atomic.Int64) http.Handler {
	completion := continuousUpstream(dispatched)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/unreachable/models" {
			connection, _, err := writer.(http.Hijacker).Hijack()
			if err == nil {
				connection.Close()
			}
			return
		}
		if request.URL.Path != selfHostedPath+"/models" && request.URL.Path != selfHostedPath+"/chat/completions" {
			writer.WriteHeader(404)
			return
		}
		forwarded := request.Clone(request.Context())
		forwarded.URL.Path = "/v1" + request.URL.Path[len(selfHostedPath):]
		completion.ServeHTTP(writer, forwarded)
	})
}
