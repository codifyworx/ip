# Deployment

This service runs from `/app/ip` on `codifyworx.com` and serves two public entrypoints:

- `https://codifyworx.com/ip/` via `codify-ip` on `127.0.0.1:3010`
- `https://ifconfig.fyi/` via `ifconfig-fyi` on `127.0.0.1:3011`

Both containers are built from the same source tree and image. The only functional difference is `BASE_PATH`:

- `codify-ip`: `BASE_PATH=/ip`
- `ifconfig-fyi`: `BASE_PATH=/`

## Host Prerequisites

The host needs:

- nginx
- outbound access to:
  - `github.com`
  - `download.db-ip.com`

Container deployment, the trusted Compose file, and the monthly DB-IP refresh timer are owned by `codifyworx/github-runners`.
The application checkout lives at `/app/ip`, but the app repo must not own production Compose execution.

## First-Time Host Bootstrap

Bootstrap the trusted runner repo first:

```bash
mkdir -p /app
git clone git@github.com:codifyworx/github-runners.git /app/github-runners
cd /app/github-runners
```

Prepare the IP bridge/source paths and install the trusted refresh timer:

```bash
sudo projects/container-security/scripts/prepare-host-paths.sh ip
sudo cp systemd/codify-ip-geoip-update.service /etc/systemd/system/
sudo cp systemd/codify-ip-geoip-update.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now codify-ip-geoip-update.timer
projects/ip/scripts/refresh-dbip-lite.sh
```

The trusted deploy creates or updates `/app/ip`, downloads DB-IP Lite data, builds the image, and starts:

- `codify-ip` listening on `127.0.0.1:3010`
- `ifconfig-fyi` listening on `127.0.0.1:3011`

## Rate Limiting

The root-hosted `ifconfig.fyi` service enables app-level fixed-window rate limiting by default:

- 20 requests per minute per client IP
- no burst allowance
- no separate concurrent-request allowance
- `/healthz` is excluded

Excess requests receive `429 Too Many Requests` with `Retry-After`. The response text notes that users behind CGNAT, VPNs, corporate proxies, or shared NAT may share the same public IP and therefore the same quota.

The default applies when `BASE_PATH=/`. Set `RATE_LIMIT_REQUESTS_PER_MINUTE` to a non-negative integer to override it; `0` disables the limiter.

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

## CI Deploy

`main` deploys through GitHub Actions in [.github/workflows/deploy.yml](.github/workflows/deploy.yml).

The CI deploy job:

1. runs on the repo-scoped `codify-ip-ci` request runner
2. writes a `deploy-main` request with the exact `main` SHA that passed CI
3. waits for the trusted host-side harness in `codifyworx/github-runners`

The trusted harness verifies `codifyworx/ip`, `refs/heads/main`, and the exact SHA before updating `/app/ip`.
It uses trusted runner-repo code for DB-IP downloads and trusted runner-repo Compose for container rebuild/restart.

## Manual Update

For a manual production update, run the trusted refresh/deploy harness from the runner repo:

```bash
cd /app/github-runners
projects/ip/scripts/refresh-dbip-lite.sh
```

For local development only, use the app repo Compose file on a development machine:

```bash
./scripts/update-dbip-lite.sh geoip
docker compose up -d --build
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
