package server_test

import (
	"testing"

	"poke/internal/server"
)

// TestConfigParsePopulatesCommands verifies Parse composes command and listener parsers.
func TestConfigParsePopulatesCommands(t *testing.T) {
	input := []byte(`
commands:
  uptime: uptime
listeners:
  http: {}
`)

	cfg, err := server.Parse(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	cmd, err := cfg.Commands.Get("uptime")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(cmd.Args) != 1 || cmd.Args[0] != "uptime" {
		t.Fatalf("args: got %#v", cmd.Args)
	}
	if cmd.ID != "uptime" {
		t.Fatalf("id: got %q", cmd.ID)
	}
}

// TestConfigParseAllowsEmptyConfig verifies Parse accepts an empty config.
func TestConfigParseAllowsEmptyConfig(t *testing.T) {
	if _, err := server.Parse([]byte(`{}`)); err != nil {
		t.Fatalf("parse: %v", err)
	}
}

// TestConfigParseRejectsInvalidListener verifies listener parser errors are propagated.
func TestConfigParseRejectsInvalidListener(t *testing.T) {
	input := []byte(`
listeners:
  bogus: {}
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for invalid listener type")
	}
}

// TestConfigParseRejectsInvalidCommand verifies command parser errors are propagated.
func TestConfigParseRejectsInvalidCommand(t *testing.T) {
	input := []byte(`
commands:
  bad:
    name: "nope"
listeners:
  http: {}
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for invalid command config")
	}
}
