#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
fail=0

print_missing() {
  printf 'missing %s: %s\n' "$1" "$2"
  fail=2
}

go_diagnostic=""
if go_diagnostic="$(
  GO="${GO:-}" \
  GO_TOOLCHAIN="${GO_TOOLCHAIN:-}" \
  GO_CACHE_DIR="${GO_CACHE_DIR:-}" \
  GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-}" \
  GO_TMP_DIR="${GO_TMP_DIR:-}" \
    "$ROOT_DIR/tools/harness/readiness/go-toolchain-readiness.sh" diagnose 2>&1
)"; then
  printf '%s\n' "$go_diagnostic"
else
  printf '%s\n' "$go_diagnostic"
  fail=2
fi

machine_state_diagnostic() {
  local name="$1"
  local configured="$2"
  local existing="$configured"
  if [[ -z "$configured" ]]; then
    printf 'missing %s: resolved machine-state path is empty\n' "$name"
    fail=2
    return
  fi
  while [[ ! -e "$existing" && "$existing" != "/" ]]; do
    existing="$(dirname -- "$existing")"
  done
  local filesystem="unknown"
  local available_bytes="unknown"
  filesystem="$(stat -f -c '%T' -- "$existing" 2>/dev/null || true)"
  available_bytes="$(df -Pk -- "$existing" 2>/dev/null | awk 'NR == 2 { printf "%.0f", $4 * 1024 }' || true)"
  printf 'ok machine-state: %s=%s filesystem=%s available_bytes=%s\n' \
    "$name" "$configured" "${filesystem:-unknown}" "${available_bytes:-unknown}"
}

machine_state_diagnostic GO_CACHE_DIR "${GO_CACHE_DIR:-}"
machine_state_diagnostic GO_MOD_CACHE_DIR "${GO_MOD_CACHE_DIR:-}"
machine_state_diagnostic GO_TMP_DIR "${GO_TMP_DIR:-}"

for legacy_cache in /tmp/cartulary-go-build /tmp/cartulary-go-mod; do
  if [[ -e "$legacy_cache" ]]; then
    printf 'advisory legacy machine-state: %s is not used or removed automatically; stop active Go/Make jobs before manual cleanup\n' "$legacy_cache"
  fi
done

node_path=""
if [[ -x "${NODE_BIN:-}" ]]; then
  node_path="$NODE_BIN"
fi
if [[ -n "$node_path" ]]; then
  node_version="$("$node_path" --version)"
  if [[ "$node_version" == "v${NODE_VERSION:-}" ]]; then
    printf 'ok node: %s %s\n' "$node_path" "$node_version"
  else
    printf 'missing node: expected v%s, found %s at %s\n' "${NODE_VERSION:-}" "$node_version" "$node_path"
    fail=2
  fi
else
  print_missing node "run make bootstrap-node-runtime"
fi

if [[ -n "$node_path" ]]; then
  if ! "$node_path" "$ROOT_DIR/tools/harness/readiness/diagnose-inotify.mjs" --advisory; then
    fail=2
  fi
fi

pnpm_path=""
if [[ -n "${PNPM:-}" && -x "$PNPM" ]]; then
  pnpm_path="$PNPM"
fi
if [[ -n "$pnpm_path" ]]; then
  pnpm_version="$(PATH="${NODE_RUNTIME_DIR:-}/bin:$PATH" COREPACK_HOME="${NODE_RUNTIME_DIR:-}/corepack" "$pnpm_path" --version 2>/dev/null || true)"
  if [[ "$pnpm_version" == "${PNPM_VERSION:-}" ]]; then
    printf 'ok pnpm: %s %s\n' "$pnpm_path" "$pnpm_version"
  else
    printf 'missing pnpm: expected %s, found %s at %s\n' "${PNPM_VERSION:-}" "${pnpm_version:-unusable}" "$pnpm_path"
    fail=2
  fi
else
  print_missing pnpm "run make frontend-toolchain"
fi

shellcheck_path=""
if [[ -x "${SHELLCHECK_BIN:-}" ]]; then
  shellcheck_path="$SHELLCHECK_BIN"
fi
if [[ -n "$shellcheck_path" ]]; then
  shellcheck_version="$("$shellcheck_path" --version 2>/dev/null | awk -F': ' '$1 == "version" { print $2; exit }')"
  if [[ "$shellcheck_version" == "${SHELLCHECK_VERSION:-}" ]]; then
    printf 'ok shellcheck: %s %s\n' "$shellcheck_path" "$shellcheck_version"
  else
    printf 'missing shellcheck: expected %s, found %s at %s\n' "${SHELLCHECK_VERSION:-}" "${shellcheck_version:-unusable}" "$shellcheck_path"
    fail=2
  fi
else
  print_missing shellcheck "run make shell-lint-toolchain"
fi

if command -v docker >/dev/null 2>&1; then
  docker_path="$(command -v docker)"
  compose_version="$(docker compose version --short 2>/dev/null || docker compose version 2>/dev/null || true)"
  if [[ -n "$compose_version" ]]; then
    printf 'ok docker compose: %s %s\n' "$docker_path" "$compose_version"
  else
    print_missing "docker compose" "install Docker with the compose plugin"
  fi
else
  print_missing "docker compose" "install Docker with the compose plugin"
fi

for tool in rg curl tar; do
  if command -v "$tool" >/dev/null 2>&1; then
    tool_path="$(command -v "$tool")"
    tool_version="$("$tool" --version 2>/dev/null | head -n 1 || true)"
    printf 'ok %s: %s %s\n' "$tool" "$tool_path" "$tool_version"
  else
    print_missing "$tool" "install $tool and ensure it is on PATH"
  fi
done

if command -v ss >/dev/null 2>&1; then
  ss_path="$(command -v ss)"
  ss_version="$(ss --version 2>&1 | head -n 1 || true)"
  printf 'ok ss: %s %s\n' "$ss_path" "$ss_version"
else
  printf 'optional missing ss: install iproute2 for local port diagnostics\n'
fi

exit "$fail"
