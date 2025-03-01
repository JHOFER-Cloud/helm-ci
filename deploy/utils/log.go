package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

// Make exec.Command mockable for testing
var execCommand = exec.Command

const (
	checkMark   = "✓"
	greenColor  = "\033[32m"
	redColor    = "\033[31m"
	resetColor  = "\033[0m"
	yellowColor = "\033[33m"
)

func init() {
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		PadLevelText:           false,
	})

	// Default to Info level
	Log.SetLevel(logrus.InfoLevel)
}

// InitLogger should be called from main after config is loaded
func InitLogger(DEBUG bool) {
	if os.Getenv("DEBUG") != "" || DEBUG {
		Log.SetLevel(logrus.DebugLevel)
	}
}

func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Log.Info(greenColor + checkMark + " " + msg + resetColor)
}

func Green(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Log.Info(greenColor + " " + msg + resetColor)
}

func WrapError(err error, format string, args ...interface{}) error {
	wrappedErr := fmt.Errorf(format+": %w", append(args, err)...)
	Log.Error(wrappedErr)
	return wrappedErr
}

func NewError(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	Log.Error(err)
	return err
}

func ColorizeKubectlDiff(diffOutput string) string {
	lines := strings.Split(diffOutput, "\n")
	var colorized []string

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			colorized = append(colorized, greenColor+line+resetColor)
		case strings.HasPrefix(line, "-"):
			colorized = append(colorized, redColor+line+resetColor)
		case strings.HasPrefix(line, "~"):
			colorized = append(colorized, yellowColor+line+resetColor)
		default:
			colorized = append(colorized, line)
		}
	}

	return strings.Join(colorized, "\n")
}

func ShowResourceDiff(current, proposed []byte, debug bool) error {
	currentFile, err := os.CreateTemp("", "current-*.yaml")
	if err != nil {
		return NewError("failed to create temp file: %v", err)
	}
	defer os.Remove(currentFile.Name())

	proposedFile, err := os.CreateTemp("", "proposed-*.yaml")
	if err != nil {
		return NewError("failed to create temp file: %v", err)
	}
	defer os.Remove(proposedFile.Name())

	if err := os.WriteFile(currentFile.Name(), current, 0644); err != nil {
		return NewError("failed to write current state: %v", err)
	}
	if err := os.WriteFile(proposedFile.Name(), proposed, 0644); err != nil {
		return NewError("failed to write proposed state: %v", err)
	}

	if debug {
		Log.Debugln("Current YAML:")
		fmt.Println(string(current))

		Log.Debugln("Proposed YAML:")
		fmt.Println(string(proposed))
	}

	// Use the mockable execCommand instead of exec.Command directly
	diffCmd := execCommand("kubectl", "diff", "-f", currentFile.Name(), "-f", proposedFile.Name())
	output, err := diffCmd.CombinedOutput()

	if len(output) > 0 {
		fmt.Println(ColorizeKubectlDiff(string(output)))
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			return NewError("failed to generate diff: %v", err)
		}
	}

	return nil
}
