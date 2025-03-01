package deployment

import (
	"helm-ci/deploy/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelperProcess isn't a real test. It's used as a helper process for mocking
// exec functionality
func TestHelperProcess(t *testing.T) {
	// Implementation remains the same
}

func TestExtractYAMLContent_MultipleManifests(t *testing.T) {
	common := Common{}

	helmOutput := `
MANIFEST:
---
# Source: chart/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-1
spec:
  ports:
  - port: 80
---
# Source: chart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-2
spec:
  replicas: 2
***
NOTES:
Thank you for installing the chart!
`

	expected := `
---
# Source: chart/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: example-1
spec:
  ports:
  - port: 80
---
# Source: chart/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-2
spec:
  replicas: 2`

	yaml, err := common.ExtractYAMLContent([]byte(helmOutput))
	if err != nil {
		t.Fatalf("ExtractYAMLContent failed: %v", err)
	}

	// Normalize the strings for comparison by trimming all lines
	normalizedYaml := normalizeWhitespace(string(yaml))
	normalizedExpected := normalizeWhitespace(expected)

	if normalizedYaml != normalizedExpected {
		t.Errorf("ExtractYAMLContent with multiple manifests doesn't match expected.\nGot:\n%s\n\nExpected:\n%s",
			string(yaml), expected)
	}
}

func TestHelmDeployer_GetRootCAArgs(t *testing.T) {
	// Create a HelmDeployer instance with a test config
	helmDeployer := HelmDeployer{
		Common: Common{
			Config: &config.Config{
				RootCA: "/path/to/ca.crt",
			},
		},
	}

	// Test the GetRootCAArgs method
	args := helmDeployer.GetRootCAArgs()

	// Note: Current implementation returns an empty array as the code in the method is commented out
	if len(args) != 0 {
		t.Errorf("Expected empty args, got %v", args)
	}
}

// Helper function to normalize whitespace for comparison
func normalizeWhitespace(s string) string {
	// Split by lines, trim each line, and rejoin
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	// Join with a standard separator and trim the whole string
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func TestProcessValuesFileWithVault_NoVault(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "test-values")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test values file
	valuesContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  config: simple-value
`
	valuesFile := filepath.Join(tmpDir, "values.yaml")
	if err := os.WriteFile(valuesFile, []byte(valuesContent), 0644); err != nil {
		t.Fatalf("Failed to write test values file: %v", err)
	}

	// Create a Common instance with no Vault config
	common := Common{
		Config: &config.Config{
			// VaultURL is intentionally empty
		},
	}

	// Process the file
	processedFile, err := common.ProcessValuesFileWithVault(valuesFile)
	if err != nil {
		t.Fatalf("ProcessValuesFileWithVault failed: %v", err)
	}

	// With no Vault URL configured, the original file should be returned
	if processedFile != valuesFile {
		t.Errorf("Expected the original file path to be returned, got %q", processedFile)
		// Clean up any temporary file created
		os.Remove(processedFile)
	}
}

func TestExtractYAMLContent_EmptyInput(t *testing.T) {
	common := Common{}
	yaml, err := common.ExtractYAMLContent([]byte(""))
	if err != nil {
		t.Fatalf("ExtractYAMLContent failed with empty input: %v", err)
	}

	if len(yaml) != 0 {
		t.Errorf("Expected empty output for empty input, got: %s", string(yaml))
	}
}

func TestExtractYAMLContent_NoManifestMarker(t *testing.T) {
	common := Common{}
	input := `This is some text
without a MANIFEST: marker
but with some other content`

	yaml, err := common.ExtractYAMLContent([]byte(input))
	if err != nil {
		t.Fatalf("ExtractYAMLContent failed with input without manifest marker: %v", err)
	}

	if len(yaml) != 0 {
		t.Errorf("Expected empty output for input without manifest marker, got: %s", string(yaml))
	}
}

func TestHelmDeployer_GetTraefikDashboardArgs(t *testing.T) {
	testCases := []struct {
		name     string
		config   *config.Config
		expected []string
	}{
		{
			name: "traefik dashboard enabled",
			config: &config.Config{
				TraefikDashboard: true,
				IngressHost:      "dashboard.example.com",
			},
			expected: []string{
				"--set", "ingressRoute.dashboard.matchRule=Host(`dashboard.example.com`)",
				"--set", "ingressRoute.dashboard.entryPoints[0]=websecure",
			},
		},
		{
			name: "traefik dashboard disabled",
			config: &config.Config{
				TraefikDashboard: false,
				IngressHost:      "dashboard.example.com",
			},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deployer := HelmDeployer{
				Common: Common{
					Config: tc.config,
				},
			}

			args := deployer.GetTraefikDashboardArgs()

			// Compare args length
			if len(args) != len(tc.expected) {
				t.Fatalf("Expected %d args, got %d", len(tc.expected), len(args))
			}

			// Compare each arg
			for i, arg := range args {
				if arg != tc.expected[i] {
					t.Errorf("Arg %d: expected %q, got %q", i, tc.expected[i], arg)
				}
			}
		})
	}
}
