# codifyworx IP

Small `ifconfig.me`-style service for `https://codifyworx.com/ip/` and `https://ifconfig.fyi/`.

It returns the caller's public IP plus ISP, ASN, and geolocation metadata from DB-IP Lite databases.

See [DEPLOY.md](DEPLOY.md) for host bootstrap, nginx, timer, CI, and validation steps.

## Endpoints

```text
/ip/          HTML summary
/ip/ip        plain text IP
/ip/json      structured JSON
/ip/geo       structured JSON with geo fields when configured
/ip/headers   request headers as JSON
/ip/all       text summary
/ip/all.json  JSON summary
/ip/ua        user agent
/ip/lang      accept-language
/ip/encoding  accept-encoding
/ip/mime      accept header
/ip/forwarded X-Forwarded-For
/ip/healthz   health check
```

On `ifconfig.fyi`, the same handlers run at the domain root:

```text
/           HTML summary
/ip         plain text IP
/json       structured JSON
/geo        structured JSON with geo fields when configured
/headers    request headers as JSON
/all        text summary
/all.json   JSON summary
/healthz    health check
```

## Rate Limiting

`ifconfig.fyi` defaults to a fixed-window rate limit of 20 requests per minute per client IP. There is no burst allowance and no separate concurrent-request allowance. `/healthz` is not rate limited.

When a client exceeds the limit, the service returns `429 Too Many Requests` with `Retry-After` and a plain-text response noting that users behind CGNAT, VPNs, corporate proxies, or shared NAT may share a public IP.

The limit is enabled by default when `BASE_PATH=/`. Set `RATE_LIMIT_REQUESTS_PER_MINUTE` to a non-negative integer to override it; `0` disables the app-level limiter.

## Run Locally

```bash
go run .
curl http://127.0.0.1:8080/ip/ip
curl http://127.0.0.1:8080/ip/json
```

## Docker

```bash
docker compose up --build
curl http://127.0.0.1:3010/ip/json
```

## GeoIP

GeoIP is backed by DB-IP Lite MMDB databases. Without databases, the service still returns IP and header data but cannot show ISP, ASN, or location.

Download/update the monthly free databases:

```bash
./scripts/update-dbip-lite.sh geoip
```

Mount DB-IP databases at:

```text
geoip/dbip-city-lite.mmdb
geoip/dbip-asn-lite.mmdb
```

The compose file sets:

```text
GEOIP_CITY_DB=/geoip/dbip-city-lite.mmdb
GEOIP_ASN_DB=/geoip/dbip-asn-lite.mmdb
```

Do not commit `.mmdb` files.

The Lite databases are provided by DB-IP under CC BY 4.0. Keep the DB-IP attribution visible in the rendered page.

## Reverse Proxy

For `https://codifyworx.com/ip/`, route `/ip` and `/ip/*` to the container on `127.0.0.1:3010`.

Keep `BASE_PATH=/ip` unless the reverse proxy strips the `/ip` prefix before forwarding.

The service trusts forwarded client IP headers only when the socket peer is inside `TRUSTED_PROXY_CIDRS`. Set this to your Docker/proxy networks, for example:

```text
TRUSTED_PROXY_CIDRS=127.0.0.1/32,::1/128,172.16.0.0/12,10.0.0.0/8,192.168.0.0/16
```

## Traefik Example

```yaml
labels:
  - traefik.enable=true
  - traefik.http.routers.codify-ip.rule=Host(`codifyworx.com`) && PathPrefix(`/ip`)
  - traefik.http.routers.codify-ip.entrypoints=websecure
  - traefik.http.routers.codify-ip.tls.certresolver=letsencrypt
  - traefik.http.services.codify-ip.loadbalancer.server.port=8080
```

## Deployment

`main` CI deploys through the trusted codifyworx runner request bridge after tests pass.
The workflow writes a `deploy-main` request; host-side code in `codifyworx/github-runners` verifies the repo, branch, and exact SHA before touching `/app/ip`.

The trusted deploy harness updates DB-IP Lite databases under `/app/ip/geoip`, rebuilds from the production Dockerfile, restarts the containers with the trusted Compose file in the runner repo, and health-checks `/ip/healthz` and `/healthz`.
The trusted Compose file runs two services from the same image: `codify-ip` for path-based `codifyworx.com/ip/` on `127.0.0.1:3010`, and `ifconfig-fyi` for root-hosted `ifconfig.fyi/` on `127.0.0.1:3011`.
