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

package templates

import (
	"helm-ci/deploy/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessDomainTemplate(t *testing.T) {
	// Test cases for embedded templates and custom files
	testCases := []struct {
		name         string
		config       *config.Config
		expectError  bool
		checkContent string
	}{
		{
			name: "embedded default template",
			config: &config.Config{
				DomainTemplate: "default",
				IngressHosts:    []string{"example.com", "www.example.com"},
			},
			expectError:  false,
			checkContent: "host: example.com",
		},
		{
			name: "embedded bitnami template",
			config: &config.Config{
				DomainTemplate: "bitnami",
				IngressHosts:    []string{"example.com", "www.example.com"},
			},
			expectError:  false,
			checkContent: "hostname: example.com",
		},
		{
			name: "nonexistent template",
			config: &config.Config{
				DomainTemplate: "nonexistent",
				IngressHosts:    []string{"example.com"},
			},
			expectError: true,
		},
		{
			name: "empty ingress hosts",
			config: &config.Config{
				DomainTemplate: "default",
				IngressHosts:    []string{},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultPath, err := ProcessDomainTemplate(tc.config)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ProcessDomainTemplate failed: %v", err)
			}

			// Clean up the temporary file
			defer os.Remove(resultPath)

			// If we expect content, check it
			if tc.checkContent != "" {
				content, err := os.ReadFile(resultPath)
				if err != nil {
					t.Fatalf("Failed to read result file: %v", err)
				}

				if !strings.Contains(string(content), tc.checkContent) {
					t.Errorf("Expected content to contain %q, got:\n%s",
						tc.checkContent, string(content))
				}
			}
		})
	}
}

func TestProcessDomainTemplate_CustomFile(t *testing.T) {
	// Create a custom template file
	tmpDir, err := os.MkdirTemp("", "domain-templates")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple test template
	testTemplate := `ingress:
  enabled: true
  hosts:
{{- range $i, $host := .IngressHosts }}
  - host: {{ $host }}
{{- end }}`

	customPath := filepath.Join(tmpDir, "custom.yml")
	if err := os.WriteFile(customPath, []byte(testTemplate), 0644); err != nil {
		t.Fatalf("Failed to write custom template: %v", err)
	}

	// Test with a custom template file
	config := &config.Config{
		DomainTemplate: customPath,
		IngressHosts:    []string{"example.com", "www.example.com"},
	}

	resultPath, err := ProcessDomainTemplate(config)
	if err != nil {
		t.Fatalf("ProcessDomainTemplate failed with custom file: %v", err)
	}
	defer os.Remove(resultPath)

	// Check content
	content, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}

	if !strings.Contains(string(content), "host: example.com") {
		t.Errorf("Expected content to contain 'host: example.com', got:\n%s", string(content))
	}
}

func TestGetEmbeddedTemplate(t *testing.T) {
	// Test cases for GetEmbeddedTemplate
	testCases := []struct {
		name     string
		template string
		found    bool
	}{
		{
			name:     "default template",
			template: "default",
			found:    true,
		},
		{
			name:     "bitnami template",
			template: "bitnami",
			found:    true,
		},
		{
			name:     "vault template",
			template: "vault",
			found:    true,
		},
		{
			name:     "nonexistent template",
			template: "nonexistent",
			found:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, found := GetEmbeddedTemplate(tc.template)

			if found != tc.found {
				t.Errorf("Expected found=%v, got %v", tc.found, found)
			}

			if tc.found {
				if content == "" {
					t.Errorf("Expected non-empty content for template %s", tc.template)
				}
			}
		})
	}
}

func TestListEmbeddedTemplates(t *testing.T) {
	templates := ListEmbeddedTemplates()

	// Check that we have at least the three known templates
	expectedTemplates := []string{"default", "bitnami", "vault"}
	for _, expected := range expectedTemplates {
		found := false
		for _, actual := range templates {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected template %s not found in list: %v", expected, templates)
		}
	}
}
