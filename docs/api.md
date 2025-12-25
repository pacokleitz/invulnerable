# Invulnerable API Reference

This document provides detailed information about the Invulnerable REST API.

## Base URL

```
http://api-host/api/v1
```

## Authentication

When OAuth2 Proxy is enabled, all API requests must include authentication headers automatically added by the proxy:
- `X-Auth-Request-User`
- `X-Auth-Request-Email`
- `Authorization` (Bearer token)

## Endpoints

### Scans

#### Submit Scan Results

```http
POST /scans
Content-Type: application/json
```

**Request Body:**
```json
{
  "image": "nginx:latest",
  "digest": "sha256:abc123...",
  "sbom_format": "cyclonedx",
  "sbom": { ... },
  "vulnerabilities": [
    {
      "cve_id": "CVE-2023-1234",
      "severity": "high",
      "package_name": "libssl",
      "package_version": "1.1.1",
      "fixed_version": "1.1.2",
      "description": "...",
      "cvss_score": 7.5
    }
  ]
}
```

**Response:**
```json
{
  "id": 123,
  "image_id": 45,
  "scan_time": "2024-01-15T10:30:00Z",
  "vulnerabilities_count": 15
}
```

#### List Scans

```http
GET /scans?limit=20&offset=0&image_id=45
```

**Query Parameters:**
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `image_id` (optional): Filter by image ID

**Response:**
```json
{
  "scans": [
    {
      "id": 123,
      "image_id": 45,
      "image_name": "nginx:latest",
      "digest": "sha256:abc123...",
      "scan_time": "2024-01-15T10:30:00Z",
      "vulnerabilities_count": 15,
      "critical_count": 2,
      "high_count": 5,
      "medium_count": 6,
      "low_count": 2
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0
}
```

#### Get Scan Details

```http
GET /scans/{id}
```

**Response:**
```json
{
  "id": 123,
  "image_id": 45,
  "image_name": "nginx:latest",
  "digest": "sha256:abc123...",
  "scan_time": "2024-01-15T10:30:00Z",
  "sbom_format": "cyclonedx",
  "vulnerabilities": [
    {
      "id": 456,
      "cve_id": "CVE-2023-1234",
      "severity": "high",
      "package_name": "libssl",
      "package_version": "1.1.1",
      "fixed_version": "1.1.2",
      "status": "active",
      "first_detected": "2024-01-10T08:00:00Z"
    }
  ]
}
```

#### Get SBOM

```http
GET /scans/{id}/sbom
```

**Response:** Returns the raw SBOM document (CycloneDX or SPDX JSON)

#### Compare Scans (Diff)

```http
GET /scans/{id}/diff
```

Compares the specified scan with the previous scan of the same image.

**Response:**
```json
{
  "current_scan_id": 123,
  "previous_scan_id": 120,
  "current_scan_time": "2024-01-15T10:30:00Z",
  "previous_scan_time": "2024-01-14T10:30:00Z",
  "summary": {
    "new_vulnerabilities": 3,
    "fixed_vulnerabilities": 5,
    "persistent_vulnerabilities": 10,
    "total_current": 13,
    "total_previous": 15
  },
  "new": [
    {
      "cve_id": "CVE-2024-0001",
      "severity": "critical",
      "package_name": "curl",
      "package_version": "7.88.0"
    }
  ],
  "fixed": [
    {
      "cve_id": "CVE-2023-9999",
      "severity": "high",
      "package_name": "openssl",
      "package_version": "3.0.0"
    }
  ],
  "persistent": [ ... ]
}
```

### Vulnerabilities

#### List Vulnerabilities

```http
GET /vulnerabilities?severity=critical&status=active&limit=20&offset=0
```

**Query Parameters:**
- `severity` (optional): Filter by severity (critical, high, medium, low, negligible)
- `status` (optional): Filter by status (active, fixed, ignored, accepted)
- `cve` (optional): Search by CVE ID
- `package` (optional): Search by package name
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:**
```json
{
  "vulnerabilities": [
    {
      "id": 456,
      "cve_id": "CVE-2023-1234",
      "severity": "high",
      "package_name": "libssl",
      "package_version": "1.1.1",
      "fixed_version": "1.1.2",
      "status": "active",
      "affected_images_count": 3,
      "first_detected": "2024-01-10T08:00:00Z",
      "last_seen": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 250,
  "limit": 20,
  "offset": 0
}
```

#### Get Vulnerability Details

```http
GET /vulnerabilities/{cve}
```

**Response:**
```json
{
  "cve_id": "CVE-2023-1234",
  "severity": "high",
  "cvss_score": 7.5,
  "description": "Buffer overflow vulnerability in libssl...",
  "published_date": "2023-05-01T00:00:00Z",
  "affected_images": [
    {
      "image_id": 45,
      "image_name": "nginx:latest",
      "package_name": "libssl",
      "package_version": "1.1.1",
      "fixed_version": "1.1.2",
      "status": "active",
      "first_detected": "2024-01-10T08:00:00Z"
    }
  ]
}
```

#### Update Vulnerability Status

```http
PATCH /vulnerabilities/{id}
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "accepted",
  "notes": "Risk accepted for legacy system - will upgrade in Q2"
}
```

**Response:**
```json
{
  "id": 456,
  "status": "accepted",
  "notes": "Risk accepted for legacy system - will upgrade in Q2",
  "updated_at": "2024-01-15T14:30:00Z"
}
```

### Images

#### List Images

```http
GET /images?limit=20&offset=0
```

**Query Parameters:**
- `limit` (optional): Number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset (default: 0)

**Response:**
```json
{
  "images": [
    {
      "id": 45,
      "name": "nginx:latest",
      "registry": "docker.io",
      "repository": "library/nginx",
      "tag": "latest",
      "digest": "sha256:abc123...",
      "last_scan_time": "2024-01-15T10:30:00Z",
      "total_scans": 30,
      "active_vulnerabilities": 13,
      "critical_count": 2,
      "high_count": 5
    }
  ],
  "total": 50,
  "limit": 20,
  "offset": 0
}
```

#### Get Image Details

```http
GET /images/{id}
```

**Response:**
```json
{
  "id": 45,
  "name": "nginx:latest",
  "registry": "docker.io",
  "repository": "library/nginx",
  "tag": "latest",
  "digest": "sha256:abc123...",
  "first_seen": "2024-01-01T10:00:00Z",
  "last_scan_time": "2024-01-15T10:30:00Z",
  "total_scans": 30,
  "active_vulnerabilities": 13,
  "vulnerability_summary": {
    "critical": 2,
    "high": 5,
    "medium": 4,
    "low": 2,
    "negligible": 0
  }
}
```

#### Get Image Scan History

```http
GET /images/{id}/history?limit=20&offset=0
```

**Response:**
```json
{
  "scans": [
    {
      "id": 123,
      "scan_time": "2024-01-15T10:30:00Z",
      "digest": "sha256:abc123...",
      "vulnerabilities_count": 13,
      "critical_count": 2,
      "high_count": 5,
      "changes_from_previous": {
        "new": 3,
        "fixed": 5
      }
    }
  ],
  "total": 30,
  "limit": 20,
  "offset": 0
}
```

### Metrics

#### Get Dashboard Metrics

```http
GET /metrics
```

**Response:**
```json
{
  "summary": {
    "total_images": 50,
    "total_scans": 1250,
    "total_vulnerabilities": 450,
    "active_vulnerabilities": 280
  },
  "by_severity": {
    "critical": 15,
    "high": 45,
    "medium": 120,
    "low": 80,
    "negligible": 20
  },
  "recent_scans": [
    {
      "id": 123,
      "image_name": "nginx:latest",
      "scan_time": "2024-01-15T10:30:00Z",
      "vulnerabilities_count": 13
    }
  ],
  "trend": {
    "labels": ["2024-01-08", "2024-01-09", "2024-01-10", ...],
    "critical": [12, 13, 15, ...],
    "high": [40, 42, 45, ...],
    "medium": [115, 118, 120, ...],
    "low": [75, 78, 80, ...]
  }
}
```

## Error Responses

All endpoints return standard HTTP status codes:

- `200 OK` - Success
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

**Error Response Format:**
```json
{
  "error": "Invalid request",
  "message": "Field 'image' is required",
  "code": "VALIDATION_ERROR"
}
```

## Rate Limiting

Currently, there are no rate limits enforced. This may change in future versions.

## Pagination

List endpoints support pagination using `limit` and `offset` parameters:

- `limit`: Number of results to return (default: 50, max: 100)
- `offset`: Number of results to skip (default: 0)

Responses include pagination metadata:
```json
{
  "total": 250,
  "limit": 20,
  "offset": 40
}
```

## Best Practices

1. **Use pagination** for list endpoints to avoid large responses
2. **Filter results** using query parameters when possible
3. **Cache responses** where appropriate (e.g., SBOM documents)
4. **Handle errors gracefully** and check HTTP status codes
5. **Use HTTPS** in production environments
6. **Store sensitive data** (API tokens) securely

## Examples

### Bash/cURL Examples

**Get recent scans:**
```bash
curl -X GET "http://api/v1/scans?limit=10" | jq
```

**Submit scan results:**
```bash
curl -X POST "http://api/v1/scans" \
  -H "Content-Type: application/json" \
  -d @scan-results.json
```

**Update vulnerability status:**
```bash
curl -X PATCH "http://api/v1/vulnerabilities/456" \
  -H "Content-Type: application/json" \
  -d '{"status":"accepted","notes":"Risk accepted"}'
```

### Python Example

```python
import requests

API_BASE = "http://api/v1"

# Get active critical vulnerabilities
response = requests.get(
    f"{API_BASE}/vulnerabilities",
    params={"severity": "critical", "status": "active"}
)
vulnerabilities = response.json()["vulnerabilities"]

for vuln in vulnerabilities:
    print(f"{vuln['cve_id']}: {vuln['package_name']} in {vuln['affected_images_count']} images")
```

### Go Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Scan struct {
    ID               int    `json:"id"`
    ImageName        string `json:"image_name"`
    VulnCount        int    `json:"vulnerabilities_count"`
}

func main() {
    resp, _ := http.Get("http://api/v1/scans?limit=10")
    defer resp.Body.Close()

    var result struct {
        Scans []Scan `json:"scans"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    for _, scan := range result.Scans {
        fmt.Printf("%s: %d vulnerabilities\n", scan.ImageName, scan.VulnCount)
    }
}
```

## Support

For API questions or issues:
- GitHub Issues: https://github.com/pacokleitz/invulnerable/issues
- Email: kpaco@proton.me
