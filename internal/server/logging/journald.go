package logging

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const journaldSocketPath = "/run/systemd/journal/socket"
const journaldSocketEnvVar = "POKE_JOURNALD_SOCKET"

// journaldHandler writes slog records to systemd-journald.
type journaldHandler struct {
	identifier string
	opts       *slog.HandlerOptions
	writer     *net.UnixConn
	fallback   slog.Handler
	attrs      []slog.Attr
	groups     []string
}

// newJournaldHandler creates a native journald handler.
func newJournaldHandler(identifier string, opts *slog.HandlerOptions, fallback slog.Handler) (*journaldHandler, error) {
	writer, err := connectJournaldSocket(resolveJournaldSocketPath())
	if err != nil {
		return nil, err
	}

	return &journaldHandler{
		identifier: identifier,
		opts:       opts,
		writer:     writer,
		fallback:   fallback,
	}, nil
}

// resolveJournaldSocketPath returns journald socket path, optionally overridden for tests.
func resolveJournaldSocketPath() string {
	if value, exists := os.LookupEnv(journaldSocketEnvVar); exists {
		path := strings.TrimSpace(value)
		if path != "" {
			return path
		}
	}
	return journaldSocketPath
}

// Enabled checks whether the configured level allows the record.
func (h *journaldHandler) Enabled(_ context.Context, level slog.Level) bool {
	leveler := h.opts.Level
	if leveler == nil {
		return level >= slog.LevelInfo
	}
	return level >= leveler.Level()
}

// Handle emits one slog record to journald, falling back when delivery fails.
func (h *journaldHandler) Handle(ctx context.Context, rec slog.Record) error {
	fields := map[string]string{
		"MESSAGE":           rec.Message,
		"PRIORITY":          journaldPriority(rec.Level),
		"SYSLOG_IDENTIFIER": h.identifier,
		"POKE_LEVEL":        rec.Level.String(),
	}

	if !rec.Time.IsZero() {
		fields["POKE_TIME"] = rec.Time.Format("2006-01-02T15:04:05.999999999Z07:00")
	}

	if h.opts.AddSource && rec.PC != 0 {
		addSourceFields(fields, rec.PC)
	}

	for _, attr := range h.attrs {
		appendAttr(fields, h.opts, h.groups, attr)
	}
	rec.Attrs(func(attr slog.Attr) bool {
		appendAttr(fields, h.opts, h.groups, attr)
		return true
	})

	payload := encodeJournaldEntry(fields)
	if _, err := h.writer.Write(payload); err != nil {
		if h.fallback != nil {
			return h.fallback.Handle(ctx, rec)
		}
		return err
	}

	return nil
}

// WithAttrs returns a handler with additional attributes.
func (h *journaldHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nextAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	nextAttrs = append(nextAttrs, h.attrs...)
	nextAttrs = append(nextAttrs, attrs...)

	nextFallback := h.fallback
	if nextFallback != nil {
		nextFallback = nextFallback.WithAttrs(attrs)
	}

	return &journaldHandler{
		identifier: h.identifier,
		opts:       h.opts,
		writer:     h.writer,
		fallback:   nextFallback,
		attrs:      nextAttrs,
		groups:     h.groups,
	}
}

// WithGroup returns a handler that nests future attributes under group.
func (h *journaldHandler) WithGroup(name string) slog.Handler {
	if strings.TrimSpace(name) == "" {
		return h
	}

	nextGroups := make([]string, 0, len(h.groups)+1)
	nextGroups = append(nextGroups, h.groups...)
	nextGroups = append(nextGroups, name)

	nextFallback := h.fallback
	if nextFallback != nil {
		nextFallback = nextFallback.WithGroup(name)
	}

	return &journaldHandler{
		identifier: h.identifier,
		opts:       h.opts,
		writer:     h.writer,
		fallback:   nextFallback,
		attrs:      h.attrs,
		groups:     nextGroups,
	}
}

// connectJournaldSocket opens a unixgram connection to journald.
func connectJournaldSocket(path string) (*net.UnixConn, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	addr := &net.UnixAddr{
		Name: path,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// addSourceFields appends source metadata from program counter.
func addSourceFields(fields map[string]string, pc uintptr) {
	frame, ok := runtime.CallersFrames([]uintptr{pc}).Next()
	if !ok {
		return
	}

	fields["CODE_FILE"] = frame.File
	fields["CODE_LINE"] = strconv.Itoa(frame.Line)
	fields["CODE_FUNC"] = frame.Function
}

// appendAttr flattens an attr into journald fields.
func appendAttr(fields map[string]string, opts *slog.HandlerOptions, groups []string, attr slog.Attr) {
	if opts.ReplaceAttr != nil {
		attr = opts.ReplaceAttr(groups, attr)
	}

	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	if attr.Value.Kind() == slog.KindGroup {
		nextGroups := groups
		if attr.Key != "" {
			nextGroups = appendCopy(groups, attr.Key)
		}

		for _, child := range attr.Value.Group() {
			appendAttr(fields, opts, nextGroups, child)
		}
		return
	}

	if attr.Key == "" {
		return
	}

	fieldKey := buildCustomFieldKey(groups, attr.Key)
	fields[fieldKey] = attrValueString(attr.Value)
}

// appendCopy returns a new slice containing src values and tail value.
func appendCopy(src []string, tail string) []string {
	out := make([]string, 0, len(src)+1)
	out = append(out, src...)
	out = append(out, tail)
	return out
}

// journaldPriority maps slog levels to journald priority values.
func journaldPriority(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "3"
	case level >= slog.LevelWarn:
		return "4"
	case level >= slog.LevelInfo:
		return "6"
	default:
		return "7"
	}
}

// buildCustomFieldKey converts grouped attr keys into valid journald field names.
func buildCustomFieldKey(groups []string, key string) string {
	parts := make([]string, 0, len(groups)+1)
	for _, group := range groups {
		if strings.TrimSpace(group) != "" {
			parts = append(parts, group)
		}
	}
	parts = append(parts, key)

	name := strings.Join(parts, "_")
	name = sanitizeJournaldFieldName(name)

	if !strings.HasPrefix(name, "POKE_") {
		name = "POKE_" + name
	}
	return name
}

// sanitizeJournaldFieldName normalizes a field name to journald-compatible format.
func sanitizeJournaldFieldName(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "FIELD"
	}

	out := sanitizeFieldToken(strings.ToUpper(trimmed))
	if out == "" {
		return "FIELD"
	}
	return ensureValidFieldPrefix(out)
}

// sanitizeFieldToken converts unsupported characters to underscores.
func sanitizeFieldToken(token string) string {
	var b strings.Builder
	b.Grow(len(token))
	for _, ch := range token {
		if isJournaldFieldChar(ch) {
			b.WriteRune(ch)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}

// isJournaldFieldChar reports whether rune is valid in journald field names.
func isJournaldFieldChar(ch rune) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// ensureValidFieldPrefix rewrites invalid first field characters.
func ensureValidFieldPrefix(field string) string {
	first := field[0]
	if first == '_' || (first >= '0' && first <= '9') {
		return "F_" + field
	}
	return field
}

// attrValueString renders a slog value for journald transport.
func attrValueString(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format("2006-01-02T15:04:05.999999999Z07:00")
	case slog.KindAny:
		return fmt.Sprint(value.Any())
	default:
		return value.String()
	}
}

// encodeJournaldEntry encodes journald fields into the native datagram format.
func encodeJournaldEntry(fields map[string]string) []byte {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, key := range keys {
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(strings.ReplaceAll(fields[key], "\n", "\\n"))
		b.WriteByte('\n')
	}
	return []byte(b.String())
}
