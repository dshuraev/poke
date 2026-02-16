package listener

import (
	"context"
	"fmt"
	"poke/internal/server/request"
	"sort"

	"github.com/goccy/go-yaml"
)

type RequestSource[T any] interface {
	Listen(ctx context.Context, cfg T, ch chan<- request.CommandRequest) error
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

// StartAll starts all configured listeners and returns the started instances.
func (lc ListenerConfig) StartAll(ctx context.Context, ch chan<- request.CommandRequest) ([]Listener, error) {
	if len(lc.listeners) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(lc.listeners))
	for listenerType := range lc.listeners {
		keys = append(keys, listenerType)
	}
	sort.Strings(keys)

	started := make([]Listener, 0, len(keys))
	for _, listenerType := range keys {
		entry := lc.listeners[listenerType]
		switch listenerType {
		case "http":
			httpListener, ok := entry.listener.(*HTTPListener)
			if !ok {
				return nil, fmt.Errorf("listener http: invalid listener type %T", entry.listener)
			}
			cfg, ok := entry.config.(HTTPListenerConfig)
			if !ok {
				return nil, fmt.Errorf("listener http: invalid config type %T", entry.config)
			}
			if err := httpListener.Listen(ctx, cfg, ch); err != nil {
				return nil, fmt.Errorf("listener http: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported listener type %q", listenerType)
		}
		started = append(started, entry)
	}

	return started, nil
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
