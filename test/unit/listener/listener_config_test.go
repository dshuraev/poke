package listener_test

import (
	"testing"

	"poke/internal/server/listener"

	"github.com/goccy/go-yaml"
)

func TestListenerConfigUnmarshalHTTP(t *testing.T) {
	input := []byte(`
http:
  host: 127.0.0.1
  port: 9001
  auth:
    api_token:
      token: "secret"
`)

	var cfg listener.ListenerConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestListenerConfigUnmarshalEmpty(t *testing.T) {
	var cfg listener.ListenerConfig
	if err := yaml.Unmarshal([]byte(`{}`), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func TestListenerConfigUnknownType(t *testing.T) {
	var cfg listener.ListenerConfig
	if err := yaml.Unmarshal([]byte(`{udp: {}}`), &cfg); err == nil {
		t.Fatalf("expected error for unknown listener type")
	}
}
