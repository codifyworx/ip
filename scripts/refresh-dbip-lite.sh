#!/usr/bin/env sh
set -eu

app_dir="${APP_DIR:-/app/ip}"
codify_health_url="${CODIFY_HEALTH_URL:-http://127.0.0.1:3010/ip/healthz}"
ifconfig_health_url="${IFCONFIG_HEALTH_URL:-http://127.0.0.1:3011/healthz}"

cd "$app_dir"
./scripts/update-dbip-lite.sh geoip
docker compose up -d --build
curl -fsS "$codify_health_url"
curl -fsS "$ifconfig_health_url"
