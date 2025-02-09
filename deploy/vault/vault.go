package vault

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"
)

type VaultConfig struct {
	URL           string
	Token         string
	BasePath      string
	InsecureHTTPS bool
	KVVersion     int
}

type VaultClient struct {
	config VaultConfig
	client *http.Client
}

func NewVaultClient(config VaultConfig) *VaultClient {
	if config.KVVersion == 0 {
		config.KVVersion = 2 // Default to KV v2 if not specified
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureHTTPS,
		},
	}

	return &VaultClient{
		config: config,
		client: &http.Client{Transport: tr},
	}
}

// formatVaultPath formats the path according to KV version
func (v *VaultClient) formatVaultPath(path string) string {
	// Remove any leading/trailing slashes
	path = strings.Trim(path, "/")

	// Split path into mount and key path
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return path // Return as is if can't split
	}

	mount, keyPath := parts[0], parts[1]

	// Format according to KV version
	if v.config.KVVersion == 2 {
		// KV v2 format: /v1/[mount]/data/[path]
		return fmt.Sprintf("%s/v1/%s/data/%s", v.config.URL, mount, keyPath)
	}

	// KV v1 format: /v1/[mount]/[path]
	return fmt.Sprintf("%s/v1/%s/%s", v.config.URL, mount, keyPath)
}

// getSecret retrieves a secret from Vault
func (v *VaultClient) getSecret(path string) (map[string]interface{}, error) {
	fullPath := v.formatVaultPath(path)
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("X-Vault-Token", v.config.Token)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get secret, status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"` // For KV v2
		} `json:"data"`
		// Direct data for KV v1
		DirectData map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Handle different response structures for KV v1 and v2
	if v.config.KVVersion == 2 {
		return result.Data.Data, nil
	}
	return result.DirectData, nil
}

// ProcessYAMLWithVaultTemplates processes a YAML file and replaces Vault templates
func (v *VaultClient) ProcessYAMLWithVaultTemplates(yamlData []byte) ([]byte, error) {
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	processed, err := v.processNode(data)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(processed)
}

func (v *VaultClient) processNode(node interface{}) (interface{}, error) {
	switch n := node.(type) {
	case map[string]interface{}:
		for key, value := range n {
			processed, err := v.processNode(value)
			if err != nil {
				return nil, err
			}
			n[key] = processed
		}
		return n, nil

	case []interface{}:
		for i, value := range n {
			processed, err := v.processNode(value)
			if err != nil {
				return nil, err
			}
			n[i] = processed
		}
		return n, nil

	case string:
		if strings.HasPrefix(n, "<<vault.") {
			return v.resolveVaultTemplate(n)
		}
		return n, nil

	default:
		return n, nil
	}
}

func (v *VaultClient) resolveVaultTemplate(template string) (interface{}, error) {
	// Remove the <<vault. prefix and get path
	path := strings.TrimPrefix(template, "<<vault.")

	// Split the path to separate the secret key
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid vault template format: %s", template)
	}

	secretKey := parts[len(parts)-1]
	secretPath := strings.Join(parts[:len(parts)-1], "/")

	// If basePath is set, prepend it unless the path is absolute
	if v.config.BasePath != "" && !strings.HasPrefix(secretPath, "/") {
		secretPath = fmt.Sprintf("%s/%s", strings.Trim(v.config.BasePath, "/"), secretPath)
	}

	secret, err := v.getSecret(secretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret from path %s: %v", secretPath, err)
	}

	value, ok := secret[secretKey]
	if !ok {
		return nil, fmt.Errorf("key %s not found in secret at path %s", secretKey, secretPath)
	}

	// Handle multiline strings
	if str, ok := value.(string); ok && strings.Contains(str, "\n") {
		return formatMultilineString(str), nil
	}

	return value, nil
}

func formatMultilineString(s string) string {
	// If the string contains newlines, format it as a YAML multiline string
	if strings.Contains(s, "\n") {
		return fmt.Sprintf("|\n%s", indentString(s, 2))
	}
	return s
}

func indentString(s string, spaces int) string {
	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
