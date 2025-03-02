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
	"helm-ci/deploy/utils"
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
		return nil, utils.NewError("invalid vault placeholder format: must be enclosed in <<>>")
	}

	inner := strings.TrimPrefix(strings.TrimSuffix(placeholder, ">>"), "<<")
	if !strings.HasPrefix(inner, "vault.") {
		return nil, utils.NewError("invalid vault placeholder format: must start with vault")
	}

	parts := strings.Split(strings.TrimPrefix(inner, "vault."), "/")
	if len(parts) < 2 {
		return nil, utils.NewError("invalid vault path format: must have at least one path segment and key")
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
