package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"model-relay/internal/relay"
)

func main() {
	directory := flag.String("data-dir", "", "isolated schema2 compatibility installation")
	key := flag.String("key-file", "", "isolated material")
	output := flag.String("output", "", "new pre-migration archive")
	flag.Parse()
	result, err := relay.Upgrade(*directory, *key, *output)
	if err != nil {
		fmt.Fprintln(os.Stderr, "schema2 compatibility upgrade failed")
		os.Exit(1)
	}
	if json.NewEncoder(os.Stdout).Encode(result) != nil {
		os.Exit(1)
	}
}
