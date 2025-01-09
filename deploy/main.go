package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
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
	Namespace   string
	ReleaseName string
	IngressHost string
	GitHubToken string
	GitHubRepo  string
	GitHubOwner string
	Domain      string
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
	flag.StringVar(&cfg.GitHubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flag.StringVar(&cfg.GitHubRepo, "github-repo", "", "GitHub repository name")
	flag.StringVar(&cfg.GitHubOwner, "github-owner", "", "GitHub repository owner")
	flag.StringVar(&cfg.Domain, "domain", "", "Ingress domain")
	flag.BoolVar(&cfg.Custom, "custom", false, "Custom Kubernetes deployment")
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
		c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.Domain)
	} else {
		c.ReleaseName = c.AppName
		c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.Domain)
	}
}

func (c *Config) Deploy() error {
	if c.Custom {
		return c.deployCustom()
	}
	return c.deployHelm()
}

func (c *Config) deployHelm() error {
	// Add the Helm repository
	addRepoCmd := exec.Command("helm", "repo", "add", c.AppName, c.Repository)
	addRepoCmd.Stdout = os.Stdout
	addRepoCmd.Stderr = os.Stderr
	if err := addRepoCmd.Run(); err != nil {
		return fmt.Errorf("failed to add Helm repository: %v", err)
	}

	// Update the Helm repository
	updateRepoCmd := exec.Command("helm", "repo", "update")
	updateRepoCmd.Stdout = os.Stdout
	updateRepoCmd.Stderr = os.Stderr
	if err := updateRepoCmd.Run(); err != nil {
		return fmt.Errorf("failed to update Helm repository: %v", err)
	}

	// Determine the values files to use
	valuesFiles := []string{filepath.Join(c.ValuesPath, "common.yml")}
	stageValuesFile := filepath.Join(c.ValuesPath, fmt.Sprintf("%s.yml", c.Stage))
	valuesFiles = append(valuesFiles, stageValuesFile)

	// Deploy the Helm chart
	args := []string{
		"upgrade", "--install", c.ReleaseName, c.Chart,
		"--namespace", c.Namespace,
		"--set", fmt.Sprintf("ingress.host=%s", c.IngressHost),
	}

	for _, valuesFile := range valuesFiles {
		args = append(args, "--values", valuesFile)
	}

	if c.Version != "" {
		args = append(args, "--version", c.Version)
	}

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *Config) deployCustom() error {
	// Assuming custom Kubernetes manifests are in the values path
	manifests, err := filepath.Glob(filepath.Join(c.ValuesPath, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to find manifests: %v", err)
	}

	for _, manifest := range manifests {
		cmd := exec.Command("kubectl", "apply", "-f", manifest, "-n", c.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}

func main() {
	cfg := parseFlags()
	cfg.setupNames()
	cfg.PrintConfig()

	if err := cfg.Deploy(); err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deployment succeeded")
}
