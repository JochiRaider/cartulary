#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/check-generate-drift.sh"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

missing_sqlc="$ROOT_DIR/tmp/generate-drift-missing-sqlc"
rm -f "$missing_sqlc"

set +e
output="$(SQLC_BIN="$missing_sqlc" "$SCRIPT" 2>&1)"
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  fail "missing SQLC_BIN: expected generate-drift failure"
fi

assert_contains "$output" "generate-drift requires an executable SQLC_BIN at $missing_sqlc" "missing SQLC_BIN diagnostic"
assert_contains "$output" "run make codegen-toolchain before generate-drift" "missing SQLC_BIN setup guidance"
assert_not_contains "$output" "bootstrap sqlc tool" "missing SQLC_BIN bootstrap avoidance"
assert_not_contains "$output" "generate sqlc" "missing SQLC_BIN generation avoidance"

node_bin="${NODE_BIN:-node}"

prepare_otel_contract_fixture() {
  local fixture_root="$1"

  mkdir -p \
    "$fixture_root/apps/web/src/app" \
    "$fixture_root/tools/otel"
  cp -a "$ROOT_DIR/contracts" "$fixture_root/contracts"
  cp "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" "$fixture_root/tools/otel/generate-otel-contracts.mjs"
  cp "$ROOT_DIR/apps/web/src/app/otelBoundary.test.ts" "$fixture_root/apps/web/src/app/otelBoundary.test.ts"
}

mutate_json_file() {
  local file="$1"
  local expression="$2"

  "$node_bin" --input-type=module - "$file" "$expression" <<'NODE'
import { readFileSync, writeFileSync } from "node:fs";

const [file, expression] = process.argv.slice(2);
const data = JSON.parse(readFileSync(file, "utf8"));
Function("data", expression)(data);
writeFileSync(file, `${JSON.stringify(data, null, 2)}\n`);
NODE
}

otel_contract_tmp="$(mktemp -d "$ROOT_DIR/tmp/otel-contract-fixtures.XXXXXX")"
cleanup_paths+=("$otel_contract_tmp")

valid_otel_fixture="$otel_contract_tmp/valid"
prepare_otel_contract_fixture "$valid_otel_fixture"
"$node_bin" "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" --root "$valid_otel_fixture" --check >/dev/null

stale_sha_fixture="$otel_contract_tmp/stale-sha"
prepare_otel_contract_fixture "$stale_sha_fixture"
mutate_json_file \
  "$stale_sha_fixture/contracts/otel/otel_source_snapshot.v1.json" \
  'data.semconv_generated_constants.generator_source_sha = "0000000000000000000000000000000000000000";'
set +e
output="$("$node_bin" "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" --root "$stale_sha_fixture" --check 2>&1)"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "OTel stale generator SHA fixture: expected failure"
fi
assert_contains "$output" "generator_source_sha must match tools/otel/generate-otel-contracts.mjs" "OTel stale generator SHA diagnostic"

missing_generator_fixture="$otel_contract_tmp/missing-generator"
prepare_otel_contract_fixture "$missing_generator_fixture"
rm "$missing_generator_fixture/tools/otel/generate-otel-contracts.mjs"
set +e
output="$("$node_bin" "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" --root "$missing_generator_fixture" --check 2>&1)"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "OTel missing generator fixture: expected failure"
fi
assert_contains "$output" "generator_source_ref file is missing: tools/otel/generate-otel-contracts.mjs" "OTel missing generator diagnostic"

stale_probe_fixture="$otel_contract_tmp/stale-probe"
prepare_otel_contract_fixture "$stale_probe_fixture"
mutate_json_file \
  "$stale_probe_fixture/contracts/otel/import_boundary.json" \
  'data.browser_runtime_probe.evidence = "apps/web/src/otelBoundary.test.ts::OpenTelemetry browser boundary";'
set +e
output="$("$node_bin" "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" --root "$stale_probe_fixture" --check 2>&1)"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "OTel stale browser probe fixture: expected failure"
fi
assert_contains "$output" "browser_runtime_probe.evidence file is missing: apps/web/src/otelBoundary.test.ts" "OTel stale browser probe diagnostic"

missing_probe_name_fixture="$otel_contract_tmp/missing-probe-name"
prepare_otel_contract_fixture "$missing_probe_name_fixture"
mutate_json_file \
  "$missing_probe_name_fixture/contracts/otel/import_boundary.json" \
  'data.browser_runtime_probe.evidence = "apps/web/src/app/otelBoundary.test.ts::Missing OpenTelemetry boundary test";'
set +e
output="$("$node_bin" "$ROOT_DIR/tools/otel/generate-otel-contracts.mjs" --root "$missing_probe_name_fixture" --check 2>&1)"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "OTel missing browser probe test-name fixture: expected failure"
fi
assert_contains "$output" "browser_runtime_probe.evidence test name is absent" "OTel missing browser probe test-name diagnostic"

scratch_tool_dir="$(mktemp -d "$ROOT_DIR/tmp/generate-drift-smoke.XXXXXX")"
cleanup_paths+=("$scratch_tool_dir")
fake_sqlc="$scratch_tool_dir/sqlc"
fake_go="$scratch_tool_dir/go"
node_runtime_fixture="$scratch_tool_dir/node-runtime"
node_archive_fixture="$scratch_tool_dir/node-archives"

cat >"$fake_sqlc" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "generate" ]]; then
  echo "unexpected sqlc invocation: $*" >&2
  exit 2
fi
EOF

cat >"$fake_go" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "run" && "${2:-}" == "./tools/contractgen" ]]; then
  exit 0
fi

echo "unexpected go invocation: $*" >&2
exit 2
EOF

chmod +x "$fake_sqlc" "$fake_go"
mkdir -p "$node_runtime_fixture/bin" "$node_archive_fixture"
ln -s "$node_bin" "$node_runtime_fixture/bin/node"
if [[ -x "${PNPM:-$ROOT_DIR/tmp/node-runtime/bin/pnpm}" ]]; then
  ln -s "${PNPM:-$ROOT_DIR/tmp/node-runtime/bin/pnpm}" "$node_runtime_fixture/bin/pnpm"
fi

scratch_manifest="$scratch_tool_dir/scratch-inputs.json"
cat >"$scratch_manifest" <<'EOF'
{
  "schema_id": "cartulary.generate_drift_scratch_inputs.v1",
  "generated_paths": ["internal/gen/contracts"],
  "copy_paths": ["scripts/harness-contract.mjs", "tools/definitely-missing-generate-drift-input"],
  "placeholder_dirs": []
}
EOF

set +e
output="$(
  SQLC_BIN="$fake_sqlc" \
  GO="$fake_go" \
  GO_CACHE_DIR="$scratch_tool_dir/go-cache" \
  GO_MOD_CACHE_DIR="$scratch_tool_dir/go-mod" \
  GENERATE_DRIFT_SCRATCH_INPUT_MANIFEST="${scratch_manifest#"$ROOT_DIR"/}" \
    "$SCRIPT" 2>&1
)"
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  fail "scratch input manifest: expected missing declared input failure"
fi

assert_contains \
  "$output" \
  "required generate-drift input missing: tools/definitely-missing-generate-drift-input" \
  "scratch input manifest missing input diagnostic"

set +e
output="$(
  SQLC_BIN="$fake_sqlc" \
  GO="$fake_go" \
  GO_CACHE_DIR="$scratch_tool_dir/go-cache" \
  GO_MOD_CACHE_DIR="$scratch_tool_dir/go-mod" \
  NODE_RUNTIME_DIR="$node_runtime_fixture" \
  CARTULARY_NODE_ARCHIVE_DIR="$node_archive_fixture" \
  CARTULARY_BOOTSTRAP_NODE_PLATFORM="definitely-unsupported" \
    "$SCRIPT" 2>&1
)"
status=$?
set -e

if [[ "$status" -ne 0 ]]; then
  fail "scratch Make include smoke: expected generate-drift success with fake generators, got output: $output"
fi

assert_not_contains \
  "$output" \
  "No rule to make target 'tools/task_surface.generated.mk'" \
  "scratch Make include smoke"
