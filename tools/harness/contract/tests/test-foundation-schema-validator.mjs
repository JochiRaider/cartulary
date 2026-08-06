#!/usr/bin/env node
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  cpSync,
  mkdirSync,
  mkdtempSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const testDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(testDir, "..", "..", "..", "..");
const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-foundation-validator-"));
try {
  const contractRoot = path.join(scratch, "tools", "harness", "contract");
  mkdirSync(path.join(contractRoot, "generated"), { recursive: true });
  for (const relative of [
    "tools/harness/contract/harness-contract.mjs",
    "tools/harness/contract/artifact-writer.mjs",
    "tools/harness/contract/generated/foundation-schema-validators.cjs",
  ]) {
    cpSync(path.join(repoRoot, relative), path.join(scratch, relative));
  }
  const probe = path.join(scratch, "probe.mjs");
  writeFileSync(
    probe,
    `import { validateSchemaSync } from "./tools/harness/contract/harness-contract.mjs";\n` +
      `const zeroCounts = { steps: 0, tests: 0, failed: 0, non_test: 0, non_test_failed: 0, packages: 0 };\n` +
      `const zeroAccounting = { authoritative: 0, support: 0, raw: 0, tooling_support: 0, unowned_regression: 0, unmapped: 0, authoritative_failed: 0, support_failed: 0, raw_failed: 0, tooling_support_failed: 0, unowned_regression_failed: 0, unmapped_failed: 0, missing: 0 };\n` +
      `validateSchemaSync("cartulary.tool_run_summary.v5", { schema_id: "cartulary.tool_run_summary.v5", target: "bootstrap-node-runtime", command: { cwd: ".", argv: ["make", "bootstrap-node-runtime"], make_target: "bootstrap-node-runtime", env: {} }, status: "pass", exit_code: 0, started_at: "2026-01-01T00:00:00Z", completed_at: "2026-01-01T00:00:00Z", duration_ms: 0, output_mode: "summary", result_root: ".cartulary/test-results", run_id: "foundation", run_root: ".cartulary/test-results/foundation", summary_artifacts: [], log_artifacts: [], work_units: [], evidence_targets: [], helper_units: [], counts: zeroCounts, step_accounting: zeroAccounting, failure_class: null, failure_reason: null, failures: [], slowest: [], warnings: [], rerun_commands: [], scheduler_timing: null, extensions: {} });\n` +
      `try { validateSchemaSync("cartulary.tool_run_summary.v5", {}); } catch (error) {\n` +
      `  if (!String(error.message).includes("cartulary.tool_run_summary.v5 validation failed")) throw error;\n` +
      `  process.stdout.write("foundation validator accepted a valid summary and rejected an invalid summary\\n");\n` +
      `  process.exit(0);\n` +
      `}\nthrow new Error("foundation validator accepted an invalid summary");\n`,
  );
  const output = execFileSync(process.execPath, [probe], {
    cwd: scratch,
    env: { HOME: scratch, PATH: process.env.PATH ?? "" },
    encoding: "utf8",
  });
  assert.match(output, /accepted a valid summary and rejected an invalid summary/u);
} finally {
  rmSync(scratch, { recursive: true, force: true });
}

process.stdout.write("foundation schema validator checks passed\n");
