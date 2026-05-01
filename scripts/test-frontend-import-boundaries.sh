#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
CHECKER="$ROOT_DIR/scripts/check-frontend-import-boundaries.mjs"
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
    fail "$label: expected output to contain [$needle], got [$haystack]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle], got [$haystack]"
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

write_config() {
  local case_root="$1"

  mkdir -p "$case_root/tools"
  cat >"$case_root/tools/frontend_import_boundaries.json" <<'JSON'
{
  "schema_id": "cartulary.frontend_import_boundaries.v1",
  "scan_roots": [
    "apps/web/src",
    "apps/web/e2e",
    "packages/grid-adapter/src",
    "packages/protocol-ts/src"
  ],
  "scan_excludes": [
    "packages/protocol-ts/src/generated/**"
  ],
  "rules": [
    {
      "id": "frontend-grid-vendor-boundary",
      "level": "error",
      "message": "Import react-data-grid only through @cartulary/grid-adapter.",
      "allowed_importers": [
        "packages/grid-adapter/src/**"
      ],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "react-data-grid",
          "include_subpaths": true
        }
      ]
    },
    {
      "id": "frontend-generated-protocol-boundary",
      "level": "warning",
      "message": "Import generated protocol artifacts only through the @cartulary/protocol-ts facade.",
      "allowed_importers": [
        "packages/protocol-ts/src/index.ts"
      ],
      "restricted_imports": [
        {
          "kind": "package",
          "name": "@cartulary/protocol-ts/generated",
          "include_subpaths": true
        },
        {
          "kind": "path_prefix",
          "path": "packages/protocol-ts/src/generated"
        }
      ]
    }
  ]
}
JSON
}

prepare_case_root() {
  local name="$1"
  local case_root="$tmp_dir/$name"

  mkdir -p \
    "$case_root/apps/web/src" \
    "$case_root/apps/web/e2e" \
    "$case_root/packages/grid-adapter/src" \
    "$case_root/packages/protocol-ts/src/generated"
  write_config "$case_root"
  printf '%s\n' "$case_root"
}

run_checker() {
  local case_root="$1"
  shift

  "$NODE_BIN" "$CHECKER" --root "$case_root" --config tools/frontend_import_boundaries.json "$@"
}

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/frontend-import-boundaries.XXXXXX")"
cleanup_paths+=("$tmp_dir")

allowed_grid_root="$(prepare_case_root allowed-grid)"
cat >"$allowed_grid_root/packages/grid-adapter/src/index.tsx" <<'TS'
import { DataGrid } from "react-data-grid";
import "react-data-grid/lib/styles.css";

export const grid = DataGrid;
TS
allowed_grid_output="$(assert_passes "allowed grid adapter import" run_checker "$allowed_grid_root")"
assert_contains "$allowed_grid_output" "frontend import boundaries verified" "allowed grid adapter output"

blocked_grid_root="$(prepare_case_root blocked-grid)"
cat >"$blocked_grid_root/apps/web/src/GridLeak.tsx" <<'TS'
import { DataGrid } from "react-data-grid";

export const grid = DataGrid;
TS
blocked_grid_output="$(assert_fails "blocked app grid import" run_checker "$blocked_grid_root")"
assert_contains "$blocked_grid_output" "frontend-grid-vendor-boundary" "blocked grid rule"
assert_contains "$blocked_grid_output" "apps/web/src/GridLeak.tsx" "blocked grid file"

generated_package_root="$(prepare_case_root generated-package)"
cat >"$generated_package_root/apps/web/src/generatedProtocol.ts" <<'TS'
import { contractArtifactIndex } from "@cartulary/protocol-ts/generated";

export const contracts = contractArtifactIndex;
TS
generated_package_output="$(assert_passes "generated package warning" run_checker "$generated_package_root")"
assert_contains "$generated_package_output" "warning: frontend-generated-protocol-boundary" "generated package warning"
generated_package_error_output="$(assert_fails "generated warning as error" run_checker "$generated_package_root" --warnings-as-errors)"
assert_contains "$generated_package_error_output" "error: frontend-generated-protocol-boundary" "generated warning promoted to error"

generated_relative_root="$(prepare_case_root generated-relative)"
cat >"$generated_relative_root/apps/web/e2e/support.ts" <<'TS'
import { contractArtifactIndex } from "../../../packages/protocol-ts/src/generated/contracts";

export const contracts = contractArtifactIndex;
TS
generated_relative_output="$(assert_passes "generated relative warning" run_checker "$generated_relative_root")"
assert_contains "$generated_relative_output" "warning: frontend-generated-protocol-boundary" "generated relative warning"
assert_contains "$generated_relative_output" "apps/web/e2e/support.ts" "generated relative file"

facade_root="$(prepare_case_root facade)"
cat >"$facade_root/apps/web/src/contracts.ts" <<'TS'
import { parseContractArtifact } from "@cartulary/protocol-ts";

export const parse = parseContractArtifact;
TS
facade_output="$(assert_passes "protocol facade import" run_checker "$facade_root")"
assert_contains "$facade_output" "frontend import boundaries verified" "facade import output"
assert_not_contains "$facade_output" "frontend-generated-protocol-boundary" "facade import must not warn"

protocol_owner_root="$(prepare_case_root protocol-owner)"
cat >"$protocol_owner_root/packages/protocol-ts/src/index.ts" <<'TS'
import { contractArtifactIndex } from "./generated/index.js";

export const contracts = contractArtifactIndex;
TS
protocol_owner_output="$(assert_passes "protocol facade generated import" run_checker "$protocol_owner_root")"
assert_contains "$protocol_owner_output" "frontend import boundaries verified" "protocol owner generated import output"
