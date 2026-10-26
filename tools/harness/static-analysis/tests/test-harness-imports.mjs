import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { collectHarnessImportBoundaryViolations } from "../harness-import-boundary.mjs";
test("current harness imports satisfy semantic owner boundaries", () => {
  const report = collectHarnessImportBoundaryViolations(path.resolve(import.meta.dirname, "../../../.."));
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
    for (const file of ["browser/index.mjs", "browser/private.mjs", "test-catalog/index.mjs", "test-catalog/private.mjs", "backend/private.mjs", "evidence-accounting/private.mjs"])
      write(`tools/harness/${file}`, "export const value = 1;\n");
    write("tools/harness/scheduler/adapters/browser.mjs", 'export { value } from "../../browser/private.mjs";\n');
    write("tools/harness/readiness/check.mjs", 'import { value } from "../browser/index.mjs";\nconst text = \'import "../browser/private.mjs"\';\n');
    save();
    assert.deepEqual(collectHarnessImportBoundaryViolations(root).violations, []);
    write("tools/harness/readiness/check.mjs", 'import "../browser/private.mjs";\nimport "../backend/private.mjs";\nimport "../evidence-accounting/private.mjs";\nimport "../core/private.mjs";\nimport "../frontend/private.mjs";\n');
    for (const owner of ["core", "frontend"])
      write(`tools/harness/${owner}/private.mjs`, "export {};\n");
    write("tools/harness/scheduler/unauthorized.mjs", 'import "../browser/index.mjs";\nimport "../test-catalog/private.mjs";\n');
    write("tools/harness/unrecognized/child.mjs", "export {};\n");
    const rules = new Set(collectHarnessImportBoundaryViolations(root).violations.map((entry) => entry.rule));
    for (const rule of ["forbidden_private_browser_import", "forbidden_private_backend_import", "forbidden_private_test_catalog_import", "forbidden_scheduler_private_browser_import", "forbidden_unknown_harness_owner_root", "forbidden_private_evidence_accounting_import", "forbidden_private_catch_all_import"])
      assert.ok(rules.has(rule), rule);
    registry.facades[0].paths.push("tools/harness/browser/missing.mjs");
    save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /missing facade/);
    registry.facades[0].paths.pop();
    registry.facades.push({ ...registry.facades[0], key: "duplicate" });
    save();
    assert.throws(() => collectHarnessImportBoundaryViolations(root), /more than one facade/);
  }
  finally {
    rmSync(root, { recursive: true, force: true });
  }
});
