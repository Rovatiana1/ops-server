#!/usr/bin/env bash
# migrate.sh — applies SQL migrations to PostgreSQL
# Usage: ./migrate.sh [up|down]

set -euo pipefail

DIRECTION="${1:-up}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load .env if present
if [ -f "${SCRIPT_DIR}/../../.env" ]; then
  export "$(grep -v '^#' "${SCRIPT_DIR}/../../.env" | xargs)"
fi

DB_HOST="${OPS_SERVER_DATABASE_HOST:-localhost}"
DB_PORT="${OPS_SERVER_DATABASE_PORT:-5432}"
DB_USER="${OPS_SERVER_DATABASE_USER:-postgres}"
DB_PASS="${OPS_SERVER_DATABASE_PASSWORD:-postgres}"
DB_NAME="${OPS_SERVER_DATABASE_NAME:-ops-server}"

export PGPASSWORD="$DB_PASS"

PSQL="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"

echo "── Migrations ($DIRECTION) ────────────────────────────────"
echo "   Host : $DB_HOST:$DB_PORT"
echo "   DB   : $DB_NAME"

if [ "$DIRECTION" = "up" ]; then
  for file in "$SCRIPT_DIR"/*.sql; do
    echo "  applying: $(basename "$file")"
    $PSQL -f "$file"
  done
  echo "✅  All migrations applied."

elif [ "$DIRECTION" = "down" ]; then
  echo "⚠️  Down migrations are not automated."
  echo "   Please apply rollback scripts manually."
  exit 1

else
  echo "Usage: $0 [up|down]"
  exit 1
fi
