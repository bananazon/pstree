GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
PSTREE_VERSION := 0.9.2

GOOS ?= $(shell uname | tr '[:upper:]' '[:lower:]')
GOARCH ?=$(shell arch)

.PHONY: all build install

all: build install

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: build OS ARCH
build: guard-PSTREE_VERSION mod-tidy clean
	@echo "================================================="
	@echo "Building pstree"
	@echo "=================================================\n"

	@if [ ! -d "bin" ]; then \
		mkdir "bin"; \
	fi
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o "bin/pstree"
	# sleep 2
	# tar -czvf "pstree_${PSTREE_VERSION}_${GOOS}_${GOARCH}.tgz" bin; \

.PHONY: clean
clean:
	@echo "================================================="
	@echo "Cleaning pstree"
	@echo "=================================================\n"
	@if [ -f bin/pstree ]; then \
		rm -f bin/pstree; \
	fi
	@if [ -f pstree ]; then \
		rm -f pstree; \
	fi; \

.PHONY: clean-all
clean-all: clean
	@echo "================================================="
	@echo "Cleaning tarballs"
	@echo "=================================================\n"
	@rm -f *.tgz 2>/dev/null

.PHONY: install
install:
	@echo "================================================="
	@echo "Installing raindrop in ${GOPATH}/bin"
	@echo "================================================="

	GOOS=${GOOS} GOARCH=${GOARCH} go install

#
# Test targets
#
.PHONY: test test-all test-short test-race test-bench test-cover test-build test-clean

test-build:
	@echo "================================================="
	@echo "Building test binary"
	@echo "=================================================\n"
	go build -o pstree.testbin .

test-clean:
	@echo "================================================="
	@echo "Cleaning up test binary"
	@echo "=================================================\n"
	@rm -f pstree.testbin

test: test-build
	@echo "================================================="
	@echo "Running standard tests"
	@echo "=================================================\n"
	go test -v ./...
	@$(MAKE) test-clean

test-all: test-build
	@echo "================================================="
	@echo "Running all tests"
	@echo "=================================================\n"
	go test -v ./...
	go test -race ./...
	go test -bench=. ./...
	go test -coverprofile=coverage.out ./...
	@echo "\nTo view coverage report in browser:\n  go tool cover -html=coverage.out"
	@$(MAKE) test-clean

test-short: test-build
	@echo "================================================="
	@echo "Running short tests (skipping integration tests)"
	@echo "=================================================\n"
	go test -short ./...
	@$(MAKE) test-clean

test-race: test-build
	@echo "================================================="
	@echo "Running tests with race detector"
	@echo "=================================================\n"
	go test -race ./...
	@$(MAKE) test-clean

test-bench: test-build
	@echo "================================================="
	@echo "Running benchmarks"
	@echo "=================================================\n"
	go test -bench=. ./...
	@$(MAKE) test-clean

test-cover: test-build
	@echo "================================================="
	@echo "Generating test coverage report"
	@echo "=================================================\n"
	go test -coverprofile=coverage.out ./...
	@echo "\nTo view coverage report in browser:\n  go tool cover -html=coverage.out"
	@$(MAKE) test-clean

#
# General targets
#
guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi
