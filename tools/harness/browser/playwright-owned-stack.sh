#!/usr/bin/env bash

resolve_playwright_owned_stack_env() {
  local root_dir="$1"

  if [[ -z "${root_dir}" ]]; then
    echo "resolve_playwright_owned_stack_env requires <repo_root>" >&2
    return 2
  fi

  local expected_profile_id="${CARTULARY_BROWSER_RUNTIME_PROFILE_ID:-default}"
  local actual_profile_id="${CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID:-}"
  if [[ -z "${actual_profile_id}" ]]; then
    echo "browser stack metadata is missing CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID" >&2
    return 2
  fi
  if [[ "${expected_profile_id}" != "${actual_profile_id}" ]]; then
    echo "browser runtime profile attach mismatch: expected ${expected_profile_id}, stack is ${actual_profile_id}" >&2
    return 2
  fi

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

  PLAYWRIGHT_OWNED_STACK_COMMON_ENV=(
    env
    CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER=1
    PLAYWRIGHT_WORKERS="${PLAYWRIGHT_WORKERS:-2}"
    COREPACK_HOME="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/corepack"
    PATH="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin:${PATH}"
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID="${expected_profile_id}"
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_ID="${actual_profile_id}"
    CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT="${CARTULARY_WEB_E2E_RUNTIME_PROFILE_FINGERPRINT:-}"
  )
  if [[ -n "${CARTULARY_WEB_E2E_API_ORIGIN:-}" ]]; then
    PLAYWRIGHT_OWNED_STACK_COMMON_ENV+=(CARTULARY_WEB_E2E_API_ORIGIN="${CARTULARY_WEB_E2E_API_ORIGIN}")
  fi
  if [[ -n "${CARTULARY_WEB_E2E_PUBLIC_ORIGIN:-}" ]]; then
    PLAYWRIGHT_OWNED_STACK_COMMON_ENV+=(CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${CARTULARY_WEB_E2E_PUBLIC_ORIGIN}")
  fi
}
