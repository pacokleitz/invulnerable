# Tilt Local Development Guide

This guide explains how to use [Tilt](https://tilt.dev/) for local development with Invulnerable on Docker Desktop's built-in Kubernetes.

## Prerequisites

1. **Docker Desktop** with Kubernetes enabled
   ```bash
   # Verify Docker is running
   docker ps

   # Enable Kubernetes:
   # 1. Open Docker Desktop
   # 2. Go to Settings > Kubernetes
   # 3. Check "Enable Kubernetes"
   # 4. Click "Apply & Restart"

   # Verify Kubernetes is running
   kubectl cluster-info --context docker-desktop
   ```

2. **kubectl** - Kubernetes CLI
   ```bash
   # macOS
   brew install kubectl

   # Linux
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x kubectl
   sudo mv kubectl /usr/local/bin/

   # Verify installation
   kubectl version --client
   ```

4. **Helm** - Kubernetes package manager
   ```bash
   # macOS
   brew install helm

   # Linux
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

   # Verify installation
   helm version
   ```

5. **Tilt** - Local development tool
   ```bash
   # macOS
   brew install tilt

   # Linux
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

   # Verify installation
   tilt version
   ```

## Quick Start

### 1. Enable Docker Desktop Kubernetes

```bash
# Switch to docker-desktop context
kubectl config use-context docker-desktop

# Verify cluster is running
kubectl cluster-info
```

### 2. Run Setup Script

The setup script will configure your environment:

```bash
./scripts/setup-tilt.sh
```

This will:
- Verify Docker Desktop Kubernetes is running
- Switch to docker-desktop context
- Configure /etc/hosts for invulnerable.local

### 3. Start Tilt

```bash
# HTTP mode (default)
tilt up

# OR with HTTPS enabled
tilt up -- --enable-https
```

Tilt will automatically deploy:
- nginx Ingress Controller
- cert-manager (if --enable-https)
- PostgreSQL
- Invulnerable application (backend, frontend, controller)
- OAuth2 Proxy

### Manual Setup (Optional)

If you prefer to set up components manually instead of using the setup script:

#### Install nginx Ingress Controller

```bash
# Install nginx-ingress using Helm
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

kubectl create namespace ingress-nginx

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --set controller.hostPort.enabled=true \
  --set controller.service.type=NodePort \
  --set controller.watchIngressWithoutClass=true \
  --wait

# Verify ingress is ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

#### (Optional) Install cert-manager for HTTPS

For local development with HTTPS/TLS:

```bash
# Add jetstack Helm repository
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Create namespace
kubectl create namespace cert-manager

# Install cert-manager with CRDs
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --set installCRDs=true \
  --wait

# Verify cert-manager is ready
kubectl wait --namespace cert-manager \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/instance=cert-manager \
  --timeout=120s

# Create self-signed ClusterIssuer for local dev
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF

# Create namespace if not exists
kubectl create namespace invulnerable || true

# Create certificate for invulnerable.local
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: invulnerable-local-cert
  namespace: invulnerable
spec:
  secretName: invulnerable-tls
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
    - invulnerable.local
    - "*.invulnerable.local"
EOF

# Wait for certificate to be ready
kubectl wait --namespace invulnerable \
  --for=condition=ready certificate/invulnerable-local-cert \
  --timeout=60s
```

**Note**: Tilt can handle this automatically with the `--enable-https` flag. This manual setup is only needed if you're not using Tilt's automatic deployment.

## Using Tilt

### Start Development Environment

```bash
# Add Bitnami Helm repository
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Create namespace
kubectl create namespace invulnerable

# Install PostgreSQL (optional - Tilt will do this)
# This step is automatic when you run Tilt
```

### 5. Configure /etc/hosts

Add local DNS entry for the ingress:

```bash
# Add to /etc/hosts
echo "127.0.0.1 invulnerable.local" | sudo tee -a /etc/hosts
```

### 6. Configure OAuth2 (Optional but Recommended)

For a better local dev experience, set up Google OAuth:

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create OAuth 2.0 Client ID (Web application)
3. Add authorized redirect URI: `http://invulnerable.local/oauth2/callback`
4. Update `tilt/values.yaml`:

```yaml
oauth2Proxy:
  clientID: "your-google-client-id.apps.googleusercontent.com"
  clientSecret: "your-google-client-secret"
  config:
    provider: "google"
    emailDomains:
      - "yourdomain.com"  # or your personal email domain
```

**Alternative**: The default config allows any email (insecure but convenient for local dev).

### 7. Start Tilt

```bash
# Start Tilt (opens web UI at http://localhost:10350)
tilt up

# Or run in CI mode (no web UI)
tilt ci
```

Tilt will:
- Build all Docker images (backend, frontend, controller, scanner)
- Deploy PostgreSQL
- Deploy Invulnerable with Helm (including ingress + oauth2-proxy)
- Set up port forwards
- Watch for code changes and auto-rebuild

### 8. Access the Application

**Via Ingress (recommended):**
```
http://invulnerable.local
```

**Direct access (bypasses OAuth):**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- OAuth2 Proxy: http://localhost:4180
- PostgreSQL: localhost:5432

**Tilt Web UI:**
```
http://localhost:10350
```

## Development Workflow

### Live Reload

Tilt automatically rebuilds and updates when you change code:

**Backend (Go):**
- Edit files in `backend/`
- Tilt syncs changes and rebuilds
- Pod restarts automatically

**Frontend (React):**
- Edit files in `frontend/src/`
- Tilt syncs changes
- Vite HMR updates the browser

**Controller (Go):**
- Edit files in `controller/`
- Tilt rebuilds and restarts

### View Logs

**In Tilt UI:**
- Click on any resource to see logs in real-time

**In Terminal:**
```bash
# Backend logs
tilt logs invulnerable-backend

# Frontend logs
tilt logs invulnerable-frontend

# Controller logs
tilt logs invulnerable-controller

# All logs
tilt logs
```

### Debugging

**Port Forward to PostgreSQL:**
```bash
# Already set up by Tilt on localhost:5432
psql -h localhost -U invulnerable -d invulnerable
# Password: invulnerable
```

**Execute into Pods:**
```bash
# Backend pod
kubectl exec -it -n invulnerable deployment/invulnerable-backend -- sh

# Frontend pod
kubectl exec -it -n invulnerable deployment/invulnerable-frontend -- sh
```

**Check Resources:**
```bash
# All resources
kubectl get all -n invulnerable

# Pods
kubectl get pods -n invulnerable

# ImageScan CRDs
kubectl get imagescans -n invulnerable

# Ingress
kubectl get ingress -n invulnerable
```

### Database Migrations

Migrations run automatically via init container. To run manually:

```bash
# Port forward to PostgreSQL (already done by Tilt)
# Then run migrations from backend directory
cd backend
make migrate-up
```

### Create ImageScan Resources

```bash
# Create a test ImageScan
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: nginx-test
  namespace: invulnerable
spec:
  image: "nginx:latest"
  schedule: "*/5 * * * *"  # Every 5 minutes for testing
  resources:
    requests:
      memory: "256Mi"
      cpu: "250m"
EOF

# Check status
kubectl get imagescans -n invulnerable
kubectl describe imagescan nginx-test -n invulnerable

# Check created CronJob
kubectl get cronjobs -n invulnerable
```

## Enabling HTTPS/TLS (Optional)

If you installed cert-manager (step 3), you can enable HTTPS for local development.

### Using tilt/values-https.yaml

The `tilt/values-https.yaml` file contains HTTPS-specific configuration:
- TLS enabled on ingress
- SSL redirect enabled
- Cookie security enabled
- HTTPS redirect URL for OAuth2

**Option 1: Replace HTTP values**
```bash
# Back up original
cp tilt/values.yaml tilt/values-http.yaml

# Copy HTTPS values
cp tilt/values-https.yaml tilt/values.yaml

# Restart Tilt
tilt down && tilt up
```

**Option 2: Edit Tiltfile to use HTTPS values**
```python
# In Tiltfile, change the flags line:
flags=[
    '--values=./tilt/values-https.yaml',  # Changed from tilt/values.yaml
    # ... rest of flags
]
```

### Update OAuth Redirect URI

If using real Google OAuth credentials, update your OAuth app:
1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Edit your OAuth 2.0 Client ID
3. Update authorized redirect URI to: `https://invulnerable.local/oauth2/callback`

### Access with HTTPS

After enabling HTTPS:
```
https://invulnerable.local
```

**Browser Certificate Warning**: You'll see a security warning because the certificate is self-signed. This is normal for local development:
1. Click "Advanced"
2. Click "Proceed to invulnerable.local (unsafe)"

The warning won't appear in production with a proper certificate from Let's Encrypt.

### Switch Back to HTTP

To switch back to HTTP:
```bash
# Restore HTTP values
cp tilt/values-http.yaml tilt/values.yaml

# Or edit tilt/values.yaml and set:
# oauth2Proxy.config.cookieSecure: false
# oauth2Proxy.config.redirectUrl: "http://invulnerable.local/oauth2/callback"
# ingress.tls: []

# Restart Tilt
tilt down && tilt up
```

## Tilt Commands

```bash
# Start Tilt with web UI
tilt up

# Start without web UI
tilt up --stream

# Run in CI mode (exits when done)
tilt ci

# Stop Tilt (keeps cluster running)
tilt down

# View args
tilt args

# Trigger a rebuild
tilt trigger invulnerable-backend

# Disable a resource
tilt disable invulnerable-frontend

# Enable a resource
tilt enable invulnerable-frontend
```

## Cleanup

### Stop Tilt (Keep Cluster)

```bash
tilt down
```

### Delete kind Cluster

```bash
# Delete the cluster entirely
kind delete cluster --name invulnerable
```

### Remove /etc/hosts Entry

```bash
# Edit /etc/hosts and remove:
# 127.0.0.1 invulnerable.local
sudo nano /etc/hosts
```

## Customization

### Change Images

Edit `Tiltfile` to modify which images are built:

```python
# Example: Skip building scanner
# Comment out the docker_build for scanner
```

### Change Helm Values

Edit `tilt/values.yaml` to customize deployment:

```yaml
# Example: Change backend replicas
backend:
  replicaCount: 2
```

### Use Different Registry

By default, Tilt uses `localhost:5001`. To change:

```bash
tilt up -- --registry=my-registry.io/invulnerable
```

## Troubleshooting

### Ingress Not Working

```bash
# Check ingress controller
kubectl get pods -n ingress-nginx

# Check ingress resource
kubectl describe ingress -n invulnerable

# Test direct access
curl http://localhost:3000  # Should work
curl http://invulnerable.local  # Should work through ingress
```

### OAuth2 Login Not Working

```bash
# Check oauth2-proxy logs
kubectl logs -n invulnerable deployment/invulnerable-oauth2-proxy

# Verify configuration
kubectl get configmap -n invulnerable
kubectl get secret -n invulnerable

# Test oauth2-proxy directly
curl http://localhost:4180/ping
```

### PostgreSQL Connection Issues

```bash
# Check PostgreSQL is running
kubectl get pods -n invulnerable | grep postgres

# Test connection
kubectl run -it --rm psql-test --image=postgres:15 --restart=Never -- \
  psql -h postgres-postgresql.invulnerable.svc.cluster.local -U invulnerable -d invulnerable
```

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n invulnerable

# Describe problematic pod
kubectl describe pod <pod-name> -n invulnerable

# Check events
kubectl get events -n invulnerable --sort-by='.lastTimestamp'
```

### Image Pull Errors

```bash
# Make sure images are built
tilt trigger <resource-name>

# Check Tilt logs
tilt logs <resource-name>
```

## Performance Tips

1. **Disable Unused Resources**: Use `tilt disable` to skip resources you're not working on
2. **Use Snapshots**: Tilt can save cluster state between sessions
3. **Optimize Live Updates**: Adjust `live_update` rules in Tiltfile for faster sync
4. **Resource Limits**: Lower resource requests in `tilt/values.yaml` for faster scheduling

## Additional Resources

- [Tilt Documentation](https://docs.tilt.dev/)
- [kind Documentation](https://kind.sigs.k8s.io/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

## FAQ

**Q: Can I use kind instead of Docker Desktop?**
A: Yes! Change the context in Tiltfile to `allow_k8s_contexts('kind-invulnerable')` and create a kind cluster with the appropriate port mappings.

**Q: How do I reset everything?**
A: `tilt down && kubectl delete namespace invulnerable ingress-nginx cert-manager --ignore-not-found && tilt up`

**Q: Can I use Tilt in production?**
A: No, Tilt is for local development only. Use proper CI/CD pipelines for production.

**Q: How do I add custom resources?**
A: Add them to the Tiltfile using `k8s_yaml()` or `helm_resource()`.

**Q: Why is OAuth2 required?**
A: The Helm chart enforces authentication when ingress is enabled. For local dev without OAuth, use port-forward instead.
