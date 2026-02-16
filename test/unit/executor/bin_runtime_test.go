package executor_test

import (
	"context"
	"testing"
	"time"

	"poke/internal/server/executor"
)

func TestExecuteBinarySuccess(t *testing.T) {
	cmd := executor.Command{
		ID:       "ok",
		Name:     "ok",
		Args:     []string{"sh", "-c", "printf hello"},
		Env:      executor.NewEnvDefault(),
		Executor: "bin",
	}

	result := executor.ExecuteBinary(context.Background(), cmd)
	if result.Error != nil {
		t.Fatalf("execute: %v", result.Error)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit_code: got %d want 0", result.ExitCode)
	}
	if string(result.Output) != "hello" {
		t.Fatalf("output: got %q want %q", string(result.Output), "hello")
	}
}

func TestExecuteBinaryRejectsCommandWithoutArgs(t *testing.T) {
	cmd := executor.Command{ID: "bad", Name: "bad"}

	result := executor.ExecuteBinary(context.Background(), cmd)
	if result.Error == nil {
		t.Fatalf("expected error")
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit_code: got %d want -1", result.ExitCode)
	}
}

func TestExecuteBinaryReturnsErrorForMissingBinary(t *testing.T) {
	cmd := executor.Command{
		ID:       "missing",
		Name:     "missing",
		Args:     []string{"__poke_binary_that_does_not_exist__"},
		Env:      executor.NewEnvDefault(),
		Executor: "bin",
	}

	result := executor.ExecuteBinary(context.Background(), cmd)
	if result.Error == nil {
		t.Fatalf("expected error")
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit_code: got %d want -1", result.ExitCode)
	}
}

func TestExecuteBinaryHonorsTimeout(t *testing.T) {
	cmd := executor.Command{
		ID:       "timeout",
		Name:     "timeout",
		Args:     []string{"sh", "-c", "sleep 5"},
		Env:      executor.NewEnvDefault(),
		Executor: "bin",
		Timeout:  20 * time.Millisecond,
	}

	started := time.Now()
	result := executor.ExecuteBinary(context.Background(), cmd)
	elapsed := time.Since(started)

	if result.Error == nil {
		t.Fatalf("expected timeout error")
	}
	if elapsed >= 2*time.Second {
		t.Fatalf("expected timeout to stop quickly, elapsed=%v", elapsed)
	}
}
