package deployment

import (
	"fmt"
	"helm-ci/deploy/templates"
	"helm-ci/deploy/utils"
	"os"
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

	if d.Config.TraefikDashboard {
		if len(d.Config.IngressHosts) > 0 {
			// Create a Host expression with all domains
			hostRules := make([]string, len(d.Config.IngressHosts))
			for i, host := range d.Config.IngressHosts {
				hostRules[i] = fmt.Sprintf("Host(`%s`)", host)
			}
			hostExpression := strings.Join(hostRules, " || ")

			args = append(args,
				"--set", fmt.Sprintf("ingressRoute.dashboard.matchRule=%s", hostExpression),
				"--set", "ingressRoute.dashboard.entryPoints[0]=websecure",
			)
		} else {
			// If no domains are specified, don't set a host rule
			utils.Log.Warning("Traefik dashboard enabled but no domains specified")
			args = append(args,
				"--set", "ingressRoute.dashboard.entryPoints[0]=websecure",
			)
		}
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
	args = append(args, "upgrade", "--install", d.Config.ReleaseName)

	// Check if the repository is an OCI registry
	if strings.HasPrefix(d.Config.Repository, "oci://") {
		args = append(args, fmt.Sprintf("%s/%s", d.Config.Repository, d.Config.Chart))
	} else {
		args = append(args, fmt.Sprintf("%s/%s", d.Config.AppName, d.Config.Chart))
		// Add helm repo for all apps
		repoAddCmd := d.Cmd.Command("helm", "repo", "add", d.Config.AppName, d.Config.Repository)
		if err := d.Cmd.Run(repoAddCmd); err != nil {
			return utils.NewError("failed to add Helm repository: %v", err)
		}

		repoUpdateCmd := d.Cmd.Command("helm", "repo", "update")
		if err := d.Cmd.Run(repoUpdateCmd); err != nil {
			return utils.NewError("failed to update Helm repository: %v", err)
		}
	}

	args = append(args, "--namespace", d.Config.Namespace, "--create-namespace")

	// Process domain template if domains are specified
	if len(d.Config.Domains) > 0 {
		domainValuesFile, err := templates.ProcessDomainTemplate(d.Config)
		if err != nil {
			return err
		}
		if domainValuesFile != "" {
			defer os.Remove(domainValuesFile)
			args = append(args, "--values", domainValuesFile)
		}
	}

	// Process and add values files with Vault templating
	commonValuesFile := filepath.Join(d.Config.ValuesPath, "common.yml")
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

	stageValuesFile := filepath.Join(d.Config.ValuesPath, fmt.Sprintf("%s.yml", d.Config.Stage))
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
	if d.Config.Version != "" {
		args = append(args, "--version", d.Config.Version)
	}

	// Add Traefik dashboard args if applicable
	if strings.Contains(d.Config.AppName, "traefik") {
		args = append(args, d.GetTraefikDashboardArgs()...)
	}

	// Add root CA args
	args = append(args, d.GetRootCAArgs()...)

	// Show diff first
	utils.Green("Showing differences:")
	if err := d.GetDiff(args, true); err != nil {
		return err
	}

	// Check if we should proceed
	if !utils.ConfirmDeployment(d.Config.DEBUG) {
		return utils.NewError("Deployment cancelled by user")
	}

	// Proceed with actual deployment
	cmd := d.Cmd.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return d.Cmd.Run(cmd)
}
