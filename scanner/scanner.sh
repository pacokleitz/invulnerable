#!/bin/bash
set -e

# Scanner script for Syft + Grype
# This script runs as a Kubernetes CronJob to scan container images

# Use mounted workspace for temporary files
export TMPDIR=/tmp/syft
mkdir -p "$TMPDIR"

# Configuration from environment variables
IMAGE="${SCAN_IMAGE}"
API_ENDPOINT="${API_ENDPOINT:-http://backend.invulnerable.svc.cluster.local:8080}"
SBOM_FORMAT="${SBOM_FORMAT:-cyclonedx}"

if [ -z "$IMAGE" ]; then
    echo "Error: SCAN_IMAGE environment variable is required"
    exit 1
fi

echo "========================================="
echo "Starting scan for image: $IMAGE"
echo "========================================="

# Create temporary directory for outputs
TEMP_DIR=$(mktemp -d)
SBOM_FILE="$TEMP_DIR/sbom.json"
GRYPE_FILE="$TEMP_DIR/grype.json"

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Step 1: Generate SBOM with Syft
echo "Step 1: Generating SBOM with Syft..."
syft "$IMAGE" -o "${SBOM_FORMAT}-json" > "$SBOM_FILE"

if [ $? -ne 0 ]; then
    echo "Error: Syft SBOM generation failed"
    exit 1
fi

echo "SBOM generated successfully"

# Step 2: Scan SBOM with Grype
echo "Step 2: Scanning SBOM with Grype..."

# Build Grype flags
GRYPE_FLAGS="-o json"
if [ "${ONLY_FIXABLE_VULNS:-false}" = "true" ]; then
    echo "Filtering to only show vulnerabilities with available fixes"
    GRYPE_FLAGS="$GRYPE_FLAGS --only-fixed"
fi

# Run Grype with optional filtering
grype "sbom:$SBOM_FILE" $GRYPE_FLAGS > "$GRYPE_FILE"

if [ $? -ne 0 ]; then
    echo "Error: Grype scan failed"
    exit 1
fi

echo "Grype scan completed successfully"

# Step 3: Prepare payload
echo "Step 3: Preparing payload for API..."

# Extract SBOM version if available
SBOM_VERSION=$(jq -r '.bomFormat + " " + .specVersion' "$SBOM_FILE" 2>/dev/null || echo "unknown")

# Extract image digest from Grype source target
IMAGE_DIGEST=$(jq -r '.source.target.imageID // .source.target.repoDigests[0] // empty' "$GRYPE_FILE" 2>/dev/null || echo "")

# Create JSON payload by building it in pieces to avoid ARG_MAX issues
PAYLOAD_FILE="$TEMP_DIR/payload.json"
META_FILE="$TEMP_DIR/meta.json"

# Create metadata JSON
jq -n \
    --arg image "$IMAGE" \
    --arg sbom_format "$SBOM_FORMAT" \
    --arg sbom_version "$SBOM_VERSION" \
    --arg image_digest "$IMAGE_DIGEST" \
    --arg webhook_url "${WEBHOOK_URL:-}" \
    --arg webhook_format "${WEBHOOK_FORMAT:-}" \
    --arg webhook_min_severity "${WEBHOOK_MIN_SEVERITY:-}" \
    --arg webhook_only_fixable "${WEBHOOK_ONLY_FIXABLE:-true}" \
    --arg sla_critical "${SLA_CRITICAL:-7}" \
    --arg sla_high "${SLA_HIGH:-30}" \
    --arg sla_medium "${SLA_MEDIUM:-90}" \
    --arg sla_low "${SLA_LOW:-180}" \
    --arg imagescan_namespace "${IMAGESCAN_NAMESPACE:-}" \
    --arg imagescan_name "${IMAGESCAN_NAME:-}" \
    '{
        image: $image,
        sbom_format: $sbom_format,
        sbom_version: $sbom_version,
        image_digest: (if $image_digest != "" then $image_digest else null end),
        webhook_config: (
            if $webhook_url != "" then {
                url: $webhook_url,
                format: $webhook_format,
                min_severity: $webhook_min_severity,
                only_fixable: ($webhook_only_fixable == "true")
            } else null end
        ),
        sla_config: {
            critical: ($sla_critical | tonumber),
            high: ($sla_high | tonumber),
            medium: ($sla_medium | tonumber),
            low: ($sla_low | tonumber)
        },
        imagescan_context: (
            if $imagescan_namespace != "" and $imagescan_name != "" then {
                namespace: $imagescan_namespace,
                name: $imagescan_name
            } else null end
        )
    }' > "$META_FILE"

# Merge metadata with SBOM and Grype results
jq -s '.[0] + {sbom: .[1], grype_result: .[2]}' "$META_FILE" "$SBOM_FILE" "$GRYPE_FILE" > "$PAYLOAD_FILE"

# Step 4: Send to API
echo "Step 4: Sending results to API at $API_ENDPOINT/api/v1/scans"

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d @"$PAYLOAD_FILE" \
    "$API_ENDPOINT/api/v1/scans")

if [ "$HTTP_CODE" -ge 200 ] && [ "$HTTP_CODE" -lt 300 ]; then
    echo "✓ Scan results successfully uploaded (HTTP $HTTP_CODE)"
    echo "========================================="
    echo "Scan completed successfully for: $IMAGE"
    echo "========================================="
    exit 0
else
    echo "✗ Failed to upload scan results (HTTP $HTTP_CODE)"
    exit 1
fi
