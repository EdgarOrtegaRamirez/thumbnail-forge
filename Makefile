.PHONY: build test test-race bench fixtures clean install lint fmt tidy build-all run

# Build the binary
build:
	go build -o thumbnail-forge .

# Run all tests (excludes bench fixtures with .c files)
test:
	go test ./internal/... ./cmd/... -v

# Run tests with race detector
test-race:
	go test -race ./internal/... ./cmd/...

# Run specific handler tests
test-image:
	go test ./internal/handlers/... -run TestImageHandler -v

test-code:
	go test ./internal/handlers/... -run TestCodeHandler -v

test-pdf:
	go test ./internal/handlers/... -run TestPDFHandler -v

test-video:
	go test ./internal/handlers/... -run TestVideoHandler -v

test-audio:
	go test ./internal/handlers/... -run TestAudioHandler -v

test-detect:
	go test ./internal/detect/... -v

test-terminal:
	go test ./internal/terminal/... -v

# Run benchmarks (requires fixtures)
bench:
	go test ./internal/handlers/... -bench=. -benchmem -benchtime=1s -timeout=600s

# Generate benchmark fixtures
fixtures:
	python3 generate_bench_fixtures.py

# Clean build artifacts
clean:
	rm -f thumbnail-forge
	rm -f thumbnail-forge-*
	rm -rf tests/bench/

# Install binary to /usr/local/bin
install:
	sudo mv thumbnail-forge /usr/local/bin/

# Run golangci-lint
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -s -w .

# Tidy modules
tidy:
	go mod tidy

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o thumbnail-forge-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o thumbnail-forge-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o thumbnail-forge-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o thumbnail-forge-windows-amd64.exe .

# Run the tool
run:
	./thumbnail-forge generate tests/fixtures/sample.go --terminal
