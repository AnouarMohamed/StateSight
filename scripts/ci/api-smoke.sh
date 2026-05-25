#!/usr/bin/env bash
set -euo pipefail

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${REDIS_URL:?REDIS_URL is required}"

command -v curl >/dev/null || { printf 'curl is required\n' >&2; exit 127; }
command -v jq >/dev/null || { printf 'jq is required\n' >&2; exit 127; }

api_port="${API_PORT:-18080}"
base_url="http://127.0.0.1:${api_port}"
application_id="44444444-4444-4444-4444-444444444444"
rule_name="CI resource scope ${GITHUB_RUN_ID:-local}-${RANDOM}"
api_log="$(mktemp)"
api_pid=""

cleanup() {
  if [[ -n "${api_pid}" ]] && kill -0 "${api_pid}" 2>/dev/null; then
    kill "${api_pid}"
    wait "${api_pid}" || true
  fi
  rm -f "${api_log}"
}
trap cleanup EXIT

go run ./scripts/migrate
go run ./scripts/migrate
go run ./scripts/seed

API_PORT="${api_port}" AUTH_REQUIRED=false go run ./apps/api >"${api_log}" 2>&1 &
api_pid=$!

for _ in {1..30}; do
  if curl --fail --silent "${base_url}/readyz" >/dev/null; then
    break
  fi
  if ! kill -0 "${api_pid}" 2>/dev/null; then
    cat "${api_log}" >&2
    exit 1
  fi
  sleep 1
done

curl --fail --silent "${base_url}/readyz" >/dev/null || {
  cat "${api_log}" >&2
  exit 1
}

rule_response="$(
  curl --fail --silent --show-error \
    --request POST "${base_url}/api/v1/applications/${application_id}/ignore-rules" \
    --header "Content-Type: application/json" \
    --data "{\"name\":\"${rule_name}\",\"match_expression\":\"spec.replicas\",\"resource_ref\":\"apps/v1/Deployment:payments/ledger-api\",\"reason\":\"CI smoke verification\"}"
)"
rule_id="$(jq --exit-status --raw-output '.data.id' <<<"${rule_response}")"
jq --exit-status \
  --arg app "${application_id}" \
  '.success and .data.application_id == $app and .data.active == true' \
  <<<"${rule_response}" >/dev/null

details_response="$(curl --fail --silent --show-error "${base_url}/api/v1/applications/${application_id}")"
jq --exit-status --arg id "${rule_id}" \
  '.data.ignore_rules | any(.id == $id and .active == true)' \
  <<<"${details_response}" >/dev/null

update_response="$(
  curl --fail --silent --show-error \
    --request PATCH "${base_url}/api/v1/applications/${application_id}/ignore-rules/${rule_id}" \
    --header "Content-Type: application/json" \
    --data '{"active":false}'
)"
jq --exit-status '.success and .data.active == false' <<<"${update_response}" >/dev/null
