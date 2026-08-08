#!/bin/sh
set -eu

origin=${1:-}
if [ -z "$origin" ]; then
  echo "Usage: $0 https://your-domain.example" >&2
  exit 1
fi
origin=${origin%/}
case "$origin" in
  https://*) ;;
  *) echo "Production smoke test requires an HTTPS URL" >&2; exit 1 ;;
esac

curl --fail --silent --show-error "$origin/healthz" >/dev/null
curl --fail --silent --show-error "$origin/" >/dev/null
curl --fail --silent --show-error "$origin/api/v1/categories" >/dev/null
echo "Smoke test passed: $origin"
