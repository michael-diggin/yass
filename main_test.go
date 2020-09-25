package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {

	args := []string{"deploy", "8080"}
	command := "hello"
	var out bytes.Buffer
	var outErr bytes.Buffer
	err := run(args, command, &out, &outErr)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Fatalf("Got %s", out.String())
	}
	if !strings.Contains(outErr.String(), "") {
		t.Fatalf("%s", outErr.String())
	}
}

func TestRunWithBadCommand(t *testing.T) {
	args := []string{"hello"}
	command := "hello"
	var out bytes.Buffer
	var outErr bytes.Buffer
	err := run(args, command, &out, &outErr)
	if err.Error() != "Incorrect command: usage 'yass deploy PORT'" {
		t.Fatalf("Unexpected err: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("Got %s", out.String())
	}
	if outErr.String() != "" {
		t.Fatalf("Got %s", outErr.String())
	}
}
