# Invulnerable ImageScans with ytt

This directory contains ytt templates for generating ImageScan resources for all Invulnerable components.

## Files

- **`imagescan.yaml`** - ytt template with loop to generate ImageScan resources
- **`values.yaml`** - Data values file containing all Invulnerable images and configuration

## Prerequisites

Install ytt:
```bash
# macOS
brew install ytt

# Linux
wget -O- https://github.com/vmware-tanzu/carvel-ytt/releases/latest/download/ytt-linux-amd64 > /tmp/ytt && \
  chmod +x /tmp/ytt && sudo mv /tmp/ytt /usr/local/bin/ytt

# Or via go
go install github.com/vmware-tanzu/carvel-ytt/cmd/ytt@latest
```

## Usage

### Generate ImageScans for all components

```bash
ytt -f imagescan.yaml -f values.yaml
```

This will generate 4 ImageScan resources:
- `invulnerable-backend-scan`
- `invulnerable-frontend-scan`
- `invulnerable-controller-scan`
- `invulnerable-scanner-scan`

### Apply to cluster

```bash
ytt -f imagescan.yaml -f values.yaml | kubectl apply -f -
```

### Override values

You can override values using the `--data-value` flag:

```bash
# Change namespace
ytt -f imagescan.yaml -f values.yaml \
  --data-value namespace=production

# Change webhook secret
ytt -f imagescan.yaml -f values.yaml \
  --data-value webhook.secretRef.name=teams-webhook

# Change schedule for all images
ytt -f imagescan.yaml -f values.yaml \
  --data-value images[0].schedule.cron="0 2 * * *"
```

### Custom values file

Create a custom values overlay:

```yaml
#@data/values
---
#@overlay/match-child-defaults missing_ok=True
images:
  - name: backend
    schedule:
      cron: "0 2 * * *"  # Daily at 2 AM instead of every 5 minutes
```

Apply with overlay:
```bash
ytt -f imagescan.yaml -f values.yaml -f my-values.yaml
```

## Configuration

Edit `values.yaml` to customize:

### Per-Image Configuration
- `name` - Component name (used for ImageScan name)
- `image` - Full image reference with tag
- `schedule.cron` - Cron schedule
- `schedule.suspend` - Temporarily pause scheduled scans
- `workspaceSize` - Disk space allocation

### Global Configuration
- `namespace` - Kubernetes namespace
- `timeZone` - Timezone for cron schedules
- `webhook` - Webhook notification settings
- `registryPolling` - Registry polling configuration
- `resources` - CPU/memory limits
- `scannerImage` - Scanner container image

## Examples

### Production configuration

```yaml
#@data/values
---
images:
  - name: backend
    image: ghcr.io/pacokleitz/invulnerable-backend:0.1.0
    schedule:
      cron: "0 2 * * *"  # Daily at 2 AM
      suspend: false
    workspaceSize: "10Gi"

webhook:
  secretRef:
    name: teams-webhook
  format: teams
  scanCompletion:
    enabled: true
    minSeverity: Critical  # Only critical vulnerabilities
  statusChange:
    enabled: true
    minSeverity: High

namespace: production
timeZone: "America/New_York"
```

### Development/testing configuration

```yaml
#@data/values
---
images:
  - name: backend
    image: ghcr.io/pacokleitz/invulnerable-backend:latest
    schedule:
      cron: "*/15 * * * *"  # Every 15 minutes
      suspend: false
    workspaceSize: "5Gi"

webhook:
  secretRef:
    name: webhook-test
  format: slack
  scanCompletion:
    enabled: true
    minSeverity: Low  # All vulnerabilities

namespace: dev
registryPolling:
  enabled: true
  interval: 2m  # Check every 2 minutes for fast feedback
```

## Webhook Setup

Before applying, create the webhook secret:

```bash
# For Slack
kubectl create secret generic slack-webhook \
  --from-literal=url='https://hooks.slack.com/services/YOUR/WEBHOOK/TOKEN' \
  -n invulnerable

# For Teams
kubectl create secret generic teams-webhook \
  --from-literal=url='https://outlook.office.com/webhook/YOUR/WEBHOOK/TOKEN' \
  -n invulnerable
```

## Verifying Output

Check the generated YAML before applying:

```bash
ytt -f imagescan.yaml -f values.yaml > generated.yaml
cat generated.yaml
```

## Cleanup

Delete all generated ImageScans:

```bash
ytt -f imagescan.yaml -f values.yaml | kubectl delete -f -
```

Or delete individually:
```bash
kubectl delete imagescan invulnerable-backend-scan -n invulnerable
kubectl delete imagescan invulnerable-frontend-scan -n invulnerable
kubectl delete imagescan invulnerable-controller-scan -n invulnerable
kubectl delete imagescan invulnerable-scanner-scan -n invulnerable
```
