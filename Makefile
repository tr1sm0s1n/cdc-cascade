ifneq (,$(wildcard ./.env))
    include .env
    export
endif

COMPOSE := docker compose

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
	@docker exec -it tda-postgres psql -d $(DB_NAME) -U $(DB_USER) -W
