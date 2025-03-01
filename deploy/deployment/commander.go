package deployment

import (
	"os/exec"
)

// Commander provides an interface for executing commands
// This allows us to mock exec.Command in tests
type Commander interface {
	Command(name string, args ...string) *exec.Cmd
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
}

// RealCommander implements Commander using actual exec functionality
type RealCommander struct{}

// Command creates a new exec.Cmd
func (c *RealCommander) Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// CombinedOutput runs the command and returns its combined stdout and stderr
func (c *RealCommander) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

// Run starts the command and waits for it to complete
func (c *RealCommander) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Output runs the command and returns its stdout
func (c *RealCommander) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}
