package dispatch

import (
	"fmt"
	"poke/internal/server/executor"
	"sort"

	"github.com/goccy/go-yaml"
)

type CommandRegistry struct {
	cmds map[string]executor.Command
}

// UnmarshalYAML parses commands config per docs/configuration/command.md.
func (reg *CommandRegistry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw yaml.MapSlice
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if raw == nil {
		reg.cmds = map[string]executor.Command{}
		return nil
	}

	cmds := make(map[string]executor.Command, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		id, ok := item.Key.(string)
		if !ok {
			return fmt.Errorf("command id must be string, got %T", item.Key)
		}
		if id == "" {
			return fmt.Errorf("command id must not be empty")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate command id %q", id)
		}
		seen[id] = struct{}{}

		cmd, err := decodeCommandConfig(item.Value)
		if err != nil {
			return fmt.Errorf("command %s: %w", id, err)
		}
		cmd.ID = id
		cmds[id] = cmd
	}

	reg.cmds = cmds
	return nil
}

func NewCommandRegistry(cmds map[string]executor.Command) *CommandRegistry {
	if cmds == nil {
		cmds = make(map[string]executor.Command)
	}
	return &CommandRegistry{cmds: cmds}
}

func (reg *CommandRegistry) Register(id string, cmd executor.Command) {
	if reg.cmds == nil {
		reg.cmds = make(map[string]executor.Command)
	}
	reg.cmds[id] = cmd
}

func (reg *CommandRegistry) Get(id string) (executor.Command, error) {
	cmd, exists := reg.cmds[id]
	if exists {
		return cmd, nil
	}
	return executor.Command{}, fmt.Errorf("command with ID %s not found", id)
}

// ExecutorNames returns the unique executor names used by registered commands.
func (reg *CommandRegistry) ExecutorNames() []string {
	if reg == nil || len(reg.cmds) == 0 {
		return nil
	}

	defaultExecutor := executor.NewCommandDefault().Executor
	executors := make(map[string]struct{}, len(reg.cmds))
	for _, cmd := range reg.cmds {
		name := cmd.Executor
		if name == "" {
			name = defaultExecutor
		}
		if name != "" {
			executors[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(executors))
	for name := range executors {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// decodeCommandConfig unmarshals a per-command config node into a Command.
func decodeCommandConfig(rawConfig interface{}) (executor.Command, error) {
	var cmd executor.Command
	if rawConfig == nil {
		if err := yaml.Unmarshal([]byte(`{}`), &cmd); err != nil {
			return executor.Command{}, err
		}
		return cmd, nil
	}

	data, err := yaml.Marshal(rawConfig)
	if err != nil {
		return executor.Command{}, err
	}

	if err := yaml.Unmarshal(data, &cmd); err != nil {
		return executor.Command{}, err
	}

	return cmd, nil
}
