# Ruang Tenang API

[![Build and Deploy](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml/badge.svg)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml)

Backend API untuk aplikasi Ruang Tenang.

## Checklist Quickstart

- [x] `.env` sudah dibuat dari `.env.example`.
- [ ] PostgreSQL aktif dan kredensial valid.
- [ ] Migrasi berhasil dijalankan.
- [ ] Seeder dev berhasil dijalankan.
- [ ] Server berjalan di port target.
- [ ] Demo state sudah di-refresh lewat `make seed-demo`.

Gunakan verifikasi otomatis berikut untuk mengecek 5 item di atas sekaligus:

```bash
make quickstart-check
```

## Tech Stack

- Go 1.24, Gin, GORM, PostgreSQL
- golang-migrate, JWT, Swagger
- Zap, Viper

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

## Prasyarat

- Go 1.24+
- PostgreSQL 14+
- Make

## Setup Cepat

1. Salin env:

```bash
cp .env.example .env
```

2. Isi kredensial database di `.env`.

Catatan production:
- set `FRONTEND_URL=https://ruang-tenang.site`
- pastikan `CORS_ALLOWED_ORIGINS` memuat origin frontend

3. Install tools, setup, dan jalankan server:

```bash
make install-tools
make setup
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
make seed-demo      # Reset + seed curated demo state
make seed-prod      # Seed database for production
make quickstart-check # Verify quickstart checklist (db/migrate/seed/server/demo)
make help           # Show all commands
```

## Command Penting

```bash
make run            # Jalankan server
make build          # Build binary
make swagger        # Generate Swagger docs
make migrate-up     # Jalankan migrasi
make migrate-down   # Rollback migrasi terakhir
make seed-dev       # Seeder development
make seed-demo      # Reset + seeder dev dengan profile demo curated
make seed-prod      # Seeder production
make quickstart-check # Verifikasi otomatis checklist quickstart backend
make help           # Daftar command
```

`make seed-demo` sekarang menjalankan `seed-dev --reset --profile demo` untuk menyiapkan state khusus live demo.

## Catatan Refactor Migration

- Perubahan schema yang mengupdate tabel existing sudah digabung langsung ke migration create tabel asal untuk seluruh histori migration.
- Migration yang sudah tidak terpakai telah dihapus.
- Seluruh file migration aktif sekarang mengikuti pola satu tabel per migration.
- Karena histori migration dirombak, lakukan reset database lalu jalankan ulang `make migrate-up` dari awal chain sebelum menjalankan seeder.

## Dokumentasi API

- Swagger UI: http://localhost:8080/swagger/index.html

## Akun Uji (Seeder Dev)

- Demo Utama: demo.utama@ruang-tenang.com / password
- Demo Cadangan: demo.cadangan@ruang-tenang.com / password
- Admin: admin@ruang-tenang.com / password
- Moderator: moderator@ruang-tenang.com / password
- Member: gading@gmail.com / password
- Member: dery@gmail.com / password
- Member: andhika@gmail.com / password

Notes:
- Di development, admin diseed via `production.SeedAdminUser`.
- Kredensial admin bisa dioverride lewat `SEED_ADMIN_EMAIL` dan `SEED_ADMIN_PASSWORD` (legacy: `ADMIN_EMAIL`, `ADMIN_PASSWORD`).

## License

MIT
