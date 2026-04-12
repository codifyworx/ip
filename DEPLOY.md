# Deployment

This service runs from `/app/ip` on `codifyworx.com` and serves two public entrypoints:

- `https://codifyworx.com/ip/` via `codify-ip` on `127.0.0.1:3010`
- `https://ifconfig.fyi/` via `ifconfig-fyi` on `127.0.0.1:3011`

Both containers are built from the same source tree and image. The only functional difference is `BASE_PATH`:

- `codify-ip`: `BASE_PATH=/ip`
- `ifconfig-fyi`: `BASE_PATH=/`

## Host Prerequisites

The host needs:

- Docker with `docker compose`
- nginx
- git
- outbound access to:
  - `github.com`
  - `download.db-ip.com`

The deployment checkout lives at `/app/ip`.

## First-Time Host Bootstrap

Clone the repo:

```bash
mkdir -p /app
git clone https://github.com/codifyworx/ip.git /app/ip
cd /app/ip
```

Fetch the DB-IP Lite databases and start the containers:

```bash
./scripts/update-dbip-lite.sh geoip
docker compose up -d --build
```

This creates:

- `codify-ip` listening on `127.0.0.1:3010`
- `ifconfig-fyi` listening on `127.0.0.1:3011`

## nginx

`codifyworx.com/ip/` is expected to proxy `/ip` and `/ip/*` to `127.0.0.1:3010`.

For `ifconfig.fyi`, install the dedicated vhost:

```bash
cp deploy/nginx/ifconfig.fyi.conf /etc/nginx/sites-available/ifconfig.fyi.conf
ln -sf /etc/nginx/sites-available/ifconfig.fyi.conf /etc/nginx/sites-enabled/ifconfig.fyi.conf
mkdir -p /var/www/letsencrypt/.well-known/acme-challenge
nginx -t
systemctl reload nginx
```

The checked-in `deploy/nginx/ifconfig.fyi.conf` does two things:

- serves ACME challenges from `/var/www/letsencrypt`
- redirects normal HTTP traffic to HTTPS

It also defines the TLS vhost that proxies `ifconfig.fyi` to `127.0.0.1:3011`.

### Shared TLS Frontend

The current production host does not terminate web TLS directly in a normal `listen 443 ssl` site.

Instead:

- nginx `stream` listens on `:443`
- TURN traffic is split by ALPN
- normal web TLS is proxied to nginx `:8443`
- nginx web vhosts on `:8443` use `proxy_protocol`

That means the `ifconfig.fyi` HTTPS server block must match the existing host pattern:

```nginx
listen 8443 ssl http2 proxy_protocol;
listen [::]:8443 ssl http2 proxy_protocol;
```

Do not use `certbot --nginx` blindly on this host. It is safer to use `certbot certonly --webroot` and keep the shared ingress edits explicit.

### Let's Encrypt

With the ACME webroot location active on port 80, request the certificate:

```bash
certbot certonly --webroot \
  -w /var/www/letsencrypt \
  -d ifconfig.fyi \
  -d www.ifconfig.fyi \
  --cert-name ifconfig.fyi \
  --non-interactive \
  --agree-tos
```

The production cert paths are:

```text
/etc/letsencrypt/live/ifconfig.fyi/fullchain.pem
/etc/letsencrypt/live/ifconfig.fyi/privkey.pem
```

After adding or changing the `8443` TLS vhost, run:

```bash
nginx -t
systemctl reload nginx
```

If the host continues serving stale HTTP behavior after a reload, do a full restart:

```bash
systemctl restart nginx
```

## GeoIP Refresh Timer

Install the systemd unit and timer so DB-IP Lite updates happen monthly:

```bash
cp deploy/systemd/codify-ip-geoip-update.service /etc/systemd/system/
cp deploy/systemd/codify-ip-geoip-update.timer /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now codify-ip-geoip-update.timer
```

The timer runs `scripts/refresh-dbip-lite.sh`, which:

1. downloads the latest DB-IP Lite MMDB files into `/app/ip/geoip`
2. rebuilds the image
3. restarts the compose services
4. health-checks both listeners

## CI Deploy

`main` deploys through GitHub Actions in [.github/workflows/ci.yml](.github/workflows/ci.yml).

Required secret:

- `PUBLISH_SSH_KEY_BASE64` or `PUBLISH_SSH_KEY`

Optional repository or organization variables:

- `PUBLISH_HOST=codifyworx.com`
- `PUBLISH_USER=root`

The CI deploy job:

1. ensures `/app/ip` exists
2. clones the repo if needed
3. resets the checkout to `origin/main`
4. runs `./scripts/refresh-dbip-lite.sh`

Because CI does a hard reset to `origin/main`, any manual edits made directly on the host are temporary until they are committed and pushed.

## Manual Update

For a manual update from the host checkout:

```bash
cd /app/ip
git fetch origin main
git reset --hard origin/main
./scripts/refresh-dbip-lite.sh
```

For a local unpushed change, copy the changed files into `/app/ip`, then rebuild:

```bash
cd /app/ip
docker compose up -d --build
curl -fsS http://127.0.0.1:3010/ip/healthz
curl -fsS http://127.0.0.1:3011/healthz
```

## Validation

Host-local checks:

```bash
curl -fsS http://127.0.0.1:3010/ip/healthz
curl -fsS http://127.0.0.1:3011/healthz
curl -fsS http://127.0.0.1:3010/ip/ip
curl -fsS http://127.0.0.1:3011/ip
```

Public checks:

```bash
curl -I http://ifconfig.fyi/
curl -I http://www.ifconfig.fyi/
curl -I https://codifyworx.com/ip/
curl -I https://ifconfig.fyi/
curl -I https://www.ifconfig.fyi/
curl -fsS https://codifyworx.com/ip/json
curl -fsS https://ifconfig.fyi/json
```
