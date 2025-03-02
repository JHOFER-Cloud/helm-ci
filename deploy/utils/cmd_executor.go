package utils

import (
	"os"
	"os/exec"
)

// CommandExecutor is an interface for executing shell commands
// This allows us to mock command execution in tests
type CommandExecutor interface {
	// Execute runs a command with args and returns the combined output and any error
	Execute(name string, args ...string) ([]byte, error)

	// ExecuteWithStdio runs a command with stdin/stdout/stderr connected to os.Stdin/os.Stdout/os.Stderr
	ExecuteWithStdio(name string, args ...string) error

	// CommandExists checks if a command exists in the PATH
	CommandExists(name string) bool
}

// RealCommandExecutor uses the actual os/exec package to run commands
type RealCommandExecutor struct{}

// Execute runs a shell command and returns its combined output
func (e *RealCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// ExecuteWithStdio runs a command with the specified standard I/O
func (e *RealCommandExecutor) ExecuteWithStdio(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// CommandExists checks if a command exists in the PATH
func (e *RealCommandExecutor) CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Default executor for production use
var DefaultExecutor CommandExecutor = &RealCommandExecutor{}

// MockCommandExecutor mocks command execution for testing
type MockCommandExecutor struct {
	// MockOutputs maps command+args strings to their expected outputs and errors
	MockOutputs map[string]struct {
		Output []byte
		Err    error
	}

	// ExecutionLog records all commands executed
	ExecutionLog []struct {
		Command string
		Args    []string
	}
}

// NewMockCommandExecutor creates a new mock executor
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		MockOutputs: make(map[string]struct {
			Output []byte
			Err    error
		}),
		ExecutionLog: make([]struct {
			Command string
			Args    []string
		}, 0),
	}
}

// Execute returns mock outputs for the given command and args
func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	// Record this execution
	m.ExecutionLog = append(m.ExecutionLog, struct {
		Command string
		Args    []string
	}{
		Command: name,
		Args:    args,
	})

	// Generate a key for the mock outputs map
	key := name
	for _, arg := range args {
		key += " " + arg
	}

	// Return the mock output if it exists
	if mock, ok := m.MockOutputs[key]; ok {
		return mock.Output, mock.Err
	}

	// If we have a catch-all command, use that
	if mock, ok := m.MockOutputs[name]; ok {
		return mock.Output, mock.Err
	}

	// Default to empty output and nil error
	return []byte{}, nil
}

// ExecuteWithStdio mocks execution with stdio
func (m *MockCommandExecutor) ExecuteWithStdio(name string, args ...string) error {
	// Record this execution
	m.ExecutionLog = append(m.ExecutionLog, struct {
		Command string
		Args    []string
	}{
		Command: name,
		Args:    args,
	})

	// Generate a key for the mock outputs map
	key := name
	for _, arg := range args {
		key += " " + arg
	}

	// Return the mock error if it exists
	if mock, ok := m.MockOutputs[key]; ok {
		if len(mock.Output) > 0 {
			os.Stdout.Write(mock.Output)
		}
		return mock.Err
	}

	// If we have a catch-all command, use that
	if mock, ok := m.MockOutputs[name]; ok {
		if len(mock.Output) > 0 {
			os.Stdout.Write(mock.Output)
		}
		return mock.Err
	}

	// Default to nil error
	return nil
}

// CommandExists always returns true for mock executor
func (m *MockCommandExecutor) CommandExists(name string) bool {
	return true
}

// MockOutput adds a mock output for a command
func (m *MockCommandExecutor) MockOutput(command string, output []byte, err error) {
	m.MockOutputs[command] = struct {
		Output []byte
		Err    error
	}{
		Output: output,
		Err:    err,
	}
}

// GetExecutionCount returns the number of times a command was executed
func (m *MockCommandExecutor) GetExecutionCount(command string) int {
	count := 0
	for _, exec := range m.ExecutionLog {
		if exec.Command == command {
			count++
		}
	}
	return count
}

// HasExecuted checks if a specific command with args was executed
func (m *MockCommandExecutor) HasExecuted(command string, args ...string) bool {
	for _, exec := range m.ExecutionLog {
		if exec.Command != command || len(exec.Args) != len(args) {
			continue
		}

		match := true
		for i, arg := range args {
			if exec.Args[i] != arg {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}
	return false
}

// Reset clears all execution logs
func (m *MockCommandExecutor) Reset() {
	m.ExecutionLog = m.ExecutionLog[:0]
}
