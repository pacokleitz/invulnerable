# Testing Strategy for Vulnerability Scanning

## Fixture-Based Testing

We use **fixture files** instead of running real scans during tests to ensure:

1. **Deterministic results** - CVE databases change over time, fixtures don't
2. **Fast tests** - No need to run real Grype scans
3. **Controlled scenarios** - Test specific edge cases with artificial data
4. **Reproducible** - Same fixtures produce same results every time

## Test Coverage

### Scan Processing Tests (`internal/api/scans_fixture_test.go`)

Tests the core vulnerability filtering logic:
- ✅ `TestFilterVulnerabilitiesByOnlyFixed_Mixed` - Filter fixed/unfixed from mixed set
- ✅ `TestFilterVulnerabilitiesByOnlyFixed_AllFixed` - Handle all-fixed scenario
- ✅ `TestFilterVulnerabilitiesByOnlyFixed_AllUnfixed` - Handle all-unfixed scenario
- ✅ `TestFilterVulnerabilitiesByOnlyFixed_Disabled` - Test with filtering disabled
- ✅ `TestSeverityCounts_WithFixtures` - Verify severity counting with/without filtering

### Webhook Notification Tests (`internal/notifier/notifier_fixture_test.go`)

Tests webhook notification behavior with fixtures:
- ✅ `TestWebhookNotification_OnlyFixed_Mixed` - Webhook with mixed CVEs (onlyFixed=true)
- ✅ `TestWebhookNotification_OnlyFixed_AllUnfixed` - Webhook skipped when no fixed CVEs
- ✅ `TestWebhookNotification_OnlyFixed_Disabled` - Webhook includes unfixed when disabled
- ✅ `TestStatusChangeNotification_OnlyFixed` - Status change notifications respect onlyFixed

## Available Fixtures

### grype-output-mixed.json
**Purpose**: Test filtering and severity handling
**Contents**:
- 1 Critical with fix
- 1 High without fix
- 1 Medium with fix
- 1 Low without fix

**Expected behavior when `onlyFixed=true`**:
- Only 2 CVEs pass filter (Critical + Medium)
- Webhook sends notification for 2 vulnerabilities

### grype-output-all-fixed.json
**Purpose**: Test scenario where all CVEs are fixable
**Contents**:
- 1 Critical with fix
- 1 High with fix

**Expected behavior when `onlyFixed=true`**:
- All 2 CVEs pass filter
- Webhook sends notification for 2 vulnerabilities

### grype-output-all-unfixed.json
**Purpose**: Test scenario where no CVEs are fixable
**Contents**:
- 1 Critical without fix
- 1 High without fix

**Expected behavior when `onlyFixed=true`**:
- 0 CVEs pass filter
- Webhook is **NOT** triggered (TotalVulns=0)

## Creating New Fixtures

When you need to test a new scenario:

1. **Option A: Generate from real scan**
   ```bash
   grype nginx:latest -o json > testdata/grype-output-nginx.json
   ```

2. **Option B: Create manually**
   - Copy an existing fixture
   - Modify CVE IDs, packages, severities as needed
   - Set `fix.versions` to `[]` for unfixed, `["1.2.3"]` for fixed

3. **Minimal fixture** (remove unnecessary fields):
   ```json
   {
     "matches": [
       {
         "vulnerability": {
           "id": "CVE-2024-XXX",
           "severity": "High",
           "fix": {"versions": ["1.2.3"], "state": "fixed"}
         },
         "artifact": {
           "name": "package-name",
           "version": "1.0.0"
         }
       }
     ]
   }
   ```

## Best Practices

1. ✅ **DO** use fixtures for testing business logic
2. ✅ **DO** keep fixtures minimal (only required fields)
3. ✅ **DO** name CVEs descriptively (CVE-2024-FIXED-1, CVE-2024-UNFIXED-1)
4. ✅ **DO** document what each fixture tests in README.md
5. ❌ **DON'T** run real scans in unit/integration tests
6. ❌ **DON'T** rely on external CVE databases in tests
7. ❌ **DON'T** use fixtures for E2E tests (those should use real scans)

## Migration Testing

The `testhelpers.go` now automatically discovers and runs all migrations:
- No need to update code when adding new migrations
- Consistent with production (init container also auto-discovers)
- Just add `NNN_migration_name.up.sql` to `migrations/` folder
