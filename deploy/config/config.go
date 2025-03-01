package config

import (
	"flag"
	"fmt"
	"helm-ci/deploy/utils"
	"os"
	"reflect"
)

// Config holds all the configuration for the deployment
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

// ParseFlags parses command line flags and returns a Config
func ParseFlags() *Config {
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

	// Validate required flags
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

// PrintConfig prints the current configuration
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

// SetupNames configures namespace and release name based on the config
func (c *Config) SetupNames() {
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
