package auth

import (
	"crypto/subtle"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

// APITokenConfig configures API token authentication.
//
// Exactly one token source must be configured:
// - token: literal token in config
// - env: environment variable containing the token
// - file: file path containing the token
type APITokenConfig struct {
	token string
	env   string
	file  string
}

// apiTokenSourceKind identifies which credential source was configured.
type apiTokenSourceKind int

const (
	// apiTokenSourceKindUnknown is a sentinel for invalid/unset source selection.
	apiTokenSourceKindUnknown apiTokenSourceKind = iota
	// apiTokenSourceKindToken indicates the literal `token:` field was configured.
	apiTokenSourceKindToken
	// apiTokenSourceKindEnv indicates the `env:` field was configured.
	apiTokenSourceKindEnv
	// apiTokenSourceKindFile indicates the `file:` field was configured.
	apiTokenSourceKindFile
)

// apiTokenSource captures the resolved token plus its provenance for diagnostics.
type apiTokenSource struct {
	kind     apiTokenSourceKind
	token    string
	envName  string
	filePath string
}

// UnmarshalYAML parses an api_token config block and resolves the effective token.
//
// Tokens are trimmed with strings.TrimSpace to avoid accidental whitespace from YAML
// indentation, env var values, or trailing newlines in files.
func (cfg *APITokenConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type apiTokenConfigInput struct {
		Token *string `yaml:"token"`
		Env   *string `yaml:"env"`
		File  *string `yaml:"file"`
	}

	*cfg = APITokenConfig{}

	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	if _, hasLegacyListeners := raw["listeners"]; hasLegacyListeners {
		return fmt.Errorf("api_token listeners is no longer supported; configure auth under listeners.<type>.auth")
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}

	var in apiTokenConfigInput
	if err := yaml.Unmarshal(data, &in); err != nil {
		return err
	}

	src, err := resolveAPITokenSource(in.Token, in.Env, in.File)
	if err != nil {
		return err
	}

	cfg.token = src.token
	cfg.env = src.envName
	cfg.file = src.filePath
	return nil
}

// Validate checks ctx against the configured API token.
func (cfg *APITokenConfig) Validate(ctx *AuthContext) error {
	if ctx == nil {
		return fmt.Errorf("auth context is required")
	}
	if ctx.AuthKind != AuthTypeAPIToken {
		return fmt.Errorf("auth method mismatch: expected %q got %q", AuthTypeAPIToken, ctx.AuthKind)
	}
	if cfg.token == "" {
		return fmt.Errorf("api_token is not configured")
	}

	if subtle.ConstantTimeCompare([]byte(cfg.token), []byte(ctx.APIToken)) != 1 {
		return fmt.Errorf("invalid api token")
	}

	return nil
}

// resolveAPITokenSource selects and resolves the configured token source.
func resolveAPITokenSource(rawToken *string, rawEnv *string, rawFile *string) (apiTokenSource, error) {
	kind, err := selectAPITokenSourceKind(rawToken, rawEnv, rawFile)
	if err != nil {
		return apiTokenSource{}, err
	}

	switch kind {
	case apiTokenSourceKindToken:
		token, err := resolveAPITokenFromLiteral(rawToken)
		if err != nil {
			return apiTokenSource{}, err
		}
		return apiTokenSource{kind: kind, token: token}, nil
	case apiTokenSourceKindEnv:
		token, envName, err := resolveAPITokenFromEnv(rawEnv)
		if err != nil {
			return apiTokenSource{}, err
		}
		return apiTokenSource{kind: kind, token: token, envName: envName}, nil
	case apiTokenSourceKindFile:
		token, filePath, err := resolveAPITokenFromFile(rawFile)
		if err != nil {
			return apiTokenSource{}, err
		}
		return apiTokenSource{kind: kind, token: token, filePath: filePath}, nil
	default:
		return apiTokenSource{}, fmt.Errorf("api_token: unknown token source")
	}
}

// selectAPITokenSourceKind enforces that exactly one of token, env, or file is set.
func selectAPITokenSourceKind(rawToken *string, rawEnv *string, rawFile *string) (apiTokenSourceKind, error) {
	tokenSet := rawToken != nil
	envSet := rawEnv != nil
	fileSet := rawFile != nil

	tokenSources := 0
	if tokenSet {
		tokenSources++
	}
	if envSet {
		tokenSources++
	}
	if fileSet {
		tokenSources++
	}

	if tokenSources == 0 {
		return apiTokenSourceKindUnknown, fmt.Errorf("api_token requires exactly one of token, env, or file")
	}
	if tokenSources != 1 {
		return apiTokenSourceKindUnknown, fmt.Errorf("api_token must not combine token, env, and/or file")
	}

	if tokenSet {
		return apiTokenSourceKindToken, nil
	}
	if envSet {
		return apiTokenSourceKindEnv, nil
	}
	return apiTokenSourceKindFile, nil
}

// resolveAPITokenFromLiteral validates and normalizes the literal token source.
func resolveAPITokenFromLiteral(raw *string) (string, error) {
	if raw == nil {
		return "", fmt.Errorf("api_token token is required")
	}
	token := strings.TrimSpace(*raw)
	if token == "" {
		return "", fmt.Errorf("api_token token must not be empty")
	}
	return token, nil
}

// resolveAPITokenFromEnv loads and validates the token from an environment variable.
func resolveAPITokenFromEnv(raw *string) (token string, envName string, err error) {
	if raw == nil {
		return "", "", fmt.Errorf("api_token env is required")
	}

	envName = strings.TrimSpace(*raw)
	if envName == "" {
		return "", "", fmt.Errorf("api_token env must not be empty")
	}
	value, ok := os.LookupEnv(envName)
	if !ok {
		return "", "", fmt.Errorf("api_token env %q is not set", envName)
	}

	token = strings.TrimSpace(value)
	if token == "" {
		return "", "", fmt.Errorf("api_token env %q is empty", envName)
	}

	return token, envName, nil
}

// resolveAPITokenFromFile loads and validates the token from a file on disk.
func resolveAPITokenFromFile(raw *string) (token string, filePath string, err error) {
	if raw == nil {
		return "", "", fmt.Errorf("api_token file is required")
	}

	filePath = strings.TrimSpace(*raw)
	if filePath == "" {
		return "", "", fmt.Errorf("api_token file must not be empty")
	}

	data, err := os.ReadFile(filePath) // #nosec G304 -- by design, comes from config
	if err != nil {
		return "", "", fmt.Errorf("api_token file: %w", err)
	}

	token = strings.TrimSpace(string(data))
	if token == "" {
		return "", "", fmt.Errorf("api_token file %q is empty", filePath)
	}

	return token, filePath, nil
}
