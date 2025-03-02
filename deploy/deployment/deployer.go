package deployment

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"helm-ci/deploy/config"
	"helm-ci/deploy/utils"
	"helm-ci/deploy/vault"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// Deployer interface for different deployment strategies
type Deployer interface {
	Deploy() error
}

// Common contains shared functionality for all deployers
type Common struct {
	Config *config.Config
	Cmd    Commander
}

// NewCommon creates a new Common with default configuration
func NewCommon(cfg *config.Config) Common {
	return Common{
		Config: cfg,
		Cmd:    &RealCommander{},
	}
}

// ProcessValuesFileWithVault processes a values file with Vault templating
func (c *Common) ProcessValuesFileWithVault(filename string) (string, error) {
	// If Vault URL is not configured, return the original file
	if c.Config.VaultURL == "" {
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
		c.Config.VaultURL,
		c.Config.VaultToken,
		c.Config.VaultBasePath,
		c.Config.VaultKVVersion,
		c.Config.VaultInsecureTLS,
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

	if c.Config.DEBUG {
		utils.Log.Debugln("Processed content:")
		fmt.Println(processedContent)
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

// SetupRootCA sets up the root CA certificate
func (c *Common) SetupRootCA() error {
	if c.Config.RootCA == "" {
		return nil
	}

	utils.Log.Infof("Setting up Root CA from: %s\n", c.Config.RootCA)

	var certData []byte
	var err error

	// Check if RootCA is a URL
	if strings.HasPrefix(c.Config.RootCA, "http://") || strings.HasPrefix(c.Config.RootCA, "https://") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(c.Config.RootCA)
		if err != nil {
			return utils.NewError("failed to download root CA: %v", err)
		}
		defer resp.Body.Close()

		certData, err = io.ReadAll(resp.Body)
		if err != nil {
			return utils.NewError("failed to read root CA from URL: %v", err)
		}
	} else {
		certData, err = os.ReadFile(c.Config.RootCA)
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
	utils.Log.Infof("Creating namespace: %s\n", c.Config.Namespace)
	var nsBuffer bytes.Buffer
	createNsCmd := c.Cmd.Command("kubectl", "create", "namespace", c.Config.Namespace, "--dry-run=client", "-o", "yaml")
	createNsCmd.Stdout = &nsBuffer

	if err := c.Cmd.Run(createNsCmd); err != nil {
		return utils.NewError("failed to create namespace yaml: %v", err)
	}

	applyNsCmd := c.Cmd.Command("kubectl", "apply", "-f", "-")
	applyNsCmd.Stdin = bytes.NewReader(nsBuffer.Bytes())

	if err := c.Cmd.Run(applyNsCmd); err != nil {
		return utils.NewError("failed to apply namespace: %v", err)
	}

	// Create secret
	utils.Log.Infof("Creating CA secret in namespace: %s\n", c.Config.Namespace)
	var secretBuffer bytes.Buffer
	secretCmd := c.Cmd.Command("kubectl", "create", "secret", "generic",
		"custom-root-ca",
		"--from-file=ca.crt="+tmpFile.Name(),
		"-n", c.Config.Namespace,
		"--dry-run=client",
		"-o", "yaml")
	secretCmd.Stdout = &secretBuffer

	if err := c.Cmd.Run(secretCmd); err != nil {
		return utils.NewError("failed to create secret yaml: %v", err)
	}

	applySecretCmd := c.Cmd.Command("kubectl", "apply", "-f", "-")
	applySecretCmd.Stdin = bytes.NewReader(secretBuffer.Bytes())

	if err := c.Cmd.Run(applySecretCmd); err != nil {
		return utils.NewError("failed to apply secret: %v", err)
	}

	utils.Success("Root CA setup completed successfully")
	return nil
}

// ExtractYAMLContent extracts YAML content from helm output
func (c *Common) ExtractYAMLContent(helmOutput []byte) ([]byte, error) {
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

// GetDiff gets the diff between current and proposed state
func (c *Common) GetDiff(args []string, isHelm bool) error {
	if isHelm {
		currentCmd := c.Cmd.Command("helm", "get", "manifest", c.Config.ReleaseName, "-n", c.Config.Namespace)
		current, err := c.Cmd.Output(currentCmd)
		if err != nil {
			utils.Log.Info("No existing release found. Showing what would be installed:")
			dryRunArgs := append(args, "--dry-run")
			cmd := c.Cmd.Command("helm", dryRunArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := c.Cmd.Run(cmd); err != nil {
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
		proposedCmd := c.Cmd.Command("helm", dryRunArgs...)
		proposed, err := c.Cmd.Output(proposedCmd)
		if err != nil {
			// Add more detailed error information
			utils.Log.Errorf("Failed to get proposed state: %v", err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				utils.Log.Errorf("Stderr: %s", string(exitErr.Stderr))
			}
			return utils.NewError("failed to get proposed state: %w", err)
		}

		proposedYAML, err := c.ExtractYAMLContent(proposed)
		if err != nil {
			return utils.NewError("failed to extract YAML content: %v", err)
		}

		return utils.ShowResourceDiff(current, proposedYAML, c.Config.DEBUG)
	} else {
		for _, manifest := range args {
			cmd := c.Cmd.Command("kubectl", "diff", "-f", manifest, "-n", c.Config.Namespace)
			output, err := c.Cmd.CombinedOutput(cmd)

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
