#!/usr/bin/env bash
set -euo pipefail

fail=0

print_missing() {
  printf 'missing %s: %s\n' "$1" "$2"
  fail=1
}

go_path="${GO:-}"
if [[ -n "$go_path" && ! -x "$go_path" ]] && command -v "$go_path" >/dev/null 2>&1; then
  go_path="$(command -v "$go_path")"
fi
if [[ -n "$go_path" && -x "$go_path" ]]; then
  go_version_line="$("$go_path" version)"
  go_version="${go_version_line#go version }"
  go_version="${go_version%% *}"
  if [[ "$go_version" == go1.26* ]]; then
    printf 'ok go: %s %s\n' "$go_path" "$go_version"
  else
    printf 'missing go: expected Go 1.26, found %s at %s\n' "$go_version" "$go_path"
    fail=1
  fi
else
  print_missing go "install Go 1.26 or set GO=/path/to/go"
fi

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
    fail=1
  fi
else
  print_missing node "run make bootstrap-node-runtime"
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
    fail=1
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
    fail=1
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
