# Tilt Quick Reference

## 🎯 What is Tilt?

Tilt is a local development tool that:
- Works with Docker Desktop Kubernetes
- Builds Docker images with live reload
- Deploys your Helm chart automatically
- Provides a web UI for logs and resources
- Enables fast inner-loop development

## 📁 Files Created

```
.
├── Tiltfile                    # Main Tilt configuration
├── tilt/
│   ├── values.yaml             # Helm values for local dev (HTTP)
│   ├── values-https.yaml       # Helm values for HTTPS mode
│   ├── postgres-values.yaml    # PostgreSQL config
│   ├── README.md               # Quick start guide
│   ├── reference.md            # This file
│   ├── components.md           # Component overview
│   └── ingress-explained.md    # Ingress architecture
├── .tiltignore                 # Files to ignore for live updates
├── TILT.md                     # Complete documentation
└── scripts/
    └── setup-tilt.sh          # One-time setup script
```

## 🚀 Quick Start

```bash
# 1. Enable Kubernetes in Docker Desktop Settings

# 2. Run setup script
./scripts/setup-tilt.sh

# 3. Start Tilt
tilt up

# Access at http://invulnerable.local
```

## 🔑 Key Features

### Live Reload Configured

**Backend (Go):**
- Syncs: `backend/cmd/`, `backend/internal/`
- Auto-rebuilds on `.go` file changes
- Hot reload in cluster

**Frontend (React):**
- Syncs: `frontend/src/`, `frontend/index.html`
- Vite HMR enabled
- Instant updates

**Controller (Go):**
- Syncs: `controller/api/`, `controller/internal/`
- Auto-rebuilds on changes
- Restarts automatically

### Port Forwards

Automatically set up:
- `localhost:3000` → Frontend
- `localhost:8080` → Backend API
- `localhost:4180` → OAuth2 Proxy
- `localhost:5432` → PostgreSQL

### Resources Deployed

- ✅ PostgreSQL (Bitnami chart)
- ✅ Backend (with auto-scaling disabled)
- ✅ Frontend
- ✅ Controller (with leader election disabled)
- ✅ OAuth2 Proxy (configured for local dev)
- ✅ Ingress (nginx)
- ✅ ImageScan CRDs

## 🌐 Access Points

| Service | URL | Notes |
|---------|-----|-------|
| **Tilt UI** | http://localhost:10350 | Dashboard for logs/resources |
| **Application (via ingress)** | http://invulnerable.local | With OAuth2 authentication |
| **Frontend (direct)** | http://localhost:3000 | Bypasses OAuth |
| **Backend (direct)** | http://localhost:8080 | Bypasses OAuth |
| **PostgreSQL** | localhost:5432 | user: invulnerable, pass: invulnerable |

## 📝 Common Commands

```bash
# Start Tilt with UI
tilt up

# Start without UI (stream logs)
tilt up --stream

# Stop Tilt (keeps cluster)
tilt down

# View logs
tilt logs <resource-name>

# Trigger rebuild
tilt trigger <resource-name>

# Disable/enable resource
tilt disable <resource-name>
tilt enable <resource-name>

# Clean everything
tilt down
kubectl delete namespace invulnerable ingress-nginx cert-manager --ignore-not-found
```

## 🔧 Configuration

### OAuth2 Proxy

Default config in `tilt/values.yaml`:
- Provider: Google OAuth
- Allows ANY email (emailDomains: "*")
- Cookie security disabled for HTTP
- **⚠️ INSECURE - for local dev only!**

**To use real Google OAuth:**
1. Create OAuth app at https://console.cloud.google.com/apis/credentials
2. Add redirect URI: `http://invulnerable.local/oauth2/callback`
3. Update `tilt/values.yaml`:
   ```yaml
   oauth2Proxy:
     clientID: "your-id.apps.googleusercontent.com"
     clientSecret: "your-secret"
     config:
       emailDomains:
         - "yourdomain.com"
   ```

### Resource Limits

Scaled down for local dev:
- Backend: 128Mi-512Mi RAM, 100m-500m CPU
- Frontend: 64Mi-256Mi RAM, 50m-200m CPU
- Controller: 64Mi-256Mi RAM, 50m-200m CPU
- PostgreSQL: 128Mi-512Mi RAM, 100m-500m CPU

### Disable Components

Edit `Tiltfile` to comment out unwanted components:
```python
# Example: Skip controller
# docker_build(
#   ref=f'{registry}/invulnerable-controller',
#   ...
# )
```

## 🐛 Troubleshooting

### "Kubernetes not available"
```bash
# Enable Kubernetes in Docker Desktop:
# Settings > Kubernetes > Enable Kubernetes

# Verify it's running
kubectl cluster-info --context docker-desktop
```

### "ingress not working"
```bash
kubectl get pods -n ingress-nginx
# Should see controller pod running
```

### "invulnerable.local not resolving"
```bash
# Check /etc/hosts
grep invulnerable /etc/hosts
# Should see: 127.0.0.1 invulnerable.local

# If missing:
echo "127.0.0.1 invulnerable.local" | sudo tee -a /etc/hosts
```

### "OAuth login fails"
```bash
# Use direct access (bypasses OAuth)
open http://localhost:3000

# Or check oauth2-proxy logs
kubectl logs -n invulnerable deployment/invulnerable-oauth2-proxy
```

### "PostgreSQL connection error"
```bash
# Check PostgreSQL is running
kubectl get pods -n invulnerable | grep postgres

# Test connection
kubectl port-forward -n invulnerable svc/postgres-postgresql 5432:5432
psql -h localhost -U invulnerable -d invulnerable
```

## 📚 Learn More

- Full documentation: [TILT.md](../TILT.md)
- Tilt docs: https://docs.tilt.dev/
- Docker Desktop docs: https://docs.docker.com/desktop/kubernetes/
- Helm docs: https://helm.sh/docs/

## 💡 Pro Tips

1. **Keep Tilt UI open** - Best way to monitor logs and resources
2. **Use direct ports** - Faster than ingress for development
3. **Disable unused resources** - Speeds up development cycle
4. **Use Tilt snapshots** - Save cluster state between sessions
5. **Customize live_update** - Adjust sync patterns for your workflow

## 🔄 Development Workflow

```bash
# 1. Start Tilt (one time)
tilt up

# 2. Make changes to code
vim backend/internal/api/handler.go

# 3. Tilt auto-rebuilds and reloads
# Watch progress in Tilt UI: http://localhost:10350

# 4. Test changes
curl http://localhost:8080/api/v1/metrics

# 5. When done
tilt down
```

## ⚡ Performance

Tilt with live_update is **much faster** than regular rebuilds:

| Method | Backend Rebuild | Frontend Rebuild |
|--------|----------------|------------------|
| **Without Tilt** | 30-60s | 30-45s |
| **With Tilt live_update** | 2-5s | <1s (HMR) |

## 🎓 Next Steps

1. Read [TILT.md](TILT.md) for complete documentation
2. Configure real OAuth credentials for testing auth flows
3. Create ImageScan CRDs to test scanning
4. Explore Tilt UI features and resource management
5. Customize Tiltfile for your workflow
