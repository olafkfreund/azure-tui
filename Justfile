# Justfile for azure-tui

# Build the main TUI/CLI binary
default:
	go build -o azure-tui ./cmd/main.go

# Run the TUI/CLI interactively
run:
	go run ./cmd/main.go

# Run all tests (if any)
test:
	go test ./...

# Format all Go code
fmt:
	gofmt -w .

# Tidy up go.mod and go.sum
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -f azure-tui

# Lint (requires golangci-lint installed)
lint:
	golangci-lint run ./...
