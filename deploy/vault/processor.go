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

	return result, nil
}
