#!/usr/bin/env bash
set -euo pipefail
unset NODE_OPTIONS NODE_PATH DEBUG PWDEBUG
for review_env_name in ${!OTEL_@}; do unset "$review_env_name"; done
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../../.." && pwd)"
node="$root/tmp/node-runtime/bin/node"
condition=missing
expected_version="$(sed -nE 's/^[[:space:]]*"node_version": *"([0-9]+\.[0-9]+\.[0-9]+)",?$/\1/p' "$root/tools/toolchain_pins.json")"
installed_version=""
if [[ -x "$node" ]]; then condition=incompatible; installed_version="$("$node" --version 2>/dev/null || true)"; fi
if [[ -z "$expected_version" || "$installed_version" != "v$expected_version" ]]; then
  command_id="${1//-/_}"
  printf '{"schema_id":"cartulary.ui_review_command_result.v2","command_id":"cartulary.harness.command.%s.v1","session_id":null,"operation_id":null,"state":null,"epoch":null,"status":"error","exit_code":2,"failures":[{"failure_class":"config","failure_reason":"configuration_error","diagnostic_code":"tool_configuration","phase":"prerequisites","subject_id":"node","condition":"%s","recovery_id":"bootstrap_node"}],"receipt":null,"bundle_id":null,"private_refs":[]}\n' "$command_id" "$condition"
  exit 2
fi
exec "$node" "$root/tools/harness/browser/ui-review/cli.mjs" "$@"
