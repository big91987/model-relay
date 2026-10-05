package relay

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCrashProcessHelper(t *testing.T) {
	if os.Getenv("MODEL_RELAY_TEST_CRASH_HELPER") != "1" {
		return
	}
	store, err := Open(os.Getenv("MODEL_RELAY_TEST_DATA"), os.Getenv("MODEL_RELAY_TEST_KEY"))
	if err != nil {
		t.Fatal("helper open failed")
	}
	defer store.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("helper listener unavailable")
	}
	server := New(store, Config{Authority: listener.Addr().String(), Policy: Policy{TestAuthority: os.Getenv("MODEL_RELAY_TEST_AUTHORITY")}})
	fmt.Println("HELPER_URL=http://" + listener.Addr().String())
	httpServer := &http.Server{Handler: server, ReadHeaderTimeout: 5 * time.Second}
	if err = httpServer.Serve(listener); err != http.ErrServerClosed {
		t.Fatal("helper serve failed")
	}
}

func TestProcessCrashRecovery(t *testing.T) {
	entered := make(chan struct{})
	canceled := make(chan struct{})
	value := newFixture(t, func(writer http.ResponseWriter, request *http.Request) {
		close(entered)
		<-request.Context().Done()
		close(canceled)
	})
	value.http.Close()
	if err := value.store.Close(); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.run=^TestCrashProcessHelper$")
	command.Env = append(os.Environ(), "MODEL_RELAY_TEST_CRASH_HELPER=1", "MODEL_RELAY_TEST_DATA="+value.directory, "MODEL_RELAY_TEST_KEY="+value.keyFile, "MODEL_RELAY_TEST_AUTHORITY="+strings.TrimPrefix(value.upstream.URL, "http://"))
	output, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	command.Stderr = io.Discard
	if err = command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { command.Process.Kill(); command.Wait() })
	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "HELPER_URL=") {
				ready <- strings.TrimPrefix(scanner.Text(), "HELPER_URL=")
				return
			}
		}
		ready <- ""
	}()
	var address string
	select {
	case address = <-ready:
		if address == "" {
			t.Fatal("crash helper failed to start")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("crash helper startup timeout")
	}
	completed := make(chan struct{})
	go func() {
		request, _ := http.NewRequest("POST", address+"/v1/chat/completions", strings.NewReader(ordinary))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+value.key)
		response, _ := http.DefaultClient.Do(request)
		if response != nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		close(completed)
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("child request did not reach upstream")
	}
	if err = command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err = command.Wait(); err == nil {
		t.Fatal("crash did not terminate abnormally")
	}
	select {
	case <-completed:
	case <-time.After(5 * time.Second):
		t.Fatal("crash did not close client")
	}
	select {
	case <-canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("crash did not terminate upstream")
	}
	reopened, err := Open(value.directory, value.keyFile)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	var state string
	var duration *int64
	if err = reopened.DB.QueryRow("SELECT state,duration_ms FROM requests ORDER BY rowid DESC LIMIT 1").Scan(&state, &duration); err != nil || state != "process_interrupted" || duration != nil {
		t.Fatal("crashed durable request falsely finalized")
	}
}
