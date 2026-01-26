package executor

// The result of command execution
type Result struct {
	Output   []byte
	ExitCode int
	Error    error
}
