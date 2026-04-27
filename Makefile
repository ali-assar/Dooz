include .env
MIGRATE=migrate -path=migration -database "$(DATABASE_HOST)" -verbose

APP_NAME=dooz
BINARY_NAME=bin/$(APP_NAME)
GO=/usr/local/go/bin/go

.PHONY: build run wire devtools clean swagger db-migrate-up db-migrate-down db-seed cron help

wire:
	@echo "Checking for wire..."
	@GOPATH=$$($(GO) env GOPATH); \
	WIRE_BIN="$$GOPATH/bin/wire"; \
	if [ ! -f "$$WIRE_BIN" ]; then \
		echo "wire not found. Installing..."; \
		$(GO) install github.com/google/wire/cmd/wire@latest; \
	fi; \
	echo "Generating Wire code..."; \
	cd cmd && $$WIRE_BIN

build: wire
	@echo "Building application..."
	@mkdir -p bin
	$(GO) build -o $(BINARY_NAME) ./cmd
	@echo "Build complete: $(BINARY_NAME)"

run: wire
	@echo "Running application..."
	$(GO) run ./cmd

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f cmd/wire_gen.go
	@echo "Clean complete"

devtools:
	@echo "Installing devtools"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) install github.com/google/wire/cmd/wire@latest
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

swagger:
	@echo "Checking for swag..."
	@GOPATH=$$($(GO) env GOPATH); \
	SWAG_BIN="$$GOPATH/bin/swag"; \
	if [ ! -f "$$SWAG_BIN" ]; then \
		echo "swag not found. Installing..."; \
		$(GO) install github.com/swaggo/swag/cmd/swag@latest; \
	fi; \
	echo "Generating Swagger documentation..."; \
	$$SWAG_BIN fmt; \
	$$SWAG_BIN init --parseDependency -g ./cmd/main.go -o ./docs; \
	echo "Swagger docs generated in ./docs"

db-migrate-up:
	$(MIGRATE) up

db-migrate-down:
	$(MIGRATE) down

db-create-migration:
	@read -p "What is the name of migration?" NAME; \
	${MIGRATE} create -ext sql -seq -dir migration $$NAME

db-seed:
	@echo "Seeding database..."
	$(GO) run ./cmd --seed --seed-count=50

cron:
	$(GO) run ./cmd --cron

fmt:
	gofumpt -l -w .

help:
	@echo "Available targets:"
	@echo "  make wire            - Generate Wire dependency injection code"
	@echo "  make build           - Build the application binary"
	@echo "  make run             - Run the application"
	@echo "  make swagger         - Generate Swagger docs"
	@echo "  make db-migrate-up   - Run database migrations"
	@echo "  make db-seed         - Seed database"
	@echo "  make cron            - Run cron jobs"
	@echo "  make clean           - Clean build artifacts"
