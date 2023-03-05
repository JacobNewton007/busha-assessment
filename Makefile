#!make
include .envrc


# ==================================================================================== # 
# HELPERS
# ==================================================================================== #

## help: print this help
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== # 
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api: db/migrations/up
	go run ./main.go -db-dsn=${BUSHA_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${BUSHA_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration file for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running migrations...'
	migrate -path ./migrations -database ${BUSHA_DB_DSN} up

## db/migrations/down: apply all up database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running migrations...'
	migrate -path ./migrations -database ${BUSHA_DB_DSN} down

# ==================================================================================== # 
# QUALITY CONTROL
# ==================================================================================== #

.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running test...'
	go test -race -vet=off ./...

.PHONY: audit
vendor:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify 
	@echo 'Vendoring dependencies...'
	go mod vendor

# ==================================================================================== # 
# BUILD
# ==================================================================================== #

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build/main.go
build/api:
	@echo 'Building main...'
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go build  -o=./bin/api ./main.go
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./main.go 
.PHONY: start
start: db/migrations/up
		./bin/api -db-dsn=${BUSHA_DB_DSN}

build/dev:
	docker-compose up

docker/build:
	docker build -t app-prod . --target production

docker/start:
	docker run -p 4000:4000 --name app-prod app-prod