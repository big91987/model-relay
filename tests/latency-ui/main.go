package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"time"

	"model-relay/tests/latencyfixture"
)

func journey(width int) error {
	fixture, err := latencyfixture.New()
	if err != nil {
		return err
	}
	defer fixture.Close()
	fixture.Register("controlled", &latencyfixture.Scenario{})
	team, err := fixture.Admin("POST", "tenants", map[string]any{"name": "Latency focused UI team"}, 201)
	if err != nil {
		return err
	}
	tenant, _ := team["id"].(string)
	capability := make([]byte, 32)
	if _, err = rand.Read(capability); err != nil {
		return err
	}
	token := hex.EncodeToString(capability)
	control := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" || request.Header.Get("X-Test-Capability") != token {
			writer.WriteHeader(403)
			return
		}
		if err := fixture.Restart(); err != nil {
			writer.WriteHeader(500)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(map[string]string{"origin": fixture.Gateway.URL})
	}))
	defer control.Close()
	command := exec.Command("node", "tests/browser/latency-focus.mjs")
	command.Env = append(os.Environ(), "RELAY_LATENCY_WIDTH="+fmt.Sprint(width), "RELAY_LATENCY_ORIGIN="+fixture.Gateway.URL, "RELAY_LATENCY_UPSTREAM="+fixture.Upstream.URL, "RELAY_LATENCY_PASSWORD="+fixture.Password, "RELAY_LATENCY_CREDENTIAL="+latencyfixture.Credential, "RELAY_LATENCY_TENANT="+tenant, "RELAY_LATENCY_CONTROL="+control.URL, "RELAY_LATENCY_CAPABILITY="+token)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	started := time.Now()
	err = command.Run()
	fmt.Printf("LATENCY_UI width=%d elapsed_ms=%d ok=%t\n", width, time.Since(started).Milliseconds(), err == nil)
	if err != nil {
		return errors.New("focused latency UI failed; see safe receipt")
	}
	return nil
}
func main() {
	for _, width := range []int{1280, 390} {
		if err := journey(width); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
