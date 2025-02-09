package vault

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	token      string
	basePath   string
	kvVersion  int
	httpClient *http.Client
}

func NewClient(baseURL, token, basePath string, kvVersion int, insecureTLS bool) (*Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureTLS,
		},
	}

	return &Client{
		baseURL:    baseURL,
		token:      token,
		basePath:   basePath,
		kvVersion:  kvVersion,
		httpClient: &http.Client{Transport: tr},
	}, nil
}

func (c *Client) GetSecret(placeholder string) (string, error) {
	vPath, err := ParseVaultPath(placeholder)
	if err != nil {
		return "", err
	}

	vPath.BasePath = c.basePath
	secretPath := vPath.BuildSecretPath()

	url := fmt.Sprintf("%s/v1/%s", c.baseURL, secretPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vault request failed: %s, status: %d", string(body), resp.StatusCode)
	}

	var result struct {
		Data struct {
			Data map[string]string `json:"data"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	value, ok := result.Data.Data[vPath.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret", vPath.Key)
	}

	return value, nil
}
