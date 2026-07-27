#!/usr/bin/env bash

resolve_playwright_owned_stack_env() {
  local root_dir="$1"

  if [[ -z "${root_dir}" ]]; then
    echo "resolve_playwright_owned_stack_env requires <repo_root>" >&2
    return 2
  fi

  if [[ "${CARTULARY_BROWSER_SERVICE_REQUIREMENT:-}" != "test-services" ]]; then
    echo "Playwright browser execution currently requires a managed test-services session" >&2
    return 2
  fi

  local expected_profile_id="${CARTULARY_BROWSER_RUNTIME_PROFILE_ID:-default}"
  PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$root_dir/tmp/node-runtime}"
  PLAYWRIGHT_OWNED_STACK_PNPM_BIN="${PNPM:-${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin/pnpm}"
  PLAYWRIGHT_OWNED_STACK_NODE_BIN="${NODE_BIN:-}"

  if [[ ! -x "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" ]]; then
    echo "repo-local pnpm was not found at ${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}; run make frontend-toolchain" >&2
    return 1
  fi

  if [[ -z "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" ]]; then
    if [[ -x "${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin/node" ]]; then
      PLAYWRIGHT_OWNED_STACK_NODE_BIN="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin/node"
    else
      PLAYWRIGHT_OWNED_STACK_NODE_BIN="node"
    fi
  fi

  local stack_file="${CARTULARY_WEB_E2E_STACK_JSON_FILE:-}"
  if [[ -z "${stack_file}" ]]; then
    echo "Playwright attachment requires CARTULARY_WEB_E2E_STACK_JSON_FILE" >&2
    return 2
  fi
  eval "$(
    "${PLAYWRIGHT_OWNED_STACK_NODE_BIN}" \
      "${root_dir}/tools/harness/browser/browser-session-evidence.mjs" \
      attach "${stack_file}"
  )"

  local actual_profile_id="${CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID:-}"
  if [[ "${expected_profile_id}" != "${actual_profile_id}" ]]; then
    echo "browser runtime profile attach mismatch: expected ${expected_profile_id}, stack is ${actual_profile_id}" >&2
    return 2
  fi

  PLAYWRIGHT_OWNED_STACK_COMMON_ENV=(
    env
    CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED=1
    CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER="${CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER}"
    CARTULARY_PLAYWRIGHT_STATE_DIR="${CARTULARY_PLAYWRIGHT_STATE_DIR}"
    PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-2}"
    COREPACK_HOME="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/corepack"
    PATH="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin:${PATH}"
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID="${expected_profile_id}"
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID="${actual_profile_id}"
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT="${CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT:-}"
    CARTULARY_WEB_E2E_STACK_JSON_FILE="${CARTULARY_WEB_E2E_STACK_JSON_FILE}"
    CARTULARY_WEB_E2E_STACK_SHA256="${CARTULARY_WEB_E2E_STACK_SHA256}"
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS="${CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS}"
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_REF="${CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_REF}"
    CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_SHA256="${CARTULARY_WEB_E2E_STARTUP_DIAGNOSTICS_SHA256}"
  )
  PLAYWRIGHT_OWNED_STACK_COMMON_ENV+=(
    CARTULARY_WEB_E2E_API_ORIGIN="${CARTULARY_WEB_E2E_API_ORIGIN}"
    CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${CARTULARY_WEB_E2E_PUBLIC_ORIGIN}"
  )
}
