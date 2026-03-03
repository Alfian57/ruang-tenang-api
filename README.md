# Ruang Tenang API

[![CI Tests](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/ci-tests.yml/badge.svg)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/ci-tests.yml)
[![Build and Deploy](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml/badge.svg)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml)
[![Coverage Gate](https://img.shields.io/badge/coverage%20gate-%E2%89%A5%2090%25-brightgreen)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/ci-tests.yml)

Backend API untuk aplikasi Ruang Tenang - Platform Kesehatan Mental.

## Testing Pyramid

Target komposisi testing yang dipakai:
- Unit Test: 70% - 80% (fokus logika bisnis/core)
- Integration Test: 15% - 20% (repository/db + handler integration flow)
- E2E/API Test: 5% - 10% (router + public/protected API smoke)

Gunakan command berikut:
- `make test-unit`
- `make test-integration`
- `make test-e2e`
- `make test-pyramid`

Catatan:
- CI menjalankan `make test-pyramid` di setiap push/PR.
- Coverage total diproteksi dengan gate minimal 90%.

### Test Classification Rules (Wajib Diikuti)

Gunakan aturan berikut saat menambah test baru agar komposisi pyramid tetap konsisten:

- **Unit (`make test-unit`)**
   - Tempatkan di package: `./pkg/...`, `./internal/config`, `./internal/dto`, `./internal/middleware`, `./internal/model`, `./internal/service`
   - Fokus: pure logic, helper, mapper, validation, edge-case branch
   - Hindari ketergantungan DB/network

- **Integration (`make test-integration`)**
   - Tempatkan di: `./internal/database`, `./internal/repository`, dan flow integrasi handler
   - Untuk handler integration flow, gunakan suffix nama test file: `*_integration_test.go`
   - Fokus: interaksi lintas layer (DB/repository/handler) dengan setup test yang realistis

- **E2E/API (`make test-e2e`)**
   - Tempatkan di: `./internal/router`, `./cmd/api`, `./cmd/server`
   - Untuk smoke API end-to-end, gunakan nama yang jelas (contoh: `api_e2e_test.go`)
   - Fokus: route wiring, middleware chain, public/protected endpoint smoke

Prinsip umum:
- Jika test dapat berjalan tanpa DB dan tanpa wiring router penuh → **Unit**
- Jika test memverifikasi query/transaction/ORM behavior → **Integration**
- Jika test memverifikasi endpoint dari route hingga response contract → **E2E/API**

### PR Checklist (Testing)

Sebelum membuat / merge PR, pastikan:

- [ ] Test baru diklasifikasikan benar ke **Unit / Integration / E2E** sesuai rules di atas
- [ ] Handler integration flow menggunakan suffix `*_integration_test.go` bila relevan
- [ ] Menjalankan command yang sesuai perubahan (`make test-unit`, `make test-integration`, `make test-e2e`)
- [ ] Untuk perubahan lintas layer, menjalankan `make test-pyramid`
- [ ] Perubahan tidak menurunkan target komposisi test pyramid

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
make test           # Run tests
make test-unit      # Run unit-focused tests
make test-integration # Run integration-focused tests
make test-e2e       # Run E2E/API smoke tests
make test-pyramid   # Run unit + integration + e2e suite
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

After running development seeder (`make seed` with default mode):
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
