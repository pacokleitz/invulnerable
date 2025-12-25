# Invulnerable Helm Chart

This Helm chart deploys the Invulnerable container vulnerability scanner and management platform on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- External PostgreSQL database (not included in this chart)

## Installation

### Quick Start

```bash
# Install with default values
helm install invulnerable ./helm/invulnerable

# Install in a specific namespace
helm install invulnerable ./helm/invulnerable -n invulnerable --create-namespace

# Install with custom values
helm install invulnerable ./helm/invulnerable -f custom-values.yaml
```

### Configuration

The following table lists the configurable parameters and their default values.

#### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Override full name | `""` |
| `image.registry` | Global image registry (can be overridden per component) | `""` |

#### Frontend

| Parameter | Description | Default |
|-----------|-------------|---------|
| `frontend.enabled` | Enable frontend deployment | `true` |
| `frontend.replicaCount` | Number of frontend replicas | `2` |
| `frontend.image.registry` | Frontend image registry (overrides global) | `""` |
| `frontend.image.repository` | Frontend image repository | `invulnerable-frontend` |
| `frontend.image.tag` | Frontend image tag | `latest` |
| `frontend.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `frontend.service.type` | Service type | `ClusterIP` |
| `frontend.service.port` | Service port | `80` |

#### Backend

| Parameter | Description | Default |
|-----------|-------------|---------|
| `backend.enabled` | Enable backend deployment | `true` |
| `backend.replicaCount` | Number of backend replicas | `2` |
| `backend.image.registry` | Backend image registry (overrides global) | `""` |
| `backend.image.repository` | Backend image repository | `invulnerable-backend` |
| `backend.image.tag` | Backend image tag | `latest` |
| `backend.database.host` | PostgreSQL host | `postgres.default.svc.cluster.local` |
| `backend.database.port` | PostgreSQL port | `5432` |
| `backend.database.user` | Database user | `invulnerable` |
| `backend.database.password` | Database password | `changeme` |
| `backend.database.name` | Database name | `invulnerable` |
| `backend.autoscaling.enabled` | Enable HPA | `true` |
| `backend.autoscaling.minReplicas` | Minimum replicas | `2` |
| `backend.autoscaling.maxReplicas` | Maximum replicas | `10` |

#### Scanner

| Parameter | Description | Default |
|-----------|-------------|---------|
| `scanner.enabled` | Enable scanner cronjob | `true` |
| `scanner.image.registry` | Scanner image registry (overrides global) | `""` |
| `scanner.image.repository` | Scanner image repository | `invulnerable-scanner` |
| `scanner.schedule` | Cron schedule for scanning | `"0 2 * * *"` |
| `scanner.images` | List of images to scan (supports single or multiple) | `["nginx:latest", "alpine:latest", "ubuntu:22.04"]` |
| `scanner.successfulJobsHistoryLimit` | Number of successful jobs to keep | `3` |
| `scanner.failedJobsHistoryLimit` | Number of failed jobs to keep | `3` |

#### Ingress

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts[0].host` | Hostname | `invulnerable.local` |

### Example Custom Values

Create a `custom-values.yaml` file:

```yaml
# Global registry configuration
# All components will pull images from this registry
image:
  registry: "ghcr.io/myorg"  # GitHub Container Registry
  # Or use: docker.io, gcr.io/my-project, myregistry.azurecr.io, etc.

# Use production database
backend:
  database:
    host: postgres.production.svc.cluster.local
    password: super-secret-password
    # Or use existing secret
    existingSecret: postgres-credentials
    passwordKey: password

# Configure ingress for your domain
ingress:
  hosts:
    - host: invulnerable.example.com
      paths:
        - path: /api
          pathType: Prefix
          backend: backend
        - path: /
          pathType: Prefix
          backend: frontend
  tls:
    - secretName: invulnerable-tls
      hosts:
        - invulnerable.example.com

# Configure scanner to scan your images
scanner:
  schedule: "0 2 * * *"  # Daily at 2 AM
  images:
    - myregistry.io/app:latest
    - myregistry.io/api:latest
    - myregistry.io/worker:latest
```

Install with custom values:

```bash
helm install invulnerable ./helm/invulnerable -f custom-values.yaml
```

### Registry Configuration Examples

**Using a global registry:**

```yaml
# All components will use ghcr.io/myorg
image:
  registry: "ghcr.io/myorg"

# Results in:
# - Frontend: ghcr.io/myorg/invulnerable-frontend:latest
# - Backend: ghcr.io/myorg/invulnerable-backend:latest
# - Scanner: ghcr.io/myorg/invulnerable-scanner:latest
```

**Per-component registry override:**

```yaml
# Global registry
image:
  registry: "ghcr.io/myorg"

# Override only for scanner (e.g., using AWS ECR for scanner)
scanner:
  image:
    registry: "123456789012.dkr.ecr.us-east-1.amazonaws.com"
    repository: "invulnerable-scanner"

# Results in:
# - Frontend: ghcr.io/myorg/invulnerable-frontend:latest
# - Backend: ghcr.io/myorg/invulnerable-backend:latest
# - Scanner: 123456789012.dkr.ecr.us-east-1.amazonaws.com/invulnerable-scanner:latest
```

**Using different registries:**

```bash
# GitHub Container Registry
--set image.registry=ghcr.io/myorg

# Docker Hub (explicit)
--set image.registry=docker.io/myorg

# Google Container Registry
--set image.registry=gcr.io/my-project

# Azure Container Registry
--set image.registry=myregistry.azurecr.io

# AWS Elastic Container Registry
--set image.registry=123456789012.dkr.ecr.us-east-1.amazonaws.com

# Private registry
--set image.registry=registry.mycompany.com
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade invulnerable ./helm/invulnerable -f custom-values.yaml

# Upgrade to a new chart version
helm upgrade invulnerable ./helm/invulnerable --version 0.2.0
```

## Uninstalling

```bash
helm uninstall invulnerable
```

## Database Setup

This chart does not include a PostgreSQL deployment. You must provide an external PostgreSQL database.

### Quick PostgreSQL Setup (for testing)

```bash
# Install PostgreSQL using Bitnami chart
helm install postgres bitnami/postgresql \
  --set auth.username=invulnerable \
  --set auth.password=changeme \
  --set auth.database=invulnerable

# Get the connection details
export POSTGRES_PASSWORD=$(kubectl get secret --namespace default postgres-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
echo "PostgreSQL password: $POSTGRES_PASSWORD"
```

### Production Database

For production, use a managed PostgreSQL service (AWS RDS, Google Cloud SQL, Azure Database) or deploy PostgreSQL with proper backups and high availability.

## Security Considerations

- **Never commit secrets to version control**
- Use Kubernetes secrets or external secret management (Vault, AWS Secrets Manager)
- Enable TLS for ingress in production
- Use `imagePullSecrets` for private registries
- Review and adjust resource limits based on your workload
- Enable network policies to restrict traffic
- **All containers run as non-root by default** (UID 1000)
  - Pod security contexts enforce `runAsNonRoot: true`
  - Container security contexts drop all capabilities
  - `allowPrivilegeEscalation` is disabled
  - Frontend runs on port 8080 (non-privileged port)

## Troubleshooting

### Check pod status

```bash
kubectl get pods -l app.kubernetes.io/instance=invulnerable
```

### View logs

```bash
# Backend logs
kubectl logs -l app.kubernetes.io/component=backend -f

# Frontend logs
kubectl logs -l app.kubernetes.io/component=frontend -f

# Scanner logs
kubectl logs -l app.kubernetes.io/component=scanner -f
```

### Test connectivity

```bash
# Port forward to backend
kubectl port-forward svc/invulnerable-backend 8080:8080

# Test health endpoint
curl http://localhost:8080/health
```

## Contributing

See the main project README for contribution guidelines.

## License

See the main project LICENSE file.
