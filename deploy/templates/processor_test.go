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
	// Create test domain templates
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

	templatePath := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(templatePath, []byte(testTemplate), 0644); err != nil {
		t.Fatalf("Failed to write test template: %v", err)
	}

	// Test cases
	testCases := []struct {
		name         string
		config       *config.Config
		expectError  bool
		checkContent string
	}{
		{
			name: "basic template processing",
			config: &config.Config{
				DomainsTemplate: templatePath,
				IngressHosts:    []string{"example.com", "www.example.com"},
			},
			expectError:  false,
			checkContent: "host: example.com",
		},
		{
			name: "empty ingress hosts",
			config: &config.Config{
				DomainsTemplate: templatePath,
				IngressHosts:    []string{},
			},
			expectError: false,
		},
		{
			name: "nonexistent template",
			config: &config.Config{
				DomainsTemplate: "nonexistent.yml",
				IngressHosts:    []string{"example.com"},
			},
			expectError: true,
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
