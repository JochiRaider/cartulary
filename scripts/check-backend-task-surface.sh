#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"

fail() {
  echo "$*" >&2
  exit 1
}

extract_target_block() {
  local target="$1"
  awk -v target="$target" '
    $0 ~ "^" target ":" { in_block=1; next }
    in_block && /^[^[:space:]].*:/ { exit }
    in_block { print }
  ' "$makefile"
}

check_heavy_line="$(sed -n 's/^check-heavy:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_heavy_line" ]]; then
  fail "Makefile must define check-heavy prerequisites"
fi

read -r -a heavy_prereqs <<<"$check_heavy_line"
service_backed_targets=(
  backend-store
  backend-integration
  backend-integration-support
  backend-process
  backend-process-support
)
for target in "${service_backed_targets[@]}"; do
  for prereq in "${heavy_prereqs[@]}"; do
    if [[ "$prereq" == "$target" ]]; then
      fail "check-heavy must not include service-backed target $target"
    fi
  done
done

if ! rg -q '^backend-store:' "$makefile"; then
  fail "Makefile must define backend-store"
fi

check_service_block="$(extract_target_block check-service-backed)"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

for target in "${service_backed_targets[@]}"; do
  if ! printf '%s\n' "$check_service_block" | rg -q "(^|[[:space:]])$target($|[[:space:]])"; then
    fail "check-service-backed must invoke $target"
  fi
done

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])backend-store($|[[:space:]])'; then
  fail "test-fast must invoke backend-store"
fi
