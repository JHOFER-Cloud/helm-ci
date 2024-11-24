# Usage

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik -n traefik --create-namespace
```

Put common.yml, dev.yml and live.yml in helm/values

```yaml
# .github/workflows/deploy.yml
name: Deploy Application

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  # Example with Helm
  deploy-with-helm:
    uses: JHOFER-Cloud/helm-ci/.github/workflows/k8s-deploy-template.yml@main
    with:
      app_name: my-helm-app
      deployment_type: helm
      helm_repository: https://charts.bitnami.com/bitnami
      helm_chart: nginx
      helm_version: 13.2.24
      values_path: helm/values
      ingress_domain: company.com
    secrets:
      KUBE_CONFIG_DEV: ${{ secrets.KUBE_CONFIG_DEV }}
      KUBE_CONFIG_LIVE: ${{ secrets.KUBE_CONFIG_LIVE }}

  # Example with raw manifests
  deploy-raw-app:
    uses: JHOFER-Cloud/helm-ci/.github/workflows/k8s-deploy-template.yml@main
    with:
      app_name: my-raw-app
      deployment_type: manifest
      image: my-registry.com/my-app:${{ github.sha }}
      port: 3000
      ingress_domain: company.com
    secrets:
      KUBE_CONFIG_DEV: ${{ secrets.KUBE_CONFIG_DEV }}
      KUBE_CONFIG_LIVE: ${{ secrets.KUBE_CONFIG_LIVE }}
```
