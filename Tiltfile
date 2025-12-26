# Tiltfile for local development with Docker Desktop Kubernetes
# https://docs.tilt.dev/

# Allow k8s contexts for Docker Desktop
allow_k8s_contexts('docker-desktop')

# Suppress warning for scanner image (used by dynamically created Jobs)
update_settings(suppress_unused_image_warnings=['localhost:5001/invulnerable-scanner'])

# Load Tilt extensions
load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create')

# Configuration
config.define_string('registry', args=False, usage='Docker registry for images')
config.define_bool('enable-https', args=False, usage='Enable HTTPS with cert-manager')
config.define_bool('enable-oidc', args=False, usage='Enable OIDC with Dex provider')
cfg = config.parse()
registry = cfg.get('registry', 'localhost:5001')
enable_https = cfg.get('enable-https', False)
enable_oidc = cfg.get('enable-oidc', False)

# Helm repositories should be added via setup-tilt.sh
# If you haven't run setup yet, run: ./scripts/setup-tilt.sh

# Deploy nginx Ingress Controller
# For Docker Desktop: Uses LoadBalancer + port-forward to expose to localhost
# Traffic flow: localhost:80 -> port-forward -> LoadBalancer -> ingress pod:80 -> services
print('📦 Deploying nginx Ingress Controller...')
helm_resource(
    name='ingress-nginx',
    chart='ingress-nginx/ingress-nginx',
    namespace='ingress-nginx',
    flags=[
        '--create-namespace',
        '--set=controller.service.type=LoadBalancer',
        '--set=controller.watchIngressWithoutClass=true',
    ],
    labels=['infrastructure'],
    resource_deps=[],
)

# Port forward ingress to localhost (required for Docker Desktop on macOS)
# Using unprivileged ports 8080:80 and 8443:443 to avoid permission issues
local_resource(
    'ingress-localhost',
    serve_cmd='kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80 8443:443',
    resource_deps=['ingress-nginx'],
    labels=['infrastructure'],
)

# Optionally deploy cert-manager for HTTPS
if enable_https:
    print('🔒 Deploying cert-manager for HTTPS support...')
    helm_resource(
        name='cert-manager',
        chart='jetstack/cert-manager',
        namespace='cert-manager',
        flags=[
            '--create-namespace',
            '--set=installCRDs=true',
        ],
        labels=['infrastructure'],
        resource_deps=[],
    )

    # Create self-signed ClusterIssuer and Certificate
    k8s_yaml(blob('''
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: invulnerable-local-cert
  namespace: invulnerable
spec:
  secretName: invulnerable-tls
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
    - invulnerable.local
    - "*.invulnerable.local"
'''))

    k8s_resource(
        'selfsigned-issuer',
        labels=['infrastructure'],
        resource_deps=['cert-manager'],
    )

    k8s_resource(
        'invulnerable-local-cert',
        labels=['infrastructure'],
        resource_deps=['cert-manager', 'selfsigned-issuer'],
    )

# Optionally deploy Dex OIDC provider for local testing
if enable_oidc:
    print('🔐 Deploying Dex OIDC provider for local authentication...')
    helm_resource(
        name='dex',
        chart='dex/dex',
        namespace='dex',
        flags=[
            '--create-namespace',
            '--values=./tilt/dex-values.yaml',
        ],
        labels=['infrastructure', 'auth'],
        resource_deps=['ingress-nginx'],
    )

# Create namespace (depends on cert-manager if HTTPS enabled)
if enable_https:
    namespace_create('invulnerable', allow_duplicates=True)
else:
    namespace_create('invulnerable')

# Build Docker images with live updates
docker_build(
    ref=registry + '/invulnerable-backend',
    context='./backend',
    dockerfile='./backend/Dockerfile',
    live_update=[
        # Sync Go source files
        sync('./backend', '/app'),
        # Rebuild on Go file changes
        run('cd /app && go build -o /app/server ./cmd/server', trigger=['./backend/**/*.go']),
    ],
    # Only rebuild when these files change (relative to context)
    only=[
        'cmd/',
        'internal/',
        'migrations/',
        'go.mod',
        'go.sum',
        'Dockerfile',
    ],
)

docker_build(
    ref=registry + '/invulnerable-frontend',
    context='./frontend',
    dockerfile='./frontend/Dockerfile',
    live_update=[
        # Sync source files for Vite HMR
        sync('./frontend/src', '/app/src'),
        sync('./frontend/index.html', '/app/index.html'),
        sync('./frontend/package.json', '/app/package.json'),
    ],
    # Only rebuild when these files change (relative to context)
    only=[
        'src/',
        'index.html',
        'package.json',
        'package-lock.json',
        'tsconfig.json',
        'tsconfig.node.json',
        'vite.config.ts',
        'tailwind.config.js',
        'postcss.config.js',
        'nginx.conf',
        'Dockerfile',
    ],
)

docker_build(
    ref=registry + '/invulnerable-controller',
    context='./controller',
    dockerfile='./controller/Dockerfile',
    live_update=[
        sync('./controller', '/workspace'),
        run('cd /workspace && make build', trigger=['./controller/**/*.go']),
    ],
    # Only rebuild when these files change (relative to context)
    only=[
        'api/',
        'internal/',
        'cmd/',
        'go.mod',
        'go.sum',
        'Makefile',
    ],
)

docker_build(
    ref=registry + '/invulnerable-scanner',
    context='./scanner',
    dockerfile='./scanner/Dockerfile',
    # Only rebuild when these files change (relative to context)
    only=[
        'scan.sh',
        'Dockerfile',
    ],
)

# Deploy PostgreSQL
print('🐘 Deploying PostgreSQL...')
helm_resource(
    name='postgres',
    chart='bitnami/postgresql',
    namespace='invulnerable',
    flags=[
        '--values=./tilt/postgres-values.yaml',
    ],
    port_forwards=['5432:5432'],
    labels=['database'],
    resource_deps=['ingress-nginx'],  # Wait for ingress to be ready
)

# Deploy Invulnerable with Helm
print('🚀 Deploying Invulnerable application...')

# Choose values file based on mode
if enable_oidc:
    values_file = './tilt/values-oidc.yaml'
elif enable_https:
    values_file = './tilt/values-https.yaml'
else:
    values_file = './tilt/values.yaml'

# Set up dependencies
helm_deps = ['postgres']
if enable_https:
    helm_deps.append('cert-manager')
if enable_oidc:
    helm_deps.append('dex')

# Render Helm chart and apply as k8s_yaml
# This allows Tilt to scan for image references
yaml = local(
    'helm template invulnerable ./helm/invulnerable ' +
    '--namespace invulnerable ' +
    '--values=' + values_file + ' ' +
    '--set=image.registry=' + registry + ' ' +
    '--set=backend.image.repository=invulnerable-backend ' +
    '--set=backend.image.tag=latest ' +
    '--set=backend.image.pullPolicy=IfNotPresent ' +
    '--set=frontend.image.repository=invulnerable-frontend ' +
    '--set=frontend.image.tag=latest ' +
    '--set=frontend.image.pullPolicy=IfNotPresent ' +
    '--set=controller.image.repository=invulnerable-controller ' +
    '--set=controller.image.tag=latest ' +
    '--set=controller.image.pullPolicy=IfNotPresent ' +
    '--set=scanner.image.repository=invulnerable-scanner ' +
    '--set=scanner.image.tag=latest ' +
    '--set=scanner.image.pullPolicy=IfNotPresent',
    quiet=True  # Suppress verbose output during tilt down
)

k8s_yaml(yaml)

# Group resources under 'invulnerable' label
k8s_resource(
    objects=[
        'invulnerable:namespace',
        'invulnerable:serviceaccount',
    ],
    new_name='invulnerable-setup',
    labels=['app'],
    resource_deps=helm_deps,
)

k8s_resource(
    workload='invulnerable-backend',
    labels=['app'],
    resource_deps=helm_deps,
)

k8s_resource(
    workload='invulnerable-frontend',
    labels=['app'],
    resource_deps=helm_deps,
)

k8s_resource(
    workload='invulnerable-controller',
    labels=['app'],
    resource_deps=helm_deps,
)

k8s_resource(
    workload='invulnerable-oauth2-proxy',
    labels=['auth'],
    resource_deps=helm_deps,
)

# Port forwards for direct access (bypassing ingress)
local_resource(
    'port-forward-backend',
    serve_cmd='kubectl port-forward -n invulnerable svc/invulnerable-backend 8081:8080',
    resource_deps=['invulnerable-backend'],
    labels=['port-forwards'],
    readiness_probe=probe(
        period_secs=10,
        exec=exec_action(['curl', '-f', 'http://localhost:8081/health']),
    ),
)

local_resource(
    'port-forward-frontend',
    serve_cmd='kubectl port-forward -n invulnerable svc/invulnerable-frontend 3000:80',
    resource_deps=['invulnerable-frontend'],
    labels=['port-forwards'],
)

local_resource(
    'port-forward-oauth2-proxy',
    serve_cmd='kubectl port-forward -n invulnerable svc/invulnerable-oauth2-proxy 4180:4180',
    resource_deps=['invulnerable-oauth2-proxy'],
    labels=['port-forwards'],
)

# Setup instructions
access_url = 'https://invulnerable.local:8443' if enable_https else 'http://invulnerable.local:8080'
https_note = '\\n⚠️  Using self-signed certificate - browser will show warning (expected)' if enable_https else ''
cert_manager_status = '  ✓ cert-manager (HTTPS enabled)' if enable_https else '  - cert-manager (HTTP mode)'
dex_url = 'https://dex.invulnerable.local:8443' if enable_https else 'http://dex.invulnerable.local:8080'
dex_status = '  ✓ Dex OIDC provider ({})'.format(dex_url) if enable_oidc else '  - Dex OIDC provider (disabled)'
auth_note = ''

if enable_oidc:
    auth_note = '''
OIDC Authentication (Dex):
  - Login at: {url}
  - Test users:
    * admin@invulnerable.local / password
    * user@invulnerable.local / password
    * test@example.com / password
  - Dex UI: {dex_url}/dex'''.format(url=access_url, dex_url=dex_url)
else:
    auth_note = '''
OAuth2 Proxy configured for local development.'''

instruction_text = '''✓ Tilt setup complete!

Infrastructure:
  ✓ nginx Ingress Controller
  {cert_status}
  {dex_status}
  ✓ PostgreSQL

Access the application:
  - Main: {url}{note}
  - Backend API: {url}/api
  - Direct Frontend: http://localhost:3000
  - Direct Backend: http://localhost:8081
  - PostgreSQL: localhost:5432

Make sure to add to /etc/hosts:
  127.0.0.1 invulnerable.local{dex_host}

{auth_note}

See TILT.md for more information.'''.format(
    cert_status=cert_manager_status,
    dex_status=dex_status,
    url=access_url,
    note=https_note,
    dex_host=' dex.invulnerable.local' if enable_oidc else '',
    auth_note=auth_note
)

local_resource(
    'setup-instructions',
    cmd='echo "{}"'.format(instruction_text),
    auto_init=True,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['info'],
)

# Watch for changes in Helm chart
watch_file('./helm/invulnerable/values.yaml')
watch_file('./helm/invulnerable/templates/')
watch_file('./tilt/values.yaml')
watch_file('./tilt/values-https.yaml')
watch_file('./tilt/values-oidc.yaml')
watch_file('./tilt/dex-values.yaml')
