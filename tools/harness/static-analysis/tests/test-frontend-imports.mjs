import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import { spawnSync } from "node:child_process";
test("frontend owner boundaries survive aliasing extraction and comments", () => {
  const repository = path.resolve(import.meta.dirname, "../../../..");
  const scratch = mkdtempSync(path.join(os.tmpdir(), "frontend-owner-boundaries-"));
  const owner = JSON.parse(readFileSync(path.join(repository, "tools/frontend_import_boundaries.json"), "utf8"));
  const identities = ["frontend-grid-vendor-boundary", "frontend-timeline-no-assessment-owner", "frontend-assessment-owner-no-timeline-imports", "frontend-workbook-policy-transport-boundary"];
  const rules = owner.rules.filter((rule) => identities.includes(rule.id));
  assert.equal(rules.length, identities.length);
  const cases = [
    ["workbook/timeline/models/renamed.ts", 'import type { AssessmentCreateDraft as Local } from "../../models/assessmentWorkbookModel"; export type Extracted = Local;', "frontend-timeline-no-assessment-owner"],
    ["workbook/components/AssessmentWorkbookSurface.tsx", 'export { row as Renamed } from "../timeline/models/row";', "frontend-assessment-owner-no-timeline-imports"],
    ["networkFlow/renamed.ts", 'export const load = () => import("react-data-grid");', "frontend-grid-vendor-boundary"],
    ["workbook/policies/renamed.ts", 'import * as local from "../../services/browserApi"; export const capability = local;', "frontend-workbook-policy-transport-boundary"],
    ["workbook/timeline/models/renamed.ts", '// AssessmentCreateDraft react-data-grid authorization\nimport type { Row as Local } from "./row"; export type Extracted = Local;', null],
  ];
  try {
    for (const [index, [file, source, expected]] of cases.entries()) {
      const root = path.join(scratch, String(index));
      const target = path.join(root, "apps/web/src", file);
      mkdirSync(path.dirname(target), { recursive: true });
      writeFileSync(target, source);
      writeFileSync(path.join(root, "boundaries.json"), JSON.stringify({
        schema_id: owner.schema_id, scan_roots: ["apps/web/src"], scan_excludes: [],
        acyclic_import_graphs: [], singleton_imports: [], rules, raw_design_token_literal_checks: [],
      }));
      const childEnv = { ...process.env };
      delete childEnv.CARTULARY_TEST_TARGET;
      const child = spawnSync(process.execPath, [path.join(repository, "tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs"), "--root", root, "--config", "boundaries.json"], { env: childEnv, encoding: "utf8", timeout: 10000 });
      assert.equal(child.error, undefined);
      assert.equal(child.status, expected === null ? 0 : 1, child.stdout + child.stderr);
      if (expected !== null)
        assert.ok(child.stderr.includes(expected), child.stderr);
    }
  }
  finally {
    rmSync(scratch, { recursive: true, force: true });
  }
});
