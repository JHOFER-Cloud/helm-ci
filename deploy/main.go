package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
)

type Config struct {
	Stage            string
	AppName          string
	Environment      string
	PRNumber         string
	ValuesPath       string
	Custom           bool
	Chart            string
	Version          string
	Repository       string
	Namespace        string
	ReleaseName      string
	IngressHost      string
	GitHubToken      string
	GitHubRepo       string
	GitHubOwner      string
	Domain           string
	TraefikDashboard bool
	RootCA           string
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
	flag.BoolVar(&cfg.Custom, "custom", false, "Custom Kubernetes deployment")
	flag.BoolVar(&cfg.TraefikDashboard, "traefik-dashboard", false, "Deploy Traefik dashboard")
	flag.StringVar(&cfg.RootCA, "root-ca", "", "Path to root CA certificate")
	flag.Parse()

	if cfg.AppName == "" {
		fmt.Println("app name is required")
		os.Exit(1)
	}

	if cfg.Stage == "" {
		fmt.Println("stage is required")
		os.Exit(1)
	}

	if cfg.Environment == "" {
		fmt.Println("environment is required")
		os.Exit(1)
	}

	return cfg
}

func (c *Config) PrintConfig() {
	fmt.Println("Current Configuration:")
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name
		fmt.Printf("- %s: %v\n", fieldName, field.Interface())
	}
}

func (c *Config) setupNames() {
	if c.Stage == "live" {
		c.Namespace = c.AppName
	} else {
		c.Namespace = c.AppName + "-dev"
	}

	if c.Stage == "dev" && c.PRNumber != "" {
		c.ReleaseName = fmt.Sprintf("%s-pr-%s", c.AppName, c.PRNumber)
		c.IngressHost = fmt.Sprintf("%s-pr-%s.%s", c.AppName, c.PRNumber, c.Domain)
	} else {
		c.ReleaseName = c.AppName
		c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.Domain)
	}
}

func (c *Config) setupRootCA() error { // FIX: This is not working!!
	if c.RootCA == "" {
		return nil
	}

	fmt.Printf("Setting up Root CA from: %s\n", c.RootCA)

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
			return fmt.Errorf("failed to download root CA: %v", err)
		}
		defer resp.Body.Close()

		certData, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read root CA from URL: %v", err)
		}
	} else {
		certData, err = os.ReadFile(c.RootCA)
		if err != nil {
			return fmt.Errorf("failed to read root CA file: %v", err)
		}
	}

	// Create namespace
	fmt.Printf("Creating namespace: %s\n", c.Namespace)
	var nsBuffer bytes.Buffer
	createNsCmd := exec.Command("kubectl", "create", "namespace", c.Namespace, "--dry-run=client", "-o", "yaml")
	createNsCmd.Stdout = &nsBuffer

	if err := createNsCmd.Run(); err != nil {
		return fmt.Errorf("failed to create namespace yaml: %v", err)
	}

	applyNsCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyNsCmd.Stdin = bytes.NewReader(nsBuffer.Bytes())

	if err := applyNsCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply namespace: %v", err)
	}

	// Create secret
	fmt.Printf("Creating CA secret in namespace: %s\n", c.Namespace)
	var secretBuffer bytes.Buffer
	secretCmd := exec.Command("kubectl", "create", "secret", "generic",
		"custom-root-ca",
		"--from-literal=ca.crt="+string(certData),
		"-n", c.Namespace,
		"--dry-run=client",
		"-o", "yaml")
	secretCmd.Stdout = &secretBuffer

	if err := secretCmd.Run(); err != nil {
		return fmt.Errorf("failed to create secret yaml: %v", err)
	}

	applySecretCmd := exec.Command("kubectl", "apply", "-f", "-")
	applySecretCmd.Stdin = bytes.NewReader(secretBuffer.Bytes())

	if err := applySecretCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply secret: %v", err)
	}

	fmt.Println("Root CA setup completed successfully")
	return nil
}

func (c *Config) Deploy() error {
	if c.Custom {
		return c.deployCustom()
	}
	return c.deployHelm()
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

func (c *Config) getRootCAArgs() []string {
	var args []string

	if c.RootCA != "" {
		if strings.Contains(c.AppName, "traefik") {
			args = append(args,
				"--set", "additionalArguments[0]=--serverstransport.rootcas=/usr/local/share/ca-certificates/ca.crt",
				"--set", "additionalVolumes[0].name=custom-root-ca",
				"--set", "additionalVolumes[0].secret.secretName=custom-root-ca",
				"--set", "additionalVolumeMounts[0].name=custom-root-ca",
				"--set", "additionalVolumeMounts[0].mountPath=/usr/local/share/ca-certificates/ca.crt",
				"--set", "additionalVolumeMounts[0].subPath=ca.crt",
			)
		} else {
			args = append(args,
				"--set", "extraVolumes[0].name=custom-root-ca",
				"--set", "extraVolumes[0].secret.secretName=custom-root-ca",
				"--set", "extraVolumeMounts[0].name=custom-root-ca",
				"--set", "extraVolumeMounts[0].mountPath=/etc/ssl/certs/ca.crt",
				"--set", "extraVolumeMounts[0].subPath=ca.crt",
			)
		}
	}

	return args
}

func (c *Config) deployHelm() error {
	if err := c.setupRootCA(); err != nil {
		return err
	}

	var args []string
	// Start with the basic command
	args = append(args, "upgrade", "--install", c.ReleaseName, c.Chart)
	args = append(args, "--namespace", c.Namespace, "--create-namespace")

	// Add helm repo for all apps
	if err := exec.Command("helm", "repo", "add", c.AppName, c.Repository).Run(); err != nil {
		return fmt.Errorf("failed to add Helm repository: %v", err)
	}

	if err := exec.Command("helm", "repo", "update").Run(); err != nil {
		return fmt.Errorf("failed to update Helm repository: %v", err)
	}

	args = append(args, "--set", fmt.Sprintf("ingress.host=%s", c.IngressHost))

	// Add values files
	commonValuesFile := filepath.Join(c.ValuesPath, "common.yml")
	if _, err := os.Stat(commonValuesFile); err == nil {
		args = append(args, "--values", commonValuesFile)
	}

	stageValuesFile := filepath.Join(c.ValuesPath, fmt.Sprintf("%s.yml", c.Stage))
	if _, err := os.Stat(stageValuesFile); err == nil {
		args = append(args, "--values", stageValuesFile)
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

	cmd := exec.Command("helm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (c *Config) deployCustom() error {
	manifests, err := filepath.Glob(filepath.Join(c.ValuesPath, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to find manifests: %v", err)
	}

	for _, manifest := range manifests {
		cmd := exec.Command("kubectl", "apply", "-f", manifest, "-n", c.Namespace)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply manifest %s: %v", manifest, err)
		}
	}

	return nil
}

func main() {
	cfg := parseFlags()
	cfg.setupNames()
	cfg.PrintConfig()

	if err := cfg.Deploy(); err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deployment succeeded")
}
