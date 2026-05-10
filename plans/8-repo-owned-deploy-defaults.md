# Repo-Owned Deploy Defaults

**Issue:** #8
**Branch:** fix/dbrowning2-repo-owned-deploy-defaults
**Status:** ready-for-pr

## Objective

Make the standard self-hosted deployment behavior part of the repository workflow instead of requiring GitHub variables for normal operation.

## Steps

- [x] Enable deploy by default.
- [x] Keep `SELF_HOSTED_DEPLOY=false` as an explicit pause switch.
- [x] Define `DEPLOY_PATH=/app/ip` in workflow code.
- [x] Validate workflow syntax and whitespace.

## Decisions

- Do not manually deploy as part of this work.
- The production path for this repo is `/app/ip`.
