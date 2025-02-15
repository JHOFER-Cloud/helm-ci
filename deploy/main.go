package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"helm-ci/deploy/utils"
	"helm-ci/deploy/vault"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppName          string
	Chart            string
	Custom           bool
	CustomNameSpace  string
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
	VaultInsecureTLS bool
	VaultToken       string
	VaultURL         string
	Version          string
	VaultKVVersion   int
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
	flag.StringVar(&cfg.CustomNameSpace, "custom-namespace", "", "Custom K8s Namespace")
	flag.BoolVar(&cfg.Custom, "custom", false, "Custom Kubernetes deployment")
	flag.BoolVar(&cfg.TraefikDashboard, "traefik-dashboard", false, "Deploy Traefik dashboard")
	flag.StringVar(&cfg.RootCA, "root-ca", "", "Path to root CA certificate")
	flag.BoolVar(&cfg.PRDeployments, "pr-deployments", true, "Enable PR deployments")
	flag.StringVar(&cfg.VaultURL, "vault-url", "", "Vault server URL")
	flag.StringVar(&cfg.VaultToken, "vault-token", os.Getenv("VAULT_TOKEN"), "Vault authentication token")
	flag.StringVar(&cfg.VaultBasePath, "vault-base-path", "", "Base path for Vault secrets")
	flag.BoolVar(&cfg.VaultInsecureTLS, "vault-insecure-tls", false, "Allow insecure TLS connections to Vault (not recommended for production)")
	flag.IntVar(&cfg.VaultKVVersion, "vault-kv-version", 2, "Vault KV version (1 or 2)")
	flag.BoolVar(&cfg.DEBUG, "debug", false, "DEBUG output; THIS MAY OUTPUT SECRETS!!!")
	flag.Parse()

	if cfg.AppName == "" {
		utils.NewError("app name is required")
		os.Exit(1)
	}

	if cfg.Stage == "" {
		utils.NewError("stage is required")
		os.Exit(1)
	}

	if cfg.Environment == "" {
		utils.NewError("environment is required")
		os.Exit(1)
	}

	return cfg
}

func (c *Config) PrintConfig() {
	utils.Log.Info("Current Configuration:")
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		// Don't print sensitive values
		if fieldName == "VaultToken" || fieldName == "GitHubToken" {
			utils.Log.Info(fmt.Sprintf("%s: [REDACTED]", fieldName))
		} else {
			utils.Log.Info(fmt.Sprintf("%s: %v", fieldName, field.Interface()))
		}
	}
}

func (c *Config) setupNames() {
	if c.CustomNameSpace != "" {
		c.Namespace = c.CustomNameSpace
	} else if c.Stage == "live" {
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

func (c *Config) processValuesFileWithVault(filename string) (string, error) {
	// If Vault URL is not configured, return the original file
	if c.VaultURL == "" {
		utils.Log.Debug("No Vault URL configured, using original values file")
		return filename, nil
	}

	// Read the original values file
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", utils.NewError("failed to read values file %s: %v", filename, err)
	}

	// Create Vault client
	vaultClient, err := vault.NewClient(
		c.VaultURL,
		c.VaultToken,
		c.VaultBasePath,
		c.VaultKVVersion,
		c.VaultInsecureTLS,
	)
	if err != nil {
		return "", utils.NewError("failed to initialize vault client: %w", err)
	}

	// Process the content using the new method
	processedContent, err := vaultClient.ProcessString(string(content))
	if err != nil {
		return "", utils.NewError("failed to process vault templates in file %s: %w", filename, err)
	}

	// Check if this is a Kubernetes Secret
	if strings.Contains(processedContent, "kind: Secret") {
		// Parse the YAML
		var secret map[string]interface{}
		err = yaml.Unmarshal([]byte(processedContent), &secret)
		if err != nil {
			return "", utils.NewError("failed to parse Secret YAML: %v", err)
		}

		// Get the data section
		if data, ok := secret["data"].(map[string]interface{}); ok {
			// Base64 encode each value
			for k, v := range data {
				if str, ok := v.(string); ok {
					data[k] = base64.StdEncoding.EncodeToString([]byte(str))
				}
			}
			// Update the secret with encoded values
			secret["data"] = data
		}

		// Convert back to YAML
		yamlBytes, err := yaml.Marshal(secret)
		if err != nil {
			return "", utils.NewError("failed to marshal Secret YAML: %v", err)
		}
		processedContent = string(yamlBytes)
	}

	if c.DEBUG {
		utils.Log.Debugln("Processed content:")
		fmt.Println(processedContent)

		utils.Green("Looks good, deploy now? (Y/n): ")
		var response string
		fmt.Scanln(&response)

		if response == "n" || response == "N" {
			os.Exit(1)
		}
	}

	// Create a temporary file for the processed values
	tmpFile, err := os.CreateTemp("", "values-*.yml")
	if err != nil {
		return "", utils.NewError("failed to create temporary file: %v", err)
	}

	// Write the processed content to the temporary file
	if err := os.WriteFile(tmpFile.Name(), []byte(processedContent), 0644); err != nil {
		os.Remove(tmpFile.Name()) // Clean up the temp file if write fails
		return "", utils.NewError("failed to write processed values: %v", err)
	}

	utils.Log.Infof("Successfully processed values file: %s", tmpFile.Name())
	return tmpFile.Name(), nil
}

func (c *Config) setupRootCA() error {
	if c.RootCA == "" {
		return nil
	}

	utils.Log.Infof("Setting up Root CA from: %s\n", c.RootCA)

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
			return utils.NewError("failed to download root CA: %v", err)
		}
		defer resp.Body.Close()

		certData, err = io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewError("failed to read root CA from URL: %v", err)
		}
	} else {
		certData, err = os.ReadFile(c.RootCA)
		if err != nil {
			return utils.NewError("failed to read root CA file: %v", err)
		}
	}

	// Create a temporary file to store the certificate data
	tmpFile, err := os.CreateTemp("", "root-ca-*.crt")
	if err != nil {
		return utils.NewError("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(certData); err != nil {
		return utils.NewError("failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return utils.NewError("failed to close temporary file: %v", err)
	}

	// Create namespace
	utils.Log.Infof("Creating namespace: %s\n", c.Namespace)
	var nsBuffer bytes.Buffer
	createNsCmd := exec.Command("kubectl", "create", "namespace", c.Namespace, "--dry-run=client", "-o", "yaml")
	createNsCmd.Stdout = &nsBuffer

	if err := createNsCmd.Run(); err != nil {
		return utils.NewError("failed to create namespace yaml: %v", err)
	}

	applyNsCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyNsCmd.Stdin = bytes.NewReader(nsBuffer.Bytes())

	if err := applyNsCmd.Run(); err != nil {
		return utils.NewError("failed to apply namespace: %v", err)
	}

	// Create secret
	utils.Log.Infof("Creating CA secret in namespace: %s\n", c.Namespace)
	var secretBuffer bytes.Buffer
	secretCmd := exec.Command("kubectl", "create", "secret", "generic",
		"custom-root-ca",
		"--from-file=ca.crt="+tmpFile.Name(),
		"-n", c.Namespace,
		"--dry-run=client",
		"-o", "yaml")
	secretCmd.Stdout = &secretBuffer

	if err := secretCmd.Run(); err != nil {
		return utils.NewError("failed to create secret yaml: %v", err)
	}

	applySecretCmd := exec.Command("kubectl", "apply", "-f", "-")
	applySecretCmd.Stdin = bytes.NewReader(secretBuffer.Bytes())

	if err := applySecretCmd.Run(); err != nil {
		return utils.NewError("failed to apply secret: %v", err)
	}

	utils.Success("Root CA setup completed successfully")
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
		currentCmd := exec.Command("helm", "get", "manifest", c.ReleaseName, "-n", c.Namespace)
		current, err := currentCmd.Output()
		if err != nil {
			utils.Log.Info("No existing release found. Showing what would be installed:")
			dryRunArgs := append(args, "--dry-run")
			cmd := exec.Command("helm", dryRunArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// Add more detailed error information
				utils.Log.Errorf("Dry-run failed: %v", err)
				if exitErr, ok := err.(*exec.ExitError); ok {
					utils.Log.Errorf("Stderr: %s", string(exitErr.Stderr))
				}
				return fmt.Errorf("failed to get proposed state: %w", err)
			}
			return nil
		}

		dryRunArgs := append(args, "--dry-run")
		proposedCmd := exec.Command("helm", dryRunArgs...)
		proposed, err := proposedCmd.Output()
		if err != nil {
			// Add more detailed error information
			utils.Log.Errorf("Failed to get proposed state: %v", err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				utils.Log.Errorf("Stderr: %s", string(exitErr.Stderr))
			}
			return utils.NewError("failed to get proposed state: %w", err)
		}

		proposedYAML, err := extractYAMLContent(proposed)
		if err != nil {
			return utils.NewError("failed to extract YAML content: %v", err)
		}

		return utils.ShowResourceDiff(current, proposedYAML, c.DEBUG)
	} else {
		for _, manifest := range args {
			cmd := exec.Command("kubectl", "diff", "-f", manifest, "-n", c.Namespace)
			output, err := cmd.CombinedOutput()

			utils.Green("\nDiff for %s:\n", manifest)
			fmt.Println(utils.ColorizeKubectlDiff(string(output)))

			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
					return utils.NewError("failed to get diff for %s: %v", manifest, err)
				}
			}
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
// add logic that looks for a free index instead of overwriting [0]
func (c *Config) getRootCAArgs() []string {
	var args []string

	if c.RootCA != "" {
		// args = append(args,
		// 	"--set", "volumes[0].name=custom-root-ca",
		// 	"--set", "volumes[0].secretName=custom-root-ca",
		// 	"--set", "volumes[0].mountPath=/etc/ssl/certs",
		// 	"--set", "volumes[0].subPath=ca.crt",
		// )
	}

	return args
}

func (c *Config) deployHelm() error {
	if err := c.setupRootCA(); err != nil {
		return err
	}

	var args []string
	args = append(args, "upgrade", "--install", c.ReleaseName)

	// Check if the repository is an OCI registry
	if strings.HasPrefix(c.Repository, "oci://") {
		args = append(args, fmt.Sprintf("%s/%s", c.Repository, c.Chart))
	} else {
		args = append(args, fmt.Sprintf("%s/%s", c.AppName, c.Chart))
		// Add helm repo for all apps
		if err := exec.Command("helm", "repo", "add", c.AppName, c.Repository).Run(); err != nil {
			return utils.NewError("failed to add Helm repository: %v", err)
		}

		if err := exec.Command("helm", "repo", "update").Run(); err != nil {
			return utils.NewError("failed to update Helm repository: %v", err)
		}
	}

	args = append(args, "--namespace", c.Namespace, "--create-namespace")

	if c.Domain != "" {
		if strings.Contains(c.AppName, "vault") {
			args = append(args, "--set", fmt.Sprintf("server.ingress.hosts[0].host=%s", c.IngressHost))
		} else {
			args = append(args, "--set", fmt.Sprintf("ingress.host=%s", c.IngressHost))
		}
	}
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
	utils.Green("Showing differences:")
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
		return utils.NewError("failed to find manifests: %v", err)
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

	// Check if namespace exists, create if it doesn't
	cmd := exec.Command("kubectl", "get", "namespace", c.Namespace)
	if err := cmd.Run(); err != nil {
		utils.Green("Namespace %s does not exist, creating it...", c.Namespace)
		cmd = exec.Command("kubectl", "create", "namespace", c.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return utils.NewError("failed to create namespace %s: %v", c.Namespace, err)
		}
	}

	// Show diff first
	utils.Green("Showing differences:")
	if err := c.getDiff(processedManifests, false); err != nil {
		return err
	}

	// Proceed with actual deployment
	for _, manifest := range processedManifests {
		cmd := exec.Command("kubectl", "apply", "-f", manifest, "-n", c.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return utils.NewError("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}

func main() {
	cfg := parseFlags()
	cfg.setupNames()
	cfg.PrintConfig()
	utils.InitLogger(cfg.DEBUG)

	if err := cfg.Deploy(); err != nil {
		utils.NewError("Deployment failed: %v\n", err)
		os.Exit(1)
	}

	utils.Success("Deployment succeeded")
}
