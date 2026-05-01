#!/bin/sh
set -e

# =============================================================================
# Ruang Tenang API Entrypoint
# =============================================================================
# Environment variables for database operations:
#   RUN_MIGRATE=true     - Run database migrations before starting
#   RUN_MIGRATE_FRESH=true - Drop all tables and re-migrate (WARNING: destroys data)
#   RUN_SEEDER=true      - Run the single presentation database seeder
# =============================================================================

echo "🚀 Starting Ruang Tenang API..."

SEEDER_BIN="./seeder"

# Database migrations
if [ "$RUN_MIGRATE_FRESH" = "true" ]; then
  echo "⚠️  WARNING: Running fresh migration - this will destroy all data!"
  ./migrate fresh
  if [ "$RUN_SEEDER" = "true" ]; then
    echo "🌱 Running database seeders..."
    "$SEEDER_BIN"
  fi
elif [ "$RUN_MIGRATE" = "true" ]; then
  echo "📦 Running database migrations..."
  ./migrate up
  if [ "$RUN_SEEDER" = "true" ]; then
    echo "🌱 Running database seeders..."
    "$SEEDER_BIN"
  fi
elif [ "$RUN_SEEDER" = "true" ]; then
  echo "🌱 Running database seeders..."
  "$SEEDER_BIN"
fi

echo "✅ Starting server..."
exec "$@"
