.PHONY: help
.DEFAULT_GOAL := help

ifeq ($(shell uname), Darwin)
	SED_CMD := gsed
else ifeq ($(shell uname -s | cut -c 1-5), Linux)
	SED_CMD := sed
endif

PSQL=PGPASSWORD=postgres psql -U postgres -h localhost -p 5432
PROJECT=miqio-proxy
DB=miqio_lp_proxy
DB_TEST=$(DB)_test
DB_CONN_STRING=postgres://postgres:postgres@localhost:5432/$(DB)?sslmode=disable
DB_TEST_CONN_STRING=postgres://postgres:postgres@localhost:5432/$(DB_TEST)?sslmode=disable

dev: ## Run the server in dev mode
	go run . server

build: ## Build the binary
	CGO_ENABLED=0 go build -o bin/$(PROJECT) .

setup: ## Install and setup dependencies
	brew list golang-migrate &>/dev/null || brew install golang-migrate
	brew list mockery &>/dev/null || brew install mockery
	brew list golangci-lint &>/dev/null || brew install golangci-lint
	brew list diffutils &>/dev/null || brew install diffutils
	brew list gnu-sed &>/dev/null || brew install gnu-sed
	go get -u github.com/spf13/cobra@latest
	go install github.com/spf13/cobra-cli@latest
	go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

db-create: ## Create databases
	@$(PSQL) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB)'" | grep -q 1 || $(PSQL) -c "CREATE DATABASE "$(DB)""
	@$(PSQL) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_TEST)'" | grep -q 1 || $(PSQL) -c "CREATE DATABASE "$(DB_TEST)""

db-migrate: ## Migrate databases
	@count=$$(find ./db/migrations -type f -not -name '.keep' | wc -l); \
		if [ $$count -eq 0 ]; then \
			echo "No migrations to run: ./db/migrations folder is empty"; \
		else \
			migrate -database $(DB_CONN_STRING) -path db/migrations up; \
			migrate -database $(DB_TEST_CONN_STRING) -path db/migrations up; \
		fi

db-reset: ## Reset databases
	@$(PSQL) -tc "DROP DATABASE IF EXISTS $(DB)"
	@$(PSQL) -tc "DROP DATABASE IF EXISTS $(DB_TEST)"
	@$(MAKE) db-create
	@$(MAKE) db-migrate

db-down: ## Migrate DOWN databases
	POSTGRESQL_URL="$(DB_CONN_STRING)" \
	bash -c 'migrate -database $(POSTGRESQL_URL) -path db/migrations down'
	POSTGRESQL_URL="$(DB_TEST_CONN_STRING)" \
	bash -c 'migrate -database $(POSTGRESQL_URL) -path db/migrations down'

db-up: ## Migrate UP databases
	POSTGRESQL_URL="$(DB_CONN_STRING)" \
	bash -c 'migrate -database $${POSTGRESQL_URL} -path db/migrations up'
	POSTGRESQL_URL="$(DB_TEST_CONN_STRING)" \
	bash -c 'migrate -database $${POSTGRESQL_URL} -path db/migrations up'

db-setup: db-create db-migrate ## Setup the databases (create and migrate)

lint: ## Run linter
	golangci-lint run

fix-fieldalignment: ## Fix field alignment for structs
	fieldalignment -fix ./...

gen-mocks: ## Generate Interface mocks
	mockery --all

test: ## Run tests
	@go clean -testcache
	@go test -p=1 -count=1 -json -v ./... 2>&1 | tee /tmp/gotest.log | gotestfmt

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
