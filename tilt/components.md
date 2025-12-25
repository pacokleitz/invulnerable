# Tilt Components Overview

This document explains all the components included in the Tilt setup for local development.

## 🏗️ Infrastructure Components

### 1. **nginx Ingress Controller** ✅ Auto-installed

**What it does**: Routes external HTTP/HTTPS traffic to your services based on hostname and path.

**Installation**: Automatically installed by `setup-tilt.sh`

**Configuration**:
```yaml
# Installed in ingress-nginx namespace
# Watches for Ingress resources
# Maps port 80 (HTTP) and 443 (HTTPS) from host to cluster
```

**Verify**:
```bash
kubectl get pods -n ingress-nginx
# Should see: ingress-nginx-controller pod running
```

**Why you need it**: Ingress is required to access your application at `http://invulnerable.local` instead of using port-forwards.

---

### 2. **cert-manager** ⚙️ Optional

**What it does**: Automatically provisions and manages TLS certificates for HTTPS.

**Installation**: Optional during `setup-tilt.sh` (prompts you), or install manually (see TILT.md step 3)

**Configuration**:
- Creates self-signed ClusterIssuer for local dev
- Generates TLS certificate for `invulnerable.local`
- Stores certificate in `invulnerable-tls` secret

**When to use**:
- ✅ Testing HTTPS locally
- ✅ Testing OAuth flows with HTTPS (some providers require it)
- ✅ Debugging TLS issues
- ❌ Not needed for basic HTTP development

**Verify**:
```bash
# Check cert-manager is installed
kubectl get pods -n cert-manager

# Check certificate is ready
kubectl get certificate -n invulnerable
# Should see: invulnerable-local-cert READY=True
```

**Why it's optional**: Local development can use HTTP, but some features (like certain OAuth providers) work better with HTTPS.

---

### 3. **PostgreSQL** ✅ Auto-deployed by Tilt

**What it does**: Database for storing scan results, vulnerabilities, and metadata.

**Installation**: Deployed automatically by Tilt using Bitnami Helm chart

**Configuration**:
- Database: `invulnerable`
- Username: `invulnerable`
- Password: `invulnerable`
- Port: 5432
- Persistence: Disabled for faster dev cycles

**Verify**:
```bash
kubectl get pods -n invulnerable | grep postgres
# Should see: postgres-postgresql pod running
```

**Access**:
```bash
# Already port-forwarded by Tilt to localhost:5432
psql -h localhost -U invulnerable -d invulnerable
# Password: invulnerable
```

---

## 🎯 Application Components

### 4. **Backend** (Go + Echo)

**Deployed by Tilt**: Yes
**Live Reload**: Yes - syncs Go code changes and rebuilds
**Port Forward**: 8080 → localhost:8080

**Image**: Built from `backend/Dockerfile`
**Registry**: `localhost:5001/invulnerable-backend`

---

### 5. **Frontend** (React + Vite)

**Deployed by Tilt**: Yes
**Live Reload**: Yes - Vite HMR for instant updates
**Port Forward**: 8080 → localhost:3000

**Image**: Built from `frontend/Dockerfile`
**Registry**: `localhost:5001/invulnerable-frontend`

---

### 6. **Controller** (Kubernetes CRD)

**Deployed by Tilt**: Yes
**Live Reload**: Yes - rebuilds on Go changes
**Port Forward**: None (background process)

**What it does**: Watches ImageScan CRDs and creates CronJobs for scanning

**Image**: Built from `controller/Dockerfile`
**Registry**: `localhost:5001/invulnerable-controller`

---

### 7. **Scanner** (Syft + Grype)

**Deployed by Tilt**: Image built only
**Live Reload**: No (runs as CronJobs)

**What it does**: Scans container images for vulnerabilities
**Image**: Built from `scanner/Dockerfile`
**Registry**: `localhost:5001/invulnerable-scanner`

**Note**: Scanner doesn't run continuously - it's triggered by CronJobs created by the controller.

---

## 🔐 Security Components

### 8. **OAuth2 Proxy** ✅ Auto-deployed by Tilt

**What it does**: Provides authentication layer for the application

**Configuration**: Configured in `tilt/values.yaml`
- Default: Allows any email (insecure, dev only)
- Can be configured with real Google/GitHub/etc. OAuth

**Port Forward**: 4180 → localhost:4180

**Why you need it**: The Helm chart enforces authentication when ingress is enabled. OAuth2 Proxy satisfies this requirement.

---

## 📋 Component Matrix

| Component | Required? | Auto-installed? | Purpose |
|-----------|-----------|-----------------|---------|
| **nginx-ingress** | ✅ Yes | ✅ Yes (setup script) | Routes traffic to services |
| **cert-manager** | ⚙️ Optional | ⚙️ Optional (prompted) | HTTPS/TLS certificates |
| **PostgreSQL** | ✅ Yes | ✅ Yes (Tilt) | Database |
| **Backend** | ✅ Yes | ✅ Yes (Tilt) | API server |
| **Frontend** | ✅ Yes | ✅ Yes (Tilt) | Web UI |
| **Controller** | ✅ Yes | ✅ Yes (Tilt) | ImageScan CRD manager |
| **Scanner** | ✅ Yes | ✅ Yes (Tilt) | Vulnerability scanner |
| **OAuth2 Proxy** | ✅ Yes | ✅ Yes (Tilt) | Authentication |

---

## 🔄 Development Modes

### Mode 1: HTTP (Default)

**Files used**:
- `tilt/values.yaml` (HTTP config)

**Components**:
- ✅ nginx-ingress
- ❌ cert-manager (not needed)
- ✅ All application components
- ✅ OAuth2 Proxy (HTTP mode)

**Access**: `http://invulnerable.local`

**Best for**: Fast development, no HTTPS needed

---

### Mode 2: HTTPS

**Files used**:
- `tilt/values-https.yaml` (HTTPS config)

**Components**:
- ✅ nginx-ingress
- ✅ cert-manager (required)
- ✅ All application components
- ✅ OAuth2 Proxy (HTTPS mode)

**Access**: `https://invulnerable.local`

**Best for**: Testing OAuth providers that require HTTPS, debugging TLS issues

---

## 🚀 Quick Reference

### Start Development (HTTP)

```bash
./scripts/setup-tilt.sh
# Choose "N" for cert-manager (HTTP only)
tilt up
# Access: http://invulnerable.local
```

### Start Development (HTTPS)

```bash
./scripts/setup-tilt.sh
# Choose "Y" for cert-manager
# Edit Tiltfile to use tilt/values-https.yaml OR:
cp tilt/values-https.yaml tilt/values.yaml
tilt up
# Access: https://invulnerable.local
```

### Check What's Installed

```bash
# nginx-ingress
kubectl get pods -n ingress-nginx

# cert-manager (if installed)
kubectl get pods -n cert-manager

# Application components
kubectl get pods -n invulnerable

# ImageScan CRDs
kubectl get imagescans -n invulnerable

# All ingress resources
kubectl get ingress -A
```

---

## 📚 Additional Resources

- **Full Setup Guide**: TILT.md
- **Quick Reference**: .tilt-reference.md
- **Main README**: README.md (Development section)

---

## ❓ FAQ

**Q: Do I need cert-manager?**
A: No, not for HTTP development. Only if you want HTTPS locally.

**Q: Can I disable OAuth2 Proxy?**
A: No, the Helm chart requires it when ingress is enabled. But you can bypass it using direct port-forwards (localhost:3000, localhost:8080).

**Q: Why does the setup script prompt for cert-manager?**
A: It's optional. We ask so you can choose HTTP (faster, simpler) or HTTPS (more realistic) mode.

**Q: Can I use a real domain instead of invulnerable.local?**
A: Yes, but you'll need to configure DNS or modify /etc/hosts, and update the Tilt values files.

**Q: What's the difference between tilt/values.yaml and tilt/values-https.yaml?**
A:
- `tilt/values.yaml`: HTTP, no TLS, cookieSecure=false
- `tilt/values-https.yaml`: HTTPS, TLS enabled, cookieSecure=true, requires cert-manager
