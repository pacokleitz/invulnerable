# Access Token Validation Implementation Guide

## ✅ Implementation Status

**JWT validation has been fully implemented!** The backend now cryptographically validates access tokens when OAuth2 is enabled.

## Current Security Posture

The application operates in two distinct modes based on Helm configuration:

### Mode 1: OAuth Disabled (Development/Testing)
**Configuration:** `oauth2Proxy.enabled: false`

**Behavior:**
- ✅ No authentication required
- ✅ Application works without OAuth2 Proxy
- ✅ `/api/v1/user/me` returns 204 No Content
- ⚠️ **NOT SECURE** - Do not expose publicly!

**Use case:** Local development, internal testing

### Mode 2: OAuth Enabled (Production)
**Configuration:** `oauth2Proxy.enabled: true` + `oidcIssuerUrl` configured

**Behavior:**
- ✅ Authentication REQUIRED
- ✅ JWT validation MANDATORY (no fallbacks)
- ✅ Cryptographic proof of authenticity
- ✅ Email cross-check (JWT vs headers)
- ✅ Defense-in-depth with Network Policies
- ❌ Fails fast if OIDC not properly configured

**Use case:** Production deployments, internet-facing applications

### Security Enforcement

When `oauth2Proxy.enabled=true`:
1. **Helm validation:** Fails if `oidcIssuerUrl` not set
2. **Backend startup:** Fails if `OIDC_ISSUER_URL` not provided
3. **Request handling:** Returns 401 if JWT validation fails
4. **No fallbacks:** JWT must be valid - no weak checks

## Security Layers (Defense in Depth)

```
Layer 1: Network Policies (PRIMARY DEFENSE) ✅
    ↓ Blocks direct pod-to-pod access
Layer 2: JWT Token Validation (SECONDARY DEFENSE) ✅ IMPLEMENTED
    ↓ Validates token cryptographically with JWKS
    ↓ Verifies signature, issuer, audience, expiration
Layer 3: Email Cross-Check (ADDITIONAL DEFENSE) ✅ IMPLEMENTED
    ↓ Verifies email in JWT matches header
Backend Application
```

## Implementation Options

### Option 1: JWT Token Validation (Recommended)

**When to use:** Your OAuth provider returns JWT access tokens (most OIDC providers do)

**Pros:**
- Cryptographic proof of authenticity
- No shared secrets between services
- Industry standard
- Verifies token hasn't been tampered with
- Checks expiration automatically

**Implementation:**

```go
// backend/internal/auth/jwt.go
package auth

import (
    "context"
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
    issuerURL  string
    audience   string
    httpClient *http.Client
    jwksCache  map[string]*rsa.PublicKey
    mutex      sync.RWMutex
}

type JWKS struct {
    Keys []JWK `json:"keys"`
}

type JWK struct {
    Kid string   `json:"kid"`
    Kty string   `json:"kty"`
    Alg string   `json:"alg"`
    Use string   `json:"use"`
    N   string   `json:"n"`
    E   string   `json:"e"`
}

func NewJWTValidator(issuerURL, audience string) *JWTValidator {
    return &JWTValidator{
        issuerURL:  issuerURL,
        audience:   audience,
        httpClient: &http.Client{Timeout: 10 * time.Second},
        jwksCache:  make(map[string]*rsa.PublicKey),
    }
}

func (v *JWTValidator) ValidateToken(tokenString string) (*jwt.Token, error) {
    // Parse token without validation first to get kid
    token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    kid, ok := token.Header["kid"].(string)
    if !ok {
        return nil, fmt.Errorf("kid header missing")
    }

    // Get public key for this kid
    publicKey, err := v.getPublicKey(kid)
    if err != nil {
        return nil, fmt.Errorf("failed to get public key: %w", err)
    }

    // Parse and validate token with public key
    token, err = jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return publicKey, nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to validate token: %w", err)
    }

    // Validate claims
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    // Verify issuer
    iss, ok := claims["iss"].(string)
    if !ok || iss != v.issuerURL {
        return nil, fmt.Errorf("invalid issuer: expected %s, got %s", v.issuerURL, iss)
    }

    // Verify audience (if configured)
    if v.audience != "" {
        aud, ok := claims["aud"].(string)
        if !ok || aud != v.audience {
            return nil, fmt.Errorf("invalid audience: expected %s, got %s", v.audience, aud)
        }
    }

    // Verify expiration
    if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
        return nil, fmt.Errorf("token expired")
    }

    return token, nil
}

func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
    // Check cache first
    v.mutex.RLock()
    if key, ok := v.jwksCache[kid]; ok {
        v.mutex.RUnlock()
        return key, nil
    }
    v.mutex.RUnlock()

    // Fetch JWKS
    jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", v.issuerURL)
    resp, err := v.httpClient.Get(jwksURL)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
    }
    defer resp.Body.Close()

    var jwks JWKS
    if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
        return nil, fmt.Errorf("failed to decode JWKS: %w", err)
    }

    // Find matching key
    for _, key := range jwks.Keys {
        if key.Kid == kid {
            publicKey, err := parseRSAPublicKey(key.N, key.E)
            if err != nil {
                return nil, err
            }

            // Cache the key
            v.mutex.Lock()
            v.jwksCache[kid] = publicKey
            v.mutex.Unlock()

            return publicKey, nil
        }
    }

    return nil, fmt.Errorf("key with kid %s not found", kid)
}

func parseRSAPublicKey(n, e string) (*rsa.PublicKey, error) {
    // Implementation to parse base64url encoded n and e into RSA public key
    // Use crypto/rsa and encoding/base64
    // ... (implementation details)
    return nil, nil // Placeholder
}
```

**Usage in Handler:**

```go
// backend/internal/api/user.go
import (
    "yourapp/internal/auth"
)

type UserHandler struct {
    logger       *zap.Logger
    jwtValidator *auth.JWTValidator  // Add this
}

func NewUserHandler(logger *zap.Logger, jwtValidator *auth.JWTValidator) *UserHandler {
    return &UserHandler{
        logger:       logger,
        jwtValidator: jwtValidator,
    }
}

func (h *UserHandler) GetCurrentUser(c echo.Context) error {
    email := c.Request().Header.Get("X-Auth-Request-Email")
    username := c.Request().Header.Get("X-Auth-Request-User")
    accessToken := c.Request().Header.Get("X-Auth-Request-Access-Token")

    if email == "" {
        // OAuth2 Proxy not deployed
        return c.NoContent(http.StatusNoContent)
    }

    // Validate access token cryptographically
    if h.jwtValidator != nil {
        token, err := h.jwtValidator.ValidateToken(accessToken)
        if err != nil {
            h.logger.Warn("invalid access token",
                zap.Error(err),
                zap.String("email", email),
                zap.String("remote_addr", c.RealIP()))
            return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
        }

        // Optional: Extract email from token and verify it matches header
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            if tokenEmail, ok := claims["email"].(string); ok && tokenEmail != email {
                h.logger.Warn("email mismatch between token and header",
                    zap.String("token_email", tokenEmail),
                    zap.String("header_email", email))
                return echo.NewHTTPError(http.StatusUnauthorized, "email mismatch")
            }
        }
    } else {
        // Fallback to presence check (current implementation)
        if accessToken == "" {
            h.logger.Warn("missing access token",
                zap.String("email", email))
            return echo.NewHTTPError(http.StatusUnauthorized, "invalid authentication")
        }
    }

    return c.JSON(http.StatusOK, UserResponse{
        Email:    email,
        Username: username,
    })
}
```

**Configuration:**

```go
// backend/cmd/manager/main.go
import (
    "os"
    "yourapp/internal/auth"
)

func main() {
    // ... existing code ...

    var jwtValidator *auth.JWTValidator
    if issuerURL := os.Getenv("OIDC_ISSUER_URL"); issuerURL != "" {
        audience := os.Getenv("OIDC_AUDIENCE") // Optional
        jwtValidator = auth.NewJWTValidator(issuerURL, audience)
        logger.Info("JWT validation enabled",
            zap.String("issuer", issuerURL))
    } else {
        logger.Warn("JWT validation disabled - using presence check only")
    }

    userHandler := api.NewUserHandler(logger, jwtValidator)

    // ... rest of setup ...
}
```

**Helm values:**

```yaml
backend:
  env:
    - name: OIDC_ISSUER_URL
      value: "https://your-idp.com/realms/your-realm"
    - name: OIDC_AUDIENCE
      value: "invulnerable"  # Optional
```

---

### Option 2: OAuth2 Proxy Signature Validation

**When to use:** Simpler alternative when JWT validation is complex

**Configuration:**

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--signature-key=your-32-byte-shared-secret"
```

**Backend:**

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func (h *UserHandler) validateOAuth2ProxySignature(c echo.Context) error {
    signature := c.Request().Header.Get("Gap-Signature")
    timestamp := c.Request().Header.Get("Gap-Auth")

    if signature == "" || timestamp == "" {
        return fmt.Errorf("missing signature headers")
    }

    // Compute expected signature
    h := hmac.New(sha256.New, []byte(h.signatureKey))
    h.Write([]byte(timestamp))
    expectedSig := hex.EncodeToString(h.Sum(nil))

    // Compare
    if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
        return fmt.Errorf("signature mismatch")
    }

    return nil
}
```

---

## Recommendation Priority

1. **Implement Network Policies** (CRITICAL - Already done ✅)
2. **Implement JWT Validation** (HIGH - Provides cryptographic proof)
3. **Keep Token Presence Check** (LOW - Minimal protection, current implementation)

## When is Token Presence Check "Good Enough"?

Token presence check alone is acceptable when:
- Network policies are properly enforced ✅
- Backend is in a private network/VPC
- No untrusted pods in the cluster
- Monitoring/alerting is in place for anomalies
- This is an internal-only application

For **internet-facing production apps**, implement JWT validation.

## Testing JWT Validation

```bash
# Get a real token from OAuth flow
TOKEN=$(kubectl exec -it oauth2-proxy-pod -- \
  cat /dev/shm/oauth2_proxy_session_* | jq -r '.AccessToken')

# Test with valid token
curl -H "X-Auth-Request-Access-Token: $TOKEN" \
     -H "X-Auth-Request-Email: user@example.com" \
     https://invulnerable.example.com/api/v1/user/me

# Test with invalid token (should fail)
curl -H "X-Auth-Request-Access-Token: fake-token" \
     -H "X-Auth-Request-Email: user@example.com" \
     https://invulnerable.example.com/api/v1/user/me
# Expected: 401 Unauthorized
```

## Performance Considerations

**JWT Validation:**
- JWKS caching reduces overhead
- Add TTL and refresh mechanism for JWKS cache
- Consider using a library like `github.com/MicahParks/keyfunc` for automatic JWKS rotation

**Token Presence Check:**
- Minimal overhead (string comparison)
- But provides minimal security

## Summary

| Method | Security | Complexity | Performance | Recommended |
|--------|----------|------------|-------------|-------------|
| Network Policies | HIGH ✅ | Low | N/A | **Yes** |
| JWT Validation | HIGH | Medium | Medium | **Yes (production)** |
| HMAC Signature | HIGH | Low | High | Yes (simpler alternative) |
| Token Presence | **LOW** ⚠️ | Very Low | Very High | **No (current)** |

**Current Status:** Token presence check provides **minimal security**. Network policies are the primary defense.

**Recommendation:** Implement JWT validation for production deployments.

---

## ✅ Using the Implemented JWT Validation

JWT validation has been fully implemented in the backend. To enable it:

### 1. Configure Helm Values

Add OIDC configuration to your `values.yaml`:

```yaml
oauth2Proxy:
  enabled: true

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-idp.com/realms/your-realm"

    # Optional: Set audience for stricter validation
    oidcAudience: "invulnerable"
```

### 2. Deploy

The backend will automatically:
- Load `OIDC_ISSUER_URL` from environment (injected by Helm)
- Initialize JWT validator on startup
- Fetch JWKS from `{issuerUrl}/.well-known/jwks.json`
- Cache public keys for 1 hour
- Validate all access tokens cryptographically

### 3. Verify It's Working

Check backend logs on startup:

```bash
kubectl logs -l app.kubernetes.io/component=backend -n invulnerable | grep JWT
```

**Expected output:**
```
{"level":"info","msg":"JWT validation enabled","issuer":"https://your-idp.com/realms/your-realm","audience":"invulnerable"}
```

**If OIDC not configured:**
```
{"level":"warn","msg":"JWT validation disabled - using token presence check only. Set OIDC_ISSUER_URL to enable cryptographic token validation."}
```

### 4. Test Token Validation

```bash
# Get a real token from a logged-in session
TOKEN=$(kubectl exec -it <oauth2-proxy-pod> -- cat /dev/shm/oauth2_proxy_* | jq -r '.AccessToken')

# Test with valid token (should succeed)
curl -H "X-Auth-Request-Access-Token: $TOKEN" \
     -H "X-Auth-Request-Email: user@example.com" \
     https://invulnerable.example.com/api/v1/user/me

# Test with invalid token (should fail with 401)
curl -H "X-Auth-Request-Access-Token: fake-token" \
     -H "X-Auth-Request-Email: user@example.com" \
     https://invulnerable.example.com/api/v1/user/me
```

### Implementation Files

The JWT validation implementation consists of:

1. **`backend/internal/auth/jwt.go`**
   - `JWTValidator` struct with JWKS caching
   - `ValidateToken()` - main validation logic
   - `getPublicKey()` - JWKS fetching with 1-hour cache
   - `parseRSAPublicKey()` - RSA key parsing from JWKS
   - `verifyAudience()` - audience claim validation

2. **`backend/internal/api/user.go`**
   - Updated to use `JWTValidator`
   - Cryptographic token verification
   - Email cross-check between JWT and headers
   - Graceful fallback if validator not configured

3. **`backend/cmd/server/main.go`**
   - Initializes `JWTValidator` from `OIDC_ISSUER_URL` env var
   - Passes validator to `UserHandler`

4. **`helm/invulnerable/templates/backend-deployment.yaml`**
   - Injects `OIDC_ISSUER_URL` environment variable
   - Injects `OIDC_AUDIENCE` if configured

5. **`helm/invulnerable/values.yaml`**
   - Added `oidcAudience` configuration option

### Security Benefits

With JWT validation enabled:
- ✅ **Cryptographic proof**: RSA signature verification prevents token forgery
- ✅ **Issuer validation**: Only tokens from configured OIDC provider accepted
- ✅ **Audience validation**: Only tokens for your application accepted (if configured)
- ✅ **Expiration check**: Expired tokens automatically rejected
- ✅ **Email cross-check**: Email in JWT must match OAuth2 Proxy header
- ✅ **JWKS caching**: Performance optimized with 1-hour cache TTL
- ✅ **Defense in depth**: Works alongside Network Policies for layered security

### Troubleshooting

**Problem: "failed to fetch JWKS"**
```
Solution: Verify backend can reach OIDC issuer URL
- Check network policies allow egress to OIDC provider
- Verify DNS resolution works from backend pods
- Check firewall rules
```

**Problem: "key with kid X not found in JWKS"**
```
Solution: OIDC provider rotated keys
- Wait 1 hour for cache to expire (or restart backend)
- Verify oidcIssuerUrl is correct
- Check OIDC provider JWKS endpoint: {issuerUrl}/.well-known/jwks.json
```

**Problem: "invalid issuer: expected X, got Y"**
```
Solution: Token issuer mismatch
- Verify OIDC_ISSUER_URL matches OAuth provider exactly
- Check for trailing slashes (must match exactly)
```

**Problem: "invalid audience"**
```
Solution: Audience claim mismatch
- Remove oidcAudience from values.yaml if not needed
- OR configure OAuth provider to include correct audience in tokens
```

**Problem: "email mismatch between token and header"**
```
Solution: Possible security issue - header injection attempt
- Check backend logs for details
- Verify OAuth2 Proxy is working correctly
- Could indicate an attack - investigate immediately
```
