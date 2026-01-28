package executor_test

import (
	"testing"

	"github.com/goccy/go-yaml"
	"poke/internal/server/executor"
)

func TestEnvRoundTrip(t *testing.T) {
	original := executor.Env{
		Strategy: executor.EnvStrategyOverride,
		Vals: executor.EnvMap{
			"FOO": "hello world",
			"BAR": "42",
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got executor.Env
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Strategy != original.Strategy {
		t.Fatalf("strategy: got %q want %q", got.Strategy, original.Strategy)
	}
	if !envMapEqual(got.Vals, original.Vals) {
		t.Fatalf("vals: got %#v want %#v", got.Vals, original.Vals)
	}
}

func TestEnvUnmarshalStringify(t *testing.T) {
	input := []byte(`strategy: extend
vals:
  FOO: 42
  BAR: true
  BAZ: null
  123: 456
`)

	var got executor.Env
	if err := yaml.Unmarshal(input, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Strategy != executor.EnvStrategyExtend {
		t.Fatalf("strategy: got %q want %q", got.Strategy, executor.EnvStrategyExtend)
	}

	want := executor.EnvMap{
		"FOO": "42",
		"BAR": "true",
		"BAZ": "",
		"123": "456",
	}

	if !envMapEqual(got.Vals, want) {
		t.Fatalf("vals: got %#v want %#v", got.Vals, want)
	}
}

func envMapEqual(a, b executor.EnvMap) bool {
	if len(a) != len(b) {
		return false
	}
	for key, val := range a {
		if b[key] != val {
			return false
		}
	}
	return true
}
