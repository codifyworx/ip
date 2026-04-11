#!/usr/bin/env sh
set -eu

out_dir="${1:-geoip}"
dbip_date="${DBIP_DATE:-$(date -u +%Y-%m)}"
base_url="${DBIP_BASE_URL:-https://download.db-ip.com/free}"

mkdir -p "$out_dir"
tmp_dir="$(mktemp -d "${out_dir}/.dbip.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

download_db() {
  name="$1"
  target="$2"
  url="${base_url}/dbip-${name}-lite-${dbip_date}.mmdb.gz"
  archive="${tmp_dir}/dbip-${name}-lite.mmdb.gz"
  output="${tmp_dir}/${target}"

  echo "Downloading ${url}"
  curl -fsSL "$url" -o "$archive"
  gzip -dc "$archive" > "$output"
  mv "$output" "${out_dir}/${target}"
}

download_db city dbip-city-lite.mmdb
download_db asn dbip-asn-lite.mmdb

echo "Updated DB-IP Lite databases in ${out_dir}"
