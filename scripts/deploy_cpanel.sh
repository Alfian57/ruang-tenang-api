#!/usr/bin/env bash
#
# deploy_cpanel.sh — Rakit bundle deployment Golang API (cPanel + Passenger) .
#
# Menghasilkan:
#   ./upload-api/        (isi siap upload ke ~/ruang-tenang-api/)
#   ./upload-api.zip     (arsip siap upload)
#
# Script ini otomatis:
#   1. Cross-compile biner statis Linux amd64 (server, migrate, seeder).
#   2. Menyalin template app.js (reverse proxy Node.js) ke bundle.
#   3. Menyalin migrations/, storage/, dan menyiapkan uploads/ kosong.
#
# Pemakaian:
#   bash scripts/deploy_cpanel.sh
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

OUT_DIR="${REPO_ROOT}/upload-api"
ZIP_PATH="${REPO_ROOT}/upload-api.zip"
# Nama folder utama di dalam zip (= Application Root cPanel). Saat diextract
# akan terbentuk satu folder ini dengan seluruh isi di dalamnya.
ARCHIVE_ROOT="ruang-tenang-api"
APP_JS_SRC="${REPO_ROOT}/deployment/app.js"

red()    { printf '\033[31m%s\033[0m\n' "$*"; }
green()  { printf '\033[32m%s\033[0m\n' "$*"; }
blue()   { printf '\033[34m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
die()    { red "ERROR: $*"; exit 1; }

# ---------------------------------------------------------------------------
# 1. Preflight
# ---------------------------------------------------------------------------
blue "== Preflight =="
command -v go  >/dev/null 2>&1 || die "go tidak ditemukan."
command -v zip >/dev/null 2>&1 || die "zip tidak ditemukan."

[ -f "${APP_JS_SRC}" ] || die "Template app.js tidak ditemukan di ${APP_JS_SRC}"
[ -d "migrations" ]    || die "Folder migrations/ tidak ditemukan."
[ -d "storage" ]       || die "Folder storage/ tidak ditemukan."

# ---------------------------------------------------------------------------
# 2. Cross-compile biner
# ---------------------------------------------------------------------------
blue "== Cross-compile (linux/amd64) =="
export CGO_ENABLED=0 GOOS=linux GOARCH=amd64

go build -trimpath -o app-main    ./cmd/server/main.go
go build -trimpath -o app-migrate ./cmd/migrate/main.go
go build -trimpath -o app-seeder  ./cmd/seeder/main.go

green "   app-main / app-migrate / app-seeder selesai."

# ---------------------------------------------------------------------------
# 3. Rakit bundle
# ---------------------------------------------------------------------------
blue "== Rakit bundle =="
rm -rf "${OUT_DIR}" "${ZIP_PATH}"
mkdir -p "${OUT_DIR}/uploads"

cp app-main app-migrate app-seeder "${OUT_DIR}/"
cp "${APP_JS_SRC}" "${OUT_DIR}/app.js"
cp -r migrations storage "${OUT_DIR}/"

chmod +x "${OUT_DIR}/app-main" "${OUT_DIR}/app-migrate" "${OUT_DIR}/app-seeder"

[ -f "${OUT_DIR}/app.js" ]            || die "app.js tidak ada di bundle."
[ -f "${OUT_DIR}/app-main" ]          || die "app-main tidak ada di bundle."
[ -d "${OUT_DIR}/migrations" ]        || die "migrations/ tidak ada di bundle."
[ -d "${OUT_DIR}/storage" ]           || die "storage/ tidak ada di bundle."
[ -d "${OUT_DIR}/uploads" ]           || die "uploads/ tidak ada di bundle."

# ---------------------------------------------------------------------------
# 4. Zip (dengan satu folder utama di dalam arsip)
# ---------------------------------------------------------------------------
blue "== Zip =="
STAGE_DIR="$(mktemp -d)"
mv "${OUT_DIR}" "${STAGE_DIR}/${ARCHIVE_ROOT}"
( cd "${STAGE_DIR}" && zip -qr "${ZIP_PATH}" "${ARCHIVE_ROOT}" )
mv "${STAGE_DIR}/${ARCHIVE_ROOT}" "${OUT_DIR}"
rm -rf "${STAGE_DIR}"

green ""
green "Selesai."
green "  Bundle : ${OUT_DIR}"
green "  Zip    : ${ZIP_PATH}"
green "Isi zip diextract menjadi folder '${ARCHIVE_ROOT}/'."
green "Upload ${ZIP_PATH##*/} ke direktori home (~) cPanel, extract, buat .env,"
green "lalu jalankan migration/seeder melalui Terminal dan RESTART app."
