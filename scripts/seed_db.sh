#!/usr/bin/env bash
set -euo pipefail

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL must be set"
  exit 1
fi

psql "$DATABASE_URL" -f internal/infrastructure/postgres/seed.sql

echo "seed applied"
