.PHONY: all app tidy run test

# Default target: build the executable
all: app

# Rule to build the target executable
app:
	go build -o bin/app cmd/main.go

# Clean target: remove the target executable
tidy:
	rm -f bin/*
	go mod tidy
	go fmt ./...

# Run target: build and run the target executable
run:
	go run cmd/main.go

# Test target: run Go tests for the project
test:
	go test ./...
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0 run
