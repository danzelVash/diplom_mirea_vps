BINARY=core-backend
BUILD_DIR=.build
GOCACHE=$(BUILD_DIR)/gocache

.PHONY: build run clean test

build:
	mkdir -p $(BUILD_DIR)
	env GOCACHE=$(abspath $(GOCACHE)) go build -o $(BUILD_DIR)/$(BINARY) ./cmd/core

run:
	mkdir -p data $(BUILD_DIR)
	env GOCACHE=$(abspath $(GOCACHE)) go run ./cmd/core --listen :8080 --data-dir ./data

test:
	mkdir -p $(BUILD_DIR)
	env GOCACHE=$(abspath $(GOCACHE)) go test ./...

clean:
	rm -rf $(BUILD_DIR)
