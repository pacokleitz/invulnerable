# Network Policy Architecture

## Overview

When network policies are enabled, the backend is protected by restricting which pods can access it. This prevents attackers from bypassing OAuth2 authentication by directly accessing the backend from compromised pods.

## Traffic Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                         Kubernetes Cluster                            │
│                                                                        │
│  ┌─────────────────┐                                                  │
│  │ Ingress         │  ✅ ALLOWED                                      │
│  │ Controller      │────────────────┐                                 │
│  │ (nginx)         │                │                                 │
│  └─────────────────┘                │                                 │
│         │                            │                                 │
│         │ Enforces OAuth2            ▼                                 │
│         │ via auth-url          ┌──────────────┐                      │
│         │                        │   Backend    │                      │
│         └───────────────────────▶│   :8080      │                      │
│                                  │              │                      │
│  ┌─────────────────┐             │ Network      │                      │
│  │ Scanner Pod     │  ✅ ALLOWED │ Policy       │                      │
│  │ (Job)           │────────────▶│ Protected    │                      │
│  │ component:      │             │              │                      │
│  │   scanner       │             └──────────────┘                      │
│  └─────────────────┘                    ▲                              │
│                                          │                              │
│  ┌─────────────────┐                    │                              │
│  │ Prometheus      │  ✅ OPTIONAL       │                              │
│  │ (if enabled)    │────────────────────┘                              │
│  └─────────────────┘                                                   │
│                                                                        │
│  ┌─────────────────┐                                                  │
│  │ Random Pod      │  ❌ BLOCKED                                      │
│  │ (Attacker)      │─────────X──────────────────────────────────      │
│  └─────────────────┘         Cannot reach backend                     │
│                                                                        │
└──────────────────────────────────────────────────────────────────────┘
```

## Allowed Traffic Sources

The network policy allows traffic to the backend from **3 sources only**:

### 1. Ingress Controller (Required)
**Purpose:** User traffic through authenticated ingress

**Selector:**
```yaml
- from:
  - namespaceSelector:
      matchLabels:
        kubernetes.io/metadata.name: ingress-nginx
  ports:
  - protocol: TCP
    port: 8080
```

**Why:** Ingress controller enforces OAuth2 authentication via `auth-url` annotation. All user requests must go through this path.

**Traffic:**
- ✅ GET /api/v1/scans (user viewing scans)
- ✅ GET /api/v1/vulnerabilities (user viewing vulnerabilities)
- ✅ PATCH /api/v1/vulnerabilities/123 (user updating vulnerability status)
- ✅ All authenticated user traffic

### 2. Scanner Pods (Required)
**Purpose:** Scanner Jobs posting scan results

**Selector:**
```yaml
- from:
  - podSelector:
      matchLabels:
        app.kubernetes.io/name: invulnerable-scanner
  ports:
  - protocol: TCP
    port: 8080
```

**Why:** Scanner pods are created as Jobs by the ImageScan controller. They need to POST scan results to the backend after completing vulnerability scans. Using the specific `app.kubernetes.io/name` label ensures only invulnerable scanner pods are allowed (not any random scanner).

**Traffic:**
- ✅ POST /api/v1/scans (submitting scan results)

**Labels on scanner pods:**
```bash
kubectl get pods -n invulnerable -l app.kubernetes.io/name=invulnerable-scanner
```

### 3. Monitoring (Optional)
**Purpose:** Prometheus metrics scraping

**Selector:**
```yaml
- from:
  - namespaceSelector:
      matchLabels:
        name: monitoring
    podSelector:
      matchLabels:
        app: prometheus
  ports:
  - protocol: TCP
    port: 8080
```

**Why:** If you have Prometheus deployed, it needs to scrape `/metrics` endpoint.

**Traffic:**
- ✅ GET /api/v1/metrics (Prometheus scraping)

**Configuration:**
```yaml
networkPolicy:
  enabled: true
  allowMonitoring: true
  monitoringNamespaceLabel:
    name: monitoring
  monitoringPodLabel:
    app: prometheus
```

## Blocked Traffic

**All other pods in the cluster are blocked:**

```bash
# This will timeout (connection refused)
kubectl run attacker-pod --rm -it --image=curlimages/curl -n invulnerable -- \
  curl --connect-timeout 5 http://invulnerable-backend:8080/api/v1/scans

# Output: curl: (28) Connection timeout after 5000 ms
```

**Why this is critical:**
- Prevents pod-to-pod bypass of OAuth2
- Prevents compromised pods from accessing backend
- Prevents lateral movement in cluster
- Enforces all traffic through authenticated ingress

## Security Impact

### Without Network Policies ❌

```
Attacker Pod ──────────────────┐
                                │
User Browser ───┐               ├──▶ Backend
                │               │    (No protection!)
Scanner Pod ────┼───────────────┘
                │
Ingress ────────┘ (OAuth2 bypassed!)
```

**Attack scenario:**
1. Attacker compromises any pod in cluster
2. Attacker makes direct HTTP request to backend service
3. Attacker forges `X-Auth-Request-Email` header
4. **Backend trusts header and returns data** 😱

### With Network Policies ✅

```
Attacker Pod ────────X (BLOCKED)

User Browser ───┐
                │
Scanner Pod ────┼──▶ Backend
                │    (Protected!)
Ingress ────────┘ (OAuth2 enforced)
```

**Attack scenario:**
1. Attacker compromises any pod in cluster
2. Attacker makes direct HTTP request to backend service
3. **Network policy blocks connection** ✅
4. Attacker cannot reach backend at all ✅

## Why Scanner Pods Are Trusted

You might wonder: **"If scanner pods can access the backend, can't an attacker label their pod as a scanner?"**

**Answer:** No, because:

1. **RBAC prevents label spoofing:**
   - Regular users cannot create pods with arbitrary labels
   - Only the ImageScan controller (with ServiceAccount) can create scanner Jobs
   - Controller is trusted and creates Jobs with proper labels

2. **Scanner authentication:**
   - Scanner pods still go through backend API validation
   - Scanner endpoints (`POST /api/v1/scans`) validate the data
   - Invalid scan data is rejected

3. **Scanner scope is limited:**
   - Scanners can only POST scan results
   - They cannot read other scans or modify vulnerabilities
   - Attack surface is minimal

4. **Network policy is defense-in-depth:**
   - Primary defense: RBAC + proper labels
   - Secondary defense: Backend API validation
   - Network policy prevents unauthorized pods from even trying

## Configuration

### Enable Network Policies (Production)

```yaml
# values.yaml
networkPolicy:
  enabled: true
  ingressControllerNamespaceLabel:
    kubernetes.io/metadata.name: ingress-nginx
```

### Customize Ingress Controller

For different ingress controllers, adjust the namespace selector:

```yaml
# Traefik
ingressControllerNamespaceLabel:
  kubernetes.io/metadata.name: traefik

# Istio
ingressControllerNamespaceLabel:
  kubernetes.io/metadata.name: istio-system
```

### Add Custom Allowed Pods

```yaml
networkPolicy:
  additionalIngressRules:
    - from:
      - podSelector:
          matchLabels:
            app: my-custom-app
      ports:
      - protocol: TCP
        port: 8080
```

## Verification

### Test Network Policy is Working

```bash
# 1. Regular pods should be blocked
kubectl run test-blocked --rm -it --image=curlimages/curl -n invulnerable -- \
  curl --connect-timeout 5 http://invulnerable-backend:8080/api/v1/metrics

# Expected: Connection timeout ✅

# 2. Check scanner pods can still POST results
kubectl logs -n invulnerable -l app.kubernetes.io/component=scanner --tail=20

# Expected: "scan created successfully" ✅

# 3. Check ingress traffic works (via browser or curl through ingress)
curl http://invulnerable.local:8080/api/v1/scans

# Expected: Redirect to OAuth login or scan results ✅
```

### View Network Policy

```bash
kubectl get networkpolicy -n invulnerable
kubectl describe networkpolicy invulnerable-backend -n invulnerable
```

## Troubleshooting

### Scanner Jobs Failing

**Symptom:**
```
Error: failed to create scan: Post "http://invulnerable-backend:8080/api/v1/scans": context deadline exceeded
```

**Cause:** Network policy blocking scanner pods

**Solution:** Verify scanner pods have the correct label:
```bash
kubectl get pods -n invulnerable -l app.kubernetes.io/name=invulnerable-scanner --show-labels
```

Expected label: `app.kubernetes.io/name=invulnerable-scanner`

If label is missing, check controller Job template.

### Ingress Traffic Blocked

**Symptom:** 502 Bad Gateway when accessing application

**Cause:** Network policy blocking ingress controller

**Solution:** Verify ingress controller namespace label:
```bash
kubectl get namespace ingress-nginx --show-labels
```

Update `ingressControllerNamespaceLabel` in values.yaml to match.

## Summary

| Source | Label/Namespace | Traffic | Authenticated |
|--------|-----------------|---------|---------------|
| **User Browser** | Via Ingress | All API calls | ✅ OAuth2 |
| **Scanner Pods** | `name=invulnerable-scanner` | POST scan results | ⚠️ Trusted system |
| **Prometheus** | `app=prometheus` | GET /metrics | ⚠️ Monitoring only |
| **Random Pods** | Any other | ❌ Blocked | N/A |

**Key Takeaway:** Network policies enforce that all user traffic goes through OAuth2-authenticated ingress, while still allowing trusted system components (scanners, monitoring) to function.
