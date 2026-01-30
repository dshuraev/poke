package executor

import (
	"fmt"
	"time"
)

// `Command` struct represents an executable command that is registered with
// poke server.
type Command struct {
	ID          string        `yaml:"-"`                     // Unique identifier for the command, used for loookup
	Name        string        `yaml:"name,omitempty"`        // Human-readable name of the command, not necessarily unique
	Description string        `yaml:"description,omitempty"` // Human-readable command description
	Args        []string      `yaml:"args,omitempty"`        // Command arguments
	Executor    string        `yaml:"executor,omitempty"`    // Command executor, used to lookup the executor for command
	Env         Env           `yaml:"env,omitempty"`         // Environmental configuration: vars, merge strategy
	Timeout     time.Duration `yaml:"timeout,omitempty"`     // Command timeout, 0 = no timeout, use with caution
}

const defaultExecutorName = "bin"

// NewCommandDefault returns a Command populated with default executor and env.
func NewCommandDefault() Command {
	return Command{
		Executor: defaultExecutorName,
		Env:      NewEnvDefault(),
	}
}

// UnmarshalYAML parses command config per docs/configuration/command.md.
// Supported forms:
// - string: single argument
// - list: list of arguments
// - object: full command specification
//
// Important: Command.ID is NOT unmarshalled - it has to be set by caller
func (cmd *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*cmd = NewCommandDefault()

	args, handled, err := unmarshalCommandArgsAsString(unmarshal)
	if err != nil {
		return err
	}
	if handled {
		cmd.Args = args
		return nil
	}

	args, handled, err = unmarshalCommandArgsAsList(unmarshal)
	if err != nil {
		return err
	}
	if handled {
		cmd.Args = args
		return nil
	}

	inCmd, err := unmarshalCommandSpec(unmarshal)
	if err != nil {
		return err
	}

	applyCommandOverrides(cmd, inCmd)
	return validateCommandArgs(cmd.Args)
}

// MarshalYAML renders commands in short or object form depending on fields set.
func (cmd Command) MarshalYAML() (interface{}, error) {
	defaultEnv := NewEnvDefault()
	envIsDefault := cmd.Env.Strategy == defaultEnv.Strategy && len(cmd.Env.Vals) == 0
	executorIsDefault := cmd.Executor == "" || cmd.Executor == defaultExecutorName

	if cmd.Name == "" &&
		cmd.Description == "" &&
		cmd.Timeout == 0 &&
		envIsDefault &&
		executorIsDefault {
		if len(cmd.Args) == 1 {
			return cmd.Args[0], nil
		}
		return cmd.Args, nil
	}

	type commandAlias Command
	return commandAlias(cmd), nil
}

// unmarshalCommandArgsAsString tries the single-argument shorthand form.
func unmarshalCommandArgsAsString(unmarshal func(interface{}) error) ([]string, bool, error) {
	var asString string
	if err := unmarshal(&asString); err != nil {
		return nil, false, nil
	}

	return []string{asString}, true, nil
}

// unmarshalCommandArgsAsList tries the list-of-arguments shorthand form.
func unmarshalCommandArgsAsList(unmarshal func(interface{}) error) ([]string, bool, error) {
	var asList []string
	if err := unmarshal(&asList); err != nil {
		return nil, false, nil
	}

	return asList, true, nil
}

// unmarshalCommandSpec parses the full command specification form.
func unmarshalCommandSpec(unmarshal func(interface{}) error) (Command, error) {
	type commandAlias Command
	var in commandAlias
	if err := unmarshal(&in); err != nil {
		return Command{}, err
	}

	return Command(in), nil
}

// applyCommandOverrides merges non-empty values from inCmd onto cmd.
func applyCommandOverrides(cmd *Command, inCmd Command) {
	if inCmd.Name != "" {
		cmd.Name = inCmd.Name
	}
	if inCmd.Description != "" {
		cmd.Description = inCmd.Description
	}
	if inCmd.Executor != "" {
		cmd.Executor = inCmd.Executor
	}
	if len(inCmd.Args) > 0 {
		cmd.Args = inCmd.Args
	}

	cmd.Timeout = inCmd.Timeout
	if inCmd.Env.Strategy != "" || len(inCmd.Env.Vals) > 0 {
		cmd.Env = inCmd.Env
	}
}

// validateCommandArgs ensures commands are always configured with arguments.
func validateCommandArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command has no arguments")
	}

	return nil
}
