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

package main

import (
	"helm-ci/deploy/config"
	"helm-ci/deploy/deployment"
	"helm-ci/deploy/utils"
	"os"
)

func main() {
	// Parse command line flags
	cfg := config.ParseFlags()

	// Setup namespace and release name
	cfg.SetupNames()

	// Print configuration
	cfg.PrintConfig()

	// Initialize logger
	utils.InitLogger(cfg.DEBUG)

	// Create a common instance with the config
	common := deployment.NewCommon(cfg)

	// Create deployer based on config
	var deployer deployment.Deployer
	if cfg.Custom {
		deployer = &deployment.CustomDeployer{Common: common}
	} else {
		deployer = &deployment.HelmDeployer{Common: common}
	}

	// Run deployment
	if err := deployer.Deploy(); err != nil {
		utils.NewError("Deployment failed: %v\n", err)
		os.Exit(1)
	}

	utils.Success("Deployment succeeded")
}
