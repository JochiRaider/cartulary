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

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir] = process.argv.slice(2);
const { runFallowBatch } = await import(pathToFileURL(
  `${rootDir}/tools/harness/static-analysis/fallow-static-cli.mjs`,
).href);

let active = 0;
let peak = 0;
const completed = [];
const delays = new Map([
  ["dead-code", 45],
  ["package-surface-dead-code", 10],
  ["dupes", 30],
  ["health", 20],
]);
const specs = [...delays.keys()].map((name) => ({
  reportRoot: "/fixture",
  name,
  args: [name],
  outputFile: `/fixture/${name}.json`,
}));
const results = await runFallowBatch(specs, async (_root, name, args, outputFile) => {
  active += 1;
  peak = Math.max(peak, active);
  await new Promise((resolve) => setTimeout(resolve, delays.get(name)));
  active -= 1;
  completed.push(name);
  if (name === "dupes") {
    throw new Error("simulated start failure");
  }
  return {
    name,
    command: ["fallow", ...args],
    outputFile,
    status: name === "dead-code" || name === "dupes" ? "fail" : "pass",
    exitCode: name === "dead-code" || name === "dupes" ? 1 : 0,
  };
});
if (peak !== 4) {
  throw new Error(`expected four overlapping Fallow children, got ${peak}`);
}
if (completed.join(",") === specs.map((spec) => spec.name).join(",")) {
  throw new Error("fixture must complete out of authored order");
}
if (results.map((result) => result.name).join(",") !== specs.map((spec) => spec.name).join(",")) {
  throw new Error("concurrent Fallow results did not retain authored order");
}
if (results.filter((result) => result.status === "fail").map((result) => result.name).join(",") !== "dead-code,dupes") {
  throw new Error("simultaneous Fallow failures were not retained in authored order");
}
EOF

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import { readFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const [rootDir] = process.argv.slice(2);
const { collectBlockingPackageSurfaceFindings } = await import(pathToFileURL(
  `${rootDir}/tools/harness/static-analysis/fallow-static-cli.mjs`,
).href);
const owner = JSON.parse(readFileSync(
  `${rootDir}/tools/fallow/reachability_owner.json`,
  "utf8",
));
const findings = collectBlockingPackageSurfaceFindings(
  {
    unused_types: [
      {
        path: "packages/ui-contracts/src/designTokens.ts",
        export_name: "CartularyDesignTokenVarName",
      },
      {
        path: "packages/ui-contracts/src/designTokens.ts",
        export_name: "CartularyDefaultThemeId",
      },
    ],
  },
  owner.blocking_package_surfaces,
);
if (findings.length !== 1 || findings[0].symbol !== "CartularyDesignTokenVarName") {
  throw new Error(`unexpected blocking package-surface findings: ${JSON.stringify(findings)}`);
}
EOF

bash "$ROOT_DIR/tools/harness/static-analysis/tests/test-protocol-ts-dead-code-check.sh"

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

reachability_root="$(mktemp -d)"
cleanup_paths+=("$reachability_root")

mkdir -p \
  "$reachability_root/apps/web/src/testing" \
  "$reachability_root/apps/web/src/measurement" \
  "$reachability_root/apps/web/public/assets/fonts/inter" \
  "$reachability_root/packages/example/src" \
  "$reachability_root/tools/fallow" \
  "$reachability_root/tools/harness/static-analysis" \
  "$reachability_root/tools/harness/test-support" \
  "$reachability_root/tools/release-evidence"

cat >"$reachability_root/package.json" <<'JSON'
{
  "private": true,
  "workspaces": [
    "apps/web",
    "packages/*"
  ],
  "devDependencies": {
    "@cyclonedx/cdxgen": "12.3.1",
    "unused-tool": "1.0.0"
  }
}
JSON

cat >"$reachability_root/apps/web/package.json" <<'JSON'
{
  "name": "@cartulary/web",
  "private": true,
  "devDependencies": {
    "vitest": "4.1.4",
    "vite": "8.0.8"
  }
}
JSON

cat >"$reachability_root/packages/example/package.json" <<'JSON'
{
  "name": "@cartulary/example",
  "private": true
}
JSON

cat >"$reachability_root/.fallowrc.json" <<'JSON'
{
  "$schema": "https://raw.githubusercontent.com/fallow-rs/fallow/main/schema.json",
  "entry": [
    "apps/web/src/main.tsx"
  ],
  "workspaces": {
    "patterns": ["apps/web", "packages/*"]
  },
  "rules": {
    "unused-files": "warn",
    "unused-exports": "warn",
    "unused-types": "warn",
    "unused-dependencies": "off",
    "unused-dev-dependencies": "warn",
    "unlisted-dependencies": "off",
    "unresolved-imports": "warn",
    "policy-violation": "off"
  }
}
JSON

cat >"$reachability_root/apps/web/vite.config.ts" <<'TS'
export default {};
TS

cat >"$reachability_root/apps/web/index.html" <<'HTML'
<html>
  <head>
    <link rel="preload" href="/assets/fonts/inter/InterVariable.woff2" as="font" />
    <link rel="stylesheet" href="/assets/fonts/fonts.css" />
  </head>
  <body>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
HTML

cat >"$reachability_root/apps/web/measurement.html" <<'HTML'
<html><body><script src="./src/measurement/main.ts" type="module"></script></body></html>
HTML

cat >"$reachability_root/apps/web/src/measurement/main.ts" <<'TS'
import { fixture } from "./fixture";
console.log(fixture);
TS

cat >"$reachability_root/apps/web/src/measurement/fixture.ts" <<'TS'
export const fixture = "measurement";
TS

cat >"$reachability_root/apps/web/src/measurement/unreachable.ts" <<'TS'
export const unusedMeasurement = true;
TS

cat >"$reachability_root/apps/web/public/assets/fonts/fonts.css" <<'CSS'
@font-face {
  font-family: Inter;
  src: url("./inter/InterVariable.woff2") format("woff2");
}
CSS

printf 'font\n' >"$reachability_root/apps/web/public/assets/fonts/inter/InterVariable.woff2"

cat >"$reachability_root/apps/web/src/main.tsx" <<'TS'
import { connect, usedValue, type Reader } from "./surface";
const reader: Reader = { read: () => usedValue };
connect(reader, (value) => console.log(value));
export const app = "cartulary";
TS

cat >"$reachability_root/apps/web/src/surface.ts" <<'TS'
export interface Reader { read(): string }
export function connect(reader: Reader, callback: (value: string) => void) {
  callback(reader.read());
}
export const usedValue = "used";
export const unusedValue = "dead";
export type UnusedType = { dead: true };
TS

cat >"$reachability_root/apps/web/src/fixtureSupport.ts" <<'TS'
export const testOnly = "test";
TS

cat >"$reachability_root/apps/web/src/surface.test.ts" <<'TS'
import { testOnly } from "./fixtureSupport";
if (testOnly !== "test") throw new Error("fixture failed");
TS

cat >"$reachability_root/apps/web/src/testing/testSetup.ts" <<'TS'
export const setup = "vitest";
TS

cat >"$reachability_root/apps/web/src/testing/testSetup.dom.ts" <<'TS'
export const setupDom = "vitest-dom";
TS

cat >"$reachability_root/tools/harness/static-analysis/example-cli.mjs" <<'JS'
export const exampleCli = true;
JS

cat >"$reachability_root/tools/harness/static-analysis/example-test.mjs" <<'JS'
export const exampleTest = true;
JS

cat >"$reachability_root/tools/harness/test-support/example-dynamic.mjs" <<'JS'
export const importedDynamicPeer = true;
export const modeledDynamicExport = true;
export const ordinaryUnusedExport = true;
JS

cat >"$reachability_root/tools/harness/test-support/example-direct.mjs" <<'JS'
import { importedDynamicPeer } from "./example-dynamic.mjs";

export const exampleDirect = importedDynamicPeer;
JS

cat >"$reachability_root/tools/release-evidence/generate-sbom-license-evidence.mjs" <<'JS'
import { spawnSync } from "node:child_process";

export function runCdxgen(pnpm = "pnpm") {
  return spawnSync(pnpm, ["exec", "cdxgen", "--version"]);
}
JS

cat >"$reachability_root/tools/task_surface_owner.json" <<'JSON'
{
  "schema_id": "cartulary.task_surface_owner.v1",
  "targets": [
    {
      "name": "example-tool",
      "backing_scripts": [
        "tools/harness/static-analysis/example-cli.mjs"
      ]
    }
  ],
  "harness_checks": [
    {
      "name": "harness-smoke-example",
      "backing_scripts": [
        "tools/harness/static-analysis/example-test.mjs"
      ],
      "command": [
        "node",
        "./tools/harness/static-analysis/example-test.mjs"
      ]
    }
  ],
  "make_recipes": {
    "example-tool": {
      "type": "node_tool"
    }
  }
}
JSON

cat >"$reachability_root/tools/fallow/reachability_owner.json" <<'JSON'
{
  "schema_id": "cartulary.fallow_reachability_owner.v1",
  "base_config": {
    "path": ".fallowrc.json"
  },
  "blocking_web_surface": {
    "root": "apps/web",
    "rules": ["unused_files", "unused_exports", "unused_types"]
  },
  "blocking_package_surfaces": {
    "packages": [{
      "package_name": "@cartulary/example",
      "entrypoint": "apps/web/src/main.tsx",
      "own_test_globs": ["packages/example/src/**/*.test.ts"]
    }],
    "rules": ["unused_exports", "unused_types"]
  },
  "task_surface": {
    "owner_path": "tools/task_surface_owner.json",
    "script_extensions": [".cjs", ".cts", ".js", ".jsx", ".mjs", ".mts", ".ts", ".tsx"],
    "required_node_tool_backing_scripts": true
  },
  "harness_entrypoints": {
    "files": [
      "tools/harness/test-support/example-direct.mjs"
    ]
  },
  "harness_dynamic_exports": [
    {
      "file": "tools/harness/test-support/example-dynamic.mjs",
      "exports": ["modeledDynamicExport"],
      "owner": "fixture shell dynamic import",
      "evidence": "fixture coverage for reachability-owner dynamic export handling",
      "removal_trigger": "remove when fixture no longer needs dynamic export modeling"
    }
  ],
  "vitest": {
    "config_file": "apps/web/vite.config.ts",
    "setup_files": [
      "apps/web/src/testing/testSetup.ts",
      "apps/web/src/testing/testSetup.dom.ts"
    ],
    "test_entry_globs": [
      "apps/web/src/**/*.test.ts",
      "apps/web/src/**/*.test.tsx"
    ]
  },
  "vite_public_assets": {
    "public_root": "apps/web/public",
    "html_entry_files": ["apps/web/index.html", "apps/web/measurement.html"],
    "url_prefixes": ["/assets/"],
    "always_used_files": ["apps/web/public/assets/fonts/fonts.css"]
  },
  "executable_tooling_dependencies": [
    {
      "package_name": "@cyclonedx/cdxgen",
      "owner_script": "tools/release-evidence/generate-sbom-license-evidence.mjs",
      "command": ["pnpm", "exec", "cdxgen"]
    }
  ],
  "static_policy": {
    "runtime_enabled": false,
    "per_file_suppression_growth": false
  },
  "extensions": {}
}
JSON

resolved_config="$reachability_root/resolved-fallowrc.json"
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$reachability_root" "$resolved_config" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir, fixtureRoot, resolvedConfig] = process.argv.slice(2);
const { buildResolvedFallowConfig } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-reachability.mjs`).href
);
const result = buildResolvedFallowConfig({
  root: fixtureRoot,
  outputFile: resolvedConfig,
});
if (result.stats.vite_module_entry_points !== 2 ||
    !result.config.entry.includes("apps/web/src/measurement/main.ts")) {
  throw new Error("expected production and measurement module roots from HTML");
}
if (result.stats.task_surface_entry_points < 2) {
  throw new Error("expected task-surface scripts in resolved Fallow config");
}
if (result.stats.harness_entry_points !== 1) {
  throw new Error("expected harness entrypoint scripts in resolved Fallow config");
}
if (result.stats.harness_dynamic_export_files !== 1 || result.stats.harness_dynamic_exports !== 1) {
  throw new Error("expected harness dynamic export modeling in resolved Fallow config");
}
EOF

reachability_output="$reachability_root/fallow-reachability.json"
"$NODE_BIN" "$FALLOW_SCRIPT" dead-code \
  --root "$reachability_root" \
  --config "$resolved_config" \
  --format json \
  --quiet \
  --no-cache \
  --output-file "$reachability_output" >/dev/null

"$NODE_BIN" --input-type=module - "$reachability_output" "$ROOT_DIR" "$reachability_root" <<'EOF'
import assert from "node:assert/strict";
import { readFileSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const [file, rootDir, fixtureRoot] = process.argv.slice(2);
const { collectBlockingWebSurfaceFindings, hasUsableFallowOutput } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-static-cli.mjs`).href
);
const owner = JSON.parse(readFileSync(`${fixtureRoot}/tools/fallow/reachability_owner.json`, "utf8"));
const data = JSON.parse(readFileSync(file, "utf8"));
const blocked = collectBlockingWebSurfaceFindings(data, owner.blocking_web_surface);
assert.deepEqual(blocked.map(({rule, path, symbol}) => [rule, path, symbol]), [
  ["unused_files", "apps/web/src/measurement/unreachable.ts", null],
  ["unused_exports", "apps/web/src/surface.ts", "unusedValue"],
  ["unused_types", "apps/web/src/surface.ts", "UnusedType"],
]);
const clean = { ...data, unused_files: [], unused_exports: [], unused_types: [],
  unused_class_members: [{path: "apps/web/src/surface.ts", name: "read"}],
  duplicate_exports: [{path: "apps/web/src/surface.ts", name: "connect"}] };
assert.deepEqual(collectBlockingWebSurfaceFindings(clean, owner.blocking_web_surface), []);
assert.throws(() => collectBlockingWebSurfaceFindings({}, owner.blocking_web_surface), /missing unused_files/);
assert.throws(() => collectBlockingWebSurfaceFindings({...clean, unused_files: [{path: "apps/web/../escape.ts"}]}, owner.blocking_web_surface), /invalid relative path/);
assert.equal(hasUsableFallowOutput({outputFile: file, exitCode: 0}), true);
assert.equal(hasUsableFallowOutput({outputFile: file, exitCode: 1}), true);
for (const exitCode of [2, 127, 128])
  assert.equal(hasUsableFallowOutput({outputFile: file, exitCode}), false);
const malformed = `${fixtureRoot}/malformed.json`;
writeFileSync(malformed, "{broken");
assert.equal(hasUsableFallowOutput({outputFile: malformed, exitCode: 0}), false);
writeFileSync(malformed, "{}");
assert.equal(hasUsableFallowOutput({outputFile: malformed, exitCode: 0}), false);
writeFileSync(malformed, '{"error":true}');
assert.equal(hasUsableFallowOutput({outputFile: malformed, exitCode: 1}), false);
const text = JSON.stringify(data);
for (const file of ["main.ts", "fixture.ts"]) {
  if (text.includes(`apps/web/src/measurement/${file}`)) {
    throw new Error(`live measurement module reported unused: ${file}`);
  }
}
if (!text.includes("apps/web/src/measurement/unreachable.ts")) {
  throw new Error("unreachable measurement peer must remain detectable");
}
if (text.includes("apps/web/src/testing/testSetup.ts")) {
  throw new Error("expected Vitest setup file to be owner-reachable");
}
if (text.includes("apps/web/src/testing/testSetup.dom.ts")) {
  throw new Error("expected DOM Vitest setup file to be owner-reachable");
}
if (text.includes("apps/web/public/assets/fonts/fonts.css")) {
  throw new Error("expected Vite public stylesheet to be owner-reachable");
}
if (text.includes('"/assets/fonts/fonts.css"')) {
  throw new Error("expected Vite public asset URL to be validated before Fallow");
}
if (text.includes("@cyclonedx/cdxgen")) {
  throw new Error("expected executable tooling dependency to be owner-reachable");
}
if (text.includes("tools/harness/test-support/example-direct.mjs")) {
  throw new Error("expected harness entrypoint script to be owner-reachable");
}
if (text.includes("modeledDynamicExport")) {
  throw new Error("expected harness dynamic export to be modeled by config");
}
if (!text.includes("ordinaryUnusedExport")) {
  throw new Error("expected unrelated unused export to remain reported");
}
if (!text.includes("unused-tool")) {
  throw new Error("expected unrelated unused dev dependency to remain reported");
}
EOF

missing_asset_root="$(mktemp -d)"
cleanup_paths+=("$missing_asset_root")
cp -R "$reachability_root/." "$missing_asset_root/"
rm -f "$missing_asset_root/apps/web/public/assets/fonts/fonts.css"
if "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$missing_asset_root" "$missing_asset_root/resolved.json" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir, fixtureRoot, resolvedConfig] = process.argv.slice(2);
const { buildResolvedFallowConfig } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-reachability.mjs`).href
);
buildResolvedFallowConfig({ root: fixtureRoot, outputFile: resolvedConfig });
EOF
then
  fail "expected missing public asset to fail Fallow reachability owner validation"
fi

missing_script_root="$(mktemp -d)"
cleanup_paths+=("$missing_script_root")
cp -R "$reachability_root/." "$missing_script_root/"
rm -f "$missing_script_root/tools/harness/static-analysis/example-cli.mjs"
if "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$missing_script_root" "$missing_script_root/resolved.json" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir, fixtureRoot, resolvedConfig] = process.argv.slice(2);
const { buildResolvedFallowConfig } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-reachability.mjs`).href
);
buildResolvedFallowConfig({ root: fixtureRoot, outputFile: resolvedConfig });
EOF
then
  fail "expected missing task-surface backing script to fail Fallow reachability owner validation"
fi

missing_setup_root="$(mktemp -d)"
cleanup_paths+=("$missing_setup_root")
cp -R "$reachability_root/." "$missing_setup_root/"
rm -f "$missing_setup_root/apps/web/src/testing/testSetup.ts"
if "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$missing_setup_root" "$missing_setup_root/resolved.json" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir, fixtureRoot, resolvedConfig] = process.argv.slice(2);
const { buildResolvedFallowConfig } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-reachability.mjs`).href
);
buildResolvedFallowConfig({ root: fixtureRoot, outputFile: resolvedConfig });
EOF
then
  fail "expected missing Vitest setup file to fail Fallow reachability owner validation"
fi

missing_entrypoint_root="$(mktemp -d)"
cleanup_paths+=("$missing_entrypoint_root")
cp -R "$reachability_root/." "$missing_entrypoint_root/"
rm -f "$missing_entrypoint_root/tools/harness/test-support/example-direct.mjs"
if "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$missing_entrypoint_root" "$missing_entrypoint_root/resolved.json" <<'EOF'
import { pathToFileURL } from "node:url";

const [rootDir, fixtureRoot, resolvedConfig] = process.argv.slice(2);
const { buildResolvedFallowConfig } = await import(
  pathToFileURL(`${rootDir}/tools/harness/static-analysis/fallow-reachability.mjs`).href
);
buildResolvedFallowConfig({ root: fixtureRoot, outputFile: resolvedConfig });
EOF
then
  fail "expected missing harness entrypoint to fail Fallow reachability owner validation"
fi
