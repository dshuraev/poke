package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"
)

// New builds a structured logger from config.
func New(cfg Config) (*slog.Logger, error) {
	normalized := withRuntimeDefaults(cfg)
	if err := normalized.validate(); err != nil {
		return nil, err
	}

	level, err := parseLevel(normalized.Level)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		AddSource: normalized.AddSource,
		Level:     level,
	}

	handler := newHandler(normalized, opts)
	logger := slog.New(handler)

	staticAttrs := staticFieldAttrs(normalized.StaticFields)
	if len(staticAttrs) > 0 {
		logger = logger.With(staticAttrs...)
	}

	return logger, nil
}

// withRuntimeDefaults applies runtime defaults for direct logger construction.
func withRuntimeDefaults(cfg Config) Config {
	out := cfg

	if strings.TrimSpace(out.Level) == "" {
		out.Level = defaultLevel
	} else {
		out.Level = normalizeToken(out.Level)
	}

	if strings.TrimSpace(out.Format) == "" {
		out.Format = defaultFormat
	} else {
		out.Format = normalizeToken(out.Format)
	}

	if strings.TrimSpace(out.Sink.Type) == "" {
		out.Sink.Type = defaultSinkType
	} else {
		out.Sink.Type = normalizeToken(out.Sink.Type)
	}

	if out.Sink.Type == "journald" {
		if out.Sink.Journald == nil {
			out.Sink.Journald = &JournaldSinkConfig{
				Fallback: defaultJournaldFallbk,
			}
		}
		if strings.TrimSpace(out.Sink.Journald.Fallback) == "" {
			out.Sink.Journald.Fallback = defaultJournaldFallbk
		} else {
			out.Sink.Journald.Fallback = normalizeToken(out.Sink.Journald.Fallback)
		}
	}

	return out
}

// parseLevel maps a config level token to slog level.
func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported logging level %q", level)
	}
}

// newHandler builds a handler for the configured format and sink.
func newHandler(cfg Config, opts *slog.HandlerOptions) slog.Handler {
	fallback := newOutputHandler(cfg.Format, resolveFallbackOutput(cfg.Sink), opts)

	if cfg.Sink.Type == "journald" && cfg.Sink.Journald != nil {
		handler, err := newJournaldHandler(cfg.Sink.Journald.Identifier, opts, fallback)
		if err == nil {
			return handler
		}
	}

	return fallback
}

// newOutputHandler builds a text or JSON handler for a plain output writer.
func newOutputHandler(format string, output io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if format == "text" {
		return slog.NewTextHandler(output, opts)
	}
	return slog.NewJSONHandler(output, opts)
}

// resolveFallbackOutput resolves the configured fallback output writer.
func resolveFallbackOutput(sink SinkConfig) *os.File {
	switch sink.Type {
	case "stdout", "journald":
		return os.Stdout
	default:
		return os.Stdout
	}
}

// staticFieldAttrs converts static_fields into deterministic slog attrs.
func staticFieldAttrs(fields map[string]string) []any {
	if len(fields) == 0 {
		return nil
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	attrs := make([]any, 0, len(keys)*2)
	for _, key := range keys {
		attrs = append(attrs, key, fields[key])
	}
	return attrs
}
