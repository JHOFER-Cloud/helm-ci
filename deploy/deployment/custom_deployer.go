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
// FIX: metadata.namespace in manifest file has to be set/replaced
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
	stageManifests, err := filepath.Glob(filepath.Join(d.Config.ValuesPath, d.Config.Stage, "*.y*ml"))
	if err != nil {
		return utils.NewError("failed to glob stage manifests: %w", err)
	}

	commonManifests, err := filepath.Glob(filepath.Join(d.Config.ValuesPath, "common", "*.y*ml"))
	if err != nil {
		return utils.NewError("failed to glob common manifests: %w", err)
	}
	manifests := append(stageManifests, commonManifests...)

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
	cmd := d.Cmd.Command("kubectl", "get", "namespace", d.Config.Namespace)
	if err := d.Cmd.Run(cmd); err != nil {
		utils.Green("Namespace %s does not exist, creating it...", d.Config.Namespace)
		cmd = d.Cmd.Command("kubectl", "create", "namespace", d.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to create namespace %s: %v", d.Config.Namespace, err)
		}
	}

	// Show diff first
	utils.Green("Showing differences:")
	if err := d.GetDiff(processedManifests, false); err != nil {
		return err
	}

	// Check if we should proceed
	if !utils.ConfirmDeployment(d.Config.DEBUG) {
		return utils.NewError("Deployment cancelled by user")
	}

	// Proceed with actual deployment
	for _, manifest := range processedManifests {
		cmd := d.Cmd.Command("kubectl", "apply", "-f", manifest, "-n", d.Config.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := d.Cmd.Run(cmd); err != nil {
			return utils.NewError("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}
