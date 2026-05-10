# Move Deploy Automation To Self-Hosted Runner

**Issue:** #4
**Branch:** feat/dbrowning2-self-hosted-deploy
**Status:** ready-for-pr

## Objective

Move codifyworx/ip deployment out of the GitHub-hosted test job and into a separate self-hosted deployment job that follows the PAWS SSH secret contract and refuses to overwrite a dirty live checkout.

## Steps

- [x] Split deploy into a self-hosted workflow job.
- [x] Keep tests and container security as deploy prerequisites.
- [x] Preserve `PUBLISH_SSH_KEY_BASE64` preferred / `PUBLISH_SSH_KEY` fallback.
- [x] Replace live `git reset --hard` with a dirty-checking deploy path.
- [x] Validate workflow syntax and whitespace.

## Decisions

- Do not manually deploy as part of this work.
- Keep existing `main`, schedule, and manual workflow triggers.

## Where I Left Off

Last updated: 2026-05-09
Workflow patch is in place and validation passed; PR is next.
