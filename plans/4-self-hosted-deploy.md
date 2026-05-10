# Move Deploy Automation To Self-Hosted Runner

**Issue:** #4
**Branch:** feat/dbrowning2-self-hosted-deploy
**Status:** ready-for-pr

## Objective

Move codifyworx/ip deployment out of the GitHub-hosted test job and into a separate self-hosted deployment job that follows the PAWS SSH secret contract and refuses to overwrite a dirty live checkout.

## Steps

- [x] Split deploy into a self-hosted workflow.
- [x] Keep tests and container security as deploy prerequisites.
- [x] Preserve `PUBLISH_SSH_KEY_BASE64` preferred / `PUBLISH_SSH_KEY` fallback.
- [x] Replace live `git reset --hard` with a dirty-checking deploy path.
- [x] Validate workflow syntax and whitespace.
- [x] Keep CI green when no self-hosted runner is registered.

## Decisions

- Do not manually deploy as part of this work.
- Keep CI responsible for tests and Trivy.
- Trigger deploy after a successful CI run on `main` or by manual dispatch.
- Require `SELF_HOSTED_DEPLOY=true` before scheduling the self-hosted deploy job.

## Where I Left Off

Last updated: 2026-05-09
Deploy was split from CI so the repository can keep passing CI before a self-hosted runner is registered.
