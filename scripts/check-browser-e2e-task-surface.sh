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

read -r -a prereqs <<<"$check_heavy_line"
browser_targets=()
for prereq in "${prereqs[@]}"; do
  if [[ "$prereq" == browser-e2e* ]]; then
    browser_targets+=("$prereq")
  fi
done

if [[ "${#browser_targets[@]}" -ne 1 ]]; then
  fail "check-heavy must include exactly one browser-e2e* prerequisite, found: ${browser_targets[*]:-none}"
fi

if [[ "${browser_targets[0]}" != "browser-e2e-webserver-backed" ]]; then
  fail "check-heavy must use browser-e2e-webserver-backed as its only browser prerequisite, found: ${browser_targets[0]}"
fi

if ! rg -q '^browser-e2e-webserver-backed:' "$makefile"; then
  fail "Makefile must define browser-e2e-webserver-backed"
fi
