# HELM CI/CD

## Usage

Example repo with a test apache helm and manifest deployment:
<https://github.com/JHOFER-Cloud/helm-test>

## Vault Integration

This tool supports HashiCorp Vault integration for secret management, allowing you to reference Vault secrets in your YAML files using placeholders.

### Placeholder Format

`<<vault.path/to/secret/KEY>>`

kv2: `vault_base_path` + placeholder = /v1/`vault_base_path`/path/to/secret

### Examples

```yaml
# deploy worklow
jobs:
  deploy-nginx:
    uses: JHOFER-Cloud/helm-ci/.github/workflows/k8s-deploy-template.yml@main
    with:
      #rest of config
      vault_url: https://vault.dev
      vault_base_path: vaultMountPath
      vault_kv_version: 2
```

#### Single Line Secret

```yaml
# values.yaml
database:
  password: <<vault.database/credentials/DB_PASSWORD>>
```

#### Multi-line Secret (e.g., JSON configuration)

```yaml
# values.yaml
renovate:
  config: <<vault.renovate/common/CONFIG>>
```

When the secret in Vault contains multi-line content (like JSON), it will be automatically formatted as a YAML block:

```yaml
# Processed output
renovate:
  config: |
    {
      "platform": "github",
      "token": "secret123",
      "autodiscover": true
    }
```

#### Vault Versions Support

Supports both KV v1 and KV v2 secret engines
Default version is KV v2

## Tests

```bash
go test ./deploy/vault/... -v
```

## EVERYTHING BELOW MAY BE OUTDATED

### Custom values, no chart

```yaml
# workflow.yml
# For custom values only
on:
  push:
    branches: [main]
  pull_request:
    branches: [main, dev*]

jobs:
  deploy:
    uses: JHOFER-Cloud/helm-ci/.github/workflows/k8s-deploy-template.yml@main
    with:
      app_name: my-app
      custom: true
      values_path: helm/values
      ingress_domain: company.com
    secrets:
      KUBE_CONFIG_DEV: ${{ secrets.KUBE_CONFIG_DEV }}
      KUBE_CONFIG_LIVE: ${{ secrets.KUBE_CONFIG_LIVE }}

# helm/values/common.yml
replicaCount: 1
image:
  repository: my-registry.com/my-app
  tag: latest

service:
  port: 3000

ingress:
  enabled: true
  host: { { .IngressHost } }
```

### Helm chart

```yaml
# .github/workflows/deploy.yml
jobs:
  deploy-nginx:
    uses: JHOFER-Cloud/helm-ci/.github/workflows/k8s-deploy-template.yml@main
    with:
      app_name: my-nginx
      helm_repository: https://charts.bitnami.com/bitnami
      helm_chart: nginx
      helm_version: 13.2.24
      values_path: helm/values
      ingress_domain: company.com
    secrets:
      KUBE_CONFIG_DEV: ${{ secrets.KUBE_CONFIG_DEV }}
      KUBE_CONFIG_LIVE: ${{ secrets.KUBE_CONFIG_LIVE }}
```
