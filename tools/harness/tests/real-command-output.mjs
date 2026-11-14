import assert from "node:assert/strict";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { parseStrictJSON, readRetainedArtifactBytes, validateSchemaSync } from "../contract/index.mjs";
import { validateRunId } from "../contract/harness-contract.mjs";
import { readCanonicalUnitEvents } from "../evidence-accounting/index.mjs";

const decoder = new TextDecoder("utf-8", { fatal: true, ignoreBOM: true });
const maximumJSONBytes = 16 * 1024 * 1024;
const read = (root, relative) => readRetainedArtifactBytes(root, relative, maximumJSONBytes);
const decode = (bytes) => decoder.decode(bytes);
const json = (root, relative, schema) => {
  const value = parseStrictJSON(decode(read(root, relative)), relative);
  validateSchemaSync(schema, value);
  return value;
};
const lineCount = (text) => text === "" ? 0 : text.replace(/\n$/u, "").split("\n").length;

// This matrix exercises graph-backed commands only. The expected identity comes
// from the invocation, never from stdout or a summary that is being checked.
export async function assertRealCommandOutput({ repoRoot, targetName, mode, runRoot, runID, stdout, stderr }) {
  assert.ok(path.isAbsolute(runRoot), "expected run root must be absolute");
  assert.equal(path.basename(runRoot), validateRunId(runID));
  const policy = parseStrictJSON(decode(read(repoRoot, "tools/task_surface_manifest.json")))
    .targets.find((entry) => entry.name === targetName);
  assert.equal(policy?.output_policy?.summary_schema, "cartulary.harness_run_summary.v1");
  assert.equal(stderr.length, 0, `${targetName}: successful stderr must be empty`);
  const output = decode(stdout);
  if (mode === "summary") {
    const budget = policy.output_policy.success_budget;
    assert.ok(lineCount(output) <= budget.stdout_lines, "human stdout line budget");
    assert.ok(stdout.length <= budget.stdout_bytes, "human stdout byte budget");
    const lines = output.split("\n").filter((line) => line.startsWith("[GRAPH] "));
    assert.equal(lines.length, 1, "expected one graph summary");
    const field = (name) => {
      const values = [...lines[0].matchAll(new RegExp(`(?:^| )${name}=([^ ]+)`, "gu"))];
      assert.equal(values.length, 1, `expected one graph ${name}`);
      return values[0][1];
    };
    assert.equal(field("target"), targetName);
    assert.equal(field("status"), "pass");
    assert.equal(path.resolve(repoRoot, field("run_root")), runRoot, "reported run root must match invocation");
  } else {
    assert.equal(mode, "machine", "closed assertion mode");
    assert.ok(output.startsWith("{") && output.endsWith("}\n"), "machine output must be one object followed by one LF");
  }
  const manifest = json(runRoot, "run-manifest.json", "cartulary.harness_run_manifest.v3");
  const run = json(runRoot, "run-summary.json", "cartulary.harness_run_summary.v1");
  const target = json(runRoot, `target-summaries/${targetName}.json`, "cartulary.harness_target_summary.v1");
  assert.equal(manifest.run_id, runID, "manifest run identity");
  assert.equal(run.run_id, runID, "summary run identity");
  assert.equal(manifest.target, targetName);
  assert.equal(manifest.command_id, policy.command_id);
  assert.equal(target.command_id, policy.command_id);
  for (const summary of [run, target]) {
    assert.equal(summary.target, targetName);
    assert.equal(summary.status, "pass");
    assert.equal(summary.failure_class, null);
    assert.equal(summary.failure_reason, null);
  }
  if (mode === "machine") {
    const published = parseStrictJSON(output, "machine stdout");
    validateSchemaSync("cartulary.harness_run_summary.v1", published);
    assert.deepEqual(published, run, "machine output must equal its retained summary");
  }
  // Keep the streaming owner: large valid runs must not become a single string.
  for await (const _event of readCanonicalUnitEvents(path.join(runRoot, "unit-events.ndjson"))) continue;
}

export function assertInvalidTargetOutput({ status, stdout, stderr }) {
  assert.equal(status, 2, "invalid target-plan must exit 2");
  assert.equal(stdout.length, 0, "invalid target-plan must not emit partial JSON");
  assert.ok(stderr.length > 0 && stderr.length <= 8192, "bounded usage diagnostic");
  const diagnostic = decode(stderr);
  assert.match(diagnostic, /failure_reason=usage_error/u);
  assert.match(diagnostic, /TARGET must name a declared target/u);
}

if (process.argv[1] && import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href) {
  const [mode, repoRoot, targetName, captureRoot, runRoot, runID] = process.argv.slice(2);
  const stdout = read(captureRoot, "stdout.log");
  const stderr = read(captureRoot, "stderr.log");
  if (mode === "invalid-usage") {
    assertInvalidTargetOutput({ status: Number(runRoot), stdout, stderr });
  } else {
    await assertRealCommandOutput({ repoRoot, targetName, mode, runRoot, runID, stdout, stderr });
  }
}
