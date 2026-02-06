package auth

import (
	"crypto/subtle"
	"fmt"
	"os"
	"strings"
)

// APITokenConfig configures API token authentication.
//
// Exactly one token source must be configured:
// - token: literal token in config
// - env: environment variable containing the token
// - file: file path containing the token
//
// Listeners, when non-empty, restricts which listener types accept this auth kind.
type APITokenConfig struct {
	Listeners []string

	token string
	env   string
	file  string
}

type apiTokenSourceKind int

const (
	apiTokenSourceKindUnknown apiTokenSourceKind = iota
	apiTokenSourceKindToken
	apiTokenSourceKindEnv
	apiTokenSourceKindFile
)

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
		Listeners []string `yaml:"listeners"`
		Token     *string  `yaml:"token"`
		Env       *string  `yaml:"env"`
		File      *string  `yaml:"file"`
	}

	*cfg = APITokenConfig{}

	var in apiTokenConfigInput
	if err := unmarshal(&in); err != nil {
		return err
	}

	if err := cfg.setListeners(in.Listeners); err != nil {
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

// setListeners normalizes listener types, validates entries, and de-duplicates values.
func (cfg *APITokenConfig) setListeners(raw []string) error {
	if len(raw) == 0 {
		cfg.Listeners = nil
		return nil
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		normalized := normalizeListenerType(v)
		if normalized == "" {
			return fmt.Errorf("api_token listeners must not contain empty entries")
		}
		if _, exists := seen[normalized]; exists {
			return fmt.Errorf("api_token listeners contains duplicate %q", normalized)
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	cfg.Listeners = out
	return nil
}

// normalizeListenerType lowercases and trims a listener type for comparison.
func normalizeListenerType(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// allowsListenerType reports whether the config accepts the given listener type.
func (cfg *APITokenConfig) allowsListenerType(listenerType string) bool {
	if len(cfg.Listeners) == 0 {
		return true
	}

	normalized := normalizeListenerType(listenerType)
	for _, allowed := range cfg.Listeners {
		if allowed == normalized {
			return true
		}
	}
	return false
}

// Validate checks ctx against the configured API token and listener allow-list.
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
	if !cfg.allowsListenerType(ctx.ListenerType) {
		return fmt.Errorf("api_token is not enabled for listener %q", ctx.ListenerType)
	}

	if subtle.ConstantTimeCompare([]byte(cfg.token), []byte(ctx.APIToken)) != 1 {
		return fmt.Errorf("invalid api token")
	}

	return nil
}
