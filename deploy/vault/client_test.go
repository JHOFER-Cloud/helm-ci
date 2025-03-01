package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		token       string
		basePath    string
		kvVersion   int
		insecureTLS bool
		expectError bool
	}{
		{
			name:        "valid client kv1",
			baseURL:     "https://vault.example.com",
			token:       "test-token",
			basePath:    "secret",
			kvVersion:   KVv1,
			insecureTLS: false,
			expectError: false,
		},
		{
			name:        "valid client kv2",
			baseURL:     "https://vault.example.com",
			token:       "test-token",
			basePath:    "secret",
			kvVersion:   KVv2,
			insecureTLS: true,
			expectError: false,
		},
		{
			name:        "invalid kv version",
			baseURL:     "https://vault.example.com",
			token:       "test-token",
			basePath:    "secret",
			kvVersion:   3, // Invalid
			insecureTLS: false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.baseURL, tc.token, tc.basePath, tc.kvVersion, tc.insecureTLS)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}

			if client.baseURL != tc.baseURL {
				t.Errorf("Expected baseURL %q, got %q", tc.baseURL, client.baseURL)
			}

			if client.token != tc.token {
				t.Errorf("Expected token %q, got %q", tc.token, client.token)
			}

			if client.basePath != tc.basePath {
				t.Errorf("Expected basePath %q, got %q", tc.basePath, client.basePath)
			}

			if client.kvVersion != tc.kvVersion {
				t.Errorf("Expected kvVersion %d, got %d", tc.kvVersion, client.kvVersion)
			}
		})
	}
}

func TestGetSecret(t *testing.T) {
	// Create a test server that simulates Vault responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request has the correct token
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Handle different paths
		switch r.URL.Path {
		case "/v1/secret/data/test/path":
			// KV v2 response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"data": {
						"test-key": "test-value"
					}
				}
			}`))
		case "/v1/secret/test/path":
			// KV v1 response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"data": {
					"test-key": "test-value"
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test cases
	testCases := []struct {
		name        string
		client      *Client
		placeholder string
		expected    string
		expectError bool
	}{
		{
			name: "kv v2 valid secret",
			client: &Client{
				baseURL:    server.URL,
				token:      "test-token",
				basePath:   "secret",
				kvVersion:  KVv2,
				httpClient: http.DefaultClient,
			},
			placeholder: "<<vault.test/path/test-key>>",
			expected:    "test-value",
			expectError: false,
		},
		{
			name: "kv v1 valid secret",
			client: &Client{
				baseURL:    server.URL,
				token:      "test-token",
				basePath:   "secret",
				kvVersion:  KVv1,
				httpClient: http.DefaultClient,
			},
			placeholder: "<<vault.test/path/test-key>>",
			expected:    "test-value",
			expectError: false,
		},
		{
			name: "invalid placeholder format",
			client: &Client{
				baseURL:    server.URL,
				token:      "test-token",
				basePath:   "secret",
				kvVersion:  KVv2,
				httpClient: http.DefaultClient,
			},
			placeholder: "not-a-vault-placeholder",
			expectError: true,
		},
		{
			name: "invalid token",
			client: &Client{
				baseURL:    server.URL,
				token:      "invalid-token",
				basePath:   "secret",
				kvVersion:  KVv2,
				httpClient: http.DefaultClient,
			},
			placeholder: "<<vault.test/path/test-key>>",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := tc.client.GetSecret(tc.placeholder)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetSecret failed: %v", err)
			}

			if value != tc.expected {
				t.Errorf("Expected value %q, got %q", tc.expected, value)
			}
		})
	}
}
