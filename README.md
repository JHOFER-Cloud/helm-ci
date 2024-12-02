# HELM CI/CD

## Usage

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik -n traefik --create-namespace
```

### Custom values, no chart

```yaml
# workflow.yml
# For custom values only
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