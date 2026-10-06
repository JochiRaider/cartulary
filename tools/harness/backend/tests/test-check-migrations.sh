#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "${path}"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "${label}: expected output to contain [${needle}]"
  fi
}

sha256_file() {
  local path="$1"

  sha256sum "$path" | awk '{print $1}'
}

make_history_repo() {
  local dir="$1"
  local second_name="${2:-00002_owner_repair.sql}"
  local first_hash
  local second_hash

  mkdir -p "${dir}/db/migrations" "${dir}/tools"
  cat >"${dir}/db/migrations/00001_bootstrap.sql" <<'SQL'
-- +goose Up
SELECT 1;
-- +goose Down
SELECT 1;
SQL
  cat >"${dir}/db/migrations/${second_name}" <<'SQL'
-- +goose Up
SELECT 2;
-- +goose Down
SELECT 2;
SQL
  first_hash="$(sha256_file "${dir}/db/migrations/00001_bootstrap.sql")"
  second_hash="$(sha256_file "${dir}/db/migrations/${second_name}")"
  cat >"${dir}/tools/migration_history_manifest.json" <<JSON
{
  "schema_id": "cartulary.migration_history_manifest.v1",
  "migration_root": "db/migrations",
  "immutable_through_version": 1,
  "entries": [
    {
      "version": 1,
      "filename": "00001_bootstrap.sql",
      "sha256": "${first_hash}",
      "historical_phase_shaped": false
    },
    {
      "version": 2,
      "filename": "${second_name}",
      "sha256": "${second_hash}",
      "historical_phase_shaped": false
    }
  ]
}
JSON
}

run_history_check() {
  local repo="$1"
  local output
  local status

  set +e
  output="$(node "${ROOT_DIR}/tools/harness/generated-artifacts/database-contract-drift/migration-history-cli.mjs" --root "$repo" 2>&1)"
  status=$?
  set -e

  printf '%s' "$output" >"${repo}/history-output.log"
  printf '%s' "$status" >"${repo}/history-status"
}

set_history_entry() {
  local repo="$1"
  local version="$2"
  local field="$3"
  local value="$4"

  node - "$repo/tools/migration_history_manifest.json" "$version" "$field" "$value" <<'NODE'
const [file, versionRaw, field, valueRaw] = process.argv.slice(2);
const fs = require("node:fs");
const manifest = JSON.parse(fs.readFileSync(file, "utf8"));
const version = Number.parseInt(versionRaw, 10);
const entry = manifest.entries.find((candidate) => candidate.version === version);
if (!entry) {
  throw new Error(`missing migration history entry ${version}`);
}
entry[field] = field === "historical_phase_shaped" ? valueRaw === "true" : valueRaw;
fs.writeFileSync(file, `${JSON.stringify(manifest, null, 2)}\n`);
NODE
}

history_valid_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migration-history-valid.XXXXXX")"
cleanup_paths+=("$history_valid_dir")
make_history_repo "$history_valid_dir"
run_history_check "$history_valid_dir"
assert_equals "$(cat "${history_valid_dir}/history-status")" "0" "valid history status"
assert_contains "$(cat "${history_valid_dir}/history-output.log")" "migration history check passed: 2 migrations" "valid history output"

history_gap_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migration-history-gap.XXXXXX")"
cleanup_paths+=("$history_gap_dir")
make_history_repo "$history_gap_dir"
mv "${history_gap_dir}/db/migrations/00002_owner_repair.sql" "${history_gap_dir}/db/migrations/00003_owner_repair.sql"
run_history_check "$history_gap_dir"
assert_equals "$(cat "${history_gap_dir}/history-status")" "1" "history gap status"
assert_contains "$(cat "${history_gap_dir}/history-output.log")" "creates migration version gap" "history gap diagnostic"

history_marker_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migration-history-marker.XXXXXX")"
cleanup_paths+=("$history_marker_dir")
make_history_repo "$history_marker_dir"
sed -i '/-- +goose Down/d' "${history_marker_dir}/db/migrations/00002_owner_repair.sql"
second_hash="$(sha256_file "${history_marker_dir}/db/migrations/00002_owner_repair.sql")"
set_history_entry "$history_marker_dir" 2 "sha256" "$second_hash"
run_history_check "$history_marker_dir"
assert_equals "$(cat "${history_marker_dir}/history-status")" "1" "history marker status"
assert_contains "$(cat "${history_marker_dir}/history-output.log")" "must contain a '-- +goose Down' marker" "history marker diagnostic"

history_hash_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migration-history-hash.XXXXXX")"
cleanup_paths+=("$history_hash_dir")
make_history_repo "$history_hash_dir"
set_history_entry "$history_hash_dir" 2 "sha256" "0000000000000000000000000000000000000000000000000000000000000000"
run_history_check "$history_hash_dir"
assert_equals "$(cat "${history_hash_dir}/history-status")" "1" "history hash status"
assert_contains "$(cat "${history_hash_dir}/history-output.log")" "hash drift" "history hash diagnostic"

history_phase_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migration-history-phase.XXXXXX")"
cleanup_paths+=("$history_phase_dir")
make_history_repo "$history_phase_dir" "00002_phase12_owner_repair.sql"
set_history_entry "$history_phase_dir" 2 "historical_phase_shaped" "true"
run_history_check "$history_phase_dir"
assert_equals "$(cat "${history_phase_dir}/history-status")" "1" "history future phase status"
assert_contains "$(cat "${history_phase_dir}/history-output.log")" "uses a phase-shaped name after immutable historical migrations" "history future phase diagnostic"
