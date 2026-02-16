package auth

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// Auth configures and dispatches authentication validators by auth kind.
//
// Configuration is loaded from each listener's `auth` node (see docs/configuration/auth.md).
type Auth struct {
	Validators map[string]Validator
}

// Validator checks a request-scoped AuthContext against a configured credential.
type Validator interface {
	Validate(ctx *AuthContext) error
}

// UnmarshalYAML parses the `auth` block and instantiates validators by auth kind.
func (auth *Auth) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*auth = Auth{Validators: map[string]Validator{}}

	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	if raw == nil {
		return nil
	}

	validators := make(map[string]Validator, len(raw))
	for authKind, rawConfig := range raw {
		switch authKind {
		case AuthTypeAPIToken:
			cfg := new(APITokenConfig)
			if err := decodeAuthConfig(rawConfig, cfg); err != nil {
				return fmt.Errorf("auth %s: %w", authKind, err)
			}
			validators[authKind] = cfg
		default:
			return fmt.Errorf("unsupported auth method %q", authKind)
		}
	}

	auth.Validators = validators
	return nil
}

// Validate routes ctx to the configured validator for ctx.AuthKind.
func (auth *Auth) Validate(ctx *AuthContext) error {
	if auth == nil || len(auth.Validators) == 0 {
		return fmt.Errorf("no auth validators configured")
	}
	if ctx == nil {
		return fmt.Errorf("auth context is required")
	}
	if ctx.AuthKind == "" {
		return fmt.Errorf("auth kind is required")
	}

	validator, exists := auth.Validators[ctx.AuthKind]
	if !exists {
		return fmt.Errorf("unsupported auth method %q", ctx.AuthKind)
	}
	return validator.Validate(ctx)
}

// decodeAuthConfig unmarshals a per-auth-kind config node into target.
//
// This helper exists because Auth.UnmarshalYAML first parses the `auth` node into a
// generic map, then re-hydrates each auth-kind subtree into its strongly-typed config.
func decodeAuthConfig(rawConfig interface{}, target interface{}) error {
	if rawConfig == nil {
		return yaml.Unmarshal([]byte(`{}`), target)
	}

	data, err := yaml.Marshal(rawConfig)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, target)
}
