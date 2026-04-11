#!/usr/bin/env sh
set -eu

app_dir="${APP_DIR:-/app/ip}"
health_url="${HEALTH_URL:-http://127.0.0.1:3010/ip/healthz}"

cd "$app_dir"
./scripts/update-dbip-lite.sh geoip
docker compose up -d --build ip
curl -fsS "$health_url"
