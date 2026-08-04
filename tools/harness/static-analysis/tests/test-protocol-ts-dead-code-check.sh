#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
FALLOW_SCRIPT="$ROOT_DIR/node_modules/fallow/bin/fallow"
CHECK_SCRIPT="$ROOT_DIR/tools/harness/static-analysis/protocol-ts-dead-code-check.mjs"
fixture_root="$(mktemp -d)"

cleanup() {
  rm -rf "$fixture_root"
}
trap cleanup EXIT

mkdir -p \
  "$fixture_root/apps/web/src" \
  "$fixture_root/contracts/protocol-ts" \
  "$fixture_root/packages/protocol-ts/src/entrypoints" \
  "$fixture_root/packages/protocol-ts/src/generated"

cat >"$fixture_root/package.json" <<'JSON'
{
  "private": true,
  "workspaces": ["apps/web", "packages/*"]
}
JSON

cat >"$fixture_root/.fallowrc.json" <<'JSON'
{
  "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
  "entry": ["apps/web/src/main.ts"],
  "workspaces": { "patterns": ["apps/web", "packages/*"] },
  "ignorePatterns": ["packages/protocol-ts/src/generated/**"],
  "ignoreExports": [
    { "file": "packages/protocol-ts/src/index.ts", "exports": ["*"] }
  ],
  "publicPackages": ["@cartulary/protocol-ts"],
  "rules": {
    "unused-files": "warn",
    "unused-exports": "warn",
    "unused-types": "warn",
    "unused-dependencies": "off"
  },
  "overrides": [
    {
      "files": ["packages/protocol-ts/src/generated/**"],
      "rules": {
        "unused-files": "off",
        "unused-exports": "off",
        "unused-types": "off"
      }
    }
  ]
}
JSON

cat >"$fixture_root/apps/web/package.json" <<'JSON'
{
  "name": "@cartulary/web",
  "private": true,
  "dependencies": { "@cartulary/protocol-ts": "workspace:*" }
}
JSON

cat >"$fixture_root/apps/web/src/main.ts" <<'TS'
import { liveValue } from "@cartulary/protocol-ts";
import type { LiveType } from "@cartulary/protocol-ts";

export const appValue: LiveType = liveValue;
TS

cat >"$fixture_root/packages/protocol-ts/package.json" <<'JSON'
{
  "name": "@cartulary/protocol-ts",
  "private": true,
  "type": "module",
  "exports": {
    ".": "./src/index.ts",
    "./family": "./src/entrypoints/family.ts"
  }
}
JSON

cat >"$fixture_root/packages/protocol-ts/src/index.ts" <<'TS'
export { liveValue, type LiveType, unusedExport } from "./live";
TS

cat >"$fixture_root/packages/protocol-ts/src/live.ts" <<'TS'
export const liveValue = "live";
export type LiveType = string;
export const unusedExport = "unused export";
export const unusedValue = "unused value";
export type UnusedType = { readonly unused: true };
TS

cat >"$fixture_root/packages/protocol-ts/src/unused-file.ts" <<'TS'
const unreachable = "unused file";
void unreachable;
TS

cat >"$fixture_root/packages/protocol-ts/src/compatibility.compile.ts" <<'TS'
export const compileOnlyValue = "compile fixture";
export type CompileOnlyType = string;
TS

cat >"$fixture_root/packages/protocol-ts/src/generated/family-types.ts" <<'TS'
export type OwnerSelectedType = { readonly id: string };
export type OwnerSelectedFutureType = { readonly generation: number };
TS

cat >"$fixture_root/packages/protocol-ts/src/entrypoints/family.ts" <<'TS'
export type * from "../generated/family-types.js";
TS

cat >"$fixture_root/contracts/protocol-ts/frontend-entrypoints.v2.json" <<'JSON'
{
  "entrypoints": [
    {
      "specifier": "@cartulary/protocol-ts/family",
      "authored_path": "packages/protocol-ts/src/entrypoints/family.ts",
      "generated_module_allowlist": [
        "packages/protocol-ts/src/generated/family-types.ts"
      ]
    }
  ]
}
JSON

output="$fixture_root/result.json"
if "$NODE_BIN" "$CHECK_SCRIPT" \
  --root "$fixture_root" \
  --fallow-bin "$FALLOW_SCRIPT" \
  --output "$output" >/dev/null 2>"$fixture_root/stderr.log"; then
  echo "expected fixture dead-code gate to fail" >&2
  exit 1
fi

"$NODE_BIN" --input-type=module - "$output" <<'EOF'
import { readFileSync } from "node:fs";

const [output] = process.argv.slice(2);
const report = JSON.parse(readFileSync(output, "utf8"));
const text = JSON.stringify(report.findings);
for (const expected of ["unused-file.ts", "unusedExport", "unusedValue", "UnusedType"]) {
  if (!text.includes(expected)) {
    throw new Error(`missing expected dead-code finding ${expected}`);
  }
}
for (const forbidden of ["liveValue", "LiveType", "compileOnlyValue", "CompileOnlyType", "OwnerSelectedType", "OwnerSelectedFutureType"]) {
  if (text.includes(forbidden)) {
    throw new Error(`unexpected dead-code finding ${forbidden}`);
  }
}
if (report.authored_suppressions !== 0) {
  throw new Error("authored suppressions must remain zero");
}
if (report.automated_compile_fixture_exports !== 1) {
  throw new Error("compile fixture exemption was not derived automatically");
}
if (report.automated_generated_pass_through_exports !== 2) {
  throw new Error("generated pass-through exemption was not derived from the owner");
}
EOF
