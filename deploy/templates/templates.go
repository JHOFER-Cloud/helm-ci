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

package templates

import (
	_ "embed" // Import embed for directive
)

// Embedded template files for domains

//go:embed domains/default.yml
var DefaultDomainTemplate string

//go:embed domains/bitnami.yml
var BitnamiDomainTemplate string

//go:embed domains/vault.yml
var VaultDomainTemplate string

// GetEmbeddedTemplate returns the content of a built-in template by name
// Returns the content and a boolean indicating if the template was found
func GetEmbeddedTemplate(name string) (string, bool) {
	switch name {
	case "default":
		return DefaultDomainTemplate, true
	case "bitnami":
		return BitnamiDomainTemplate, true
	case "vault":
		return VaultDomainTemplate, true
	default:
		return "", false
	}
}

// ListEmbeddedTemplates returns the names of all built-in templates
func ListEmbeddedTemplates() []string {
	return []string{
		"default",
		"bitnami",
		"vault",
	}
}
