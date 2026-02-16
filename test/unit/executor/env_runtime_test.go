package executor_test

import (
	"testing"

	"poke/internal/server/executor"
)

func TestEnvGetInheritIncludesParentVariables(t *testing.T) {
	const key = "POKE_TEST_ENV_INHERIT"
	const value = "inherit-value"
	t.Setenv(key, value)

	got := executor.Env{Strategy: executor.EnvStrategyInherit}.Get()
	if got[key] != value {
		t.Fatalf("%s: got %q want %q", key, got[key], value)
	}
}

func TestEnvGetIsolateUsesOnlyProvidedValues(t *testing.T) {
	const parentKey = "POKE_TEST_ENV_ISOLATE_PARENT"
	t.Setenv(parentKey, "parent-value")

	got := executor.Env{
		Strategy: executor.EnvStrategyIsolate,
		Vals: executor.EnvMap{
			"POKE_TEST_ENV_ISOLATE": "isolated",
		},
	}.Get()

	if len(got) != 1 {
		t.Fatalf("len: got %d want 1", len(got))
	}
	if got[parentKey] != "" {
		t.Fatalf("%s should not be inherited, got %q", parentKey, got[parentKey])
	}
	if got["POKE_TEST_ENV_ISOLATE"] != "isolated" {
		t.Fatalf("POKE_TEST_ENV_ISOLATE: got %q want %q", got["POKE_TEST_ENV_ISOLATE"], "isolated")
	}
}

func TestEnvGetExtendKeepsParentValuesOnConflict(t *testing.T) {
	const sharedKey = "POKE_TEST_ENV_EXTEND_SHARED"
	const parentOnlyKey = "POKE_TEST_ENV_EXTEND_PARENT_ONLY"
	const extensionOnlyKey = "POKE_TEST_ENV_EXTEND_EXTENSION_ONLY"

	t.Setenv(sharedKey, "parent")
	t.Setenv(parentOnlyKey, "parent-only")

	got := executor.Env{
		Strategy: executor.EnvStrategyExtend,
		Vals: executor.EnvMap{
			sharedKey:        "extension",
			extensionOnlyKey: "extension-only",
		},
	}.Get()

	if got[sharedKey] != "parent" {
		t.Fatalf("shared key: got %q want %q", got[sharedKey], "parent")
	}
	if got[parentOnlyKey] != "parent-only" {
		t.Fatalf("parent-only key: got %q want %q", got[parentOnlyKey], "parent-only")
	}
	if got[extensionOnlyKey] != "extension-only" {
		t.Fatalf("extension-only key: got %q want %q", got[extensionOnlyKey], "extension-only")
	}
}

func TestEnvGetOverrideReplacesParentValuesOnConflict(t *testing.T) {
	const sharedKey = "POKE_TEST_ENV_OVERRIDE_SHARED"
	const parentOnlyKey = "POKE_TEST_ENV_OVERRIDE_PARENT_ONLY"
	const overrideOnlyKey = "POKE_TEST_ENV_OVERRIDE_ONLY"

	t.Setenv(sharedKey, "parent")
	t.Setenv(parentOnlyKey, "parent-only")

	got := executor.Env{
		Strategy: executor.EnvStrategyOverride,
		Vals: executor.EnvMap{
			sharedKey:       "override",
			overrideOnlyKey: "override-only",
		},
	}.Get()

	if got[sharedKey] != "override" {
		t.Fatalf("shared key: got %q want %q", got[sharedKey], "override")
	}
	if got[parentOnlyKey] != "parent-only" {
		t.Fatalf("parent-only key: got %q want %q", got[parentOnlyKey], "parent-only")
	}
	if got[overrideOnlyKey] != "override-only" {
		t.Fatalf("override-only key: got %q want %q", got[overrideOnlyKey], "override-only")
	}
}

func TestEnvMapToListSortsKeys(t *testing.T) {
	env := executor.EnvMap{
		"B": "2",
		"A": "1",
		"C": "3",
	}

	got := env.ToList()
	want := []string{"A=1", "B=2", "C=3"}

	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d: got %q want %q", i, got[i], want[i])
		}
	}
}
