#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"

usage() {
  echo "usage: run-service-backed-schedule-target.sh --target <target> --phase-label <label> [--projection <target>] --service-wrapper <test-services|none>" >&2
  exit 2
}

target=""
phase_label=""
projection=""
service_wrapper=""

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --target)
      target="${2:-}"
      shift 2
      ;;
    --phase-label)
      phase_label="${2:-}"
      shift 2
      ;;
    --projection)
      projection="${2:-}"
      shift 2
      ;;
    --service-wrapper)
      service_wrapper="${2:-}"
      shift 2
      ;;
    *)
      usage
      ;;
  esac
done

if [[ -z "$target" || -z "$phase_label" || -z "$service_wrapper" ]]; then
  usage
fi

if [[ -z "$projection" ]]; then
  projection="$target"
fi

node_bin="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}"
if [[ ! -x "$node_bin" ]]; then
  node_bin="node"
fi

make_bin="${MAKE:-make}"
test_output_script="${TEST_OUTPUT_SCRIPT:-$ROOT_DIR/scripts/lib/test-output.sh}"
task_surface_manifest="${TASK_SURFACE_MANIFEST:-$ROOT_DIR/tools/task_surface_manifest.json}"
run_phase_script="${RUN_PHASE_SCRIPT:-$ROOT_DIR/scripts/lib/run-phase.sh}"
schedule_script="${RUN_SERVICE_BACKED_SCHEDULE_SCRIPT:-$ROOT_DIR/scripts/run-service-backed-schedule.mjs}"
schedule_manifest="${SERVICE_BACKED_SCHEDULE_MANIFEST:-$ROOT_DIR/tools/service_backed_schedule_manifest.json}"

scheduler_command=(
  env
  "MAKE=$make_bin"
  "NODE_BIN=$node_bin"
  "TEST_OUTPUT_SCRIPT=$test_output_script"
  "TASK_SURFACE_MANIFEST=$task_surface_manifest"
  "$run_phase_script" "$phase_label" --
  "$node_bin" "$schedule_script"
  --target "$target"
  --manifest "$schedule_manifest"
  --defer-summary
)

status=0
case "$service_wrapper" in
  test-services)
    test_services_bin="${TEST_SERVICES_BIN:-}"
    if [[ -z "$test_services_bin" ]]; then
      echo "TEST_SERVICES_BIN is required for --service-wrapper test-services" >&2
      exit 2
    fi
    "$test_services_bin" run -- "${scheduler_command[@]}" || status=$?
    ;;
  none)
    "${scheduler_command[@]}" || status=$?
    ;;
  *)
    usage
    ;;
esac

if [[ "$status" -eq 0 ]]; then
  requested=pass
else
  requested=fail
fi

summary_status=0
NODE_BIN="$node_bin" TASK_SURFACE_MANIFEST="$task_surface_manifest" \
  "$test_output_script" target-summary "$target" "$requested" --projection "$projection" || summary_status=$?

if [[ "$status" -ne 0 ]]; then
  exit "$status"
fi
exit "$summary_status"
