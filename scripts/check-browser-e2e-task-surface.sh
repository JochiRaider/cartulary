#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
functional_script="$repo_root/scripts/run-browser-e2e-functional.sh"
stateful_script="$repo_root/scripts/run-browser-e2e-stateful.sh"
webserver_backed_script="$repo_root/scripts/run-browser-e2e-webserver-backed.sh"
node_bin="${NODE_BIN:-node}"

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

browser_functional_block="$(awk '
  /^browser-e2e-functional:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$browser_functional_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-functional block"
fi
if ! printf '%s\n' "$browser_functional_block" | grep -Fq './scripts/run-browser-e2e-functional.sh'; then
  fail "browser-e2e-functional must delegate to scripts/run-browser-e2e-functional.sh"
fi

browser_stateful_block="$(awk '
  /^browser-e2e-stateful:/ { in_block=1; next }
  in_block && /^[^[:space:]].*:/ { exit }
  in_block { print }
' "$makefile")"
if [[ -z "$browser_stateful_block" ]]; then
  fail "Makefile must define a non-empty browser-e2e-stateful block"
fi
if ! printf '%s\n' "$browser_stateful_block" | grep -Fq './scripts/run-browser-e2e-stateful.sh'; then
  fail "browser-e2e-stateful must delegate to scripts/run-browser-e2e-stateful.sh"
fi

if ! [[ -f "$functional_script" ]]; then
  fail "missing scripts/run-browser-e2e-functional.sh"
fi
if ! [[ -f "$stateful_script" ]]; then
  fail "missing scripts/run-browser-e2e-stateful.sh"
fi

if ! grep -Fq 'phase1 authoritative browser_functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must execute Phase 1 browser_functional rows through the manifest"
fi
if grep -Fq 'e2e/phase1.spec.ts' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must not raw-select e2e/phase1.spec.ts"
fi
if ! grep -Fq 'phase2 authoritative browser_functional' "$functional_script"; then
  fail "scripts/run-browser-e2e-functional.sh must execute Phase 2 browser_functional rows through the manifest"
fi

if ! grep -Fq 'phase1 authoritative browser_stateful' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must execute Phase 1 browser_stateful rows through the manifest"
fi
if grep -Fq 'e2e/phase1.clock.spec.ts' "$stateful_script"; then
  fail "scripts/run-browser-e2e-stateful.sh must not raw-select e2e/phase1.clock.spec.ts"
fi

if ! grep -Fq '"$ROOT_DIR/scripts/run-browser-e2e-functional.sh"' "$webserver_backed_script"; then
  fail "scripts/run-browser-e2e-webserver-backed.sh must delegate to scripts/run-browser-e2e-functional.sh"
fi

if ! "$node_bin" - "$repo_root" <<'EOF'
const fs = require("fs");
const path = require("path");

const root = process.argv[2];
for (const [phase, expectedFor] of [
  [
    "phase1",
    (entry) => (entry.id === "E-1-04" ? "browser_stateful" : "browser_functional"),
  ],
  ["phase2", () => "browser_functional"],
]) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"),
  );
  for (const entry of manifest.e2e ?? []) {
    if (entry.coverage !== "authoritative" || entry.runner !== "playwright") {
      continue;
    }
    const expected = expectedFor(entry);
    if (entry.execution_dependency !== expected) {
      console.error(
        `${phase} authoritative e2e row ${entry.id} must declare execution_dependency=${expected}`,
      );
      process.exit(1);
    }
  }
}
EOF
then
  fail "Phase 1 and Phase 2 authoritative browser manifest rows must carry the canonical execution_dependency for their layer"
fi
