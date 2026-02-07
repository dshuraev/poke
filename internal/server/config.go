package server

import (
	"fmt"
	"poke/internal/server/auth"
	"poke/internal/server/dispatch"
	"poke/internal/server/listener"

	"github.com/goccy/go-yaml"
)

// Config represents the top-level server configuration.
type Config struct {
	Commands  dispatch.CommandRegistry `yaml:"commands"`
	Listeners listener.ListenerConfig  `yaml:"listeners"`
	Auth      auth.Auth                `yaml:"auth"`
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
	type configInput struct {
		Commands  *dispatch.CommandRegistry `yaml:"commands"`
		Listeners *listener.ListenerConfig  `yaml:"listeners"`
		Auth      *auth.Auth                `yaml:"auth"`
	}

	var in configInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Auth == nil {
		return fmt.Errorf("auth is required")
	}

	if in.Commands == nil {
		var empty dispatch.CommandRegistry
		if err := yaml.Unmarshal([]byte(`{}`), &empty); err != nil {
			return err
		}
		cfg.Commands = empty
	} else {
		cfg.Commands = *in.Commands
	}

	if in.Listeners == nil {
		var empty listener.ListenerConfig
		if err := yaml.Unmarshal([]byte(`{}`), &empty); err != nil {
			return err
		}
		cfg.Listeners = empty
	} else {
		cfg.Listeners = *in.Listeners
	}

	cfg.Auth = *in.Auth
	return nil
}
