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

	// Create deployer based on config
	var deployer deployment.Deployer
	if cfg.Custom {
		deployer = &deployment.CustomDeployer{Common: deployment.Common{Config: cfg}}
	} else {
		deployer = &deployment.HelmDeployer{Common: deployment.Common{Config: cfg}}
	}

	// Run deployment
	if err := deployer.Deploy(); err != nil {
		utils.NewError("Deployment failed: %v\n", err)
		os.Exit(1)
	}

	utils.Success("Deployment succeeded")
}
