## -----------------------------------------------------------------------------
## lazytask - Developer Makefile
## -----------------------------------------------------------------------------
## Usage:
##   make help
##   make run
##   make migrate seed
##   make check
##
## Notes:
## - All paths and commands are overridable at invocation time:
##     make run DB_PATH=/tmp/lazytask/app.db
## - This Makefile intentionally exposes granular targets for daily workflows.
## -----------------------------------------------------------------------------

SHELL := /bin/bash

# Avoid name collision with files named like target names.
.PHONY: help info env print-%
.PHONY: tools-check tools-install deps tidy
.PHONY: fmt vet test test-race lint check ci
.PHONY: build run run-dev run-db
.PHONY: migrate migrate-up migrate-status migrate-down migrate-reset
.PHONY: goose-create
.PHONY: seed bootstrap
.PHONY: sqlc-generate sqlc-validate
.PHONY: clean clean-build-cache clean-test-cache clean-mod-cache

## -----------------------------------------------------------------------------
## Core variables (override with `make <target> VAR=value`)
## -----------------------------------------------------------------------------
APP_NAME            ?= lazytask
MAIN_PKG            ?= ./cmd/app
OUT_DIR             ?= ./bin
BIN_PATH            ?= $(OUT_DIR)/$(APP_NAME)
DB_PATH             ?= $(HOME)/.config/$(APP_NAME)/app.db
DEV_DB_PATH         ?= /tmp/$(APP_NAME)/app.db

GO                  ?= go
GOFLAGS             ?=
GOCACHE_DIR         ?= /tmp/go-build

GOOSE_PKG           ?= github.com/pressly/goose/v3/cmd/goose
GOOSE_MIG_DIR       ?= internal/infrastructure/db/migrations
GOOSE_DRIVER        ?= sqlite3

SQLC_CONFIG         ?= sqlc.yaml

## Tool binaries (used by optional local installs)
GOBIN               ?= $(HOME)/go/bin
GOOSE_BIN           ?= $(GOBIN)/goose
SQLC_BIN            ?= $(GOBIN)/sqlc

## Extra flags pass-through (example: make run ARGS="--seed")
ARGS                ?=

## -----------------------------------------------------------------------------
## Meta targets
## -----------------------------------------------------------------------------

help: ## Show available targets with descriptions
	@echo ""
	@echo "lazytask Makefile"
	@echo "--------------"
	@echo "Key variables:"
	@echo "  APP_NAME=$(APP_NAME)"
	@echo "  DB_PATH=$(DB_PATH)"
	@echo "  DEV_DB_PATH=$(DEV_DB_PATH)"
	@echo "  BIN_PATH=$(BIN_PATH)"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_.-]+:.*## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make run"
	@echo "  make migrate seed"
	@echo "  make run DB_PATH=/tmp/lazytask.db"
	@echo "  make goose-create NAME=add_new_table"
	@echo ""

info: ## Print resolved Make variables for debugging
	@echo "APP_NAME=$(APP_NAME)"
	@echo "MAIN_PKG=$(MAIN_PKG)"
	@echo "OUT_DIR=$(OUT_DIR)"
	@echo "BIN_PATH=$(BIN_PATH)"
	@echo "DB_PATH=$(DB_PATH)"
	@echo "DEV_DB_PATH=$(DEV_DB_PATH)"
	@echo "GO=$(GO)"
	@echo "GOFLAGS=$(GOFLAGS)"
	@echo "GOCACHE_DIR=$(GOCACHE_DIR)"
	@echo "GOOSE_PKG=$(GOOSE_PKG)"
	@echo "GOOSE_MIG_DIR=$(GOOSE_MIG_DIR)"
	@echo "GOOSE_DRIVER=$(GOOSE_DRIVER)"
	@echo "SQLC_CONFIG=$(SQLC_CONFIG)"
	@echo "GOBIN=$(GOBIN)"
	@echo "GOOSE_BIN=$(GOOSE_BIN)"
	@echo "SQLC_BIN=$(SQLC_BIN)"

env: ## Print Go environment values used by this project
	@$(GO) env GOPATH GOMOD GOCACHE GOOS GOARCH

print-%: ## Print any variable value, e.g. `make print-DB_PATH`
	@echo '$*=$($*)'

## -----------------------------------------------------------------------------
## Dependencies and toolchain
## -----------------------------------------------------------------------------

deps: ## Download module dependencies
	@GOCACHE=$(GOCACHE_DIR) $(GO) mod download

tidy: ## Tidy go.mod/go.sum
	@GOCACHE=$(GOCACHE_DIR) $(GO) mod tidy

tools-check: ## Check if goose and sqlc binaries exist in GOBIN
	@test -x "$(GOOSE_BIN)" && echo "goose: OK ($(GOOSE_BIN))" || echo "goose: MISSING ($(GOOSE_BIN))"
	@test -x "$(SQLC_BIN)" && echo "sqlc : OK ($(SQLC_BIN))" || echo "sqlc : MISSING ($(SQLC_BIN))"

tools-install: ## Install goose and sqlc CLI tools into GOBIN
	@GOBIN=$(GOBIN) GOCACHE=$(GOCACHE_DIR) $(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	@GOBIN=$(GOBIN) GOCACHE=$(GOCACHE_DIR) $(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installed tools to $(GOBIN)"

## -----------------------------------------------------------------------------
## Quality gates
## -----------------------------------------------------------------------------

fmt: ## Format all Go files
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -n "$$files" ]; then \
		gofmt -w $$files; \
	fi

vet: ## Run go vet on all packages
	@GOCACHE=$(GOCACHE_DIR) $(GO) vet ./...

test: ## Run tests for all packages
	@GOCACHE=$(GOCACHE_DIR) $(GO) test ./...

test-race: ## Run tests with race detector
	@GOCACHE=$(GOCACHE_DIR) $(GO) test -race ./...

lint: vet ## Alias for vet (extend later with golangci-lint if desired)
	@true

check: fmt vet test build ## Full local validation pipeline
	@echo "check: OK"

ci: tidy check ## CI-like validation pipeline
	@echo "ci: OK"

## -----------------------------------------------------------------------------
## Build and run
## -----------------------------------------------------------------------------

build: ## Build binary into ./bin
	@mkdir -p "$(OUT_DIR)"
	@GOCACHE=$(GOCACHE_DIR) $(GO) build $(GOFLAGS) -o "$(BIN_PATH)" "$(MAIN_PKG)"
	@echo "Built $(BIN_PATH)"

run: ## Run app using DB_PATH
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(MAIN_PKG)" --db-path "$(DB_PATH)" $(ARGS)

run-dev: ## Run app with temporary dev database path
	@mkdir -p "$$(dirname "$(DEV_DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(MAIN_PKG)" --db-path "$(DEV_DB_PATH)" $(ARGS)

run-db: ## Run built binary with DB_PATH (requires make build first)
	@test -x "$(BIN_PATH)" || (echo "Binary not found. Run: make build" && exit 1)
	@"$(BIN_PATH)" --db-path "$(DB_PATH)" $(ARGS)

## -----------------------------------------------------------------------------
## Database and migration operations
## -----------------------------------------------------------------------------

migrate: migrate-up ## Alias for migrate-up
	@true

migrate-up: ## Run embedded app migrations against DB_PATH
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(MAIN_PKG)" --migrate --db-path "$(DB_PATH)"

migrate-status: ## Show migration status using goose CLI package runner
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(GOOSE_PKG)" -dir "$(GOOSE_MIG_DIR)" "$(GOOSE_DRIVER)" "$(DB_PATH)" status

migrate-down: ## Roll back one migration using goose CLI package runner
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(GOOSE_PKG)" -dir "$(GOOSE_MIG_DIR)" "$(GOOSE_DRIVER)" "$(DB_PATH)" down

migrate-reset: ## Roll back all migrations using goose CLI package runner
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(GOOSE_PKG)" -dir "$(GOOSE_MIG_DIR)" "$(GOOSE_DRIVER)" "$(DB_PATH)" reset

goose-create: ## Create a new SQL migration file (usage: make goose-create NAME=add_feature)
	@test -n "$(NAME)" || (echo "NAME is required. Example: make goose-create NAME=add_tasks_index" && exit 1)
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(GOOSE_PKG)" -dir "$(GOOSE_MIG_DIR)" create "$(NAME)" sql

seed: ## Seed sample data using DB_PATH
	@mkdir -p "$$(dirname "$(DB_PATH)")"
	@GOCACHE=$(GOCACHE_DIR) $(GO) run "$(MAIN_PKG)" --seed --db-path "$(DB_PATH)"

bootstrap: migrate-up seed ## Run migrations and seed sample data
	@echo "bootstrap: complete"

## -----------------------------------------------------------------------------
## SQLC
## -----------------------------------------------------------------------------

sqlc-generate: ## Generate code from sqlc.yaml
	@sqlc generate -f "$(SQLC_CONFIG)"

sqlc-validate: ## Validate sqlc configuration and SQL
	@sqlc vet -f "$(SQLC_CONFIG)"

## -----------------------------------------------------------------------------
## Cleanup
## -----------------------------------------------------------------------------

clean: ## Remove local build artifacts
	@rm -rf "$(OUT_DIR)"
	@echo "Removed $(OUT_DIR)"

clean-build-cache: ## Remove Go build cache used by this Makefile
	@rm -rf "$(GOCACHE_DIR)"
	@echo "Removed $(GOCACHE_DIR)"

clean-test-cache: ## Clean Go test cache
	@GOCACHE=$(GOCACHE_DIR) $(GO) clean -testcache
	@echo "Test cache cleaned"

clean-mod-cache: ## Clean module download cache (can be expensive to refill)
	@GOCACHE=$(GOCACHE_DIR) $(GO) clean -modcache
	@echo "Module cache cleaned"

