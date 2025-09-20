BINARY_NAME=nomad-mcp-pack
CMD_PATH=github.com/leefowlercu/nomad-mcp-pack
INSTALL_DIR=~/go/bin
VERSION ?= 0.1.0-dev

all: build

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build -ldflags="-X '$(CMD_PATH)/cmd.version=$(VERSION)'" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Artifact: $(BINARY_NAME) built successfully."

test: 
	@echo "Running all tests..."
	@go test ./...
	@echo "Tests finished."

test-unit:
	@echo "Running unit tests only..."
	@go test -short ./...
	@echo "Unit tests finished."

test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/integration
	@echo "Integration tests finished."

test-integration-verbose:
	@echo "Running integration tests with verbose output..."
	@go test -v -timeout 60s ./tests/integration
	@echo "Integration tests finished."

registry-init:
	@echo "Initializing registry submodule..."
	@git submodule update --init --recursive
	@echo "Registry submodule initialized."

registry-up: registry-init
	@echo "Starting local MCP Registry..."
	@cd tests/integration/registry && docker-compose up -d
	@echo "Registry starting... Check http://localhost:8080/v0/health"

registry-down:
	@echo "Stopping local MCP Registry..."
	@cd tests/integration/registry && docker-compose down
	@echo "Registry stopped."

registry-logs:
	@echo "Showing MCP Registry logs..."
	@cd tests/integration/registry && docker-compose logs -f

registry-update:
	@echo "Updating registry submodule to latest..."
	@cd tests/integration/registry && git fetch --tags && git checkout $$(git describe --tags --abbrev=0)
	@git add tests/integration/registry
	@echo "Registry submodule updated. Commit the change to lock in the new version."

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME) watch.json stdout.log stderr.log
	@rm -rf ./packs/
	@echo "Cleaned."

rebuild: clean build
	@echo "Rebuild complete."

install: build
	@echo "Installing $(BINARY_NAME)..."
	@mv ./$(BINARY_NAME) $(INSTALL_DIR)/
	@echo "$(BINARY_NAME) installed successfully."

.PHONY: all build test test-unit test-integration test-integration-verbose registry-init registry-up registry-down registry-logs registry-update clean rebuild install
