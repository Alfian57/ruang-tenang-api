# Ruang Tenang API

<!-- CI/CD: 2025-12-13 22:29 WIB -->

REST API untuk aplikasi Ruang Tenang, dibangun dengan Golang dan Gin Framework.

## Tech Stack

- Go 1.24+
- Gin Framework
- PostgreSQL
- JWT Authentication
- Docker

## Development

```bash
# Copy environment file
cp .env.example .env

# Install dependencies
go mod download

# Run development server
go run ./cmd/server
```

## Docker

```bash
# Build image
docker build -t ruang-tenang-api .

# Run container
docker run -p 8080:8080 --env-file .env ruang-tenang-api
```

## Project Structure

```
.
├── cmd/
│   └── server/         # Application entry point
├── configs/            # Configuration files
├── internal/
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   ├── repository/     # Database operations
│   └── services/       # Business logic
├── migrations/         # Database migrations
├── Dockerfile
└── go.mod
```

## API Documentation

API documentation available at `/swagger` when running in development mode.

## Deployment

This project uses GitHub Actions for CI/CD. On push to `main` branch:
1. Docker image is built and pushed to GitHub Container Registry
2. VPS pulls the latest image and restarts the container
