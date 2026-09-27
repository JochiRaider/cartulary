#!/usr/bin/env bash
set -euo pipefail
unset NODE_OPTIONS NODE_PATH DEBUG PWDEBUG
for review_env_name in ${!OTEL_@}; do unset "$review_env_name"; done
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../../../.." && pwd)"
node="$root/tmp/node-runtime/bin/node"
if [[ ! -x "$node" ]]; then
  command_id="${1//-/_}"
  printf '{"schema_id":"cartulary.ui_review_command_result.v1","command_id":"cartulary.harness.command.%s.v1","session_id":null,"operation_id":null,"state":null,"epoch":null,"status":"error","exit_code":2,"failures":[{"failure_class":"config","failure_reason":"configuration_error","diagnostic_code":"tool_configuration"}],"receipt":null,"bundle_id":null,"private_refs":[]}\n' "$command_id"
  exit 2
fi
exec "$node" "$root/tools/harness/browser/ui-review/cli.mjs" "$@"
