package logging

import (
	"fmt"
	"strings"
)

const (
	defaultLevel          = "info"
	defaultFormat         = "text"
	defaultSinkType       = "stdout"
	defaultJournaldFallbk = "stdout"
)

var (
	allowedLevels = map[string]struct{}{
		"debug": {},
		"info":  {},
		"warn":  {},
		"error": {},
	}
	allowedFormats = map[string]struct{}{
		"json": {},
		"text": {},
	}
	allowedSinkTypes = map[string]struct{}{
		"stdout":   {},
		"journald": {},
	}
	allowedJournaldFallbacks = map[string]struct{}{
		"stdout": {},
	}
)

// Config defines server logging settings from docs/configuration/logging.md.
type Config struct {
	Level        string            `yaml:"level,omitempty"`
	Format       string            `yaml:"format,omitempty"`
	AddSource    bool              `yaml:"add_source,omitempty"`
	StaticFields map[string]string `yaml:"static_fields,omitempty"`
	Sink         SinkConfig        `yaml:"sink,omitempty"`
}

// SinkConfig defines the active log sink and sink-specific options.
type SinkConfig struct {
	Type     string              `yaml:"type,omitempty"`
	Journald *JournaldSinkConfig `yaml:"journald,omitempty"`
}

// JournaldSinkConfig defines journald-only sink options.
type JournaldSinkConfig struct {
	Identifier string `yaml:"identifier,omitempty"`
	Fallback   string `yaml:"fallback,omitempty"`
}

// UnmarshalYAML parses and validates logging config, applying documented defaults.
func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type configInput struct {
		Level        *string           `yaml:"level"`
		Format       *string           `yaml:"format"`
		AddSource    *bool             `yaml:"add_source"`
		StaticFields map[string]string `yaml:"static_fields"`
		Sink         *SinkConfig       `yaml:"sink"`
	}

	*cfg = Config{
		Level:  defaultLevel,
		Format: defaultFormat,
		Sink: SinkConfig{
			Type: defaultSinkType,
		},
	}

	var in configInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Level != nil {
		cfg.Level = normalizeToken(*in.Level)
	}
	if in.Format != nil {
		cfg.Format = normalizeToken(*in.Format)
	}
	if in.AddSource != nil {
		cfg.AddSource = *in.AddSource
	}
	if in.StaticFields != nil {
		cfg.StaticFields = in.StaticFields
	}
	if in.Sink != nil {
		cfg.Sink = *in.Sink
	}

	return cfg.validate()
}

// UnmarshalYAML parses sink options and applies sink-level defaults.
func (cfg *SinkConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type sinkInput struct {
		Type     *string             `yaml:"type"`
		Journald *JournaldSinkConfig `yaml:"journald"`
	}

	*cfg = SinkConfig{
		Type: defaultSinkType,
	}

	var in sinkInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Type != nil {
		cfg.Type = normalizeToken(*in.Type)
	}
	if in.Journald != nil {
		cfg.Journald = in.Journald
	}

	return cfg.validate()
}

// UnmarshalYAML parses journald sink options and applies journald defaults.
func (cfg *JournaldSinkConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type journaldInput struct {
		Identifier *string `yaml:"identifier"`
		Fallback   *string `yaml:"fallback"`
	}

	*cfg = JournaldSinkConfig{
		Fallback: defaultJournaldFallbk,
	}

	var in journaldInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Identifier != nil {
		cfg.Identifier = strings.TrimSpace(*in.Identifier)
	}
	if in.Fallback != nil {
		cfg.Fallback = normalizeToken(*in.Fallback)
	}

	return nil
}

func (cfg Config) validate() error {
	if _, ok := allowedLevels[cfg.Level]; !ok {
		return fmt.Errorf("logging level must be one of debug, info, warn, error")
	}
	if _, ok := allowedFormats[cfg.Format]; !ok {
		return fmt.Errorf("logging format must be one of json or text")
	}
	if err := cfg.Sink.validate(); err != nil {
		return err
	}
	return nil
}

func (cfg SinkConfig) validate() error {
	if _, ok := allowedSinkTypes[cfg.Type]; !ok {
		return fmt.Errorf("logging sink type must be one of stdout or journald")
	}
	if cfg.Type != "journald" {
		return nil
	}
	if cfg.Journald == nil {
		return fmt.Errorf("logging sink journald config is required when sink type is journald")
	}
	identifier := strings.TrimSpace(cfg.Journald.Identifier)
	if identifier == "" {
		return fmt.Errorf("logging sink journald identifier is required")
	}
	if _, ok := allowedJournaldFallbacks[cfg.Journald.Fallback]; !ok {
		return fmt.Errorf("logging sink journald fallback must be stdout")
	}
	return nil
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
