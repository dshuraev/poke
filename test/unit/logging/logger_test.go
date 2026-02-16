package logging_test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"poke/internal/server/logging"
)

// TestNewWritesToStdoutSink verifies stdout sink emits configured text logs.
func TestNewWritesToStdoutSink(t *testing.T) {
	logs := captureStdout(t, func() error {
		logger, err := logging.New(logging.Config{
			Level:  "info",
			Format: "text",
			StaticFields: map[string]string{
				"service": "poke",
			},
			Sink: logging.SinkConfig{Type: "stdout"},
		})
		if err != nil {
			return err
		}

		logger.Info("hello", "event", "stdout_sink")
		return nil
	})

	if !strings.Contains(logs, "hello") {
		t.Fatalf("expected message in logs, got %q", logs)
	}
	if !strings.Contains(logs, "event=stdout_sink") {
		t.Fatalf("expected event in logs, got %q", logs)
	}
	if !strings.Contains(logs, "service=poke") {
		t.Fatalf("expected static field in logs, got %q", logs)
	}
}

// TestNewFallsBackWhenJournaldSocketUnavailable verifies journald sink fallback behavior.
func TestNewFallsBackWhenJournaldSocketUnavailable(t *testing.T) {
	// Force journald socket dial failure deterministically.
	t.Setenv("POKE_JOURNALD_SOCKET", filepath.Join(t.TempDir(), "missing-journald.sock"))

	logs := captureStdout(t, func() error {
		logger, err := logging.New(logging.Config{
			Level:  "info",
			Format: "text",
			Sink: logging.SinkConfig{
				Type: "journald",
				Journald: &logging.JournaldSinkConfig{
					Identifier: "poke-test",
					Fallback:   "stdout",
				},
			},
		})
		if err != nil {
			return err
		}

		logger.Info("hello journald", "event", "journald_fallback")
		return nil
	})

	if !strings.Contains(logs, "hello journald") {
		t.Fatalf("expected message in fallback logs, got %q", logs)
	}
	if !strings.Contains(logs, "event=journald_fallback") {
		t.Fatalf("expected event in fallback logs, got %q", logs)
	}
}

// captureStdout captures stdout during run and returns written output.
func captureStdout(t *testing.T, run func() error) string {
	t.Helper()

	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = oldStdout
	}()

	if err := run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}

	return string(data)
}
