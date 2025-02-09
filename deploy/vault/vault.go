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
}

type VaultClient struct {
	config VaultConfig
	client *http.Client
}

func NewVaultClient(config VaultConfig) *VaultClient {
	// Create custom transport with configurable TLS settings
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

// getSecret retrieves a secret from Vault
func (v *VaultClient) getSecret(path string) (map[string]interface{}, error) {
	fullPath := fmt.Sprintf("%s/v1/%s", v.config.URL, strings.TrimPrefix(path, "/"))
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
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Data, nil
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
	// Remove the <<vault. prefix
	path := strings.TrimPrefix(template, "<<vault.")

	// If basePath is set, prepend it unless the path is absolute
	if v.config.BasePath != "" && !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("%s/%s", strings.Trim(v.config.BasePath, "/"), path)
	}

	// Split path into secret path and key
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid vault template format: %s", template)
	}

	secretKey := parts[len(parts)-1]
	secretPath := strings.Join(parts[:len(parts)-1], "/")

	secret, err := v.getSecret(secretPath)
	if err != nil {
		return nil, err
	}

	value, ok := secret[secretKey]
	if !ok {
		return nil, fmt.Errorf("key %s not found in secret %s", secretKey, secretPath)
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
