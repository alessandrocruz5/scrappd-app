# Storage Service Testing Guide

This document explains how to run tests for the R2 storage implementation.

## Test Structure

The storage service includes three types of tests:

1. **Unit Tests** (`storage_test.go`) - Mock-based tests that don't require R2 connection
2. **Integration Tests** (`storage_integration_test.go`) - Tests that require actual R2 credentials
3. **Benchmarks** - Performance tests for storage operations

## Running Tests

### Unit Tests (No R2 Required)

Unit tests use mocks and don't require actual R2 credentials. They test:
- Configuration validation
- Mock storage implementation
- Interface compliance
- Context cancellation
- File operations logic
- Content type handling
- Presigned URL expiry settings

```bash
# Run all unit tests in the services package
cd backend
go test ./internal/services/ -v

# Run with coverage
go test ./internal/services/ -v -cover

# Run with detailed coverage report
go test ./internal/services/ -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Integration Tests (R2 Required)

Integration tests require actual Cloudflare R2 credentials and will perform real operations against your R2 bucket.

**Warning**: Integration tests will create and delete files in your R2 bucket. Use a test bucket to avoid affecting production data.

#### Prerequisites

1. Create a test R2 bucket in Cloudflare dashboard
2. Generate R2 API tokens
3. Set environment variables

#### Environment Variables

```bash
export STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
export STORAGE_ACCESS_KEY_ID=<your-access-key-id>
export STORAGE_SECRET_ACCESS_KEY=<your-secret-access-key>
export STORAGE_BUCKET_NAME=scrappd-images-test  # Use a test bucket!
export STORAGE_REGION=auto
```

Or create a `.env.test` file:

```env
STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
STORAGE_ACCESS_KEY_ID=<your-access-key-id>
STORAGE_SECRET_ACCESS_KEY=<your-secret-access-key>
STORAGE_BUCKET_NAME=scrappd-images-test
STORAGE_REGION=auto
```

#### Running Integration Tests

```bash
# Run integration tests
cd backend
go test -tags=integration ./internal/services/ -v

# Run specific integration test
go test -tags=integration ./internal/services/ -v -run TestIntegrationR2StorageUpload

# Run with timeout (useful for network issues)
go test -tags=integration ./internal/services/ -v -timeout 5m
```

Integration tests cover:
- File upload with auto-generated keys
- File upload with custom keys
- File download and content verification
- File deletion
- Presigned URL generation
- File existence checking
- Listing files by prefix
- Large file handling (1MB+)
- Concurrent operations
- Complete workflow scenarios

### Benchmarks

Benchmark tests measure performance of storage operations.

```bash
# Run benchmarks
cd backend
go test ./internal/services/ -bench=. -benchmem

# Run specific benchmark
go test ./internal/services/ -bench=BenchmarkMockStorage -benchmem

# Run benchmarks multiple times for accuracy
go test ./internal/services/ -bench=. -benchmem -benchtime=10s -count=5
```

Sample output:
```
BenchmarkMockStorage/Upload-8     5000000    250 ns/op    128 B/op    2 allocs/op
BenchmarkMockStorage/Download-8   10000000   120 ns/op     64 B/op    1 allocs/op
```

## Test Commands Quick Reference

```bash
# All unit tests
make test-unit

# All integration tests (requires R2)
make test-integration

# All tests (unit + integration)
make test-all

# Tests with coverage
make test-coverage

# Benchmarks
make benchmark

# Clean test cache
go clean -testcache
```

## Makefile Targets

Add these targets to your `backend/Makefile`:

```makefile
# Test targets
.PHONY: test-unit test-integration test-all test-coverage benchmark

test-unit:
	@echo "Running unit tests..."
	go test ./internal/services/ -v

test-integration:
	@echo "Running integration tests (requires R2 credentials)..."
	go test -tags=integration ./internal/services/ -v -timeout 5m

test-all: test-unit test-integration

test-coverage:
	@echo "Running tests with coverage..."
	go test ./internal/services/ -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

benchmark:
	@echo "Running benchmarks..."
	go test ./internal/services/ -bench=. -benchmem -benchtime=5s
```

## Continuous Integration

### GitHub Actions Example

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run unit tests
        working-directory: ./backend
        run: go test ./internal/services/ -v -cover

  integration-tests:
    runs-on: ubuntu-latest
    # Only run integration tests on main branch or when labeled
    if: github.ref == 'refs/heads/main' || contains(github.event.pull_request.labels.*.name, 'integration-test')
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Run integration tests
        working-directory: ./backend
        env:
          STORAGE_ENDPOINT: ${{ secrets.R2_ENDPOINT }}
          STORAGE_ACCESS_KEY_ID: ${{ secrets.R2_ACCESS_KEY_ID }}
          STORAGE_SECRET_ACCESS_KEY: ${{ secrets.R2_SECRET_ACCESS_KEY }}
          STORAGE_BUCKET_NAME: ${{ secrets.R2_TEST_BUCKET }}
          STORAGE_REGION: auto
        run: go test -tags=integration ./internal/services/ -v -timeout 5m
```

## Test Coverage Goals

Aim for these coverage targets:

- **Overall**: 80%+ coverage
- **Critical paths**: 90%+ coverage (upload, download, delete)
- **Error handling**: 100% coverage
- **Configuration**: 100% coverage

Check coverage:

```bash
cd backend
go test ./internal/services/ -coverprofile=coverage.out
go tool cover -func=coverage.out

# Or generate HTML report
go tool cover -html=coverage.out
```

## Writing New Tests

### Unit Test Template

```go
func TestNewFeature(t *testing.T) {
    ctx := context.Background()

    t.Run("descriptive test name", func(t *testing.T) {
        mock := &MockStorage{
            UploadFunc: func(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
                // Mock implementation
                return "test-key", nil
            },
        }

        // Test your feature
        result, err := mock.Upload(ctx, nil, "test.txt", "text/plain")

        assert.NoError(t, err)
        assert.Equal(t, "test-key", result)
    })
}
```

### Integration Test Template

```go
// +build integration

func TestIntegrationNewFeature(t *testing.T) {
    cfg := loadTestConfig(t)
    logger := logrus.New()
    logger.SetOutput(os.Stdout)

    storage, err := NewR2Storage(cfg, logger)
    require.NoError(t, err)

    ctx := context.Background()

    t.Run("test with real R2", func(t *testing.T) {
        // Perform real R2 operations
        key, err := storage.Upload(ctx, file, "test.txt", "text/plain")
        require.NoError(t, err)

        // Always cleanup
        defer func() {
            _ = storage.Delete(ctx, key)
        }()
    })
}
```

## Troubleshooting Tests

### Common Issues

#### 1. Integration Tests Skip

```
Skipping integration tests: R2 credentials not configured
```

**Solution**: Set the required environment variables:
```bash
export STORAGE_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
export STORAGE_ACCESS_KEY_ID=<key>
export STORAGE_SECRET_ACCESS_KEY=<secret>
export STORAGE_BUCKET_NAME=<bucket>
```

#### 2. Authentication Errors

```
Error: AccessDenied
```

**Solution**:
- Verify your access key and secret are correct
- Ensure the API token has read/write permissions
- Check that the bucket name is correct

#### 3. Network Timeouts

```
Error: context deadline exceeded
```

**Solution**:
- Increase test timeout: `go test -timeout 10m`
- Check your internet connection
- Verify the endpoint URL is correct

#### 4. Test Cache Issues

```
Tests showing outdated results
```

**Solution**:
```bash
go clean -testcache
go test ./internal/services/ -v
```

## Best Practices

1. **Always cleanup**: Use `defer` to delete test files
2. **Use test buckets**: Never run integration tests against production buckets
3. **Isolate tests**: Each test should be independent
4. **Mock external dependencies**: Use mocks for unit tests
5. **Test error cases**: Don't just test happy paths
6. **Use meaningful names**: Test names should describe what they test
7. **Keep tests fast**: Unit tests should complete in milliseconds
8. **Document complex tests**: Add comments for non-obvious test logic

## Performance Benchmarks

Expected performance (approximate):

- **Upload (small file)**: ~100-500ms
- **Download (small file)**: ~50-200ms
- **Delete**: ~50-100ms
- **GetURL**: ~5-20ms (local operation)
- **Exists**: ~50-100ms
- **List (100 files)**: ~100-300ms

Actual performance depends on:
- Network latency
- File size
- R2 region
- Concurrent operations

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [testify Documentation](https://github.com/stretchr/testify)
- [Cloudflare R2 Documentation](https://developers.cloudflare.com/r2/)
- [AWS SDK Go v2 Testing](https://aws.github.io/aws-sdk-go-v2/docs/unit-testing/)
