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
