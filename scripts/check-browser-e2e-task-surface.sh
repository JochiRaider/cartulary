#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"

fail() {
  echo "$*" >&2
  exit 1
}

check_heavy_line="$(sed -n 's/^check-heavy:[[:space:]]*//p' "$makefile" | head -n 1)"
if [[ -z "$check_heavy_line" ]]; then
  fail "Makefile must define check-heavy prerequisites"
fi

read -r -a heavy_prereqs <<<"$check_heavy_line"
browser_targets=()
for prereq in "${heavy_prereqs[@]}"; do
  if [[ "$prereq" == browser-e2e* ]]; then
    browser_targets+=("$prereq")
  fi
done

if [[ "${#browser_targets[@]}" -ne 0 ]]; then
  fail "check-heavy must not include browser-e2e* prerequisites, found: ${browser_targets[*]}"
fi

check_service_block="$(awk '
  /^check-service-backed:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$check_service_block" ]]; then
  fail "Makefile must define a non-empty check-service-backed block"
fi

service_browser_targets=()
while IFS= read -r line; do
  while IFS= read -r target; do
    [[ -n "$target" ]] && service_browser_targets+=("$target")
  done < <(printf '%s\n' "$line" | grep -o 'browser-e2e[^[:space:]]*' || true)
done <<<"$check_service_block"

if [[ "${#service_browser_targets[@]}" -ne 1 ]]; then
  fail "check-service-backed must invoke exactly one browser-e2e* target, found: ${service_browser_targets[*]:-none}"
fi

if [[ "${service_browser_targets[0]}" != "browser-e2e-webserver-backed" ]]; then
  fail "check-service-backed must use browser-e2e-webserver-backed as its only browser target, found: ${service_browser_targets[0]}"
fi

if ! rg -q '^browser-e2e-webserver-backed:' "$makefile"; then
  fail "Makefile must define browser-e2e-webserver-backed"
fi
