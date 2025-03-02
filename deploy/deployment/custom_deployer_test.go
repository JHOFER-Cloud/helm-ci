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
	"bytes"
	"helm-ci/deploy/config"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUpdateNamespaces(t *testing.T) {
	// Create a test directory
	tmpDir, err := os.MkdirTemp("", "namespace-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test cases
	testCases := []struct {
		name     string
		content  string
		expected map[string]string // Map of paths to expected namespace values
	}{
		{
			name: "update_existing_namespace",
			content: `apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: old-namespace
spec:
  selector:
    app: test
  ports:
  - port: 80
    targetPort: 8080`,
			expected: map[string]string{
				"metadata.namespace": "test-namespace",
			},
		},
		{
			name: "add_missing_namespace",
			content: `apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  selector:
    app: test
  ports:
  - port: 80
    targetPort: 8080`,
			expected: map[string]string{
				"metadata.namespace": "test-namespace",
			},
		},
		{
			name: "multi_document_yaml",
			content: `apiVersion: v1
kind: Service
metadata:
  name: service-1
  namespace: old-namespace-1
spec:
  selector:
    app: test-1
---
apiVersion: v1
kind: Service
metadata:
  name: service-2
spec:
  selector:
    app: test-2`,
			expected: map[string]string{
				"metadata.namespace": "test-namespace", // All resources should have the same namespace
			},
		},
		{
			name: "non_k8s_yaml",
			content: `foo: bar
baz: qux`,
			expected: map[string]string{}, // No changes expected
		},
		{
			name:     "empty_yaml",
			content:  "",
			expected: map[string]string{}, // No changes expected
		},
		{
			name: "yaml_with_comments",
			content: `# This is a test service
apiVersion: v1
kind: Service
metadata:
  name: test-service
  # This is where we set the namespace
  namespace: old-namespace
spec:
  # Service spec follows
  selector:
    app: test
  ports:
  - port: 80
    targetPort: 8080`,
			expected: map[string]string{
				"metadata.namespace": "test-namespace",
			},
		},
		{
			name: "yaml_with_anchors_and_aliases",
			content: `apiVersion: v1
kind: Service
metadata: &metadata
  name: test-service
  namespace: old-namespace
  labels:
    app: test
spec:
  selector: *metadata`,
			expected: map[string]string{
				"metadata.namespace": "test-namespace",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create input file
			inputFile := filepath.Join(tmpDir, tc.name+".yaml")
			if err := os.WriteFile(inputFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Create deployer with test configuration
			deployer := &CustomDeployer{
				Common: Common{
					Config: &config.Config{
						Namespace: "test-namespace",
					},
				},
			}

			// Call the function being tested
			resultFile, err := deployer.updateNamespaces(inputFile)
			if err != nil {
				// We don't expect errors for these test cases
				t.Fatalf("updateNamespaces failed: %v", err)
			}

			// Clean up on return if a new file was created
			if resultFile != inputFile {
				defer os.Remove(resultFile)
			}

			// For non-modified files, verify the original file was returned
			if len(tc.expected) == 0 {
				if resultFile != inputFile {
					t.Errorf("Expected non-modified file to return original path, got different path")
				}
				return
			}

			// For files we expect to be modified, verify they were changed
			if resultFile == inputFile && len(tc.expected) > 0 {
				t.Errorf("Expected file to be modified but got original path")
				return
			}

			// Read the result file
			resultContent, err := os.ReadFile(resultFile)
			if err != nil {
				t.Fatalf("Failed to read result file: %v", err)
			}

			// Verify the namespaces in the YAML content
			verifyNamespaces(t, resultContent, "test-namespace")
		})
	}
}

// Helper function to verify namespaces in YAML content
func verifyNamespaces(t *testing.T, content []byte, expectedNamespace string) {
	t.Helper()

	// Split into documents
	yamlDocs := bytes.Split(content, []byte("---"))

	for _, docBytes := range yamlDocs {
		if len(bytes.TrimSpace(docBytes)) == 0 {
			continue // Skip empty documents
		}

		var doc map[string]interface{}
		err := yaml.Unmarshal(docBytes, &doc)
		if err != nil {
			// Some documents might be invalid YAML or not have expected structure
			continue
		}

		// Check if this document has metadata
		metadata, ok := doc["metadata"].(map[string]interface{})
		if !ok {
			continue
		}

		// Check namespace if the document has metadata
		namespace, ok := metadata["namespace"].(string)
		if !ok {
			t.Errorf("Document with metadata is missing namespace: %v", metadata)
			continue
		}

		if namespace != expectedNamespace {
			t.Errorf("Expected namespace '%s', got '%s'", expectedNamespace, namespace)
		}
	}
}

func TestMalformedYAML(t *testing.T) {
	// Create a test directory
	tmpDir, err := os.MkdirTemp("", "malformed-yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	malformed := `apiVersion: v1
kind: Service
metadata:
  name: test-service
spec:
  - bad indentation
  selector:
    app: test`

	// Create input file
	inputFile := filepath.Join(tmpDir, "malformed.yaml")
	if err := os.WriteFile(inputFile, []byte(malformed), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create deployer with test configuration
	deployer := &CustomDeployer{
		Common: Common{
			Config: &config.Config{
				Namespace: "test-namespace",
			},
		},
	}

	// Call the function being tested - should not error or infinite loop
	resultFile, err := deployer.updateNamespaces(inputFile)
	if err != nil {
		t.Fatalf("updateNamespaces failed on malformed YAML: %v", err)
	}

	// The result should be the original file since we can't process it
	if resultFile != inputFile {
		t.Errorf("Expected original file path for malformed YAML, got different file")
		os.Remove(resultFile) // Clean up the unexpected file
	}
}

func TestDeploy_UpdatesNamespaces(t *testing.T) {
	// Create a test directory structure
	tmpDir, err := os.MkdirTemp("", "deploy-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the stage and common directories
	stageDir := filepath.Join(tmpDir, "dev")
	commonDir := filepath.Join(tmpDir, "common")
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		t.Fatalf("Failed to create stage directory: %v", err)
	}
	if err := os.MkdirAll(commonDir, 0755); err != nil {
		t.Fatalf("Failed to create common directory: %v", err)
	}

	// Create test manifests
	stageManifest := `apiVersion: v1
kind: Service
metadata:
  name: stage-service
  namespace: old-namespace
spec:
  selector:
    app: test`

	commonManifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: common-config
spec:
  data:
    key: value`

	stageFile := filepath.Join(stageDir, "service.yaml")
	commonFile := filepath.Join(commonDir, "config.yaml")

	if err := os.WriteFile(stageFile, []byte(stageManifest), 0644); err != nil {
		t.Fatalf("Failed to write stage manifest: %v", err)
	}
	if err := os.WriteFile(commonFile, []byte(commonManifest), 0644); err != nil {
		t.Fatalf("Failed to write common manifest: %v", err)
	}

	// Create mock commander
	mockCmd := NewMockCommander()

	// Mock the kubectl commands
	// First call to "kubectl get namespace" should succeed
	mockCmd.AddResponse("kubectl:get:namespace", []byte("test-namespace"), nil)

	// Mock the kubectl diff command (used for GetDiff)
	mockCmd.AddResponse("kubectl:diff", []byte("No differences"), nil)

	// Mock the kubectl apply commands
	mockCmd.AddResponse("kubectl:apply", []byte("service/stage-service configured"), nil)
	mockCmd.AddResponse("kubectl:apply", []byte("configmap/common-config created"), nil)

	// Create the custom deployer with our test config
	deployer := &CustomDeployer{
		Common: Common{
			Config: &config.Config{
				ValuesPath: tmpDir,
				Stage:      "dev",
				Namespace:  "test-namespace",
				DEBUG:      false, // No confirmation prompt
			},
			Cmd: mockCmd,
		},
	}

	// Call the Deploy method
	err = deployer.Deploy()
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}

	// Verify kubectl apply was called for each manifest
	applyCount := 0
	for _, cmd := range mockCmd.Commands {
		if cmd.Name == "kubectl" && len(cmd.Args) >= 2 && cmd.Args[0] == "apply" {
			applyCount++

			// Check that namespace flag is used
			hasNamespaceFlag := false
			for i := 0; i < len(cmd.Args)-1; i++ {
				if cmd.Args[i] == "-n" && cmd.Args[i+1] == "test-namespace" {
					hasNamespaceFlag = true
					break
				}
			}

			if !hasNamespaceFlag {
				t.Errorf("kubectl apply command missing namespace flag: %v", cmd.Args)
			}
		}
	}

	if applyCount != 2 {
		t.Errorf("Expected kubectl apply to be called 2 times, got %d", applyCount)
	}
}

// Test deep nesting in YAML
func TestDeepNesting(t *testing.T) {
	// Create a test directory
	tmpDir, err := os.MkdirTemp("", "deep-nesting-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Nested YAML content
	nestedContent := `apiVersion: v1
kind: List
items:
- apiVersion: v1
  kind: Service
  metadata:
    name: nested-service
    namespace: old-namespace
  spec:
    selector:
      app: test
- apiVersion: v1
  kind: ConfigMap
  metadata:
    name: nested-config
  data:
    key: value`

	// Create input file
	inputFile := filepath.Join(tmpDir, "nested.yaml")
	if err := os.WriteFile(inputFile, []byte(nestedContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create deployer with test configuration
	deployer := &CustomDeployer{
		Common: Common{
			Config: &config.Config{
				Namespace: "test-namespace",
			},
		},
	}

	// Call the function being tested
	resultFile, err := deployer.updateNamespaces(inputFile)
	if err != nil {
		t.Fatalf("updateNamespaces failed: %v", err)
	}

	// Clean up on return
	defer func() {
		if resultFile != inputFile {
			os.Remove(resultFile)
		}
	}()

	// This test is mainly to verify that the nested YAML doesn't cause errors
	// The namespace changes would be in the items array, which we don't currently traverse
}
