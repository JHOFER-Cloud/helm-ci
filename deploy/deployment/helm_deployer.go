// Copyright 2025 Josef Hofer (JHOFER-Cloud)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	commonValuesPattern := filepath.Join(d.Config.ValuesPath, "common.y*ml")
	matches, err := filepath.Glob(commonValuesPattern)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		processedFile, err := d.ProcessValuesFileWithVault(matches[0])
		if err != nil {
			return err
		}
		if processedFile != matches[0] {
			defer os.Remove(processedFile)
		}
		args = append(args, "--values", processedFile)
	}

	stageValuesPattern := filepath.Join(d.Config.ValuesPath, fmt.Sprintf("%s.y*ml", d.Config.Stage))
	matches, err = filepath.Glob(stageValuesPattern)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		processedFile, err := d.ProcessValuesFileWithVault(matches[0])
		if err != nil {
			return err
		}
		if processedFile != matches[0] {
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

	// Check if release exists
	existingReleaseCmd := d.Cmd.Command("helm", "get", "manifest", d.Config.ReleaseName, "-n", d.Config.Namespace)
	_, existingErr := d.Cmd.Output(existingReleaseCmd)
	isFirstInstall := existingErr != nil

	// First, try to detect if the chart contains CRDs
	hasCRDs := false

	if isFirstInstall {
		// Run a limited template command to check for CRDs
		templateArgs := []string{"template", d.Config.ReleaseName}

		// Add the chart reference
		if strings.HasPrefix(d.Config.Repository, "oci://") {
			templateArgs = append(templateArgs, fmt.Sprintf("%s/%s", d.Config.Repository, d.Config.Chart))
		} else {
			templateArgs = append(templateArgs, fmt.Sprintf("%s/%s", d.Config.AppName, d.Config.Chart))
		}

		templateArgs = append(templateArgs, "--include-crds", "--namespace", d.Config.Namespace)
		if d.Config.Version != "" {
			templateArgs = append(templateArgs, "--version", d.Config.Version)
		}

		// Use --dry-run for the template check
		templateArgs = append(templateArgs, "--dry-run")

		utils.Log.Info("Checking if chart contains CRDs...")
		templateCmd := d.Cmd.Command("helm", templateArgs...)
		output, err := d.Cmd.Output(templateCmd)

		// Check if the output contains CustomResourceDefinition
		if err == nil && strings.Contains(string(output), "CustomResourceDefinition") {
			utils.Log.Info("Chart contains CRDs. Skipping diff preview to avoid CRD issues.")
			hasCRDs = true
		}
	}

	// Try to show diff preview if not a first installation with CRDs
	if !isFirstInstall || !hasCRDs {
		utils.Green("Showing differences:")
		diffErr := d.GetDiff(args, true)

		// If diff fails and this is a first install, check if it's a CRD-related error
		if diffErr != nil && isFirstInstall {
			errMsg := diffErr.Error()
			if strings.Contains(errMsg, "no matches for kind") &&
				strings.Contains(errMsg, "ensure CRDs are installed first") {
				utils.Log.Warning("Diff failed due to missing CRDs. Proceeding directly with installation.")
				hasCRDs = true
			} else {
				// For non-CRD related errors, return the error
				return diffErr
			}
		} else if diffErr != nil {
			// For existing installations, any diff error is a real error
			return diffErr
		}

		// Check if we should proceed (unless we've determined we need to skip the diff)
		if !hasCRDs {
			if !utils.ConfirmDeployment(d.Config.DEBUG) {
				return utils.NewError("Deployment cancelled by user")
			}
		}
	}

	// For charts with CRDs on first install, we've already decided to proceed
	if hasCRDs && isFirstInstall {
		utils.Log.Info("Proceeding with installation (CRDs will be installed automatically)...")
		if !utils.ConfirmDeployment(d.Config.DEBUG) {
			return utils.NewError("Deployment cancelled by user")
		}
	}

	// Proceed with actual deployment
	cmd := d.Cmd.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return d.Cmd.Run(cmd)
}
