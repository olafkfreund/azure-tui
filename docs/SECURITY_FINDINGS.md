# Security Findings to Address Later

The following security issues were found by `gosec` and temporarily disabled in CI:

## File Permission Issues (G301, G306)

### Directory Permissions (G301)
- **Files**: `internal/config/config.go:129`, `internal/terraform/terraform.go:65`
- **Issue**: Using `0755` permissions instead of `0750` or less
- **Fix**: Change to `0750` permissions for directories

```go
// Current
os.MkdirAll(configDir, 0755)

// Recommended 
os.MkdirAll(configDir, 0750)
```

### File Permissions (G306)
- **Files**: `internal/terraform/terraform.go:146,164`
- **Issue**: Using `0644` permissions instead of `0600` or less
- **Fix**: Change to `0600` permissions for files

```go
// Current
os.WriteFile(filePath, []byte(content), 0644)

// Recommended
os.WriteFile(filePath, []byte(content), 0600)
```

## File Inclusion Issues (G304)

### Variable File Paths
- **Files**: Various locations where `os.Open()` and `os.ReadFile()` use variable paths
- **Issue**: Potential path traversal vulnerabilities
- **Mitigation**: These are mostly config files in known locations, but should validate paths

## Command Injection Issues (G204)

### Azure CLI Commands
- **Files**: Multiple files in `internal/azure/` packages
- **Issue**: Using `exec.Command()` with user input
- **Status**: These are legitimate Azure CLI calls, but input should be validated

## Re-enabling Security Checks

To re-enable security and linting checks in CI:

1. Uncomment the sections in `.github/workflows/ci.yml`
2. Address the security findings above
3. Use `#nosec` comments for false positives

## Local Testing

Run security checks locally:
```bash
# Check linting and security
just check-code

# Full quality assurance including security
just qa-full
```
