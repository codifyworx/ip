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
curl http://127.0.0.1:8080/ip/json
```

## GeoIP

GeoIP is optional. Without databases, the service still returns IP and header data.

Mount MaxMind databases at:

```text
geoip/GeoLite2-City.mmdb
geoip/GeoLite2-ASN.mmdb
```

The compose file sets:

```text
GEOIP_CITY_DB=/geoip/GeoLite2-City.mmdb
GEOIP_ASN_DB=/geoip/GeoLite2-ASN.mmdb
```

Do not commit `.mmdb` files.

## Reverse Proxy

For `https://codifyworx.com/ip/`, route `/ip` and `/ip/*` to the container on port `8080`.

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
