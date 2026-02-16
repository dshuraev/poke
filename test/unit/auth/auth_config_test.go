package auth_test

import (
	"testing"

	"poke/internal/server/auth"

	"github.com/goccy/go-yaml"
)

func TestAuthUnmarshalAPITokenAndValidate(t *testing.T) {
	input := []byte(`
api_token:
  token: "secret"
`)

	var cfg auth.Auth
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("http", "secret")
	if err := cfg.Validate(&ctx); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestAuthValidateNoValidatorsConfigured(t *testing.T) {
	var cfg auth.Auth
	if err := yaml.Unmarshal([]byte(`{}`), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("http", "secret")
	if err := cfg.Validate(&ctx); err == nil {
		t.Fatalf("expected error when no validators configured")
	}
}

func TestAuthUnmarshalUnknownAuthMethod(t *testing.T) {
	var cfg auth.Auth
	if err := yaml.Unmarshal([]byte(`{unknown: {}}`), &cfg); err == nil {
		t.Fatalf("expected error for unknown auth method")
	}
}

func TestAuthUnmarshalAPITokenNullConfig(t *testing.T) {
	var cfg auth.Auth
	if err := yaml.Unmarshal([]byte(`{api_token: null}`), &cfg); err == nil {
		t.Fatalf("expected error for null api_token config")
	}
}
