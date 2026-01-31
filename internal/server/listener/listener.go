package listener

import (
	"context"
	"fmt"
	"poke/internal/server/request"

	"github.com/goccy/go-yaml"
)

type RequestSource[T any] interface {
	Listen(ctx context.Context, cfg T, ch chan<- request.CommandRequest)
}

type Listener struct {
	listener interface{}
	config   interface{}
}

type ListenerConfig struct {
	listeners map[string]Listener
}

// UnmarshalYAML parses listener config per docs/configuration/listener.md.
func (lc *ListenerConfig) UnmarshalYAML(unmarshall func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshall(&raw); err != nil {
		return err
	}

	if raw == nil {
		lc.listeners = map[string]Listener{}
		return nil
	}

	listeners := make(map[string]Listener, len(raw))
	for listenerType, rawConfig := range raw {
		switch listenerType {
		case "http":
			var cfg HTTPListenerConfig
			if err := decodeListenerConfig(rawConfig, &cfg); err != nil {
				return fmt.Errorf("listener http: %w", err)
			}

			listeners[listenerType] = Listener{
				listener: &HTTPListener{},
				config:   cfg,
			}
		default:
			return fmt.Errorf("unsupported listener type %q", listenerType)
		}
	}

	lc.listeners = listeners
	return nil
}

// decodeListenerConfig unmarshals a per-listener config node into a target struct.
func decodeListenerConfig(rawConfig interface{}, target interface{}) error {
	if rawConfig == nil {
		return yaml.Unmarshal([]byte(`{}`), target)
	}

	data, err := yaml.Marshal(rawConfig)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, target)
}
