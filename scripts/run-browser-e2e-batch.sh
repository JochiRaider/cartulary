#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/playwright-owned-stack.sh"

MANIFEST="${BROWSER_E2E_BATCH_MANIFEST:-$ROOT_DIR/tools/browser_e2e_batch_manifest.json}"
TEST_OUTPUT_HELPER="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/scripts/lib/test-output.sh}"

usage() {
  echo "usage: run-browser-e2e-batch.sh <stage>" >&2
  exit 2
}

if [[ "$#" -ne 1 ]]; then
  usage
fi

stage="$1"
node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

resolve_playwright_owned_stack_env "$ROOT_DIR"

stage_children="$(
  "$node_bin" - "$MANIFEST" "$stage" <<'EOF'
const fs = require("node:fs");

const [manifestPath, stageName] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
if (manifest.schema_id !== "cartulary.browser_e2e_batch_manifest.v2") {
  throw new Error(`${manifestPath} must declare schema_id cartulary.browser_e2e_batch_manifest.v2`);
}
const matches = (manifest.stages ?? []).filter((entry) => entry?.name === stageName);
if (matches.length !== 1) {
  throw new Error(`expected exactly one browser E2E batch stage ${stageName}, found ${matches.length}`);
}
const stage = matches[0];
const children = stage.children ?? [];
if (!Array.isArray(children) || children.length === 0) {
  throw new Error(`browser E2E batch stage ${stageName} must declare children[]`);
}
process.stdout.write(children.join(","));
EOF
)"

mapfile -t stage_groups < <(
  "$node_bin" - "$MANIFEST" "$stage" <<'EOF'
const fs = require("node:fs");

const [manifestPath, stageName] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
const stage = (manifest.stages ?? []).find((entry) => entry?.name === stageName);
if (!stage) {
  throw new Error(`missing browser E2E batch stage ${stageName}`);
}
const groups = stage.groups ?? [];
if (!Array.isArray(groups) || groups.length === 0) {
  throw new Error(`browser E2E batch stage ${stageName} must declare groups[]`);
}
const allowedKinds = new Set([
  "webserver-backed",
  "duration_balanced_specs",
  "functional",
  "support",
  "stateful",
  "measurement",
  "visual",
]);
for (const [index, group] of groups.entries()) {
  if (!group || typeof group !== "object" || Array.isArray(group)) {
    throw new Error(`browser E2E batch stage ${stageName} group ${index + 1} must be an object`);
  }
  for (const key of ["name", "target", "kind"]) {
    if (typeof group[key] !== "string" || group[key].trim() === "") {
      throw new Error(`browser E2E batch stage ${stageName} group ${index + 1} must declare ${key}`);
    }
  }
  if (!allowedKinds.has(group.kind)) {
    throw new Error(`browser E2E batch group ${group.name} has unsupported kind ${group.kind}`);
  }
  const workers = group.workers === undefined ? "default" : String(group.workers);
  const resetBefore = group.reset_before === undefined ? "" : String(group.reset_before);
  process.stdout.write([
    group.name,
    group.target,
    group.kind,
    workers,
    resetBefore,
  ].join("\t") + "\n");
}
EOF
)

run_target_summary() {
  local target="$1"
  local status="$2"
  local children="${3:-}"

  if [[ -n "$children" ]]; then
    NODE_BIN="$node_bin" "$TEST_OUTPUT_HELPER" target-summary "$target" "$status" --projection "$target"
    return $?
  fi

  NODE_BIN="$node_bin" "$TEST_OUTPUT_HELPER" target-summary "$target" "$status"
}

run_group() {
  local target="$1"
  local kind="$2"
  local workers="$3"
  shift 3

  local -a group_env=(
    env
    "CARTULARY_TEST_TARGET=$target"
    "NODE_BIN=$PLAYWRIGHT_OWNED_STACK_NODE_BIN"
  )

  if [[ "$workers" != "default" ]]; then
    group_env+=("PLAYWRIGHT_WORKERS=$workers")
  fi

  case "$kind" in
    webserver-backed)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-webserver-backed.sh"
      ;;
    duration_balanced_specs)
      if [[ "$target" == "browser-e2e-webserver-backed" ]]; then
        "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-webserver-backed.sh"
      else
        "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-functional.sh"
      fi
      ;;
    functional)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-functional.sh"
      ;;
    support)
      local -a support_env=(
        "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"
        "CARTULARY_TEST_TARGET=$target"
        "NODE_BIN=$PLAYWRIGHT_OWNED_STACK_NODE_BIN"
      )
      if [[ "$workers" != "default" ]]; then
        support_env+=("PLAYWRIGHT_WORKERS=$workers")
      fi
      "${support_env[@]}" \
        "$ROOT_DIR/scripts/lib/run-playwright-phase.sh" \
        "browser-e2e-support raw" \
        -- \
        "$PLAYWRIGHT_OWNED_STACK_PNPM_BIN" --dir apps/web exec playwright test \
        e2e/phase2.support.spec.ts e2e/phase3.support.spec.ts
      ;;
    stateful)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-stateful.sh"
      ;;
    measurement)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-measurement.sh"
      ;;
    visual)
      "${group_env[@]}" "$ROOT_DIR/scripts/run-browser-e2e-visual.sh"
      ;;
    *)
      echo "unsupported browser E2E batch group kind ${kind}" >&2
      return 2
      ;;
  esac
}

if [[ -n "${CARTULARY_TEST_TARGET:-}" && "${CARTULARY_TEST_TARGET}" == "$stage" ]]; then
  NODE_BIN="$node_bin" "$TEST_OUTPUT_HELPER" target-start "$stage" --children "$stage_children" || true
fi

overall_status=0
for group_row in "${stage_groups[@]}"; do
  IFS=$'\t' read -r group_name target kind workers reset_before <<<"$group_row"

  if [[ -n "$reset_before" ]]; then
    env CARTULARY_TEST_TARGET="${CARTULARY_TEST_TARGET:-$stage}" \
      NODE_BIN="$node_bin" \
      "$ROOT_DIR/scripts/reset-web-e2e-stack.sh" --label "$reset_before"
  fi

  set +e
  run_group "$target" "$kind" "$workers"
  group_status=$?
  set -e

  if [[ "$group_status" -eq 0 ]]; then
    run_target_summary "$target" pass || group_status=$?
  else
    run_target_summary "$target" fail || true
  fi

  if [[ "$group_status" -ne 0 && "$overall_status" -eq 0 ]]; then
    overall_status="$group_status"
  fi
done

if [[ -n "${CARTULARY_TEST_TARGET:-}" && "${CARTULARY_TEST_TARGET}" == "$stage" ]]; then
  if [[ "$overall_status" -eq 0 ]]; then
    run_target_summary "$stage" pass "$stage_children" || overall_status=$?
  else
    run_target_summary "$stage" fail "$stage_children" || true
  fi
fi

exit "$overall_status"
