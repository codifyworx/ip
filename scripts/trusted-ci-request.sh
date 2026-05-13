#!/usr/bin/env bash
set -euo pipefail

ci_root="${CODIFYWORX_CI_ROOT:-/app/github-runners}"
request_dir="${CODIFY_IP_CI_REQUEST_DIR:-${ci_root}/bridges/ip/requests}"
result_dir="${CODIFY_IP_CI_RESULT_DIR:-${ci_root}/bridges/ip/results}"
mode="${CODIFY_IP_VALIDATION_MODE:-container-security}"
timeout_seconds="${CODIFY_IP_VALIDATION_TIMEOUT_SECONDS:-3600}"
poll_seconds="${CODIFY_IP_VALIDATION_POLL_SECONDS:-5}"

repo="${GITHUB_REPOSITORY:-codifyworx/ip}"
sha="${CODIFY_IP_REQUEST_SHA:-${GITHUB_SHA:-}}"
ref="${CODIFY_IP_REQUEST_REF:-${GITHUB_REF:-}}"
run_id="${GITHUB_RUN_ID:-manual}"
run_attempt="${GITHUB_RUN_ATTEMPT:-1}"

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

require_command jq

if [ -n "${GITHUB_ACTIONS:-}" ]; then
  [ -n "${CODIFY_IP_CI_REQUEST_DIR:-}" ] || die "CODIFY_IP_CI_REQUEST_DIR is missing; this job is not running on the trusted request runner."
  [ -n "${CODIFY_IP_CI_RESULT_DIR:-}" ] || die "CODIFY_IP_CI_RESULT_DIR is missing; this job is not running on the trusted request runner."
fi

[[ "${sha}" =~ ^[0-9a-fA-F]{40}$ ]] || die "GITHUB_SHA must be a 40-character hex SHA."
[[ -n "${ref}" ]] || die "GITHUB_REF is required."
[[ "${run_id}" =~ ^[0-9]+$ ]] || die "GITHUB_RUN_ID must be numeric."
[[ "${run_attempt}" =~ ^[0-9]+$ ]] || die "GITHUB_RUN_ATTEMPT must be numeric."
case "${mode}" in
  container-security|deploy-main) ;;
  *) die "CODIFY_IP_VALIDATION_MODE must be container-security or deploy-main." ;;
esac

mkdir -p "${request_dir}" "${result_dir}"
[ -w "${request_dir}" ] || die "Request directory is not writable by the runner: ${request_dir}"
[ -r "${result_dir}" ] || die "Result directory is not readable by the runner: ${result_dir}"

request_id="${run_id}-${run_attempt}-${mode}-${sha}"
request_path="${request_dir}/${request_id}.json"
tmp_path="${request_path}.tmp.$$"
result_path="${result_dir}/${request_id}.result.json"

jq -n \
  --arg repo "${repo}" \
  --arg sha "${sha}" \
  --arg ref "${ref}" \
  --arg run_id "${run_id}" \
  --arg run_attempt "${run_attempt}" \
  --arg mode "${mode}" \
  '{
    schema: 1,
    repo: $repo,
    sha: $sha,
    ref: $ref,
    run_id: $run_id,
    run_attempt: $run_attempt,
    mode: $mode
  }' > "${tmp_path}"
mv "${tmp_path}" "${request_path}"

echo "Queued trusted ${mode} request: ${request_path}"
echo "Waiting for result: ${result_path}"

deadline=$((SECONDS + timeout_seconds))
while [ "${SECONDS}" -lt "${deadline}" ]; do
  if [ -s "${result_path}" ]; then
    status="$(jq -r '.status // "unknown"' "${result_path}")"
    reason="$(jq -r '.reason // ""' "${result_path}")"
    artifact_path="$(jq -r '.artifact_path // ""' "${result_path}")"
    echo "trusted validation status: ${status}"
    if [ -n "${reason}" ]; then
      echo "reason: ${reason}"
    fi
    if [ -n "${artifact_path}" ]; then
      echo "artifact_path=${artifact_path}"
    fi
    jq . "${result_path}"
    [ "${status}" = "success" ]
    exit $?
  fi
  sleep "${poll_seconds}"
done

die "Timed out waiting for trusted ${mode} result after ${timeout_seconds}s."
