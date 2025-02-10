package vault

import (
	"fmt"
	"strings"
)

type VaultPath struct {
	BasePath string
	Path     string
	Key      string
	Version  int
}

func ParseVaultPath(placeholder string) (*VaultPath, error) {
	// Check for proper <<...>> format
	if !strings.HasPrefix(placeholder, "<<") || !strings.HasSuffix(placeholder, ">>") {
		return nil, fmt.Errorf("invalid vault placeholder format: must be enclosed in <<>>")
	}

	inner := strings.TrimPrefix(strings.TrimSuffix(placeholder, ">>"), "<<")
	if !strings.HasPrefix(inner, "vault.") {
		return nil, fmt.Errorf("invalid vault placeholder format: must start with vault")
	}

	parts := strings.Split(strings.TrimPrefix(inner, "vault."), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid vault path format: must have at least one path segment and key")
	}

	key := parts[len(parts)-1]
	path := strings.Join(parts[:len(parts)-1], "/")

	return &VaultPath{Path: path, Key: key}, nil
}

func (vp *VaultPath) BuildSecretPath() string {
	vp.BasePath = strings.TrimSuffix(vp.BasePath, "/")
	if vp.Version == KVv2 {
		return fmt.Sprintf("%s/data/%s", vp.BasePath, vp.Path)
	}
	return fmt.Sprintf("%s/%s", vp.BasePath, vp.Path)
}
