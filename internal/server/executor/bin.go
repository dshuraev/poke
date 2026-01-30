package executor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
)

func ExecuteBinary(ctx context.Context, cmd Command) Result {
	log.Printf("executor/bin: start command %s[%s]", cmd.ID, cmd.Name)
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
		log.Printf("executor/bin: invalid command %s[%s]: %v", cmd.ID, cmd.Name, err)
		return Result{
			ExitCode: -1,
			Error:    err,
		}
	}

	log.Printf("executor/bin: exec argv=%v", cmd.Args)
	// #nosec G204 -- commands are configured by trusted config after validation.
	cmdExec := exec.CommandContext(cmdCtx, cmd.Args[0], cmd.Args[1:]...)
	cmdExec.Env = cmd.Env.Get().ToList()
	output, err := cmdExec.CombinedOutput()

	if cmdExec.ProcessState == nil {
		execErr := err
		if execErr == nil {
			execErr = errors.New("unknown error")
		}
		log.Printf("executor/bin: failed to execute %s[%s]: %v", cmd.ID, cmd.Name, execErr)
		return Result{
			Output:   output,
			ExitCode: -1,
			Error:    fmt.Errorf("command %s[%s] failed to execute: %w", cmd.ID, cmd.Name, execErr),
		}
	}

	exitCode := cmdExec.ProcessState.ExitCode()
	if err != nil {
		log.Printf("executor/bin: exit %d for %s[%s]: %v", exitCode, cmd.ID, cmd.Name, err)
	} else {
		log.Printf("executor/bin: exit %d for %s[%s]", exitCode, cmd.ID, cmd.Name)
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
