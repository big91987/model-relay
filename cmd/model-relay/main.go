package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"model-relay/internal/relay"
)

var buildVersion = "dev"

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: model-relay init|serve|backup|restore|reset-password")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	directory := flags.String("data-dir", "data", "installation directory")
	keyFile := flags.String("key-file", "relay.key", "separate encryption key")
	listen := flags.String("listen", "127.0.0.1:5544", "loopback listen authority")
	input := flags.String("input", "", "backup input")
	output := flags.String("output", "", "backup output")
	passwordStdin := flags.Bool("password-stdin", false, "read replacement password from standard input")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected arguments")
	}
	switch args[0] {
	case "init":
		password, err := relay.Init(*directory, *keyFile)
		if err != nil {
			return err
		}
		fmt.Println("Administrator password (shown once): " + password)
		return nil
	case "reset-password":
		if !*passwordStdin {
			return errors.New("--password-stdin required")
		}
		input, err := io.ReadAll(io.LimitReader(os.Stdin, 1027))
		if err != nil {
			return errors.New("cannot read password input")
		}
		password := strings.TrimSuffix(strings.TrimSuffix(string(input), "\n"), "\r")
		if err := relay.ResetAdminPassword(*directory, *keyFile, password); err != nil {
			return err
		}
		fmt.Println("Password reset complete.")
		return nil
	case "backup":
		if *output == "" {
			return errors.New("--output required")
		}
		return relay.Backup(*directory, *output)
	case "restore":
		if *input == "" {
			return errors.New("--input required")
		}
		return relay.Restore(*input, *directory, *keyFile)
	case "serve":
		host, _, err := net.SplitHostPort(*listen)
		if err != nil {
			return errors.New("invalid listen authority")
		}
		address := net.ParseIP(host)
		if address == nil || !address.IsLoopback() {
			return errors.New("v0.1 requires an explicit loopback address")
		}
		store, err := relay.Open(*directory, *keyFile)
		if err != nil {
			return err
		}
		defer store.Close()
		listener, err := net.Listen("tcp", *listen)
		if err != nil {
			return errors.New("listen failed; installation retained")
		}
		defer listener.Close()
		root, cancel := context.WithCancel(context.Background())
		defer cancel()
		server := &http.Server{Handler: relay.New(store, relay.Config{Authority: listener.Addr().String(), Version: buildVersion}), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 * 1024, BaseContext: func(net.Listener) context.Context { return root }}
		signals, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		done := make(chan struct{})
		shutdownDone := make(chan struct{})
		go func() {
			defer close(shutdownDone)
			select {
			case <-signals.Done():
				shutdown, release := context.WithTimeout(context.Background(), 10*time.Second)
				defer release()
				if server.Shutdown(shutdown) != nil {
					cancel()
					server.Close()
				}
			case <-done:
			}
		}()
		fmt.Println("Model Relay listening on http://" + listener.Addr().String() + "/admin/")
		err = server.Serve(listener)
		close(done)
		<-shutdownDone
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	default:
		return errors.New("unknown command")
	}
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Model Relay operation failed:", err)
		os.Exit(1)
	}
}
