.PHONY: build test lint run clean install

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
GO_VERSION ?= $(shell go version | awk '{print $$3}')

# 构建目标
BINARY_NAME = trunk
BUILD_DIR = build
MAIN_PATH = main.go

# ldflags
LDFLAGS = -ldflags "\
	-X 'github.com/spelens-gud/trunk/internal.Version=$(VERSION)' \
	-X 'github.com/spelens-gud/trunk/internal.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/spelens-gud/trunk/internal.BuildDate=$(BUILD_DATE)' \
	-X 'github.com/spelens-gud/trunk/internal.GoVersion=$(GO_VERSION)'"

build:
	@echo "构建 $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

run:
	go run $(MAIN_PATH)

install:
	go install $(LDFLAGS) $(MAIN_PATH)

clean:
	rm -rf $(BUILD_DIR)/
