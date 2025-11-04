.PHONY: all build test clean fmt vet

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

clean:
	@echo "🧹 Cleaning..."
	go clean