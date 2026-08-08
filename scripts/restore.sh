#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 DATABASE.dump UPLOADS.tar.gz" >&2
  exit 1
fi
database_dump=$1
uploads_archive=$2
[ -f "$database_dump" ] && [ -f "$uploads_archive" ] || { echo "Backup file not found" >&2; exit 1; }

compose="docker compose --env-file .env.production -f docker-compose.prod.yml"
$compose exec -T postgres pg_restore -U veloham -d veloham --clean --if-exists < "$database_dump"
$compose run --rm -T --no-deps -v "$(cd "$(dirname "$uploads_archive")" && pwd):/restore:ro" backend sh -c "tar -xzf /restore/$(basename "$uploads_archive") -C /app/uploads"
echo "Restore completed"
