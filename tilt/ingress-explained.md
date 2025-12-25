# How Ingress Works in kind

This document explains how the nginx Ingress Controller works in the kind (Kubernetes in Docker) setup.

## 🌐 Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│ Your Computer (macOS/Linux/Windows)                             │
│                                                                  │
│  Browser → invulnerable.local                                   │
│           ↓ (/etc/hosts: 127.0.0.1)                            │
│        localhost:80                                              │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ Docker Container (kind-invulnerable-control-plane)              │
│                                                                  │
│  Port 80 (mapped via extraPortMappings)                         │
│     ↓                                                            │
│  ┌──────────────────────────────────────────┐                  │
│  │ nginx-ingress-controller Pod             │                  │
│  │ (bound to host port 80 via hostPort)     │                  │
│  │                                           │                  │
│  │  ┌─────────────────────────────────────┐ │                  │
│  │  │ Nginx Reverse Proxy                 │ │                  │
│  │  │                                     │ │                  │
│  │  │ Routes based on Ingress rules:     │ │                  │
│  │  │ - /api → backend service:8080      │ │                  │
│  │  │ - /    → frontend service:80       │ │                  │
│  │  └─────────────────────────────────────┘ │                  │
│  └──────────────────────────────────────────┘                  │
│                   │                                              │
│                   ├──────────────────┬─────────────────┐        │
│                   ▼                  ▼                 ▼        │
│         ┌─────────────────┐  ┌──────────────┐  ┌─────────────┐ │
│         │ backend-service │  │ frontend-svc │  │ oauth2-svc  │ │
│         │   ClusterIP     │  │  ClusterIP   │  │ ClusterIP   │ │
│         └────────┬────────┘  └──────┬───────┘  └──────┬──────┘ │
│                  ▼                  ▼                  ▼        │
│         ┌─────────────────┐  ┌──────────────┐  ┌─────────────┐ │
│         │ backend pods    │  │ frontend pods│  │ oauth2 pods │ │
│         └─────────────────┘  └──────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## 🔑 Key Components

### 1. kind extraPortMappings (kind-config.yaml)

```yaml
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80    # Port inside kind container
        hostPort: 80         # Port on your computer
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
```

**What it does**: Maps your computer's ports 80/443 to the kind Docker container's ports 80/443.

**Equivalent**: Similar to `docker run -p 80:80 -p 443:443`

### 2. nginx Ingress Controller with hostPort

```python
# Tiltfile
helm_resource(
    name='ingress-nginx',
    flags=[
        '--set=controller.hostPort.enabled=true',
    ],
)
```

**What it does**: Configures the nginx ingress controller pod to bind directly to the node's ports 80 and 443.

**Pod spec created**:
```yaml
spec:
  containers:
  - name: controller
    ports:
    - containerPort: 80
      hostPort: 80      # ← Binds to node port
    - containerPort: 443
      hostPort: 443     # ← Binds to node port
```

### 3. Ingress Resources

```yaml
# Created by Helm chart
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: invulnerable
spec:
  rules:
  - host: invulnerable.local
    http:
      paths:
      - path: /api
        backend:
          service:
            name: invulnerable-backend
            port:
              number: 8080
      - path: /
        backend:
          service:
            name: invulnerable-frontend
            port:
              number: 80
```

**What it does**: Defines routing rules for the ingress controller.

## 🔀 Complete Traffic Flow

### Step-by-Step Request Flow

1. **Browser Request**
   ```
   User types: http://invulnerable.local
   ```

2. **DNS Resolution**
   ```
   /etc/hosts: invulnerable.local → 127.0.0.1
   ```

3. **TCP Connection**
   ```
   Browser connects to 127.0.0.1:80
   ```

4. **Docker Port Mapping**
   ```
   Docker forwards: localhost:80 → kind-container:80
   (via extraPortMappings in kind-config.yaml)
   ```

5. **hostPort Binding**
   ```
   kind node port 80 → nginx-ingress-controller pod port 80
   (via hostPort in pod spec)
   ```

6. **Ingress Routing**
   ```
   nginx reads Ingress rules and routes to:
   - /api/* → backend-service:8080
   - /* → frontend-service:80
   ```

7. **Service Load Balancing**
   ```
   Service distributes to backend/frontend pods
   ```

## 🆚 Alternative Approaches (Not Used)

### NodePort Without hostPort

```yaml
# If we used ONLY NodePort (without hostPort):
service:
  type: NodePort

# Would create:
service/ingress-nginx-controller NodePort 10.96.x.x:80 -> 30080/TCP

# Access would require:
http://invulnerable.local:30080  # ← Need to specify port!
```

**Why we don't use this**: NodePorts use random high ports (30000-32767), requiring extra port mappings.

### LoadBalancer

```yaml
# If we used LoadBalancer:
service:
  type: LoadBalancer

# Would require:
# - MetalLB or cloud provider
# - More complex setup
```

**Why we don't use this**: Overkill for local development.

### Port Forward

```bash
# Manual port forward:
kubectl port-forward -n ingress-nginx service/ingress-nginx-controller 80:80
```

**Why we don't use this**: Requires manual command, doesn't persist.

## 💡 Why hostPort is Perfect for kind

| Requirement | hostPort Solution |
|-------------|-------------------|
| **Use standard ports** (80/443) | ✅ Direct binding to ports 80/443 |
| **No manual port forwards** | ✅ Automatic via pod binding |
| **Simple DNS** (invulnerable.local) | ✅ Works with standard HTTP port |
| **HTTPS support** | ✅ Port 443 mapped |
| **Multiple sites** | ✅ Ingress rules handle routing |
| **Realistic** | ✅ Similar to production ingress |

## 🧪 Verify It's Working

### 1. Check kind port mappings

```bash
docker ps --filter name=kind-invulnerable-control-plane
# Look for: 0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
```

### 2. Check ingress controller pod

```bash
kubectl get pods -n ingress-nginx
# Should see: ingress-nginx-controller-xxx Running

kubectl get pods -n ingress-nginx -o yaml | grep -A 5 hostPort
# Should see: hostPort: 80 and hostPort: 443
```

### 3. Check ingress resources

```bash
kubectl get ingress -n invulnerable
# Should see: invulnerable.local with backend/frontend rules
```

### 4. Test connectivity

```bash
# Should work (via ingress):
curl http://invulnerable.local

# Should also work (direct to service via port-forward):
kubectl port-forward -n invulnerable svc/invulnerable-frontend 8080:80
curl http://localhost:8080
```

## 🔧 Troubleshooting

### "Connection refused" on invulnerable.local

**Check /etc/hosts**:
```bash
grep invulnerable /etc/hosts
# Should see: 127.0.0.1 invulnerable.local
```

**Check port mappings**:
```bash
docker port kind-invulnerable-control-plane
# Should see: 80/tcp -> 0.0.0.0:80
```

**Check ingress controller**:
```bash
kubectl get pods -n ingress-nginx
# All pods should be Running
```

### Ingress routes not working

**Check ingress rules**:
```bash
kubectl describe ingress -n invulnerable
# Verify host and path rules
```

**Check backend services exist**:
```bash
kubectl get svc -n invulnerable
# Should see: invulnerable-backend and invulnerable-frontend
```

**Check ingress controller logs**:
```bash
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
```

## 📚 Further Reading

- [kind Ingress Documentation](https://kind.sigs.k8s.io/docs/user/ingress/)
- [Kubernetes Ingress Concept](https://kubernetes.io/docs/concepts/services-networking/ingress/)
- [nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [hostPort vs NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport)

## 🎓 Summary

The ingress setup uses **hostPort** binding, which:
1. ✅ Provides direct access to standard ports (80/443)
2. ✅ Works automatically (no manual port-forwards)
3. ✅ Supports clean URLs (no port numbers needed)
4. ✅ Routes multiple services via Ingress rules
5. ✅ Simulates production ingress behavior

The traffic path is:
```
Browser → localhost:80 → kind-node:80 → ingress-pod:80 → service → pods
```

All configured declaratively in kind-config.yaml and Tiltfile!
