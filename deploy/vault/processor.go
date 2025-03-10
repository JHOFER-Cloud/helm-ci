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
	"helm-ci/deploy/utils"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var vaultPlaceholderRegex = regexp.MustCompile(`<<vault\.[^>]+>>`)

func (c *Client) ProcessString(input string) (string, error) {
	// First process vault placeholders
	result := input
	matches := vaultPlaceholderRegex.FindAllString(input, -1)

	for _, placeholder := range matches {
		secretValue, err := c.GetSecret(placeholder)
		if err != nil {
			return "", utils.NewError("failed to process placeholder %s: %w", placeholder, err)
		}

		// Check if the secret value contains newlines
		if strings.Contains(secretValue, "\n") {
			// Find the indentation of the placeholder
			re := regexp.MustCompile(`(?m)^(\s*).*` + regexp.QuoteMeta(placeholder) + `.*$`)
			match := re.FindStringSubmatch(input)
			if len(match) > 1 {
				indent := match[1]
				// Add the literal block scalar indicator and indent the content
				lines := strings.Split(secretValue, "\n")
				for i := 0; i < len(lines); i++ {
					if lines[i] != "" {
						lines[i] = indent + "  " + lines[i] // Add two spaces for content indentation
					}
				}
				// Replace the placeholder with the block scalar
				blockValue := "|\n" + strings.Join(lines, "\n")
				result = strings.ReplaceAll(result, placeholder, blockValue)
			}
		} else {
			// For single-line values, just replace directly
			result = strings.ReplaceAll(result, placeholder, secretValue)
		}
	}

	// Then handle Kubernetes Secret base64 encoding if needed
	if strings.Contains(result, "kind: Secret") {
		var secret map[string]interface{}
		if err := yaml.Unmarshal([]byte(result), &secret); err != nil {
			return "", utils.NewError("failed to parse Secret YAML: %v", err)
		}

		// Get the data section and ensure it's a map
		if data, ok := secret["data"].(map[string]interface{}); ok {
			newData := make(map[string]interface{})
			for k, v := range data {
				newData[k] = v
			}
			secret["data"] = newData
		}

		// Convert back to YAML
		yamlBytes, err := yaml.Marshal(secret)
		if err != nil {
			return "", utils.NewError("failed to marshal Secret YAML: %v", err)
		}
		result = string(yamlBytes)
	}

	return result, nil
}
