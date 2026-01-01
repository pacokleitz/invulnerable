# Invulnerable Helm Chart

This Helm chart deploys the Invulnerable container vulnerability scanner and management platform on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- External PostgreSQL database (not included in this chart)
- External S3 compatible bucket (not included in this chart)

## Installation

### Quick Start

**⚠️ Important:** Invulnerable requires an Ingress controller to route traffic between frontend and backend services.

```bash
# Install with default values (ingress enabled at invulnerable.local)
helm install invulnerable ./helm/invulnerable -n invulnerable --create-namespace

# Add to /etc/hosts (or C:\Windows\System32\drivers\etc\hosts on Windows)
echo "127.0.0.1 invulnerable.local" | sudo tee -a /etc/hosts

# Access at http://invulnerable.local
```

**For production with public access and authentication, see "Security Considerations" section below.**

## Security Considerations

**⚠️ IMPORTANT: Security & Access Configuration**

The application requires an Ingress controller to function properly (routes frontend and backend traffic).

**Default Configuration (Development/Testing):**
- ✅ Ingress is **ENABLED** by default at `invulnerable.local`
- ⚠️ OAuth2 Proxy is **DISABLED** by default (no authentication)
- ⚠️ **Not suitable for production or public networks!**

**⚠️ Note on CVE Status Tracking Without Authentication:**

The vulnerability status tracking feature (marking CVEs as fixed, ignored, etc.) works without OAuth2 authentication, but **all changes will be recorded as "unknown" user** in the audit history. For production deployments requiring:
- **Audit trails** - Know who made each decision
- **User accountability** - Track status changes to specific users
- **Compliance requirements** - Non-repudiation of security decisions

You **MUST** enable OAuth2 Proxy to properly capture user identity in the change history.

**For Production with Public Access:**

You **MUST** enable oauth2Proxy for authentication:

```yaml
# ✅ SECURE: Both ingress and authentication enabled
oauth2Proxy:
  enabled: true
  # ... configure OAuth provider

ingress:
  enabled: true
  # ... configure domain and TLS
```

**❌ NEVER do this in production:**
```yaml
# ❌ INSECURE: Public access without authentication
ingress:
  enabled: true
oauth2Proxy:
  enabled: false  # DON'T DO THIS!
```

## Architecture & Routing

The frontend application makes API calls to `/api/v1/*` as relative paths, meaning it expects the backend to be accessible at the same hostname/IP. The application **requires** proper routing to function:

```
Browser Request Flow:
  http://your-domain/          → Frontend Service (serves React SPA)
  http://your-domain/api/v1/*  → Backend Service (API endpoints)
```

**Routing is handled by:**
- **Kubernetes Ingress** - Routes `/api` prefix to backend service, everything else to frontend
- **OAuth2 Proxy** (when enabled) - Sits in front and handles authentication before routing

**Important:** The frontend Docker image uses Nginx but **does NOT** include proxy configuration to route `/api` requests to the backend. This routing must be provided by the Kubernetes Ingress resource or an external reverse proxy.

### Configuration

The following table lists the configurable parameters and their default values.

#### Global Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Override full name | `""` |
| `image.registry` | Global image registry (can be overridden per component) | `ghcr.io/pacokleitz` |

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
| `frontend.service.targetPort` | Service target port | `8080` |
| `frontend.resources.requests.memory` | Memory request | `64Mi` |
| `frontend.resources.requests.cpu` | CPU request | `50m` |
| `frontend.resources.limits.memory` | Memory limit | `256Mi` |
| `frontend.resources.limits.cpu` | CPU limit | `200m` |

#### Backend

| Parameter | Description | Default |
|-----------|-------------|---------|
| `backend.enabled` | Enable backend deployment | `true` |
| `backend.replicaCount` | Number of backend replicas | `2` |
| `backend.image.registry` | Backend image registry (overrides global) | `""` |
| `backend.image.repository` | Backend image repository | `invulnerable-backend` |
| `backend.image.tag` | Backend image tag | `latest` |
| `backend.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `backend.service.type` | Service type | `ClusterIP` |
| `backend.service.port` | Service port | `8080` |
| `backend.service.targetPort` | Service target port | `8080` |
| `backend.frontendURL` | Frontend URL for webhook notifications | `""` |
| `backend.database.host` | PostgreSQL host | `postgres.default.svc.cluster.local` |
| `backend.database.port` | PostgreSQL port | `5432` |
| `backend.database.user` | Database user | `invulnerable` |
| `backend.database.password` | Database password | `changeme` |
| `backend.database.name` | Database name | `invulnerable` |
| `backend.database.sslmode` | PostgreSQL SSL mode | `disable` |
| `backend.database.existingSecret` | Use existing secret for database password | `""` |
| `backend.database.passwordKey` | Key in existing secret for password | `password` |
| `backend.s3.endpoint` | S3 endpoint URL | `""` |
| `backend.s3.bucket` | S3 bucket name | `invulnerable` |
| `backend.s3.region` | S3 region | `us-east-1` |
| `backend.s3.accessKey` | S3 access key | `""` |
| `backend.s3.secretKey` | S3 secret key | `""` |
| `backend.s3.useSSL` | Use SSL for S3 connection | `true` |
| `backend.s3.existingSecret` | Use existing secret for S3 credentials | `""` |
| `backend.s3.accessKeyKey` | Key in existing secret for access key | `access-key` |
| `backend.s3.secretKeyKey` | Key in existing secret for secret key | `secret-key` |
| `backend.autoscaling.enabled` | Enable HPA | `true` |
| `backend.autoscaling.minReplicas` | Minimum replicas | `2` |
| `backend.autoscaling.maxReplicas` | Maximum replicas | `10` |
| `backend.autoscaling.targetCPUUtilizationPercentage` | Target CPU % | `70` |
| `backend.autoscaling.targetMemoryUtilizationPercentage` | Target memory % | `80` |
| `backend.resources.requests.memory` | Memory request | `128Mi` |
| `backend.resources.requests.cpu` | CPU request | `100m` |
| `backend.resources.limits.memory` | Memory limit | `512Mi` |
| `backend.resources.limits.cpu` | CPU limit | `500m` |

#### Controller

| Parameter | Description | Default |
|-----------|-------------|---------|
| `controller.enabled` | Enable the ImageScan controller | `true` |
| `controller.replicaCount` | Number of controller replicas | `1` |
| `controller.image.registry` | Controller image registry (overrides global) | `""` |
| `controller.image.repository` | Controller image repository | `invulnerable-controller` |
| `controller.image.tag` | Controller image tag | `latest` |
| `controller.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `controller.leaderElection.enabled` | Enable leader election for HA | `true` |
| `controller.rbac.clusterWide` | Controller watches all namespaces (false = own namespace only) | `false` |
| `controller.resources.requests.memory` | Memory request | `128Mi` |
| `controller.resources.requests.cpu` | CPU request | `100m` |
| `controller.resources.limits.memory` | Memory limit | `512Mi` |
| `controller.resources.limits.cpu` | CPU limit | `500m` |

#### Scanner

| Parameter | Description | Default |
|-----------|-------------|---------|
| `scanner.image.registry` | Scanner image registry (overrides global) | `""` |
| `scanner.image.repository` | Scanner image repository (used by ImageScan CRDs) | `invulnerable-scanner` |
| `scanner.image.tag` | Scanner image tag | `latest` |
| `scanner.image.pullPolicy` | Image pull policy | `IfNotPresent` |

#### OAuth2 Proxy (Authentication)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `oauth2Proxy.enabled` | Enable OAuth2 authentication | `false` |
| `oauth2Proxy.replicaCount` | Number of OAuth2 proxy replicas | `2` |
| `oauth2Proxy.image.tag` | OAuth2 proxy image tag | `v7.13.0` |
| `oauth2Proxy.image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `oauth2Proxy.clientID` | OAuth client ID | `""` |
| `oauth2Proxy.clientSecret` | OAuth client secret | `""` |
| `oauth2Proxy.cookieSecret` | Cookie encryption secret (32 chars) | `""` |
| `oauth2Proxy.existingSecret` | Use existing secret for credentials | `""` |
| `oauth2Proxy.config.provider` | OAuth provider (google, github, oidc, etc.) | `"oidc"` |
| `oauth2Proxy.config.oidcIssuerUrl` | OIDC issuer URL | `""` |
| `oauth2Proxy.config.loginUrl` | OAuth login URL (for non-OIDC providers) | `""` |
| `oauth2Proxy.config.redeemUrl` | OAuth redeem URL (for non-OIDC providers) | `""` |
| `oauth2Proxy.config.validateUrl` | OAuth validate URL (for non-OIDC providers) | `""` |
| `oauth2Proxy.config.redirectUrl` | OAuth redirect URL | `"https://invulnerable.local/oauth2/callback"` |
| `oauth2Proxy.config.emailDomains` | Restrict to email domains | `[]` |
| `oauth2Proxy.config.cookieName` | OAuth2 cookie name | `_oauth2_proxy` |
| `oauth2Proxy.config.cookieSecure` | Require HTTPS for cookies | `true` |
| `oauth2Proxy.config.cookieDomains` | Cookie domain restrictions | `[]` |
| `oauth2Proxy.config.skipProviderButton` | Skip provider selection page | `false` |
| `oauth2Proxy.config.whitelist` | Whitelist domains for redirects | `[]` |
| `oauth2Proxy.config.configFile` | Inline OAuth2 config (overrides other settings) | `""` |
| `oauth2Proxy.config.extraEnv` | Extra environment variables | `[]` |
| `oauth2Proxy.config.extraArgs` | Extra command line arguments | `[]` |
| `oauth2Proxy.resources.requests.memory` | Memory request | `64Mi` |
| `oauth2Proxy.resources.requests.cpu` | CPU request | `50m` |
| `oauth2Proxy.resources.limits.memory` | Memory limit | `256Mi` |
| `oauth2Proxy.resources.limits.cpu` | CPU limit | `200m` |

#### Ingress

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress (⚠️ REQUIRED for app to function) | `true` |
| `ingress.className` | Ingress class name | `nginx` |
| `ingress.hosts[0].host` | Hostname | `invulnerable.local` |
| `ingress.tls` | TLS configuration | `[]` |

#### Security & Service Account

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name (auto-generated if empty) | `""` |
| `imagePullSecrets` | Image pull secrets for private registries | `[]` |
| `podSecurityContext.runAsNonRoot` | Run pods as non-root | `true` |
| `podSecurityContext.runAsUser` | User ID to run pods | `1000` |
| `podSecurityContext.runAsGroup` | Group ID to run pods | `1000` |
| `podSecurityContext.fsGroup` | File system group ID | `1000` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.capabilities.drop` | Dropped capabilities | `[ALL]` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `false` |
| `securityContext.runAsNonRoot` | Run as non-root | `true` |
| `securityContext.runAsUser` | User ID | `1000` |

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

# Enable and configure ingress (REQUIRED for the application to function)
ingress:
  enabled: true
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

# Controller is enabled by default
# After installation, create ImageScan CRDs to define images to scan
```

**After installing**, create ImageScan resources to scan your images:

```bash
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: app-scan
  namespace: invulnerable
spec:
  image: "myregistry.io/app:latest"
  schedule: "0 2 * * *"  # Daily at 2 AM
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
---
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: api-scan
  namespace: invulnerable
spec:
  image: "myregistry.io/api:latest"
  schedule: "0 */6 * * *"  # Every 6 hours
---
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: worker-scan
  namespace: invulnerable
spec:
  image: "myregistry.io/worker:latest"
  schedule: "0 3 * * *"  # Daily at 3 AM
EOF
```

Install with custom values:

```bash
helm install invulnerable ./helm/invulnerable -f custom-values.yaml
```

### Authentication with OAuth2

Invulnerable supports OAuth2 authentication for protecting access to the application. OAuth2 provides:
- **Access control** - Restrict who can access the platform
- **User identity tracking** - Capture user email/username in vulnerability status change audit logs
- **Compliance** - Enable proper audit trails for security decisions

Without OAuth2 enabled, all CVE status changes are recorded as "unknown" user in the history.

See [examples/](examples/) for provider-specific configurations.

**Enable with Google:**

```bash
# Generate cookie secret
COOKIE_SECRET=$(openssl rand -base64 32 | head -c 32)

# Install with Google OAuth and ingress
helm install invulnerable ./helm/invulnerable \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host="invulnerable.example.com" \
  --set oauth2Proxy.enabled=true \
  --set oauth2Proxy.clientID="your-id.apps.googleusercontent.com" \
  --set oauth2Proxy.clientSecret="your-secret" \
  --set oauth2Proxy.cookieSecret="$COOKIE_SECRET" \
  --set oauth2Proxy.config.provider="google" \
  --set oauth2Proxy.config.redirectUrl="https://invulnerable.example.com/oauth2/callback" \
  --set oauth2Proxy.config.emailDomains[0]="example.com"
```

**Or use a values file:**

```yaml
# values-with-auth.yaml
ingress:
  enabled: true
  hosts:
    - host: invulnerable.example.com

oauth2Proxy:
  enabled: true
  clientID: "your-client-id"
  clientSecret: "your-client-secret"
  cookieSecret: "your-32-char-secret"
  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-provider.com"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"
    emailDomains:
      - "example.com"
```

For more examples, see the [examples/](examples/) directory which includes configurations for Google, GitHub, and Keycloak.

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
helm upgrade invulnerable ./helm/invulnerable --version 0.1.0
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
