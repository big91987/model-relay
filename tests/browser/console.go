package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"model-relay/internal/relay"
)

func consoleJourney() error {
	root, err := os.MkdirTemp("", "relay-console-browser-")
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
	defer store.Close()
	upstreamListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("console controlled upstream listener unavailable: %w", err)
	}
	var dispatched atomic.Int64
	upstream := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer controlled-console-upstream" {
			writer.WriteHeader(401)
			return
		}
		if request.Method == "GET" && request.URL.Path == "/test/facts" {
			writer.Header().Set("Content-Type", "application/json")
			json.NewEncoder(writer).Encode(map[string]any{"dispatches": dispatched.Load()})
			return
		}
		if request.Method == "GET" && request.URL.Path == "/v1/models" {
			writer.Header().Set("Content-Type", "application/json")
			fmt.Fprint(writer, `{"object":"list","data":[{"id":"controlled-model","object":"model"}]}`)
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
		if len(input.Messages) > 0 && (input.Messages[0].Content == "controlled double normal" || input.Messages[0].Content == "controlled double SSE") {
			select {
			case <-time.After(350 * time.Millisecond):
			case <-request.Context().Done():
				return
			}
		}
		if input.Stream {
			writer.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(writer, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"controlled SSE response\"}}]}\r\n\r\n")
			writer.(http.Flusher).Flush()
			if len(input.Messages) > 0 && input.Messages[0].Content == "controlled cancel input" {
				select {
				case <-request.Context().Done():
					return
				case <-time.After(5 * time.Second):
					return
				}
			}
			fmt.Fprint(writer, "data: {\"choices\":[],\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":4,\"total_tokens\":7}}\r\n\r\ndata: [DONE]\r\n\r\n")
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		fmt.Fprint(writer, `{"id":"controlled-response","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"controlled upstream response"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":4,"total_tokens":7}}`)
	})}
	go upstream.Serve(upstreamListener)
	defer upstream.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("console real Go listener unavailable: %w", err)
	}
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: relay.New(store, relay.Config{Authority: listener.Addr().String(), Version: "console-controlled-verification", Policy: relay.Policy{TestAuthority: upstreamListener.Addr().String()}})}
	go server.Serve(listener)
	defer server.Close()
	command := exec.Command("node", "tests/browser/console.mjs")
	command.Env = append(os.Environ(), "RELAY_CONSOLE_ORIGIN=http://"+listener.Addr().String(), "RELAY_CONSOLE_PASSWORD="+password, "RELAY_CONSOLE_UPSTREAM=http://"+upstreamListener.Addr().String()+"/v1")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
