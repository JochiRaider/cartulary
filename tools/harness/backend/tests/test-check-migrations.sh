#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/generated-artifacts/database-contract-drift/check-migrations.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "${label}: expected output not to contain [${needle}]"
  fi
}

assert_file_absent_or_empty() {
  local path="$1"
  local label="$2"

  if [[ -s "$path" ]]; then
    fail "${label}: expected ${path} to be absent or empty"
  fi
}

count_lines() {
  local text="$1"
  local pattern="$2"

  printf '%s\n' "$text" | grep -Fxc "$pattern" | tr -d ' '
}

sha256_file() {
  local path="$1"

  sha256sum "$path" | awk '{print $1}'
}

write_fakes() {
  local dir="$1"

  mkdir -p "${dir}/bin"

  cat >"${dir}/bin/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"${DOCKER_LOG:?}"
EOF
  chmod +x "${dir}/bin/docker"

  cat >"${dir}/wait-postgres.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"${WAIT_LOG:?}"
EOF
  chmod +x "${dir}/wait-postgres.sh"

  cat >"${dir}/migrate" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf '%s|%s\n' "${CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN:?}" "$*" >>"${MIGRATE_LOG:?}"
if [[ "$*" == *"up-by-one"* ]]; then
  exit 91
fi
EOF
  chmod +x "${dir}/migrate"
}

make_migrations() {
  local dir="$1"
  shift
  local name

  mkdir -p "$dir"
  for name in "$@"; do
    touch "${dir}/${name}"
  done
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

run_check() {
  local work_dir="$1"
  local migrations_dir="$2"
  local output
  local status

  : >"${work_dir}/docker.log"
  : >"${work_dir}/wait.log"
  : >"${work_dir}/migrate.log"
  : >"${work_dir}/docker-compose.yml"
  : >"${work_dir}/config.toml"

  set +e
  output="$(
    PATH="${work_dir}/bin:${PATH}" \
    DOCKER_LOG="${work_dir}/docker.log" \
    WAIT_LOG="${work_dir}/wait.log" \
    MIGRATE_LOG="${work_dir}/migrate.log" \
    CARTULARY_MIGRATE_BIN="${work_dir}/migrate" \
    CARTULARY_MIGRATIONS_DIR="${migrations_dir}" \
    CARTULARY_DEV_SERVICES_SCRIPT="${work_dir}/wait-postgres.sh" \
    CARTULARY_COMPOSE_FILE="${work_dir}/docker-compose.yml" \
    CONFIG_FILE="${work_dir}/config.toml" \
      "$SCRIPT" 2>&1
  )"
  status=$?
  set -e

  printf '%s' "$output" >"${work_dir}/output.log"
  printf '%s' "$status" >"${work_dir}/status"
}

commands_from_log() {
  local path="$1"

  sed 's/^.*|//' "$path"
}

normal_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migrations-normal.XXXXXX")"
cleanup_paths+=("$normal_dir")
write_fakes "$normal_dir"
make_migrations "${normal_dir}/migrations" \
  00001_phase0_bootstrap.sql \
  00002_phase4_record_envelope_backfill.sql \
  00003_phase4_workbook_surfaces.sql \
  00004_phase4_assessments_core02.sql \
  00005_phase4_evidence_blob_routes.sql
run_check "$normal_dir" "${normal_dir}/migrations"
assert_equals "$(cat "${normal_dir}/status")" "0" "normal migration check status"
normal_output="$(cat "${normal_dir}/output.log")"
normal_commands="$(commands_from_log "${normal_dir}/migrate.log")"
assert_contains "$normal_output" "migration verification: empty database apply to head" "normal empty database output"
assert_contains "$normal_output" "migration verification: upgrade path from pre-record-envelope boundary" "normal record boundary output"
assert_contains "$normal_output" "migration verification: upgrade path from pre-assessments-Core02 boundary" "normal assessments boundary output"
assert_contains "$normal_output" "migration verification: upgrade path from penultimate boundary" "normal penultimate output"
assert_equals "$(count_lines "$normal_commands" "up")" "4" "normal up command count"
assert_equals "$(count_lines "$normal_commands" "up-to 1")" "1" "normal record boundary version"
assert_equals "$(count_lines "$normal_commands" "up-to 3")" "1" "normal assessments boundary version"
assert_equals "$(count_lines "$normal_commands" "up-to 4")" "1" "normal penultimate boundary version"
assert_not_contains "$normal_commands" "up-by-one" "normal migration commands"

single_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migrations-single.XXXXXX")"
cleanup_paths+=("$single_dir")
write_fakes "$single_dir"
make_migrations "${single_dir}/migrations" 00001_phase0_bootstrap.sql
run_check "$single_dir" "${single_dir}/migrations"
assert_equals "$(cat "${single_dir}/status")" "0" "single migration check status"
single_output="$(cat "${single_dir}/output.log")"
single_commands="$(commands_from_log "${single_dir}/migrate.log")"
assert_contains "$single_output" "skipping pre-record-envelope boundary" "single record boundary skip"
assert_contains "$single_output" "skipping pre-assessments-Core02 boundary" "single assessments boundary skip"
assert_contains "$single_output" "only one migration exists; running best-available boundary" "single penultimate fallback"
assert_equals "$(count_lines "$single_commands" "up")" "2" "single up command count"
assert_not_contains "$single_commands" "up-to" "single migration commands"
assert_not_contains "$single_commands" "up-by-one" "single migration commands"

malformed_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migrations-malformed.XXXXXX")"
cleanup_paths+=("$malformed_dir")
write_fakes "$malformed_dir"
make_migrations "${malformed_dir}/migrations" 00001_phase0_bootstrap.sql not_a_goose_migration.sql
run_check "$malformed_dir" "${malformed_dir}/migrations"
assert_equals "$(cat "${malformed_dir}/status")" "1" "malformed migration check status"
assert_contains "$(cat "${malformed_dir}/output.log")" "invalid migration filename \"not_a_goose_migration.sql\"" "malformed migration diagnostic"
assert_file_absent_or_empty "${malformed_dir}/migrate.log" "malformed migrate log"

missing_anchor_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-migrations-missing-anchor.XXXXXX")"
cleanup_paths+=("$missing_anchor_dir")
write_fakes "$missing_anchor_dir"
make_migrations "${missing_anchor_dir}/migrations" \
  00001_phase0_bootstrap.sql \
  00002_phase4_record_envelope_backfill.sql \
  00003_phase4_workbook_surfaces.sql
run_check "$missing_anchor_dir" "${missing_anchor_dir}/migrations"
assert_equals "$(cat "${missing_anchor_dir}/status")" "1" "missing anchor migration check status"
assert_contains "$(cat "${missing_anchor_dir}/output.log")" "missing migration anchor for pre-assessments-Core02 boundary" "missing anchor diagnostic"
assert_file_absent_or_empty "${missing_anchor_dir}/migrate.log" "missing anchor migrate log"

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
