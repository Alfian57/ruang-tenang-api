# Ruang Tenang API

Backend API untuk aplikasi Ruang Tenang - Platform Kesehatan Mental.

## Tech Stack

- **Go 1.24** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM library
- **PostgreSQL** - Database
- **golang-migrate** - Database migrations
- **JWT** - Authentication
- **Swagger** - API documentation
- **Zap** - Logging
- **Viper** - Configuration management

## Project Structure

```
├── cmd/
│   ├── server/         # Main server entry point
│   └── seeder/         # Database seeder
├── internal/
│   ├── config/         # Configuration
│   ├── database/       # Database connection
│   ├── dto/            # Data Transfer Objects
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Middleware (auth, cors, logger)
│   ├── models/         # GORM models
│   ├── repositories/   # Data access layer
│   ├── router/         # Route definitions
│   └── services/       # Business logic
├── migrations/         # SQL migration files
├── pkg/
│   ├── logger/         # Zap logger setup
│   └── utils/          # Utility functions (JWT, password)
└── docs/               # Swagger generated docs
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- Make

### Installation

1. Clone the repository
2. Copy environment file:
   ```bash
   cp .env.example .env
   ```
3. Update `.env` with your database credentials

4. Install required tools:
   ```bash
   make install-tools
   ```

5. Run setup (download deps, migrate, seed):
   ```bash
   make setup
   ```

6. Start the server:
   ```bash
   make run
   ```

### Available Commands

```bash
make run            # Run the server
make build          # Build binary
make test           # Run tests
make swagger        # Generate Swagger docs
make migrate-up     # Run migrations
make migrate-down   # Rollback last migration
make seed           # Seed database
make help           # Show all commands
```

## API Documentation

After starting the server, visit:
- Swagger UI: http://localhost:8080/swagger/index.html

## Test Accounts

After running seeder:
- **Admin**: admin@ruangtenang.id / admin123
- **Member**: john@example.com / member123

## License

MIT
