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
  http:
    auth:
      api_token:
        token: "secret"
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

// TestConfigParseRejectsMissingListenerAuth verifies Parse rejects listeners without auth.
func TestConfigParseRejectsMissingListenerAuth(t *testing.T) {
	input := []byte(`
listeners:
  http: {}
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for missing listener auth block")
	}
}

// TestConfigParseRejectsLegacyTopLevelAuth verifies Parse rejects deprecated top-level auth.
func TestConfigParseRejectsLegacyTopLevelAuth(t *testing.T) {
	input := []byte(`
listeners:
  http:
    auth:
      api_token:
        token: "secret"
auth:
  api_token:
    token: "legacy-secret"
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for deprecated top-level auth block")
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
  http:
    auth:
      api_token:
        token: "secret"
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for invalid command config")
	}
}
