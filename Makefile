# Paths
BINARY_NAME=bin/splitExpense
SQLC_CONFIG=sqlc.json

# Commands
GO=go
SQLC=sqlc
AIR=air

.PHONY: all build run dev sqlc clean tidy fmt

all: 
	install-sqlc
	sqlc
	build


# Build the Go binary
build:
	@echo ">> Building the binary..."
	$(GO) build -o $(BINARY_NAME)

# Run the compiled binary
run: build
	@echo ">> Running the app..."
	./$(BINARY_NAME)

# Start live reload with Air
dev:
	@echo ">> Starting Air (live reload)..."
	$(AIR)

install-sqlc: 
	@echo ">> Installing sqlc..."
	$(GO) install github.com/kyleconroy/sqlc/cmd/sqlc@latest

# Generate code using sqlc
sqlc:
	@echo ">> Running sqlc code generation..."
	$(SQLC) generate --file $(SQLC_CONFIG)

# Clean up binaries and temp files
clean:
	@echo ">> Cleaning up..."
	rm -rf $(BINARY_NAME) tmp/

# Go mod tidy
tidy:
	@echo ">> Tidying modules..."
	$(GO) mod tidy

# Format Go code
fmt:
	@echo ">> Formatting code..."
	$(GO) fmt ./...

test:
	@echo ">> Running Integration tests..."
	go clean -testcache && go test -v ./...
