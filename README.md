# Ruang Tenang API

[![Build and Deploy](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml/badge.svg)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml)


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
│   ├── seed-dev/       # Development seeder
│   └── seed-prod/      # Production seeder
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
   - Untuk production, set `FRONTEND_URL=https://ruang-tenang.site`
   - Pastikan `CORS_ALLOWED_ORIGINS` juga memuat origin frontend (contoh: `https://ruang-tenang.site`)

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
make swagger        # Generate Swagger docs
make migrate-up     # Run migrations
make migrate-down   # Rollback last migration
make seed-dev       # Seed database for development
make seed-prod      # Seed database for production
make help           # Show all commands
```

## API Documentation

After starting the server, visit:
- Swagger UI: http://localhost:8080/swagger/index.html

## Test Accounts

After running development seeder (`make seed-dev`):
- **Admin**: admin@ruang-tenang.com / password
- **Moderator**: moderator@ruang-tenang.com / password
- **Member**: gading@gmail.com / password
- **Member**: dery@gmail.com / password
- **Member**: andhika@gmail.com / password

Notes:
- In development mode, admin user is seeded by `production.SeedAdminUser` with default email/password above.
- Admin credentials can be overridden via env: `SEED_ADMIN_EMAIL` and `SEED_ADMIN_PASSWORD` (legacy: `ADMIN_EMAIL`, `ADMIN_PASSWORD`).

## License

MIT
