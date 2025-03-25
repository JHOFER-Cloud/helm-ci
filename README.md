# HELM CI/CD (ARCHIVED)

I migrated to FluxCD

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
./run-tests.sh
```
