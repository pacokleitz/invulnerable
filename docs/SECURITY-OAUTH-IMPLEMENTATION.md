# OAuth Security Implementation Summary

## ✅ Implemented Security Fixes

### 1. JWT Token Validation (CRITICAL) ✅ IMPLEMENTED
**Files:**
- `/backend/internal/auth/jwt.go` (new)
- `/backend/internal/api/user.go` (updated)
- `/backend/cmd/server/main.go` (updated)
- `/helm/invulnerable/templates/backend-deployment.yaml` (updated)
- `/helm/invulnerable/values.yaml` (updated)

Implemented full JWT validation with cryptographic verification:
- RSA signature verification using JWKS
- Issuer validation (iss claim)
- Audience validation (aud claim, optional)
- Expiration checking (exp claim)
- Email cross-check (JWT email vs header email)
- JWKS caching (1-hour TTL)
- Graceful fallback to token presence check if OIDC not configured

**Configuration:**
```yaml
oauth2Proxy:
  config:
    oidcIssuerUrl: "https://your-idp.com/realms/your-realm"
    oidcAudience: "invulnerable"  # Optional
```

Backend automatically loads `OIDC_ISSUER_URL` from environment and initializes JWT validator.

### 2. Backend Header Validation (CRITICAL)
**File:** `/backend/internal/api/user.go`

Added validation to prevent header forgery attacks:
- Now requires `X-Auth-Request-Access-Token` header to be present
- Prevents attackers from bypassing ingress and forging user headers
- Logs security warnings when forgery attempts are detected

**Before:**
```go
email := c.Request().Header.Get("X-Auth-Request-Email")
// No validation - trusts headers blindly
```

**After:**
```go
email := c.Request().Header.Get("X-Auth-Request-Email")
accessToken := c.Request().Header.Get("X-Auth-Request-Access-Token")

if accessToken == "" {
    // Security: Possible header forgery attempt
    return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication")
}
```

### 2. Network Policy Support (CRITICAL)
**Files:**
- `/helm/invulnerable/templates/backend-networkpolicy.yaml` (new)
- `/helm/invulnerable/values.yaml` (updated)

Added Kubernetes NetworkPolicy to enforce traffic flow through OAuth2 Proxy:
- Restricts backend ingress to ingress controller only
- Prevents pod-to-pod bypass of authentication
- Optional monitoring exception for Prometheus
- Configurable for different ingress controllers

**Usage:**
```yaml
# values.yaml
oauth2Proxy:
  enabled: true

networkPolicy:
  enabled: true  # MUST be true for production
  ingressControllerNamespaceLabel:
    kubernetes.io/metadata.name: ingress-nginx
```

### 3. Security Documentation
**Files:**
- `/docs/SECURITY-REVIEW-OAUTH.md` (new) - Comprehensive security analysis
- `/docs/SECURITY-OAUTH-IMPLEMENTATION.md` (this file) - Implementation guide

---

## ⚠️ Required Production Configuration

Before deploying to production with OAuth2, you **MUST** configure:

### 1. Email Domain Restrictions
```yaml
oauth2Proxy:
  config:
    emailDomains:
      - "yourcompany.com"  # REQUIRED - do not allow wildcard (*)
```

### 2. Session Timeout
```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--cookie-refresh=1h"
      - "--cookie-expire=12h"
```

### 3. Network Policies
```yaml
networkPolicy:
  enabled: true  # REQUIRED
```

### 4. HTTPS/TLS
```yaml
ingress:
  tls:
    - secretName: invulnerable-tls
      hosts:
        - invulnerable.example.com

oauth2Proxy:
  config:
    cookieSecure: true  # Must be true for HTTPS
```

---

## 🔧 Deployment Steps

### Step 1: Generate Secrets
```bash
# Generate cookie secret (exactly 32 bytes)
COOKIE_SECRET=$(openssl rand -base64 32 | head -c 32)

# Create Kubernetes secret
kubectl create secret generic oauth2-proxy-secrets \
  --from-literal=client-id='your-oauth-client-id' \
  --from-literal=client-secret='your-oauth-client-secret' \
  --from-literal=cookie-secret="$COOKIE_SECRET" \
  --namespace invulnerable
```

### Step 2: Configure values.yaml
```yaml
oauth2Proxy:
  enabled: true
  existingSecret: "oauth2-proxy-secrets"

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-idp.com/realms/your-realm"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # CRITICAL: Restrict email domains
    emailDomains:
      - "yourcompany.com"

    # Session security
    cookieSecure: true

    # Session timeout
    extraArgs:
      - "--cookie-refresh=1h"
      - "--cookie-expire=12h"

# CRITICAL: Enable network policies
networkPolicy:
  enabled: true
  ingressControllerNamespaceLabel:
    kubernetes.io/metadata.name: ingress-nginx

# TLS configuration
ingress:
  tls:
    - secretName: invulnerable-tls
      hosts:
        - invulnerable.example.com
```

### Step 3: Deploy
```bash
helm upgrade --install invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --create-namespace \
  --values values.yaml
```

### Step 4: Verify Security

#### Test 1: Verify OAuth Flow
```bash
curl -I https://invulnerable.example.com/
# Should redirect to OAuth provider (302)
```

#### Test 2: Verify Network Policy
```bash
# Try to bypass OAuth by accessing backend directly
kubectl run test-pod --rm -it --image=curlimages/curl -n invulnerable -- \
  curl http://invulnerable-backend:8080/api/v1/user/me

# Should fail with connection timeout (network policy blocking)
```

#### Test 3: Verify Header Validation
```bash
# Try to forge headers (requires access to ingress controller)
kubectl exec -n ingress-nginx <ingress-pod> -- curl \
  -H "X-Auth-Request-Email: fake@attacker.com" \
  http://invulnerable-backend.invulnerable:8080/api/v1/user/me

# Should return 401 Unauthorized (missing access token)
```

#### Test 4: Verify Session Expiration
```bash
# Login and extract cookie
COOKIE=$(curl -c - https://invulnerable.example.com/ | grep _oauth2_proxy)

# Wait 12+ hours (or modify cookie-expire for testing)
# Try to access with old cookie
curl -b "$COOKIE" https://invulnerable.example.com/

# Should redirect to login (session expired)
```

---

## 🔒 Security Checklist

Use this checklist before production deployment:

- [ ] OAuth2 Proxy is enabled (`oauth2Proxy.enabled: true`)
- [ ] Email domain restrictions are configured (no wildcard `*`)
- [ ] Cookie secret is exactly 32 bytes
- [ ] Session timeout is configured (`cookie-expire`)
- [ ] Network policies are enabled (`networkPolicy.enabled: true`)
- [ ] TLS/HTTPS is configured with valid certificates
- [ ] `cookieSecure: true` is set
- [ ] Redirect URL matches OAuth provider configuration exactly
- [ ] Secrets are managed via `existingSecret` (not inline)
- [ ] Backend validates access token cryptographically (JWT validation enabled)
- [ ] OIDC issuer URL is configured correctly
- [ ] Backend logs show "JWT validation enabled" on startup
- [ ] Monitoring/alerting is configured for auth failures
- [ ] Network policy allows ingress controller traffic
- [ ] Network policy blocks direct pod access to backend
- [ ] OAuth provider client credentials are secured
- [ ] Regular secret rotation schedule is established

---

## 🧪 Testing Network Policies

### Before Enabling Network Policies
```bash
# Should succeed - no restrictions
kubectl run test-pod --rm -it --image=curlimages/curl -n invulnerable -- \
  curl http://invulnerable-backend:8080/api/v1/metrics
```

### After Enabling Network Policies
```bash
# Should timeout - blocked by network policy
kubectl run test-pod --rm -it --image=curlimages/curl -n invulnerable -- \
  curl --connect-timeout 5 http://invulnerable-backend:8080/api/v1/metrics

# Output: "curl: (28) Connection timeout after 5000 ms"
```

### Allow Traffic from Specific Pod
```yaml
networkPolicy:
  additionalIngressRules:
    - from:
      - podSelector:
          matchLabels:
            app: my-monitoring-app
      ports:
      - protocol: TCP
        port: 8080
```

---

## 📊 Monitoring & Alerting

### Key Metrics to Monitor

1. **Authentication Failures**
   ```promql
   rate(oauth2_proxy_authentication_errors_total[5m])
   ```

2. **Session Expirations**
   ```promql
   rate(oauth2_proxy_cookie_refresh_total[5m])
   ```

3. **Header Forgery Attempts**
   - Monitor backend logs for warnings about missing access tokens
   - Alert on: `"possible header forgery attempt"`

4. **Network Policy Drops**
   - Monitor pod network connection failures
   - May require CNI-specific monitoring (e.g., Calico, Cilium)

### Example Prometheus Alert
```yaml
- alert: OAuth2HeaderForgeryAttempt
  expr: |
    increase(
      log_messages_total{
        level="warn",
        message=~".*header forgery attempt.*"
      }[5m]
    ) > 0
  annotations:
    summary: "Possible OAuth2 header forgery detected"
    description: "Someone attempted to forge OAuth2 headers"
```

---

## 🔄 Secret Rotation

### Rotate Cookie Secret
```bash
# Generate new secret
NEW_SECRET=$(openssl rand -base64 32 | head -c 32)

# Update Kubernetes secret
kubectl patch secret oauth2-proxy-secrets -n invulnerable \
  -p "{\"data\":{\"cookie-secret\":\"$(echo -n $NEW_SECRET | base64)\"}}"

# Restart OAuth2 Proxy to pick up new secret
kubectl rollout restart deployment/invulnerable-oauth2-proxy -n invulnerable

# Note: This will invalidate all existing sessions
```

### Rotate OAuth Client Credentials
```bash
# 1. Create new client credentials in OAuth provider
# 2. Update Kubernetes secret
kubectl patch secret oauth2-proxy-secrets -n invulnerable \
  -p "{\"data\":{\"client-secret\":\"$(echo -n $NEW_CLIENT_SECRET | base64)\"}}"

# 3. Restart OAuth2 Proxy
kubectl rollout restart deployment/invulnerable-oauth2-proxy -n invulnerable
```

---

## 🚨 Incident Response

### Suspected Header Forgery Attack
1. Check backend logs for forgery warnings
2. Verify network policies are enabled
3. Verify backend version includes header validation
4. Review network policy configuration
5. Check for unauthorized access to ingress controller

### Suspected Session Hijacking
1. Rotate cookie secret immediately (invalidates all sessions)
2. Review access logs for suspicious activity
3. Enable `--request-logging=true` in OAuth2 Proxy
4. Reduce `cookie-expire` time
5. Consider implementing IP-based session binding

### OAuth Provider Compromise
1. Revoke client credentials at OAuth provider
2. Generate new client credentials
3. Update Kubernetes secret
4. Restart OAuth2 Proxy
5. Force all users to re-authenticate
6. Review audit logs for unauthorized access

---

## 📚 Additional Resources

- [OAuth2 Proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Security Review Document](./SECURITY-REVIEW-OAUTH.md)
- [Authentication Setup Guide](./AUTHENTICATION.md)
