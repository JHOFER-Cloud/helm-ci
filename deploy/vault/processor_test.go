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

package vault

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProcessString(t *testing.T) {
	// Create a test server that simulates Vault responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/v1/secret/data/db/credentials":
			// KV v2 response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"password": "super-secret-password"
					}
				}
			}`))
		case "/v1/secret/data/config/json":
			// KV v2 response with multiline JSON
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"value": "{\n  \"key1\": \"value1\",\n  \"key2\": \"value2\"\n}"
					}
				}
			}`))
		case "/v1/secret/data/app/config":
			// KV v2 response with multiple placeholders
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"username": "admin",
						"password": "password123",
						"url": "https://example.com"
					}
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a vault client for testing
	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		basePath:   "secret",
		kvVersion:  KVv2,
		httpClient: http.DefaultClient,
	}

	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "simple value replacement",
			input:       "password: <<vault.db/credentials/password>>",
			expected:    "password: super-secret-password",
			expectError: false,
		},
		{
			name: "multiline JSON replacement",
			input: `config:
  json: <<vault.config/json/value>>`,
			expected: `config:
  json: |
    {
      "key1": "value1",
      "key2": "value2"
    }`,
			expectError: false,
		},
		{
			name: "kubernetes secret",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
type: Opaque
data:
  password: <<vault.db/credentials/password>>`,
			expectError: false, // We'll check the structure rather than exact string
		},
		{
			name:        "invalid placeholder",
			input:       "password: <<vault.invalid>>", // Missing path separator, will be caught by regex but fail in processing
			expectError: true,
		},
		// NEW TEST CASES
		{
			name: "multiple placeholders in one file",
			input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  DB_USERNAME: <<vault.app/config/username>>
  DB_PASSWORD: <<vault.app/config/password>>
  API_URL: <<vault.app/config/url>>`,
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  DB_USERNAME: admin
  DB_PASSWORD: password123
  API_URL: https://example.com`,
			expectError: false,
		},
		{
			name: "placeholder with complex indentation",
			input: `config:
  database:
    credentials:
      password: <<vault.db/credentials/password>>
  settings:
    json: <<vault.config/json/value>>`,
			expected: `config:
  database:
    credentials:
      password: super-secret-password
  settings:
    json: |
      {
        "key1": "value1",
        "key2": "value2"
      }`,
			expectError: false,
		},
		{
			name: "multiline value with mixed indentation",
			input: `
config:
  # This is a comment
  settings:
    data: <<vault.config/json/value>>
    other: value
    nested:
      option: nested-value`,
			expectError: false, // We'll verify structure matches expectations
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.ProcessString(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ProcessString failed: %v", err)
			}

			// For kubernetes secret test case, verify the structure
			if strings.Contains(tc.input, "kind: Secret") {
				var secret map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &secret); err != nil {
					t.Fatalf("Failed to parse result as YAML: %v", err)
				}

				// Verify it's a Secret
				if kind, ok := secret["kind"].(string); !ok || kind != "Secret" {
					t.Errorf("Expected kind: Secret, got: %v", secret["kind"])
				}

				// Verify data is present
				data, ok := secret["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("data section not found or not a map")
				}

				// Verify password is present
				if _, ok := data["password"]; !ok {
					t.Errorf("password not found in data section")
				}

				return
			}

			// For the multiline value with mixed indentation case, we need to do a more complex verification
			if tc.name == "multiline value with mixed indentation" {
				// Parse both the result and expected structure to compare
				var resultYaml map[string]interface{}
				if err := yaml.Unmarshal([]byte(result), &resultYaml); err != nil {
					t.Fatalf("Failed to parse result as YAML: %v", err)
				}

				// Check that the multiline content was properly inserted
				config, ok := resultYaml["config"].(map[string]interface{})
				if !ok {
					t.Fatalf("config section not found or not a map")
				}

				settings, ok := config["settings"].(map[string]interface{})
				if !ok {
					t.Fatalf("settings section not found or not a map")
				}

				// The data field should be a string that contains the JSON content
				data, ok := settings["data"].(string)
				if !ok {
					t.Fatalf("data section not found or not a string")
				}

				if !strings.Contains(data, "key1") || !strings.Contains(data, "value1") {
					t.Errorf("Expected JSON content in data field, got: %s", data)
				}

				// Verify the other fields remained unchanged
				if settings["other"] != "value" {
					t.Errorf("Expected settings.other to be 'value', got: %v", settings["other"])
				}

				// Verify nested structure remained intact
				nested, ok := settings["nested"].(map[string]interface{})
				if !ok {
					t.Fatalf("nested section not found or not a map")
				}

				if nested["option"] != "nested-value" {
					t.Errorf("Expected nested.option to be 'nested-value', got: %v", nested["option"])
				}

				return
			}

			// For other test cases, compare strings directly
			// Normalize line endings and whitespace for comparison
			normalizedResult := strings.ReplaceAll(strings.TrimSpace(result), "\r\n", "\n")
			normalizedExpected := strings.ReplaceAll(strings.TrimSpace(tc.expected), "\r\n", "\n")

			if normalizedResult != normalizedExpected {
				t.Errorf("ProcessString output doesn't match expected.\nGot:\n%s\n\nExpected:\n%s",
					normalizedResult, normalizedExpected)
			}
		})
	}
}

// Test handling of placeholders inside Kubernetes Secret YAML
func TestProcessString_KubernetesSecretWithMultipleValues(t *testing.T) {
	// Create a test server that simulates Vault responses for Secret values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/v1/secret/data/k8s/secret" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"username": "admin",
						"password": "super-secret",
						"api-key": "api-123-xyz"
					}
				}
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a vault client for testing
	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		basePath:   "secret",
		kvVersion:  KVv2,
		httpClient: http.DefaultClient,
	}

	// Test with kubernetes secret with multiple values
	input := `
apiVersion: v1
kind: Secret
metadata:
  name: api-credentials
  namespace: default
type: Opaque
data:
  username: <<vault.k8s/secret/username>>
  password: <<vault.k8s/secret/password>>
  api-key: <<vault.k8s/secret/api-key>>
`

	result, err := client.ProcessString(input)
	if err != nil {
		t.Fatalf("ProcessString failed: %v", err)
	}

	// Parse the result as YAML
	var secret map[string]interface{}
	if err := yaml.Unmarshal([]byte(result), &secret); err != nil {
		t.Fatalf("Failed to parse result as YAML: %v", err)
	}

	// Verify it's a Secret
	if kind, ok := secret["kind"].(string); !ok || kind != "Secret" {
		t.Errorf("Expected kind: Secret, got: %v", secret["kind"])
	}

	// Verify data section
	data, ok := secret["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data section not found or not a map")
	}

	// Check each field exists
	expectedFields := []string{"username", "password", "api-key"}
	for _, field := range expectedFields {
		if _, ok := data[field]; !ok {
			t.Errorf("Field %s not found in data section", field)
		}
	}

	// Ensure type is correct
	if typeName, ok := secret["type"].(string); !ok || typeName != "Opaque" {
		t.Errorf("Expected type: Opaque, got: %v", secret["type"])
	}
}

// Test the behavior when an input has multiple placeholders and one of them fails
func TestProcessString_PartialFailure(t *testing.T) {
	// Create a test server that simulates Vault responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// One valid path, one missing path
		if r.URL.Path == "/v1/secret/data/app/valid" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"key": "valid-value"
					}
				}
			}`))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a vault client for testing
	client := &Client{
		baseURL:    server.URL,
		token:      "test-token",
		basePath:   "secret",
		kvVersion:  KVv2,
		httpClient: http.DefaultClient,
	}

	// Input with one valid and one invalid placeholder
	input := `
config:
  valid: <<vault.app/valid/key>>
  invalid: <<vault.app/missing/key>>
`

	// This should fail because one placeholder can't be resolved
	_, err := client.ProcessString(input)
	if err == nil {
		t.Errorf("Expected error for partial failure, got nil")
	}
}
