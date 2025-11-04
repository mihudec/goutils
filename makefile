.PHONY: all build test clean fmt vet coverage

# Default target
all: fmt test

build:
	@echo "🚀 Building..."
	go build ./...

test:
	@echo "🧪 Running tests..."
	go test -v ./...

fmt:
	@echo "✨ Formatting code..."
	go fmt ./...

vet:
	@echo "🔍 Running govet..."
	go vet ./...

coverage:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	@echo "Coverage written to coverage.out"

clean:
	@echo "🧹 Cleaning..."
	go clean
