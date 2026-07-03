#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
FALLOW_SCRIPT="$ROOT_DIR/node_modules/fallow/bin/fallow"
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

if [[ ! -f "$FALLOW_SCRIPT" ]]; then
  fail "missing Fallow package script at $FALLOW_SCRIPT; run make frontend-install"
fi

case_root="$(mktemp -d)"
cleanup_paths+=("$case_root")

mkdir -p \
  "$case_root/node_modules" \
  "$case_root/apps/web/src" \
  "$case_root/packages/grid-adapter/src" \
  "$case_root/tools/fallow"

cat >"$case_root/package.json" <<'JSON'
{
  "private": true,
  "workspaces": [
    "apps/web",
    "packages/*"
  ],
  "dependencies": {
    "react-data-grid": "7.0.0-beta.59"
  }
}
JSON

cat >"$case_root/apps/web/package.json" <<'JSON'
{
  "name": "@cartulary/web",
  "private": true,
  "dependencies": {
    "@cartulary/grid-adapter": "workspace:*"
  }
}
JSON

cat >"$case_root/packages/grid-adapter/package.json" <<'JSON'
{
  "name": "@cartulary/grid-adapter",
  "private": true,
  "dependencies": {
    "react-data-grid": "7.0.0-beta.59"
  }
}
JSON

cat >"$case_root/apps/web/src/bad.ts" <<'TS'
import DataGrid from "react-data-grid";

export const bad = DataGrid;
TS

cat >"$case_root/packages/grid-adapter/src/index.tsx" <<'TS'
import DataGrid from "react-data-grid";

export const allowed = DataGrid;
TS

cp "$ROOT_DIR/tools/fallow/cartulary-boundaries.rulepack.json" \
  "$case_root/tools/fallow/cartulary-boundaries.rulepack.json"

cat >"$case_root/.fallowrc.json" <<'JSON'
{
  "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
  "entry": [
    "apps/web/src/bad.ts",
    "packages/grid-adapter/src/index.tsx"
  ],
  "rules": {
    "unused-files": "off",
    "unused-exports": "off",
    "unused-types": "off",
    "unused-dependencies": "off",
    "unlisted-dependencies": "off",
    "unresolved-imports": "off",
    "policy-violation": "warn"
  },
  "rulePacks": ["tools/fallow/cartulary-boundaries.rulepack.json"]
}
JSON

output="$case_root/fallow-policy.json"
"$NODE_BIN" "$FALLOW_SCRIPT" dead-code \
  --root "$case_root" \
  --config "$case_root/.fallowrc.json" \
  --format json \
  --quiet \
  --no-cache \
  --policy-violations \
  --output-file "$output" >/dev/null

"$NODE_BIN" --input-type=module - "$output" <<'EOF'
import { readFileSync } from "node:fs";

const [file] = process.argv.slice(2);
const data = JSON.parse(readFileSync(file, "utf8"));
const matches = [];
const adapterMatches = [];

function walk(value) {
  if (!value || typeof value !== "object") {
    return;
  }
  if (Array.isArray(value)) {
    for (const item of value) {
      walk(item);
    }
    return;
  }
  const text = JSON.stringify(value);
  if (
    text.includes("cartulary-boundaries") &&
    text.includes("apps-web-no-react-data-grid") &&
    text.includes("apps/web/src/bad.ts")
  ) {
    matches.push(value);
  }
  if (
    text.includes("cartulary-boundaries") &&
    text.includes("apps-web-no-react-data-grid") &&
    text.includes("packages/grid-adapter/src/index.tsx")
  ) {
    adapterMatches.push(value);
  }
  for (const child of Object.values(value)) {
    walk(child);
  }
}

walk(data);
if (matches.length < 1) {
  throw new Error("expected apps/web react-data-grid policy violation");
}
if (adapterMatches.length !== 0) {
  throw new Error("expected no grid-adapter react-data-grid policy violation");
}
EOF
