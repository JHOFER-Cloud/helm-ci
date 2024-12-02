package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	flag.StringVar(&cfg.Chart, "chart", "", "Helm chart (optional)")
	flag.StringVar(&cfg.Version, "version", "", "Chart version (optional)")
	flag.StringVar(&cfg.Repository, "repo", "", "Helm repository (optional)")
	flag.StringVar(&cfg.BaseDomain, "domain", "", "Base domain")
	flag.StringVar(&cfg.GitHubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flag.StringVar(&cfg.GitHubRepo, "github-repo", "", "GitHub repository name")
	flag.StringVar(&cfg.GitHubOwner, "github-owner", "", "GitHub repository owner")
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

	// If chart and repository are provided, it's not a custom deployment
	cfg.Custom = cfg.Chart == "" || cfg.Repository == ""

	return cfg
}

func (c *Config) PrintConfig() {
	fmt.Println("Current Configuration:")
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		fmt.Printf("- %s: %v\n", fieldName, field.Interface())
	}
}

func (c *Config) setupNames() {
	// Set namespace based on stage
	if c.Stage == "live" {
		c.Namespace = c.AppName
	} else {
		c.Namespace = c.AppName + "-dev"
	}

	// For PRs, only modify the release name to include PR number
	if c.Stage == "dev" && c.PRNumber != "" {
		c.ReleaseName = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.BaseDomain)
	} else {
		c.ReleaseName = c.AppName
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
	c.PrintConfig()

	if !c.Custom && (c.Chart == "" || c.Repository == "") {
		return fmt.Errorf("chart and repository required when not using custom values")
	}

	if !c.Custom {
		// Add and update helm repo
		repoName := strings.Split(c.Chart, "/")[0]
		cmd := exec.Command("helm", "repo", "add", repoName, c.Repository)
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

	// Create helm upgrade command without --wait first
	args := []string{
		"upgrade", "--install",
		c.ReleaseName,
	}

	if c.Custom {
		args = append(args, ".")
	} else {
		args = append(args, c.Chart)
	}

	args = append(args,
		"--namespace", c.Namespace,
		"--create-namespace",
	)

	if !c.Custom && c.Version != "" {
		args = append(args, "--version", c.Version)
	}

	// Add values files
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

	// First deploy without --wait
	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error deploying helm chart: %v", err)
	}

	// Then wait for rollout
	fmt.Println("Waiting for deployments to roll out...")
	cmd = exec.Command("kubectl", "get", "deploy", "-n", c.Namespace, "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error getting deployments: %v", err)
	}

	deployments := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, deployment := range deployments {
		if deployment == "" {
			continue
		}
		fmt.Printf("Waiting for %s...\n", deployment)
		cmd = exec.Command("kubectl", "rollout", "status", "-n", c.Namespace, deployment, "--timeout=10m")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error waiting for deployment %s: %v", deployment, err)
		}
	}

	return nil
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

	// Output URL for GitHub Actions
	f, err := os.OpenFile(os.Getenv("GITHUB_ENV"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// handle error
	}
	defer f.Close()
	fmt.Fprintf(f, "URL=https://%s\n", cfg.IngressHost)
}
