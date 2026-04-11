# codifyworx IP

Small `ifconfig.me`-style service for `https://codifyworx.com/ip/`.

It returns the caller's public IP plus optional ISP, ASN, and geolocation metadata when MaxMind databases are mounted.

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

`main` CI deploys to `/app/ip` on `codifyworx.com` after tests pass.
The deploy step updates DB-IP Lite databases under `/app/ip/geoip`, rebuilds the container, restarts it, and health-checks `/ip/healthz`.

It expects the same SSH secret pattern used by other Codifyworx projects:

```text
PUBLISH_SSH_KEY_BASE64 or PUBLISH_SSH_KEY
```

Optional repo/org variables:

```text
PUBLISH_HOST=codifyworx.com
PUBLISH_USER=root
```
