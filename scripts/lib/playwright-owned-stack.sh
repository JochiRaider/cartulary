#!/usr/bin/env bash

resolve_playwright_owned_stack_env() {
  local root_dir="$1"

  if [[ -z "${root_dir}" ]]; then
    echo "resolve_playwright_owned_stack_env requires <repo_root>" >&2
    return 2
  fi

  PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-$root_dir/tmp/node-runtime}"
  PLAYWRIGHT_OWNED_STACK_PNPM_BIN="${PNPM:-}"
  PLAYWRIGHT_OWNED_STACK_NODE_BIN="${NODE_BIN:-}"

  if [[ -z "${PLAYWRIGHT_OWNED_STACK_PNPM_BIN}" ]]; then
    if command -v pnpm >/dev/null 2>&1; then
      PLAYWRIGHT_OWNED_STACK_PNPM_BIN="$(command -v pnpm)"
    elif [[ -x "$HOME/.local/share/pnpm/pnpm" ]]; then
      PLAYWRIGHT_OWNED_STACK_PNPM_BIN="$HOME/.local/share/pnpm/pnpm"
    else
      echo "pnpm was not provided and could not be discovered" >&2
      return 1
    fi
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
    PATH="${PLAYWRIGHT_OWNED_STACK_NODE_RUNTIME_DIR}/bin:${PATH}"
  )
}
