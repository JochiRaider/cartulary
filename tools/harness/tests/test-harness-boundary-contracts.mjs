import { runContractSuite } from "./contract-suite-support.mjs";
runContractSuite("boundaries");

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../contract/index.mjs";
import { validateInspectorFieldLayouts } from "../generated-artifacts/design-presentation/design-presentation.mjs";

test("inspector presentation v2 rejects invalid authored field layouts", () => {
  const root = path.resolve(import.meta.dirname, "../../..");
  const presentation = JSON.parse(readFileSync(path.join(root, "contracts/design/presentation.v2.json"), "utf8"));
  const layouts = presentation.inspector.field_layout_overrides;
  const viewSchemas = path.join(root, "contracts/view-schemas");
  validateSchemaSync("cartulary.design_presentation.v2", presentation);
  validateInspectorFieldLayouts(layouts, viewSchemas);
  for (const candidate of [
    [...layouts, { ...layouts[0], layout: "narrative" }],
    [{ ...layouts[0], view_schema_id: "unknown" }],
    [{ ...layouts[0], field_key: "timeline.nonexistent" }],
    [{ ...layouts[0], layout: "unknown" }],
  ]) assert.throws(() => validateInspectorFieldLayouts(candidate, viewSchemas));
  const unknownLayout = structuredClone(presentation);
  unknownLayout.inspector.field_layout_overrides[0].layout = "unknown";
  assert.throws(() => validateSchemaSync("cartulary.design_presentation.v2", unknownLayout));
});

import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { collectHarnessImportBoundaryViolations } from "../static-analysis/harness-import-boundary.mjs";

test("current harness imports satisfy semantic owner boundaries", () => {
  const report = collectHarnessImportBoundaryViolations(path.resolve(import.meta.dirname, "../../.."));
  assert.deepEqual(report.violations, [], JSON.stringify(report.violations, null, 2));
});

test("harness boundaries reject private imports, unrelated exceptions, unknown owners and invalid facades", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "cartulary-harness-boundaries-"));
  const write = (name, body) => {
    const file = path.join(root, name);
    mkdirSync(path.dirname(file), { recursive: true });
    writeFileSync(file, body);
  };
  const registry = { schema_id: "cartulary.harness_helper_ownership.v1", facades: [
    { key: "browser_owner_facade", boundary_group: "browser", paths: ["tools/harness/browser/index.mjs"], allowed_consumers: ["tools/harness/scheduler/adapters/browser.mjs"] },
    { key: "test_catalog", boundary_group: "test_catalog", paths: ["tools/harness/test-catalog/index.mjs"], allowed_consumers: ["tools/harness/readiness/check.mjs"] },
  ] };
  const save = () => write("tools/harness_helper_ownership.json", JSON.stringify(registry));
  try {
    for (const file of ["browser/index.mjs", "browser/private.mjs", "test-catalog/index.mjs", "test-catalog/private.mjs", "backend/private.mjs", "evidence-accounting/private.mjs"]) write(`tools/harness/${file}`, "export const value = 1;\n");
    write("tools/harness/scheduler/adapters/browser.mjs", 'export { value } from "../../browser/private.mjs";\n');
    write("tools/harness/readiness/check.mjs", 'import { value } from "../browser/index.mjs";\nconst text = \'import "../browser/private.mjs"\';\n');
    save();
    assert.deepEqual(collectHarnessImportBoundaryViolations(root).violations, []);
    write("tools/harness/readiness/check.mjs", 'import "../browser/private.mjs";\nimport "../backend/private.mjs";\nimport "../evidence-accounting/private.mjs";\nimport "../core/private.mjs";\nimport "../frontend/private.mjs";\n');
    for (const owner of ["core", "frontend"]) write(`tools/harness/${owner}/private.mjs`, "export {};\n");
    write("tools/harness/scheduler/unauthorized.mjs", 'import "../browser/index.mjs";\nimport "../test-catalog/private.mjs";\n');
    write("tools/harness/unrecognized/child.mjs", "export {};\n");
    const rules = new Set(collectHarnessImportBoundaryViolations(root).violations.map((entry) => entry.rule));
    for (const rule of ["forbidden_private_browser_import", "forbidden_private_backend_import", "forbidden_private_test_catalog_import", "forbidden_scheduler_private_browser_import", "forbidden_unknown_harness_owner_root", "forbidden_private_evidence_accounting_import", "forbidden_private_catch_all_import"]) assert.ok(rules.has(rule), rule);
    registry.facades[0].paths.push("tools/harness/browser/missing.mjs"); save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /missing facade/);
    registry.facades[0].paths.pop();
    registry.facades.push({ ...registry.facades[0], key: "duplicate" }); save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /more than one facade/);
  } finally { rmSync(root, { recursive: true, force: true }); }
});


import { spawnSync } from "node:child_process";
import { browserAcquisitionLaunchArguments } from "../scheduler/adapters/browser.mjs";

test("browser adapter launches the owning acquisition dispatcher and rejects invalid ownership", () => {
  const command = browserAcquisitionLaunchArguments("missing-acquisition", "invalid-launch", process.execPath, ["--version"]);
  const childEnv = { ...process.env };
  delete childEnv.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  const child = spawnSync(process.execPath, command, { encoding: "utf8", env: childEnv });
  assert.notEqual(child.status, 0, "an import-only facade must never count as a successful browser producer");
  assert.match(child.stderr, /browser acquisition launch failed/);
});
