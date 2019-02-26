# Go
GOCMD=go
GOBUILD=$(GOCMD) build -ldflags="-s -w"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Project
MTBULK_PROJECT=cmd/mt-bulk/*

BIN_DIR=bin
BUILD_DIR=build

CONFIG_EXAMPLE=mt-bulk.example.toml
HOSTS_EXAMPLE=hosts.example.txt

# Build
all: build
build: build-prepare build-linux-amd64 build-linux-386 build-darwin-amd64 clean

# Clean
clean:
	rm -R $(BUILD_DIR)/*

build-prepare:
	[ -d $(BUILD_DIR)/mt-bulk/certs ] || mkdir -p $(BUILD_DIR)/mt-bulk/certs
	[ -d $(BIN_DIR) ] || mkdir -p $(BIN_DIR)
	cp $(CONFIG_EXAMPLE) $(BUILD_DIR)/mt-bulk/
	cp $(HOSTS_EXAMPLE) $(BUILD_DIR)/mt-bulk/

# Cross compile
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.linux.amd64.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk

build-linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.linux.386.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.darwin.amd64.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk