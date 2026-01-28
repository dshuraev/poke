package executor

import (
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"
)

type EnvStrategy string
type EnvMap map[string]string

// Represents environmental configuration for a single `Command`.
type Env struct {
	// Env variable merge strategy, must be `EnvStrategyInherit`, `EnvStrategyIsolate`,
	// `EnvStrategyExtend`, or `EnvStrategyOverride`
	Strategy EnvStrategy `yaml:"strategy,omitempty"`
	// Map `key:val` of environmental variables provided with the command
	Vals EnvMap `yaml:"vals,omitempty"`
}

const (
	EnvStrategyInherit  EnvStrategy = "inherit"  // use parent process env
	EnvStrategyIsolate  EnvStrategy = "isolate"  // use only specified vars (clean)
	EnvStrategyExtend   EnvStrategy = "extend"   // parent env + non-replacing variable merge
	EnvStrategyOverride EnvStrategy = "override" // parent env + replacing veriable merge
)

func NewEnvDefault() Env {
	return Env{
		Strategy: EnvStrategyIsolate,
		Vals:     make(EnvMap),
	}
}

// UnmarshalYAML parses env config per docs/configuration/command.md.
// Keys and values are always interpreted as strings.
func (env *Env) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type envAlias Env

	*env = NewEnvDefault()

	var in envAlias
	if err := unmarshal(&in); err != nil {
		return err
	}

	if in.Strategy != "" {
		env.Strategy = in.Strategy
	}
	if in.Vals != nil {
		env.Vals = in.Vals
	}
	return nil
}

// UnmarshalYAML converts any YAML map into string key/value pairs.
func (env *EnvMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[interface{}]interface{}
	if err := unmarshal(&raw); err == nil {
		*env = mapToEnvMap(raw)
		return nil
	}

	var rawString map[string]interface{}
	if err := unmarshal(&rawString); err != nil {
		return err
	}

	out := make(EnvMap, len(rawString))
	for key, val := range rawString {
		out[toEnvString(key)] = toEnvString(val)
	}
	*env = out
	return nil
}

func mapToEnvMap(raw map[interface{}]interface{}) EnvMap {
	if len(raw) == 0 {
		return nil
	}
	out := make(EnvMap, len(raw))
	for key, val := range raw {
		out[toEnvString(key)] = toEnvString(val)
	}
	return out
}

func toEnvString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func (cfg Env) Get() EnvMap {
	switch cfg.Strategy {
	case EnvStrategyInherit:
		return getDefaultEnv()
	case EnvStrategyIsolate:
		result := make(EnvMap, len(cfg.Vals))
		maps.Copy(result, cfg.Vals)
		return result
	case EnvStrategyExtend:
		parent := getDefaultEnv()
		return mergeExtend(parent, cfg.Vals)
	case EnvStrategyOverride:
		parent := getDefaultEnv()
		return mergeOverwrite(parent, cfg.Vals)
	default:
		panic("Unreachable branch: invalid env merge strategy")
	}
}

func (env EnvMap) ToList() []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+env[k])
	}
	return out
}

func getDefaultEnv() EnvMap {
	env := make(EnvMap)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		} else {
			env[parts[0]] = ""
		}
	}
	return env
}

// merges two maps by overwriting `parent` with `newVals` on conflict
func mergeOverwrite(parent EnvMap, newVals EnvMap) EnvMap {
	result := make(EnvMap, len(parent)+len(newVals))
	maps.Copy(result, parent)
	maps.Copy(result, newVals)
	return result
}

// merges two maps but extending `parent` with new keys from `extension`
func mergeExtend(parent EnvMap, extension EnvMap) EnvMap {
	return mergeOverwrite(extension, parent)
}
