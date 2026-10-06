# Data dan Operasi
## Configuration

Core config: APP_ENV, PORT, JWT_SECRET, CORS_ALLOWED_ORIGINS, dan DATABASE_URL. APP_ENV, PORT, dan CORS memiliki default development; set eksplisit pada deployment. .env.example adalah daftar variable yang didukung; secret hanya di deployment secret store.

## Database

Engine database adalah MySQL 8.0.13+ atau MariaDB 10.2+. Skema di-baseline ulang menjadi migration per tabel di `migrations/` (satu file `.up.sql`/`.down.sql` per tabel, diurutkan sesuai dependency foreign key) yang mereproduksi skema PostgreSQL final (termasuk tabel yang dipakai model seperti organizations, notifications, premium_plans, dan forum_posts). Migration dijalankan dengan golang-migrate dan bersifat fresh-start untuk MySQL; tidak ada jalur upgrade in-place dari PostgreSQL. Seeder presentation dapat membuat/reset fixture dan hanya aman untuk local/staging. migrate-fresh serta RUN_MIGRATE_FRESH=true menghapus data.

## Seed dan demo

make seed membuat katalog, akun, konten, community, billing, B2B, moderation, dan state demo. Katalog musik presentation menyertakan delapan kategori dan komposisi Incompetech CC BY 4.0; sumber, atribusi, serta thumbnail tercatat di `docs/SEED_CONTENT.md`. Password admin dapat dioverride SEED_ADMIN_PASSWORD; akun demo bukan kredensial production.

## AI, cache, dan file

Prompt AI di-embed dari prompts/. DeepSeek API key, endpoint, dan model harus dikonfigurasi per environment. Upload disimpan di path yang dilayani route static-safe; jangan mengekspos directory traversal atau credential. Cache clear tersedia melalui route development/admin dan tidak boleh dibuka sembarangan.

## Container dan CI

Docker builder berbasis golang:1.25-alpine, menjalankan Swag, build server/seeder/migrate, lalu runtime non-root Alpine. Entrypoint dapat menjalankan migration/seeder sebelum server. Workflow GitHub Actions dan variable/secret deployment adalah sumber kebenaran operasi.
