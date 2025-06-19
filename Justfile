# Justfile for azure-tui
# Azure TUI - Terminal User Interface for Azure Resource Management

# Variables
APP_NAME := "azure-tui"
BUILD_DIR := "build"
VERSION := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
LDFLAGS := "-X main.version=" + VERSION + " -s -w"

# Default task - build for current platform
default: build

# Show available tasks
help:
	@just --list

# =============================================================================
# BUILD TASKS
# =============================================================================

# Build the main TUI binary for current platform
build:
	@echo "Building {{APP_NAME}} for current platform..."
	go build -ldflags "{{LDFLAGS}}" -o {{APP_NAME}} ./cmd/main.go
	@echo "✅ Build complete: {{APP_NAME}}"

# Build optimized release binary
build-release:
	@echo "Building optimized release binary..."
	go build -ldflags "{{LDFLAGS}}" -trimpath -o {{APP_NAME}} ./cmd/main.go
	@echo "✅ Release build complete: {{APP_NAME}}"

# Build for multiple platforms
build-all: build-linux build-windows build-macos
	@echo "✅ All platform builds complete!"

# Build Linux binary (amd64)
build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p {{BUILD_DIR}}
	GOOS=linux GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o {{BUILD_DIR}}/{{APP_NAME}}-linux-amd64 ./cmd/main.go
	@echo "✅ Linux build complete: {{BUILD_DIR}}/{{APP_NAME}}-linux-amd64"

# Build Windows executable (amd64)
build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p {{BUILD_DIR}}
	GOOS=windows GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o {{BUILD_DIR}}/{{APP_NAME}}-windows-amd64.exe ./cmd/main.go
	@echo "✅ Windows build complete: {{BUILD_DIR}}/{{APP_NAME}}-windows-amd64.exe"

# Build macOS binary (amd64)
build-macos:
	@echo "Building for macOS (amd64)..."
	@mkdir -p {{BUILD_DIR}}
	GOOS=darwin GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o {{BUILD_DIR}}/{{APP_NAME}}-darwin-amd64 ./cmd/main.go
	@echo "✅ macOS build complete: {{BUILD_DIR}}/{{APP_NAME}}-darwin-amd64"

# Build macOS binary (arm64 - Apple Silicon)
build-macos-arm:
	@echo "Building for macOS (arm64)..."
	@mkdir -p {{BUILD_DIR}}
	GOOS=darwin GOARCH=arm64 go build -ldflags "{{LDFLAGS}}" -o {{BUILD_DIR}}/{{APP_NAME}}-darwin-arm64 ./cmd/main.go
	@echo "✅ macOS ARM build complete: {{BUILD_DIR}}/{{APP_NAME}}-darwin-arm64"

# =============================================================================
# TESTING TASKS
# =============================================================================

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -v -bench=. -benchmem ./...

# Test specific package
test-pkg PKG:
	@echo "Testing package: {{PKG}}"
	go test -v ./{{PKG}}

# =============================================================================
# DEVELOPMENT TASKS
# =============================================================================

# Run the TUI interactively
run:
	@echo "Starting Azure TUI..."
	go run ./cmd/main.go

# Run with OpenAI (disable GitHub Copilot)
run-openai:
	@echo "Starting Azure TUI with OpenAI provider..."
	USE_GITHUB_COPILOT=false go run ./cmd/main.go

# Run with debug logging
run-debug:
	@echo "Starting Azure TUI with debug logging..."
	DEBUG=true go run ./cmd/main.go

# Run and watch for changes (requires entr: sudo apt install entr)
watch:
	@echo "Starting file watcher for automatic rebuild..."
	find . -name "*.go" | entr -r just run

# Format all Go code
fmt:
	@echo "Formatting Go code..."
	gofmt -w .
	@echo "✅ Code formatted"

# Format and organize imports
fmt-imports:
	@echo "Formatting and organizing imports..."
	goimports -w .
	@echo "✅ Imports organized"

# Tidy up go.mod and go.sum
tidy:
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "✅ Modules tidied"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "✅ Dependencies downloaded"

# =============================================================================
# QUALITY ASSURANCE
# =============================================================================

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	$(go env GOPATH)/bin/golangci-lint run ./...
	@echo "✅ Linting complete"

# Check for security issues (requires gosec)
security:
	@echo "Running security check..."
	$(go env GOPATH)/bin/gosec ./...
	@echo "✅ Security check complete"

# Run all quality checks
qa: fmt tidy lint test
	@echo "✅ All quality checks passed!"

# Run security and linting checks (not used in CI)
qa-full: fmt tidy lint test security
	@echo "✅ All quality checks including security passed!"

# Check only security and linting (for local development)
check-code: lint security
	@echo "✅ Code quality checks complete!"

# =============================================================================
# UTILITY TASKS
# =============================================================================

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f {{APP_NAME}} {{APP_NAME}}.exe
	rm -rf {{BUILD_DIR}}
	rm -f coverage.out coverage.html
	@echo "✅ Cleanup complete"

# Show build information
info:
	@echo "=== Build Information ==="
	@echo "App Name: {{APP_NAME}}"
	@echo "Version: {{VERSION}}"
	@echo "Go Version: $(go version)"
	@echo "Platform: $(go env GOOS)/$(go env GOARCH)"
	@echo "========================="

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✅ Development tools installed"

# =============================================================================
# DEMO AND TESTING SCRIPTS
# =============================================================================

# Run demo scripts
demo:
	@echo "Running enhanced demo..."
	./demo/demo-enhanced.sh

# Test network functionality
test-network:
	@echo "Testing network functionality..."
	./demo/demo-network.sh

# Test container functionality
test-containers:
	@echo "Testing container functionality..."
	./demo/demo-container-instance.sh

# Test storage functionality
test-storage:
	@echo "Testing storage functionality..."
	./demo/demo-storage-account.sh

# Test Key Vault functionality
test-keyvault:
	@echo "Testing Key Vault functionality..."
	./demo/demo-keyvault.sh

# =============================================================================
# RELEASE TASKS
# =============================================================================

# Test release workflow components
test-release:
	@echo "🧪 Testing release workflow..."
	./scripts/test-release-workflow.sh

# Create a release build with all platforms
release: clean qa build-all
	@echo "Creating release archive..."
	@mkdir -p {{BUILD_DIR}}/release
	@cp README.md {{BUILD_DIR}}/release/
	@cp LICENSE {{BUILD_DIR}}/release/
	@cp -r docs {{BUILD_DIR}}/release/
	@cd {{BUILD_DIR}} && tar -czf {{APP_NAME}}-{{VERSION}}-release.tar.gz release/
	@echo "✅ Release created: {{BUILD_DIR}}/{{APP_NAME}}-{{VERSION}}-release.tar.gz"

# Install binary to local system
install: build-release
	@echo "Installing {{APP_NAME}} to /usr/local/bin..."
	sudo cp {{APP_NAME}} /usr/local/bin/
	sudo chmod +x /usr/local/bin/{{APP_NAME}}
	@echo "✅ {{APP_NAME}} installed successfully!"

# Uninstall binary from local system
uninstall:
	@echo "Uninstalling {{APP_NAME}} from /usr/local/bin..."
	sudo rm -f /usr/local/bin/{{APP_NAME}}
	@echo "✅ {{APP_NAME}} uninstalled successfully!"

# =============================================================================
# RELEASE MANAGEMENT
# =============================================================================

# Create a new release (requires version tag)
create-release VERSION DESCRIPTION="":
	@echo "🚀 Creating release {{VERSION}}..."
	@if [ -z "{{DESCRIPTION}}" ]; then \
		./scripts/create-release.sh "{{VERSION}}"; \
	else \
		./scripts/create-release.sh "{{VERSION}}" "{{DESCRIPTION}}"; \
	fi

# Check if ready for release
check-release:
	@echo "🔍 Checking if repository is ready for release..."
	@echo "Running full quality assurance..."
	just qa-full
	@echo ""
	@echo "🔍 Checking git status..."
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Working directory is not clean. Please commit or stash changes."; \
		git status; \
		exit 1; \
	fi
	@echo "✅ Repository is clean"
	@echo ""
	@echo "🏷️  Recent tags:"
	@git tag --sort=-version:refname | head -5 || echo "No tags found"
	@echo ""
	@echo "✅ Repository is ready for release!"
	@echo ""
	@echo "💡 To create a release, run:"
	@echo "   just create-release v1.0.0 'Release description'"

# List recent releases
list-releases:
	@echo "📋 Recent releases:"
	@git tag --sort=-version:refname | head -10 || echo "No releases found"

# Show release status
release-status:
	@echo "📊 Release Status"
	@echo "================="
	@echo ""
	@echo "🏷️  Latest tag: $$(git describe --tags --abbrev=0 2>/dev/null || echo 'No tags')"
	@echo "📝 Current commit: $$(git rev-parse --short HEAD)"
	@echo "🌿 Current branch: $$(git branch --show-current)"
	@echo ""
	@echo "🔍 Build status:"
	@if just build > /dev/null 2>&1; then \
		echo "✅ Build: Success"; \
	else \
		echo "❌ Build: Failed"; \
	fi
	@if just test > /dev/null 2>&1; then \
		echo "✅ Tests: Pass"; \
	else \
		echo "❌ Tests: Fail"; \
	fi
