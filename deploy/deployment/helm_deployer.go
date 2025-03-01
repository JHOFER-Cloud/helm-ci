package deployment

import (
	"fmt"
	"helm-ci/deploy/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HelmDeployer implements Helm-based deployments
type HelmDeployer struct {
	Common
}

// GetTraefikDashboardArgs returns arguments for Traefik dashboard
func (d *HelmDeployer) GetTraefikDashboardArgs() []string {
	var args []string

	if d.Common.Config.TraefikDashboard {
		args = append(args,
			"--set", fmt.Sprintf("ingressRoute.dashboard.matchRule=Host(`%s`)", d.Common.Config.IngressHost),
			"--set", "ingressRoute.dashboard.entryPoints[0]=websecure",
		)
	}

	return args
}

// GetRootCAArgs returns arguments for root CA
func (d *HelmDeployer) GetRootCAArgs() []string {
	var args []string
	// Note: Current implementation is commented out in the original code
	// if d.Common.Config.RootCA != "" {
	//     args = append(args,
	//         "--set", "volumes[0].name=custom-root-ca",
	//         "--set", "volumes[0].secretName=custom-root-ca",
	//         "--set", "volumes[0].mountPath=/etc/ssl/certs",
	//         "--set", "volumes[0].subPath=ca.crt",
	//     )
	// }
	return args
}

// Deploy implements the Helm deployment
func (d *HelmDeployer) Deploy() error {
	if err := d.SetupRootCA(); err != nil {
		return err
	}

	var args []string
	args = append(args, "upgrade", "--install", d.Common.Config.ReleaseName)

	// Check if the repository is an OCI registry
	if strings.HasPrefix(d.Common.Config.Repository, "oci://") {
		args = append(args, fmt.Sprintf("%s/%s", d.Common.Config.Repository, d.Common.Config.Chart))
	} else {
		args = append(args, fmt.Sprintf("%s/%s", d.Common.Config.AppName, d.Common.Config.Chart))
		// Add helm repo for all apps
		if err := exec.Command("helm", "repo", "add", d.Common.Config.AppName, d.Common.Config.Repository).Run(); err != nil {
			return utils.NewError("failed to add Helm repository: %v", err)
		}

		if err := exec.Command("helm", "repo", "update").Run(); err != nil {
			return utils.NewError("failed to update Helm repository: %v", err)
		}
	}

	args = append(args, "--namespace", d.Common.Config.Namespace, "--create-namespace")

	if d.Common.Config.Domain != "" {
		if strings.Contains(d.Common.Config.AppName, "vault") {
			args = append(args, "--set", fmt.Sprintf("server.ingress.hosts[0].host=%s", d.Common.Config.IngressHost))
		} else {
			args = append(args, "--set", fmt.Sprintf("ingress.host=%s", d.Common.Config.IngressHost))
		}
	}

	// Process and add values files with Vault templating
	commonValuesFile := filepath.Join(d.Common.Config.ValuesPath, "common.yml")
	if _, err := os.Stat(commonValuesFile); err == nil {
		processedFile, err := d.ProcessValuesFileWithVault(commonValuesFile)
		if err != nil {
			return err
		}
		if processedFile != commonValuesFile {
			defer os.Remove(processedFile)
		}
		args = append(args, "--values", processedFile)
	}

	stageValuesFile := filepath.Join(d.Common.Config.ValuesPath, fmt.Sprintf("%s.yml", d.Common.Config.Stage))
	if _, err := os.Stat(stageValuesFile); err == nil {
		processedFile, err := d.ProcessValuesFileWithVault(stageValuesFile)
		if err != nil {
			return err
		}
		if processedFile != stageValuesFile {
			defer os.Remove(processedFile)
		}
		args = append(args, "--values", processedFile)
	}

	// Add version if specified
	if d.Common.Config.Version != "" {
		args = append(args, "--version", d.Common.Config.Version)
	}

	// Add Traefik dashboard args if applicable
	if strings.Contains(d.Common.Config.AppName, "traefik") {
		args = append(args, d.GetTraefikDashboardArgs()...)
	}

	// Add root CA args
	args = append(args, d.GetRootCAArgs()...)

	// Show diff first
	utils.Green("Showing differences:")
	if err := d.GetDiff(args, true); err != nil {
		return err
	}

	// Proceed with actual deployment
	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
