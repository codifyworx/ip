# Local Self-Hosted Deploy

**Issue:** #6
**Branch:** fix/dbrowning2-local-self-hosted-deploy
**Status:** ready-for-pr

## Objective

Make deployment run locally on the self-hosted runner instead of requiring SSH credentials from a self-hosted runner to the same publish host.

## Steps

- [x] Remove SSH credential requirements from the deploy workflow.
- [x] Deploy from the local live checkout path.
- [x] Keep dirty checkout protection.
- [x] Validate workflow syntax and whitespace.

## Decisions

- Do not manually deploy as part of this work.
- Keep deploy gated by `SELF_HOSTED_DEPLOY=true`.
- Default the live checkout path to `/app/ip`.
