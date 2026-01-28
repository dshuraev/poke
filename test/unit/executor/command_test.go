package executor_test

import (
	"testing"
	"time"

	"poke/internal/server/executor"

	"github.com/goccy/go-yaml"
)

func TestCommandUnmarshalString(t *testing.T) {
	var got executor.Command
	if err := yaml.Unmarshal([]byte(`uptime`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Args) != 1 || got.Args[0] != "uptime" {
		t.Fatalf("args: got %#v", got.Args)
	}
	if got.Executor != "bin" {
		t.Fatalf("executor: got %q", got.Executor)
	}
	if got.Timeout != 0 {
		t.Fatalf("timeout: got %v", got.Timeout)
	}
	if got.Env.Strategy != executor.EnvStrategyIsolate {
		t.Fatalf("env strategy: got %q", got.Env.Strategy)
	}
	if len(got.Env.Vals) != 0 {
		t.Fatalf("env vals: got %#v", got.Env.Vals)
	}
}

func TestCommandUnmarshalArgsList(t *testing.T) {
	var got executor.Command
	if err := yaml.Unmarshal([]byte(`["df", "-h"]`), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Args) != 2 || got.Args[0] != "df" || got.Args[1] != "-h" {
		t.Fatalf("args: got %#v", got.Args)
	}
}

func TestCommandUnmarshalObject(t *testing.T) {
	input := []byte(`
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

	var got executor.Command
	if err := yaml.Unmarshal(input, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != "Hello" {
		t.Fatalf("name: got %q", got.Name)
	}
	if got.Description != "Say hello" {
		t.Fatalf("description: got %q", got.Description)
	}
	if got.Executor != "bin" {
		t.Fatalf("executor: got %q", got.Executor)
	}
	if got.Timeout != 5*time.Second {
		t.Fatalf("timeout: got %v", got.Timeout)
	}
	if got.Env.Strategy != executor.EnvStrategyOverride {
		t.Fatalf("env strategy: got %q", got.Env.Strategy)
	}
	if got.Env.Vals["FOO"] != "bar" {
		t.Fatalf("env vals: got %#v", got.Env.Vals)
	}
}

func TestCommandMarshalShortForms(t *testing.T) {
	single := executor.Command{
		Args:     []string{"uptime"},
		Executor: "bin",
		Env:      executor.NewEnvDefault(),
	}

	data, err := yaml.Marshal(single)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got string
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != "uptime" {
		t.Fatalf("short string: got %q", got)
	}

	list := executor.Command{
		Args:     []string{"df", "-h"},
		Executor: "bin",
		Env:      executor.NewEnvDefault(),
	}
	data, err = yaml.Marshal(list)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var gotList []string
	if err := yaml.Unmarshal(data, &gotList); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(gotList) != 2 || gotList[0] != "df" || gotList[1] != "-h" {
		t.Fatalf("short list: got %#v", gotList)
	}
}

func TestCommandMarshalObjectWhenConfigured(t *testing.T) {
	cmd := executor.Command{
		Name:        "Hello",
		Description: "Say hello",
		Args:        []string{"echo", "hello"},
		Executor:    "bin",
		Env: executor.Env{
			Strategy: executor.EnvStrategyOverride,
			Vals: executor.EnvMap{
				"FOO": "bar",
			},
		},
		Timeout: 2 * time.Second,
	}

	data, err := yaml.Marshal(cmd)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]interface{}
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["name"] != "Hello" {
		t.Fatalf("name: got %#v", got["name"])
	}
	if got["description"] != "Say hello" {
		t.Fatalf("description: got %#v", got["description"])
	}
	if got["timeout"] == nil {
		t.Fatalf("timeout missing")
	}
	if got["env"] == nil {
		t.Fatalf("env missing")
	}
}

func TestCommandUnmarshalRequiresArgs(t *testing.T) {
	var got executor.Command
	if err := yaml.Unmarshal([]byte(`{name: "nope"}`), &got); err == nil {
		t.Fatalf("expected error for missing args")
	}
}
