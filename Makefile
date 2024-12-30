# Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build

# Targets
BINARY_NAME=wake
BUILD_DIR=./build/binaries
PACKAGES_DIR=./packages

# Default target
all: clean build

build: clean
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-$(shell go env GOARCH) .

# Build the project
amd64: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-amd64 .

# Cross-compile for Raspberry Pi Zero W
rpi-zero: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-rpi-zero .

# Cross-compile for Raspberry Pi 5
rpi-5: clean
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-rpi-5 .

# package all three architectures using fpm, so install fpm however you wish
packages: amd64 rpi-zero
	./package.sh

# Clean the build
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(PACKAGES_DIR)

.PHONY: all build rpi-zero rpi-5 clean