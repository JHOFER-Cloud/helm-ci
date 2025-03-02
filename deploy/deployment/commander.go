// Copyright 2025 Josef Hofer (JHOFER-Cloud)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
