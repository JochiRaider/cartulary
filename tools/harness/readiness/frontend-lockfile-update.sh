#!/usr/bin/env bash
set -euo pipefail
# shellcheck source=tools/workspace_layout.generated.sh
source "$(dirname -- "${BASH_SOURCE[0]}")/../../workspace_layout.generated.sh"
root_dir="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
pnpm="${PNPM:?PNPM is required}"
node_bin="${NODE_BIN:?NODE_BIN is required}"
# The frontend installation producer holds the same lock, beside its layout
# directory, while it installs.
lock="${root_dir}/${CARTULARY_LAYOUT_FRONTEND_INSTALL}.lock"
mkdir -p "$(dirname "$lock")"
exec {installation_lock}>"$lock"
flock -x "$installation_lock"
"$node_bin" "$(dirname "${BASH_SOURCE[0]}")/installed-readiness.mjs" invalidate
env CI=true "$pnpm" install --lockfile-only --ignore-scripts --no-frozen-lockfile --reporter=append-only --loglevel=warn
