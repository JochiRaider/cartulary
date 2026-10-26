import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { buildResolvedFallowConfig } from "../fallow-reachability.mjs";
import { collectBlockingWebSurfaceFindings } from "../fallow-policy.mjs";
import { createE2EReachabilityFixture, createReachabilityFixture, createRulepackFixture } from "./fallow-fixtures.mjs";
import { repository } from "./fixtures.mjs";

function analyze(fixture, config, args = []) {
  const child = spawnSync(process.execPath, [
    path.join(repository, "node_modules/fallow/bin/fallow"), "dead-code",
    "--root", fixture.root, "--config", config, "--format", "json", "--quiet", "--no-cache",
    "--output-file", path.join(fixture.root, "report.json"), ...args,
  ], { encoding: "utf8", timeout: 30000 });
  assert.equal(child.error, undefined);
  assert.equal(child.status, 0, child.stdout + child.stderr);
  return fixture.readJSON("report.json");
}

test("Fallow analyzer enforces the authored boundary rulepack", (t) => {
  const fixture = createRulepackFixture(t);
  const text = JSON.stringify(analyze(fixture, path.join(fixture.root, ".fallowrc.json"), ["--policy-violations"]));
  for (const expected of ["cartulary-boundaries", "apps-web-no-react-data-grid", "apps/web/src/bad.ts"]) assert.ok(text.includes(expected), expected);
  assert.equal(text.includes("packages/grid-adapter/src/index.tsx"), false);
});

test("Fallow analyzer preserves live roots and reports dead peers", (t) => {
  const fixture = createReachabilityFixture(t);
  const config = path.join(fixture.root, "resolved.json");
  buildResolvedFallowConfig({ root: fixture.root, outputFile: config });
  const report = analyze(fixture, config);
  const owner = fixture.readJSON("tools/fallow/reachability_owner.json");
  assert.deepEqual(collectBlockingWebSurfaceFindings(report, owner.blocking_web_surface).map(({ rule, path, symbol }) => [rule, path, symbol]), [
    ["unused_files", "apps/web/src/measurement/unreachable.ts", null],
    ["unused_exports", "apps/web/src/surface.ts", "unusedValue"],
    ["unused_types", "apps/web/src/surface.ts", "UnusedType"],
  ]);
  for (const file of ["packages/example/src/contract.compile.ts", "tools/harness/static-analysis/shell-child.mjs"]) assert.equal(report.unused_files.some((entry) => entry.path === file), false, file);
  for (const file of ["packages/example/src/orphan.ts", "tools/harness/static-analysis/orphan.mjs"]) assert.ok(report.unused_files.some((entry) => entry.path === file), file);
  const text = JSON.stringify(report);
  for (const live of [
    "apps/web/src/measurement/main.ts", "apps/web/src/measurement/fixture.ts",
    "apps/web/src/testing/testSetup.ts", "apps/web/src/testing/testSetup.dom.ts",
    "apps/web/public/assets/fonts/fonts.css", '"/assets/fonts/fonts.css"', "@cyclonedx/cdxgen",
    "tools/harness/test-support/example-direct.mjs", "modeledDynamicExport",
  ]) assert.equal(text.includes(live), false, live);
  for (const dead of ["apps/web/src/measurement/unreachable.ts", "ordinaryUnusedExport", "unused-tool"]) assert.ok(text.includes(dead), dead);
});

test("Fallow analyzer qualifies authored E2E roots and exposes blanket-root blindness", (t) => {
  const fixture = createE2EReachabilityFixture(t);
  const config = path.join(fixture.root, "resolved.json");
  buildResolvedFallowConfig({ root: fixture.root, outputFile: config });
  const report = analyze(fixture, config);
  const owner = fixture.readJSON("tools/fallow/reachability_owner.json");
  const e2eFindings = collectBlockingWebSurfaceFindings(report, owner.blocking_web_surface)
    .filter((entry) => entry.path.startsWith("apps/web/e2e/"));
  assert.deepEqual(e2eFindings.map(({ rule, path, symbol }) => [rule, path, symbol]), [
    ["unused_files", "apps/web/e2e/support/orphan.ts", null],
    ["unused_exports", "apps/web/e2e/support/client.ts", "unusedValue"],
    ["unused_types", "apps/web/e2e/support/client.ts", "UnusedType"],
  ]);
  const blanket = fixture.readJSON("resolved.json");
  blanket.entry.push("apps/web/e2e/**/*.ts");
  fixture.write("blanket.json", blanket);
  const hidden = analyze(fixture, path.join(fixture.root, "blanket.json"));
  assert.equal(hidden.unused_files.some((entry) => entry.path === "apps/web/e2e/support/orphan.ts"), false);
});
