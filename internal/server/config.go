package server

import (
	"fmt"
	"poke/internal/server/dispatch"
	"poke/internal/server/listener"
	"poke/internal/server/logging"

	"github.com/goccy/go-yaml"
)

// Config represents the top-level server configuration.
type Config struct {
	Commands  dispatch.CommandRegistry `yaml:"commands"`
	Listeners listener.ListenerConfig  `yaml:"listeners"`
	Logging   logging.Config           `yaml:"logging"`
}

type configInput struct {
	Commands  *dispatch.CommandRegistry `yaml:"commands"`
	Listeners *listener.ListenerConfig  `yaml:"listeners"`
	Logging   *logging.Config           `yaml:"logging"`
}

// Parse unmarshals raw config bytes into a Config.
func Parse(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// UnmarshalYAML composes command and listener block parsers per docs/configuration/server.md.
func (cfg *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if err := rejectLegacyTopLevelAuth(raw); err != nil {
		return err
	}

	in, err := decodeConfigInput(raw)
	if err != nil {
		return err
	}

	commands, err := parseCommandRegistryOrDefault(in.Commands)
	if err != nil {
		return err
	}

	listeners, err := parseListenerConfigOrDefault(in.Listeners)
	if err != nil {
		return err
	}

	logCfg, err := parseLoggingConfigOrDefault(in.Logging)
	if err != nil {
		return err
	}

	cfg.Commands = commands
	cfg.Listeners = listeners
	cfg.Logging = logCfg
	return nil
}

// rejectLegacyTopLevelAuth ensures deprecated top-level auth configuration is not used.
func rejectLegacyTopLevelAuth(raw map[string]interface{}) error {
	if _, hasLegacyAuth := raw["auth"]; hasLegacyAuth {
		return fmt.Errorf("top-level auth is no longer supported; configure auth under listeners.<type>.auth")
	}
	return nil
}

// decodeConfigInput decodes a raw YAML node into typed top-level server config blocks.
func decodeConfigInput(raw map[string]interface{}) (configInput, error) {
	data, err := yaml.Marshal(raw)
	if err != nil {
		return configInput{}, err
	}

	var in configInput
	if err := yaml.Unmarshal(data, &in); err != nil {
		return configInput{}, err
	}
	return in, nil
}

// parseCommandRegistryOrDefault returns parsed command registry or documented empty default.
func parseCommandRegistryOrDefault(input *dispatch.CommandRegistry) (dispatch.CommandRegistry, error) {
	if input != nil {
		return *input, nil
	}

	var empty dispatch.CommandRegistry
	if err := yaml.Unmarshal([]byte(`{}`), &empty); err != nil {
		return dispatch.CommandRegistry{}, err
	}
	return empty, nil
}

// parseListenerConfigOrDefault returns parsed listener config or documented empty default.
func parseListenerConfigOrDefault(input *listener.ListenerConfig) (listener.ListenerConfig, error) {
	if input != nil {
		return *input, nil
	}

	var empty listener.ListenerConfig
	if err := yaml.Unmarshal([]byte(`{}`), &empty); err != nil {
		return listener.ListenerConfig{}, err
	}
	return empty, nil
}

// parseLoggingConfigOrDefault returns parsed logging config or documented default values.
func parseLoggingConfigOrDefault(input *logging.Config) (logging.Config, error) {
	if input != nil {
		return *input, nil
	}

	var defaults logging.Config
	if err := yaml.Unmarshal([]byte(`{}`), &defaults); err != nil {
		return logging.Config{}, err
	}
	return defaults, nil
}
