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
	"bytes"
	"helm-ci/deploy/config"
	"helm-ci/deploy/utils"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ProcessDomainTemplate creates a values file from a domain template
func ProcessDomainTemplate(cfg *config.Config) (string, error) {
	if len(cfg.IngressHosts) == 0 {
		return "", nil
	}

	// Determine template content
	var templateContent string
	templatePath := cfg.DomainTemplate

	if !strings.Contains(templatePath, "/") {
		// It's not a path, try to get it from embedded templates
		content, found := GetEmbeddedTemplate(templatePath)
		if found {
			utils.Log.Debugf("Using embedded template: %s", templatePath)
			templateContent = content
		} else {
			// If not found in embedded templates, try as a path to a file
			builtinPath := filepath.Join("deploy", "templates", "domains", templatePath+".yml")
			utils.Log.Debugf("Embedded template '%s' not found, checking file: %s", templatePath, builtinPath)

			if content, err := os.ReadFile(builtinPath); err == nil {
				templateContent = string(content)
			} else {
				// Show available templates in error message
				available := ListEmbeddedTemplates()
				return "", utils.NewError("template '%s' not found. Available built-in templates: %v",
					templatePath, available)
			}
		}
	} else {
		// It's a path to a custom template file
		utils.Log.Debugf("Reading custom template file: %s", templatePath)
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return "", utils.NewError("failed to read domain template file: %v", err)
		}
		templateContent = string(content)
	}

	// Process template
	tmpl, err := template.New("domains").Parse(templateContent)
	if err != nil {
		return "", utils.NewError("failed to parse domain template: %v", err)
	}

	data := struct {
		Domains      []string
		IngressHosts []string
		Config       *config.Config
	}{
		Domains:      cfg.Domains,
		IngressHosts: cfg.IngressHosts,
		Config:       cfg,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", utils.NewError("failed to execute domain template: %v", err)
	}

	// Create temporary file for the processed template
	tmpFile, err := os.CreateTemp("", "domains-*.yml")
	if err != nil {
		return "", utils.NewError("failed to create temporary file: %v", err)
	}

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		os.Remove(tmpFile.Name())
		return "", utils.NewError("failed to write to temporary file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", utils.NewError("failed to close temporary file: %v", err)
	}

	utils.Log.Infof("Using template: %s", templatePath)
	return tmpFile.Name(), nil
}
