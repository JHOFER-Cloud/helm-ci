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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"helm-ci/deploy/utils"
	"io"
	"net/http"
)

const (
	KVv1 = 1
	KVv2 = 2
)

type Client struct {
	baseURL    string
	token      string
	basePath   string
	kvVersion  int
	httpClient *http.Client
}

func NewClient(baseURL, token, basePath string, kvVersion int, insecureTLS bool) (*Client, error) {
	if kvVersion != KVv1 && kvVersion != KVv2 {
		return nil, utils.NewError("invalid KV version: must be 1 or 2")
	}

	if insecureTLS {
		utils.Log.Warning("Skipping TLS verification. This is insecure and should not be used in production.")
	}
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
	vPath.Version = c.kvVersion
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
		return "", utils.NewError("vault request failed: %s, status: %d", string(body), resp.StatusCode)
	}

	if c.kvVersion == KVv2 {
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
			return "", utils.NewError("key %s not found in secret", vPath.Key)
		}
		return value, nil
	} else {
		var result struct {
			Data map[string]string `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		value, ok := result.Data[vPath.Key]
		if !ok {
			return "", utils.NewError("key %s not found in secret", vPath.Key)
		}
		return value, nil
	}
}
