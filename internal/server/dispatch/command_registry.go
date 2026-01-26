package dispatch

import (
	"fmt"
	"poke/internal/server/executor"
)

type CommandRegistry struct {
	cmds map[string]executor.Command
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
