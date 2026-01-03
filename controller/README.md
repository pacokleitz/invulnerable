# Invulnerable Scanner Controller

A Kubernetes controller that manages container image scanning using Custom Resource Definitions (CRDs).

## Overview

The Invulnerable Scanner Controller watches for `ImageScan` custom resources and automatically creates and manages CronJobs to scan container images for vulnerabilities. Each ImageScan resource represents a single container image with its own scanning schedule.

## Features

- **Declarative Image Scanning**: Define what images to scan using Kubernetes CRDs
- **Per-Image Schedules**: Each image can have its own cron schedule
- **Automatic CronJob Management**: Controller creates and updates CronJobs automatically
- **Resource Control**: Specify CPU/memory limits per image scan
- **SLA Compliance Tracking**: Configure remediation SLAs per severity with visual tracking
- **Webhook Notifications**: Configurable alerts to Slack/Teams with severity-based filtering
- **Private Registry Support**: Pull images from private registries using image pull secrets
- **Suspend Support**: Temporarily pause scanning for specific images
- **Status Reporting**: Track scan status and history via CRD status

## Architecture

```
┌─────────────────┐
│   ImageScan CRD │  (User creates)
└────────┬────────┘
         │
         │ watches
         ▼
┌────────────────────┐
│  Controller        │
│  (Reconciler)      │
└────────┬───────────┘
         │
         │ creates/updates
         ▼
┌────────────────────┐
│   CronJob          │  (One per ImageScan)
└────────┬───────────┘
         │
         │ runs on schedule
         ▼
┌────────────────────┐
│  Scanner Job       │  (Scans image and sends to API)
└────────────────────┘
```

## Quick Start

### 1. Install the Controller

```bash
# Using Helm
helm install invulnerable ./helm/invulnerable --namespace invulnerable --create-namespace

# The controller is enabled by default
```

### 2. Create an ImageScan Resource

```bash
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: nginx-scan
  namespace: invulnerable
spec:
  image: "nginx:latest"
  schedule: "0 2 * * *"  # Daily at 2 AM
  sbomFormat: "cyclonedx"
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"
EOF
```

### 3. View ImageScans

```bash
# List all ImageScans
kubectl get imagescans -n invulnerable

# Get detailed information
kubectl describe imagescan nginx-scan -n invulnerable

# View as YAML
kubectl get imagescan nginx-scan -n invulnerable -o yaml
```

## ImageScan CRD Reference

### Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | Yes | - | Container image to scan (e.g., "nginx:latest") |
| `schedule` | string | Yes | - | Cron schedule for scanning |
| `sbomFormat` | string | No | "cyclonedx" | SBOM format (cyclonedx or spdx) |
| `suspend` | boolean | No | false | Suspend scanning |
| `successfulJobsHistoryLimit` | int32 | No | 3 | Number of successful jobs to retain |
| `failedJobsHistoryLimit` | int32 | No | 3 | Number of failed jobs to retain |
| `resources` | ResourceRequirements | No | - | CPU/memory requests and limits |
| `workspaceSize` | string | No | "10Gi" | Temporary workspace size for image extraction |
| `apiEndpoint` | string | No | Auto-detected | Backend API endpoint |
| `scannerImage` | object | No | - | Scanner container image configuration |
| `webhooks` | object | No | - | Webhook notification configuration (scan completion & status changes) |
| `imagePullSecrets` | []LocalObjectReference | No | - | Secrets for pulling private images |
| `onlyFixed` | boolean | No | false | Only report vulnerabilities with available fixes |
| `sla` | object | No | See below | SLA remediation deadlines per severity (days) |

### Status Fields

| Field | Type | Description |
|-------|------|-------------|
| `cronJobName` | string | Name of the managed CronJob |
| `lastSuccessfulTime` | metav1.Time | Last successful scan completion |
| `conditions` | []metav1.Condition | Current status conditions |
| `observedGeneration` | int64 | Last observed generation |

### Example with All Options

```yaml
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: custom-app-scan
  namespace: invulnerable
spec:
  # Image to scan (required)
  image: "ghcr.io/myorg/myapp:v1.0.0"

  # Scan schedule (required) - cron format
  schedule: "0 2 * * *"  # Daily at 2 AM
  # schedule: "*/30 * * * *"  # Every 30 minutes
  # schedule: "0 4 * * 0"  # Weekly on Sunday at 4 AM

  # SBOM format (optional)
  sbomFormat: "cyclonedx"  # or "spdx"

  # Suspend scanning (optional)
  suspend: false

  # Job history limits (optional)
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3

  # Resource requirements (optional but recommended)
  resources:
    requests:
      memory: "512Mi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "2000m"

  # Workspace size for image extraction (optional)
  workspaceSize: "10Gi"

  # API endpoint (optional, auto-detected if not specified)
  apiEndpoint: "http://invulnerable-backend.invulnerable.svc.cluster.local:8080"

  # Scanner image (optional, uses chart defaults if not specified)
  scannerImage:
    repository: "invulnerable-scanner"
    tag: "latest"
    pullPolicy: "IfNotPresent"

  # Webhook notifications (optional)
  webhooks:
    # Scan completion notifications
    scanCompletion:
      enabled: true
      url: "https://hooks.slack.com/services/YOUR/SCAN/WEBHOOK"
      format: "slack"  # or "teams"
      minSeverity: "High"  # Critical, High, Medium, Low, Negligible
      onlyFixed: true      # Only notify for CVEs with fixes (default: true)

    # Status change notifications (when CVE status is updated via UI/API)
    statusChange:
      enabled: true
      url: "https://hooks.slack.com/services/YOUR/STATUS/WEBHOOK"  # Can be different!
      format: "slack"  # or "teams"
      minSeverity: "High"
      onlyFixed: true      # Only notify for CVEs with fixes (default: true)
      statusTransitions:  # Optional: filter by specific transitions
        - "active→fixed"
        - "active→ignored"
        - "in_progress→fixed"
      includeNoteChanges: false  # Don't notify for note-only updates

  # Image pull secrets for private registries (optional)
  imagePullSecrets:
    - name: my-registry-secret

  # Only report fixable vulnerabilities (optional)
  onlyFixed: false

  # SLA configuration for compliance tracking (optional)
  sla:
    critical: 7    # Critical vulnerabilities must be fixed within 7 days
    high: 30       # High severity within 30 days
    medium: 90     # Medium severity within 90 days
    low: 180       # Low severity within 180 days
```

## Common Use Cases

### Scan Multiple Images with Different Schedules

Create separate ImageScan resources for each image:

```bash
# Critical production app - scan every 6 hours
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: prod-api-scan
  namespace: invulnerable
spec:
  image: "myregistry.io/api:prod"
  schedule: "0 */6 * * *"
EOF

# Development images - scan daily
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: dev-api-scan
  namespace: invulnerable
spec:
  image: "myregistry.io/api:dev"
  schedule: "0 3 * * *"
EOF
```

### Temporarily Suspend Scanning

```bash
kubectl patch imagescan nginx-scan -n invulnerable \
  --type merge -p '{"spec":{"suspend":true}}'

# Resume scanning
kubectl patch imagescan nginx-scan -n invulnerable \
  --type merge -p '{"spec":{"suspend":false}}'
```

### Update Scan Schedule

```bash
kubectl patch imagescan nginx-scan -n invulnerable \
  --type merge -p '{"spec":{"schedule":"0 4 * * *"}}'
```

## Controller Configuration

The controller can be configured via Helm values:

```yaml
controller:
  enabled: true
  replicaCount: 1

  image:
    repository: invulnerable-controller
    tag: "latest"
    pullPolicy: IfNotPresent

  # Enable leader election for HA
  leaderElection:
    enabled: true

  # RBAC configuration (IMPORTANT for security)
  rbac:
    # clusterWide: false = Namespace-scoped (RECOMMENDED for least privilege)
    # clusterWide: true = Cluster-wide (watches all namespaces)
    clusterWide: false  # Default

  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
```

### RBAC and Security

The controller follows the **Principle of Least Privilege**:

- **Default (namespace-scoped)**: Controller only watches ImageScans in its deployment namespace
  - Uses `Role` and `RoleBinding` (namespace-scoped)
  - Most secure option
  - Recommended for production

- **Cluster-wide mode**: Controller watches ImageScans in all namespaces
  - Uses `ClusterRole` and `ClusterRoleBinding` (cluster-scoped)
  - Use only if you need multi-namespace scanning
  - Requires cluster-admin to install

**Important:** The controller can ONLY read ImageScans (get/list/watch). It cannot create or delete them - users manage ImageScans, the controller only reconciles them.

For detailed security documentation, see:
- [SECURITY.md](./SECURITY.md) - Comprehensive security guide
- [RBAC-CHANGES.md](./RBAC-CHANGES.md) - RBAC improvements and migration guide

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Kubernetes cluster (minikube, kind, etc.)

### Build

```bash
cd controller

# Build the binary
make build

# Build Docker image
make docker-build IMG=invulnerable-controller:dev

# Generate CRDs and code
make manifests generate
```

### Run Locally

```bash
# Install CRDs
kubectl apply -f config/crd/bases/

# Run controller locally (connects to current kubectl context)
make run
```

### Testing

```bash
# Create a test ImageScan
kubectl apply -f config/samples/invulnerable_v1alpha1_imagescan.yaml

# Check the controller logs
kubectl logs -f -l app.kubernetes.io/component=controller -n invulnerable

# Verify CronJob was created
kubectl get cronjobs -n invulnerable

# Check ImageScan status
kubectl get imagescan -n invulnerable
kubectl describe imagescan nginx-scan -n invulnerable
```

## Troubleshooting

### ImageScan Not Creating CronJob

Check the controller logs:
```bash
kubectl logs -f -l app.kubernetes.io/component=controller -n invulnerable
```

Check ImageScan status:
```bash
kubectl describe imagescan <name> -n invulnerable
```

### CronJob Not Running

Check if the ImageScan is suspended:
```bash
kubectl get imagescan <name> -n invulnerable -o yaml | grep suspend
```

Check CronJob status:
```bash
kubectl describe cronjob <cronjob-name> -n invulnerable
```

### Deleting an ImageScan

When you delete an ImageScan, the controller automatically deletes the associated CronJob:

```bash
kubectl delete imagescan nginx-scan -n invulnerable
```

## Migration from Old Scanner

If migrating from the previous ConfigMap-based scanner:

1. Install the new controller (included in Helm chart)
2. Create ImageScan resources for each image previously in the ConfigMap
3. The old scanner CronJob and ConfigMap are no longer needed

Example migration:

```bash
# Old approach (ConfigMap with multiple images)
# scanner.images:
#   - nginx:latest
#   - alpine:latest

# New approach (one ImageScan per image)
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: nginx-scan
  namespace: invulnerable
spec:
  image: "nginx:latest"
  schedule: "0 2 * * *"
---
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: alpine-scan
  namespace: invulnerable
spec:
  image: "alpine:latest"
  schedule: "0 2 * * *"
EOF
```

## License

See main project LICENSE file.
