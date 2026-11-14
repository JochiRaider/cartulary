#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
MAKE_BIN="${MAKE_BIN:-make}"
NODE_BIN="${NODE_BIN:-node}"
case_name="${1:-}"
if [[ "$#" -ne 1 ]]; then
  echo "one real-command assertion case is required" >&2
  exit 2
fi
case "$case_name" in
  source-schema|backend-unit|backend-unit-machine|owner-json|lint|build|build-machine|test-fast|check|check-machine|invalid-usage) ;;
  *) echo "unknown real-command assertion case: $case_name" >&2; exit 2 ;;
esac
if [[ "${CARTULARY_STEP_ARTIFACT_DIR:-}" != /* ]]; then
  echo "real-command assertions require a Make-owned unit artifact directory" >&2
  exit 2
fi
umask 077
TMP_DIR="${CARTULARY_STEP_ARTIFACT_DIR}/real-command-output"
RUN_PREFIX=real
RESULTS_ROOT="${TMP_DIR}/runs"
mkdir -p "$TMP_DIR"
# These artifacts belong to the enclosing unit and survive interruption and
# post-run validation failures. The normal run-retention policy owns cleanup.
printf '[REAL-TARGET] case=%s artifacts=%s\n' "$case_name" "$TMP_DIR"


# The native runtime strips Make inputs from child environments. The immutable
# parent manifest is the authority for the selected diagnostic mode.
parent_diagnostics() {
  "$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import path from "node:path";
import { pathToFileURL } from "node:url";
const [repoRoot] = process.argv.slice(2);
const { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } =
  await import(pathToFileURL(path.join(repoRoot, "tools/harness/contract/index.mjs")));
const { validateRunId } =
  await import(pathToFileURL(path.join(repoRoot, "tools/harness/contract/harness-contract.mjs")));
const parent = process.env.CARTULARY_TEST_RESULTS_DIR ?? "";
const runID = process.env.CARTULARY_TEST_RUN_ID ?? "";
if (!path.isAbsolute(parent) || validateRunId(runID) !== runID) {
  throw new Error("real-command assertions require an exact parent run identity");
}
const manifest = parseStrictJSON(readRetainedArtifactBytes(parent, `${runID}/run-manifest.json`, 1024 * 1024).toString("utf8"));
validateSchemaSync("cartulary.harness_run_manifest.v3", manifest);
if (manifest.run_id !== runID) throw new Error("real-command parent manifest identity mismatch");
process.stdout.write(manifest.instrumentation.mode);
EOF
}

run_target() {
  local target="$1"
  local label="$2"
  local target_dir="${TMP_DIR}/${label}"
  local make_args=()
  if [[ -n "${RUN_TARGET_MAKE_ARGS:-}" ]]; then
    read -r -a make_args <<<"${RUN_TARGET_MAKE_ARGS}"
  fi
  local diagnostics
  diagnostics="$(parent_diagnostics)" || return 1
  make_args+=("HARNESS_DIAGNOSTICS=${diagnostics}")
  mkdir -p "$target_dir"
  (
    cd "$ROOT_DIR"
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u BROWSER_LANES -u JSON \
      -u MAKEFLAGS -u MFLAGS \
      CARTULARY_OUTPUT_MODE=summary \
      CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
      CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target" "${make_args[@]}"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log" || {
    local status=$?
    echo "${target}: failed with status ${status}" >&2
    sed -n '1,80p' "${target_dir}/stdout.log" >&2
    sed -n '1,80p' "${target_dir}/stderr.log" >&2
    # The command-under-test outcome is retained above. A Make wrapper's 2
    # means this expected-success assertion failed, not invalid test inputs.
    return 1
  }
  "$NODE_BIN" "${ROOT_DIR}/tools/harness/tests/real-command-output.mjs" summary \
    "$ROOT_DIR" "$target" "$target_dir" "${RESULTS_ROOT}/${RUN_PREFIX}-${label}" "${RUN_PREFIX}-${label}"
}

run_machine_target() {
  local target="$1"
  local label="$2"
  local target_dir="${TMP_DIR}/${label}"
  local make_args=()
  if [[ -n "${RUN_TARGET_MAKE_ARGS:-}" ]]; then read -r -a make_args <<<"${RUN_TARGET_MAKE_ARGS}"; fi
  local diagnostics
  diagnostics="$(parent_diagnostics)" || return 1
  make_args+=("HARNESS_DIAGNOSTICS=${diagnostics}")
  mkdir -p "$target_dir"
  (
    cd "$ROOT_DIR"
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u BROWSER_LANES -u JSON \
      -u MAKEFLAGS -u MFLAGS \
      CARTULARY_OUTPUT_MODE="${RUN_TARGET_OUTPUT_MODE:-machine}" \
      CARTULARY_TEST_RESULTS_DIR="${RESULTS_ROOT}" \
      CARTULARY_TEST_RUN_ID="${RUN_PREFIX}-${label}" \
      "$MAKE_BIN" --no-print-directory "$target" "${make_args[@]}"
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log" || {
    local status=$?
    echo "${target}: machine target failed with status ${status}" >&2
    sed -n '1,80p' "${target_dir}/stdout.log" >&2
    sed -n '1,80p' "${target_dir}/stderr.log" >&2
    return 1
  }
  "$NODE_BIN" "${ROOT_DIR}/tools/harness/tests/real-command-output.mjs" machine \
    "$ROOT_DIR" "$target" "$target_dir" "${RESULTS_ROOT}/${RUN_PREFIX}-${label}" "${RUN_PREFIX}-${label}"
}

run_invalid_usage_check() {
  local target_dir="${TMP_DIR}/invalid-target-plan"
  mkdir -p "$target_dir"
  set +e
  (
    cd "$ROOT_DIR"
    env -u CARTULARY_HARNESS_IDENTITY_PREPARED \
      -u CARTULARY_TEST_RESULTS_DIR \
      -u CARTULARY_TEST_RUN_ID \
      -u CARTULARY_TEST_TARGET \
      -u OWNER -u ROWS -u VITEST_MAX_WORKERS -u BROWSER_LANES -u JSON \
      -u MAKEFLAGS -u MFLAGS \
      "$MAKE_BIN" --no-print-directory target-plan TARGET=not-a-real-target DETAIL=summary
  ) >"${target_dir}/stdout.log" 2>"${target_dir}/stderr.log"
  local status=$?
  set -e
  "$NODE_BIN" "${ROOT_DIR}/tools/harness/tests/real-command-output.mjs" invalid-usage \
    "$ROOT_DIR" target-plan "$target_dir" "$status"
}

case "$case_name" in
  source-schema) RUN_TARGET_MAKE_ARGS="CHECK=schema" run_target source-check source-schema ;;
  backend-unit) RUN_TARGET_MAKE_ARGS="SUITE=backend-unit" run_target test-slice backend-unit ;;
  backend-unit-machine) RUN_TARGET_MAKE_ARGS="SUITE=backend-unit" run_machine_target test-slice backend-unit-machine ;;
  owner-json) RUN_TARGET_OUTPUT_MODE=summary RUN_TARGET_MAKE_ARGS="OWNER=harness.command_surface ROWS=harness.command_surface.unit.embedded_asset_layout JSON=1" run_machine_target test-slice owner-json ;;
  lint) run_target lint lint ;;
  build) run_target build build ;;
  build-machine) run_machine_target build build-machine ;;
  test-fast) run_target test-fast test-fast ;;
  check) run_target check check ;;
  check-machine) run_machine_target check check-machine ;;
  invalid-usage) run_invalid_usage_check ;;
esac
