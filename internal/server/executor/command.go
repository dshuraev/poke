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

	// check if user provided a single string, like `current-dir: pwd`
	var asString string
	if err := unmarshal(&asString); err == nil {
		cmd.Args = []string{asString}
		return nil
	}

	// check if user provided a list of arguments, like `query-fs: ['df', '-h']`
	var asList []string
	if err := unmarshal(&asList); err == nil {
		cmd.Args = asList
		return nil
	}

	// fallthrough: unmarshall full spec
	type commandAlias Command
	var in commandAlias
	if err := unmarshal(&in); err != nil {
		return err
	}

	// merge non-empty with defaults
	inCmd := Command(in)
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

	if len(cmd.Args) == 0 {
		return fmt.Errorf("command has no arguments")
	}

	return nil
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

func parseArgsValue(value interface{}) ([]string, error) {
	switch raw := value.(type) {
	case nil:
		return nil, nil
	case string:
		return []string{raw}, nil
	case []string:
		return raw, nil
	case []interface{}:
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			out = append(out, toEnvString(item))
		}
		return out, nil
	default:
		return nil, fmt.Errorf("args must be a string or list of strings")
	}
}
