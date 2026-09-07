# API Contract
## Base dan auth

Base path API adalah /api/v1. Route publik dan route terlindungi dibedakan di internal/router/router_api_v1_routes.go. JWT dikirim sebagai Authorization: Bearer <token>. Middleware role/ownership menentukan akses; dokumentasi frontend bukan security boundary.

## Response

Pertahankan response envelope dan error helper yang digunakan source code. Response sukses biasanya membawa data, optional pagination/meta, dan request id; error harus mempertahankan status, message/code, serta detail validation yang dipakai client.

Pagination memakai query page, limit, dan utility terkait. Jika mengubah field pagination atau error, cek normalizer web/services/http/client.ts dan parser di mobile.

## OpenAPI workflow

Anotasi Summary, Tags, Param, Success, Failure, dan Router pada handler adalah sumber dokumentasi. Jalankan make swagger setelah perubahan lalu pastikan make swagger-check lulus. Review docs/openapi.yaml sebagai contract snapshot; jangan mengedit generated file manual.

## Route changes

Perubahan endpoint harus mencakup middleware/role, DTO/request validation, response/error schema, web service/schema, mobile datasource/model, migration/seed bila perlu, test handler/service, dan context. Upload, export CSV/PDF, webhook, dan download file bukan JSON biasa; dokumentasikan content type dan auth-nya.
