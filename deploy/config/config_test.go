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

package config

import (
	"bytes"
	"flag"
	"helm-ci/deploy/utils"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPrintConfig_SensitiveFields(t *testing.T) {
	// Create a config with sensitive and non-sensitive values
	cfg := &Config{
		AppName:         "test-app",
		Stage:           "dev",
		Environment:     "test",
		GitHubToken:     "github-secret-token",
		VaultToken:      "vault-secret-token",
		VaultURL:        "https://vault.example.com",
		CustomNameSpace: "custom-namespace",
	}

	// Capture log output
	var buf bytes.Buffer
	origLogOut := utils.Log.Out
	utils.Log.SetOutput(&buf)

	// Call PrintConfig
	cfg.PrintConfig()

	// Restore logger output
	utils.Log.SetOutput(origLogOut)

	output := buf.String()

	// Check that sensitive fields are redacted
	if strings.Contains(output, "github-secret-token") {
		t.Error("GitHubToken was not redacted in the output")
	}
	if strings.Contains(output, "vault-secret-token") {
		t.Error("VaultToken was not redacted in the output")
	}

	// Check that [REDACTED] appears for sensitive fields
	if !strings.Contains(output, "GitHubToken: [REDACTED]") {
		t.Error("Expected 'GitHubToken: [REDACTED]' in output but didn't find it")
	}
	if !strings.Contains(output, "VaultToken: [REDACTED]") {
		t.Error("Expected 'VaultToken: [REDACTED]' in output but didn't find it")
	}

	// Check that non-sensitive fields are displayed normally
	if !strings.Contains(output, "AppName: test-app") {
		t.Error("Expected 'AppName: test-app' in output but didn't find it")
	}
	if !strings.Contains(output, "VaultURL: https://vault.example.com") {
		t.Error("Expected 'VaultURL: https://vault.example.com' in output but didn't find it")
	}
}

func TestConfig_SetupNames_EdgeCases(t *testing.T) {
	// Test cases for edge cases
	testCases := []struct {
		name            string
		config          *Config
		expectedNS      string
		expectedRelease string
		expectedHosts   []string
	}{
		{
			name: "empty app name",
			config: &Config{
				AppName: "",
				Stage:   "dev",
				Domains: []string{"example.com"},
			},
			expectedNS:      "-dev",                   // Edge case: empty app name leads to weird namespace
			expectedRelease: "",                       // Empty app name
			expectedHosts:   []string{".example.com"}, // Edge case: empty app name in host
		},
		{
			name: "special characters in app name",
			config: &Config{
				AppName: "test-app_special!",
				Stage:   "dev",
				Domains: []string{"example.com"},
			},
			expectedNS:      "test-app_special!-dev",
			expectedRelease: "test-app_special!",
			expectedHosts:   []string{"test-app_special!.example.com"},
		},
		{
			name: "PR number with special characters",
			config: &Config{
				AppName:       "test-app",
				Stage:         "dev",
				PRNumber:      "42-bugfix",
				Domains:       []string{"dev.example.com"},
				PRDeployments: true,
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app-pr-42-bugfix",
			expectedHosts:   []string{"test-app-pr-42-bugfix.dev.example.com"},
		},
		{
			name: "empty domains",
			config: &Config{
				AppName:  "test-app",
				Stage:    "dev",
				PRNumber: "42",
				Domains:  []string{}, // Empty domains
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app", // Default since no PR deployments with no domains
			expectedHosts:   []string{}, // Empty with no domains
		},
		{
			name: "empty stage",
			config: &Config{
				AppName:  "test-app",
				Stage:    "", // Empty stage
				Domains:  []string{"example.com"},
				PRNumber: "42",
			},
			expectedNS:      "test-app-dev", // Default behavior without stage
			expectedRelease: "test-app",     // Default release name
			expectedHosts:   []string{"test-app.example.com"},
		},
		{
			name: "multiple domains",
			config: &Config{
				AppName: "test-app",
				Stage:   "dev",
				Domains: []string{"example.com", "example.org", "example.net"},
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app",
			expectedHosts: []string{
				"test-app.example.com",
				"test-app.example.org",
				"test-app.example.net",
			},
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

			// Compare IngressHosts arrays
			if !reflect.DeepEqual(tc.config.IngressHosts, tc.expectedHosts) {
				t.Errorf("Expected IngressHosts %v, got %v", tc.expectedHosts, tc.config.IngressHosts)
			}
		})
	}
}

func TestParseFlags_DefaultValues(t *testing.T) {
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

	// Clear environment variables before testing defaults
	for _, env := range []string{"GITHUB_TOKEN", "VAULT_TOKEN"} {
		os.Unsetenv(env)
	}

	// Set up minimal required flags
	os.Args = []string{
		"cmd",
		"--stage=dev",
		"--app=test-app",
		"--env=Development",
	}

	cfg := ParseFlags()

	// Check default values for flags that weren't specified
	defaultChecks := []struct {
		fieldName string
		expected  interface{}
	}{
		{"ValuesPath", "helm/values"},
		{"PRDeployments", true},
		{"VaultKVVersion", 2},
		{"GitHubToken", ""}, // No env var
		{"VaultToken", ""},  // No env var
		{"Chart", ""},
		{"Version", ""},
		{"Repository", ""},
		{"Domains", []string(nil)},
		{"CustomNameSpace", ""},
		{"Custom", false},
		{"TraefikDashboard", false},
		{"RootCA", ""},
		{"VaultURL", ""},
		{"VaultBasePath", ""},
		{"VaultInsecureTLS", false},
		{"DEBUG", false},
		{"DomainsTemplate", "default"}, // Added default domain template
	}

	for _, check := range defaultChecks {
		field := reflect.ValueOf(cfg).Elem().FieldByName(check.fieldName)
		if !field.IsValid() {
			t.Errorf("Field %s not found in Config struct", check.fieldName)
			continue
		}

		actual := field.Interface()
		if !reflect.DeepEqual(actual, check.expected) {
			t.Errorf("Expected default %s to be %v, got %v",
				check.fieldName, check.expected, actual)
		}
	}
}

func TestConfig_SetupNames_Combinations(t *testing.T) {
	// Test different combinations that might be confusing
	testCases := []struct {
		name            string
		config          *Config
		expectedNS      string
		expectedRelease string
		expectedHosts   []string
	}{
		{
			name: "custom namespace with PR number",
			config: &Config{
				AppName:         "test-app",
				Stage:           "dev",
				PRNumber:        "42",
				Domains:         []string{"dev.example.com"},
				PRDeployments:   true,
				CustomNameSpace: "custom-ns",
			},
			expectedNS:      "custom-ns", // Custom namespace takes precedence
			expectedRelease: "test-app-pr-42",
			expectedHosts:   []string{"test-app-pr-42.dev.example.com"},
		},
		{
			name: "live stage with custom namespace and PR",
			config: &Config{
				AppName:         "test-app",
				Stage:           "live", // Live environment
				PRNumber:        "42",   // With PR number
				Domains:         []string{"example.com"},
				PRDeployments:   true,
				CustomNameSpace: "custom-live", // And custom namespace
			},
			expectedNS:      "custom-live",                    // Custom namespace takes precedence over live
			expectedRelease: "test-app",                       // Live stage doesn't use PR in release name
			expectedHosts:   []string{"test-app.example.com"}, // Live stage doesn't use PR in host
		},
		{
			name: "live stage with PR deployments enabled",
			config: &Config{
				AppName:       "test-app",
				Stage:         "live",
				PRNumber:      "42",
				Domains:       []string{"example.com"},
				PRDeployments: true, // Enabled but should be ignored for live
			},
			expectedNS:      "test-app",                       // Live namespace
			expectedRelease: "test-app",                       // Live release (PR ignored in live)
			expectedHosts:   []string{"test-app.example.com"}, // Live host (PR ignored)
		},
		{
			name: "multiple domains with PR",
			config: &Config{
				AppName:       "test-app",
				Stage:         "dev",
				PRNumber:      "42",
				Domains:       []string{"dev1.example.com", "dev2.example.com"},
				PRDeployments: true,
			},
			expectedNS:      "test-app-dev",
			expectedRelease: "test-app-pr-42",
			expectedHosts: []string{
				"test-app-pr-42.dev1.example.com",
				"test-app-pr-42.dev2.example.com",
			},
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

			// Compare IngressHosts arrays
			if !reflect.DeepEqual(tc.config.IngressHosts, tc.expectedHosts) {
				t.Errorf("Expected IngressHosts %v, got %v", tc.expectedHosts, tc.config.IngressHosts)
			}
		})
	}
}
