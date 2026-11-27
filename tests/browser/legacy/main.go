package main

import (
	"context"
	"flag"
	"fmt"
	"model-relay/internal/app"
	"model-relay/internal/installation"

	"model-relay/internal/relay"
	upstreamclient "model-relay/internal/upstream"

	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	directory := flag.String("data-dir", "", "isolated installation")
	key := flag.String("key-file", "", "isolated material")
	authority := flag.String("test-authority", "", "explicit isolated controlled upstream")
	flag.Parse()
	store, err := installation.Open(*directory, *key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixture installation unavailable")
		os.Exit(1)
	}
	defer store.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fixture listener unavailable")
		os.Exit(1)
	}
	server := &http.Server{Handler: app.New(store, relay.Config{Version: "isolated-migration-fixture", Authority: listener.Addr().String(), Policy: upstreamclient.Policy{TestAuthority: *authority}}), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(cleanup)
	}()
	fmt.Println("ISOLATED http://" + listener.Addr().String())
	if err = server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, "fixture service failed")
		os.Exit(1)
	}
}
