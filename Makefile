# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le s390x

UI_PATH = web/ui

GOLANGCI_LINT_OPTS ?= --timeout 4m

include Makefile.common

DOCKER_IMAGE_NAME ?= argus

.PHONY: go-build
go-build: commit-prereqs common-build

.PHONY: go-test
go-test:
	go test --tagsunit,integration ./...

.PHONY: go-test-coverage
go-test-coverage:
	go test ./...  -coverpkg=./... -coverprofile ./coverage.out --tags unit,integration
	go tool cover -func ./coverage.out

.PHONY: web-install
web-install:
	cd $(UI_PATH) && { npx --yes update-browserslist-db@latest || true; } && npm install

.PHONY: web-build
web-build:
	cd $(UI_PATH) && npm run build

.PHONY: web-test
web-test:
	cd $(UI_PATH) && npm run test:coverage

.PHONY: web-lint
web-lint:
	cd $(UI_PATH) && npm run lint

.PHONY: web
web: web-install web-build

.PHONY: build
build: web common-build

.PHONY: build-all
build-all: web-build compress-web build-darwin build-freebsd build-linux build-openbsd build-windows