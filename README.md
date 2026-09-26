# Ruang Tenang API

Backend Go untuk platform kesehatan mental Ruang Tenang. API ini adalah sumber kebenaran data, autentikasi, otorisasi, migration PostgreSQL, konten, AI, komunitas, gamifikasi, billing, dan B2B. Web dan mobile memakai API yang sama.

## Mulai cepat

### Prasyarat

- Go 1.25 atau lebih baru.
- PostgreSQL 14+.
- Make.
- Tool swag dan migrate untuk workflow dokumentasi/migration; pasang dengan make install-tools.

### Setup lokal

```bash
cp .env.example .env
make install-tools
make setup
make run
```

Server berjalan pada http://localhost:8080 secara default. Health check tersedia di /health, dan Swagger UI di http://localhost:8080/swagger/index.html.

make setup mengunduh dependency, menjalankan seluruh migration, dan mengisi data demo. Pastikan database lokal sudah tersedia serta .env berisi kredensial yang benar.

## Perintah

```bash
make deps              # Download/tidy Go modules
make run               # Jalankan server
make dev               # Hot reload dengan air
make build             # Build server dan seeder
make test              # go test ./...
make swagger           # Regenerasi docs/docs.go dan docs/openapi.yaml
make swagger-check     # Periksa drift OpenAPI tanpa mengubah tracked file
make migrate-up        # Jalankan migration
make migrate-down      # Rollback migration terakhir
make migrate-fresh     # Drop dan buat ulang database (destruktif)
make seed              # Seed data presentasi
make quickstart-check  # Verifikasi env, PostgreSQL, migration, seed, health
make docker-build      # Build image API
make docker-run        # Jalankan container memakai .env
```

Perintah seeding dan migration fresh hanya untuk development/demo. Jangan jalankan terhadap database production tanpa prosedur backup dan approval operasional.

## Arsitektur

- cmd/server/ adalah entrypoint HTTP API.
- cmd/seeder/ menjalankan seeder presentation; cmd/migrate/ menjalankan migration.
- internal/features/<domain>/ memakai pembagian application, infrastructure, dan interface/http.
- internal/router/ menggabungkan route base dan /api/v1 serta wiring dependency.
- internal/middleware/ berisi auth, role, CORS, rate limit, validation, logging, dan recovery.
- internal/model/, internal/dto/, dan pkg/ berisi tipe lintas fitur/utilitas bersama.
- migrations/ berisi chain SQL yang dijalankan golang-migrate.
- prompts/ berisi prompt AI yang di-embed ke binary.
- docs/ berisi generated Swagger Go dan snapshot OpenAPI untuk review contract.

Detail domain, contract, migration, dan operasi ada di context/README.md. Aturan kerja AI agent ada di AGENTS.md.

## Domain dan role

API melayani akun/auth; artikel, stories, forum, journal, mood, musik, playlist, dan search; chat AI, journal AI context, wellness, serta crisis/moderation; XP, level, badge, daily task, leaderboard, progress map, rewards, dan XP boost; premium, top-up, Duitku Pop webhook, invoice, dan feature usage; organisasi, B2B, seat, onboarding, insight, SSO, audit; push subscription, broadcast, upload, dan endpoint admin/moderator.

Role server adalah boundary keamanan. Client tidak boleh dipercaya hanya karena route frontend menyembunyikan menu.

## Environment

Salin .env.example ke .env. LoadConfig membaca .env/environment variables dan memvalidasi variable required.

| Kelompok | Variable |
| --- | --- |
| Core (set explicitly) | APP_ENV, PORT, JWT_SECRET, CORS_ALLOWED_ORIGINS, dan DATABASE_URL atau DB_HOST/DB_PORT/DB_USER/DB_NAME |
| Runtime | APP_TIMEZONE, FRONTEND_URL, JWT_EXPIRY_HOURS |
| AI | DEEPSEEK_API_KEY, DEEPSEEK_BASE_URL, AI_CHAT_MODEL, AI_MODERATION_MODEL, AI_JOURNAL_MODEL, AI_WELLNESS_MODEL |
| Billing | API_PUBLIC_URL, DUITKU_SANDBOX, DUITKU_MERCHANT_CODE, DUITKU_API_KEY, FRONTEND_URL |
| WhatsApp | FONNTE_TOKEN |
| Chat quota | CHAT_DAILY_MESSAGE_LIMIT, CHAT_QUOTA_RESET_INTERVAL |
| Web push | VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY, VAPID_CONTACT |
| Tools/deploy | MIGRATIONS_PATH, RUN_MIGRATE, RUN_MIGRATE_FRESH, RUN_SEEDER, SEED_ADMIN_PASSWORD |

APP_ENV, PORT, dan CORS_ALLOWED_ORIGINS memiliki default development di kode; tetap set eksplisit pada deployment. APP_PORT hanya fallback legacy. Jangan commit .env, DeepSeek key, server key, JWT secret, VAPID private key, atau kredensial database.

Panduan pengaturan Duitku Pop ada di [docs/DUITKU_SETUP.md](docs/DUITKU_SETUP.md). Reset kata sandi dan OTP nomor pengguna dikirim melalui Fonnte; nomor WhatsApp diperlukan saat registrasi dan harus diverifikasi sebelum login penuh.
Untuk Fonnte, hubungkan perangkat WhatsApp lalu ambil token pada menu perangkat sesuai [dokumentasi token Fonnte](https://docs.fonnte.com/token-api-key/). Simpan token sebagai `FONNTE_TOKEN` hanya di backend. Pengiriman memakai [API pesan Fonnte](https://docs.fonnte.com/api-send-message/).

Sumber, lisensi musik, artikel, aset ilustrasi, dan cakupan data demo dijelaskan di [docs/SEED_CONTENT.md](docs/SEED_CONTENT.md).

## OpenAPI

Anotasi Swagger pada handler dan cmd/server/main.go adalah sumber kebenaran. Jalankan make swagger lalu make swagger-check. docs/openapi.yaml adalah snapshot ter-track; docs/docs.go dan output Swagger lain diperbarui oleh command yang sama. Jangan mengedit output generated secara manual.

## CI dan deployment

Workflow .github/workflows/build-and-deploy.yml membangun image API berbasis Go 1.25, menjalankan swag init/go mod tidy, membangun server/seeder/migrate, push image, dan deploy sesuai trigger workflow. Entrypoint container mendukung migration/seeder melalui flag environment.

## Data demo

Seeder presentasi membuat akun dan konten demo untuk local/staging. Password default hanya untuk demo dan dapat diganti dengan SEED_ADMIN_PASSWORD; ADMIN_PASSWORD hanya legacy. Jangan membawa akun demo atau password default ke production.

Seeder musik presentation mengisi delapan kategori dengan katalog Incompetech berlisensi CC BY 4.0. Daftar lagu, sumber, kredit, lisensi, dan ilustrasi kategori ada di [docs/SEED_CONTENT.md](docs/SEED_CONTENT.md).
