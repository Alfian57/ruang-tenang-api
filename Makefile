.PHONY: run build clean swagger migrate-up migrate-down migrate-create seed-dev seed-demo seed-prod install-tools quickstart-check

# Load environment variables
-include .env

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=ruang-tenang-api
SEEDER_DEV_NAME=seeder-dev
SEEDER_PROD_NAME=seeder-prod

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
	go install github.com/swaggo/swag/cmd/swag@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
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
	@echo "🔨 Building seeder-dev..."
	$(GOBUILD) -o $(BIN_DIR)/$(SEEDER_DEV_NAME) $(CMD_DIR)/seed-dev
	@echo "🔨 Building seeder-prod..."
	$(GOBUILD) -o $(BIN_DIR)/$(SEEDER_PROD_NAME) $(CMD_DIR)/seed-prod
	@echo "✅ Build complete!"

# Run the application
run:
	@echo "🚀 Starting server..."
	$(GOCMD) run $(CMD_DIR)/server/main.go

# Run with hot reload (requires air)
dev:
	@echo "🔄 Starting development server with hot reload..."
	air

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
	@echo "✅ Swagger docs generated!"

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

seed-dev:
	@echo "🌱 Running seeder-dev..."
	$(GOCMD) run $(CMD_DIR)/seed-dev
	@echo "🗑️  Clearing cache..."
	@curl -s -X POST http://localhost:8080/dev/cache/clear > /dev/null 2>&1 || echo "   ⚠️  Server not running, cache will be fresh on next start"
	@echo "✅ Dev seeding complete!"

seed-demo:
	@echo "🎬 Refreshing curated demo state (--reset)..."
	$(GOCMD) run $(CMD_DIR)/seed-dev --reset --profile demo
	@echo "🗑️  Clearing cache..."
	@curl -s -X POST http://localhost:8080/dev/cache/clear > /dev/null 2>&1 || echo "   ⚠️  Server not running, cache will be fresh on next start"
	@echo "✅ Curated demo state is ready!"

seed-prod:
	@echo "🌱 Running seeder-prod..."
	$(GOCMD) run $(CMD_DIR)/seed-prod
	@echo "✅ Production seeding complete!"

# Full setup (for new installations)
setup: deps migrate-up seed-dev
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
	@echo "  clean         - Clean build artifacts"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  migrate-up    - Run all migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-create- Create new migration files"
	@echo "  seed-dev      - Run development database seeder"
	@echo "  seed-demo     - Reset + seed curated demo state"
	@echo "  seed-prod     - Run production database seeder"
	@echo "  quickstart-check - Verify DB/migration/seed/server/demo checklist"
	@echo "  setup         - Full setup (deps + migrate + seed)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
