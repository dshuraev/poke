package executor

import "time"

// `Command` struct represents an executable command that is registered with
// poke server.
type Command struct {
	ID          string        // Unique identifier for the command, used for loookup
	Name        string        // Human-readable name of the command, not necessarily unique
	Description string        // Human-readable command description
	Args        []string      // Command arguments
	Executor    string        // Command executor, used to lookup the executor for command
	Env         Env           // Environmental configuration: vars, merge strategy
	Timeout     time.Duration // Command timeout, 0 = no timeout, use with caution
}
