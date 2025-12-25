# Security and RBAC Configuration

This document explains the security model and RBAC (Role-Based Access Control) configuration for the Invulnerable Controller.

## Security Principles

The controller follows the **Principle of Least Privilege**:

1. **Minimal Permissions**: Controller only gets permissions it actually needs
2. **Namespace Scoping**: By default, controller only watches its own namespace
3. **Read-Only CRDs**: Controller cannot create or delete ImageScan resources (users do that)
4. **No Cluster Admin**: Controller never needs cluster-admin permissions
5. **Non-Root Execution**: Controller runs as non-root user (UID 65532)

## RBAC Modes

### 1. Namespace-Scoped (Default - RECOMMENDED)

**Default configuration with least privilege:**

```yaml
controller:
  rbac:
    clusterWide: false  # Default
```

**Permissions granted:**
- **ImageScans** (in deployment namespace only):
  - `get`, `list`, `watch` - Read ImageScan resources
  - `update`, `patch` - Update status and finalizers
  - ❌ NO `create` or `delete` - Users manage ImageScans, not controller

- **CronJobs** (in deployment namespace only):
  - `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` - Full control over managed CronJobs

- **Events** (in deployment namespace only):
  - `create`, `patch` - Report events for status updates

**RBAC Resources Created:**
- `Role` (namespace-scoped)
- `RoleBinding` (namespace-scoped)
- `ServiceAccount`
- `Role` (leader election)
- `RoleBinding` (leader election)

**Use this mode when:**
- ✅ Single-tenant cluster
- ✅ Controller and ImageScans in same namespace
- ✅ Maximum security required
- ✅ Compliance with least privilege

**Security Benefits:**
- Controller cannot affect other namespaces
- Contained blast radius if compromised
- Suitable for multi-tenant clusters (one controller per namespace)
- Easier to audit and manage

### 2. Cluster-Wide Mode

**For multi-namespace deployments:**

```yaml
controller:
  rbac:
    clusterWide: true
```

**Permissions granted:**
- Same verbs as namespace-scoped, but **across ALL namespaces**

**RBAC Resources Created:**
- `ClusterRole` (cluster-scoped)
- `ClusterRoleBinding` (cluster-scoped)
- `ServiceAccount`
- `Role` (leader election)
- `RoleBinding` (leader election)

**Use this mode when:**
- You want to create ImageScans in multiple namespaces
- Single controller manages scans across the cluster
- Trusted environment

**Security Considerations:**
- Controller can read ImageScans in all namespaces
- Controller can create CronJobs in all namespaces
- Larger blast radius if compromised
- Requires cluster-admin to install

## Permission Breakdown

### What Controller CAN Do

| Resource | Verbs | Reason |
|----------|-------|--------|
| `imagescans` | `get`, `list`, `watch` | Read ImageScan resources to reconcile |
| `imagescans/status` | `get`, `update`, `patch` | Update status with reconciliation state |
| `imagescans/finalizers` | `update` | Add/remove finalizers for cleanup |
| `cronjobs` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` | Full lifecycle management of owned CronJobs |
| `events` | `create`, `patch` | Report reconciliation events |

### What Controller CANNOT Do

| Resource | Prohibited Verbs | Reason |
|----------|------------------|--------|
| `imagescans` | ❌ `create`, `delete` | Users create/delete ImageScans, not controller |
| `secrets` | ❌ ALL | Controller doesn't need secret access |
| `configmaps` | ❌ ALL | Controller doesn't use ConfigMaps |
| `pods` | ❌ ALL | CronJobs manage Pods, not controller |
| `deployments` | ❌ ALL | Controller doesn't manage Deployments |
| `nodes` | ❌ ALL | Controller doesn't need node access |

## Scanner Job Security

Scanner Jobs (created by CronJobs) run with **no RBAC permissions**:

```yaml
# Scanner Pods have NO ServiceAccount permissions
# They only:
# 1. Run syft/grype to scan images
# 2. POST results to backend API via HTTP
# 3. No Kubernetes API access needed
```

**Security features:**
- Run as non-root user (UID 1000)
- `allowPrivilegeEscalation: false`
- Drop all capabilities
- `runAsNonRoot: true`
- Read-only root filesystem (where possible)

## Service Account Security

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: invulnerable-controller
  # No special annotations or secrets
  # Uses TokenRequest API (short-lived tokens)
```

**Security features:**
- No long-lived service account tokens
- Uses projected volume tokens (auto-rotated)
- Scoped to minimum required permissions

## Leader Election

Leader election permissions are **namespace-scoped** even in cluster-wide mode:

```yaml
# Leader election Role (always namespace-scoped)
- resources: ["configmaps", "leases"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # Only in controller's deployment namespace
```

**Why namespace-scoped:**
- Leader election ConfigMap/Lease is in deployment namespace
- No need for cluster-wide leader election permissions
- Reduces attack surface

## Attack Surface Analysis

### If Controller is Compromised

**Namespace-scoped mode (clusterWide: false):**
- ❌ Cannot access other namespaces
- ❌ Cannot create/delete ImageScans (only reconcile existing)
- ❌ Cannot access secrets or configmaps
- ✅ Limited to creating CronJobs in deployment namespace
- ✅ Cannot escalate privileges

**Cluster-wide mode (clusterWide: true):**
- ⚠️ Can create CronJobs in any namespace
- ⚠️ Can read ImageScans across cluster
- ❌ Still cannot create/delete ImageScans
- ❌ Still cannot access secrets

### Mitigation Strategies

1. **Use namespace-scoped mode** (default)
2. **Network Policies** - Restrict controller egress
3. **Pod Security Standards** - Enforce restricted profile
4. **Image Scanning** - Scan controller image before deployment
5. **Audit Logging** - Enable Kubernetes audit logs
6. **Monitoring** - Alert on unexpected RBAC usage

## Compliance Mapping

### CIS Kubernetes Benchmark

| Control | Compliance | Implementation |
|---------|-----------|----------------|
| 5.1.5 - Minimize service account permissions | ✅ | Namespace-scoped RBAC by default |
| 5.1.6 - Verify service account token auto-mounting | ✅ | Uses projected tokens |
| 5.2.2 - Minimize privileged containers | ✅ | Non-root, no privileges |
| 5.2.3 - Minimize capabilities | ✅ | Drop all capabilities |
| 5.2.5 - Minimize root containers | ✅ | runAsNonRoot: true |

### NIST 800-190

| Control | Compliance |
|---------|-----------|
| Least Privilege | ✅ Minimal RBAC, namespace-scoped |
| Defense in Depth | ✅ Pod security + RBAC + network policies |
| Audit Logging | ✅ RBAC changes audited by K8s |

## Verification

### Check Controller Permissions

```bash
# View actual permissions
kubectl describe clusterrole invulnerable-controller-role
# OR (namespace-scoped)
kubectl describe role invulnerable-controller-role -n invulnerable

# Check what controller can do
kubectl auth can-i --as=system:serviceaccount:invulnerable:invulnerable-controller create imagescans
# Should return: no

kubectl auth can-i --as=system:serviceaccount:invulnerable:invulnerable-controller create cronjobs
# Should return: yes (in namespace mode: only in invulnerable namespace)
```

### Audit RBAC

```bash
# List all permissions for controller
kubectl get clusterrolebinding -o json | \
  jq '.items[] | select(.subjects[]?.name=="invulnerable-controller")'

# Check for overly permissive rules
kubectl get clusterrole invulnerable-controller-role -o yaml | \
  grep -E "verbs.*\*|resources.*\*"
# Should return nothing (no wildcards)
```

## Security Updates

When updating RBAC:

1. **Never add `cluster-admin`** - Controller never needs it
2. **Avoid wildcards** - No `resources: ["*"]` or `verbs: ["*"]`
3. **Document changes** - Update this file with rationale
4. **Test namespace-scoped first** - Before enabling cluster-wide
5. **Use specific API groups** - Don't use `apiGroups: ["*"]`

## Reporting Security Issues

If you discover a security issue with the controller RBAC:

1. **Do NOT open a public issue**
2. Email: security@example.com (replace with your security contact)
3. Include: RBAC rules, impact, reproduction steps
4. We will respond within 48 hours

## References

- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Principle of Least Privilege](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [NIST 800-190](https://csrc.nist.gov/publications/detail/sp/800-190/final)
