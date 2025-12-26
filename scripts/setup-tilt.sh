#!/bin/bash
# Setup script for Tilt local development with Docker Desktop
# Tilt handles nginx-ingress, cert-manager, and PostgreSQL installation

set -e  # Exit on error

echo "🚀 Setting up Docker Desktop Kubernetes for Invulnerable..."
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${RED}✗ $1 is not installed${NC}"
        echo "  Install it with: $2"
        return 1
    else
        echo -e "${GREEN}✓ $1 is installed${NC}"
        return 0
    fi
}

echo "Checking prerequisites..."
MISSING_DEPS=0

check_command docker "https://docs.docker.com/get-docker/" || MISSING_DEPS=1
check_command kubectl "brew install kubectl or https://kubernetes.io/docs/tasks/tools/" || MISSING_DEPS=1
check_command tilt "https://docs.tilt.dev/install.html" || MISSING_DEPS=1

if [ $MISSING_DEPS -eq 1 ]; then
    echo
    echo -e "${RED}Please install missing dependencies and try again.${NC}"
    exit 1
fi

echo
echo "All prerequisites installed!"
echo

# Check if Docker Desktop Kubernetes is enabled
if ! kubectl cluster-info --context docker-desktop &> /dev/null; then
    echo -e "${RED}✗ Docker Desktop Kubernetes is not running${NC}"
    echo
    echo "Please enable Kubernetes in Docker Desktop:"
    echo "  1. Open Docker Desktop"
    echo "  2. Go to Settings > Kubernetes"
    echo "  3. Check 'Enable Kubernetes'"
    echo "  4. Click 'Apply & Restart'"
    echo
    exit 1
fi

echo -e "${GREEN}✓ Docker Desktop Kubernetes is running${NC}"
echo

# Switch to docker-desktop context
echo "Switching to docker-desktop context..."
kubectl config use-context docker-desktop
echo -e "${GREEN}✓ Context switched to docker-desktop${NC}"
echo

# Verify cluster is accessible
kubectl cluster-info
echo

# Add Helm repositories for Tilt
echo "Adding Helm repositories..."
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2>/dev/null || echo -e "${GREEN}✓ ingress-nginx repo already added${NC}"
helm repo add jetstack https://charts.jetstack.io 2>/dev/null || echo -e "${GREEN}✓ jetstack repo already added${NC}"
helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || echo -e "${GREEN}✓ bitnami repo already added${NC}"
helm repo add dex https://charts.dexidp.io 2>/dev/null || echo -e "${GREEN}✓ dex repo already added${NC}"
echo "Updating Helm repositories..."
helm repo update >/dev/null 2>&1
echo -e "${GREEN}✓ Helm repositories configured${NC}"
echo

# Configure /etc/hosts
if ! grep -q "invulnerable.local" /etc/hosts; then
    echo -e "${YELLOW}⚠ Adding invulnerable.local to /etc/hosts (requires sudo)${NC}"
    echo "127.0.0.1 invulnerable.local" | sudo tee -a /etc/hosts
    echo -e "${GREEN}✓ Added invulnerable.local to /etc/hosts${NC}"
else
    echo -e "${GREEN}✓ invulnerable.local already in /etc/hosts${NC}"
fi

# Add dex.invulnerable.local for OIDC testing
if ! grep -q "dex.invulnerable.local" /etc/hosts; then
    echo -e "${YELLOW}⚠ Adding dex.invulnerable.local to /etc/hosts (requires sudo)${NC}"
    echo "127.0.0.1 dex.invulnerable.local" | sudo tee -a /etc/hosts
    echo -e "${GREEN}✓ Added dex.invulnerable.local to /etc/hosts${NC}"
else
    echo -e "${GREEN}✓ dex.invulnerable.local already in /etc/hosts${NC}"
fi

echo
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo
echo "What was configured:"
echo "  ✓ kubectl context: docker-desktop"
echo "  ✓ Helm repositories: ingress-nginx, jetstack, bitnami, dex"
echo "  ✓ /etc/hosts configured (invulnerable.local, dex.invulnerable.local)"
echo
echo "Next steps:"
echo
echo "1. Start Tilt in your preferred mode:"
echo
echo "   HTTP mode (default):"
echo -e "   ${YELLOW}tilt up${NC}"
echo
echo "   HTTPS mode (with cert-manager):"
echo -e "   ${YELLOW}tilt up -- --enable-https=true${NC}"
echo
echo "   OIDC mode (with local Dex provider):"
echo -e "   ${YELLOW}tilt up -- --enable-oidc=true${NC}"
echo
echo "2. Access the Tilt UI:"
echo -e "   ${YELLOW}http://localhost:10350${NC}"
echo
echo "3. Tilt will automatically deploy:"
echo "   - nginx Ingress Controller"
echo "   - cert-manager (if --enable-https flag used)"
echo "   - Dex OIDC provider (if --enable-oidc flag used)"
echo "   - PostgreSQL"
echo "   - Invulnerable application (backend, frontend, controller)"
echo "   - OAuth2 Proxy"
echo
echo "4. Access the application after Tilt finishes:"
echo -e "   ${YELLOW}http://invulnerable.local${NC} (HTTP mode)"
echo -e "   ${YELLOW}https://invulnerable.local${NC} (with --enable-https)"
echo
echo "For more information, see TILT.md"
echo
