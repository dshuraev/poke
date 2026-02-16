package executor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
)

// ExecuteBinary runs a configured command using os/exec and returns execution result.
func ExecuteBinary(ctx context.Context, cmd Command) Result {
	logger := slog.Default().With("component", "executor/bin")
	logger.Info("binary execution started", "event", "binary_execution_started", "command_id", cmd.ID, "command_name", cmd.Name)

	var cmdCtx context.Context
	var cancel context.CancelFunc
	if cmd.Timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, cmd.Timeout)
	} else {
		cmdCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	err := validateCommand(cmd)
	if err != nil {
		logger.Warn("invalid command", "event", "binary_command_invalid", "command_id", cmd.ID, "command_name", cmd.Name, "error", err)
		return Result{
			ExitCode: -1,
			Error:    err,
		}
	}

	logger.Debug("invoking command", "event", "binary_command_invoking", "command_id", cmd.ID, "command_name", cmd.Name, "args", cmd.Args)
	// #nosec G204 -- commands are configured by trusted config after validation.
	cmdExec := exec.CommandContext(cmdCtx, cmd.Args[0], cmd.Args[1:]...)
	cmdExec.Env = cmd.Env.Get().ToList()
	output, err := cmdExec.CombinedOutput()

	if cmdExec.ProcessState == nil {
		execErr := err
		if execErr == nil {
			execErr = errors.New("unknown error")
		}
		logger.Error("failed to execute command", "event", "binary_execution_failed_to_start", "command_id", cmd.ID, "command_name", cmd.Name, "error", execErr)
		return Result{
			Output:   output,
			ExitCode: -1,
			Error:    fmt.Errorf("command %s[%s] failed to execute: %w", cmd.ID, cmd.Name, execErr),
		}
	}

	exitCode := cmdExec.ProcessState.ExitCode()
	if err != nil {
		logger.Error("command exited with error", "event", "binary_execution_completed_with_error", "command_id", cmd.ID, "command_name", cmd.Name, "exit_code", exitCode, "error", err)
	} else {
		logger.Info("command completed", "event", "binary_execution_completed", "command_id", cmd.ID, "command_name", cmd.Name, "exit_code", exitCode)
	}
	return Result{
		Output:   output,
		ExitCode: exitCode,
		Error:    err,
	}
}

func validateCommand(cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("command %s[%s] has no arguments", cmd.ID, cmd.Name)
	}
	return nil
}
