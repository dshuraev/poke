package dispatch_test

import (
	"testing"
	"time"

	"poke/internal/server/dispatch"
	"poke/internal/server/executor"

	"github.com/goccy/go-yaml"
)

// TestCommandRegistryUnmarshalShorthand verifies string and list forms are parsed.
func TestCommandRegistryUnmarshalShorthand(t *testing.T) {
	input := []byte(`
uptime: uptime
query-fs: ["df", "-h"]
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	uptime := getCommand(t, &reg, "uptime")
	if len(uptime.Args) != 1 || uptime.Args[0] != "uptime" {
		t.Fatalf("uptime args: got %#v", uptime.Args)
	}
	if uptime.Executor != "bin" {
		t.Fatalf("uptime executor: got %q", uptime.Executor)
	}
	if uptime.Env.Strategy != executor.EnvStrategyIsolate {
		t.Fatalf("uptime env strategy: got %q", uptime.Env.Strategy)
	}

	query := getCommand(t, &reg, "query-fs")
	if len(query.Args) != 2 || query.Args[0] != "df" || query.Args[1] != "-h" {
		t.Fatalf("query-fs args: got %#v", query.Args)
	}
}

// TestCommandRegistryUnmarshalObject verifies object form is parsed and defaults apply.
func TestCommandRegistryUnmarshalObject(t *testing.T) {
	input := []byte(`
hello:
  name: "Hello"
  description: "Say hello"
  args: ["echo", "hello world"]
  executor: "bin"
  env:
    strategy: override
    vals:
      FOO: bar
  timeout: "5s"
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	cmd := getCommand(t, &reg, "hello")
	if cmd.Name != "Hello" {
		t.Fatalf("name: got %q", cmd.Name)
	}
	if cmd.Description != "Say hello" {
		t.Fatalf("description: got %q", cmd.Description)
	}
	if cmd.Executor != "bin" {
		t.Fatalf("executor: got %q", cmd.Executor)
	}
	if cmd.Timeout != 5*time.Second {
		t.Fatalf("timeout: got %v", cmd.Timeout)
	}
	if cmd.Env.Strategy != executor.EnvStrategyOverride {
		t.Fatalf("env strategy: got %q", cmd.Env.Strategy)
	}
	if cmd.Env.Vals["FOO"] != "bar" {
		t.Fatalf("env vals: got %#v", cmd.Env.Vals)
	}
}

// TestCommandRegistrySetsID ensures command IDs are populated from map keys.
func TestCommandRegistrySetsID(t *testing.T) {
	input := []byte(`
uptime: uptime
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	cmd := getCommand(t, &reg, "uptime")
	if cmd.ID != "uptime" {
		t.Fatalf("id: got %q", cmd.ID)
	}
}

// TestCommandRegistryRejectsDuplicateIDs validates duplicate command keys are rejected.
func TestCommandRegistryRejectsDuplicateIDs(t *testing.T) {
	input := []byte(`
dup: uptime
dup: ["echo", "hello"]
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err == nil {
		t.Fatalf("expected error for duplicate command ids")
	}
}

// TestCommandRegistryRejectsNonStringID validates keys must be strings.
func TestCommandRegistryRejectsNonStringID(t *testing.T) {
	input := []byte(`
1: uptime
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err == nil {
		t.Fatalf("expected error for non-string command id")
	}
}

// TestCommandRegistryRejectsEmptyID validates empty keys are rejected.
func TestCommandRegistryRejectsEmptyID(t *testing.T) {
	input := []byte(`
"": uptime
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err == nil {
		t.Fatalf("expected error for empty command id")
	}
}

// TestCommandRegistryRejectsMissingArgs validates object form requires args.
func TestCommandRegistryRejectsMissingArgs(t *testing.T) {
	input := []byte(`
nope:
  name: "nope"
`)

	var reg dispatch.CommandRegistry
	if err := yaml.Unmarshal(input, &reg); err == nil {
		t.Fatalf("expected error for missing command args")
	}
}

// getCommand fetches a command by ID and fails the test on error.
func getCommand(t *testing.T, reg *dispatch.CommandRegistry, id string) executor.Command {
	t.Helper()

	cmd, err := reg.Get(id)
	if err != nil {
		t.Fatalf("get %s: %v", id, err)
	}
	return cmd
}
