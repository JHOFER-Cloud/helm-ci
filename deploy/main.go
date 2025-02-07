package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

type Config struct {
	Stage            string
	AppName          string
	Environment      string
	PRNumber         string
	ValuesPath       string
	Custom           bool
	Chart            string
	Version          string
	Repository       string
	Namespace        string
	ReleaseName      string
	IngressHost      string
	GitHubToken      string
	GitHubRepo       string
	GitHubOwner      string
	Domain           string
	TraefikDashboard bool
	RootCA           string
	PRDeployments    bool
	DEBUG            bool
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
	if c.Stage == "live" {
		c.Namespace = c.AppName
	} else {
		c.Namespace = c.AppName + "-dev"
	}

	if c.Stage == "dev" && c.PRNumber != "" && c.PRDeployments {
		c.ReleaseName = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		if c.Domain != "" {
			c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.Domain)
		}
	} else {
		c.ReleaseName = c.AppName
		if c.Domain != "" {
			c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.Domain)
		}
	}
}

func (c *Config) setupRootCA() error {
	if c.RootCA == "" {
		return nil
	}

	fmt.Printf("Setting up Root CA from: %s\n", c.RootCA)

	var certData []byte
	var err error

	// Check if RootCA is a URL
	if strings.HasPrefix(c.RootCA, "http://") || strings.HasPrefix(c.RootCA, "https://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(c.RootCA)
		if err != nil {
			return fmt.Errorf("failed to download root CA: %v", err)
		}
		defer resp.Body.Close()

		certData, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read root CA from URL: %v", err)
		}
	} else {
		certData, err = os.ReadFile(c.RootCA)
		if err != nil {
			return fmt.Errorf("failed to read root CA file: %v", err)
		}
	}

	// Create a temporary file to store the certificate data
	tmpFile, err := os.CreateTemp("", "root-ca-*.crt")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(certData); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	// Create namespace
	fmt.Printf("Creating namespace: %s\n", c.Namespace)
	var nsBuffer bytes.Buffer
	createNsCmd := exec.Command("kubectl", "create", "namespace", c.Namespace, "--dry-run=client", "-o", "yaml")
	createNsCmd.Stdout = &nsBuffer

	if err := createNsCmd.Run(); err != nil {
		return fmt.Errorf("failed to create namespace yaml: %v", err)
	}

	applyNsCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyNsCmd.Stdin = bytes.NewReader(nsBuffer.Bytes())

	if err := applyNsCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply namespace: %v", err)
	}

	// Create secret
	fmt.Printf("Creating CA secret in namespace: %s\n", c.Namespace)
	var secretBuffer bytes.Buffer
	secretCmd := exec.Command("kubectl", "create", "secret", "generic",
		"custom-root-ca",
		"--from-file=ca.crt="+tmpFile.Name(),
		"-n", c.Namespace,
		"--dry-run=client",
		"-o", "yaml")
	secretCmd.Stdout = &secretBuffer

	if err := secretCmd.Run(); err != nil {
		return fmt.Errorf("failed to create secret yaml: %v", err)
	}

	applySecretCmd := exec.Command("kubectl", "apply", "-f", "-")
	applySecretCmd.Stdin = bytes.NewReader(secretBuffer.Bytes())

	if err := applySecretCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply secret: %v", err)
	}

	fmt.Println("Root CA setup completed successfully")
	return nil
}

func (c *Config) Deploy() error {
	if c.Custom {
		return c.deployCustom()
	}
	return c.deployHelm()
}

func extractYAMLContent(helmOutput []byte) ([]byte, error) {
	lines := strings.Split(string(helmOutput), "\n")
	var yamlLines []string
	inManifest := false

	for _, line := range lines {
		if strings.HasPrefix(line, "MANIFEST:") {
			inManifest = true
			continue
		}
		if inManifest {
			if strings.Contains(line, "***") {
				continue
			}
			if strings.HasPrefix(line, "NOTES:") {
				break
			}
			yamlLines = append(yamlLines, line)
		}
	}

	return []byte(strings.Join(yamlLines, "\n")), nil
}

func (c *Config) getDiff(args []string, isHelm bool) error {
	if isHelm {
		// Get current state
		currentCmd := exec.Command("helm", "get", "manifest", c.ReleaseName, "-n", c.Namespace)
		current, err := currentCmd.Output()
		if err != nil {
			// If release doesn't exist, show what would be installed
			fmt.Println("No existing release found. Showing what would be installed:")

			// Add --dry-run flag to show what would be installed
			dryRunArgs := append(args, "--dry-run")
			cmd := exec.Command("helm", dryRunArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}

		// Get proposed state
		dryRunArgs := append(args, "--dry-run")
		proposedCmd := exec.Command("helm", dryRunArgs...)
		proposed, err := proposedCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get proposed state: %v", err)
		}

		// Extract YAML content from Helm output
		proposedYAML, err := extractYAMLContent(proposed)
		if err != nil {
			return fmt.Errorf("failed to extract YAML content: %v", err)
		}

		// Create diff using kubectl diff
		return c.showResourceDiff(current, proposedYAML)
	} else {
		// For kubectl, use kubectl diff directly
		for _, manifest := range args {
			cmd := exec.Command("kubectl", "diff", "-f", manifest, "-n", c.Namespace)
			output, err := cmd.Output()
			if err != nil {
				// Exit code 1 means differences were found, which is expected
				if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
					return fmt.Errorf("failed to get diff for %s: %v", manifest, err)
				}
			}

			if len(output) == 0 {
				// If no diff, show what would be applied
				fmt.Printf("\nNo existing resources found for %s. Showing what would be applied:\n", manifest)
				showCmd := exec.Command("kubectl", "apply", "-f", manifest, "-n", c.Namespace, "--dry-run=client", "-o", "yaml")
				showCmd.Stdout = os.Stdout
				showCmd.Stderr = os.Stderr
				if err := showCmd.Run(); err != nil {
					return fmt.Errorf("failed to show resources for %s: %v", manifest, err)
				}
			} else {
				fmt.Printf("\nDiff for %s:\n", manifest)
				fmt.Println(string(output))
			}
		}
	}
	return nil
}

func (c *Config) showResourceDiff(current, proposed []byte) error {
	// Create temporary files for diff
	currentFile, err := os.CreateTemp("", "current-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(currentFile.Name())

	proposedFile, err := os.CreateTemp("", "proposed-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(proposedFile.Name())

	// Write resources to temp files
	if err := os.WriteFile(currentFile.Name(), current, 0644); err != nil {
		return fmt.Errorf("failed to write current state: %v", err)
	}
	if err := os.WriteFile(proposedFile.Name(), proposed, 0644); err != nil {
		return fmt.Errorf("failed to write proposed state: %v", err)
	}

	if c.DEBUG {
		// Print the contents of the files for debugging
		fmt.Println("Current YAML:")
		fmt.Println(string(current))

		fmt.Println("Proposed YAML:")
		fmt.Println(string(proposed))
	}

	// Use diff to show differences with color
	diffCmd := exec.Command("kubectl", "diff", "-f", currentFile.Name(), "-f", proposedFile.Name())
	diffCmd.Stdout = os.Stdout
	diffCmd.Stderr = os.Stderr

	err = diffCmd.Run()
	if err != nil {
		// Exit code 1 means differences were found, which is expected
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			return fmt.Errorf("failed to generate diff: %v", err)
		}
	}

	return nil
}

func (c *Config) getTraefikDashboardArgs() []string {
	var args []string

	if c.TraefikDashboard {
		args = append(args,
			"--set", fmt.Sprintf("ingressRoute.dashboard.matchRule=Host(`%s`)", c.IngressHost),
			"--set", "ingressRoute.dashboard.entryPoints[0]=websecure",
		)
	}

	return args
}

// FIX: Mounting CA file not working (secret gets created)
// spec.template.spec.containers[0].volumeMounts[2].mountPath: Required value
func (c *Config) getRootCAArgs() []string {
	var args []string

	if c.RootCA != "" {
		args = append(args,
			"--set", "volumes[0].name=custom-root-ca",
			"--set", "volumes[0].secretName=custom-root-ca",
			"--set", "volumes[0].mountPath=/etc/ssl/certs",
			"--set", "volumes[0].subPath=ca.crt",
		)
	}

	return args
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

	if c.Domain != "" {
		args = append(args, "--set", fmt.Sprintf("ingress.host=%s", c.IngressHost))
	}

	// Add values files
	commonValuesFile := filepath.Join(c.ValuesPath, "common.yml")
	if _, err := os.Stat(commonValuesFile); err == nil {
		args = append(args, "--values", commonValuesFile)
	}

	stageValuesFile := filepath.Join(c.ValuesPath, fmt.Sprintf("%s.yml", c.Stage))
	if _, err := os.Stat(stageValuesFile); err == nil {
		args = append(args, "--values", stageValuesFile)
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

	// Show diff first
	fmt.Println("Showing differences:")
	if err := c.getDiff(manifests, false); err != nil {
		return err
	}

	// Proceed with actual deployment
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
