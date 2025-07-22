.PHONY: help test test-coverage test-cross-platform build clean

help:
	@echo "Available commands:"
	@echo "  make test              - Run all Go tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-cross-platform - Run cross-platform tests with Solidity"
	@echo "  make build             - Build all Go code and examples"
	@echo "  make clean             - Clean generated files"
	@echo "  make help              - Show this help message"

test:
	@echo "=== Running Go Tests ==="
	cd go && go test ./tests ./tests/benchmark -v

test-coverage:
	@echo "=== Running Go Tests with Coverage ==="
	cd go && go test -coverprofile=coverage.out ./tests ./tests/benchmark
	@echo "=== Generating Coverage Report ==="
	cd go && go tool cover -html=coverage.out -o coverage.html
	@echo "=== Checking Coverage with Exclusions ==="
	cd go && $(shell go env GOPATH)/bin/go-test-coverage --config=.testcoverage.yml || echo "Note: go-test-coverage not found, install with: go install github.com/vladopajic/go-test-coverage/v2@latest"
	@echo "✅ Coverage report generated: go/coverage.html"

test-cross-platform:
	@echo "=== Cross-Platform SMT Test ==="
	@echo "1. Generating fresh proof data with Go..."
	cd go && go run cmd/generate_test_data.go
	@echo "2. Running Solidity test (reading Go-generated JSON)..."
	cd contracts && forge test --match-test "testGoGeneratedProofs" -vv --ffi
	@echo "✅ Cross-platform verification working!"

build:
	@echo "=== Building Go Code ==="
	cd go && go build ./...
	@echo "=== Building Examples ==="
	cd go/examples/basic && go build -o basic .
	cd go/examples/advanced && go build -o advanced .
	cd go/examples/integration && go build -o integration .
	cd go/examples/sequential && go build -o sequential .
	@echo "✅ Build completed successfully"

clean:
	@echo "=== Cleaning ==="
	rm -f contracts/test_data.json
	rm -f go/coverage.out go/coverage.html
	rm -f go/examples/*/basic go/examples/*/advanced go/examples/*/integration go/examples/*/sequential