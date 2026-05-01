# Ruang Tenang API

[![Build and Deploy](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml/badge.svg)](https://github.com/Alfian57/ruang-tenang-api/actions/workflows/build-and-deploy.yml)

Backend API untuk aplikasi Ruang Tenang.

## Checklist Quickstart

- [x] `.env` sudah dibuat dari `.env.example`.
- [ ] PostgreSQL aktif dan kredensial valid.
- [ ] Migrasi berhasil dijalankan.
- [ ] Seeder utama berhasil dijalankan.
- [ ] Server berjalan di port target.

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
│   └── seeder/         # Single presentation-ready seeder
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
make seed           # Seed database with presentation-ready data
make quickstart-check # Verify quickstart checklist (db/migrate/seed/server)
make help           # Show all commands
```

## Command Penting

```bash
make run            # Jalankan server
make build          # Build binary
make swagger        # Generate Swagger docs
make migrate-up     # Jalankan migrasi
make migrate-down   # Rollback migrasi terakhir
make seed           # Seeder utama dengan data siap presentasi
make quickstart-check # Verifikasi otomatis checklist quickstart backend
make help           # Daftar command
```

Seeder utama sudah mencakup katalog, akun, konten, komunitas, billing, B2B, moderasi, dan state demo. Untuk reset database sebelum seed, gunakan:

```bash
make seed SEED_FLAGS=--reset
```

## Catatan Refactor Migration

- Perubahan schema yang mengupdate tabel existing sudah digabung langsung ke migration create tabel asal untuk seluruh histori migration.
- Migration yang sudah tidak terpakai telah dihapus.
- Seluruh file migration aktif sekarang mengikuti pola satu tabel per migration.
- Karena histori migration dirombak, lakukan reset database lalu jalankan ulang `make migrate-up` dari awal chain sebelum menjalankan seeder.

## Dokumentasi API

- Swagger UI: http://localhost:8080/swagger/index.html

## Akun Uji (Seeder Utama)

- Admin: admin@ruang-tenang.com / password
- Mitra: mitra@ruang-tenang.com / password
- User premium personal: gading@gmail.com / password
- User premium B2B: dery@gmail.com / password
- User freemium: andhika@gmail.com / password

Notes:
- Password admin bisa dioverride lewat `SEED_ADMIN_PASSWORD` (legacy: `ADMIN_PASSWORD`).

## License

MIT
