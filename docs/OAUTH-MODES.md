# OAuth Authentication Modes

The application supports two distinct operational modes for authentication:

## Mode 1: No Authentication (Development)

**When to use:** Local development, internal testing, proof-of-concept

**Configuration:**
```yaml
# values.yaml
oauth2Proxy:
  enabled: false
```

**Behavior:**
- Application runs without any authentication
- `/api/v1/user/me` returns `204 No Content`
- All API endpoints are publicly accessible
- **WARNING:** Never expose publicly in this mode!

**Tilt usage:**
```bash
tilt up  # Uses tilt/values.yaml (OAuth disabled by default)
```

---

## Mode 2: Full OAuth with JWT Validation (Production)

**When to use:** Production deployments, internet-facing applications

**Configuration:**
```yaml
# values.yaml
oauth2Proxy:
  enabled: true

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-idp.com/realms/your-realm"  # REQUIRED!
    oidcJwksUrl: ""  # Optional - for cluster-internal JWKS access
    oidcAudience: ""  # Optional - for stricter validation

    # Client credentials
    clientID: "your-client-id"
    clientSecret: "your-client-secret"
    cookieSecret: "your-32-byte-secret"

    redirectUrl: "https://your-domain.com/oauth2/callback"
    emailDomains:
      - "yourcompany.com"  # Restrict to your domain
```

**Behavior:**
- ✅ OAuth2 Proxy enforces authentication at ingress
- ✅ Backend validates JWT tokens cryptographically
- ✅ Email claim verified against headers
- ✅ Network policies prevent direct backend access
- ❌ Returns 401 if JWT validation fails
- ❌ No fallback mechanisms

**Tilt usage (with Dex OIDC):**
```bash
tilt up -- --enable-oidc  # Uses tilt/values-oidc.yaml
```

**Backend environment variables (auto-configured by Helm):**
```bash
OAUTH_ENABLED=true
OIDC_ISSUER_URL=https://your-idp.com/realms/your-realm
OIDC_JWKS_URL=http://internal-service:5556/keys  # Optional
OIDC_AUDIENCE=your-client-id  # Optional
```

---

## Security Enforcement Matrix

| Configuration | OAuth2 Proxy | JWT Validator | Behavior |
|---------------|--------------|---------------|----------|
| `enabled: false` | Not deployed | Not initialized | ✅ No auth required |
| `enabled: true`, no `oidcIssuerUrl` | Deployed | ❌ **Helm fails** | Deployment blocked |
| `enabled: true`, has `oidcIssuerUrl` | Deployed | Initialized | ✅ Full JWT validation |
| `enabled: true`, invalid JWT | Deployed | Initialized | ❌ Returns 401 |

---

## Fail-Fast Validation

The implementation includes multiple layers of validation to prevent misconfiguration:

### 1. Helm Template Validation
```yaml
{{- if .Values.oauth2Proxy.enabled }}
{{- if not .Values.oauth2Proxy.config.oidcIssuerUrl }}
{{- fail "ERROR: oauth2Proxy.enabled=true but oidcIssuerUrl not set" }}
{{- end }}
{{- end }}
```

**Result:** `helm install` fails immediately if OAuth enabled without OIDC config

### 2. Backend Startup Validation
```go
if oauthEnabled {
    issuerURL := getEnv("OIDC_ISSUER_URL", "")
    if issuerURL == "" {
        logger.Fatal("OAUTH_ENABLED=true but OIDC_ISSUER_URL not set")
    }
}
```

**Result:** Backend pod fails to start if misconfigured

### 3. Runtime Request Validation
```go
if oauthEnabled && jwtValidator == nil {
    return 500  // Configuration error
}
if oauthEnabled {
    if err := validateJWT(token); err != nil {
        return 401  // Invalid token
    }
}
```

**Result:** Requests fail with clear error messages

---

## Testing Different Modes

### Test Mode 1: No Authentication
```bash
# Deploy without OAuth
helm install invulnerable ./helm/invulnerable \
  --set oauth2Proxy.enabled=false

# Verify
curl http://invulnerable.local/api/v1/user/me
# Expected: 204 No Content
```

### Test Mode 2: Full OAuth
```bash
# Deploy with OAuth + OIDC
helm install invulnerable ./helm/invulnerable \
  --set oauth2Proxy.enabled=true \
  --set oauth2Proxy.config.oidcIssuerUrl=https://dex.example.com \
  --set oauth2Proxy.clientID=invulnerable \
  --set oauth2Proxy.clientSecret=secret \
  --set oauth2Proxy.cookieSecret=01234567890123456789012345678901

# Verify - should redirect to OAuth provider
curl -I http://invulnerable.local/
# Expected: 302 redirect to OAuth login

# After login, verify JWT validation
curl http://invulnerable.local/api/v1/user/me \
  -H "Cookie: _oauth2_proxy=..."
# Expected: 200 OK with user info (after JWT validation)
```

### Test Invalid Configuration (Should Fail)
```bash
# Try to enable OAuth without OIDC issuer
helm install invulnerable ./helm/invulnerable \
  --set oauth2Proxy.enabled=true

# Expected: Helm error:
# "ERROR: oauth2Proxy.enabled=true but oauth2Proxy.config.oidcIssuerUrl is not set"
```

---

## Migration Guide

### Upgrading from No Auth to OAuth

1. **Configure OIDC provider** (Dex, Keycloak, Auth0, etc.)

2. **Update values.yaml:**
```yaml
oauth2Proxy:
  enabled: true
  config:
    oidcIssuerUrl: "https://your-idp.com"
    # ... other OAuth config
```

3. **Enable Network Policies:**
```yaml
networkPolicy:
  enabled: true
```

4. **Deploy:**
```bash
helm upgrade invulnerable ./helm/invulnerable -f values.yaml
```

5. **Verify:**
- Backend logs show: `"OAuth2 enabled - JWT validation active"`
- Network policy blocks direct pod access
- Application redirects to OAuth login

---

## Troubleshooting

### Problem: 500 Internal Server Error on `/api/v1/user/me`

**Cause:** OAuth enabled but JWT validator not initialized

**Check:**
```bash
kubectl logs -l app.kubernetes.io/component=backend | grep -i oauth
```

**Expected (correct):**
```
{"level":"info","msg":"OAuth2 enabled - JWT validation active","issuer":"..."}
```

**Got (incorrect):**
```
{"level":"error","msg":"OAuth2 enabled but JWT validator not configured"}
```

**Fix:** Ensure `oidcIssuerUrl` is configured in values.yaml

---

### Problem: 401 Unauthorized after OAuth login

**Cause:** JWT validation failing

**Check:**
```bash
kubectl logs -l app.kubernetes.io/component=backend | grep -i "invalid access token"
```

**Common reasons:**
1. OIDC issuer URL mismatch (external vs internal)
   - **Solution:** Set `oidcJwksUrl` to cluster-internal URL
2. Token expired
   - **Solution:** Refresh session or increase token lifetime
3. Wrong signing algorithm
   - **Solution:** Ensure OIDC provider uses RS256
4. Email claim missing
   - **Solution:** Configure OIDC scopes to include `email`

---

## Security Recommendations

1. **Production:** Always use Mode 2 (OAuth enabled)
2. **Email domains:** Restrict to your organization
3. **Network policies:** Enable for defense-in-depth
4. **HTTPS:** Required for production (cookieSecure: true)
5. **Session timeout:** Configure cookie expiration
6. **Monitoring:** Alert on authentication failures

---

## Summary

| Aspect | Mode 1: No Auth | Mode 2: OAuth + JWT |
|--------|-----------------|---------------------|
| **Security** | ❌ None | ✅ Strong |
| **Setup complexity** | ✅ Simple | ⚠️ Moderate |
| **Use case** | Dev/testing | Production |
| **JWT validation** | ❌ Disabled | ✅ Mandatory |
| **Fail-fast** | N/A | ✅ Yes |
| **Fallbacks** | N/A | ❌ None |
| **Public internet** | ❌ Never | ✅ Yes |
