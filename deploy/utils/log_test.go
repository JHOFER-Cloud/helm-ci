package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// Helper for mocking exec.Command
type mockCmd struct {
	output string
	err    error
}

func (m mockCmd) CombinedOutput() ([]byte, error) {
	return []byte(m.output), m.err
}

// Tests for existing functions
func TestSuccess(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test Success function
	testMessage := "Operation successful"
	Success("%s", testMessage)

	logOutput := buf.String()
	if !strings.Contains(logOutput, testMessage) {
		t.Errorf("Success() did not log the message. Got: %s", logOutput)
	}
}

func TestGreen(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test Green function
	testMessage := "Green message"
	Green("%s", testMessage)

	logOutput := buf.String()
	if !strings.Contains(logOutput, testMessage) {
		t.Errorf("Green() did not log the message. Got: %s", logOutput)
	}
}

func TestGreenWithFormatting(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test Green function with formatting
	Green("Count: %d, Name: %s", 42, "test")

	logOutput := buf.String()
	expectedContent := "Count: 42, Name: test"
	if !strings.Contains(logOutput, expectedContent) {
		t.Errorf("Green() did not format correctly. Expected content '%s' not found in: %s",
			expectedContent, logOutput)
	}
}

func TestSuccessWithFormatting(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test Success function with formatting
	Success("Processed %d items in %s", 10, "database")

	logOutput := buf.String()
	expectedContent := "Processed 10 items in database"
	if !strings.Contains(logOutput, expectedContent) {
		t.Errorf("Success() did not format correctly. Expected content '%s' not found in: %s",
			expectedContent, logOutput)
	}
}

func TestLogLevels(t *testing.T) {
	// Save original log level and restore it after test
	origLevel := Log.GetLevel()
	defer Log.SetLevel(origLevel)

	// Test DEBUG environment variable
	t.Run("Environment variable log level", func(t *testing.T) {
		// Set up environment
		os.Setenv("DEBUG", "true")
		defer os.Unsetenv("DEBUG")

		// Reset log level
		InitLogger(false)

		if Log.GetLevel() != logrus.DebugLevel {
			t.Errorf("Expected log level to be set to Debug with DEBUG=true, got %v", Log.GetLevel())
		}
	})

	// Test debug flag
	t.Run("Debug flag log level", func(t *testing.T) {
		// Make sure environment variable isn't affecting test
		os.Unsetenv("DEBUG")

		// Reset log level and set via flag
		InitLogger(true)

		if Log.GetLevel() != logrus.DebugLevel {
			t.Errorf("Expected log level to be set to Debug with flag, got %v", Log.GetLevel())
		}
	})
}

// NEW TESTS BELOW

func TestWrapError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Create an error to wrap
	origErr := errors.New("original error")

	// Test WrapError with simple message
	wrappedErr := WrapError(origErr, "something failed")

	// Verify that the error contains both the wrapper message and original error
	errMsg := wrappedErr.Error()
	if !strings.Contains(errMsg, "something failed") || !strings.Contains(errMsg, "original error") {
		t.Errorf("WrapError() did not properly include both messages. Got: %s", errMsg)
	}

	// Verify that the error was logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "something failed") || !strings.Contains(logOutput, "original error") {
		t.Errorf("WrapError() did not properly log the error. Log output: %s", logOutput)
	}

	// Test with formatting
	buf.Reset()
	formattedErr := WrapError(origErr, "failed processing item %d: %s", 42, "config.yaml")

	// Verify formatted error
	errMsg = formattedErr.Error()
	expectedMsg := "failed processing item 42: config.yaml: original error"
	if errMsg != expectedMsg {
		t.Errorf("WrapError() with formatting produced incorrect error. Got: %s, Want: %s", errMsg, expectedMsg)
	}
}

func TestNewError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test NewError with simple message
	simpleErr := NewError("an error occurred")

	// Verify error message
	if simpleErr.Error() != "an error occurred" {
		t.Errorf("NewError() created incorrect error message. Got: %s", simpleErr.Error())
	}

	// Verify the error was logged
	logOutput := buf.String()
	if !strings.Contains(logOutput, "an error occurred") {
		t.Errorf("NewError() did not log the error. Log output: %s", logOutput)
	}

	// Test with formatting
	buf.Reset()
	formattedErr := NewError("failed with code %d: %s", 404, "not found")

	// Verify formatted error
	expectedMsg := "failed with code 404: not found"
	if formattedErr.Error() != expectedMsg {
		t.Errorf("NewError() with formatting produced incorrect error. Got: %s, Want: %s",
			formattedErr.Error(), expectedMsg)
	}

	// Verify logging with formatting
	logOutput = buf.String()
	if !strings.Contains(logOutput, expectedMsg) {
		t.Errorf("NewError() did not properly log the formatted error. Log output: %s", logOutput)
	}
}

func TestColorizeKubectlDiff(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "added lines",
			input: "normal line\n+added line\nanother normal line",
			expected: []string{
				"normal line",
				greenColor + "+added line" + resetColor,
				"another normal line",
			},
		},
		{
			name:  "removed lines",
			input: "first line\n-removed line\nlast line",
			expected: []string{
				"first line",
				redColor + "-removed line" + resetColor,
				"last line",
			},
		},
		{
			name:  "changed lines",
			input: "unchanged\n~changed line\nunchanged line",
			expected: []string{
				"unchanged",
				yellowColor + "~changed line" + resetColor,
				"unchanged line",
			},
		},
		{
			name:  "mixed changes",
			input: "+added\n-removed\n~changed\nnormal",
			expected: []string{
				greenColor + "+added" + resetColor,
				redColor + "-removed" + resetColor,
				yellowColor + "~changed" + resetColor,
				"normal",
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: []string{
				"", // Split on an empty string gives a slice with one empty string
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ColorizeKubectlDiff(tc.input)
			resultLines := strings.Split(result, "\n")

			// Check if the number of lines matches
			if len(resultLines) != len(tc.expected) {
				t.Fatalf("Expected %d lines, got %d lines", len(tc.expected), len(resultLines))
			}

			// Check each line
			for i, expectedLine := range tc.expected {
				if resultLines[i] != expectedLine {
					t.Errorf("Line %d: expected '%s', got '%s'", i, expectedLine, resultLines[i])
				}
			}
		})
	}
}

// Mock for exec.Command - returns a custom mockCmd that returns predefined output and error
func mockExecCommand(command string, args ...string) *exec.Cmd {
	// Define the expected behavior for kubectl diff
	if command == "kubectl" && len(args) > 0 && args[0] == "diff" {
		// This is a bit of a hack, but it works for simple testing
		// We return a real exec.Command to echo, but we'll never execute it
		// Instead, our mockCmd will provide the output
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}

	// Default behavior
	return exec.Command("echo", "mock command")
}

// TestHelperProcess isn't a real test - it's used as a helper for mockExecCommand
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// We're in the mock process now
	fmt.Fprintf(os.Stdout, "+  added resource\n-  removed resource\n")
	os.Exit(1) // kubectl diff returns 1 if there are differences
}

// TestShowResourceDiff tests the ShowResourceDiff function with mocked kubectl diff
func TestShowResourceDiff(t *testing.T) {
	// Create temporary files for testing
	tmpDir, err := os.MkdirTemp("", "resource-diff-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original execCommand and restore it after test
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	// Replace execCommand with our mock
	execCommand = mockExecCommand

	// Create mock resources
	currentResource := []byte(`
apiVersion: v1
kind: Service
metadata:
  name: example
spec:
  ports:
  - port: 80
`)

	proposedResource := []byte(`
apiVersion: v1
kind: Service
metadata:
  name: example
spec:
  ports:
  - port: 8080
`)

	// Capture stdout to verify the colorized output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function we're testing
	err = ShowResourceDiff(currentResource, proposedResource, false)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// We expect no error, even though kubectl diff exits with code 1
	if err != nil {
		t.Errorf("ShowResourceDiff returned error: %v", err)
	}

	// Check that the output contains our mocked diff result with colorization
	if !strings.Contains(output, greenColor+"+  added resource"+resetColor) {
		t.Errorf("Expected colorized added resource in output, got: %s", output)
	}

	if !strings.Contains(output, redColor+"-  removed resource"+resetColor) {
		t.Errorf("Expected colorized removed resource in output, got: %s", output)
	}
}
