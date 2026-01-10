# Testing Guide

> **TL;DR:** This project maintains comprehensive test coverage across backend (Go) and frontend (JavaScript). Run `go test ./...` for backend, `npm test` for frontend. All tests run automatically in CI/CD.


---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Test Structure](#test-structure)
3. [Running Tests](#running-tests)
4. [Current Coverage](#current-coverage)
5. [Writing Tests](#writing-tests)
6. [Testing Architecture](#testing-architecture)
7. [Dependency Injection & Mocking](#dependency-injection--mocking)
8. [Integration Testing](#integration-testing)
9. [CI/CD Integration](#cicd-integration)
10. [Best Practices](#best-practices)

---

## Quick Start

### Backend Tests (Go)

```bash
# Run all unit tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Run specific service tests
go test -v ./cmd/scraper-service/
go test -v ./cmd/alerts-service/
go test -v ./cmd/archive-service/

# Run integration tests (requires Firestore emulator)
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration -v ./internal/storage/...
```

### Frontend Tests (JavaScript)

```bash
cd dataAnalysis

# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage report
npm run test:coverage
```

---

## Test Structure

### Backend (Go)

```
cmd/
├── alerts-service/
│   ├── main.go
│   └── handler_test.go           # HTTP handler tests
├── archive-service/
│   ├── main.go
│   └── handler_test.go           # HTTP handler tests
└── scraper-service/
    ├── main.go
    └── handler_test.go           # HTTP handler tests

internal/
├── models/
│   ├── alert.go
│   └── alert_test.go             # Data model tests
├── storage/
│   ├── firestore.go
│   ├── firestore_test.go         # Unit tests
│   ├── firestore_integration_test.go  # Integration tests (build tag)
│   ├── police_alerts.go
│   ├── police_alerts_test.go
│   ├── interfaces.go             # Interfaces for DI
│   ├── mock_store.go             # Mock implementations
│   └── mock_gcs.go
└── waze/
    ├── client.go
    ├── client_test.go            # HTTP client tests
    ├── interfaces.go             # Interfaces for DI
    └── mock_fetcher.go           # Mock implementations
```

**Key Principles:**
- Tests co-located with source files (`*_test.go`)
- Integration tests use build tags (`//go:build integration`)
- Mock implementations separate from tests
- Interfaces enable dependency injection

### Frontend (JavaScript)

```
dataAnalysis/
├── public/
│   ├── app.js                    # Main application (DOM glue)
│   └── config.js
├── tests/
│   ├── filters.test.js           # Filter logic tests (33 tests)
│   ├── geojson.test.js           # GeoJSON transformation (27 tests)
│   └── utils.test.js             # Utility functions (30 tests)
└── vitest.config.js              # Test configuration
```

**Architecture Decision:**
- Business logic extracted into testable functions
- DOM/Leaflet integration code in `app.js` intentionally untested
- 90 tests covering all business logic (100% coverage)

---

## Running Tests

### Backend: Unit Tests

```bash
# All packages with race detection
go test -v -race ./...

# Specific package
go test -v ./internal/waze/

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Short mode (skip long-running tests)
go test -short ./...
```

### Backend: Integration Tests

Integration tests require the Firestore emulator:

```bash
# Terminal 1: Start Firestore emulator
gcloud emulators firestore start --host-port=localhost:8080

# Terminal 2: Run integration tests
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration -v ./internal/storage/...

# Run specific integration test
go test -tags=integration -v -run TestIntegration_SavePoliceAlerts_NewAlert ./internal/storage/
```

**What Integration Tests Cover:**
- Firestore CRUD operations
- Query correctness (date ranges, filters)
- Data transformation (WazeAlert → PoliceAlert)
- Concurrent operations
- Large batch operations
- Unicode handling

### Frontend Tests

```bash
cd dataAnalysis

# Run all tests (90 tests)
npm test

# Watch mode (re-run on file changes)
npm run test:watch

# Coverage report
npm run test:coverage

# Run specific test file
npm test -- filters.test.js
```

**What Frontend Tests Cover:**
- Filter logic (subtypes, streets, verified-only)
- GeoJSON coordinate transformation
- Date/time utilities
- Edge cases (null values, empty arrays, invalid inputs)

---

## Current Coverage

### Backend Coverage (Go)

| Service/Package | Unit Coverage | Integration | Total | Status |
|-----------------|--------------|-------------|-------|--------|
| **alerts-service** | 72.2% | +15% | ~87% | ✅ Excellent |
| **scraper-service** | 47.3% | +15% | ~62% | ✅ Good |
| **archive-service** | 67.4% | +15% | ~82% | ✅ Good |
| **internal/waze** | 47.0% | - | 47% | ✅ Good |
| **internal/storage** | 8.4% | +40% | ~48% | ⚠️ Unit tests limited |
| **internal/models** | [no statements] | - | N/A | ✅ Pure data |

**Note:** Storage package has low unit test coverage because most functions interact with Firestore. Integration tests provide comprehensive coverage.

### Frontend Coverage (JavaScript)

| Component | Tests | Coverage | Status |
|-----------|-------|----------|--------|
| **Filter Logic** | 33 | 100% | ✅ Excellent |
| **GeoJSON Transform** | 27 | 100% | ✅ Excellent |
| **Utilities** | 30 | 100% | ✅ Excellent |
| **Total** | **90** | **100%** | ✅ Excellent |

**Architecture Note:** `app.js` shows 0% coverage in reports because it contains DOM/Leaflet integration code (framework glue). All business logic has been extracted and is fully tested.

### CI/CD Requirements

- **Minimum Coverage:** 20% (conservative, allows gradual improvement)
- **Target Coverage:** 60% (aspirational goal with full DI implementation)
- **Enforcement:** Tests must pass before deployment
- **Race Detection:** Enabled on all Go tests

**Coverage Tracking:** [Codecov Dashboard](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP)

---

## Writing Tests

### Backend: Table-Driven Tests (Go Best Practice)

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ProcessString(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr = %v, got error = %v", tt.wantErr, err)
            }
            
            if result != tt.expected {
                t.Errorf("expected %q, got %q", tt.expected, result)
            }
        })
    }
}
```

### Backend: HTTP Handler Testing

```go
func TestHandlerEndpoint(t *testing.T) {
    req := httptest.NewRequest("GET", "/endpoint", nil)
    w := httptest.NewRecorder()
    
    handler := http.HandlerFunc(yourHandler)
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
    
    var response map[string]interface{}
    if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    
    // Assert response fields
}
```

### Frontend: Factory Functions (JavaScript Best Practice)

```javascript
import { describe, it, expect } from 'vitest';

const createMockAlert = (overrides = {}) => ({
    UUID: `alert-${Math.random().toString(36).substr(2, 9)}`,
    Type: 'POLICE',
    Subtype: 'POLICE_VISIBLE',
    Street: 'Test Street',
    NThumbsUpLast: 5,
    ...overrides,
});

describe('Alert Filtering', () => {
    it('filters by subtype', () => {
        const alerts = [
            createMockAlert({ Subtype: 'POLICE_VISIBLE' }),
            createMockAlert({ Subtype: 'POLICE_HIDING' }),
        ];
        
        const filtered = applyFilters(alerts, { 
            subtypes: ['POLICE_VISIBLE'] 
        });
        
        expect(filtered).toHaveLength(1);
        expect(filtered[0].Subtype).toBe('POLICE_VISIBLE');
    });
});
```

---

## Testing Architecture

### Interface-Based Design

The codebase uses interfaces to enable dependency injection and testability:

```go
// internal/storage/interfaces.go
type AlertStore interface {
    SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, scrapeTime time.Time) error
    GetPoliceAlertsByDateRange(ctx context.Context, start, end time.Time) ([]models.PoliceAlert, error)
    GetPoliceAlertsByDatesWithFilters(ctx context.Context, dates []string, subtypes []string, streets []string) ([]models.PoliceAlert, error)
    Close() error
}

// internal/waze/interfaces.go
type AlertFetcher interface {
    GetAlerts(bbox string) (*models.WazeAPIResponse, error)
    GetAlertsMultipleBBoxes(bboxes []string) ([]models.WazeAlert, error)
    GetStats() *models.ScrapingStats
}
```

**Benefits:**
- Enables testing without real Firestore/Waze API
- Allows swapping implementations
- Follows dependency inversion principle
- Reduces coupling between components

### Service Architecture Patterns

| Service | Pattern | Testability | Notes |
|---------|---------|-------------|-------|
| **scraper-service** | DI via function params | ✅ High | Accepts interfaces in handler |
| **alerts-service** | Server struct with interfaces | ✅ High | Injected in main() |
| **archive-service** | Server struct with interfaces | ✅ High | Injected in main() |

**Example: Testable Handler Pattern**

```go
// Production: Inject real implementations
func main() {
    wazeClient := waze.NewClient()
    firestoreClient, _ := storage.NewFirestoreClient(ctx, projectID, collection)
    
    handler := makeScraperHandler(wazeClient, firestoreClient, bboxes)
    http.HandleFunc("/", handler)
}

// Testing: Inject mocks
func TestScraperHandler(t *testing.T) {
    mockFetcher := &waze.MockAlertFetcher{
        GetAlertsMultipleBBoxesFunc: func(bboxes []string) ([]models.WazeAlert, error) {
            return []models.WazeAlert{{UUID: "test"}}, nil
        },
    }
    
    mockStore := &storage.MockAlertStore{
        SavePoliceAlertsFunc: func(ctx context.Context, alerts []models.WazeAlert, t time.Time) error {
            return nil
        },
    }
    
    handler := makeScraperHandler(mockFetcher, mockStore, []string{"bbox"})
    
    // Test handler logic without external dependencies
}
```

---

## Dependency Injection & Mocking

### What is Dependency Injection?

**Simple Definition:** Pass dependencies as parameters instead of creating them inside functions.

**Before (Hard to Test):**
```go
func makeHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        client := waze.NewClient()        // Hard dependency
        db, _ := firestore.NewClient()    // Requires real Firestore
        
        // Cannot test without real services
    }
}
```

**After (Easy to Test):**
```go
func makeHandler(fetcher waze.AlertFetcher, store storage.AlertStore) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        alerts, err := fetcher.GetAlerts(bbox)  // Use injected interface
        store.SavePoliceAlerts(ctx, alerts)      // Use injected interface
        
        // Can test with mocks!
    }
}
```

### Mock Implementations

Mocks are fake implementations that you control completely:

```go
// internal/storage/mock_store.go
type MockAlertStore struct {
    // Function fields allow custom behavior per test
    SavePoliceAlertsFunc func(context.Context, []models.WazeAlert, time.Time) error
    GetPoliceAlertsByDateRangeFunc func(context.Context, time.Time, time.Time) ([]models.PoliceAlert, error)
    
    // CallLog tracks how many times methods were called
    CallLog struct {
        SavePoliceAlertsCalls         int
        GetPoliceAlertsByDateRangeCalls int
        LastSaveAlertsCount            int
    }
}

func (m *MockAlertStore) SavePoliceAlerts(ctx context.Context, alerts []models.WazeAlert, t time.Time) error {
    m.CallLog.SavePoliceAlertsCalls++
    m.CallLog.LastSaveAlertsCount = len(alerts)
    
    if m.SavePoliceAlertsFunc != nil {
        return m.SavePoliceAlertsFunc(ctx, alerts, t)
    }
    return nil
}
```

### Using Mocks in Tests

```go
func TestHandlerSuccess(t *testing.T) {
    // Configure mock to return specific data
    mockFetcher := &waze.MockAlertFetcher{
        GetAlertsMultipleBBoxesFunc: func(bboxes []string) ([]models.WazeAlert, error) {
            return []models.WazeAlert{
                {UUID: "alert-1", Type: "POLICE"},
                {UUID: "alert-2", Type: "POLICE"},
            }, nil
        },
    }
    
    mockStore := &storage.MockAlertStore{
        SavePoliceAlertsFunc: func(ctx context.Context, alerts []models.WazeAlert, t time.Time) error {
            return nil  // Success
        },
    }
    
    handler := makeScraperHandler(mockFetcher, mockStore, []string{"bbox"})
    
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    handler(w, req)
    
    // Verify behavior
    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }
    
    // Verify mock was called correctly
    if mockStore.CallLog.SavePoliceAlertsCalls != 1 {
        t.Errorf("expected 1 call to SavePoliceAlerts, got %d", 
            mockStore.CallLog.SavePoliceAlertsCalls)
    }
}
```

### Testing Error Scenarios

```go
func TestHandlerFetchError(t *testing.T) {
    // Mock returns an error
    mockFetcher := &waze.MockAlertFetcher{
        GetAlertsMultipleBBoxesFunc: func(bboxes []string) ([]models.WazeAlert, error) {
            return nil, errors.New("waze API connection failed")
        },
    }
    
    mockStore := &storage.MockAlertStore{}
    handler := makeScraperHandler(mockFetcher, mockStore, []string{"bbox"})
    
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    handler(w, req)
    
    // Should return error status
    if w.Code != http.StatusInternalServerError {
        t.Errorf("expected 500, got %d", w.Code)
    }
    
    // Store should NOT be called when fetch fails
    if mockStore.CallLog.SavePoliceAlertsCalls != 0 {
        t.Error("SavePoliceAlerts should not be called when fetch fails")
    }
}
```

---

## Integration Testing

### Setup: Firestore Emulator

Integration tests use the Firestore emulator for real database operations without GCP costs:

```bash
# Install (one-time)
gcloud components install cloud-firestore-emulator

# Start emulator
gcloud emulators firestore start --host-port=localhost:8080

# In another terminal, run tests
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration -v ./internal/storage/...
```

### Writing Integration Tests

```go
//go:build integration

package storage

import (
    "context"
    "os"
    "testing"
    "time"
)

func TestIntegration_SaveAndRetrieve(t *testing.T) {
    // Skip if emulator not configured
    if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
        t.Skip("FIRESTORE_EMULATOR_HOST not set")
    }
    
    ctx := context.Background()
    client, err := NewFirestoreClient(ctx, "test-project", "test_alerts")
    if err != nil {
        t.Fatalf("failed to create client: %v", err)
    }
    defer client.Close()
    
    // Test data
    alerts := []models.WazeAlert{
        {UUID: "test-1", Type: "POLICE", Subtype: "POLICE_VISIBLE"},
    }
    
    // Save
    err = client.SavePoliceAlerts(ctx, alerts, time.Now())
    if err != nil {
        t.Fatalf("failed to save: %v", err)
    }
    
    // Retrieve
    start := time.Now().Add(-1 * time.Hour)
    end := time.Now().Add(1 * time.Hour)
    retrieved, err := client.GetPoliceAlertsByDateRange(ctx, start, end)
    if err != nil {
        t.Fatalf("failed to retrieve: %v", err)
    }
    
    // Verify
    if len(retrieved) != 1 {
        t.Errorf("expected 1 alert, got %d", len(retrieved))
    }
    if retrieved[0].UUID != "test-1" {
        t.Errorf("expected UUID 'test-1', got %q", retrieved[0].UUID)
    }
}
```

### Integration Test Coverage

Current integration tests cover:
- ✅ New alert creation
- ✅ Existing alert updates
- ✅ Concurrent update handling
- ✅ Large batch operations (500+ alerts)
- ✅ Date range queries
- ✅ Multi-date queries with filters
- ✅ Unicode street name handling
- ✅ Query performance (<5s for 10,000 alerts)

---

## CI/CD Integration

### GitHub Actions Workflows

Tests run automatically on every push and pull request:

```yaml
# .github/workflows/*-ci-cd.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      # Linting
      - name: Run Go Linter
        uses: golangci/golangci-lint-action@v6
      
      # Unit tests
      - name: Run Unit Tests
        run: go test -v -race -coverprofile=coverage.out ./...
      
      # Coverage enforcement
      - name: Check Coverage Threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 20" | bc -l) )); then
            echo "::error::Coverage below 20%"
            exit 1
          fi
      
      # Upload to Codecov
      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: service-name
  
  integration-test:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      
      # Setup Firestore emulator
      - name: Set up Cloud SDK
        uses: google-github-actions/setup-gcloud@v2
        with:
          install_components: 'cloud-firestore-emulator'
      
      - name: Start Firestore Emulator
        run: |
          gcloud emulators firestore start --host-port=localhost:8080 &
          sleep 10
      
      # Run integration tests
      - name: Run Integration Tests
        env:
          FIRESTORE_EMULATOR_HOST: localhost:8080
        run: go test -tags=integration -v ./internal/storage/...
  
  build-and-deploy:
    needs: [test, integration-test]  # Both must pass
    # ... deployment steps
```

### Frontend CI

```yaml
# .github/workflows/frontend-ci.yml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install dependencies
        run: npm ci
        working-directory: dataAnalysis
      
      - name: Run Tests
        run: npm test
        working-directory: dataAnalysis
      
      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          directory: dataAnalysis/coverage/
          flags: frontend
```

### Deployment Safety

- ✅ **No deployment without passing tests**
- ✅ **Both unit and integration tests required**
- ✅ **Coverage thresholds enforced**
- ✅ **Race conditions detected**
- ✅ **Linting prevents code quality issues**

---

## Best Practices

### General Principles

1. **Test Behavior, Not Implementation**
   - Focus on what the function does, not how
   - Allows refactoring without breaking tests

2. **One Assertion Per Test** (when possible)
   - Makes test failures clearer
   - Use subtests for multiple scenarios

3. **Clear Test Names**
   ```go
   // Good
   func TestGetAlerts_InvalidBBox_ReturnsError(t *testing.T)
   
   // Bad
   func TestGetAlerts1(t *testing.T)
   ```

4. **Arrange-Act-Assert Pattern**
   ```go
   // Arrange: Set up test data
   alerts := []models.WazeAlert{...}
   
   // Act: Execute the function
   result, err := ProcessAlerts(alerts)
   
   // Assert: Verify the outcome
   if err != nil {
       t.Errorf("unexpected error: %v", err)
   }
   ```

### Go-Specific Best Practices

1. **Table-Driven Tests**
   - Scales well for multiple scenarios
   - Reduces code duplication
   - Easy to add new test cases

2. **Use `t.Helper()`** for test utilities
   ```go
   func assertNoError(t *testing.T, err error) {
       t.Helper()  // Reports line number of caller
       if err != nil {
           t.Fatalf("unexpected error: %v", err)
       }
   }
   ```

3. **Clean Up with `defer`**
   ```go
   func TestWithResource(t *testing.T) {
       resource := setupResource()
       defer resource.Close()  // Ensures cleanup
       
       // Test code
   }
   ```

4. **Use `t.Parallel()` for Independent Tests**
   ```go
   func TestIndependent(t *testing.T) {
       t.Parallel()  // Run concurrently with other tests
       // Test code
   }
   ```

### JavaScript-Specific Best Practices

1. **Factory Functions for Test Data**
   - Reduces duplication
   - Makes tests more readable
   - Easy to override specific fields

2. **Descriptive Test Suites**
   ```javascript
   describe('applyFilters', () => {
       describe('when filtering by subtype', () => {
           it('includes only matching alerts', () => {
               // Test code
           });
       });
   });
   ```

3. **Use Vitest's Built-in Matchers**
   ```javascript
   expect(result).toHaveLength(5);
   expect(result[0]).toMatchObject({ UUID: 'alert-1' });
   expect(fn).toHaveBeenCalledWith('arg');
   ```

### Common Pitfalls to Avoid

1. ❌ **Testing Too Much at Once**
   - Keep tests focused on one behavior
   - Use integration tests for multi-component flows

2. ❌ **Fragile Tests (Brittle)**
   - Don't test implementation details
   - Avoid hardcoding exact strings if not relevant

3. ❌ **Not Testing Error Paths**
   - Every error return should be tested
   - Test edge cases and boundary conditions

4. ❌ **Slow Tests**
   - Use mocks instead of real services
   - Run integration tests separately
   - Parallelize when possible

5. ❌ **Flaky Tests**
   - Avoid time-dependent assertions
   - Clean up resources properly
   - Use deterministic test data

---

## Test Types Summary

| Type | Purpose | Speed | When to Use |
|------|---------|-------|-------------|
| **Unit** | Test single function/method | Fast (<1ms) | Always - test all business logic |
| **Integration** | Test component interactions | Medium (1-10ms) | Database, API interactions |
| **E2E** | Test complete user flows | Slow (seconds) | Critical paths only |

### Current Testing Strategy

✅ **Unit Tests:** Comprehensive coverage of business logic  
✅ **Integration Tests:** Database operations with Firestore emulator  
⚪ **E2E Tests:** Not implemented (optional for portfolio project)

This testing strategy provides strong confidence in code quality while maintaining fast feedback cycles and reasonable maintenance burden.

---

## Coverage Goals & Roadmap

### Current Baselines (January 2026)

| Component | Current | Status |
|-----------|---------|--------|
| Backend Unit | 45% | ✅ Good |
| Backend Integration | 30% | ✅ Good |
| Frontend | 100% | ✅ Excellent |
| **Overall** | **~60%** | ✅ Strong |

### 6-Month Targets

| Component | Target | Actions Required |
|-----------|--------|------------------|
| Backend Unit | 55% | Add handler edge case tests |
| Backend Integration | 40% | Expand emulator test scenarios |
| Frontend | 100% | Maintain current level |
| **Overall** | **~70%** | Continue incremental improvements |

### Coverage Philosophy

**Quality Over Quantity:** We prioritize testing business logic and critical paths over achieving arbitrary coverage percentages. Some code (like pure data models and framework glue) doesn't need extensive testing.

**Pragmatic Approach:** Coverage targets balance thoroughness with development velocity. All critical paths are tested; non-critical edge cases are evaluated on a case-by-case basis.

---

## Additional Resources

### Documentation
- [Architecture Document](ARCHITECTURE.md) - System design and component interactions
- [ADR (Architectural Decision Records)](ADR.md) - Key technical decisions
- [Security Documentation](../SECURITY.md) - Security considerations

### External Resources
- [Go Testing Package](https://pkg.go.dev/testing) - Official Go testing docs
- [Vitest Documentation](https://vitest.dev/) - Frontend testing framework
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) - Best practices
- [Google Cloud Firestore Emulator](https://cloud.google.com/firestore/docs/emulator) - Local testing

### Tools
- [golangci-lint](https://golangci-lint.run/) - Go linting
- [Codecov](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP) - Coverage tracking
- [httptest](https://pkg.go.dev/net/http/httptest) - HTTP testing utilities

---

## FAQ

**Q: Why is storage package unit coverage only 8.4%?**  
A: Most storage functions interact directly with Firestore, which requires integration tests. The 40% integration coverage provides comprehensive testing of database operations.

**Q: Why doesn't app.js have tests?**  
A: Business logic has been extracted into separate modules that are fully tested. app.js contains only DOM/Leaflet integration code (framework glue), which provides minimal testing value.

**Q: Should I increase test coverage before adding features?**  
A: No. Add tests for new features as you write them. Retroactively increasing coverage is lower priority than building new functionality.

**Q: Do I need E2E tests?**  
A: Not for this project. The combination of unit tests + integration tests provides strong confidence. E2E tests are optional polish.

**Q: How do I debug failing tests?**  
```bash
# Run single test with verbose output
go test -v -run TestSpecificTest ./package/

# Run with race detector
go test -race -run TestSpecificTest ./package/

# Frontend: Run in watch mode
npm run test:watch

# Frontend: Run specific test
npm test -- filters.test.js
```

---

**Last Updated:** January 10, 2026  
**Maintained by:** Project Team  
**Coverage Dashboard:** [Codecov](https://codecov.io/gh/Lllllllleong/wazePoliceScraperGCP)
