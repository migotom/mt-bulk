# Project
MTBULK_PROJECT=cmd/mt-bulk/*
MTBULKRESTAPI_PROJECT=cmd/mt-bulk-rest-api/*

BIN_DIR=bin
BUILD_DIR=build

EXAMPLES=examples
DOCS=docs
README=README.md
VERSION=2.2.0

# Go
GOCMD=go
GOBUILD=$(GOCMD) build -ldflags="-s -w -X main.version=${VERSION}"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build
all: build
build: test build-prepare build-linux-amd64 build-linux-386 build-darwin-amd64 build-win-amd64 clean

# Clean
clean:
	rm -R $(BUILD_DIR)/*

test:
	$(GOCMD) test ./...

build-prepare:
	[ -d $(BUILD_DIR)/mt-bulk/certs ] || mkdir -p $(BUILD_DIR)/mt-bulk/certs
	[ -d $(BUILD_DIR)/mt-bulk/docs ] || mkdir -p $(BUILD_DIR)/mt-bulk/docs
	[ -d $(BUILD_DIR)/mt-bulk/examples ] || mkdir -p $(BUILD_DIR)/mt-bulk/examples
	[ -d $(BIN_DIR) ] || mkdir -p $(BIN_DIR)
	cp -R $(EXAMPLES)/* $(BUILD_DIR)/mt-bulk/examples
	cp -R $(DOCS)/* $(BUILD_DIR)/mt-bulk/docs
	cp $(README) $(BUILD_DIR)/mt-bulk/

# Cross compile
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api $(MTBULKRESTAPI_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.$(VERSION).linux.amd64.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api

build-linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api $(MTBULKRESTAPI_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.$(VERSION).linux.386.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk $(MTBULK_PROJECT)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api $(MTBULKRESTAPI_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.$(VERSION).darwin.amd64.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api

build-win-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk//mt-bulk.exe $(MTBULK_PROJECT)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api.exe $(MTBULKRESTAPI_PROJECT)
	cd $(BUILD_DIR) && zip -r -9 ../$(BIN_DIR)/mt-bulk.$(VERSION).windows.amd64.zip mt-bulk
	rm $(BUILD_DIR)/mt-bulk/mt-bulk.exe
	rm $(BUILD_DIR)/mt-bulk/mt-bulk-rest-api.exe
