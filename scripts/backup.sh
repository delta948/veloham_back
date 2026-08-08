#!/bin/sh
set -eu

backup_dir=${1:-./backups}
case "$backup_dir" in
  /|""|.) echo "Choose a dedicated backup directory" >&2; exit 1 ;;
esac
mkdir -p "$backup_dir"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
compose="docker compose --env-file .env.production -f docker-compose.prod.yml"

$compose exec -T postgres pg_dump -U veloham -d veloham -Fc > "$backup_dir/database-$stamp.dump"
$compose run --rm -T --no-deps -v "$backup_dir:/backup" backend sh -c "tar -czf /backup/uploads-$stamp.tar.gz -C /app/uploads ."
echo "Backup created in $backup_dir"
