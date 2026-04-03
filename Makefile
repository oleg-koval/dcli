.PHONY: help build build-all clean test lint install

VERSION ?= 0.1.0
BINARY_NAME = dcli
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

help:
	@echo "dcli - Docker Compose & Git CLI"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build for current OS/arch"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  fuzz         - Run fuzzer tests"
	@echo "  install      - Install locally"
	@echo "  clean        - Remove build artifacts"

build:
	go build -v -o bin/$(BINARY_NAME) -ldflags="-X 'github.com/oleg-koval/dcli/cmd.Version=$(VERSION)'" main.go

build-all:
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build -v -o bin/$(BINARY_NAME)-$${platform%/*}-$${platform#*/} \
		-ldflags="-X 'github.com/oleg-koval/dcli/cmd.Version=$(VERSION)'" main.go; \
	done

test:
	go test -v -cover ./...

lint:
	go fmt ./...
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

fuzz:
	go test -fuzz=FuzzCleanCommand ./internal/docker -fuzztime=10s
	go test -fuzz=FuzzGitReset ./internal/git -fuzztime=10s
	go test -fuzz=FuzzConfigYAML ./internal/config -fuzztime=10s

install: build
	cp bin/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	chmod +x $(HOME)/.local/bin/$(BINARY_NAME)

clean:
	rm -rf bin/
	go clean
