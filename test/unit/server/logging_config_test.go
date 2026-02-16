package server_test

import (
	"reflect"
	"testing"

	"poke/internal/server"
	"poke/internal/server/logging"
)

// TestConfigParseAppliesLoggingDefaults verifies logging defaults are applied when omitted.
func TestConfigParseAppliesLoggingDefaults(t *testing.T) {
	input := []byte(`
commands:
  uptime: uptime
`)

	cfg, err := server.Parse(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if cfg.Logging.Level != "info" {
		t.Fatalf("logging level: got %q want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Format != "text" {
		t.Fatalf("logging format: got %q want %q", cfg.Logging.Format, "text")
	}
	if cfg.Logging.Sink.Type != "stdout" {
		t.Fatalf("logging sink type: got %q want %q", cfg.Logging.Sink.Type, "stdout")
	}
}

// TestConfigParsePopulatesJournaldLogging verifies journald sink config parsing.
func TestConfigParsePopulatesJournaldLogging(t *testing.T) {
	input := []byte(`
commands:
  uptime: uptime
logging:
  level: warn
  format: text
  add_source: true
  static_fields:
    service: poke
    env: prod
  sink:
    type: journald
    journald:
      identifier: poke-server
`)

	cfg, err := server.Parse(input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	want := logging.Config{
		Level:     "warn",
		Format:    "text",
		AddSource: true,
		StaticFields: map[string]string{
			"service": "poke",
			"env":     "prod",
		},
		Sink: logging.SinkConfig{
			Type: "journald",
			Journald: &logging.JournaldSinkConfig{
				Identifier: "poke-server",
				Fallback:   "stdout",
			},
		},
	}

	if !reflect.DeepEqual(cfg.Logging, want) {
		t.Fatalf("logging: got %#v want %#v", cfg.Logging, want)
	}
}

// TestConfigParseRejectsInvalidLoggingLevel verifies level validation.
func TestConfigParseRejectsInvalidLoggingLevel(t *testing.T) {
	input := []byte(`
commands:
  uptime: uptime
logging:
  level: noisy
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for invalid logging level")
	}
}

// TestConfigParseRejectsJournaldWithoutIdentifier verifies journald identifier validation.
func TestConfigParseRejectsJournaldWithoutIdentifier(t *testing.T) {
	input := []byte(`
commands:
  uptime: uptime
logging:
  sink:
    type: journald
    journald:
      fallback: stdout
`)

	if _, err := server.Parse(input); err == nil {
		t.Fatalf("expected error for missing journald identifier")
	}
}
