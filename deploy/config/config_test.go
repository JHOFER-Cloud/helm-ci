package config

import (
	"bytes"
	"flag"
	"helm-ci/deploy/utils"
	"os"
	"strings"
	"testing"
)

func TestSetupNames(t *testing.T) {
	testCases := []struct {
		name            string
		config          *Config
		expectedNS      string
		expectedRelease string
		expectedHost    string
	}{
		{
			name: "live environment",
			config: &Config{
				AppName: "test-app",
				Stage:   "live",
				Domain:  "example.com",
			},
			expectedNS:      "test-app",
			expectedRelease: "test-app",
			expectedHost:    "test-app.example.com",
		},
		{
			name: "dev environment",
			config: &Config{
				AppName: "test-app",
				Stage:   "dev",
				Domain:  "dev.example.com",
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app",
			expectedHost:    "test-app.dev.example.com",
		},
		{
			name: "PR deployment",
			config: &Config{
				AppName:       "test-app",
				Stage:         "dev",
				PRNumber:      "42",
				Domain:        "dev.example.com",
				PRDeployments: true,
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app-pr-42",
			expectedHost:    "test-app-pr-42.dev.example.com",
		},
		{
			name: "custom namespace",
			config: &Config{
				AppName:         "test-app",
				Stage:           "dev",
				CustomNameSpace: "custom-namespace",
				Domain:          "dev.example.com",
			},
			expectedNS:      "custom-namespace",
			expectedRelease: "test-app",
			expectedHost:    "test-app.dev.example.com",
		},
		{
			name: "PR but deployments disabled",
			config: &Config{
				AppName:       "test-app",
				Stage:         "dev",
				PRNumber:      "42",
				Domain:        "dev.example.com",
				PRDeployments: false,
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app",
			expectedHost:    "test-app.dev.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.config.SetupNames()

			if tc.config.Namespace != tc.expectedNS {
				t.Errorf("Expected Namespace %q, got %q", tc.expectedNS, tc.config.Namespace)
			}
			if tc.config.ReleaseName != tc.expectedRelease {
				t.Errorf("Expected ReleaseName %q, got %q", tc.expectedRelease, tc.config.ReleaseName)
			}
			if tc.config.IngressHost != tc.expectedHost {
				t.Errorf("Expected IngressHost %q, got %q", tc.expectedHost, tc.config.IngressHost)
			}
		})
	}
}

// Testing ParseFlags is more complex because it involves flag parsing and os.Exit
// Let's add a simplified version that tests parts of it
func TestPrintConfig(t *testing.T) {
	// Create a config with some values
	cfg := &Config{
		AppName:        "test-app",
		Stage:          "dev",
		Environment:    "test",
		GitHubToken:    "secret-token",
		VaultToken:     "vault-secret",
		PRDeployments:  true,
		VaultKVVersion: 2,
	}

	// The issue is that PrintConfig uses utils.Log.Info, not fmt.Println
	// So we need to capture the logger output, not stdout
	var buf bytes.Buffer
	origLogOut := utils.Log.Out
	utils.Log.SetOutput(&buf)

	// Call PrintConfig
	cfg.PrintConfig()

	// Restore logger output
	utils.Log.SetOutput(origLogOut)

	output := buf.String()
	if output == "" {
		t.Error("Expected PrintConfig to produce output but got nothing")
	}

	// Check that sensitive info is redacted
	if strings.Contains(output, "secret-token") || strings.Contains(output, "vault-secret") {
		t.Error("Sensitive information was not redacted in the output")
	}

	// Check that [REDACTED] appears in the output for sensitive fields
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Expected [REDACTED] marker but didn't find it")
	}
}

func TestParseFlags(t *testing.T) {
	// Save original flag.CommandLine and restore it after the test
	origFlagCommandLine := flag.CommandLine
	defer func() {
		flag.CommandLine = origFlagCommandLine
	}()

	// Reset flag.CommandLine for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original os.Args and restore it after the test
	origArgs := os.Args
	defer func() {
		os.Args = origArgs
	}()

	// Set up the environment for testing
	os.Setenv("GITHUB_TOKEN", "test-token")
	os.Setenv("VAULT_TOKEN", "vault-token")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("VAULT_TOKEN")
	}()

	// We can't easily test the case where required flags are missing because ParseFlags calls os.Exit
	// Instead, we'll test with all required flags provided

	// Test case: Minimum required flags
	os.Args = []string{
		"cmd",
		"--stage=dev",
		"--app=test-app",
		"--env=Development",
	}

	cfg := ParseFlags()

	// Validate the parsed configuration
	if cfg.Stage != "dev" {
		t.Errorf("Expected Stage to be 'dev', got '%s'", cfg.Stage)
	}
	if cfg.AppName != "test-app" {
		t.Errorf("Expected AppName to be 'test-app', got '%s'", cfg.AppName)
	}
	if cfg.Environment != "Development" {
		t.Errorf("Expected Environment to be 'Development', got '%s'", cfg.Environment)
	}
	if cfg.ValuesPath != "helm/values" {
		t.Errorf("Expected default ValuesPath to be 'helm/values', got '%s'", cfg.ValuesPath)
	}
	if cfg.PRDeployments != true {
		t.Errorf("Expected default PRDeployments to be true, got %v", cfg.PRDeployments)
	}
	if cfg.VaultKVVersion != 2 {
		t.Errorf("Expected default VaultKVVersion to be 2, got %d", cfg.VaultKVVersion)
	}
	if cfg.GitHubToken != "test-token" {
		t.Errorf("Expected GitHubToken to be 'test-token', got '%s'", cfg.GitHubToken)
	}
	if cfg.VaultToken != "vault-token" {
		t.Errorf("Expected VaultToken to be 'vault-token', got '%s'", cfg.VaultToken)
	}

	// Test case: All flags provided
	os.Args = []string{
		"cmd",
		"--stage=live",
		"--app=full-app",
		"--env=Production",
		"--pr=42",
		"--values=custom/values",
		"--chart=my-chart",
		"--version=1.0.0",
		"--repo=https://charts.example.com",
		"--github-repo=my-repo",
		"--github-owner=my-owner",
		"--domain=example.com",
		"--custom-namespace=custom-ns",
		"--custom=true",
		"--traefik-dashboard=true",
		"--root-ca=/path/to/ca.crt",
		"--pr-deployments=false",
		"--vault-url=https://vault.example.com",
		"--vault-base-path=secret",
		"--vault-insecure-tls=true",
		"--vault-kv-version=1",
		"--debug=true",
	}

	// Reset flag.CommandLine for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg = ParseFlags()

	// Validate the parsed configuration
	if cfg.Stage != "live" {
		t.Errorf("Expected Stage to be 'live', got '%s'", cfg.Stage)
	}
	if cfg.AppName != "full-app" {
		t.Errorf("Expected AppName to be 'full-app', got '%s'", cfg.AppName)
	}
	if cfg.Environment != "Production" {
		t.Errorf("Expected Environment to be 'Production', got '%s'", cfg.Environment)
	}
	if cfg.PRNumber != "42" {
		t.Errorf("Expected PRNumber to be '42', got '%s'", cfg.PRNumber)
	}
	if cfg.ValuesPath != "custom/values" {
		t.Errorf("Expected ValuesPath to be 'custom/values', got '%s'", cfg.ValuesPath)
	}
	if cfg.Chart != "my-chart" {
		t.Errorf("Expected Chart to be 'my-chart', got '%s'", cfg.Chart)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", cfg.Version)
	}
	if cfg.Repository != "https://charts.example.com" {
		t.Errorf("Expected Repository to be 'https://charts.example.com', got '%s'", cfg.Repository)
	}
	if cfg.GitHubRepo != "my-repo" {
		t.Errorf("Expected GitHubRepo to be 'my-repo', got '%s'", cfg.GitHubRepo)
	}
	if cfg.GitHubOwner != "my-owner" {
		t.Errorf("Expected GitHubOwner to be 'my-owner', got '%s'", cfg.GitHubOwner)
	}
	if cfg.Domain != "example.com" {
		t.Errorf("Expected Domain to be 'example.com', got '%s'", cfg.Domain)
	}
	if cfg.CustomNameSpace != "custom-ns" {
		t.Errorf("Expected CustomNameSpace to be 'custom-ns', got '%s'", cfg.CustomNameSpace)
	}
	if cfg.Custom != true {
		t.Errorf("Expected Custom to be true, got %v", cfg.Custom)
	}
	if cfg.TraefikDashboard != true {
		t.Errorf("Expected TraefikDashboard to be true, got %v", cfg.TraefikDashboard)
	}
	if cfg.RootCA != "/path/to/ca.crt" {
		t.Errorf("Expected RootCA to be '/path/to/ca.crt', got '%s'", cfg.RootCA)
	}
	if cfg.PRDeployments != false {
		t.Errorf("Expected PRDeployments to be false, got %v", cfg.PRDeployments)
	}
	if cfg.VaultURL != "https://vault.example.com" {
		t.Errorf("Expected VaultURL to be 'https://vault.example.com', got '%s'", cfg.VaultURL)
	}
	if cfg.VaultBasePath != "secret" {
		t.Errorf("Expected VaultBasePath to be 'secret', got '%s'", cfg.VaultBasePath)
	}
	if cfg.VaultInsecureTLS != true {
		t.Errorf("Expected VaultInsecureTLS to be true, got %v", cfg.VaultInsecureTLS)
	}
	if cfg.VaultKVVersion != 1 {
		t.Errorf("Expected VaultKVVersion to be 1, got %d", cfg.VaultKVVersion)
	}
	if cfg.DEBUG != true {
		t.Errorf("Expected DEBUG to be true, got %v", cfg.DEBUG)
	}
}
