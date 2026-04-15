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

func TestPrintUnknownCommand(t *testing.T) {
	var buf bytes.Buffer
	printUnknownCommand(&buf, "wat")
	if got, want := buf.String(), "error: unknown command \"wat\"\n"; got != want {
		t.Fatalf("printUnknownCommand() = %q, want %q", got, want)
	}
}
