# Needs to be defined before including Makefile.common to auto-generate targets
DOCKER_ARCHS ?= amd64 arm64 ppc64le s390x

UI_PATH = web/ui

GOLANGCI_LINT_OPTS ?= --timeout 4m

include Makefile.common

DOCKER_IMAGE_NAME ?= argus

.PHONY: go-build
go-build: commit-prereqs common-build

.PHONY: go-test
go-test:
	go test --tags unit,integration ./...

.PHONY: go-test-coverage
go-test-coverage:
	go test ./...  -coverpkg=./... -coverprofile ./coverage.out --tags unit,integration
	go tool cover -func ./coverage.out

.PHONY: web-install
web-install:
	cd "$(UI_PATH)" && { \
		npx --yes update-browserslist-db@latest || true; \
		npm ci; \
	}

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

PLAYWRIGHT_DIR              = $(UI_PATH)/react-app
PLAYWRIGHT_TESTS_DIR        = $(PLAYWRIGHT_DIR)/tests
PLAYWRIGHT_CONFIG_FILE      = $(PLAYWRIGHT_TESTS_DIR)/playwright-test-config.yml
PLAYWRIGHT_DB_FILE          = $(PLAYWRIGHT_TESTS_DIR)/playwright-tests.db
PLAYWRIGHT_PID_FILE         = $(PLAYWRIGHT_TESTS_DIR)/playwright-argus.pid
PLAYWRIGHT_LOG_FILE         = $(PLAYWRIGHT_TESTS_DIR)/playwright-test.log
# Port the Makefile-managed Argus server listens on.
PLAYWRIGHT_PORT            ?= 8080
# URL the Playwright tests run against; defaults to the managed server.
BASE_URL                   ?= http://localhost:$(PLAYWRIGHT_PORT)

# Second, auth-enabled instance for the auth e2e specs.
PLAYWRIGHT_AUTH_CONFIG_SRC  = $(PLAYWRIGHT_TESTS_DIR)/fixtures/auth-config.yml
PLAYWRIGHT_AUTH_CONFIG_FILE = $(PLAYWRIGHT_TESTS_DIR)/playwright-auth-config.yml
PLAYWRIGHT_AUTH_DB_FILE     = $(PLAYWRIGHT_TESTS_DIR)/playwright-auth-tests.db
PLAYWRIGHT_AUTH_PID_FILE    = $(PLAYWRIGHT_TESTS_DIR)/playwright-auth-argus.pid
PLAYWRIGHT_AUTH_LOG_FILE    = $(PLAYWRIGHT_TESTS_DIR)/playwright-auth-test.log
PLAYWRIGHT_AUTH_PORT       ?= 8081
AUTH_BASE_URL              ?= http://localhost:$(PLAYWRIGHT_AUTH_PORT)
# Bootstrap admin password for the auth-enabled instance
PLAYWRIGHT_AUTH_PASSWORD   ?= playwright-admin-pw

# wait_for_argus,<base-url>,<log-file>: poll the healthcheck endpoint until it answers (max 60s).
define wait_for_argus
i=0; until curl -sf $(1)/api/v1/healthcheck >/dev/null; do \
	i=$$((i+1)); \
	if [ $$i -ge 30 ]; then \
		echo "Argus not reachable at $(1) after 60s" >&2; \
		if [ -f $(2) ]; then cat $(2) >&2; fi; \
		exit 1; \
	fi; \
	sleep 2; \
done
endef

# start_argus,<config>,<db>,<port>,<log>,<pid-file>: launch a background Argus instance.
define start_argus
./$(OUTPUT_BINARY) -config.file $(1) -data.database-file $(2) -web.listen-port $(3) -log.level DEBUG > $(4) 2>&1 & echo $$! > $(5)
endef

# stop_argus,<pid-file>: kill the pid-file's process (if running) and remove it.
define stop_argus
if [ -f $(1) ]; then kill $$(cat $(1)) 2>/dev/null || true; rm -f $(1); fi
endef

.PHONY: playwright-tests-setup
playwright-tests-setup:
	@if [ -n "$(FRESH)" ]; then \
		$(MAKE) playwright-tests-teardown; \
		rm -f ./$(OUTPUT_BINARY); \
	else \
		$(call stop_argus,$(PLAYWRIGHT_PID_FILE)); \
	fi
	@$(call stop_argus,$(PLAYWRIGHT_AUTH_PID_FILE))
	cp config.yml.example $(PLAYWRIGHT_CONFIG_FILE)
	cp $(PLAYWRIGHT_AUTH_CONFIG_SRC) $(PLAYWRIGHT_AUTH_CONFIG_FILE)
	rm -f $(PLAYWRIGHT_AUTH_DB_FILE)*
	@if [ ! -x ./$(OUTPUT_BINARY) ] || [ -n "$(FORCE)" ]; then $(MAKE) build; fi
	$(call start_argus,$(PLAYWRIGHT_CONFIG_FILE),$(PLAYWRIGHT_DB_FILE),$(PLAYWRIGHT_PORT),$(PLAYWRIGHT_LOG_FILE),$(PLAYWRIGHT_PID_FILE))
	$(call start_argus,$(PLAYWRIGHT_AUTH_CONFIG_FILE),$(PLAYWRIGHT_AUTH_DB_FILE),$(PLAYWRIGHT_AUTH_PORT),$(PLAYWRIGHT_AUTH_LOG_FILE),$(PLAYWRIGHT_AUTH_PID_FILE))
	@$(call wait_for_argus,http://localhost:$(PLAYWRIGHT_PORT),$(PLAYWRIGHT_LOG_FILE))
	@$(call wait_for_argus,http://localhost:$(PLAYWRIGHT_AUTH_PORT),$(PLAYWRIGHT_AUTH_LOG_FILE))
	@echo "Creating the bootstrap admin via first-run setup..."
	@curl -sf -X POST "http://localhost:$(PLAYWRIGHT_AUTH_PORT)/api/v1/auth/setup" \
		-H "Content-Type: application/json" \
		-d '{"username":"admin","display_name":"Administrator","password":"$(PLAYWRIGHT_AUTH_PASSWORD)"}' \
		> /dev/null

.PHONY: playwright-tests-teardown
playwright-tests-teardown:
	@if [ -f $(PLAYWRIGHT_PID_FILE) ]; then \
		kill $$(cat $(PLAYWRIGHT_PID_FILE)) 2>/dev/null || true; \
		rm -f $(PLAYWRIGHT_PID_FILE); \
	else \
		echo "No playwright Argus instance found running ($(PLAYWRIGHT_PID_FILE) missing)"; \
	fi
	@$(call stop_argus,$(PLAYWRIGHT_AUTH_PID_FILE))
	rm -f $(PLAYWRIGHT_CONFIG_FILE) $(PLAYWRIGHT_DB_FILE)* $(PLAYWRIGHT_LOG_FILE)
	rm -f $(PLAYWRIGHT_AUTH_CONFIG_FILE) $(PLAYWRIGHT_AUTH_DB_FILE)* $(PLAYWRIGHT_AUTH_LOG_FILE)

.PHONY: playwright-tests
playwright-tests:
	@echo "Waiting for Argus to be ready at $(BASE_URL)..."
	@$(call wait_for_argus,$(BASE_URL),$(PLAYWRIGHT_LOG_FILE))
	@echo "Waiting for Argus to be ready at $(AUTH_BASE_URL)..."
	@$(call wait_for_argus,$(AUTH_BASE_URL),$(PLAYWRIGHT_AUTH_LOG_FILE))
	cd $(PLAYWRIGHT_DIR) && \
	BASE_URL=$(BASE_URL) \
	AUTH_BASE_URL=$(AUTH_BASE_URL) \
	PLAYWRIGHT_AUTH_PASSWORD=$(PLAYWRIGHT_AUTH_PASSWORD) \
	PWTEST_CHILD_PROCESS_TIMEOUT=30000 \
	npx playwright test $(if $(HEADED),--headed,)

.PHONY: playwright-full
playwright-full: playwright-tests-setup
	@$(MAKE) playwright-tests; status=$$?; \
		$(MAKE) playwright-tests-teardown; \
		exit $$status
