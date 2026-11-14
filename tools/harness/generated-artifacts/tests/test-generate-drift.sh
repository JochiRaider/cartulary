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
