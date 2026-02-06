package auth_test

import (
	"os"
	"testing"

	"poke/internal/server/auth"

	"github.com/goccy/go-yaml"
)

func TestAPITokenConfigUnmarshalTokenAndValidate(t *testing.T) {
	input := []byte(`
listeners: [http]
token: "my-secret-token"
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.AuthContext{
		AuthKind:     auth.AuthTypeAPIToken,
		ListenerType: "http",
		APIToken:     "my-secret-token",
	}
	if err := cfg.Validate(&ctx); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestAPITokenConfigUnmarshalEnvAndValidate(t *testing.T) {
	const envName = "POKE_TEST_API_TOKEN"
	const envValue = "from-env"

	t.Setenv(envName, envValue)

	input := []byte(`
env: POKE_TEST_API_TOKEN
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("http", envValue)
	if err := cfg.Validate(&ctx); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestAPITokenConfigUnmarshalEnvMissing(t *testing.T) {
	_ = os.Unsetenv("POKE_TEST_API_TOKEN_MISSING")

	input := []byte(`
env: POKE_TEST_API_TOKEN_MISSING
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err == nil {
		t.Fatalf("expected error for missing env var")
	}
}

func TestAPITokenConfigUnmarshalFileAndValidate(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "poke-api-token-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString("from-file\n"); err != nil {
		_ = f.Close()
		t.Fatalf("write: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	input := []byte(`
file: "` + f.Name() + `"
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("http", "from-file")
	if err := cfg.Validate(&ctx); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestAPITokenConfigUnmarshalMultipleSources(t *testing.T) {
	input := []byte(`
token: a
env: SOME_ENV
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err == nil {
		t.Fatalf("expected error for multiple token sources")
	}
}

func TestAPITokenConfigValidateListenerNotAllowed(t *testing.T) {
	input := []byte(`
listeners: [http]
token: "x"
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("udp", "x")
	if err := cfg.Validate(&ctx); err == nil {
		t.Fatalf("expected error for disallowed listener type")
	}
}

func TestAPITokenConfigValidateInvalidToken(t *testing.T) {
	input := []byte(`
token: "x"
`)

	var cfg auth.APITokenConfig
	if err := yaml.Unmarshal(input, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	ctx := auth.NewAPITokenContext("http", "nope")
	if err := cfg.Validate(&ctx); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
