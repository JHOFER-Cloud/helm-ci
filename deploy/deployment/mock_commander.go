package deployment

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// MockCommander implements Commander for testing
type MockCommander struct {
	// Track commands that were executed
	Commands []MockCommand

	// Predefined responses for specific commands
	Responses map[string]MockResponse

	// Default response if no specific response is found
	DefaultResponse MockResponse
}

// MockCommand represents a command that was executed
type MockCommand struct {
	Name string
	Args []string
}

// MockResponse represents a mock response for a command
type MockResponse struct {
	Output []byte
	Error  error
}

// Command records the command and returns a nil exec.Cmd
// The actual command won't be executed
func (m *MockCommander) Command(name string, args ...string) *exec.Cmd {
	cmd := MockCommand{
		Name: name,
		Args: args,
	}
	m.Commands = append(m.Commands, cmd)

	// Return a real command but we'll intercept the execution
	return exec.Command("echo", "")
}

// getResponseForCommand returns the appropriate response for a command
func (m *MockCommander) getResponseForCommand() MockResponse {
	// Get the last recorded command
	if len(m.Commands) == 0 {
		return m.DefaultResponse
	}

	lastCommand := m.Commands[len(m.Commands)-1]

	// Try different key patterns to match the command

	// 1. Try exact match with full args
	fullKey := fmt.Sprintf("%s:%s", lastCommand.Name, strings.Join(lastCommand.Args, ":"))
	if response, ok := m.Responses[fullKey]; ok {
		return response
	}

	// 2. Try match with command and first two args (common for helm repo add/update)
	if len(lastCommand.Args) >= 2 {
		cmdWithArgs := fmt.Sprintf("%s:%s:%s", lastCommand.Name, lastCommand.Args[0], lastCommand.Args[1])
		if response, ok := m.Responses[cmdWithArgs]; ok {
			return response
		}
	}

	// 3. Try match with command and first arg
	if len(lastCommand.Args) >= 1 {
		cmdWithArg := fmt.Sprintf("%s:%s", lastCommand.Name, lastCommand.Args[0])
		if response, ok := m.Responses[cmdWithArg]; ok {
			return response
		}
	}

	// 4. Just the command name
	if response, ok := m.Responses[lastCommand.Name]; ok {
		return response
	}

	// No specific response found, use default
	return m.DefaultResponse
}

// CombinedOutput returns the mock output for the command
func (m *MockCommander) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	response := m.getResponseForCommand()
	return response.Output, response.Error
}

// Run returns the mock error for the command
func (m *MockCommander) Run(cmd *exec.Cmd) error {
	response := m.getResponseForCommand()
	return response.Error
}

// Output returns the mock output for the command
func (m *MockCommander) Output(cmd *exec.Cmd) ([]byte, error) {
	response := m.getResponseForCommand()
	return response.Output, response.Error
}

// NewMockCommander creates a new MockCommander with default responses
func NewMockCommander() *MockCommander {
	return &MockCommander{
		Commands:  []MockCommand{},
		Responses: make(map[string]MockResponse),
		DefaultResponse: MockResponse{
			Output: []byte(""),
			Error:  nil,
		},
	}
}

// AddResponse adds a mock response for a specific command
func (m *MockCommander) AddResponse(command string, output []byte, err error) {
	m.Responses[command] = MockResponse{
		Output: output,
		Error:  err,
	}
}

// GetCommand returns the command at the given index
func (m *MockCommander) GetCommand(index int) (MockCommand, error) {
	if index >= len(m.Commands) {
		return MockCommand{}, errors.New("command index out of range")
	}
	return m.Commands[index], nil
}

// GetLastCommand returns the last command that was executed
func (m *MockCommander) GetLastCommand() (MockCommand, error) {
	if len(m.Commands) == 0 {
		return MockCommand{}, errors.New("no commands executed")
	}
	return m.Commands[len(m.Commands)-1], nil
}

// ExitError is a mock implementation of exec.ExitError
type ExitError struct {
	Err         error
	ExitedEarly bool
	CodeValue   int // Renamed from ExitCode to avoid conflict with method
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "exit status 1"
}

func (e *ExitError) String() string {
	return e.Error()
}

func (e *ExitError) Exited() bool {
	return e.ExitedEarly
}

func (e *ExitError) ExitCode() int {
	return e.CodeValue // Return the CodeValue field instead
}
