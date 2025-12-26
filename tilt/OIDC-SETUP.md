# OpenID Connect (OIDC) Local Testing with Dex

This guide explains how to test OpenID Connect authentication locally using Dex as your OIDC provider.

## What is Dex?

[Dex](https://dexidp.io/) is an open-source OIDC (OpenID Connect) identity provider that:
- Acts as a portal to other identity providers (LDAP, SAML, OAuth2)
- Provides OIDC protocol support
- Perfect for local development and testing
- Lightweight and easy to configure

## Quick Start

### 1. Add Dex to /etc/hosts

```bash
# Add dex.invulnerable.local to your hosts file
echo "127.0.0.1 dex.invulnerable.local" | sudo tee -a /etc/hosts
```

### 2. Start Tilt with OIDC Enabled

```bash
# Start with OIDC mode
tilt up -- --enable-oidc=true

# Wait for all resources to be ready
# Open Tilt UI: http://localhost:10350
```

### 3. Test the Login Flow

```bash
# Open the application
open http://invulnerable.local

# You'll be redirected to Dex login page
# Use one of the test accounts:
# - admin@invulnerable.local / password
# - user@invulnerable.local / password
# - test@example.com / password
```

## What Gets Deployed

When you use `--enable-oidc`, Tilt deploys:

1. **Dex OIDC Provider** (namespace: `dex`)
   - Issuer URL: `http://dex.invulnerable.local/dex`
   - Access at: `http://dex.invulnerable.local`

2. **OAuth2 Proxy** (configured for OIDC)
   - Provider: `oidc`
   - Client ID: `invulnerable-local`
   - Scopes: `openid profile email groups`

3. **Ingress Routes**
   - `http://invulnerable.local` → protected by OIDC
   - `http://dex.invulnerable.local` → Dex OIDC provider

## Test Users

Three users are pre-configured in `tilt/dex-values.yaml`:

| Email | Username | Password | User ID |
|-------|----------|----------|---------|
| admin@invulnerable.local | admin | password | 08a8684b-db88-4b73-90a9-3cd1661f5466 |
| user@invulnerable.local | user | password | 41331323-6f44-45e6-b3b9-2c4b60c02be5 |
| test@example.com | testuser | password | f5d3c1a2-8b7e-4c9d-a3f1-2e6b8c9d0e1f |

All passwords are hashed with bcrypt.

### Adding More Users

Edit `tilt/dex-values.yaml`:

```yaml
staticPasswords:
- email: "newuser@example.com"
  # Generate hash: echo password | htpasswd -BinC 10 admin | cut -d: -f2
  hash: "$2a$10$..."
  username: "newuser"
  userID: "unique-uuid-here"
```

Then restart Tilt:
```bash
tilt down
tilt up -- --enable-oidc=true
```

## OAuth2 Proxy Configuration

The OIDC configuration is in `tilt/values-oidc.yaml`:

```yaml
oauth2Proxy:
  config:
    provider: "oidc"
    oidcIssuerUrl: "http://dex.invulnerable.local/dex"
    scope: "openid profile email groups"

    # Pass user info to backend
    extraArgs:
      - "--pass-access-token=true"
      - "--pass-user-headers=true"
```

## Testing API Access with OIDC

### Option 1: Use Browser Cookie

```bash
# 1. Log in via browser: http://invulnerable.local
# 2. Open DevTools > Application > Cookies
# 3. Copy the _oauth2_proxy_oidc cookie value

# 4. Use it in API requests
curl http://invulnerable.local/api/v1/metrics \
  -H "Cookie: _oauth2_proxy_oidc=<cookie-value>"
```

### Option 2: Get Access Token

OAuth2 proxy forwards the OIDC access token to your backend in headers:

```yaml
# These headers are available in your backend:
X-Auth-Request-User: admin@invulnerable.local
X-Auth-Request-Email: admin@invulnerable.local
X-Auth-Request-Access-Token: <jwt-token>
Authorization: Bearer <jwt-token>
```

### Option 3: Direct Backend Access (No Auth)

For development, bypass OAuth entirely:

```bash
# Use direct port-forward
curl http://localhost:8080/api/v1/metrics
```

## Verifying OIDC Flow

### 1. Check Dex is Running

```bash
kubectl get pods -n dex
# Should show: dex-xxxxx Running

# Check Dex logs
kubectl logs -n dex -l app.kubernetes.io/name=dex

# Test Dex discovery
curl http://dex.invulnerable.local/dex/.well-known/openid-configuration
```

### 2. Check OAuth2 Proxy Configuration

```bash
kubectl get pods -n invulnerable | grep oauth2-proxy
# Should show: invulnerable-oauth2-proxy-xxxxx Running

# Check OAuth2 proxy logs
kubectl logs -n invulnerable -l app.kubernetes.io/name=oauth2-proxy
```

### 3. Test the Full Flow

```bash
# 1. Access the app (should redirect to Dex)
curl -v http://invulnerable.local 2>&1 | grep -i location

# 2. You should see redirect to:
# Location: http://invulnerable.local/oauth2/start?rd=%2F

# 3. Follow the OAuth flow
curl -L http://invulnerable.local
# This will show the Dex login page HTML
```

## Customizing Dex Configuration

### Add Custom OIDC Claims

Edit `tilt/dex-values.yaml`:

```yaml
staticPasswords:
- email: "admin@invulnerable.local"
  hash: "$2a$10$..."
  username: "admin"
  userID: "08a8684b-db88-4b73-90a9-3cd1661f5466"
  # Add custom claims
  groups:
    - "admins"
    - "developers"
  extra:
    department: "engineering"
    role: "admin"
```

### Configure Different OAuth2 Client

Edit `tilt/dex-values.yaml`:

```yaml
staticClients:
- id: invulnerable-local
  redirectURIs:
  - 'http://invulnerable.local/oauth2/callback'
  name: 'Invulnerable Local'
  secret: dex-local-secret
  # Add scopes
  trustedPeers:
  - other-client-id
```

## Switching Between Auth Modes

### HTTP Mode (Default - No OIDC)

```bash
tilt up
# Uses: tilt/values.yaml
# Auth: Google OAuth (requires real credentials)
```

### HTTPS Mode

```bash
tilt up -- --enable-https=true
# Uses: tilt/values-https.yaml
# Auth: Google OAuth with TLS
```

### OIDC Mode (Dex)

```bash
tilt up -- --enable-oidc=true
# Uses: tilt/values-oidc.yaml
# Auth: Local Dex OIDC provider
```

## Troubleshooting

### "Failed to fetch OIDC discovery"

```bash
# Check Dex is accessible
curl http://dex.invulnerable.local/dex/.well-known/openid-configuration

# Check /etc/hosts
grep dex.invulnerable.local /etc/hosts

# Check ingress
kubectl get ingress -n dex
```

### "Invalid redirect URI"

The redirect URI in Dex must match exactly:
- Dex config: `http://invulnerable.local/oauth2/callback`
- OAuth2 proxy: `redirectUrl: "http://invulnerable.local/oauth2/callback"`

Edit `tilt/dex-values.yaml` if needed.

### "Login successful but still redirected"

Check OAuth2 proxy logs:

```bash
kubectl logs -n invulnerable -l app.kubernetes.io/name=oauth2-proxy --tail=50

# Look for errors like:
# - Cookie issues
# - OIDC token validation failures
# - Email domain restrictions
```

### "Cannot access Dex UI"

```bash
# Check Dex ingress
kubectl get ingress -n dex

# Check Dex service
kubectl get svc -n dex

# Port-forward directly to Dex
kubectl port-forward -n dex svc/dex 5556:5556
open http://localhost:5556/dex
```

## OIDC vs OAuth2

**OAuth2** (what you were using before):
- Authorization protocol
- Gets access tokens
- Requires real provider (Google, GitHub, etc.)
- Good for production

**OIDC** (what this setup uses):
- Authentication protocol built on OAuth2
- Gets ID tokens (JWT) with user info
- Can run locally (like Dex)
- Perfect for development and testing

## Production Considerations

This Dex setup is **for local development only**. For production:

1. **Use real OIDC providers**:
   - Google Identity Platform
   - Auth0
   - Okta
   - Azure AD
   - Keycloak (self-hosted)

2. **Enable HTTPS**:
   ```bash
   tilt up -- --enable-oidc=true --enable-https=true
   ```

3. **Use strong secrets**:
   - Generate random client secrets
   - Use proper cookie secrets (32+ chars)
   - Enable cookie security

4. **Restrict email domains**:
   ```yaml
   emailDomains:
     - "yourcompany.com"  # Not "*"
   ```

## Learn More

- [Dex Documentation](https://dexidp.io/docs/)
- [OAuth2 Proxy OIDC Guide](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/oauth_provider#openid-connect-provider)
- [OpenID Connect Spec](https://openid.net/connect/)
- [JWT.io](https://jwt.io/) - Debug JWT tokens

## Next Steps

1. **Test the login flow** with different users
2. **Inspect JWT tokens** in browser DevTools
3. **Add custom claims** and verify they're passed to backend
4. **Test API access** with cookies and tokens
5. **Try group-based authorization** by adding groups to users
