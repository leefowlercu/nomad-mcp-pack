BINARY_NAME=nomad-mcp-pack
CMD_PATH=github.com/leefowlercu/nomad-mcp-pack
INSTALL_DIR=~/go/bin
VERSION ?= 0.1.0-dev

all: build

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@go build -ldflags="-X '$(CMD_PATH)/cmd.version=$(VERSION)'" -o $(BINARY_NAME) $(CMD_PATH)
	@echo "$(BINARY_NAME) built successfully."

test: 
	@echo "Running tests..."
	@go test ./...
	@echo "Tests finished."

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@echo "Cleaned."

rebuild: clean build
	@echo "Rebuild complete."

install: build
	@echo "Installing $(BINARY_NAME)..."
	@mv ./$(BINARY_NAME) $(INSTALL_DIR)/
	@echo "$(BINARY_NAME) installed successfully."

.PHONY: all build test clean rebuild install
