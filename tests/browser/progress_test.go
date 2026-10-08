package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestBrowserPhasePreservesResult(t *testing.T) {
	for _, failure := range []error{nil, errors.New("secret-error")} {
		var output bytes.Buffer
		calls := 0
		result := browserPhase(&output, "routing", func() error {
			calls++
			return failure
		})
		if result != failure || calls != 1 {
			t.Fatal("diagnostic changed action or error")
		}
		text := output.String()
		if !strings.Contains(text, "state=start utc=") || !strings.Contains(text, "state=end utc=") || !strings.Contains(text, "elapsed_ms=") || strings.Contains(text, "secret") {
			t.Fatalf("invalid phase diagnostic: %s", text)
		}
	}
}
