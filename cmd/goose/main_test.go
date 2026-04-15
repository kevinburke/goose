package main

import (
	"bytes"
	"testing"

	"github.com/kevinburke/goose/lib/goose"
)

func TestPrintVersion(t *testing.T) {
	var buf bytes.Buffer
	printVersion(&buf)
	if got, want := buf.String(), goose.Version+"\n"; got != want {
		t.Fatalf("printVersion() = %q, want %q", got, want)
	}
}
