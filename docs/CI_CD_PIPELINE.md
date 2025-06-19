# Azure TUI CI/CD Pipeline Documentation

## Overview

This document describes the comprehensive CI/CD pipeline for the Azure TUI project, with special focus on search functionality testing and multi-platform deployment.

## Pipeline Structure

### 🧪 Test and Quality Checks Job

**Purpose**: Ensures code quality, runs tests, and validates search functionality
**Triggers**: Push to main/develop, Pull Requests

**Steps**:
1. **Code Formatting**: Validates Go code formatting with `gofmt`
2. **Dependency Management**: Ensures `go.mod` and `go.sum` are clean
3. **Unit Tests**: Runs all package tests including search functionality
4. **Coverage Analysis**: Generates coverage reports and uploads to Codecov
5. **Linting**: Runs `golangci-lint` for code quality
6. **Security Scanning**: Uses `gosec` for security vulnerabilities
7. **Search Integration Tests**: Validates search functionality integration
8. **Search Verification**: Confirms search symbols in compiled binary

**Search-Specific Tests**:
- Tests search engine core functionality
- Validates advanced search syntax parsing
- Verifies wildcard matching capabilities
- Tests search suggestions system
- Checks relevance scoring algorithm

### 🏗️ Build Binaries Job

**Purpose**: Creates multi-platform binaries with embedded search functionality
**Dependencies**: Requires successful test job

**Platforms**:
- Linux AMD64
- Linux ARM64  
- Windows AMD64
- macOS AMD64
- macOS ARM64

**Features**:
- Version embedding from Git tags
- Binary size optimization with `-ldflags "-s -w"`
- Checksum generation for security verification
- Artifact upload for release distribution

### 🔍 Test Search Functionality Job

**Purpose**: Dedicated testing of search capabilities
**Dependencies**: Requires successful test job

**Search Tests**:
1. **Package Compilation**: `go build ./internal/search/...`
2. **Unit Tests**: All search-related test cases
3. **Engine Tests**: Core search engine functionality
4. **Syntax Tests**: Advanced search query parsing
5. **Wildcard Tests**: Pattern matching with `*` and `?`
6. **Suggestions Tests**: Real-time search suggestions
7. **Integration Tests**: Search integration in main application
8. **Performance Benchmarks**: Search speed and memory usage
9. **Memory Leak Tests**: Long-running search operations

**Verification Steps**:
- Binary symbol verification for search integration
- Mock data testing for search algorithms
- Performance benchmarking for large datasets

### 🧩 Integration Tests Job

**Purpose**: End-to-end testing with real-world scenarios
**Triggers**: Push to main branch only

**Tests**:
- Binary execution verification
- Command-line interface testing
- Mock Azure integration testing
- Search functionality in realistic scenarios

### 📦 Release Job

**Purpose**: Automated release creation and distribution
**Triggers**: GitHub release events only

**Release Assets**:
- Multi-platform binaries
- Checksums file for verification
- Documentation and license files
- Comprehensive release notes with search features

**Release Features**:
- Automatic version tagging
- Multi-platform binary distribution
- Security checksum verification
- Feature highlighting including search capabilities

### 📢 Notification Job

**Purpose**: Status reporting and failure notifications
**Always Runs**: Regardless of other job outcomes

## Search Functionality Testing

### Core Search Features Tested

1. **Basic Search**:
   - Text matching across resource names
   - Case-insensitive matching
   - Partial string matching

2. **Advanced Search Syntax**:
   - `type:vm` - Filter by resource type
   - `location:eastus` - Filter by location
   - `tag:env=prod` - Filter by tags
   - `rg:mygroup` - Filter by resource group

3. **Wildcard Matching**:
   - `vm*` - Prefix matching
   - `*prod*` - Contains matching
   - `test?` - Single character wildcard
   - Complex patterns: `web-*-prod`

4. **Real-time Features**:
   - Search suggestions as you type
   - Instant result filtering
   - Relevance-based ranking
   - Performance optimized for large datasets

### Test Coverage Areas

- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component functionality
- **Performance Tests**: Speed and memory benchmarks
- **Edge Cases**: Empty results, special characters, large datasets
- **UI Integration**: Keyboard navigation, visual feedback

### Performance Benchmarks

The CI pipeline includes performance benchmarks for:
- Search query parsing speed
- Result filtering performance
- Memory usage during large searches
- Real-time suggestion generation speed

## Quality Gates

### Required Checks

All these checks must pass before code can be merged:

1. ✅ **Code Formatting**: Must be properly formatted
2. ✅ **Unit Tests**: All tests must pass (including search tests)
3. ✅ **Linting**: No linting errors allowed
4. ✅ **Security**: No security vulnerabilities
5. ✅ **Search Integration**: Search functionality must be properly integrated
6. ✅ **Build Success**: All platform builds must succeed

### Optional Checks

These provide additional insights but don't block merges:

- 📊 **Coverage Reports**: Track test coverage trends
- 🏃 **Performance Benchmarks**: Monitor search performance
- 🔍 **Search Feature Tests**: Comprehensive search functionality validation

## Environment Variables

### Required Variables

- `GO_VERSION`: Go version for builds (currently 1.21)
- `APP_NAME`: Application name for artifacts (azure-tui)

### Optional Variables

- `CODECOV_TOKEN`: For coverage reporting (set in repository secrets)
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

## Artifacts

### Build Artifacts

Each platform build produces:
- Compiled binary (`azure-tui-{os}-{arch}`)
- SHA256 checksum file
- Retained for 30 days

### Test Artifacts

- Coverage reports (HTML and raw data)
- Search performance benchmarks
- Test logs and results

### Release Artifacts

- All platform binaries
- Combined checksums file
- Documentation files
- Release archive

## Usage Examples

### Local Development

```bash
# Install dependencies
just deps

# Run all tests including search
just test

# Run only search tests
just test-pkg search

# Build with search functionality
just build

# Test search functionality specifically
./test_search_functionality.sh
```

### CI Debugging

```bash
# Run the same commands as CI locally
go test -v ./internal/search/...
go test -bench=. -benchmem ./internal/search/...
golangci-lint run
gosec ./...
```

## Search Functionality Verification

### Manual Testing Steps

1. Build and run the application: `just build && ./azure-tui`
2. Press `/` to enter search mode
3. Test basic search: type `vm` and press Enter
4. Test advanced syntax: `type:storage location:eastus`
5. Test wildcards: `web-*-prod`
6. Verify suggestions appear as you type
7. Use ↑/↓ to navigate results
8. Press Escape to exit search mode

### Automated Testing

The CI pipeline automatically:
- Compiles search package
- Runs all search unit tests
- Verifies search integration
- Benchmarks search performance
- Checks for memory leaks
- Validates search symbols in binary

## Security Considerations

### Build Security

- All dependencies are verified with checksums
- Security scanning with `gosec`
- No hardcoded credentials or secrets
- Minimal attack surface with optimized binaries

### Release Security

- SHA256 checksums for all binaries
- Signed releases (when configured)
- Secure artifact storage
- Verification instructions in release notes

## Monitoring and Alerts

### Success Indicators

- ✅ All tests pass including search functionality
- ✅ All platform builds succeed
- ✅ Coverage maintains or improves
- ✅ No security vulnerabilities detected
- ✅ Search performance within acceptable limits

### Failure Scenarios

- ❌ Test failures (including search tests)
- ❌ Build failures on any platform
- ❌ Linting or formatting issues
- ❌ Security vulnerabilities detected
- ❌ Search functionality regression

### Notification Channels

- GitHub commit status checks
- Pull request comments
- Action summary pages
- Email notifications (when configured)

## Future Enhancements

### Planned Improvements

1. **Enhanced Search Testing**:
   - Integration with real Azure API responses
   - Load testing with thousands of resources
   - User experience testing automation

2. **Advanced CI Features**:
   - Parallel test execution for faster feedback
   - Dependency vulnerability scanning
   - Container image building and scanning

3. **Release Automation**:
   - Automatic changelog generation
   - Package manager distribution (brew, apt, etc.)
   - Docker image releases

4. **Monitoring Integration**:
   - Performance regression detection
   - Automated performance baseline updates
   - Search analytics and usage metrics

## Troubleshooting

### Common Issues

1. **Search Tests Failing**:
   ```bash
   # Check search package compilation
   go build ./internal/search/...
   
   # Run verbose tests
   go test -v ./internal/search/...
   ```

2. **Build Failures**:
   ```bash
   # Check dependencies
   go mod tidy
   go mod verify
   
   # Test local build
   just build
   ```

3. **Integration Issues**:
   ```bash
   # Verify search integration
   strings ./azure-tui | grep -i search
   
   # Test with mock data
   go run ./cmd/main.go --test-mode
   ```

### Support Resources

- [GitHub Issues](https://github.com/your-repo/azure-tui/issues)
- [Project Documentation](./docs/)
- [Search Implementation Guide](./docs/SEARCH_IMPLEMENTATION_COMPLETE.md)
- [Contributing Guidelines](./CONTRIBUTING.md)
