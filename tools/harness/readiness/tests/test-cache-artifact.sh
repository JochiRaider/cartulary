#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="$ROOT_DIR/tools/harness/readiness/cache-artifact.sh"
NODE_BIN="${NODE_BIN:-node}"
TMP_DIR="$(mktemp -d "$ROOT_DIR/tmp/cache-artifact-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
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
    fail "$label: expected [$expected], got [$actual]"
  fi
}

json_field() {
  local file="$1"
  local expr="$2"
  "$NODE_BIN" - "$file" "$expr" <<'JS'
const fs = require("node:fs");
const [file, expr] = process.argv.slice(2);
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const result = Function("value", `return ${expr}`)(value);
if (result === null || result === undefined) {
  process.stdout.write("");
} else {
  process.stdout.write(String(result));
}
JS
}

cache_dir="$TMP_DIR/cache"
input_file="$TMP_DIR/input.txt"
output_file="$TMP_DIR/output.txt"
output_dir="$TMP_DIR/output-dir"
command_log="$TMP_DIR/command.log"
command_script="$TMP_DIR/write-output.sh"
results_dir="$TMP_DIR/results"
run_id="run"
target="cache-artifact-fixture"

cat >"$command_script" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'run\n' >>"$COMMAND_LOG"
cp "$INPUT_FILE" "$OUTPUT_FILE"
mkdir -p "$OUTPUT_DIR/empty"
cp "$INPUT_FILE" "$OUTPUT_DIR/copied.txt"
SH
chmod +x "$command_script"
printf 'value-one\n' >"$input_file"

run_cache() {
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
  CARTULARY_TEST_RUN_ID="$run_id" \
  CARTULARY_TEST_TARGET="$target" \
  INPUT_FILE="$input_file" \
  OUTPUT_FILE="$output_file" \
  OUTPUT_DIR="$output_dir" \
  COMMAND_LOG="$command_log" \
    "$SCRIPT" \
      \
      --scope "${fixture_scope:-build-artifact}" \
      --profile fixture-build \
      --cache-dir "$cache_dir" \
      --disable-env CARTULARY_BUILD_CACHE_DISABLE \
      --force-env CARTULARY_FORCE_REBUILD \
      --input "$input_file" \
      --input "$command_script" \
      --output "$output_file" \
      --output-dir "$output_dir" \
      --key "fixture=one" \
      -- "$command_script"
}

run_cache
assert_equals "$(cat "$command_log")" "run" "first miss executes command once"
record_path="$(find "$cache_dir/fixture-build" -type f -name '*.json' -print -quit)"
[[ -n "$record_path" ]] || fail "first miss did not publish a cache record"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
  validate-schema cartulary.harness_cache_record.v2 "$record_path" >/dev/null
assert_equals \
  "$(json_field "$record_path" 'value.policy')" \
  "content_addressed" \
  "record uses the canonical cache policy"
assert_equals \
  "$(json_field "$record_path" 'value.artifacts[0].relative_path')" \
  "${output_file#"$ROOT_DIR"/}" \
  "record identifies the cached output"
assert_equals "$(json_field "$record_path" 'value.artifacts[0].members.length')" "0" "file has no directory members"
assert_equals "$(json_field "$record_path" 'value.artifacts[1].members.map(member => member.relative_path).join(",")')" \
  "copied.txt,empty" "directory records both files and empty directories"

run_cache
assert_equals "$(cat "$command_log")" "run" "cache hit skips command"

printf 'value-two\n' >"$input_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "2" "input change executes command"
record_path="$(find "$cache_dir/fixture-build" -type f -name '*.json' -printf '%T@ %p\n' | sort -nr | head -1 | cut -d' ' -f2-)"

printf '{not valid json\n' >"$record_path"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "3" "corrupt record executes command"
"$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
  validate-schema cartulary.harness_cache_record.v2 "$record_path" >/dev/null

CARTULARY_BUILD_CACHE_DISABLE=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "4" "disabled cache executes command"

CARTULARY_FORCE_REBUILD=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "5" "force rebuild executes command"

rm -f "$output_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "6" "missing output executes command"

fixture_scope=readiness
set +e
CARTULARY_READINESS_CHECK_ONLY=1 run_cache >"$TMP_DIR/readiness.log" 2>&1
readiness_status=$?
set -e
assert_equals "$readiness_status" "2" "inspect-only readiness rejects a cold cache"
assert_equals "$(grep -c '^run$' "$command_log")" "6" "inspect-only readiness never invokes installation"
run_cache
CARTULARY_READINESS_CHECK_ONLY=1 run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "7" "inspect-only readiness accepts the exact warm cache"
unset fixture_scope

# A matching subset of fields is not a cache record. Mutations preserve the old
# grep identity fields while corrupting independent parts of the output contract.
run_cache
for mutation in extra_field missing_artifacts surplus_artifact wrong_kind wrong_path wrong_destination wrong_mode wrong_producer wrong_input wrong_digest missing_member wrong_member_mode wrong_member_digest semantic_failure timestamp trailing duplicate; do
  before="$(grep -c '^run$' "$command_log")"
  "$NODE_BIN" --input-type=module - "$record_path" "$mutation" <<'JS'
import { readFileSync, writeFileSync } from "node:fs";
const [file, mutation] = process.argv.slice(2);
const original = readFileSync(file, "utf8");
const prior = JSON.parse(original);
const value = structuredClone(prior);
switch (mutation) {
  case "extra_field": value.extra = true; break;
  case "missing_artifacts": value.artifacts = []; break;
  case "surplus_artifact": value.artifacts.push(structuredClone(value.artifacts[0])); break;
  case "wrong_kind": value.artifacts[0].artifact_type = "directory"; break;
  case "wrong_path": value.artifacts[0].relative_path += ".other"; break;
  case "wrong_destination": value.artifacts[0].destination_class = "run_evidence"; break;
  case "wrong_mode": value.artifacts[0].mode = "0777"; break;
  case "wrong_producer": value.artifacts[0].producer_identity = "other"; break;
  case "wrong_input": value.artifacts[0].semantic_input_digest = `sha256:${"0".repeat(64)}`; break;
  case "wrong_digest": value.artifacts[0].digest = `sha256:${"0".repeat(64)}`; break;
  case "missing_member": value.artifacts[1].members.pop(); break;
  case "wrong_member_mode": value.artifacts[1].members[0].mode = "0777"; break;
  case "wrong_member_digest": value.artifacts[1].members[0].digest = `sha256:${"0".repeat(64)}`; break;
  case "semantic_failure": value.semantic_result.status = "failed"; break;
  case "timestamp": value.created_at = "2026-99-99T99:99:99Z"; break;
}
let bytes = original;
for (const key of Object.keys(prior)) {
  if (JSON.stringify(prior[key]) !== JSON.stringify(value[key])) {
    const member = `  "${key}": ${JSON.stringify(prior[key])}`;
    if (!bytes.includes(member)) throw new Error(`fixture did not locate ${key}`);
    bytes = bytes.replace(member, `  "${key}": ${JSON.stringify(value[key])}`);
  }
}
if (mutation === "extra_field") bytes = bytes.replace('\n}\n', ',\n  "extra": true\n}\n');
if (mutation === "trailing") bytes += "unparsed bytes\n";
if (mutation === "duplicate") bytes = bytes.replace('{\n', '{\n  "policy": "none",\n');
writeFileSync(file, bytes);
JS
  run_cache
  assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 1))" "$mutation invalidates reuse"
  "$NODE_BIN" "$ROOT_DIR/tools/harness/contract/harness-contract-cli.mjs" \
    validate-schema cartulary.harness_cache_record.v2 "$record_path" >/dev/null
done
before="$(grep -c '^run$' "$command_log")"
chmod 600 "$output_file"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 1))" "changed output mode invalidates reuse"
mkdir "$output_dir/new-empty"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 2))" "new empty directory invalidates reuse"
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 2))" "complete unchanged manifest reuses"
# Exact output destinations belong to the key, even for the same source/command.
cp "$record_path" "$TMP_DIR/prior-record"
prior_record="$record_path"
output_file="$TMP_DIR/changed-output.txt"
run_cache
cmp -- "$prior_record" "$TMP_DIR/prior-record" || fail "changed output contract reused an old key"
assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 3))" "changed output contract executes"
# JSON escaping supports every filesystem control byte except NUL (not a filename).
output_file="$TMP_DIR/"$'quote"-back\\-line\n-tab\t-control\001'
run_cache
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "$((before + 4))" "escaped output name admits exact warm reuse"
"$NODE_BIN" --input-type=module - "$cache_dir/fixture-build" "$output_file" "$ROOT_DIR" <<'JS'
import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
const [directory, output, root] = process.argv.slice(2);
const records = readdirSync(directory).map(file => JSON.parse(readFileSync(`${directory}/${file}`, "utf8")));
assert.equal(records.filter(record => record.artifacts[0].relative_path === output.slice(root.length + 1)).length, 1);
JS

# Unsupported outputs cannot be hidden by a file-only directory walk or alias.
for unsafe in root_link ancestor_link member_link member_fifo; do
  original_output="$output_file"
  case "$unsafe" in
    root_link) ln -s "$original_output" "$TMP_DIR/output-link"; output_file="$TMP_DIR/output-link" ;;
    ancestor_link) ln -s "$TMP_DIR" "$TMP_DIR/ancestor-link"; output_file="$TMP_DIR/ancestor-link/changed-output.txt" ;;
    member_link) ln -s "$input_file" "$output_dir/unsafe" ;;
    member_fifo) mkfifo "$output_dir/unsafe" ;;
  esac
  before="$(grep -c '^run$' "$command_log")"
  set +e
  run_cache >"$TMP_DIR/unsafe.log" 2>&1
  status=$?
  set -e
  assert_equals "$status" "2" "$unsafe rejects unsupported artifact identity"
  assert_equals "$(grep -c '^run$' "$command_log")" "$before" "$unsafe never invokes producer"
  grep -Fq 'cache artifact' "$TMP_DIR/unsafe.log" || fail "$unsafe lost artifact diagnostic"
  output_file="$original_output"
  rm -f -- "$TMP_DIR/output-link" "$TMP_DIR/ancestor-link" "$output_dir/unsafe"
done
run_cache
assert_equals "$(grep -c '^run$' "$command_log")" "$before" "unsafe admission preserves prior valid record"

# Independent digest oracle and hash-process bound for a multi-batch input set.
# NUL records must preserve spaces, backslashes, newlines and option-like names.
batch_root="$TMP_DIR/batch"
mkdir -p "$batch_root/bin" "$batch_root/dir/empty"
real_sha256="$(command -v sha256sum)"
export REAL_SHA256="$real_sha256" HASH_CALLS="$batch_root/hash-calls"
cat >"$batch_root/bin/sha256sum" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "--zero" ]]; then
  printf '%s\n' "$(($# - 2))" >>"$HASH_CALLS"
  if [[ "${FAIL_HASH_BATCH:-0}" == "1" ]]; then exit 1; fi
fi
exec "$REAL_SHA256" "$@"
SH
chmod +x "$batch_root/bin/sha256sum"
batch_inputs=()
for ((i=0;i<260;i++)); do
  file="$batch_root/input-$i"
  printf 'content %s\n' "$i" >"$file"
  batch_inputs+=(--input "$file")
done
for name in 'space name' 'back\slash' $'line\nbreak' '-option'; do
  file="$batch_root/$name"
  printf 'special\n' >"$file"
  batch_inputs+=(--input "$file")
done
batch_inputs+=(--input "$batch_root/missing")
printf 'directory member\n' >"$batch_root/dir/member"
printf 'initial output\n' >"$batch_root/output"
cat >"$batch_root/build" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'run\n' >>"$BATCH_LOG"
printf 'output\n' >"$BATCH_OUTPUT"
SH
chmod +x "$batch_root/build"
export BATCH_LOG="$batch_root/build.log" BATCH_OUTPUT="$batch_root/output"
run_batch() {
  PATH="$batch_root/bin:$PATH" "$SCRIPT" --scope build-artifact --profile batch \
    --cache-dir "$batch_root/cache" "${batch_inputs[@]}" --input-dir "$batch_root/dir" \
    --output "$batch_root/output" --key fixture=batch -- "$batch_root/build"
}
run_batch
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$SCRIPT" "$batch_root" <<'JS'
import assert from "node:assert/strict";
import { readFileSync, readdirSync } from "node:fs";
import { createHash } from "node:crypto";
import os from "node:os";
const [root, script, dir] = process.argv.slice(2);
const hash = bytes => createHash("sha256").update(bytes).digest("hex");
const relative = file => file.startsWith(`${root}/`) ? file.slice(root.length + 1) : file;
const files = [...Array.from({length:260},(_,i)=>`${dir}/input-${i}`), ...["space name","back\\slash","line\nbreak","-option"].map(name=>`${dir}/${name}`)];
let material = `schema_id\tcartulary.harness_cache_record.v2\nscope\tbuild-artifact\nprofile\tbatch\nplatform\t${os.type()}:${os.machine()}\nhelper\tsha256:${hash(readFileSync(script))}\nkey\tfixture=batch\n`;
for (const file of files) material += `input\t${relative(file)}\tsha256:${hash(readFileSync(file))}\n`;
material += `input\t${relative(`${dir}/missing`)}\tmissing\n`;
const member = `file\t${relative(`${dir}/dir/member`)}\tsha256:${hash(readFileSync(`${dir}/dir/member`))}\n`;
material += `input-dir\t${relative(`${dir}/dir`)}\tsha256:${hash(member)}\noutput-contract\tfile\t${relative(`${dir}/output`)}\ncommand\t${dir}/build\n`;
const records = readdirSync(`${dir}/cache/batch`);
assert.equal(records.length,1);
const record = JSON.parse(readFileSync(`${dir}/cache/batch/${records[0]}`));
assert.equal(record.input_digest, `sha256:${hash(material)}`);
assert.equal(record.unit_digest, record.input_digest);
assert.equal(records[0],`${hash(material)}.json`);
const calls = readFileSync(`${dir}/hash-calls`,"utf8").trim().split("\n").map(Number);
assert.deepEqual(calls,[128,128,8,1,1,1]);
JS
run_batch
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "1" "batched warm reuse"
printf 'changed output\n' >"$batch_root/output"
run_batch
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "2" "changed output rejects reuse"
if FAIL_HASH_BATCH=1 run_batch >"$batch_root/failure.log" 2>&1; then
  fail "failed digest batch was accepted"
fi
assert_equals "$(grep -c '^run$' "$BATCH_LOG")" "2" "hash failure never runs producer"
grep -Fq 'cache hashing failed' "$batch_root/failure.log" || fail "hash failure lost its cause"

# Every content-proof stage must reject tool failure, including nested substitutions
# and a discovery tool that emits a partial result before failing.
fault_root="$TMP_DIR/faults"
mkdir -p "$fault_root/bin"
REAL_FIND="$(command -v find)"
export REAL_FIND CACHE_HELPER="$SCRIPT"
cat >"$fault_root/bin/sha256sum" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
last="${!#}"
reject=0
case "$FAULT_STAGE" in
  helper) [[ "$last" != "$CACHE_HELPER" ]] || reject=1 ;;
  input_directory) [[ "$last" != "$FAULT_ROOT/input-dir/member" ]] || reject=1 ;;
  directory_digest)
    if [[ "${1:-}" != --zero ]] && grep -Fq "/input-dir/member" "$last"; then reject=1; fi ;;
  output_before) [[ "$last" != "$FAULT_ROOT/output" ]] || reject=1 ;;
  output_after)
    if [[ -f "$FAULT_ROOT/produced" && "$last" == "$FAULT_ROOT/output" ]]; then reject=1; fi ;;
  output_aggregate)
    if [[ "${1:-}" != --zero ]] && grep -Fq $'output\t' "$last"; then reject=1; fi ;;
  artifact_digest)
    if [[ "${1:-}" != --zero && "$last" == "$FAULT_ROOT/output" ]]; then reject=1; fi ;;
esac
if [[ "$reject" == 1 ]]; then printf 'hash\n' >"$FAULT_ROOT/injected"; printf 'injected hash failure\n' >&2; exit 1; fi
exec "$REAL_SHA256" "$@"
SH
cat >"$fault_root/bin/find" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
reject=0
case "$FAULT_STAGE" in
  input_discovery) [[ "$1" != "$FAULT_ROOT/input-dir" ]] || reject=1 ;;
  output_discovery)
    if [[ -f "$FAULT_ROOT/produced" && "$1" == "$FAULT_ROOT/output-dir" ]]; then reject=1; fi ;;
  member_discovery) [[ "${2:-}" != -mindepth ]] || reject=1 ;;
esac
if [[ "$reject" == 1 ]]; then printf 'discovery\n' >"$FAULT_ROOT/injected"; printf '%s\0' "$1/member"; printf 'injected discovery failure\n' >&2; exit 1; fi
exec "$REAL_FIND" "$@"
SH
chmod +x "$fault_root/bin/sha256sum" "$fault_root/bin/find"
for stage in helper input_directory directory_digest output_before output_after output_aggregate artifact_digest input_discovery output_discovery member_discovery; do
  case_root="$fault_root/$stage"
  mkdir -p "$case_root/tmp" "$case_root/input-dir" "$case_root/output-dir"
  printf 'input\n' >"$case_root/input-dir/member"
  printf 'old\n' >"$case_root/output"
  cat >"$case_root/build" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'new\n' >"$FAULT_ROOT/output"
printf 'member\n' >"$FAULT_ROOT/output-dir/member"
printf 'produced\n' >"$FAULT_ROOT/produced"
SH
  chmod +x "$case_root/build"
  set +e
  FAULT_STAGE="$stage" FAULT_ROOT="$case_root" TMPDIR="$case_root/tmp" PATH="$fault_root/bin:$PATH" \
    "$SCRIPT" --scope build-artifact --profile fault --cache-dir "$case_root/cache" \
    --input-dir "$case_root/input-dir" --output "$case_root/output" --output-dir "$case_root/output-dir" \
    -- "$case_root/build" >"$case_root/log" 2>&1
  status=$?
  set -e
  [[ -f "$case_root/injected" ]] || fail "$stage did not inject its failure"
  assert_equals "$status" "2" "$stage rejects unavailable content proof"
  grep -Eq 'cache (hashing|directory discovery) failed' "$case_root/log" || fail "$stage lost its cause"
  case "$stage" in
    output_after|artifact_digest|output_discovery|member_discovery) [[ -f "$case_root/produced" ]] || fail "$stage did not reach publication" ;;
    *) [[ ! -e "$case_root/produced" ]] || fail "$stage invoked producer after failed proof" ;;
  esac
  [[ -z "$(find "$case_root/tmp" -mindepth 1 -print -quit)" ]] || fail "$stage leaked temporary material"
  if [[ -d "$case_root/cache" ]]; then
    [[ -z "$(find "$case_root/cache" -type f -print -quit)" ]] || fail "$stage published a partial record"
  fi
done

bash "$ROOT_DIR/tools/harness/backend/tests/test-build-go-artifact.sh"

echo "cache artifact tests passed"
