#!/usr/bin/env bash
set -u

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}"

ENV_FILE="${ENV_FILE:-.env}"
failures=0
created_server=0
server_pid=""
server_log="/tmp/ruang_tenang_api_quickstart_server.log"

pass() {
  printf 'PASS | %s\n' "$1"
}

fail() {
  printf 'FAIL | %s\n' "$1"
  failures=$((failures + 1))
}

run_step() {
  local label="$1"
  shift
  if "$@"; then
    pass "${label}"
    return 0
  fi

  fail "${label}"
  return 1
}

cleanup() {
  if [[ "${created_server}" -eq 1 && -n "${server_pid}" ]] && kill -0 "${server_pid}" >/dev/null 2>&1; then
    kill "${server_pid}" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT

echo "Backend quickstart verification"
echo "Project dir: ${ROOT_DIR}"

echo "Step 1/6: Verify env file"
if [[ -f "${ENV_FILE}" ]]; then
  pass "Env file exists (${ENV_FILE})"
else
  fail "Env file missing (${ENV_FILE})"
  echo
  echo "Summary: ${failures} issue(s) found"
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "${ENV_FILE}"
set +a

APP_PORT="${APP_PORT:-8080}"

echo "Step 2/6: Verify DB credentials and PostgreSQL availability"
if [[ -n "${DB_HOST:-}" && -n "${DB_PORT:-}" && -n "${DB_USER:-}" && -n "${DB_PASSWORD:-}" && -n "${DB_NAME:-}" ]]; then
  pass "DB credential fields are set in ${ENV_FILE}"
else
  fail "DB credential fields are incomplete in ${ENV_FILE}"
fi

if run_step "PostgreSQL responds on ${DB_HOST:-localhost}:${DB_PORT:-5432}" pg_isready -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -t 3; then
  if command -v psql >/dev/null 2>&1; then
    if env PGPASSWORD="${DB_PASSWORD}" psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -tAc "select 1" >/dev/null 2>&1; then
      pass "PostgreSQL credential test (select 1)"
    else
      fail "PostgreSQL credential test (select 1)"
    fi
  else
    fail "psql CLI is not installed"
  fi
fi

echo "Step 3/6: Run migrations"
if command -v migrate >/dev/null 2>&1; then
  pass "migrate CLI is installed"
  if make migrate-up >/tmp/ruang_tenang_migrate_up.log 2>&1; then
    pass "make migrate-up"
  else
    fail "make migrate-up"
    tail -n 20 /tmp/ruang_tenang_migrate_up.log
  fi
else
  fail "migrate CLI is not installed (run: make install-tools)"
fi

echo "Step 4/5: Run presentation seeder"
if make seed >/tmp/ruang_tenang_seed.log 2>&1; then
  pass "make seed"
else
  fail "make seed"
  tail -n 20 /tmp/ruang_tenang_seed.log
fi

echo "Step 5/5: Verify server on target port"
if curl -fsS "http://localhost:${APP_PORT}/health" >/dev/null 2>&1; then
  pass "Server already healthy on :${APP_PORT}"
else
  echo "Starting temporary server for health check on :${APP_PORT}"
  go run ./cmd/server/main.go >"${server_log}" 2>&1 &
  server_pid=$!
  created_server=1

  if curl -fsS --retry 20 --retry-delay 1 --retry-connrefused "http://localhost:${APP_PORT}/health" >/dev/null 2>&1; then
    pass "Temporary server healthy on :${APP_PORT}"
  else
    fail "Temporary server failed health check on :${APP_PORT}"
    echo "--- server log (tail) ---"
    tail -n 40 "${server_log}" || true
  fi
fi

echo
if [[ "${failures}" -gt 0 ]]; then
  echo "Summary: ${failures} issue(s) found"
  exit 1
fi

echo "Summary: all quickstart checks passed"
