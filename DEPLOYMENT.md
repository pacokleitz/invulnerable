# Deployment Guide

This document describes how to deploy Invulnerable to Kubernetes using Helm.

## Prerequisites

- Kubernetes cluster (1.19+)
- Helm 3.0+
- kubectl configured to access your cluster
- External PostgreSQL database

## Quick Start

### 1. Build Docker Images

```bash
# Build backend image
docker build -t invulnerable-backend:latest -f backend/Dockerfile backend/

# Build frontend image
docker build -t invulnerable-frontend:latest -f frontend/Dockerfile frontend/

# Build scanner image
docker build -t invulnerable-scanner:latest -f scanner/Dockerfile scanner/

# If using a remote registry, tag and push
docker tag invulnerable-backend:latest myregistry.io/invulnerable-backend:latest
docker push myregistry.io/invulnerable-backend:latest
# ... repeat for frontend and scanner
```

### 2. Install PostgreSQL (if needed)

For testing/development, you can quickly install PostgreSQL:

```bash
# Add Bitnami repo
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Install PostgreSQL
helm install postgres bitnami/postgresql \
  --set auth.username=invulnerable \
  --set auth.password=changeme \
  --set auth.database=invulnerable \
  --create-namespace \
  --namespace invulnerable
```

For production, use a managed database service (AWS RDS, Google Cloud SQL, Azure Database).

### 3. Install Invulnerable

Basic installation with default values:

```bash
helm install invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --create-namespace
```

### 4. Access the Application

**With Ingress (default):**

Add to your `/etc/hosts`:
```
127.0.0.1 invulnerable.local
```

If using Minikube:
```bash
minikube tunnel
```

Access at: http://invulnerable.local

**Without Ingress (port-forward):**

```bash
# Frontend
kubectl port-forward -n invulnerable svc/invulnerable-frontend 8080:80

# Backend
kubectl port-forward -n invulnerable svc/invulnerable-backend 8081:8080
```

Access at: http://localhost:8080

## Production Installation

### 1. Create a values file

Create `production-values.yaml`:

```yaml
# production-values.yaml

# Global registry configuration
# All components will pull from this registry
image:
  registry: "myregistry.io"  # Your private registry or ghcr.io, docker.io, etc.

# Use production images with specific tags
frontend:
  image:
    repository: invulnerable-frontend
    tag: "1.0.0"
    pullPolicy: Always
  replicaCount: 2
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"

backend:
  image:
    repository: invulnerable-backend
    tag: "1.0.0"
    pullPolicy: Always
  replicaCount: 2

  # Production database configuration
  database:
    host: "postgres.production.example.com"
    port: 5432
    user: "invulnerable"
    name: "invulnerable"
    sslmode: "require"
    # Use Kubernetes secret for password
    existingSecret: "postgres-credentials"
    passwordKey: "password"

  # Enable autoscaling
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80

  resources:
    requests:
      memory: "256Mi"
      cpu: "200m"
    limits:
      memory: "1Gi"
      cpu: "1000m"

scanner:
  image:
    repository: invulnerable-scanner
    tag: "1.0.0"
    pullPolicy: Always

  # Configure scanner for your images
  schedule: "0 2 * * *"  # 2 AM daily
  images:
    - myregistry.io/app:latest
    - myregistry.io/api:latest
    - myregistry.io/worker:latest
    - nginx:latest
    - postgres:15

# Production ingress with TLS
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
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

# Use image pull secrets for private registry
imagePullSecrets:
  - name: registry-credentials
```

### 2. Create secrets

Create database credentials secret:

```bash
kubectl create secret generic postgres-credentials \
  --from-literal=password='super-secret-password' \
  --namespace invulnerable
```

Create image pull secret (if using private registry):

```bash
kubectl create secret docker-registry registry-credentials \
  --docker-server=myregistry.io \
  --docker-username=myuser \
  --docker-password=mypassword \
  --docker-email=myemail@example.com \
  --namespace invulnerable
```

### 3. Install with production values

```bash
helm install invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --create-namespace \
  --values production-values.yaml
```

## Upgrading

```bash
# Upgrade to new version
helm upgrade invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --values production-values.yaml

# Upgrade with different values
helm upgrade invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --set backend.image.tag=1.1.0 \
  --set frontend.image.tag=1.1.0
```

## Uninstalling

```bash
helm uninstall invulnerable --namespace invulnerable
```

## Configuration Options

See [helm/invulnerable/README.md](helm/invulnerable/README.md) for all configuration options.

### Common Configurations

**Disable components:**

```yaml
# Disable scanner cronjobs
scanner:
  enabled: false

# Disable frontend (API only)
frontend:
  enabled: false
```

**Custom scanner schedule:**

```yaml
scanner:
  schedule: "0 */6 * * *"   # Every 6 hours
  images:
    - nginx:latest
    - postgres:15
```

**Resource limits:**

```yaml
backend:
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
```

## Monitoring and Troubleshooting

### Check deployment status

```bash
# Check all resources
kubectl get all -n invulnerable

# Check pod status
kubectl get pods -n invulnerable

# Describe pod
kubectl describe pod <pod-name> -n invulnerable
```

### View logs

```bash
# Backend logs
kubectl logs -f -l app.kubernetes.io/component=backend -n invulnerable

# Frontend logs
kubectl logs -f -l app.kubernetes.io/component=frontend -n invulnerable

# Scanner logs
kubectl logs -l app.kubernetes.io/component=scanner -n invulnerable

# Get logs from specific job
kubectl logs job/invulnerable-scanner-<job-id> -n invulnerable
```

### Check connectivity

```bash
# Test backend health
kubectl run curl --image=curlimages/curl -i --rm --restart=Never \
  -- curl http://invulnerable-backend.invulnerable.svc.cluster.local:8080/health

# Test from backend to database
kubectl exec -it deployment/invulnerable-backend -n invulnerable -- \
  /bin/sh -c 'nc -zv $DB_HOST $DB_PORT'
```

### Common Issues

**Backend can't connect to database:**
- Check database credentials in secret
- Verify database host is accessible from cluster
- Check network policies

**Images not pulling:**
- Verify image pull secrets are configured
- Check image repository and tag are correct
- Ensure registry is accessible from cluster

**Scanner jobs failing:**
- Check scanner logs for errors
- Verify API endpoint is accessible
- Check resource limits (scanning needs memory/CPU)

## Backup and Recovery

### Database Backup

```bash
# Backup using pg_dump
kubectl run pg-backup --image=postgres:15 -i --rm --restart=Never -- \
  pg_dump -h <db-host> -U invulnerable -d invulnerable > backup.sql
```

### Restore

```bash
# Restore from backup
kubectl run pg-restore --image=postgres:15 -i --rm --restart=Never -- \
  psql -h <db-host> -U invulnerable -d invulnerable < backup.sql
```

## Support

For issues and questions:
- GitHub Issues: https://github.com/pacokleitz/invulnerable/issues
- Documentation: See README.md and helm/invulnerable/README.md
