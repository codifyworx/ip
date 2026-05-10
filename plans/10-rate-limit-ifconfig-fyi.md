# Add rate limiting for ifconfig.fyi

**Issue:** #10
**Branch:** feat/dbrowning2-rate-limit-ifconfig-fyi
**Status:** completed

## Objective

Protect the public `ifconfig.fyi` entrypoint from low-effort abuse without making normal CLI use painful.

## Steps

- [x] Add a fixed-window per-client rate limit for root-hosted `ifconfig.fyi`.
- [x] Return HTTP 429 with `Retry-After` and shared-network/CGNAT context.
- [x] Document the configured limit and override behavior.
- [x] Validate tests, formatting, and repository diff hygiene.

## Decisions

- Use a fixed-window app-level limiter so the behavior is actually 20 requests per minute per client IP. nginx `limit_req` without burst is a leaky bucket and would reject normal quick successive CLI calls.
- Default rate limiting only for `BASE_PATH=/`, which is the `ifconfig.fyi` service. The path-based `codifyworx.com/ip` service remains unthrottled unless explicitly configured.
- Keep `/healthz` out of rate limiting so local health checks and deployment validation do not consume public request quota.

## Where I Left Off

Last updated: 2026-05-10
Implementation is complete and ready for review.
