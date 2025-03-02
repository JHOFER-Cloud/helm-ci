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

	// Determine template path
	templatePath := cfg.DomainsTemplate
	if !strings.Contains(templatePath, "/") {
		// Check if it's a built-in template
		builtinPath := filepath.Join("deploy", "templates", "domains", templatePath+".yml")
		if _, err := os.Stat(builtinPath); err == nil {
			templatePath = builtinPath
		} else {
			return "", utils.NewError("built-in template '%s' not found", cfg.DomainsTemplate)
		}
	}

	// Read template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", utils.NewError("failed to read domain template: %v", err)
	}

	// Process template
	tmpl, err := template.New("domains").Parse(string(content))
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

	return tmpFile.Name(), nil
}
