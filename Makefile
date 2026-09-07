.PHONY: run build clean test swagger swagger-check migrate-up migrate-down migrate-create seed install-tools quickstart-check

# Load environment variables
-include .env

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Tool versions used by the repository contract.
SWAG_VERSION ?= v1.16.4
MIGRATE_VERSION ?= v4.19.1

# Binary names
BINARY_NAME=ruang-tenang-api
SEEDER_NAME=seeder

# Database parameters
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Directories
CMD_DIR=./cmd
BIN_DIR=./bin
MIGRATIONS_DIR=./migrations

# Default target
all: build

# Install required tools
install-tools:
	@echo "📦 Installing required tools..."
	go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	@echo "✅ Tools installed!"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies downloaded!"

# Build the application
build: deps
	@echo "🔨 Building server..."
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/server/main.go
	@echo "🔨 Building seeder..."
	$(GOBUILD) -o $(BIN_DIR)/$(SEEDER_NAME) $(CMD_DIR)/seeder
	@echo "✅ Build complete!"

# Run the application
run:
	@echo "🚀 Starting server..."
	$(GOCMD) run $(CMD_DIR)/server/main.go

# Run with hot reload (requires air)
dev:
	@echo "🔄 Starting development server with hot reload..."
	air

# Run the Go test suite without changing module files.
test:
	$(GOCMD) test ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -f *.out
	@echo "✅ Clean complete!"
	@echo "✅ Clean complete!"

# Generate Swagger documentation
swagger:
	@echo "📚 Generating Swagger docs..."
	swag init -g $(CMD_DIR)/server/main.go -o ./docs
	@if [ -f ./docs/swagger.yaml ]; then cp ./docs/swagger.yaml ./docs/openapi.yaml; fi
	@echo "✅ Swagger docs generated!"

# Verify that the tracked OpenAPI snapshot matches the current annotations.
swagger-check:
	@echo "🔎 Checking OpenAPI snapshot drift..."
	@tmp_dir=$$(mktemp -d); \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	swag init -g $(CMD_DIR)/server/main.go -o "$$tmp_dir" >/dev/null; \
	diff -u ./docs/openapi.yaml "$$tmp_dir/swagger.yaml"
	@echo "✅ OpenAPI snapshot is up to date!"

# Database migrations
migrate-up:
	@echo "⬆️  Running migrations up..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "✅ Migrations complete!"

migrate-down:
	@echo "⬇️  Running migrations down..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "🗑️  Cleaning uploads directory..."
	@rm -rf uploads/*
	@echo "✅ Migration rolled back and uploads cleared!"

migrate-down-all:
	@echo "⬇️  Rolling back all migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	@echo "🗑️  Cleaning uploads directory..."
	@rm -rf uploads/*
	@echo "✅ All migrations rolled back and uploads cleared!"

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name
	@echo "✅ Migration files created!"

migrate-fresh:
	@echo "🔄 Refreshing database..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "🗑️  Cleaning uploads directory..."
	@rm -rf uploads/*
	@echo "✅ Database refreshed and uploads cleared!"

seed:
	@echo "🌱 Running presentation seeder..."
	$(GOCMD) run $(CMD_DIR)/seeder $(SEED_FLAGS)
	@echo "🗑️  Clearing cache..."
	@curl -s -X POST http://localhost:8080/dev/cache/clear > /dev/null 2>&1 || echo "   ⚠️  Server not running, cache will be fresh on next start"
	@echo "✅ Presentation seeding complete!"

# Full setup (for new installations)
setup: deps migrate-up seed
	@echo "✅ Setup complete! Run 'make run' to start the server."

quickstart-check:
	@echo "🧪 Running backend quickstart verification..."
	@bash ./scripts/quickstart_check.sh

# Docker commands
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t ruang-tenang-api .
	@echo "✅ Docker image built!"

docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --env-file .env ruang-tenang-api

# Help
help:
	@echo "Available targets:"
	@echo "  install-tools  - Install required Go tools (swag, migrate)"
	@echo "  deps          - Download dependencies"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run with hot reload (requires air)"
	@echo "  test          - Run the Go test suite"
	@echo "  clean         - Clean build artifacts"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  swagger-check - Check tracked OpenAPI snapshot for drift"
	@echo "  migrate-up    - Run all migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-create- Create new migration files"
	@echo "  seed          - Run the single presentation database seeder"
	@echo "                  Use: make seed SEED_FLAGS=--reset"
	@echo "  quickstart-check - Verify DB/migration/seed/server/demo checklist"
	@echo "  setup         - Full setup (deps + migrate + seed)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
