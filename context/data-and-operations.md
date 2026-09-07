# Data dan Operasi
## Configuration

Core config: APP_ENV, PORT, JWT_SECRET, CORS_ALLOWED_ORIGINS, dan database URL atau DB parts. APP_ENV, PORT, dan CORS memiliki default development; set eksplisit pada deployment. APP_PORT dipertahankan sebagai fallback legacy. .env.example adalah daftar variable yang didukung; secret hanya di deployment secret store.

## Database

Migration aktif berada di migrations/ dan dijalankan berurutan dengan golang-migrate. Migration lama yang sudah dipakai tidak boleh direwrite. Seeder presentation dapat membuat/reset fixture dan hanya aman untuk local/staging. migrate-fresh serta RUN_MIGRATE_FRESH=true menghapus data.

## Seed dan demo

make seed membuat katalog, akun, konten, community, billing, B2B, moderation, dan state demo. Password admin dapat dioverride SEED_ADMIN_PASSWORD; akun demo bukan kredensial production.

## AI, cache, dan file

Prompt AI di-embed dari prompts/. Gemini API key dan model harus dikonfigurasi per environment. Upload disimpan di path yang dilayani route static-safe; jangan mengekspos directory traversal atau credential. Cache clear tersedia melalui route development/admin dan tidak boleh dibuka sembarangan.

## Container dan CI

Docker builder berbasis golang:1.25-alpine, menjalankan Swag, build server/seeder/migrate, lalu runtime non-root Alpine. Entrypoint dapat menjalankan migration/seeder sebelum server. Workflow GitHub Actions dan variable/secret deployment adalah sumber kebenaran operasi.
