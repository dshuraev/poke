package listener_test

import (
	"testing"
	"time"

	"poke/internal/server/listener"

	"github.com/goccy/go-yaml"
)

func TestHTTPListenerConfigDefaults(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`
auth:
  api_token:
    token: "secret"
`), &cfg); err != nil {
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
	if cfg.Auth == nil {
		t.Fatalf("auth: expected configured auth block")
	}
}

func TestHTTPListenerConfigCustomValues(t *testing.T) {
	input := []byte(`
host: 0.0.0.0
port: 9000
read_timeout: 1s
write_timeout: 2s
idle_timeout: 3s
auth:
  api_token:
    token: "secret"
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
	input := []byte(`
host: ""
auth:
  api_token:
    token: "secret"
`)
	if err := yaml.Unmarshal(input, &cfg); err == nil {
		t.Fatalf("expected error for empty host")
	}
}

func TestHTTPListenerConfigRejectsPortOutOfRange(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`
port: 0
auth:
  api_token:
    token: "secret"
`), &cfg); err == nil {
		t.Fatalf("expected error for port below range")
	}
	if err := yaml.Unmarshal([]byte(`
port: 70000
auth:
  api_token:
    token: "secret"
`), &cfg); err == nil {
		t.Fatalf("expected error for port above range")
	}
}

func TestHTTPListenerConfigAcceptsPortBounds(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`
port: 1
auth:
  api_token:
    token: "secret"
`), &cfg); err != nil {
		t.Fatalf("unmarshal min: %v", err)
	}
	if cfg.Port != 1 {
		t.Fatalf("port min: got %d", cfg.Port)
	}

	if err := yaml.Unmarshal([]byte(`
port: 65535
auth:
  api_token:
    token: "secret"
`), &cfg); err != nil {
		t.Fatalf("unmarshal max: %v", err)
	}
	if cfg.Port != 65535 {
		t.Fatalf("port max: got %d", cfg.Port)
	}
}

func TestHTTPListenerConfigRejectsInvalidTimeout(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`
read_timeout: "n/a"
auth:
  api_token:
    token: "secret"
`), &cfg); err == nil {
		t.Fatalf("expected error for invalid timeout")
	}
}

func TestHTTPListenerConfigParsesAuthAPITokenFromEnv(t *testing.T) {
	const envName = "POKE_TEST_HTTP_LISTENER_AUTH_TOKEN"
	const envValue = "from-env"
	t.Setenv(envName, envValue)

	input := []byte(`
auth:
  api_token:
    env: POKE_TEST_HTTP_LISTENER_AUTH_TOKEN
`)

	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cfg.Auth == nil {
		t.Fatalf("auth: expected configured auth block")
	}
}

func TestHTTPListenerConfigRejectsMissingAuth(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{}`), &cfg); err == nil {
		t.Fatalf("expected error for missing auth")
	}
}

func TestHTTPListenerConfigRejectsEmptyAuth(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	if err := yaml.Unmarshal([]byte(`{auth: {}}`), &cfg); err == nil {
		t.Fatalf("expected error for empty auth")
	}
}

func TestHTTPListenerConfigRejectsUnknownAuthMethod(t *testing.T) {
	var cfg listener.HTTPListenerConfig
	input := []byte(`
auth:
  unknown: {}
`)
	if err := yaml.Unmarshal(input, &cfg); err == nil {
		t.Fatalf("expected error for unknown auth method")
	}
}
