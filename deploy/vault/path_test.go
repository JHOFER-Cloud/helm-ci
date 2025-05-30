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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseVaultPath(t *testing.T) {
	tests := []struct {
		name        string
		placeholder string
		want        *VaultPath
		wantErr     bool
	}{
		{
			name:        "valid placeholder",
			placeholder: "<<vault.renovate/common/API_TOKEN>>",
			want: &VaultPath{
				Path: "renovate/common",
				Key:  "API_TOKEN",
			},
			wantErr: false,
		},
		{
			name:        "invalid format - no vault prefix",
			placeholder: "<<invalid.path>>",
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "invalid format - no brackets",
			placeholder: "vault.path/key",
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "invalid format - too few parts",
			placeholder: "<<vault.key>>",
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVaultPath(tt.placeholder)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVaultPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !pathsEqual(got, tt.want) {
				t.Errorf("ParseVaultPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVaultPath_BuildSecretPath(t *testing.T) {
	tests := []struct {
		name     string
		path     *VaultPath
		expected string
	}{
		{
			name: "KV v2 path",
			path: &VaultPath{
				BasePath: "talos",
				Path:     "renovate/common",
				Version:  KVv2,
			},
			expected: "talos/data/renovate/common",
		},
		{
			name: "KV v1 path",
			path: &VaultPath{
				BasePath: "talos",
				Path:     "renovate/common",
				Version:  KVv1,
			},
			expected: "talos/renovate/common",
		},
		{
			name: "KV v2 path with trailing slash",
			path: &VaultPath{
				BasePath: "talos/",
				Path:     "renovate/common",
				Version:  KVv2,
			},
			expected: "talos/data/renovate/common",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.path.BuildSecretPath()
			if got != tt.expected {
				t.Errorf("BuildSecretPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestClient_ProcessString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/v1/talos/data/renovate/common":
			fmt.Fprintf(w, `{
                "data": {
                    "data": {
                        "API_TOKEN": "{\n  \"platform\": \"github\",\n  \"token\": \"SECRET\",\n  \"autodiscover\": \"true\"\n}",
                        "SIMPLE_TOKEN": "simple-secret"
                    }
                }
            }`)
		case "/v1/talos/data/longhorn/dev":
			fmt.Fprintf(w, `{
                "data": {
                    "data": {
                        "CIFS_USERNAME": "test-user",
                        "CIFS_PASSWORD": "test-password"
                    }
                }
            }`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		client  *Client
		input   string
		want    string
		wantErr bool
	}{
		{
			name:   "Multi-line JSON config",
			client: mustNewClient(t, server.URL, "test-token", "talos", KVv2, true),
			input: `renovate:
  configIsSecret: true
  config: <<vault.renovate/common/API_TOKEN>>
  securityContext:
    allowPrivilegeEscalation: false`,
			want: `renovate:
  configIsSecret: true
  config: |
    {
      "platform": "github",
      "token": "SECRET",
      "autodiscover": "true"
    }
  securityContext:
    allowPrivilegeEscalation: false`,
			wantErr: false,
		},
		{
			name:    "Single line value",
			client:  mustNewClient(t, server.URL, "test-token", "talos", KVv2, true),
			input:   "token: <<vault.renovate/common/SIMPLE_TOKEN>>",
			want:    "token: simple-secret",
			wantErr: false,
		},
		{
			name:   "Kubernetes Secret",
			client: mustNewClient(t, server.URL, "test-token", "talos", KVv2, true),
			input: `apiVersion: v1
kind: Secret
metadata:
  name: cifs-secret
  namespace: longhorn-system
type: Opaque
data:
  CIFS_USERNAME: <<vault.longhorn/dev/CIFS_USERNAME>>
  CIFS_PASSWORD: <<vault.longhorn/dev/CIFS_PASSWORD>>`,
			want: `apiVersion: v1
kind: Secret
metadata:
  name: cifs-secret
  namespace: longhorn-system
type: Opaque
data:
  CIFS_USERNAME: test-user
  CIFS_PASSWORD: test-password`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.ProcessString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// For non-Secret cases, compare strings directly
				if !strings.Contains(tt.input, "kind: Secret") {
					if got != tt.want {
						t.Errorf("ProcessString() =\n%v\nwant\n%v", got, tt.want)
					}
					return
				}

				// For Secret cases, parse and compare the YAML structures
				var gotSecret, wantSecret map[string]interface{}
				if err := yaml.Unmarshal([]byte(got), &gotSecret); err != nil {
					t.Fatalf("Failed to parse got YAML: %v", err)
				}
				if err := yaml.Unmarshal([]byte(tt.want), &wantSecret); err != nil {
					t.Fatalf("Failed to parse want YAML: %v", err)
				}

				// Compare the parsed structures
				if !reflect.DeepEqual(gotSecret, wantSecret) {
					t.Errorf("ProcessString() parsed content differs:\ngot = %#v\nwant = %#v", gotSecret, wantSecret)
				}
			}
		})
	}
}

// Helper function to create a client
func mustNewClient(t *testing.T, url, token, basePath string, version int, insecure bool) *Client {
	client, err := NewClient(url, token, basePath, version, insecure)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	return client
}

// Helper function to compare VaultPath structs
func pathsEqual(a, b *VaultPath) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Path == b.Path && a.Key == b.Key
}
