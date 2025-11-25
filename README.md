# Invulnerable - Cloud-Native Vulnerability Management Platform

A cloud-native vulnerability management platform that integrates with Anchore tools (Syft + Grype) for container image scanning and vulnerability detection.

## Architecture

- **Backend**: Go 1.21+ with Echo framework, sqlx, and PostgreSQL
- **Frontend**: SvelteKit with TypeScript and TailwindCSS
- **Database**: PostgreSQL with JSONB support for SBOM storage
- **Scanning**: Anchore Syft (SBOM generation) + Grype (vulnerability scanning)
- **Deployment**: Kubernetes

## Features

- ✅ SBOM generation and storage (CycloneDX/SPDX)
- ✅ Vulnerability lifecycle tracking
- ✅ Scan comparison and diff analysis
- ✅ Multi-severity vulnerability management
- ✅ Dashboard with metrics and trends
- ✅ RESTful API
- ✅ Container image inventory
- ✅ Automated scanning via Kubernetes CronJobs

## Project Structure

```
invulnerable/
├── backend/              # Go backend application
│   ├── cmd/server/      # Main application entry point
│   ├── internal/
│   │   ├── api/         # HTTP handlers
│   │   ├── db/          # Database layer (sqlx)
│   │   ├── models/      # Data models
│   │   ├── analyzer/    # Scan diff logic
│   │   └── metrics/     # Metrics service
│   ├── migrations/      # SQL migrations
│   └── Dockerfile
├── frontend/            # SvelteKit frontend
│   ├── src/
│   │   ├── routes/     # SvelteKit pages
│   │   └── lib/        # Components, API client, stores
│   └── Dockerfile
└── k8s/                # Kubernetes manifests
    ├── backend/
    ├── frontend/
    ├── postgres/
    └── cronjob/        # Syft+Grype scanner
```

## Quick Start

### Prerequisites

- Kubernetes cluster
- kubectl configured
- Docker
- [Task](https://taskfile.dev) - Install with: `brew install go-task/tap/go-task` (macOS) or see [installation guide](https://taskfile.dev/installation/)

### 1. Build Images

```bash
# Build all images at once
task build:all

# Or build individually
task build:backend
task build:frontend
task build:scanner
```

### 2. Deploy to Kubernetes

**Quick Deploy (Automated - Recommended)**
```bash
# Deploy everything including automated migrations
task deploy
```

**Or use the quickstart (build + deploy in one command)**
```bash
task quickstart
```

**Manual Deploy**
```bash
# Deploy individual components
task deploy:namespace
task deploy:postgres
task deploy:migrations
task deploy:backend
task deploy:frontend
task deploy:ingress
task deploy:cronjob

# Check deployment status
task status
```

### 3. Run Database Migrations

You have three options for running migrations:

#### **Option A: Kubernetes Job (Recommended for Production)**

This runs migrations as a one-time Kubernetes Job before deploying the app:

```bash
# Deploy migrations as a Job
task deploy:migrations

# Check migration logs
task logs:migration
```

#### **Option B: Init Container (Fully Automated)**

Migrations run automatically as an init container before each pod starts:

```bash
# Use deployment with init container instead of regular deployment
task deploy:backend-with-migrations
```

This approach automatically runs migrations every time the backend deploys, ensuring the database is always up to date.

#### **Option C: Manual Migration (Development Only)**

For local development, you can run migrations manually:

```bash
# Install golang-migrate
brew install golang-migrate  # macOS
# OR
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/  # Linux

# Port-forward to PostgreSQL (in one terminal)
task db:port-forward

# Run migrations (in another terminal)
task db:migrate-up
```

### 4. Access the Application

```bash
# Add to /etc/hosts
echo "127.0.0.1 invulnerable.local" | sudo tee -a /etc/hosts

# Port-forward ingress (if using nginx-ingress)
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80

# Access at http://invulnerable.local:8080
```

## API Endpoints

### Scans
- `POST /api/v1/scans` - Receive scan results from CronJob
- `GET /api/v1/scans` - List scans with pagination
- `GET /api/v1/scans/:id` - Get scan details
- `GET /api/v1/scans/:id/sbom` - Retrieve SBOM
- `GET /api/v1/scans/:id/diff` - Compare with previous scan

### Vulnerabilities
- `GET /api/v1/vulnerabilities` - List vulnerabilities
- `GET /api/v1/vulnerabilities/:cve` - Get CVE details
- `PATCH /api/v1/vulnerabilities/:id` - Update status/notes

### Images
- `GET /api/v1/images` - List tracked images
- `GET /api/v1/images/:id/history` - Scan history

### Metrics
- `GET /api/v1/metrics` - Dashboard metrics

## Migration Strategies Comparison

| Approach | Pros | Cons | Best For |
|----------|------|------|----------|
| **Kubernetes Job** | • Runs once before deployment<br>• Easy to track with logs<br>• Can use Helm hooks | • Requires manual trigger (unless using Helm hooks) | Production with CI/CD pipelines |
| **Init Container** | • Fully automated<br>• Runs on every deploy<br>• No manual steps | • Runs on every pod restart<br>• Slightly slower pod startup | Production with frequent deployments |
| **Manual** | • Full control<br>• Good for testing | • Manual process<br>• Error-prone | Local development only |

**Recommendation:** Use **Kubernetes Job** in CI/CD pipelines or **Init Container** for fully automated deployments.

## Configuration

### Backend Environment Variables

```bash
PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_USER=invulnerable
DB_PASSWORD=changeme
DB_NAME=invulnerable
DB_SSLMODE=disable
```

### Scanner Configuration

Edit `k8s/cronjob/cronjob.yaml` to:
- Change scan schedule (default: daily at 2 AM)
- Update image list in ConfigMap
- Modify resource limits

## Development

### Run Locally

```bash
# Run backend (requires PostgreSQL running locally)
task dev:backend

# Run frontend (in another terminal)
task dev:frontend

# Port-forward to Kubernetes PostgreSQL for local development
task db:port-forward
```

### Testing

The project includes comprehensive tests using [testify](https://github.com/stretchr/testify):

```bash
# Run all tests
task test

# Run tests with coverage report
task test:coverage

# Run only unit tests
task test:unit

# Run tests in watch mode (auto-rerun on file changes)
task test:watch

# Run linter
task lint

# Or use make (in backend directory)
cd backend
make test
make test-coverage
make lint
```

**Test Coverage:**
- ✅ Model tests (image, grype parsing)
- ✅ Database repository tests (with sqlmock)
- ✅ API handler tests
- ✅ Analyzer/diff logic tests
- ✅ Mocked database interactions

**Test Files:**
- `backend/internal/models/*_test.go` - Model unit tests
- `backend/internal/db/*_test.go` - Database layer tests
- `backend/internal/api/*_test.go` - API handler tests
- `backend/internal/analyzer/*_test.go` - Scan comparison tests

### Available Task Commands

View all available commands:
```bash
task --list
```

**Common Commands:**
- `task quickstart` - Build and deploy everything (one command)
- `task build:all` - Build all Docker images
- `task deploy` - Deploy all components to Kubernetes
- `task test` - Run all tests with race detection
- `task test:coverage` - Run tests with coverage report
- `task lint` - Run linter (golangci-lint)
- `task status` - Show deployment status
- `task logs:backend` - Stream backend logs
- `task db:migrate-up` - Run migrations manually
- `task clean` - Delete all resources

**📖 For a complete list of commands and workflows, see [.taskfile.md](./.taskfile.md)**

## Database Schema

- **images**: Container image inventory
- **scans**: Scan metadata and results
- **sboms**: SBOM documents (JSONB)
- **vulnerabilities**: CVE tracking with lifecycle
- **scan_vulnerabilities**: Junction table

## Scanner Workflow

1. CronJob runs on schedule
2. Syft generates SBOM for target image
3. Grype scans SBOM for vulnerabilities
4. Results are POSTed to `/api/v1/scans`
5. Backend processes and stores:
   - SBOM document
   - Vulnerability records
   - Scan metadata
6. Analyzer compares with previous scan
7. Lifecycle tracking updates (new/fixed/persistent)

## Vulnerability Lifecycle

- **active**: Currently present in latest scan
- **fixed**: No longer present in latest scan
- **ignored**: Manually marked as ignored
- **accepted**: Risk accepted

## Metrics

- Total images, scans, vulnerabilities
- Active vulnerabilities by severity
- Vulnerability trends (30 days)
- Recent scan activity

## Production Considerations

1. **Security**:
   - Change default PostgreSQL password
   - Enable TLS/SSL for database
   - Use proper secrets management
   - Implement authentication for API

2. **Scaling**:
   - HPA configured for backend
   - Adjust resource limits based on workload
   - Consider read replicas for PostgreSQL

3. **Monitoring**:
   - Add Prometheus metrics
   - Configure alerting
   - Log aggregation

4. **Backup**:
   - PostgreSQL backups
   - SBOM retention policy

## Task vs Make

This project uses [Task](https://taskfile.dev) instead of Make for the following benefits:

- **YAML-based**: More readable and easier to maintain than Makefile syntax
- **Cross-platform**: Works consistently on Linux, macOS, and Windows
- **Built-in features**: Variables, dependencies, file watching, parallel execution
- **Better error handling**: Clear error messages and status codes
- **Developer-friendly**: Autocomplete support and better documentation

### Install Task

```bash
# macOS
brew install go-task/tap/go-task

# Linux (via script)
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Windows (via Scoop)
scoop install task

# Or download from: https://github.com/go-task/task/releases
```

**Note:** The Makefile is still included for backwards compatibility, but Task is recommended.

## License

MIT

## Contributing

Pull requests welcome!
