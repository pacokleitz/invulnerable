# Testing Documentation

This document describes the testing setup for the Invulnerable vulnerability management platform.

## Test Framework

The project uses the following testing tools:

- **[testify](https://github.com/stretchr/testify)** - Assertions, mocking, and test suites
- **[sqlmock](https://github.com/DATA-DOG/go-sqlmock)** - SQL database mocking
- **Go built-in testing** - Standard library testing package
- **Race detector** - Concurrent code testing (`-race` flag)

## Running Tests

### Quick Start

```bash
# Run all tests
task test

# Run tests with coverage
task test:coverage

# Run tests in watch mode
task test:watch

# Run linter
task lint
```

### Alternative (Make)

```bash
cd backend
make test              # Run all tests
make test-coverage     # Generate coverage report
make lint             # Run linter
```

### Manual

```bash
cd backend

# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/models

# Run specific test
go test -run TestImage_FullName ./internal/models
```

## Test Coverage

### Models (`internal/models/*_test.go`)

**`image_test.go`**
- ✅ `TestImage_FullName` - Tests image name formatting
  - With registry
  - Without registry
  - Custom registry

**`grype_test.go`**
- ✅ `TestGrypeResult_UnmarshalJSON` - Tests Grype JSON parsing
- ✅ `TestGrypeVulnerability_WithFix` - Tests fix version handling

### Database Layer (`internal/db/*_test.go`)

**`images_test.go`**
- ✅ `TestImageRepository_Create` - Tests image creation with upsert logic
- ✅ `TestImageRepository_GetByID` - Tests fetching image by ID
- ✅ `TestImageRepository_List` - Tests listing images with stats

**`vulnerabilities_test.go`**
- ✅ `TestVulnerabilityRepository_Upsert` - Tests vulnerability upsert
- ✅ `TestVulnerabilityRepository_GetByCVE` - Tests CVE lookup
- ✅ `TestVulnerabilityRepository_MarkAsFixed` - Tests marking vulnerabilities as fixed

### API Handlers (`internal/api/*_test.go`)

**`scans_test.go`**
- ✅ `TestParseImageName` - Tests image name parsing logic
  - Docker Hub format
  - GCR format
  - Localhost registry
  - Tag defaults
- ✅ `TestNormalizeSeverity` - Tests severity normalization
- ✅ `TestScanHandler_CreateScan_ValidRequest` - Tests scan creation endpoint

### Analyzer (`internal/analyzer/*_test.go`)

**`diff_test.go`**
- ✅ `TestMakeVulnKey` - Tests vulnerability key generation
- ✅ `TestAnalyzer_CompareScan_NoPreviousScan` - Tests first scan (all new vulnerabilities)
- ✅ `TestAnalyzer_CompareScan_WithPreviousScan` - Tests scan comparison
  - New vulnerabilities detection
  - Fixed vulnerabilities detection
  - Persistent vulnerabilities detection
  - Automatic marking of fixed vulnerabilities

## Test Patterns

### Database Testing with sqlmock

```go
func TestRepository_Method(t *testing.T) {
    // Setup mock database
    db, mock := setupMockDB(t)
    defer db.Close()

    repo := NewRepository(db)

    // Define expected query and result
    mock.ExpectQuery(`SELECT \* FROM table`).
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
            AddRow(1, "test"))

    // Execute method
    result, err := repo.Method(context.Background(), 1)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "test", result.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Mocking with testify/mock

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Method(ctx context.Context, id int) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func TestWithMock(t *testing.T) {
    mockRepo := new(MockRepository)

    // Setup expectations
    mockRepo.On("Method", mock.Anything, 123).Return(nil)

    // Use mock
    err := mockRepo.Method(context.Background(), 123)

    // Verify
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Table-Driven Tests

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "output1"},
        {"case 2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## CI/CD Integration

Tests run automatically on every push and pull request via GitHub Actions.

**Workflow: `.github/workflows/test.yml`**

Jobs:
- **backend-test** - Runs all backend tests with coverage
- **backend-lint** - Runs golangci-lint
- **frontend-test** - Type checks and builds frontend

**Workflow: `.github/workflows/docker.yml`**

Builds Docker images for:
- Backend
- Frontend
- Scanner

## Coverage Goals

| Component | Current | Target |
|-----------|---------|--------|
| Models | ✅ High | 80%+ |
| Database | ✅ High | 75%+ |
| API Handlers | 🔶 Medium | 70%+ |
| Analyzer | ✅ High | 80%+ |

Generate coverage report:
```bash
task test:coverage
open backend/coverage.html
```

## Linting

### Configuration

Linter rules are defined in `backend/.golangci.yml`:

Enabled linters:
- gofmt - Code formatting
- govet - Go vet checks
- errcheck - Error checking
- staticcheck - Static analysis
- ineffassign - Ineffectual assignments
- unused - Unused code
- gosimple - Simplifications
- misspell - Spelling errors
- revive - Fast linter
- gocyclo - Cyclomatic complexity

### Running Linter

```bash
task lint

# Or manually
cd backend
golangci-lint run ./...
```

## Best Practices

1. **Write tests first** - TDD approach when possible
2. **Use table-driven tests** - For multiple test cases
3. **Mock external dependencies** - Database, HTTP clients, etc.
4. **Test error cases** - Not just happy paths
5. **Use meaningful test names** - Describe what is being tested
6. **Keep tests fast** - Use mocks instead of real databases
7. **Run race detector** - Catch concurrency issues
8. **Maintain coverage** - Aim for 70%+ coverage

## Adding New Tests

### 1. Create test file

```bash
# Test file name should match source file
# my_feature.go -> my_feature_test.go
```

### 2. Import required packages

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### 3. Write test function

```go
func TestMyFeature(t *testing.T) {
    // Arrange
    input := "test"

    // Act
    result := MyFeature(input)

    // Assert
    assert.Equal(t, "expected", result)
}
```

### 4. Run tests

```bash
task test

# Or run specific test
go test -run TestMyFeature ./path/to/package
```

## Troubleshooting

### Tests fail on CI but pass locally

- Check Go version matches CI (1.21+)
- Run with race detector: `go test -race ./...`
- Check for timing-sensitive tests

### Coverage too low

```bash
# Generate detailed coverage report
task test:coverage

# Find uncovered lines
go tool cover -html=coverage.out
```

### Mock expectations not met

```bash
# Use verbose output
go test -v ./...

# Check mock setup matches actual calls
mockRepo.AssertExpectations(t)
```

## Resources

- [Testing in Go](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Sqlmock Guide](https://github.com/DATA-DOG/go-sqlmock)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)
