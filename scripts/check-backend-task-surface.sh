#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
makefile="$repo_root/Makefile"
node_bin="${NODE_BIN:-node}"

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

if rg -q 'TestPhase0_.*_U_0_|TestPhase0_.*_I_0_|TestPhase0_.*_E_0_' "$makefile"; then
  fail "Makefile must not use regex-based Phase 0 Go selection"
fi

if ! "$node_bin" - "$repo_root" <<'EOF'
const fs = require("fs");
const path = require("path");

const root = process.argv[2];
for (const [phase, sections] of [
  [
    "phase0",
    [
      ["unit", "backend_unit"],
      ["integration", "backend_integration"],
      ["e2e", "backend_process"],
    ],
  ],
  [
    "phase1",
    [
      ["unit", "backend_unit"],
      ["integration", "backend_integration"],
    ],
  ],
]) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"),
  );
  for (const [section, dependency] of sections) {
    for (const entry of manifest[section] ?? []) {
      if (entry.coverage !== "authoritative" || entry.runner !== "go_test") {
        continue;
      }
      if (entry.execution_dependency !== dependency) {
        console.error(
          `${phase} authoritative ${section} row ${entry.id} must declare execution_dependency=${dependency}`,
        );
        process.exit(1);
      }
    }
  }
}
EOF
then
  fail "Authoritative backend manifests must carry the canonical execution_dependency for their layer"
fi

backend_unit_block="$(extract_target_block backend-unit)"
for expected in \
  'RUN_GO_MANIFEST_PHASE) "backend-unit phase0 authoritative platform" phase0 unit authoritative backend_unit --' \
  'RUN_GO_MANIFEST_PHASE) "backend-unit phase0 authoritative app" phase0 unit authoritative backend_unit --' \
  'RUN_GO_MANIFEST_PHASE) "backend-unit phase1 authoritative platform" phase1 unit authoritative backend_unit --' \
  'RUN_GO_MANIFEST_PHASE) "backend-unit phase1 authoritative auth" phase1 unit authoritative backend_unit --'
do
  if ! printf '%s\n' "$backend_unit_block" | grep -Fq "$expected"; then
    fail "backend-unit must invoke Phase 0 manifest selection: missing $expected"
  fi
done

if printf '%s\n' "$backend_unit_block" | rg -q 'TestPhase1_.*_U_1_'; then
  fail "backend-unit must not use regex-based Phase 1 Go selection"
fi

if ! printf '%s\n' "$backend_unit_block" | grep -Fq "TestSupportPhase1_"; then
  fail "backend-unit must run Phase 1 support coverage through TestSupportPhase1_"
fi

backend_integration_block="$(extract_target_block backend-integration)"
if ! printf '%s\n' "$backend_integration_block" | grep -Fq 'RUN_GO_MANIFEST_PHASE) "backend-integration phase0 authoritative" phase0 integration authoritative backend_integration --'; then
  fail "backend-integration must invoke the Phase 0 manifest selector"
fi

if ! printf '%s\n' "$backend_integration_block" | grep -Fq 'RUN_GO_MANIFEST_PHASE) "backend-integration phase1 authoritative" phase1 integration authoritative backend_integration --'; then
  fail "backend-integration must invoke the Phase 1 manifest selector"
fi

if printf '%s\n' "$backend_integration_block" | rg -q 'TestPhase1_.*_I_1_'; then
  fail "backend-integration must not use regex-based Phase 1 Go selection"
fi

backend_process_block="$(extract_target_block backend-process)"
if ! printf '%s\n' "$backend_process_block" | grep -Fq 'RUN_GO_MANIFEST_PHASE) "backend-process phase0 authoritative" phase0 e2e authoritative backend_process --'; then
  fail "backend-process must invoke the Phase 0 manifest selector"
fi

phase0_process_block="$(extract_target_block phase0-process-e2e)"
if ! printf '%s\n' "$phase0_process_block" | grep -Fq 'RUN_GO_MANIFEST_PHASE) "phase0-process-e2e" phase0 e2e authoritative backend_process --'; then
  fail "phase0-process-e2e must invoke the Phase 0 manifest selector"
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

backend_integration_support_block="$(extract_target_block backend-integration-support)"
if ! printf '%s\n' "$backend_integration_support_block" | grep -Fq "TestSupportPhase1_"; then
  fail "backend-integration-support must run Phase 1 support coverage through TestSupportPhase1_"
fi

test_fast_block="$(extract_target_block test-fast)"
if [[ -z "$test_fast_block" ]]; then
  fail "Makefile must define a non-empty test-fast block"
fi
if ! printf '%s\n' "$test_fast_block" | rg -q '(^|[[:space:]])backend-store($|[[:space:]])'; then
  fail "test-fast must invoke backend-store"
fi
