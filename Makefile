GOCMD=go
GOBUILD=$(GOCMD) build -trimpath
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v ./... -trimpath
GOFUNCTEST=$(GOCMD) test ./functionaltests -v
BIN_NAME=tupi-auth-key
BUILD_DIR=build
AUTH_PLUGIN_BIN=./$(BUILD_DIR)/auth_plugin.so

BIN_PATH=./$(BUILD_DIR)/$(BIN_NAME)
OUTFLAG=-o $(BIN_PATH)
PLUGIN_MODE_FLAG=-buildmode=plugin
AUTH_PLUGIN_FILE=plugin.go

SCRIPTS_DIR=./scripts/


.PHONY: build # - Creates the binary under the build/ directory
build: buildplugin
	$(GOBUILD) $(OUTFLAG)


.PHONY: buildplugin # - Creates the plugin .so binary under the build/ directory
buildplugin:
	$(GOBUILD) -o $(AUTH_PLUGIN_BIN) $(PLUGIN_MODE_FLAG) $(AUTH_PLUGIN_FILE)

.PHONY: test # - Run all tests
test:
	$(GOTEST)

.PHONY: functest # - Run all tests
functest: build
	$(GOFUNCTEST)

.PHONY: setupenv # - Install needed tools for tests/docs
setupenv:
	$(SCRIPTS_DIR)/env.sh setup-env

.PHONY: docs # - Build documentation
docs:
	$(SCRIPTS_DIR)/env.sh build-docs

.PHONY: coverage # - Run all tests and check coverage
coverage: cov

cov:
	$(SCRIPTS_DIR)/check_coverage.sh

.PHONY: run # - Run the program. You can use `make run ARGS="-host :9090 -root=/"`
run:
	$(GOBUILD) $(OUTFLAG)
	$(BIN_PATH) $(ARGS)

.PHONY: clean # - Remove the files created during build
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install # - Copy the binary to the path
install: build
	go install

.PHONY: uninstall # - Remove the binary from path
uninstall:
	go clean -i github.com/jucacrispim/tupi-auth-key


all: build test install

.PHONY: help  # - Show this help text
help:
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1 \2/' | expand -t20
