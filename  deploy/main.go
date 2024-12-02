package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	Stage       string
	AppName     string
	Environment string
	PRNumber    string
	ValuesPath  string
	Custom      bool
	Chart       string
	Version     string
	Repository  string
	BaseDomain  string
	Namespace   string
	ReleaseName string
	IngressHost string
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Stage, "stage", "", "Deployment stage (dev/live)")
	flag.StringVar(&cfg.AppName, "app", "", "Application name")
	flag.StringVar(&cfg.Environment, "env", "", "Environment")
	flag.StringVar(&cfg.PRNumber, "pr", "", "PR number")
	flag.StringVar(&cfg.ValuesPath, "values", "helm/values", "Path to values files")
	flag.BoolVar(&cfg.Custom, "custom", false, "Use custom values only")
	flag.StringVar(&cfg.Chart, "chart", "", "Helm chart (optional)")
	flag.StringVar(&cfg.Version, "version", "", "Chart version (optional)")
	flag.StringVar(&cfg.Repository, "repo", "", "Helm repository (optional)")
	flag.StringVar(&cfg.BaseDomain, "domain", "", "Base domain")

	flag.Parse()

	if cfg.AppName == "" {
		fmt.Println("app name is required")
		os.Exit(1)
	}

	if cfg.Stage == "" {
		fmt.Println("stage is required")
		os.Exit(1)
	}

	if cfg.Environment == "" {
		fmt.Println("environment is required")
		os.Exit(1)
	}

	if !cfg.Custom && (cfg.Chart == "" || cfg.Repository == "") {
		fmt.Println("chart and repository are required when not using custom values")
		os.Exit(1)
	}

	return cfg
}

func (c *Config) setupNames() {
	c.Namespace = c.AppName
	c.ReleaseName = c.AppName

	if c.Stage == "dev" && c.PRNumber != "" {
		c.Namespace = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		c.ReleaseName = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.BaseDomain)
	} else {
		if c.Stage == "dev" {
			c.IngressHost = fmt.Sprintf("%s.dev.%s", c.AppName, c.BaseDomain)
		} else {
			c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.BaseDomain)
		}
	}
}

func (c *Config) ensureNamespace() error {
	cmd := exec.Command("kubectl", "create", "namespace", c.Namespace, "--dry-run=client", "-o", "yaml")
	createOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error creating namespace yaml: %v", err)
	}

	cmd = exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(string(createOutput))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error applying namespace: %v", err)
	}

	return nil
}

type ValuesTemplateData struct {
	Stage       string
	AppName     string
	Environment string
	IngressHost string
}

func (c *Config) processValues() error {
	data := ValuesTemplateData{
		Stage:       c.Stage,
		AppName:     c.AppName,
		Environment: c.Environment,
		IngressHost: c.IngressHost,
	}

	if err := os.MkdirAll("processed-values", 0755); err != nil {
		return fmt.Errorf("error creating processed-values directory: %v", err)
	}

	// Process common values
	commonPath := filepath.Join(c.ValuesPath, "common.yml")
	if _, err := os.Stat(commonPath); err == nil {
		if err := processTemplate(commonPath, "processed-values/common.yml", data); err != nil {
			return fmt.Errorf("error processing common values: %v", err)
		}
	}

	// Process stage-specific values
	stagePath := filepath.Join(c.ValuesPath, fmt.Sprintf("%s.yml", c.Stage))
	if _, err := os.Stat(stagePath); err == nil {
		if err := processTemplate(stagePath, fmt.Sprintf("processed-values/%s.yml", c.Stage), data); err != nil {
			return fmt.Errorf("error processing stage values: %v", err)
		}
	}

	return nil
}

func processTemplate(inputPath, outputPath string, data ValuesTemplateData) error {
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading template file: %v", err)
	}

	tmpl, err := template.New(filepath.Base(inputPath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}

func (c *Config) deployHelm() error {
	if !c.Custom && (c.Chart == "" || c.Repository == "") {
		return fmt.Errorf("chart and repository required when not using custom values")
	}

	if !c.Custom {
		cmd := exec.Command("helm", "repo", "add", "app", c.Repository)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error adding helm repo: %v", err)
		}

		cmd = exec.Command("helm", "repo", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error updating helm repos: %v", err)
		}
	}

	args := []string{
		"upgrade", "--install",
		c.ReleaseName,
	}

	if !c.Custom {
		args = append(args, c.Chart)
	}

	args = append(args,
		"--namespace", c.Namespace,
		"--create-namespace",
		"--wait",
		"--timeout", "10m",
	)

	if !c.Custom && c.Version != "" {
		args = append(args, "--version", c.Version)
	}

	// Add default Traefik values if not in custom values
	defaultValues := fmt.Sprintf(`
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.ingress.kubernetes.io/router.entrypoints: web,websecure
    traefik.ingress.kubernetes.io/router.middlewares: traefik-strip-prefix@kubernetescrd
  hosts:
    - host: %s
      paths:
        - path: /
          pathType: Prefix
`, c.IngressHost)

	tmpFile, err := os.CreateTemp("", "default-values-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if err := os.WriteFile(tmpFile.Name(), []byte(defaultValues), 0644); err != nil {
		return err
	}

	// Add values files
	if !c.Custom {
		args = append(args, "-f", tmpFile.Name())
	}

	commonValues := filepath.Join("processed-values", "common.yml")
	if _, err := os.Stat(commonValues); err == nil {
		args = append(args, "-f", commonValues)
	}

	stageValues := filepath.Join("processed-values", fmt.Sprintf("%s.yml", c.Stage))
	if _, err := os.Stat(stageValues); err == nil {
		args = append(args, "-f", stageValues)
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func main() {
	cfg := parseFlags()

	cfg.setupNames()

	if err := cfg.ensureNamespace(); err != nil {
		fmt.Printf("Error creating namespace: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.processValues(); err != nil {
		fmt.Printf("Error processing values: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.deployHelm(); err != nil {
		fmt.Printf("Error deploying helm: %v\n", err)
		os.Exit(1)
	}
}
