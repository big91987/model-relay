package jsoncodec_test

import (
	"testing"

	"model-relay/internal/jsoncodec"
)

func TestUniqueRejectsAmbiguousJSON(t *testing.T) {
	for _, data := range [][]byte{
		[]byte(`{"name":"first","name":"second"}`),
		[]byte(`{"nested":[{"a":1,"\u0061":2}]}`),
		[]byte(`{} {}`), []byte(`{"a":`), []byte(``),
		{'{', '"', 'a', '"', ':', '"', 0xff, '"', '}'},
	} {
		if jsoncodec.Unique(data) {
			t.Fatalf("ambiguous JSON accepted: %q", data)
		}
	}
	for _, data := range []string{`{"name":"first","nested":[{"name":"second"}]}`, `[]`, `null`, `"text"`, `123`} {
		if !jsoncodec.Unique([]byte(data)) {
			t.Fatalf("valid JSON rejected: %s", data)
		}
	}
}

func TestObjectRequiresNonNullObject(t *testing.T) {
	for _, data := range []string{`null`, `[]`, `"text"`, `{`} {
		if _, ok := jsoncodec.Object([]byte(data)); ok {
			t.Fatalf("non-object accepted: %s", data)
		}
	}
	value, ok := jsoncodec.Object([]byte(`{"name":"first"}`))
	if !ok || string(value["name"]) != `"first"` {
		t.Fatal("object field lost")
	}
}
