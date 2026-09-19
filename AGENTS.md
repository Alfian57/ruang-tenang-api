# AGENTS.md — Ruang Tenang API
Dokumen ini adalah instruksi kanonis untuk AI agent dan contributor. Baca README.md, lalu buka context/ sesuai area perubahan.

CLAUDE.md, GEMINI.md, dan .github/copilot-instructions.md adalah adapter tipis yang merujuk ke dokumen ini; Cursor dan OpenCode memakai AGENTS.md sebagai instruction native.

## Sumber kebenaran

- Runtime truth: go.mod, internal/config/config.go, kode feature, router, migration aktif, Dockerfile, entrypoint, dan workflow CI.
- HTTP route truth: internal/router/router_api_v1_routes.go serta anotasi Swagger pada handler.
- Data truth: SQL migration; model/seed tidak boleh menggantikan migration.
- Contract truth: response/DTO aktual dan docs/openapi.yaml yang diregenerasi dari anotasi.
- Prompt truth: file di prompts/; jangan menanam prompt baru tersebar di service.
- File .orig, .rej, docs/docs.go, dan Swagger output adalah artefak/histori sesuai generatornya.

## Cara bekerja

1. Baca context domain yang relevan sebelum mengubah feature.
2. Ikuti alur route → middleware → interface/http handler → application service → repository/infrastructure.
3. Pertahankan role checks, ownership checks, validation, rate limit, audit, dan privacy boundary.
4. Endpoint baru/perubahan response harus memiliki DTO, anotasi Swagger, test bermakna, dan pembaruan snapshot OpenAPI.
5. Migration baru harus incremental, memiliki file up/down, aman untuk deployment berurutan, dan tidak mengubah migration historis.
6. Jangan mengedit database production, menghapus data, menjalankan migrate-fresh, atau menjalankan seeder reset tanpa instruksi eksplisit.
7. Jangan menaruh secret, token, data kesehatan mental, atau kredensial demo production di kode/log/dokumen.
8. Gunakan error/response helper dan query utility yang sudah ada; jangan membuat envelope baru tanpa contract terdokumentasi.

## Validasi

- Tentukan scope dari file/kode yang berubah terlebih dahulu.
- Default validasi hanya menjalankan pengujian dan pemeriksaan yang relevan
  dengan scope perubahan, misalnya package test yang terdampak, bukan seluruh
  repository.
- Jangan menjalankan test suite, vet, lint, build, integration test, atau
  pemeriksaan lain di luar scope perubahan tanpa konfirmasi eksplisit dari
  user terlebih dahulu. Jika validasi yang lebih luas dibutuhkan, jelaskan
  command dan alasannya lalu minta konfirmasi.
- Perubahan route/DTO: make swagger, lalu make swagger-check; review diff OpenAPI.
- Perubahan migration: jalankan pada database disposable dan uji rollback bila tersedia.
- Perubahan config/Docker: go build ./cmd/... dan review .env.example.
- Sebelum selesai, jalankan git diff --check.
- Jika dependency, PostgreSQL, atau tool tidak tersedia, laporkan error sebenarnya dan command yang belum terverifikasi.

## Generated files dan operasi

- Sumber Swagger adalah anotasi; gunakan make swagger, bukan edit manual docs/docs.go/docs/openapi.yaml.
- go mod tidy dapat mengubah go.mod/go.sum; jalankan hanya ketika dependency memang berubah.
- Entrypoint mendukung RUN_MIGRATE, RUN_MIGRATE_FRESH, dan RUN_SEEDER; fresh migration destruktif.
- PORT adalah konfigurasi port utama. APP_PORT hanya fallback kompatibilitas.
- MIGRATIONS_PATH mengubah lokasi source migration pada command migrate/container.

## Dokumentasi dan koordinasi

- Perbarui README/context/.env.example ketika feature, route, config, migration, seed, AI model, deployment, atau contract berubah.
- Perubahan API wajib dicek terhadap web service/schema dan mobile datasource/model.
- Context harus menjelaskan keputusan/invariant, bukan menyalin seluruh source.
- Jangan mengubah route atau schema hanya untuk membuat docs tampak konsisten; tandai discrepancy sebagai issue teknis terpisah.

## Code review rules

- Flag endpoint terlindungi yang kehilangan AuthMiddleware/UserMiddleware/AdminMiddleware/MitraMiddleware.
- Flag akses journal/chat/profile tanpa ownership/privacy enforcement.
- Flag migration destruktif, secret yang di-log, webhook tanpa verifikasi, dan response envelope yang tidak terdokumentasi.
