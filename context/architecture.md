# Arsitektur API
## Bootstrap

cmd/server/main.go memuat config, timezone, logger, database, dan router lalu menangani graceful shutdown. internal/router/ membangun dependency graph dan mendaftarkan route base serta /api/v1.

## Feature modules

Domain aktif berada di internal/features/<domain>/ dengan pembagian umum:
- interface/http: Gin handler, binding, status code, dan endpoint annotation;
- application: use-case/service serta orchestration;
- infrastructure: repository, external provider, dan persistence-specific code.

internal/model, internal/dto, internal/middleware, internal/shared, dan pkg dipakai lintas feature. Saat menambah feature, ikuti pola yang paling dekat.

## Request flow

Request melewati CORS/recovery/logging, rate limit, auth JWT, dan role middleware sesuai route. Handler melakukan binding/validation lalu memanggil application service. Repository mengakses GORM/PostgreSQL. Response memakai helper terstandar agar client web/mobile dapat menormalkan envelope dan error.

## External systems

DeepSeek dipakai untuk chat/moderation/journal/wellness, Midtrans untuk payment, VAPID untuk push, storage untuk upload, dan golang-migrate untuk schema. Semua credential dibaca dari environment.

## Perubahan aman

Perubahan pada shared response, middleware, user context, pagination, atau model dapat berdampak ke semua domain. Cari seluruh consumer dan perbarui OpenAPI serta client sebelum merge.
