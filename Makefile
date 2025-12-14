.PHONY: run build test clean swagger migrate-up migrate-down migrate-create seed install-tools

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=ruang-tenang-api
SEEDER_NAME=seeder

# Database parameters
DB_URL=postgres://postgres:postgres@localhost:5432/ruang_tenang?sslmode=disable

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
	@echo "🔨 Building seeder..."
	$(GOBUILD) -o $(BIN_DIR)/$(SEEDER_NAME) $(CMD_DIR)/seeder/main.go
	@echo "✅ Build complete!"

# Run the application
run:
	@echo "🚀 Starting server..."
	$(GOCMD) run $(CMD_DIR)/server/main.go

# Run with hot reload (requires air)
dev:
	@echo "🔄 Starting development server with hot reload..."
	air

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
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
	@echo "✅ Migration rolled back!"

migrate-down-all:
	@echo "⬇️  Rolling back all migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down
	@echo "✅ All migrations rolled back!"

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name
	@echo "✅ Migration files created!"

migrate-force:
	@read -p "Enter version: " version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $$version
	@echo "✅ Migration version forced!"

# Run seeder
seed:
	@echo "🌱 Running seeder..."
	$(GOCMD) run $(CMD_DIR)/seeder/main.go
	@echo "✅ Seeding complete!"

# Full setup (for new installations)
setup: deps migrate-up seed
	@echo "✅ Setup complete! Run 'make run' to start the server."

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
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  swagger       - Generate Swagger documentation"
	@echo "  migrate-up    - Run all migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-create- Create new migration files"
	@echo "  seed          - Run database seeder"
	@echo "  setup         - Full setup (deps + migrate + seed)"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
