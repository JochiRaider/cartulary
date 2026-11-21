#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"
SCRIPT="$ROOT_DIR/tools/harness/generated-artifacts/check-generate-drift.sh"
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

missing_sqlc_dir="$(cartulary_harness_mktemp_dir generate-drift-missing-sqlc.XXXXXX)"
cleanup_paths+=("$missing_sqlc_dir")
missing_sqlc="$missing_sqlc_dir/sqlc"

set +e
output="$(SQLC_BIN="$missing_sqlc" "$SCRIPT" 2>&1)"
status=$?
set -e

if [[ "$status" -eq 0 ]]; then
  fail "missing SQLC_BIN: expected generate-drift failure"
fi

assert_contains "$output" "generate-drift requires an executable SQLC_BIN at $missing_sqlc" "missing SQLC_BIN diagnostic"
assert_contains "$output" "run make generate before source-check CHECK=generated" "missing SQLC_BIN setup guidance"
assert_not_contains "$output" "bootstrap sqlc tool" "missing SQLC_BIN bootstrap avoidance"
assert_not_contains "$output" "generate sqlc" "missing SQLC_BIN generation avoidance"

node_bin="${NODE_BIN:-node}"
if command -v "$node_bin" >/dev/null 2>&1; then
  node_bin="$(command -v "$node_bin")"
fi

prepare_otel_contract_fixture() {
  local fixture_root="$1"

  mkdir -p \
    "$fixture_root/apps/web/src/app" \
    "$fixture_root/tools/otel"
  mkdir -p "$fixture_root/contracts/otel"
  local contract
  for contract in otel_source_snapshot.v1.json generated_constants_manifest.json import_boundary.json; do
    cp "$ROOT_DIR/contracts/otel/$contract" "$fixture_root/contracts/otel/$contract"
  done
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

otel_contract_tmp="$(cartulary_harness_mktemp_dir otel-contract-fixtures.XXXXXX)"
cleanup_paths+=("$otel_contract_tmp")

"$node_bin" --input-type=module - "$ROOT_DIR" "$otel_contract_tmp" <<'NODE'
import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import { pathToFileURL } from "node:url";
const [repo, fixture] = process.argv.slice(2);
const { copyGenerationInputs } = await import(pathToFileURL(path.join(repo, "tools/harness/generated-artifacts/copy-generation-inputs.mjs")));
const source = path.join(fixture, "copy-source"), output = path.join(fixture, "copy-output");
fs.mkdirSync(path.join(source, "contracts/human-material"), { recursive: true });
fs.mkdirSync(path.join(source, "contracts/README.bundle"));
fs.mkdirSync(path.join(source, "empty"));
const payload = Buffer.from(Array.from({ length: 2 * 65536 + 7 }, (_, index) => index % 251));
const input = path.join(source, "contracts/allowed.bin");
fs.writeFileSync(input, payload);
fs.writeFileSync(path.join(source, "contracts/tool.sh"), "#!/bin/sh\nexit 0\n");
fs.chmodSync(path.join(source, "contracts/tool.sh"), 0o751);
for (const name of ["README.md", "readme", "owner.MaRkDoWn", "human-material/hidden.ts", "README.bundle/hidden.ts"])
  fs.writeFileSync(path.join(source, "contracts", name), "excluded fixture input");
const prohibited = file => typeof file === "string" && file.startsWith(`${source}/`) &&
  /\/(?:human-material(?:\/|$)|readme(?:\.|\/|$)|owner\.markdown$)/iu.test(file);
const originals = Object.fromEntries(["lstatSync", "openSync", "readFileSync", "readdirSync", "copyFileSync"].map(name => [name, fs[name]]));
const opened = [];
try {
  for (const name of ["lstatSync", "openSync", "readFileSync", "readdirSync"]) {
    fs[name] = (file, ...args) => {
      assert.equal(prohibited(file), false, `probed excluded fixture input via ${name}`);
      if (name === "openSync" && typeof file === "string" && file.startsWith(`${source}/`)) opened.push(file);
      return originals[name](file, ...args);
    };
  }
  syncBuiltinESMExports();
  copyGenerationInputs(source, output, ["contracts", "contracts/allowed.bin", "empty"], ["contracts/human-material"]);
  assert.deepEqual(fs.readdirSync(path.join(output, "contracts")).sort(), ["allowed.bin", "tool.sh"]);
  assert.deepEqual(fs.readFileSync(path.join(output, "contracts/allowed.bin")), payload);
  assert.equal(fs.readFileSync(path.join(output, "contracts/tool.sh"), "utf8"), "#!/bin/sh\nexit 0\n");
  assert.equal(fs.statSync(path.join(output, "contracts/tool.sh")).mode & 0o777, 0o751);
  assert.deepEqual(fs.readdirSync(path.join(output, "empty")), []);
  assert.deepEqual(opened.sort(), [input, path.join(source, "contracts/tool.sh")].sort(), "overlapping selections copy each input once");
  const beforeRejected = opened.length;
  for (const request of ["contracts/README.md", "contracts/README.bundle/hidden.ts", "contracts/human-material/hidden.ts"]) {
    assert.throws(() => copyGenerationInputs(source, output, [request], ["contracts/human-material"]), /outside executable scope/);
  }
  for (const request of ["..", "/outside", "contracts/../allowed.bin", "contracts\\allowed.bin", "", "."]) {
    assert.throws(() => copyGenerationInputs(source, output, [request], []), /normalized repository-relative/);
  }
  assert.equal(opened.length, beforeRejected);
  assert.throws(() => copyGenerationInputs(source, output, [], []), /selection is empty/);
  assert.throws(() => copyGenerationInputs(source, path.join(source, "scratch"), ["contracts"], []), /must be disjoint/);
  assert.throws(() => copyGenerationInputs(source, output, ["absent"], []), /required generate-drift input missing: absent/);
  fs.symlinkSync("README.md", path.join(source, "contracts/link"));
  assert.throws(() => copyGenerationInputs(source, output, ["contracts/link"], []), /must be a regular file/);
  fs.symlinkSync("contracts", path.join(source, "alias"));
  assert.throws(() => copyGenerationInputs(source, output, ["alias/allowed.bin"], []), /ancestor.*symbolic links/);
  let copiedDescriptor;
  fs.copyFileSync = (from, to, ...args) => {
    copiedDescriptor = Number(from.slice("/proc/self/fd/".length));
    const result = originals.copyFileSync(from, to, ...args);
    fs.appendFileSync(input, "changed");
    return result;
  };
  assert.throws(() => copyGenerationInputs(source, path.join(fixture, "changed-copy"), ["contracts/allowed.bin"], []), /changed while copying/);
  assert.ok(Number.isInteger(copiedDescriptor));
  assert.throws(() => fs.fstatSync(copiedDescriptor), { code: "EBADF" });
} finally {
  Object.assign(fs, originals);
  syncBuiltinESMExports();
}
NODE

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

# Public generation is transactional (TH-HARNESS-REQ-004): a renderer failure
# after the SQL outputs are cleared must leave every declared generated path
# byte-identical and leave no publication transaction behind.
generate_fixture="$(cartulary_harness_mktemp_dir generate-transaction-fixture.XXXXXX)"
cleanup_paths+=("$generate_fixture")
generate_fixture_tool_dir="$(cartulary_harness_mktemp_dir generate-transaction-tools.XXXXXX)"
cleanup_paths+=("$generate_fixture_tool_dir")

prepare_generate_fixture() {
  local fixture="$1"
  local foundation_behavior="$2"
  rm -rf "$fixture"
  mkdir -p "$fixture/tools" "$fixture/internal/gen/sql" "$fixture/internal/app"
  # The driver and its input copier run from the fixture's own harness tree.
  cp -a "$ROOT_DIR/tools/harness" "$fixture/tools/harness"
  cp -p "$ROOT_DIR/tools/executable_input_policy.json" "$fixture/tools/executable_input_policy.json"
  mkdir -p "$fixture/tools/schemas"
  local layout_file
  for layout_file in workspace_layout.json workspace_layout.generated.sh workspace_layout.generated.mk workspace_layout.generated.mjs; do
    cp -p "$ROOT_DIR/tools/$layout_file" "$fixture/tools/$layout_file"
  done
  cp -p "$ROOT_DIR/tools/schemas/cartulary.workspace_layout.v1.schema.json" "$fixture/tools/schemas/"
  case "$foundation_behavior" in
    fail) printf 'process.stderr.write("fixture foundation failure\\n");\nprocess.exit(1);\n' ;;
    *) printf '// fixture foundation renderer\n' ;;
  esac >"$fixture/tools/harness/generated-artifacts/generate-foundation-schema-validators.mjs"
  printf '{"owners":[]}\n' >"$fixture/tools/test_catalog_owner.json"
  printf '{"targets":[],"harness_checks":[]}\n' >"$fixture/tools/task_surface_owner.json"
  : >"$fixture/Makefile"
  cat >"$fixture/tools/generate_drift_scratch_inputs.json" <<'JSON'
{
  "schema_id": "cartulary.generate_drift_scratch_inputs.v1",
  "generated_paths": ["internal/gen/sql"],
  "copy_paths": ["internal/gen/sql", "tools/executable_input_policy.json", "tools/harness", "tools/schemas", "tools/task_surface_owner.json", "tools/test_catalog_owner.json", "tools/workspace_layout.generated.mjs", "tools/workspace_layout.generated.mk", "tools/workspace_layout.generated.sh", "tools/workspace_layout.json"],
  "placeholder_dirs": ["internal/app"]
}
JSON
  printf '// Code generated by sqlc. DO NOT EDIT.\npackage sql\n' >"$fixture/internal/gen/sql/models.go"
  printf '// Code generated by sqlc. DO NOT EDIT.\npackage sql\n\nconst fixture = 1\n' >"$fixture/internal/gen/sql/query.sql.go"
  ln -s "$ROOT_DIR/node_modules" "$fixture/node_modules"
  git -C "$fixture" init --quiet
}

generated_tree_digest() {
  (cd "$1" && find internal/gen/sql -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum)
}

printf '#!/usr/bin/env bash\nexit 0\n' >"$generate_fixture_tool_dir/sqlc"
chmod 0755 "$generate_fixture_tool_dir/sqlc"
mkdir -p "$generate_fixture_tool_dir/go-cache" "$generate_fixture_tool_dir/go-mod" "$generate_fixture_tool_dir/go-tmp"

prepare_generate_fixture "$generate_fixture" fail
before_digest="$(generated_tree_digest "$generate_fixture")"
set +e
output="$(cd "$generate_fixture" && env \
  NODE_BIN="$node_bin" \
  SQLC_BIN="$generate_fixture_tool_dir/sqlc" \
  GO="${GO:-go}" \
  GO_CACHE_DIR="$generate_fixture_tool_dir/go-cache" \
  GO_MOD_CACHE_DIR="$generate_fixture_tool_dir/go-mod" \
  GO_TMP_DIR="$generate_fixture_tool_dir/go-tmp" \
  RUN_STEP_SCRIPT="$ROOT_DIR/tools/harness/execution/run-step.sh" \
  CARTULARY_HARNESS_REPO_ROOT="$generate_fixture" \
  bash ./tools/harness/generated-artifacts/generate-artifacts.sh 2>&1)"
status=$?
set -e
if [[ "$status" -eq 0 ]]; then
  fail "failing generation fixture: expected generate failure"
fi
if [[ "$output" != *"fixture foundation failure"* ]]; then
  printf '%s\n' "$output" | tail -20 >&2
  fail "failing generation fixture diagnostic: expected the renderer failure"
fi
after_digest="$(generated_tree_digest "$generate_fixture")"
if [[ "$before_digest" != "$after_digest" ]]; then
  fail "failing generation fixture: generated SQL outputs changed after a failed generate"
fi
if compgen -G "$generate_fixture/tmp/generated-publish.*" >/dev/null; then
  fail "failing generation fixture: a publication transaction was left behind"
fi

# The private renderer refuses to run in a tree its scratch driver did not mark.
if [[ -f "$generate_fixture/tools/harness/generated-artifacts/render-generated-artifacts.sh" ]]; then
  set +e
  output="$(cd "$generate_fixture" && env NODE_BIN="$node_bin" bash ./tools/harness/generated-artifacts/render-generated-artifacts.sh 2>&1)"
  status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "unmarked render: expected refusal"
  fi
  assert_contains "$output" "renders only inside a marked generation scratch tree" "unmarked render diagnostic"
  if [[ "$before_digest" != "$(generated_tree_digest "$generate_fixture")" ]]; then
    fail "unmarked render: generated SQL outputs changed"
  fi
fi

# The render closure check sees creates, rewrites (even byte-identical) and
# removals outside generated_paths, and ignores only scratch-state roots.
closure_fixture="$(cartulary_harness_mktemp_dir generate-closure-fixture.XXXXXX)"
cleanup_paths+=("$closure_fixture")
"$node_bin" --input-type=module - "$ROOT_DIR" "$closure_fixture" <<'NODE'
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";
const [repo, root] = process.argv.slice(2);
const { renderedInventory, undeclaredRenderedWrites } = await import(pathToFileURL(path.join(repo, "tools/harness/generated-artifacts/generated-transaction.mjs")));
const write = (relative, text) => { fs.mkdirSync(path.dirname(path.join(root, relative)), { recursive: true }); fs.writeFileSync(path.join(root, relative), text); };
write("internal/gen/sql/models.go", "package sql\n");
write("tools/input.json", "{}\n");
write("tools/removed.json", "{}\n");
const before = renderedInventory(root);
const later = new Date(Date.now() + 5000);
write("internal/gen/sql/models.go", "package sql\n// regenerated\n");
write("internal/gen/sql/new.go", "package sql\n");
write("tools/input.json", "{}\n");
fs.utimesSync(path.join(root, "tools/input.json"), later, later);
write("tools/undeclared.json", "{}\n");
fs.rmSync(path.join(root, "tools/removed.json"));
for (const scratchState of ["tmp/render.log", ".cartulary/state.json", ".git/HEAD", "node_modules/x.js"]) write(scratchState, "x\n");
const after = renderedInventory(root);
assert.deepEqual(
  undeclaredRenderedWrites(before, after, ["internal/gen/sql"]),
  ["tools/input.json", "tools/removed.json", "tools/undeclared.json"],
);
assert.deepEqual(undeclaredRenderedWrites(before, before, ["internal/gen/sql"]), []);
assert.throws(() => undeclaredRenderedWrites(before, after, ["../escape"]), /unsafe generated path/);
NODE
