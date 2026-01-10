# OAuth Security Review - Invulnerable Kubernetes Deployment

**Review Date:** 2026-01-10
**OAuth2 Proxy Version:** v7.13.0
**Reviewer:** Security Analysis

## Executive Summary

The OAuth2 implementation in the Invulnerable Helm chart is **generally secure** with proper use of OAuth2 Proxy and HTTPS. However, there are several **critical security recommendations** that should be implemented before production deployment.

**Overall Security Rating:** ⚠️ **Requires Hardening**

---

## ✅ Security Strengths

### 1. **OAuth2 Proxy Security Patches**
- ✅ Uses OAuth2 Proxy v7.13.0 with recent CVE fixes:
  - CVE-2025-47912
  - CVE-2025-58183
  - CVE-2025-58186
  - CVE-2025-64484

### 2. **Secure Cookie Configuration**
```yaml
- --cookie-httponly=true        # ✅ Prevents XSS attacks
- --cookie-samesite=lax         # ✅ CSRF protection
- --cookie-secure=true          # ✅ HTTPS-only (default)
```

### 3. **Proper Secret Management**
- ✅ Secrets stored in Kubernetes Secrets (not inline)
- ✅ Support for `existingSecret` to use external secret management
- ✅ Secrets mounted as environment variables (not command-line args)

### 4. **Authentication Header Propagation**
```yaml
- --set-xauthrequest=true           # ✅ Sets user headers
- --pass-access-token=true          # ✅ Passes OAuth access token
- --pass-authorization-header=true  # ✅ Passes auth header
- --pass-user-headers=true          # ✅ Passes user info
```

### 5. **Ingress Security**
- ✅ Separate ingress for OAuth2 endpoints
- ✅ TLS/HTTPS enforced by default
- ✅ Proper nginx auth-url and auth-signin annotations

---

## ⚠️ Critical Security Issues & Recommendations

### 1. **🔴 CRITICAL: Backend Header Trust Without Validation**

**Issue:** The backend blindly trusts `X-Auth-Request-Email` and `X-Auth-Request-User` headers without validation.

**File:** `/backend/internal/api/user.go:28-29`
```go
email := c.Request().Header.Get("X-Auth-Request-Email")
username := c.Request().Header.Get("X-Auth-Request-User")
```

**Risk:** If an attacker can bypass the ingress (e.g., through misconfigured network policies, pod-to-pod access, or service mesh), they can forge these headers and impersonate any user.

**Recommendation:**
```go
// RECOMMENDED: Add header validation
func (h *UserHandler) GetCurrentUser(c echo.Context) error {
    // Verify request came through OAuth2 Proxy by checking for access token
    accessToken := c.Request().Header.Get("X-Auth-Request-Access-Token")
    if accessToken == "" {
        h.logger.Warn("missing access token, possible header forgery attempt")
        return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
    }

    email := c.Request().Header.Get("X-Auth-Request-Email")
    username := c.Request().Header.Get("X-Auth-Request-User")

    if email == "" {
        h.logger.Warn("missing email header despite access token present")
        return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication")
    }

    // Optional: Verify access token signature if using JWT
    // This provides cryptographic proof of authentication

    return c.JSON(http.StatusOK, UserResponse{
        Email:    email,
        Username: username,
    })
}
```

**Alternative:** Use JWT token validation for stronger security.

---

### 2. **🔴 CRITICAL: No Network Policies to Enforce Traffic Flow**

**Issue:** No NetworkPolicy resources to ensure traffic can only reach the backend through OAuth2 Proxy.

**Risk:** Without network policies, pods in the cluster can directly call the backend service, bypassing OAuth2 Proxy authentication entirely.

**Recommendation:** Add NetworkPolicy to restrict backend access:

```yaml
# helm/invulnerable/templates/backend-networkpolicy.yaml
{{- if and .Values.oauth2Proxy.enabled .Values.networkPolicy.enabled }}
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: {{ include "invulnerable.fullname" . }}-backend
  namespace: {{ .Release.Namespace }}
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: backend
  policyTypes:
  - Ingress
  ingress:
  # Only allow traffic from ingress controller (which enforces OAuth)
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx  # Adjust based on your ingress controller namespace
    ports:
    - protocol: TCP
      port: 8080
  # Allow metrics scraping from monitoring
  - from:
    - namespaceSelector:
        matchLabels:
          name: monitoring  # If you have Prometheus
    ports:
    - protocol: TCP
      port: 8080
{{- end }}
```

Add to `values.yaml`:
```yaml
networkPolicy:
  enabled: true  # MUST be true for production with OAuth
```

---

### 3. **🟡 HIGH: Cookie Secret Length Not Validated**

**Issue:** The Helm chart doesn't validate that cookie secret is exactly 32 bytes.

**File:** `helm/invulnerable/templates/oauth2-proxy-secret.yaml`

**Risk:** Invalid cookie secret length can cause authentication failures or security issues.

**Recommendation:** Add validation in `_helpers.tpl`:

```yaml
{{- define "invulnerable.validateOAuth2Config" -}}
{{- if .Values.oauth2Proxy.enabled }}
  {{- if and (not .Values.oauth2Proxy.existingSecret) }}
    {{- $cookieSecret := .Values.oauth2Proxy.cookieSecret }}
    {{- if ne (len $cookieSecret) 32 }}
      {{- fail "oauth2Proxy.cookieSecret must be exactly 32 characters. Generate with: openssl rand -base64 32 | head -c 32" }}
    {{- end }}
    {{- if eq .Values.oauth2Proxy.clientID "" }}
      {{- fail "oauth2Proxy.clientID is required when oauth2Proxy is enabled" }}
    {{- end }}
    {{- if eq .Values.oauth2Proxy.clientSecret "" }}
      {{- fail "oauth2Proxy.clientSecret is required when oauth2Proxy is enabled" }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}
```

Call in deployment:
```yaml
{{- include "invulnerable.validateOAuth2Config" . }}
```

---

### 4. **🟡 HIGH: No Session Timeout Configuration**

**Issue:** No explicit session timeout or token refresh configuration.

**Risk:** Sessions may last indefinitely, increasing the window for session hijacking.

**Recommendation:** Add to `values.yaml`:

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--cookie-refresh=1h"        # Refresh cookie every hour
      - "--cookie-expire=12h"        # Session expires after 12 hours
      - "--session-store-type=cookie" # Store session in cookie (stateless)
```

For better security with Redis:
```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--session-store-type=redis"
      - "--redis-connection-url=redis://redis:6379"
      - "--cookie-expire=12h"
```

---

### 5. **🟡 HIGH: Email Domain Wildcard by Default**

**Issue:** By default, `emailDomains` is empty, allowing **ANY** authenticated user from the OAuth provider.

**File:** `helm/invulnerable/templates/oauth2-proxy-deployment.yaml:72`
```yaml
{{- else }}
- --email-domain=*  # ⚠️ Allows any email
{{- end }}
```

**Risk:** Anyone with a valid account at your OAuth provider can access the application.

**Recommendation:**
1. **Always** configure email domain restrictions in production:
```yaml
oauth2Proxy:
  config:
    emailDomains:
      - "yourcompany.com"  # REQUIRED
```

2. Update documentation to emphasize this is **mandatory** for production
3. Consider failing deployment if emailDomains is empty and environment is production

---

### 6. **🟡 MEDIUM: Assets Bypass Authentication**

**Issue:** Static assets (`/assets`, `/favicon.ico`) bypass OAuth2 authentication.

**File:** `helm/invulnerable/templates/ingress.yaml:135-148`

**Security Consideration:** This is **necessary** for the application to function (CSS/JS must load before authentication), but should be documented as a security consideration.

**Recommendation:**
- ✅ **Current implementation is acceptable** - static assets need to be public
- Document this in security documentation
- Ensure no sensitive data is served from `/assets`
- Consider Content-Security-Policy headers to mitigate XSS

---

### 7. **🟡 MEDIUM: No CSRF State Parameter Validation**

**Issue:** OAuth2 Proxy's state parameter validation is not explicitly configured.

**Recommendation:** Ensure state validation is enabled (it's enabled by default in v7.13.0, but make it explicit):

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--skip-auth-preflight=false"  # Enforce state validation
```

---

### 8. **🟢 LOW: No Rate Limiting on OAuth Endpoints**

**Issue:** No rate limiting on `/oauth2/start` and `/oauth2/callback` endpoints.

**Risk:** Potential for OAuth flow abuse or DoS on authentication endpoints.

**Recommendation:** Add nginx rate limiting:

```yaml
# In ingress annotations for oauth2-proxy ingress
nginx.ingress.kubernetes.io/limit-rps: "10"
nginx.ingress.kubernetes.io/limit-rpm: "100"
```

---

### 9. **🟢 LOW: Missing Security Headers**

**Recommendation:** Add security headers to ingress:

```yaml
ingress:
  annotations:
    nginx.ingress.kubernetes.io/configuration-snippet: |
      add_header X-Frame-Options "DENY" always;
      add_header X-Content-Type-Options "nosniff" always;
      add_header X-XSS-Protection "1; mode=block" always;
      add_header Referrer-Policy "strict-origin-when-cross-origin" always;
      add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
```

---

## 🔒 Production Deployment Checklist

Before deploying to production with OAuth2, ensure:

- [ ] **CRITICAL:** Network policies are enabled and tested
- [ ] **CRITICAL:** Email domain restrictions are configured
- [ ] **CRITICAL:** Backend validates authentication headers
- [ ] Cookie secret is exactly 32 bytes (validated)
- [ ] Session timeout is configured (12 hours recommended)
- [ ] TLS/HTTPS is properly configured with valid certificates
- [ ] Redirect URL in OAuth provider matches Helm configuration
- [ ] Secrets are managed via external secret management (Vault, Sealed Secrets, etc.)
- [ ] OAuth provider client credentials are rotated regularly
- [ ] Monitoring and alerting for authentication failures is configured
- [ ] Rate limiting is enabled on OAuth endpoints
- [ ] Security headers are configured

---

## 🧪 Testing Recommendations

### 1. Test Header Forgery Protection
```bash
# Try to forge auth headers directly to backend service
kubectl exec -it test-pod -- curl -H "X-Auth-Request-Email: attacker@evil.com" \
  http://invulnerable-backend:8080/api/v1/user/me

# Should return 401 Unauthorized (after implementing validation)
```

### 2. Test Network Policy Enforcement
```bash
# Try to access backend directly from another pod
kubectl run test-pod --rm -it --image=curlimages/curl -- \
  curl http://invulnerable-backend:8080/api/v1/metrics

# Should fail (connection refused) if network policies are working
```

### 3. Test OAuth Flow
```bash
# Access application URL
curl -I https://invulnerable.example.com/

# Should redirect to OAuth provider
# After authentication, should redirect back with session cookie
```

### 4. Test Session Expiration
```bash
# Set cookie expiration to 1 minute for testing
# Wait for expiration
# Attempt to access application
# Should redirect to OAuth login
```

---

## 📚 Additional Security Resources

- [OAuth2 Proxy Security Docs](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview#security)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [nginx Ingress Auth](https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/)

---

## Conclusion

The OAuth implementation provides a **solid foundation** for authentication but requires **critical hardening** before production use:

1. **Must implement:** Network policies and backend header validation
2. **Must configure:** Email domain restrictions and session timeouts
3. **Should add:** Security headers and rate limiting

After implementing these recommendations, the security posture will be **production-ready**.
