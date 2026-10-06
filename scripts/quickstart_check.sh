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

echo "Step 1/5: Verify env file"
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

HEALTH_PORT="${PORT:-8080}"

echo "Step 2/5: Verify DB credentials and MySQL availability"
# DATABASE_URL is the single source of truth; derive connection parts from it.
DB_USER=""; DB_PASSWORD=""; DB_NAME=""; DB_HOST="127.0.0.1"; DB_PORT="3306"
if [[ -z "${DATABASE_URL:-}" ]]; then
  fail "DATABASE_URL is not set in ${ENV_FILE}"
else
  pass "DATABASE_URL is set"
  _u="${DATABASE_URL#mysql://}"
  _creds="${_u%%@*}"
  _rest="${_u#*@}"
  DB_USER="${_creds%%:*}"
  [[ "${_creds}" == *:* ]] && DB_PASSWORD="${_creds#*:}"
  _hostport="${_rest#*tcp(}"; _hostport="${_hostport%%)*}"
  DB_HOST="${_hostport%%:*}"
  DB_PORT="${_hostport##*:}"
  DB_NAME="${_rest#*/}"; DB_NAME="${DB_NAME%%\?*}"
fi

if command -v mysqladmin >/dev/null 2>&1; then
  if MYSQL_PWD="${DB_PASSWORD:-}" mysqladmin ping -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" --silent >/dev/null 2>&1; then
    pass "MySQL responds on ${DB_HOST}:${DB_PORT}"
    if command -v mysql >/dev/null 2>&1; then
      if MYSQL_PWD="${DB_PASSWORD:-}" mysql -h "${DB_HOST}" -P "${DB_PORT}" -u "${DB_USER}" "${DB_NAME}" -e "SELECT 1" >/dev/null 2>&1; then
        pass "MySQL credential test (SELECT 1)"
      else
        fail "MySQL credential test (SELECT 1)"
      fi
    else
      fail "mysql CLI is not installed"
    fi
  else
    fail "MySQL responds on ${DB_HOST}:${DB_PORT}"
  fi
else
  fail "mysqladmin CLI is not installed (install mysql-client)"
fi

echo "Step 3/5: Run migrations"
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
if curl -fsS "http://localhost:${HEALTH_PORT}/health" >/dev/null 2>&1; then
  pass "Server already healthy on :${HEALTH_PORT}"
else
  echo "Starting temporary server for health check on :${HEALTH_PORT}"
  go run ./cmd/server/main.go >"${server_log}" 2>&1 &
  server_pid=$!
  created_server=1

  if curl -fsS --retry 20 --retry-delay 1 --retry-connrefused "http://localhost:${HEALTH_PORT}/health" >/dev/null 2>&1; then
    pass "Temporary server healthy on :${HEALTH_PORT}"
  else
    fail "Temporary server failed health check on :${HEALTH_PORT}"
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
