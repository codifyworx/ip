# Project Instructions

## Codex

Codex behavior for a project is defined by that repo's `.codex/config.toml`.
If you need different permissions or sandboxing for a session, use the appropriate Codex profile or local override.

## Current Work

**Active:**

**Paused:**

## Workflow

### Branch Structure

- `main` — stable releases only. Always deployable.
- `develop` — beta/integration branch. Features accumulate here and are tested together before promotion to main.
- `bug/<username>-<desc>` or `feat/<username>-<desc>` — individual work branches, created off `develop`

### Issues
- All bugs and features must have a GitHub issue (prefixed `bug:` or `feat:`)
- Open an issue before starting work; close it when the work is merged
- Do not leave issues open indefinitely — if work is abandoned, close the issue with a note explaining why

### Development Flow

1. **Open a GitHub issue** (`bug:` or `feat:` prefix)
2. **Create a branch off `develop`**: `feat/<username>-<desc>` or `bug/<username>-<desc>`
3. **Create a plan** in `plans/<issue-number>-<desc>.md` for features. Bug fixes can use the GitHub issue as the plan.
4. **Develop on the branch**
5. **Squash merge into `develop`** with a single descriptive commit referencing the issue
6. **Delete the feature/bug branch** after merge
7. **Close the GitHub issue** after merge
8. **Cut beta releases from `develop`** (e.g., `1.2.0-beta.1`, `1.2.0-beta.2`) for testing
9. **When stable, squash merge `develop` into `main`** and tag a stable release

### Multi-Agent Worktrees

- If more than one AI agent is working on the repo, each agent must use a unique git worktree.
- Do not run multiple agents in the same checked-out directory.
- Recommended pattern:

```sh
git worktree add ../<repo>-codex -b codex/main
git worktree add ../<repo>-claude -b claude/main
```

- Parallel work should flow through branches, commits, diffs, and merges, not through multiple agents sharing one working tree.

## Engineering Standards

### Standardize By Composition

- When a UI pattern, behavior, API shape, validation rule, or workflow becomes a project standard, implement it as a reusable component, helper, schema, or shared test fixture.
- Prefer consuming shared primitives over copying markup, styling, event handling, formatting, validation, or API glue into each screen/module.
- Treat duplicated implementations of a project standard as a review concern even when the current behavior looks correct.
- If a standard primitive cannot fit a specific use case, extend the primitive or document why the local implementation must differ.
- For UI work, standardized surfaces should compose shared pieces for artwork, primary actions, status badges, ratings, progress, empty/loading/error states, and accessibility behavior instead of hand-rolling lookalike cards or controls.

### Dependency Hygiene

- Use the latest stable package versions available from the upstream package manager when adding or refreshing dependencies.
- Do not introduce stale pins, abandoned packages, pre-release builds, nightly builds, or canary builds unless the issue documents the reason and expected removal path.
- Treat known-vulnerable direct dependencies as review blockers. Update to the latest stable fixed release instead of carrying a vulnerable version forward.
- For vulnerable transitive dependencies, prefer updating the direct dependency or lockfile to a fixed stable version. Overrides are only a documented bridge when no fixed parent release exists.
- Lockfiles are for reproducibility, not for preserving old vulnerable versions after a stable fixed release exists.
- If compatibility forces a non-latest or vulnerable package version, document the exception with the risk, mitigation, owner, and follow-up trigger.

### Container Security Scanning

- Repos that build containers must run Trivy image scans in CI before images are deployed, published, or treated as release candidates.
- Every Dockerfile, generated container, CI helper image, runner image, validation image, and deployable runtime image must be represented in the scan matrix.
- Trivy scans should include OS and library vulnerabilities and fail on fixed `HIGH` and `CRITICAL` findings.
- Ignore unfixed vulnerabilities by default to avoid vendor-feed noise, but do not ignore fixed vulnerable packages.
- Use the latest stable Trivy action and current stable Docker GitHub Actions when adding or refreshing the workflow.
- CLAWS or other deployed-state scanners still matter. Trivy is the pre-merge/pre-deploy guardrail; deployed scanners confirm what is actually running.
- When a new Dockerfile or image-producing workflow is added, updating the Trivy matrix is part of the same change.

### Trusted CI Containers

- On codifyworx self-hosted runners, any CI harness image, runner image, validation image, Compose file, Docker-running script, or host-side script that CI creates or runs must be owned by the trusted `github-runners` infrastructure repo.
- Application repositories may submit CI requests, app source, test source, and ordinary non-privileged commands, but they must not define the Dockerfile, Compose file, bind mounts, privileges, Docker socket access, or host paths used by trusted-host CI.
- Production or local-development Dockerfiles may remain in application repos, but trusted-runner workflows must access them only through `github-runners` scripts with explicit allowlists. Repo workflows must not run arbitrary Docker build/run/Compose commands on the trusted host.
- Workflows should target stable org-level runner labels such as `codifyworx-ci`, with optional capability labels like `android`; do not bake hostnames such as `nuc01` into app repo workflow labels.
- The live `github-runners` repo is trusted infrastructure. It is writable only by the infrastructure owner; other developers request changes through issues or reviewed patches.

## Safety

### Destructive Actions
- Always ask before deleting anything, including uninstalling apps or removing user data
- Never reboot devices or restart critical services without explicit permission
- Never push to main/develop or merge without explicit permission

### Secret Material (Hard Deny)
- Never read, print, cat, grep, copy, or summarize contents of private keys or secret files
- Forbidden patterns: `*.key`, `*.pem`, `*.p12`, `*.kdbx`, `*.jks`, `.env*`, `*/secrets/*`, `*/age-keys/*`
- Allowed: reference path existence and permissions metadata (`ls -l`, `stat`) without opening contents
- If a task requires reading a secret file, stop and ask the user to provide the needed value
- Never commit secrets, credentials, keys, tokens to the repository

### Device & Hardware
- All app fixes must be done in code and deployed via a build. Device recovery (removing a bricked module, clearing boot-blocking state) is allowed when the device cannot boot.
- Always force-stop the app before installing a new APK

### Android Build & Deploy
- **One package name**: No `applicationIdSuffix`. Debug and release share the same application ID.
- **One signing key**: Debug and release use the same signing config. Release keystore available locally (gitignored `local.properties`) and on CI (via secrets).
- **All builds through CI**: `beta.N` tags build debug (debuggable). Stable tags build release (minified). Local builds are for compile-checking only.
- **Updates via the app**: Use the in-app update flow. Direct `adb install` from CI artifacts is the escape hatch when the app is broken — discuss first.
- **Signing key changes for system apps**: Never replace a system priv-app APK (Magisk module overlay) with a differently-signed APK and reboot. Android caches the package signature — a mismatch hangs the device at the boot animation.
- **Magisk module identity**: A Magisk module has one APK slot and one matching permissions XML. Do not alternate between `.debug` and release package names in that slot.
- **Org package naming**: If the repo lives under `codifyworx`, prefer Android identifiers under `dev.codifyworx.*` rather than personal namespaces.
- **OTA/update safety model**: For risky install flows, cancellation must be real cancellation, rollback must complete before the UI returns to normal, and required verification failures must block any “ready to reboot” state.

## AI Session Rules

- Updates to CLAUDE.md may be made directly on main (exception to branch rules)
- Don't leave uncommitted changes on develop
- `codex.out.md` and `claude.out.md` are scratch review artifacts and should remain gitignored

## Project-Specific Notes

<!-- Add project-specific instructions below -->
