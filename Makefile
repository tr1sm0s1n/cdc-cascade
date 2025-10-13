ifneq (,$(wildcard ./.env))
    include .env
    export
endif

COMPOSE := docker compose
GO ?= go
GOFMT ?= gofmt "-s"
GOFILES := $(shell find . -name "*.go")

.PHONY: help
## help: Display available targets.
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: up
## up: Build and start the containers
up:
	@$(COMPOSE) up --build -d

.PHONY: down
## down: Stop and remove the containers
down:
	@$(COMPOSE) down

## enter: Enter the database.
.PHONY: enter
enter:
	@docker exec -it $(DB_HOST) psql -d $(DB_NAME) -U $(DB_USER) -W

## connect: Create Debezium connector for DB.
.PHONY: connect
connect:
	@docker exec -it $(API_HOST) sh ./opt/scripts/debezium-setup.sh

.PHONY: tidy
## tidy: Clean and tidy dependencies.
tidy:
	@$(GO) mod tidy -v

.PHONY: fmt
## fmt: Format Go files.
fmt:
	@$(GOFMT) -w $(GOFILES)
