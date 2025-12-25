# Authentication with OAuth2 Proxy

This document explains how to enable authentication for Invulnerable using OAuth2 Proxy.

## Overview

OAuth2 Proxy provides authentication using various OAuth 2.0 / OIDC providers. When enabled, all access to the Invulnerable frontend and backend API requires authentication.

**Current Version:** v7.13.0 (includes security fixes for CVE-2025-47912, CVE-2025-58183, CVE-2025-58186, and CVE-2025-64484)

**Supported Providers:**
- Google
- GitHub
- GitLab
- Azure AD / Microsoft
- Keycloak
- Auth0
- Okta
- Any OIDC-compliant provider

## Quick Start

### 1. Generate Cookie Secret

```bash
# Generate a random cookie secret
openssl rand -base64 32 | head -c 32
```

Save this value - you'll need it for configuration.

### 2. Configure OAuth Provider

Choose your provider and follow the setup instructions below:

- [Google](#google-oauth)
- [GitHub](#github-oauth)
- [Keycloak](#keycloak-oidc)
- [Azure AD](#azure-ad)
- [Generic OIDC](#generic-oidc-provider)

### 3. Enable OAuth2 Proxy in Helm

```yaml
# values.yaml
oauth2Proxy:
  enabled: true

  clientID: "your-client-id"
  clientSecret: "your-client-secret"
  cookieSecret: "your-generated-cookie-secret"

  config:
    provider: "oidc"  # or google, github, azure, etc.
    oidcIssuerUrl: "https://your-provider.com/oauth2/issuer"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # Optional: Restrict to specific email domains
    emailDomains:
      - "example.com"
```

### 4. Install or Upgrade

```bash
helm upgrade --install invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --values values.yaml
```

## Provider Configurations

### Google OAuth

**1. Create OAuth 2.0 credentials:**
- Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
- Create OAuth 2.0 Client ID
- Application type: Web application
- Authorized redirect URIs: `https://invulnerable.example.com/oauth2/callback`

**2. Configure Helm values:**

```yaml
oauth2Proxy:
  enabled: true
  clientID: "xxxxxxxxxxxx.apps.googleusercontent.com"
  clientSecret: "your-client-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "google"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # Optional: Restrict to your organization's domain
    emailDomains:
      - "example.com"
```

**3. Scopes (automatically included):**
- `openid`
- `email`
- `profile`

### GitHub OAuth

**1. Create OAuth App:**
- Go to [GitHub Settings > Developer settings > OAuth Apps](https://github.com/settings/developers)
- Click "New OAuth App"
- Homepage URL: `https://invulnerable.example.com`
- Authorization callback URL: `https://invulnerable.example.com/oauth2/callback`

**2. Configure Helm values:**

```yaml
oauth2Proxy:
  enabled: true
  clientID: "your-github-client-id"
  clientSecret: "your-github-client-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "github"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # Optional: Restrict to organization
    extraArgs:
      - "--github-org=your-org-name"
      # Or restrict to specific team
      # - "--github-team=your-org-name/team-name"
```

**3. For GitHub Enterprise:**

```yaml
oauth2Proxy:
  config:
    provider: "github"
    loginUrl: "https://github.company.com/login/oauth/authorize"
    redeemUrl: "https://github.company.com/login/oauth/access_token"
    validateUrl: "https://github.company.com/api/v3/user"
    extraArgs:
      - "--github-endpoint=https://github.company.com/api/v3"
```

### Keycloak (OIDC)

**1. Create Client in Keycloak:**
- Realm Settings > Clients > Create
- Client ID: `invulnerable`
- Client Protocol: `openid-connect`
- Access Type: `confidential`
- Valid Redirect URIs: `https://invulnerable.example.com/oauth2/callback`
- Save and copy the Client Secret from Credentials tab

**2. Configure Helm values:**

```yaml
oauth2Proxy:
  enabled: true
  clientID: "invulnerable"
  clientSecret: "your-keycloak-client-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://keycloak.example.com/realms/your-realm"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    extraArgs:
      - "--scope=openid email profile"
      # Optional: Skip OIDC discovery and set endpoints manually
      # - "--skip-oidc-discovery=false"
```

**3. Keycloak Configuration Tips:**
- Ensure `email` scope is included in the client
- Map user attributes (email, name) in Client Scopes
- For role-based access, use group/role mappers

### Azure AD

**1. Register Application in Azure:**
- Go to [Azure Portal > Azure Active Directory > App registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps)
- New registration
- Redirect URI: `https://invulnerable.example.com/oauth2/callback`
- Create Client Secret in Certificates & secrets

**2. Configure Helm values:**

```yaml
oauth2Proxy:
  enabled: true
  clientID: "your-azure-application-id"
  clientSecret: "your-azure-client-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "azure"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # Azure tenant ID
    extraArgs:
      - "--azure-tenant=your-tenant-id"
      # Or use common for multi-tenant
      # - "--azure-tenant=common"
```

**3. For Azure AD B2C:**

```yaml
oauth2Proxy:
  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-tenant.b2clogin.com/your-tenant.onmicrosoft.com/your-policy/v2.0/"
    extraArgs:
      - "--scope=openid profile email"
```

### GitLab

**1. Create OAuth Application:**
- GitLab > User Settings > Applications
- Redirect URI: `https://invulnerable.example.com/oauth2/callback`
- Scopes: `read_user`, `openid`, `email`, `profile`

**2. Configure Helm values:**

```yaml
oauth2Proxy:
  enabled: true
  clientID: "your-gitlab-application-id"
  clientSecret: "your-gitlab-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "gitlab"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    # For self-hosted GitLab
    loginUrl: "https://gitlab.example.com/oauth/authorize"
    redeemUrl: "https://gitlab.example.com/oauth/token"
    validateUrl: "https://gitlab.example.com/api/v4/user"

    extraArgs:
      - "--scope=read_user openid email profile"
```

### Generic OIDC Provider

For Auth0, Okta, or any OIDC-compliant provider:

```yaml
oauth2Proxy:
  enabled: true
  clientID: "your-client-id"
  clientSecret: "your-client-secret"
  cookieSecret: "generate-with-openssl-rand"

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://your-provider.com/.well-known/openid-configuration"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"

    extraArgs:
      - "--scope=openid email profile"
      - "--oidc-email-claim=email"
      - "--oidc-groups-claim=groups"
```

## Advanced Configuration

### Email Domain Restrictions

Restrict access to specific email domains:

```yaml
oauth2Proxy:
  config:
    emailDomains:
      - "example.com"
      - "company.com"
```

### Cookie Domain Configuration

For subdomain access (e.g., `*.example.com`):

```yaml
oauth2Proxy:
  config:
    cookieDomains:
      - ".example.com"  # Note the leading dot
```

### Skip Authentication for Specific Paths

Allow unauthenticated access to health endpoints:

```yaml
oauth2Proxy:
  config:
    extraArgs:
      - "--skip-auth-regex=^/api/health$"
      - "--skip-auth-regex=^/healthz$"
```

### Using Kubernetes Secrets

For production, store credentials in Kubernetes secrets:

```bash
# Create secret
kubectl create secret generic oauth2-proxy-secrets \
  --from-literal=client-id='your-client-id' \
  --from-literal=client-secret='your-client-secret' \
  --from-literal=cookie-secret='your-cookie-secret' \
  --namespace invulnerable

# Reference in values.yaml
oauth2Proxy:
  enabled: true
  existingSecret: "oauth2-proxy-secrets"

  config:
    provider: "oidc"
    oidcIssuerUrl: "https://provider.com"
    redirectUrl: "https://invulnerable.example.com/oauth2/callback"
```

### Custom Inline Configuration

For complex configurations, provide inline oauth2-proxy config (TOML/INI format).
This automatically creates a ConfigMap and overrides all command-line arguments:

```yaml
oauth2Proxy:
  enabled: true
  existingSecret: "oauth2-proxy-secrets"

  config:
    configFile: |
      provider = "oidc"
      oidc_issuer_url = "https://provider.com"
      redirect_url = "https://invulnerable.example.com/oauth2/callback"

      email_domains = [
        "example.com"
      ]

      upstreams = [
        "file:///dev/null"
      ]

      skip_auth_regex = [
        "^/api/health$"
      ]

      set_xauthrequest = true
      pass_access_token = true
```

## Testing Authentication

### 1. Verify OAuth2 Proxy is Running

```bash
kubectl get pods -l app.kubernetes.io/component=oauth2-proxy -n invulnerable
kubectl logs -f -l app.kubernetes.io/component=oauth2-proxy -n invulnerable
```

### 2. Test Authentication Flow

1. Navigate to `https://invulnerable.example.com`
2. You should be redirected to your OAuth provider
3. After authentication, you'll be redirected back to Invulnerable
4. Check that you can access both frontend and API

### 3. Verify Headers

The backend should receive these headers from authenticated requests:

- `X-Auth-Request-User` - Username
- `X-Auth-Request-Email` - User email
- `Authorization` - Bearer token (if pass-access-token is enabled)

## Troubleshooting

### "Invalid Redirect URI"

**Problem:** OAuth provider rejects the redirect.

**Solution:**
- Ensure `redirectUrl` in values.yaml matches OAuth app configuration exactly
- Check for trailing slashes
- Verify HTTPS is used (required by most providers)

### "Cookie Secret Error"

**Problem:** Cookie secret is invalid or missing.

**Solution:**
```bash
# Generate new cookie secret (must be 32 bytes)
openssl rand -base64 32 | head -c 32
```

### "403 Forbidden" After Login

**Problem:** User authenticated but access denied.

**Solution:**
- Check `emailDomains` configuration
- Verify user's email domain is allowed
- Check oauth2-proxy logs for specific error

```bash
kubectl logs -f -l app.kubernetes.io/component=oauth2-proxy -n invulnerable
```

### HTTPS/TLS Issues

**Problem:** Redirects fail or cookies not set.

**Solution:**
- OAuth2 requires HTTPS in production
- Set `cookieSecure: true` for HTTPS
- Set `cookieSecure: false` for HTTP (development only)
- Configure TLS in ingress:

```yaml
ingress:
  tls:
    - secretName: invulnerable-tls
      hosts:
        - invulnerable.example.com
```

### Provider-Specific Issues

**Google:**
- Ensure email scope is requested
- Check Google Cloud Console for API quotas

**GitHub:**
- Verify organization/team names are correct
- Check that OAuth app has access to organization data

**Keycloak:**
- Ensure client is confidential (not public)
- Verify realm name is correct in issuer URL
- Check that email scope is mapped

## Security Best Practices

1. **Always use HTTPS in production**
   ```yaml
   oauth2Proxy:
     config:
       cookieSecure: true
   ```

2. **Restrict email domains**
   ```yaml
   oauth2Proxy:
     config:
       emailDomains:
         - "yourcompany.com"
   ```

3. **Use Kubernetes secrets for credentials**
   - Never commit secrets to git
   - Use `existingSecret` instead of inline values

4. **Rotate cookie secrets regularly**
   ```bash
   # Generate new secret
   NEW_SECRET=$(openssl rand -base64 32 | head -c 32)

   # Update secret
   kubectl patch secret oauth2-proxy-secrets -n invulnerable \
     -p "{\"data\":{\"cookie-secret\":\"$(echo -n $NEW_SECRET | base64)\"}}"

   # Restart oauth2-proxy
   kubectl rollout restart deployment/invulnerable-oauth2-proxy -n invulnerable
   ```

5. **Enable audit logging**
   ```yaml
   oauth2Proxy:
     config:
       extraArgs:
         - "--request-logging=true"
         - "--auth-logging=true"
   ```

## Disabling Authentication

To temporarily disable authentication:

```yaml
oauth2Proxy:
  enabled: false
```

Then upgrade the Helm release:

```bash
helm upgrade invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --values values.yaml
```

## Further Reading

- [OAuth2 Proxy Documentation](https://oauth2-proxy.github.io/oauth2-proxy/)
- [Provider Configuration](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview)
- [nginx Ingress External Auth](https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/)
