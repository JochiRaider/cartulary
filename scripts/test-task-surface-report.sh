#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
REPORTER="$ROOT_DIR/scripts/print-task-surface-report.mjs"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_passes() {
  local label="$1"
  shift

  local output
  if ! output="$("$@" 2>&1)"; then
    fail "$label: expected success, got output: $output"
  fi
  printf '%s' "$output"
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

valid_output="$(assert_passes "current task-surface report" "$NODE_BIN" "$REPORTER" --check)"
assert_contains "$valid_output" "Cartulary task-surface report" "current report header"
assert_contains "$valid_output" "browser-e2e-measurement" "current report measurement target"
assert_contains "$valid_output" "phase-map execution dependencies" "current report phase dependency section"

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/task-surface-report.XXXXXX")"
cleanup_paths+=("$tmp_dir")
makefile_copy="$tmp_dir/Makefile"
manifest_copy="$tmp_dir/task_surface_manifest.json"
cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"

run_report_copy() {
  CARTULARY_TASK_SURFACE_MAKEFILE="$makefile_copy" \
  CARTULARY_TASK_SURFACE_MANIFEST="$manifest_copy" \
    "$NODE_BIN" "$REPORTER" --check
}

printf '\n.PHONY: unclassified-target\nunclassified-target:\n\t@true\n' >>"$makefile_copy"
unclassified_output="$(assert_fails "unclassified target drift" run_report_copy)"
assert_contains "$unclassified_output" "unclassified-target is missing task-surface classification" "unclassified target output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
printf '\n.PHONY: undocumented-public\nundocumented-public:\n\t@true\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.targets.push({
  name: "undocumented-public",
  classification: "public",
  included_in: ["helper_only"]
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
undocumented_output="$(assert_fails "public target without help" run_report_copy)"
assert_contains "$undocumented_output" "public target undocumented-public is missing a help entry" "undocumented public output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
const target = manifest.targets.find((entry) => entry.name === "task-surface-report");
target.backing_scripts = ["scripts/missing-task-surface-helper.mjs"];
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
missing_script_output="$(assert_fails "missing backing script" run_report_copy)"
assert_contains "$missing_script_output" "backing script missing: scripts/missing-task-surface-helper.mjs" "missing backing script output"

cp "$ROOT_DIR/Makefile" "$makefile_copy"
cp "$ROOT_DIR/tools/task_surface_manifest.json" "$manifest_copy"
printf '\n.PHONY: undeclared-script-ref\nundeclared-script-ref:\n\t@node ./scripts/check-toolchain-pins.mjs\n' >>"$makefile_copy"
"$NODE_BIN" - "$manifest_copy" <<'EOF'
const { readFileSync, writeFileSync } = require("node:fs");
const manifestPath = process.argv[2];
const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
manifest.targets.push({
  name: "undeclared-script-ref",
  classification: "helper_only",
  included_in: ["helper_only"],
  backing_scripts: []
});
writeFileSync(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
undeclared_script_output="$(assert_fails "undeclared script reference" run_report_copy)"
assert_contains "$undeclared_script_output" "references scripts/check-toolchain-pins.mjs" "undeclared script reference output"
