# Update Container Security and PAWS Standards

**Issue:** #1
**Branch:** feat/dbrowning2-container-security-paws
**Status:** in-progress

## Objective

Update the codifyworx/ip container build so the deployed image is rebuilt with the current stable Go toolchain, then bring the repo up to the current PAWS project instructions and container scanning standard.

## Steps

- [x] Apply current PAWS project files.
- [x] Add the PAWS container-security workflow and repo image matrix.
- [x] Update the Go build/runtime image to clear fixed Go vulnerabilities.
- [x] Run formatting, tests, workflow validation, container build, and Trivy scan.
- [ ] Commit, push, merge, and redeploy `/app/ip` on codifyworx.com.

## Decisions

- This repo currently deploys from `main`, so the feature branch is based on `main`.
- Use PAWS wording verbatim for `CLAUDE.md`/`AGENTS.md`.
- Use current stable Go channels in Docker and CI instead of patch-level toolchain pins; Trivy enforces fixed high/critical findings.

## Where I Left Off

Last updated: 2026-05-09
Implementation validated locally. Next step is commit, push, merge, deploy, and verify the running containers.
