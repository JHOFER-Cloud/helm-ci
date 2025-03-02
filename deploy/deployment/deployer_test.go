package deployment

import (
	"errors"
	"helm-ci/deploy/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func normalizeString(s string) string {
	// Normalize whitespace and line endings for comparison
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func TestExtractYAMLContent_MultipleManifests(t *testing.T) {
	common := Common{
		Config: &config.Config{},
		Cmd:    &RealCommander{},
	}

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

	expected := `---
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

	// Normalize the strings for comparison
	normalizedYaml := normalizeString(string(yaml))
	normalizedExpected := normalizeString(expected)

	if normalizedYaml != normalizedExpected {
		t.Errorf("ExtractYAMLContent with multiple manifests doesn't match expected.\nGot:\n%s\n\nExpected:\n%s",
			string(yaml), expected)
	}
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
		Cmd: &RealCommander{},
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

// NEW TESTS BELOW

func TestProcessValuesFileWithVault_FileReadError(t *testing.T) {
	// Create a Common instance with no Vault config
	common := Common{
		Config: &config.Config{
			VaultURL: "https://vault.example.com", // Set this so we don't return early
		},
		Cmd: &RealCommander{},
	}

	// Try to process a non-existent file
	nonExistentFile := "/path/to/nonexistent/file.yaml"
	_, err := common.ProcessValuesFileWithVault(nonExistentFile)

	// Expect an error for the non-existent file
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to read values file") {
		t.Errorf("Expected error message to contain 'failed to read values file', got: %v", err)
	}
}

func TestSetupRootCA_FileReadError(t *testing.T) {
	// Create a config with invalid root CA path
	cfg := &config.Config{
		RootCA:    "/path/to/nonexistent/ca.crt",
		Namespace: "test-namespace",
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Create a Common instance with the mock commander
	common := Common{
		Config: cfg,
		Cmd:    mockCmd,
	}

	// Call SetupRootCA
	err := common.SetupRootCA()

	// Expect an error for the non-existent CA file
	if err == nil {
		t.Errorf("Expected error for non-existent CA file, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to read root CA file") {
		t.Errorf("Expected error message to contain 'failed to read root CA file', got: %v", err)
	}
}

func TestSetupRootCA_CommandErrors(t *testing.T) {
	// Create a temporary CA file
	tmpDir, err := os.MkdirTemp("", "root-ca-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	caFile := filepath.Join(tmpDir, "ca.crt")
	if err := os.WriteFile(caFile, []byte("TEST CA CERTIFICATE"), 0644); err != nil {
		t.Fatalf("Failed to write test CA file: %v", err)
	}

	// Create a config with the test CA file
	cfg := &config.Config{
		RootCA:    caFile,
		Namespace: "test-namespace",
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Make the first command (create namespace yaml) fail
	mockCmd.AddResponse("kubectl:create", nil, errors.New("namespace creation failed"))

	// Create a Common instance with the mock commander
	common := Common{
		Config: cfg,
		Cmd:    mockCmd,
	}

	// Call SetupRootCA
	err = common.SetupRootCA()

	// Expect an error
	if err == nil {
		t.Errorf("Expected error when kubectl command fails, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to create namespace yaml") {
		t.Errorf("Expected error message to contain 'failed to create namespace yaml', got: %v", err)
	}
}

func TestHelmDeployer_Deploy_RepoAddError(t *testing.T) {
	// Create a config for helm deployment
	cfg := &config.Config{
		AppName:     "test-app",
		Chart:       "test-chart",
		ReleaseName: "test-release",
		Namespace:   "test-namespace",
		Repository:  "https://charts.example.com", // non-OCI repo
		Domains:     []string{"example.com"},
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Make the helm repo add command fail
	mockCmd.AddResponse("helm:repo:add", nil, errors.New("failed to add repo"))

	// Create a HelmDeployer with the mock commander
	deployer := &HelmDeployer{
		Common: Common{
			Config: cfg,
			Cmd:    mockCmd,
		},
	}

	// Call Deploy
	err := deployer.Deploy()

	// Expect an error
	if err == nil {
		t.Fatalf("Expected error when helm repo add fails, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to add Helm repository") {
		t.Errorf("Expected error message to contain 'failed to add Helm repository', got: %v", err)
	}
}

func TestHelmDeployer_Deploy_RepoUpdateError(t *testing.T) {
	// Create a config for helm deployment
	cfg := &config.Config{
		AppName:     "test-app",
		Chart:       "test-chart",
		ReleaseName: "test-release",
		Namespace:   "test-namespace",
		Repository:  "https://charts.example.com", // non-OCI repo
		Domains:     []string{"example.com"},
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Make repo add succeed but repo update fail
	// The key pattern is "helm:repo" for repo add, and we need to match the second argument
	mockCmd.AddResponse("helm:repo:add", []byte("repository added"), nil)
	mockCmd.AddResponse("helm:repo:update", nil, errors.New("failed to update repo"))

	// Create a HelmDeployer with the mock commander
	deployer := &HelmDeployer{
		Common: Common{
			Config: cfg,
			Cmd:    mockCmd,
		},
	}

	// Call Deploy
	err := deployer.Deploy()

	// Expect an error
	if err == nil {
		t.Fatalf("Expected error when helm repo update fails, got nil")
	}

	// Now that we've checked err is not nil, we can safely call Error()
	if !strings.Contains(err.Error(), "failed to update Helm repository") {
		t.Errorf("Expected error message to contain 'failed to update Helm repository', got: %v", err)
	}
}

func TestHelmDeployer_Deploy_OciRepository(t *testing.T) {
	// Create a config with OCI repository
	cfg := &config.Config{
		AppName:     "test-app",
		Chart:       "test-chart",
		ReleaseName: "test-release",
		Namespace:   "test-namespace",
		Repository:  "oci://registry.example.com", // OCI repo
		Domains:     []string{"example.com"},
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// For OCI repos, we skip repo add/update and go straight to get manifest
	mockCmd.AddResponse("helm:get:manifest", []byte(""), errors.New("release not found"))
	mockCmd.AddResponse("helm:upgrade", []byte("release deployed"), nil)

	// Create a HelmDeployer with the mock commander
	deployer := &HelmDeployer{
		Common: Common{
			Config: cfg,
			Cmd:    mockCmd,
		},
	}

	// Call Deploy
	err := deployer.Deploy()
	// Should not error
	if err != nil {
		t.Errorf("Unexpected error with OCI repository: %v", err)
	}

	// Verify we didn't try to add/update the repository
	for _, cmd := range mockCmd.Commands {
		if cmd.Name == "helm" && len(cmd.Args) > 0 && cmd.Args[0] == "repo" {
			t.Errorf("Should not call helm repo commands with OCI repository, but called: %s %v", cmd.Name, cmd.Args)
		}
	}

	// Verify the helm upgrade command uses the OCI repository format
	ociFormatFound := false
	expectedChartPath := "oci://registry.example.com/test-chart"

	for _, cmd := range mockCmd.Commands {
		if cmd.Name == "helm" && len(cmd.Args) > 2 && cmd.Args[0] == "upgrade" {
			for _, arg := range cmd.Args {
				if arg == expectedChartPath {
					ociFormatFound = true
					break
				}
			}
		}
	}

	if !ociFormatFound {
		t.Errorf("Did not find expected OCI chart reference %q in commands", expectedChartPath)
	}
}

func TestCustomDeployer_Deploy_NoManifests(t *testing.T) {
	// Create a temporary directory with no manifest files
	tmpDir, err := os.MkdirTemp("", "custom-test-empty")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config with the empty directory
	cfg := &config.Config{
		AppName:    "test-app",
		Namespace:  "test-namespace",
		ValuesPath: tmpDir, // Directory with no manifests
	}

	// Create a mock commander instead of a real one
	mockCmd := NewMockCommander()

	// Add mock responses for the namespace check (which runs even with no manifests)
	mockCmd.AddResponse("kubectl:get:namespace", nil, errors.New("not found"))
	mockCmd.AddResponse("kubectl:create:namespace", []byte("namespace created"), nil)

	// Create a CustomDeployer with the mock commander
	deployer := &CustomDeployer{
		Common: Common{
			Config: cfg,
			Cmd:    mockCmd,
		},
	}

	// Call Deploy
	err = deployer.Deploy()
	// Should not error
	if err != nil {
		t.Errorf("Unexpected error with empty manifests directory: %v", err)
	}

	// Verify that kubectl commands were called with expected arguments
	foundNamespaceCheck := false
	for _, cmd := range mockCmd.Commands {
		if cmd.Name == "kubectl" && len(cmd.Args) >= 3 &&
			cmd.Args[0] == "get" && cmd.Args[1] == "namespace" && cmd.Args[2] == "test-namespace" {
			foundNamespaceCheck = true
			break
		}
	}

	if !foundNamespaceCheck {
		t.Errorf("Expected kubectl get namespace command to be called")
	}
}

func TestGetDiff_KubectlDiffError(t *testing.T) {
	// Create a config
	cfg := &config.Config{
		Namespace: "test-namespace",
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Make kubectl diff return error other than exit code 1
	exitErr := &ExitError{
		Err:       errors.New("kubectl diff failed"),
		CodeValue: 2, // Something other than 1, which is expected for diffs
	}
	mockCmd.AddResponse("kubectl:diff", []byte("error output"), exitErr)

	// Create a Common instance
	common := Common{
		Config: cfg,
		Cmd:    mockCmd,
	}

	// Call GetDiff with kubectl (isHelm=false)
	err := common.GetDiff([]string{"manifest.yml"}, false)

	// Expect an error
	if err == nil {
		t.Errorf("Expected error when kubectl diff fails with non-1 exit code, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to get diff") {
		t.Errorf("Expected error message to contain 'failed to get diff', got: %v", err)
	}
}

func TestGetDiff_HelmDryRunError(t *testing.T) {
	// Create a config
	cfg := &config.Config{
		ReleaseName: "test-release",
		Namespace:   "test-namespace",
	}

	// Create a mock commander
	mockCmd := NewMockCommander()

	// Make helm get manifest fail (expected)
	mockCmd.AddResponse("helm:get", []byte(""), errors.New("release not found"))

	// Make helm dry-run fail
	exitErr := &ExitError{
		Err:       errors.New("helm dry-run failed"),
		CodeValue: 1,
	}
	mockCmd.AddResponse("helm:upgrade", []byte("error output"), exitErr)

	// Create a Common instance
	common := Common{
		Config: cfg,
		Cmd:    mockCmd,
	}

	// Call GetDiff with helm
	err := common.GetDiff([]string{"upgrade", "--install", "test-release"}, true)

	// Expect an error
	if err == nil {
		t.Errorf("Expected error when helm dry-run fails, got nil")
	}

	// Verify the error message contains expected text
	if !strings.Contains(err.Error(), "failed to get proposed state") {
		t.Errorf("Expected error message to contain 'failed to get proposed state', got: %v", err)
	}
}

func TestHelmDeployer_GetRootCAArgs(t *testing.T) {
	// Create a HelmDeployer instance with a test config with root CA
	helmDeployer := HelmDeployer{
		Common: Common{
			Config: &config.Config{
				RootCA: "/path/to/ca.crt",
			},
			Cmd: &RealCommander{},
		},
	}

	// Test the GetRootCAArgs method
	args := helmDeployer.GetRootCAArgs()

	// Verify the return value
	if len(args) != 0 {
		t.Errorf("Expected empty args as implementation is commented out, got %v", args)
	}
}
