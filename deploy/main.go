package main

import (
	"flag"
	"fmt"
	"helm-ci/deploy/vault"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

type Config struct {
	AppName          string
	Chart            string
	Custom           bool
	DEBUG            bool
	Domain           string
	Environment      string
	GitHubOwner      string
	GitHubRepo       string
	GitHubToken      string
	IngressHost      string
	Namespace        string
	PRDeployments    bool
	PRNumber         string
	ReleaseName      string
	Repository       string
	RootCA           string
	Stage            string
	TraefikDashboard bool
	ValuesPath       string
	VaultBasePath    string
	VaultToken       string
	VaultURL         string
	Version          string
	VaultInsecureTLS bool
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
	flag.BoolVar(&cfg.TraefikDashboard, "traefik-dashboard", false, "Deploy Traefik dashboard")
	flag.StringVar(&cfg.RootCA, "root-ca", "", "Path to root CA certificate")
	flag.BoolVar(&cfg.PRDeployments, "pr-deployments", true, "Enable PR deployments")
	// New Vault-related flags
	flag.StringVar(&cfg.VaultURL, "vault-url", "", "Vault server URL")
	flag.StringVar(&cfg.VaultToken, "vault-token", os.Getenv("VAULT_TOKEN"), "Vault authentication token")
	flag.StringVar(&cfg.VaultBasePath, "vault-base-path", "", "Base path for Vault secrets")
	flag.BoolVar(&cfg.VaultInsecureTLS, "vault-insecure-tls", false, "Allow insecure TLS connections to Vault (not recommended for production)")
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
		// Don't print sensitive values
		if fieldName == "VaultToken" || fieldName == "GitHubToken" {
			fmt.Printf("- %s: [REDACTED]\n", fieldName)
		} else {
			fmt.Printf("- %s: %v\n", fieldName, field.Interface())
		}
	}
}

func (c *Config) setupNames() {
	if c.Stage == "live" {
		c.Namespace = c.AppName
	} else {
		c.Namespace = c.AppName + "-dev"
	}

	if c.Stage == "dev" && c.PRNumber != "" && c.PRDeployments {
		c.ReleaseName = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.Domain)
	} else {
		c.ReleaseName = c.AppName
		c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.Domain)
	}
}

func (c *Config) processValuesFileWithVault(filename string) (string, error) {
	if c.VaultURL == "" || c.VaultToken == "" {
		// If Vault is not configured, return the original file
		return filename, nil
	}

	if c.VaultInsecureTLS {
		fmt.Println("Warning: Using insecure TLS for Vault connections. This is not recommended for production use.")
	}

	vaultClient := vault.NewVaultClient(vault.VaultConfig{
		URL:           c.VaultURL,
		Token:         c.VaultToken,
		BasePath:      c.VaultBasePath,
		InsecureHTTPS: c.VaultInsecureTLS,
	})

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read values file %s: %v", filename, err)
	}

	processed, err := vaultClient.ProcessYAMLWithVaultTemplates(content)
	if err != nil {
		return "", fmt.Errorf("failed to process vault templates in %s: %v", filename, err)
	}

	// Create a temporary file with processed content
	tempFile, err := os.CreateTemp("", "values-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	if _, err := tempFile.Write(processed); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to write processed values: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to close temp file: %v", err)
	}

	return tempFile.Name(), nil
}

// [Previous setupRootCA function remains unchanged]
func (c *Config) setupRootCA() error {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil
}

func (c *Config) Deploy() error {
	if c.Custom {
		return c.deployCustom()
	}
	return c.deployHelm()
}

// [Previous extractYAMLContent function remains unchanged]
func extractYAMLContent(helmOutput []byte) ([]byte, error) {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil, nil
}

// [Previous getDiff function remains unchanged]
func (c *Config) getDiff(args []string, isHelm bool) error {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil
}

// [Previous showResourceDiff function remains unchanged]
func (c *Config) showResourceDiff(current, proposed []byte) error {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil
}

// [Previous getTraefikDashboardArgs function remains unchanged]
func (c *Config) getTraefikDashboardArgs() []string {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil
}

// [Previous getRootCAArgs function remains unchanged]
func (c *Config) getRootCAArgs() []string {
	// [Previous implementation remains unchanged]
	// ... [Keep the existing implementation]
	return nil
}

func (c *Config) deployHelm() error {
	if err := c.setupRootCA(); err != nil {
		return err
	}

	var args []string
	args = append(args, "upgrade", "--install", c.ReleaseName, c.Chart)
	args = append(args, "--namespace", c.Namespace, "--create-namespace")

	// Add helm repo for all apps
	if err := exec.Command("helm", "repo", "add", c.AppName, c.Repository).Run(); err != nil {
		return fmt.Errorf("failed to add Helm repository: %v", err)
	}

	if err := exec.Command("helm", "repo", "update").Run(); err != nil {
		return fmt.Errorf("failed to update Helm repository: %v", err)
	}

	args = append(args, "--set", fmt.Sprintf("ingress.host=%s", c.IngressHost))

	// Process and add values files with Vault templating
	commonValuesFile := filepath.Join(c.ValuesPath, "common.yml")
	if _, err := os.Stat(commonValuesFile); err == nil {
		processedFile, err := c.processValuesFileWithVault(commonValuesFile)
		if err != nil {
			return err
		}
		if processedFile != commonValuesFile {
			defer os.Remove(processedFile)
		}
		args = append(args, "--values", processedFile)
	}

	stageValuesFile := filepath.Join(c.ValuesPath, fmt.Sprintf("%s.yml", c.Stage))
	if _, err := os.Stat(stageValuesFile); err == nil {
		processedFile, err := c.processValuesFileWithVault(stageValuesFile)
		if err != nil {
			return err
		}
		if processedFile != stageValuesFile {
			defer os.Remove(processedFile)
		}
		args = append(args, "--values", processedFile)
	}

	// Add version if specified
	if c.Version != "" {
		args = append(args, "--version", c.Version)
	}

	// Add Traefik dashboard args if applicable
	if strings.Contains(c.AppName, "traefik") {
		args = append(args, c.getTraefikDashboardArgs()...)
	}

	// Add root CA args
	args = append(args, c.getRootCAArgs()...)

	// Show diff first
	fmt.Println("Showing differences:")
	if err := c.getDiff(args, true); err != nil {
		return err
	}

	// Proceed with actual deployment
	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *Config) deployCustom() error {
	manifests, err := filepath.Glob(filepath.Join(c.ValuesPath, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to find manifests: %v", err)
	}

	// Process manifests with Vault templating
	processedManifests := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		processedFile, err := c.processValuesFileWithVault(manifest)
		if err != nil {
			return err
		}
		if processedFile != manifest {
			defer os.Remove(processedFile)
		}
		processedManifests = append(processedManifests, processedFile)
	}

	// Show diff first
	fmt.Println("Showing differences:")
	if err := c.getDiff(processedManifests, false); err != nil {
		return err
	}

	// Proceed with actual deployment
	for _, manifest := range processedManifests {
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
