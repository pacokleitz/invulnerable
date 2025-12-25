# Tiltfile for local development with Docker Desktop Kubernetes
# https://docs.tilt.dev/

# Allow k8s contexts for Docker Desktop
allow_k8s_contexts('docker-desktop')

# Load Tilt extensions
load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create')

# Configuration
config.define_string('registry', args=False, usage='Docker registry for images')
config.define_bool('enable-https', args=True, usage='Enable HTTPS with cert-manager')
config.define_bool('enable-oidc', args=True, usage='Enable OIDC with Dex provider')
cfg = config.parse()
registry = cfg.get('registry', 'localhost:5001')
enable_https = cfg.get('enable-https', False)
enable_oidc = cfg.get('enable-oidc', False)

# Add Helm repositories
helm_repo('ingress-nginx', 'https://kubernetes.github.io/ingress-nginx', labels=['infrastructure'])
helm_repo('jetstack', 'https://charts.jetstack.io', labels=['infrastructure'])
helm_repo('bitnami', 'https://charts.bitnami.com/bitnami', labels=['infrastructure'])
helm_repo('dex', 'https://charts.dexidp.io', labels=['infrastructure'])

# Deploy nginx Ingress Controller
# For Docker Desktop: Uses hostPort to bind directly to ports 80/443
# Traffic flow: localhost:80 -> ingress pod:80 -> services
print('📦 Deploying nginx Ingress Controller...')
helm_resource(
    name='ingress-nginx',
    chart='ingress-nginx/ingress-nginx',
    namespace='ingress-nginx',
    flags=[
        '--create-namespace',
        '--set=controller.hostPort.enabled=true',  # Binds to localhost ports 80/443
        '--set=controller.service.type=NodePort',
        '--set=controller.watchIngressWithoutClass=true',
    ],
    labels=['infrastructure'],
    resource_deps=[],
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
    ref=f'{registry}/invulnerable-backend',
    context='./backend',
    dockerfile='./backend/Dockerfile',
    live_update=[
        # Sync Go source files
        sync('./backend', '/app'),
        # Rebuild on Go file changes
        run('cd /app && go build -o /app/server ./cmd/server', trigger=['./backend/**/*.go']),
    ],
    # Only rebuild on these file changes
    only=[
        './backend/cmd/',
        './backend/internal/',
        './backend/go.mod',
        './backend/go.sum',
    ],
)

docker_build(
    ref=f'{registry}/invulnerable-frontend',
    context='./frontend',
    dockerfile='./frontend/Dockerfile',
    live_update=[
        # Sync source files for Vite HMR
        sync('./frontend/src', '/app/src'),
        sync('./frontend/index.html', '/app/index.html'),
        sync('./frontend/package.json', '/app/package.json'),
    ],
    only=[
        './frontend/src/',
        './frontend/index.html',
        './frontend/package.json',
        './frontend/tsconfig.json',
        './frontend/vite.config.ts',
        './frontend/tailwind.config.js',
    ],
)

docker_build(
    ref=f'{registry}/invulnerable-controller',
    context='./controller',
    dockerfile='./controller/Dockerfile',
    live_update=[
        sync('./controller', '/workspace'),
        run('cd /workspace && make build', trigger=['./controller/**/*.go']),
    ],
    only=[
        './controller/api/',
        './controller/internal/',
        './controller/cmd/',
        './controller/go.mod',
        './controller/go.sum',
    ],
)

docker_build(
    ref=f'{registry}/invulnerable-scanner',
    context='./scanner',
    dockerfile='./scanner/Dockerfile',
    only=[
        './scanner/scan.sh',
        './scanner/Dockerfile',
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

helm_resource(
    name='invulnerable',
    chart='./helm/invulnerable',
    namespace='invulnerable',
    flags=[
        f'--values={values_file}',
        f'--set=image.registry={registry}',
        '--set=backend.image.tag=latest',
        '--set=frontend.image.tag=latest',
        '--set=controller.image.tag=latest',
        '--set=scanner.image.tag=latest',
    ],
    resource_deps=helm_deps,
    labels=['app'],
)

# Port forwards for easy access
k8s_resource(
    workload='invulnerable-backend',
    port_forwards=['8080:8080'],
    labels=['app'],
    resource_deps=['invulnerable'],
)

k8s_resource(
    workload='invulnerable-frontend',
    port_forwards=['3000:8080'],
    labels=['app'],
    resource_deps=['invulnerable'],
)

k8s_resource(
    workload='invulnerable-controller',
    labels=['app'],
    resource_deps=['invulnerable'],
)

k8s_resource(
    workload='invulnerable-oauth2-proxy',
    port_forwards=['4180:4180'],
    labels=['auth'],
    resource_deps=['invulnerable'],
)

# Setup instructions
access_url = 'https://invulnerable.local' if enable_https else 'http://invulnerable.local'
https_note = '\\n⚠️  Using self-signed certificate - browser will show warning (expected)' if enable_https else ''
cert_manager_status = '  ✓ cert-manager (HTTPS enabled)' if enable_https else '  - cert-manager (HTTP mode)'
dex_status = '  ✓ Dex OIDC provider (http://dex.invulnerable.local)' if enable_oidc else '  - Dex OIDC provider (disabled)'
auth_note = ''

if enable_oidc:
    auth_note = '''
OIDC Authentication (Dex):
  - Login at: {url}
  - Test users:
    * admin@invulnerable.local / password
    * user@invulnerable.local / password
    * test@example.com / password
  - Dex UI: http://dex.invulnerable.local/dex'''.format(url=access_url)
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
  - Direct Backend: http://localhost:8080
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
