#!/bin/sh
set -e

# =============================================================================
# Ruang Tenang API Entrypoint
# =============================================================================
# Environment variables for database operations:
#   RUN_MIGRATE=true     - Run database migrations before starting
#   RUN_MIGRATE_FRESH=true - Drop all tables and re-migrate (WARNING: destroys data)
#   RUN_SEEDER=true      - Run database seeders
# =============================================================================

echo "🚀 Starting Ruang Tenang API..."

# Database migrations
if [ "$RUN_MIGRATE_FRESH" = "true" ]; then
  echo "⚠️  WARNING: Running fresh migration - this will destroy all data!"
  ./migrate fresh
  if [ "$RUN_SEEDER" = "true" ]; then
    echo "🌱 Running database seeders..."
    ./seeder-prod
  fi
elif [ "$RUN_MIGRATE" = "true" ]; then
  echo "📦 Running database migrations..."
  ./migrate up
  if [ "$RUN_SEEDER" = "true" ]; then
    echo "🌱 Running database seeders..."
    ./seeder-prod
  fi
elif [ "$RUN_SEEDER" = "true" ]; then
  echo "🌱 Running database seeders..."
  ./seeder-prod
fi

echo "✅ Starting server..."
exec "$@"
