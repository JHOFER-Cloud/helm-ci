package deployment

import (
	"helm-ci/deploy/utils"
	"os"
	"path/filepath"
)

// CustomDeployer implements custom Kubernetes manifest deployments
type CustomDeployer struct {
	Common
}

// Deploy implements the custom deployment
func (d *CustomDeployer) Deploy() error {
	manifests, err := filepath.Glob(filepath.Join(d.Common.Config.ValuesPath, "*.yml"))
	if err != nil {
		return utils.NewError("failed to find manifests: %v", err)
	}

	// Process manifests with Vault templating
	processedManifests := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		processedFile, err := d.ProcessValuesFileWithVault(manifest)
		if err != nil {
			return err
		}
		if processedFile != manifest {
			defer os.Remove(processedFile)
		}
		processedManifests = append(processedManifests, processedFile)
	}

	// Check if namespace exists, create if it doesn't
	cmd := d.Cmd.Command("kubectl", "get", "namespace", d.Common.Config.Namespace)
	if err := d.Cmd.Run(cmd); err != nil {
		utils.Green("Namespace %s does not exist, creating it...", d.Common.Config.Namespace)
		cmd = d.Cmd.Command("kubectl", "create", "namespace", d.Common.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to create namespace %s: %v", d.Common.Config.Namespace, err)
		}
	}

	// Show diff first
	utils.Green("Showing differences:")
	if err := d.GetDiff(processedManifests, false); err != nil {
		return err
	}

	// Proceed with actual deployment
	for _, manifest := range processedManifests {
		cmd := d.Cmd.Command("kubectl", "apply", "-f", manifest, "-n", d.Common.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}
