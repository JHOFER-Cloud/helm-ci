// deploy/main.go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "text/template"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Stage        string
    AppName      string
    Environment  string
    PRNumber     string
    ValuesPath   string
    Chart        string
    Version      string
    Repository   string
    BaseDomain   string
    Namespace    string
    ReleaseName  string
    IngressHost  string
    DeployType   string // "helm" or "manifest"
    Image        string
    Port         int
}

func main() {
    cfg := parseFlags()
    
    cfg.setupNames()
    
    if err := cfg.ensureNamespace(); err != nil {
        fmt.Printf("Error creating namespace: %v\n", err)
        os.Exit(1)
    }

    switch cfg.DeployType {
    case "helm":
        if err := cfg.processValues(); err != nil {
            fmt.Printf("Error processing values: %v\n", err)
            os.Exit(1)
        }
        if err := cfg.deployHelm(); err != nil {
            fmt.Printf("Error deploying helm: %v\n", err)
            os.Exit(1)
        }
    case "manifest":
        if err := cfg.deployManifests(); err != nil {
            fmt.Printf("Error deploying manifests: %v\n", err)
            os.Exit(1)
        }
    default:
        fmt.Printf("Unknown deployment type: %s\n", cfg.DeployType)
        os.Exit(1)
    }
}

func parseFlags() *Config {
    cfg := &Config{}
    
    flag.StringVar(&cfg.Stage, "stage", "", "Deployment stage (dev/live)")
    flag.StringVar(&cfg.AppName, "app", "", "Application name")
    flag.StringVar(&cfg.Environment, "env", "", "Environment")
    flag.StringVar(&cfg.PRNumber, "pr", "", "PR number")
    flag.StringVar(&cfg.ValuesPath, "values", "helm/values", "Path to values files")
    flag.StringVar(&cfg.Chart, "chart", "", "Helm chart")
    flag.StringVar(&cfg.Version, "version", "", "Chart version")
    flag.StringVar(&cfg.Repository, "repo", "", "Helm repository")
    flag.StringVar(&cfg.BaseDomain, "domain", "", "Base domain")
    flag.StringVar(&cfg.DeployType, "type", "manifest", "Deployment type (helm/manifest)")
    flag.StringVar(&cfg.Image, "image", "", "Container image")
    flag.IntVar(&cfg.Port, "port", 8080, "Container port")
    
    flag.Parse()
    return cfg
}

func (c *Config) ensureNamespace() error {
    cmd := exec.Command("kubectl", "create", "namespace", c.Namespace, "--dry-run=client", "-o", "yaml")
    output, err := cmd.Output()
    if err != nil {
        return err
    }

    cmd = exec.Command("kubectl", "apply", "-f", "-")
    cmd.Stdin = strings.NewReader(string(output))
    return cmd.Run()
}

func (c *Config) deployManifests() error {
    // Create temporary directory for processed manifests
    tmpDir := "processed-manifests"
    if err := os.MkdirAll(tmpDir, 0755); err != nil {
        return err
    }

    // Generate deployment manifest
    deployment := fmt.Sprintf(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  namespace: %s
spec:
  replicas: %d
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
`, c.ReleaseName, c.Namespace, c.Environment == "development" ? 1 : 3, 
   c.ReleaseName, c.ReleaseName, c.AppName, c.Image, c.Port)

    // Generate service manifest
    service := fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: %s
  namespace: %s
spec:
  ports:
  - port: 80
    targetPort: %d
  selector:
    app: %s
`, c.ReleaseName, c.Namespace, c.Port, c.ReleaseName)

    // Generate Traefik IngressRoute manifest
    ingressRoute := fmt.Sprintf(`
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: %s
  namespace: %s
spec:
  entryPoints:
    - web
    - websecure
  routes:
    - match: Host(%q)
      kind: Rule
      services:
        - name: %s
          port: 80
`, c.ReleaseName, c.Namespace, c.IngressHost, c.ReleaseName)

    // Write manifests to files
    manifests := map[string]string{
        "deployment.yaml": deployment,
        "service.yaml":    service,
        "ingress.yaml":   ingressRoute,
    }

    for filename, content := range manifests {
        path := filepath.Join(tmpDir, filename)
        if err := os.WriteFile(path, []byte(content), 0644); err != nil {
            return err
        }
    }

    // Apply manifests
    cmd := exec.Command("kubectl", "apply", "-f", tmpDir)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return err
    }

    // Wait for deployment
    cmd = exec.Command("kubectl", "rollout", "status", "deployment", c.ReleaseName, 
        "-n", c.Namespace, "--timeout=300s")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func (c *Config) setupNames() {
    baseName := fmt.Sprintf("%s-%s", c.AppName, c.Stage)
    
    if c.Environment == "development" {
        c.Namespace = fmt.Sprintf("%s-%s", baseName, c.PRNumber)
        c.ReleaseName = fmt.Sprintf("%s-%s", baseName, c.PRNumber)
        c.IngressHost = fmt.Sprintf("%s-%s.%s.%s", c.AppName, c.PRNumber, c.Stage, c.BaseDomain)
    } else {
        c.Namespace = baseName
        c.ReleaseName = baseName
        c.IngressHost = fmt.Sprintf("%s.%s", c.AppName, c.BaseDomain)
    }
}

func (c *Config) processValues() error {
    tmpDir := "processed-values"
    if err := os.MkdirAll(tmpDir, 0755); err != nil {
        return err
    }

    // Process common values
    if err := c.processFile("common.yml", tmpDir); err != nil {
        return err
    }

    // Process stage-specific values
    if err := c.processFile(fmt.Sprintf("%s.yml", c.Stage), tmpDir); err != nil {
        return err
    }

    return nil
}

func (c *Config) processFile(filename, tmpDir string) error {
    sourcePath := filepath.Join(c.ValuesPath, filename)
    if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
        return nil // Skip if file doesn't exist
    }

    data := map[string]interface{}{
        "Namespace":   c.Namespace,
        "ReleaseName": c.ReleaseName,
        "IngressHost": c.IngressHost,
        "PRNumber":    c.PRNumber,
    }

    tmpl, err := template.ParseFiles(sourcePath)
    if err != nil {
        return err
    }

    outPath := filepath.Join(tmpDir, filename)
    out, err := os.Create(outPath)
    if err != nil {
        return err
    }
    defer out.Close()

    return tmpl.Execute(out, data)
}

func (c *Config) deployHelm() error {
    args := []string{
        "upgrade", "--install",
        c.ReleaseName,
        c.Chart,
        "--namespace", c.Namespace,
        "--create-namespace",
        "--wait",
        "--timeout", "10m",
    }

    if c.Version != "" {
        args = append(args, "--version", c.Version)
    }

    // Add values files
    commonValues := filepath.Join("processed-values", "common.yml")
    if _, err := os.Stat(commonValues); err == nil {
        args = append(args, "-f", commonValues)
    }

    stageValues := filepath.Join("processed-values", fmt.Sprintf("%s.yml", c.Stage))
    if _, err := os.Stat(stageValues); err == nil {
        args = append(args, "-f", stageValues)
    }

    cmd := exec.Command("helm", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    return cmd.Run()
}
