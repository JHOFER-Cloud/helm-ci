package vault

import (
	"fmt"
	"regexp"
	"strings"
)

var vaultPlaceholderRegex = regexp.MustCompile(`<<vault\.[^>]+>>`)

func (c *Client) ProcessString(input string) (string, error) {
	result := input

	matches := vaultPlaceholderRegex.FindAllString(input, -1)
	for _, placeholder := range matches {
		secretValue, err := c.GetSecret(placeholder)
		if err != nil {
			return "", fmt.Errorf("failed to process placeholder %s: %w", placeholder, err)
		}

		result = strings.ReplaceAll(result, placeholder, secretValue)
	}

	return result, nil
}
