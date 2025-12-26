# Tilt Quick Start

## 🎯 One-Command Setup

```bash
# 1. Enable Kubernetes in Docker Desktop, then run setup
./scripts/setup-tilt.sh

# 2. Start Tilt in your preferred mode:

# HTTP mode (default - requires Google OAuth)
tilt up

# HTTPS mode (with cert-manager)
tilt up -- --enable-https=true

# OIDC mode (with local Dex provider - best for testing!)
tilt up -- --enable-oidc=true

# 3. Access at http://localhost:10350 (Tilt UI)
#    and http://invulnerable.local (application)
```

## 🚀 What Tilt Deploys

Tilt automatically deploys everything you need:

### Infrastructure (automatic)
- ✅ **nginx Ingress Controller** - Routes traffic to services
- ✅ **PostgreSQL** - Database
- ⚙️ **cert-manager** - TLS certificates (only with `--enable-https`)
- ⚙️ **Dex OIDC Provider** - Local authentication (only with `--enable-oidc`)

### Application (automatic)
- ✅ **Backend** - Go API server with live reload
- ✅ **Frontend** - React UI with Vite HMR
- ✅ **Controller** - ImageScan CRD manager
- ✅ **Scanner** - Vulnerability scanner (image only)
- ✅ **OAuth2 Proxy** - Authentication layer

## 🔧 Modes

### HTTP Mode (Default)
```bash
tilt up
```
- Access: `http://invulnerable.local`
- Auth: Google OAuth (requires real credentials)
- Fast setup, no certificate warnings

### HTTPS Mode
```bash
tilt up -- --enable-https=true
```
- Access: `https://invulnerable.local`
- Auth: Google OAuth with TLS
- Self-signed certificate (browser warning expected)

### OIDC Mode (Recommended for Testing)
```bash
tilt up -- --enable-oidc=true
```
- Access: `http://invulnerable.local`
- Auth: Local Dex OIDC provider
- **No external OAuth setup required!**
- Test users:
  - `admin@invulnerable.local / password`
  - `user@invulnerable.local / password`
  - `test@example.com / password`
- See [OIDC-SETUP.md](OIDC-SETUP.md) for details

## 📚 Full Documentation

See [TILT.md](TILT.md) for complete documentation.

## 💡 Tips

**View logs**: Click any resource in Tilt UI (http://localhost:10350)
**Bypass OAuth**: Use direct ports (localhost:3000, localhost:8080)
**Switch modes**: `tilt down`, then `tilt up` with/without `--enable-https`
