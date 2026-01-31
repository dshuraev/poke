package listener_test

import (
	"testing"
	"time"

	"poke/internal/server/listener"

	"github.com/goccy/go-yaml"
)

func TestHTTPListenerConfigDefaults(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{}`), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("host: got %q", cfg.Host)
	}
	if cfg.Port != 8008 {
		t.Fatalf("port: got %d", cfg.Port)
	}
	if cfg.ReadTimeout != 0 {
		t.Fatalf("read_timeout: got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 0 {
		t.Fatalf("write_timeout: got %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 0 {
		t.Fatalf("idle_timeout: got %v", cfg.IdleTimeout)
	}
}

func TestHTTPListenerConfigCustomValues(t *testing.T) {
	input := []byte(`
host: 0.0.0.0
port: 9000
read_timeout: 1s
write_timeout: 2s
idle_timeout: 3s
`)

	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Fatalf("host: got %q", cfg.Host)
	}
	if cfg.Port != 9000 {
		t.Fatalf("port: got %d", cfg.Port)
	}
	if cfg.ReadTimeout != time.Second {
		t.Fatalf("read_timeout: got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 2*time.Second {
		t.Fatalf("write_timeout: got %v", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 3*time.Second {
		t.Fatalf("idle_timeout: got %v", cfg.IdleTimeout)
	}
}

func TestHTTPListenerConfigRejectsEmptyHost(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{host: ""}`), &cfg); err == nil {
		t.Fatalf("expected error for empty host")
	}
}

func TestHTTPListenerConfigRejectsPortOutOfRange(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{port: 0}`), &cfg); err == nil {
		t.Fatalf("expected error for port below range")
	}
	if err := yaml.Unmarshal([]byte(`{port: 70000}`), &cfg); err == nil {
		t.Fatalf("expected error for port above range")
	}
}

func TestHTTPListenerConfigAcceptsPortBounds(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{port: 1}`), &cfg); err != nil {
		t.Fatalf("unmarshal min: %v", err)
	}
	if cfg.Port != 1 {
		t.Fatalf("port min: got %d", cfg.Port)
	}

	if err := yaml.Unmarshal([]byte(`{port: 65535}`), &cfg); err != nil {
		t.Fatalf("unmarshal max: %v", err)
	}
	if cfg.Port != 65535 {
		t.Fatalf("port max: got %d", cfg.Port)
	}
}

func TestHTTPListenerConfigRejectsInvalidTimeout(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{read_timeout: "n/a"}`), &cfg); err == nil {
		t.Fatalf("expected error for invalid timeout")
	}
}
