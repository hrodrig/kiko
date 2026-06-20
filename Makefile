BINARY      := kiko
MODULE      := github.com/hrodrig/kiko
DIST        := dist
VERSION_RAW ?= $(shell cat VERSION 2>/dev/null | tr -d '\\n\\r')
VERSION     := $(patsubst v%,%,$(VERSION_RAW))
TAG         := v$(VERSION)
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILDDATE   := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BRANCH      := $(shell git symbolic-ref --short HEAD 2>/dev/null || echo unknown)

COVERAGE_MIN  ?= 80
GRYPE_FAIL_ON ?= high
OPENBSD_ARCH  ?= amd64

LDFLAGS := -s -w \
	-X '$(MODULE)/internal/version.Version=$(VERSION)' \
	-X '$(MODULE)/internal/version.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/version.BuildDate=$(BUILDDATE)' \
	-X '$(MODULE)/internal/version.Branch=$(BRANCH)'

.PHONY: help build install test cover cover-check lint lint-fix gocyclo govulncheck grype security
.PHONY: docker-build docker-scan docker-buildx
.PHONY: release-check release snapshot
.PHONY: port-freebsd-sync port-openbsd-sync dist-freebsd dist-openbsd
.PHONY: clean run tools

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for current platform
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/$(BINARY)

install: ## Install to $GOBIN
	go install -trimpath -ldflags "$(LDFLAGS)" ./cmd/$(BINARY)

test: ## Run tests with race detector
	go test -count=1 -race ./...

cover: ## Run tests + coverage report
	go test -count=1 -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1

cover-check: cover ## Run tests + coverage gate
	@go tool cover -func=coverage.out | tail -1 | awk '{print $$NF}' | tr -d '%' | \
		while read pct; do \
			if [ "$$(echo "$$pct < $(COVERAGE_MIN)" | bc -l)" -eq 1 ]; then \
				echo "FAIL: coverage $$pct% < $(COVERAGE_MIN)%"; exit 1; \
			fi; \
			echo "PASS: coverage $$pct% >= $(COVERAGE_MIN)%"; \
		done

lint: check-mapstructure-pin fmt-check vet gocyclo ## Run all linters

lint-fix: ## Auto-fix formatting
	gofmt -s -w .

check-mapstructure-pin: ## Verify go-viper/mapstructure >= v2.4.0 (CVE pin)
	@if grep -q 'go-viper/mapstructure/v2 v2\.4\.' go.mod; then \
		echo "PASS: mapstructure pinned"; \
	else \
		echo "FAIL: mapstructure not pinned to v2.4.0"; \
		exit 1; \
	fi

fmt-check: ## Check gofmt compliance
	@if [ -n "$$(gofmt -s -l .)" ]; then \
		echo "FAIL: unformatted files:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	echo "PASS: gofmt -s"

vet: ## Run go vet
	go vet ./...

gocyclo: ## Check cyclomatic complexity (max 14)
	@which gocyclo >/dev/null 2>&1 || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@if gocyclo -over 14 . | grep .; then \
		echo "FAIL: functions exceed gocyclo limit 14"; \
		exit 1; \
	fi
	echo "PASS: gocyclo <= 14"

govulncheck: ## Check Go vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

grype: ## Directory CVE scan
	@if command -v grype >/dev/null 2>&1; then \
		grype --fail-on $(GRYPE_FAIL_ON) --exclude './dist/**,./$(BINARY)' .; \
	else \
		echo "grype not installed, skipping"; \
	fi

security: govulncheck gocyclo grype ## Run all security checks

tools: ## Install security tooling
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

docker-build: ## Build Docker image (local)
	@DOCKER_BUILDKIT=1 docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILDDATE=$(BUILDDATE) \
		--build-arg BRANCH=$(BRANCH) \
		-t $(BINARY):local .

docker-buildx: ## Build multi-arch Docker image
	@docker buildx build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILDDATE=$(BUILDDATE) \
		--build-arg BRANCH=$(BRANCH) \
		--platform linux/amd64,linux/arm64 \
		-t $(BINARY):local .

docker-scan: docker-build ## Build + Grype image scan
	@which grype >/dev/null 2>&1 || (echo "grype not installed, skipping scan"; exit 0)
	grype --fail-on $(GRYPE_FAIL_ON) $(BINARY):local

release-check: semver-check fmt-check vet cover-check security ## Release gate: all quality checks

release: ## Release via GoReleaser (main branch only)
	@if [ "$(BRANCH)" != "main" ]; then echo "FAIL: releases from main only"; exit 1; fi
	make release-check
	goreleaser release --clean

snapshot: ## GoReleaser snapshot to dist/
	KIKO_SNAPSHOT_VERSION=$(VERSION)-next goreleaser release --snapshot --clean

semver-check: ## Validate VERSION is semver
	@if ! grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$$' VERSION; then \
		echo "FAIL: VERSION must be MAJOR.MINOR.PATCH (got: $$(cat VERSION))"; \
		exit 1; \
	fi
	echo "PASS: VERSION $(VERSION)"

port-freebsd-sync: ## Sync VERSION to FreeBSD port
	@sed -i '' 's/PORTVERSION=.*/PORTVERSION=	$(VERSION)/' contrib/freebsd/Makefile

port-openbsd-sync: ## Sync VERSION to OpenBSD port
	@sed -i '' \
		-e 's/DISTNAME=.*/DISTNAME=	kiko-$(VERSION)/' \
		-e 's/PKGNAME=.*/PKGNAME=	kiko-$(VERSION)/' \
		contrib/openbsd/port/Makefile

dist-freebsd: ## Cross-compile FreeBSD tarball
	@mkdir -p $(DIST)
	GOOS=freebsd GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o /tmp/kiko_freebsd_amd64 ./cmd/kiko
	tar czf $(DIST)/kiko_$(VERSION)_freebsd_amd64.tar.gz -C /tmp kiko_freebsd_amd64

dist-openbsd: ## Cross-compile OpenBSD tarball
	@mkdir -p $(DIST)
	GOOS=openbsd GOARCH=$(OPENBSD_ARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o /tmp/kiko_openbsd_$(OPENBSD_ARCH) ./cmd/kiko
	tar czf $(DIST)/kiko_$(VERSION)_openbsd_$(OPENBSD_ARCH).tar.gz -C /tmp kiko_openbsd_$(OPENBSD_ARCH)

.PHONY: e2e

e2e: ## Docker Compose E2E (kiko + PostgreSQL + curl)
	sh testing/e2e.sh

run: ## Build and run locally
	go run -ldflags "$(LDFLAGS)" ./cmd/kiko serve

clean: ## Remove build artifacts
	rm -rf $(BINARY) bin/ $(DIST)/ coverage.out
