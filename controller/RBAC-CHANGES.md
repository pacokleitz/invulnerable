# RBAC Security Improvements

## Summary

The controller now follows the **Principle of Least Privilege** with configurable RBAC scoping.

## What Changed

### Before (❌ Security Issues)

```yaml
# ClusterRole with excessive permissions
kind: ClusterRole
rules:
- apiGroups: [invulnerable.io]
  resources: [imagescans]
  verbs: [get, list, watch, create, update, patch, delete]  # ❌ Too many verbs
  # ❌ Cluster-wide access to ALL namespaces
  # ❌ Controller can create/delete ImageScans (should be user's job)
```

**Problems:**
1. ❌ **Cluster-wide access by default** - Controller could affect any namespace
2. ❌ **Excessive permissions** - Controller could `create` and `delete` ImageScans
3. ❌ **No least privilege** - More permissions than needed
4. ❌ **Security risk** - If compromised, controller could impact entire cluster
5. ❌ **Compliance issues** - Violates CIS Kubernetes Benchmark 5.1.5

### After (✅ Secure)

```yaml
# Default: Namespace-scoped Role (RECOMMENDED)
kind: Role  # ✅ Namespace-scoped, not cluster-wide
metadata:
  namespace: invulnerable  # ✅ Limited to deployment namespace
rules:
- apiGroups: [invulnerable.io]
  resources: [imagescans]
  verbs: [get, list, watch]  # ✅ Read-only (no create/delete)
- apiGroups: [invulnerable.io]
  resources: [imagescans/status]
  verbs: [get, update, patch]  # ✅ Can update status only
- apiGroups: [batch]
  resources: [cronjobs]
  verbs: [get, list, watch, create, update, patch, delete]  # ✅ Full control of owned resources
```

**Improvements:**
1. ✅ **Namespace-scoped by default** - Controller only affects deployment namespace
2. ✅ **Read-only ImageScans** - Controller cannot create/delete CRDs (users do that)
3. ✅ **Configurable scope** - Can enable cluster-wide mode if needed
4. ✅ **Least privilege** - Minimal permissions required
5. ✅ **Compliance** - Meets CIS Kubernetes Benchmark requirements

## Configuration

### Default (Namespace-Scoped - RECOMMENDED)

```yaml
# values.yaml
controller:
  rbac:
    clusterWide: false  # Default
```

**Creates:**
- `Role` (namespace-scoped)
- `RoleBinding` (namespace-scoped)

**Permissions:**
- ImageScans: Read-only in deployment namespace
- CronJobs: Full control in deployment namespace
- Events: Create/patch in deployment namespace

**Use when:**
- ✅ Single namespace deployment
- ✅ Maximum security required
- ✅ Multi-tenant clusters
- ✅ Compliance requirements

### Optional (Cluster-Wide)

```yaml
# values.yaml
controller:
  rbac:
    clusterWide: true
```

**Creates:**
- `ClusterRole` (cluster-scoped)
- `ClusterRoleBinding` (cluster-scoped)

**Permissions:**
- ImageScans: Read-only across ALL namespaces
- CronJobs: Full control across ALL namespaces
- Events: Create/patch across ALL namespaces

**Use when:**
- You need to create ImageScans in multiple namespaces
- Single controller for entire cluster
- Trusted environment

## Permission Changes

### ImageScans Resource

| Verb | Before | After | Reason |
|------|--------|-------|--------|
| `get` | ✅ | ✅ | Read ImageScan specs |
| `list` | ✅ | ✅ | List ImageScans for reconciliation |
| `watch` | ✅ | ✅ | Watch for ImageScan changes |
| `create` | ✅ ❌ | ❌ ✅ | **REMOVED** - Users create ImageScans, not controller |
| `update` | ✅ | ❌ | **REMOVED** - Only status updates needed |
| `patch` | ✅ | ❌ | **REMOVED** - Only status patches needed |
| `delete` | ✅ ❌ | ❌ ✅ | **REMOVED** - Users delete ImageScans, not controller |

### ImageScans/Status Subresource

| Verb | Before | After | Reason |
|------|--------|-------|--------|
| `get` | ✅ | ✅ | Read current status |
| `update` | ✅ | ✅ | Update status after reconciliation |
| `patch` | ✅ | ✅ | Patch status conditions |

### CronJobs Resource

| Verb | Before | After | Reason |
|------|--------|-------|--------|
| All | ✅ | ✅ | Full control needed for owned resources |

**Note:** Controller creates and owns CronJobs, so it needs full lifecycle permissions.

## Security Benefits

### Attack Surface Reduction

**If controller is compromised:**

| Attack | Before | After (namespace-scoped) |
|--------|--------|--------------------------|
| Access other namespaces | ✅ Possible | ❌ Blocked |
| Create malicious ImageScans | ✅ Possible | ❌ Blocked |
| Delete user ImageScans | ✅ Possible | ❌ Blocked |
| Create CronJobs in other namespaces | ✅ Possible | ❌ Blocked |
| Access secrets | ❌ Blocked | ❌ Blocked |
| Escalate privileges | ❌ Blocked | ❌ Blocked |

### Compliance

| Standard | Control | Before | After |
|----------|---------|--------|-------|
| CIS Kubernetes Benchmark | 5.1.5 - Minimize service account permissions | ❌ Failed | ✅ Pass |
| NIST 800-190 | Least Privilege | ❌ Failed | ✅ Pass |
| PCI-DSS | Requirement 7 - Restrict access | ⚠️ Partial | ✅ Pass |

## Migration Guide

### Existing Deployments

If you're upgrading from the old RBAC configuration:

```bash
# 1. Upgrade Helm chart (automatically updates RBAC)
helm upgrade invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --reuse-values

# 2. Verify new permissions
kubectl describe role invulnerable-controller-role -n invulnerable

# 3. Verify controller can still reconcile
kubectl get imagescans -n invulnerable
kubectl logs -f -l app.kubernetes.io/component=controller -n invulnerable

# 4. Test creating an ImageScan as a user (not controller)
kubectl apply -f - <<EOF
apiVersion: invulnerable.io/v1alpha1
kind: ImageScan
metadata:
  name: test-scan
  namespace: invulnerable
spec:
  image: "nginx:latest"
  schedule: "0 2 * * *"
EOF

# 5. Verify CronJob was created by controller
kubectl get cronjobs -n invulnerable | grep test-scan
```

### Enabling Cluster-Wide Mode

If you need cluster-wide mode:

```bash
helm upgrade invulnerable ./helm/invulnerable \
  --namespace invulnerable \
  --set controller.rbac.clusterWide=true \
  --reuse-values
```

**Warning:** This grants cluster-wide permissions. Only enable if necessary.

## Verification

### Check Current RBAC Mode

```bash
# Check if using Role (namespace-scoped)
kubectl get role invulnerable-controller-role -n invulnerable
# If found: namespace-scoped mode ✅

# Check if using ClusterRole (cluster-wide)
kubectl get clusterrole invulnerable-controller-role
# If found: cluster-wide mode ⚠️
```

### Verify Permissions

```bash
# Test: Can controller create ImageScans? (should be NO)
kubectl auth can-i create imagescans \
  --as=system:serviceaccount:invulnerable:invulnerable-controller \
  -n invulnerable
# Expected: no ✅

# Test: Can controller update ImageScan status? (should be YES)
kubectl auth can-i patch imagescans/status \
  --as=system:serviceaccount:invulnerable:invulnerable-controller \
  -n invulnerable
# Expected: yes ✅

# Test: Can controller create CronJobs? (should be YES)
kubectl auth can-i create cronjobs \
  --as=system:serviceaccount:invulnerable:invulnerable-controller \
  -n invulnerable
# Expected: yes ✅
```

## References

- [Kubernetes RBAC Best Practices](https://kubernetes.io/docs/concepts/security/rbac-good-practices/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [NIST SP 800-190](https://csrc.nist.gov/publications/detail/sp/800-190/final)
- [controller/SECURITY.md](./SECURITY.md) - Detailed security documentation
