# Test Fixtures

This directory contains test fixture files for integration testing.

## Grype Output Fixtures

These JSON files contain example Grype scan outputs with artificial CVEs for testing purposes.

### grype-output-mixed.json
Contains a mix of vulnerabilities with different severities and fix states:
- CVE-2024-1234: Critical severity, **with fix** (openssl)
- CVE-2024-5678: High severity, **without fix** (curl)
- CVE-2024-9999: Medium severity, **with fix** (libxml2)
- CVE-2024-1111: Low severity, **without fix** (bash)

Use this fixture to test:
- Severity filtering
- OnlyFixed filtering (should filter out 2 unfixed CVEs)
- Webhook notifications with different configurations

### grype-output-all-fixed.json
Contains only vulnerabilities with fixes available:
- CVE-2024-FIXED-1: Critical severity, with fix
- CVE-2024-FIXED-2: High severity, with fix

Use this fixture to test:
- OnlyFixed filtering (all should pass)
- Webhook notifications when all CVEs are fixable

### grype-output-all-unfixed.json
Contains only vulnerabilities without fixes:
- CVE-2024-UNFIXED-1: Critical severity, no fix
- CVE-2024-UNFIXED-2: High severity, no fix

Use this fixture to test:
- OnlyFixed filtering (all should be filtered out)
- Webhook notifications when onlyFixed=true (should not trigger)

## Usage

Load fixtures in tests:
```go
data, err := os.ReadFile("testdata/grype-output-mixed.json")
if err != nil {
    t.Fatal(err)
}
var grypeOutput models.GrypeResult
json.Unmarshal(data, &grypeOutput)
```
