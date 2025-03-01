package utils

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestSuccess(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	origOutput := Log.Out
	Log.Out = &buf
	defer func() { Log.Out = origOutput }()

	// Test Success function
	testMessage := "Operation successful"
	Success(testMessage)

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
	Green(testMessage)

	logOutput := buf.String()
	if !strings.Contains(logOutput, testMessage) {
		t.Errorf("Green() did not log the message. Got: %s", logOutput)
	}
}

func TestShowResourceDiff(t *testing.T) {
	// Skip actual execution since we can't easily mock the kubectl diff command in this context
	t.Skip("Skipping ShowResourceDiff test - requires mocking exec.Command")
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
