#!/usr/bin/env bash
set -euo pipefail
stamp="${FRONTEND_INSTALL_STAMP:?FRONTEND_INSTALL_STAMP is required}"
pnpm="${PNPM:?PNPM is required}"
node_bin="${NODE_BIN:?NODE_BIN is required}"
mkdir -p "$(dirname "$stamp")"
exec {installation_lock}>"$(dirname "$stamp")/install.lock"
flock -x "$installation_lock"
"$node_bin" "$(dirname "${BASH_SOURCE[0]}")/installed-readiness.mjs" invalidate
env CI=true "$pnpm" install --lockfile-only --ignore-scripts --no-frozen-lockfile --reporter=append-only --loglevel=warn
