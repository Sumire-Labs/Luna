.PHONY: build run clean test deps dev

# Build variables
BINARY_NAME=luna-bot
MAIN_PATH=cmd/bot/main.go

# Build the application
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Development mode with hot reload
dev:
	go run $(MAIN_PATH)

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)