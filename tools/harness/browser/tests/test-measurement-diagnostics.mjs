import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { linkSync, mkdirSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";
import test from "node:test";
import { artifactDigest, measurementContext, writeMeasurementArtifact } from "./measurement-fixtures.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
test("measurement CLI admits current artifacts and rejects historical generations without changing bytes", async (t) => {
    const { measurementEvidenceRoot, persist } = measurementContext(t);
    const summary = persist();
    const summaryArtifact = writeMeasurementArtifact(measurementEvidenceRoot, "inspection-summary.json", "frontend_measurement_summary", summary);
    const aggregateDirectory = path.join(measurementEvidenceRoot, "browser-e2e-measurement");
    mkdirSync(aggregateDirectory, { mode: 0o700 });
    const aggregateFile = path.join(aggregateDirectory, "frontend-measurement-aggregate.json");
    const aggregate = { schema_id: "cartulary.frontend_measurement_aggregate.v5", target_id: "browser-e2e-measurement", status: "qualified", profile_groups: [{
                fixture_profile_id: summary.fixture_profile_id, snapshot_key: summary.snapshot_key, builder_count: 1, clone_count: 1, scheduler_overlap_count: 0,
                build_artifact: summary.build_artifact, summary_artifacts: [summaryArtifact.ref], rollups: [{ predicate_id: summary.rollup.predicate_id, sample_count: summary.rollup.sample_count, statistics: summary.rollup.statistics, threshold_ms: summary.rollup.threshold_ms, outcome: summary.rollup.outcome }],
            }] };
    const targetEvidence = {
        schema_id: "cartulary.browser_target_result.v9", target_id: aggregate.target_id, status: "pass", generated_at: "2026-10-14T00:00:00Z",
        group_results: [{ group_id: "measurement", browser_session_id: "measurement", ref: "group.json", sha256: summaryArtifact.ref.sha256 }],
        sessions: [{ browser_session_id: "measurement", runtime_profile_id: "measurement", service_requirement: "test-services", artifacts: [
                    { kind: "stack_v7", ref: "stack.json", sha256: summaryArtifact.ref.sha256 },
                    { kind: "startup_diagnostics_v2", ref: "startup.json", sha256: summaryArtifact.ref.sha256 },
                ] }],
        artifacts: [{ kind: `frontend_measurement_aggregate_${aggregate.schema_id.split(".").at(-1)}`, ref: "frontend-measurement-aggregate.json", sha256: summaryArtifact.ref.sha256 }],
    };
    validateSchemaSync(targetEvidence.schema_id, targetEvidence);
    assert.throws(() => validateSchemaSync(targetEvidence.schema_id, { ...targetEvidence, schema_id: "cartulary.browser_target_result.v4" }), /validation failed/u);
    assert.throws(() => validateSchemaSync(targetEvidence.schema_id, { ...targetEvidence, artifacts: [{ ...targetEvidence.artifacts[0], kind: "frontend_measurement_aggregate_v3" }] }), /validation failed/u);
    const inspect = (resultsDir = path.dirname(measurementEvidenceRoot), runID = path.basename(measurementEvidenceRoot)) => {
      const exactRunSource = `
      import fs from "node:fs";
      import { syncBuiltinESMExports } from "node:module";
      const readdir = fs.readdirSync;
      fs.readdirSync = (directory, ...args) => {
        if (String(directory) === ${JSON.stringify(path.dirname(measurementEvidenceRoot))}) {
          process.stderr.write("unrequested sibling run discovery");
          process.exit(97);
        }
        return readdir(directory, ...args);
      };
      syncBuiltinESMExports();
      process.argv = ${JSON.stringify([process.execPath, "tools/harness/diagnostics/explain-run-cli.mjs", "--results-dir", resultsDir, "--run-id", runID, "--detail", "measurement"])};
      await import(${JSON.stringify(new URL("../../diagnostics/explain-run-cli.mjs", import.meta.url).href)});
    `;
      return spawnSync(process.execPath, ["--input-type=module", "-e", exactRunSource], { cwd: root, encoding: "utf8", timeout: 30000 });
    };
    writeFileSync(aggregateFile, JSON.stringify(aggregate));
    const current = inspect();
    assert.equal(current.status, 0, current.stderr);
    assert.match(current.stdout, /generation=current/u);
    assert.match(current.stdout, /MEASUREMENT-STAGE/u);
    assert.doesNotMatch(current.stdout, /MEASUREMENT-COMPARE/u);
    const exactDirectory = inspect(measurementEvidenceRoot);
    assert.equal(exactDirectory.status, 0, exactDirectory.stderr);
    assert.equal(exactDirectory.stdout, current.stdout);
    const invalidID = inspect(path.dirname(measurementEvidenceRoot), "../other-run");
    assert.equal(invalidID.status, 2);
    assert.match(invalidID.stderr, /^invalid RUN_ID/u);
    const summaryFile = path.join(measurementEvidenceRoot, summaryArtifact.ref.path);
    linkSync(summaryFile, path.join(measurementEvidenceRoot, "hardlink.json"));
    const hardlink = inspect();
    assert.equal(hardlink.status, 11, "diagnostics must reject hard-linked evidence");
    rmSync(path.join(measurementEvidenceRoot, "hardlink.json"));
    symlinkSync(measurementEvidenceRoot, path.join(measurementEvidenceRoot, "alias"));
    for (const relative of [`./${summaryArtifact.ref.path}`, `alias/${summaryArtifact.ref.path}`]) {
        const invalid = structuredClone(aggregate);
        invalid.profile_groups[0].summary_artifacts[0].path = relative;
        writeFileSync(aggregateFile, JSON.stringify(invalid));
        assert.equal(inspect().status, 11, "diagnostics must reject noncanonical and aliased references");
    }
    rmSync(path.join(measurementEvidenceRoot, "alias"));
    const summaryBytes = readFileSync(summaryFile);
    const duplicateSummary = Buffer.from(`{"schema_id":${JSON.stringify(summary.schema_id)},${summaryBytes.toString("utf8").slice(1)}`);
    const ambiguous = structuredClone(aggregate);
    ambiguous.profile_groups[0].summary_artifacts[0].sha256 = artifactDigest(duplicateSummary);
    writeFileSync(summaryFile, duplicateSummary);
    writeFileSync(aggregateFile, JSON.stringify(ambiguous));
    assert.equal(inspect().status, 11, "digest-valid duplicate members cannot become diagnostic evidence");
    assert.deepEqual(readFileSync(summaryFile), duplicateSummary);
    writeFileSync(summaryFile, summaryBytes);
    const duplicate = `{"schema_id":${JSON.stringify(aggregate.schema_id)},${JSON.stringify(aggregate).slice(1)}`;
    writeFileSync(aggregateFile, duplicate);
    assert.equal(inspect().status, 11, "diagnostics must reject duplicate JSON members");
    assert.equal(readFileSync(aggregateFile, "utf8"), duplicate);
    const forbidden = structuredClone(aggregate);
    // A hostile path token under a synthetic run; the child guard forbids I/O.
    const forbiddenDirectory = "docs";
    forbidden.profile_groups[0].summary_artifacts[0].path = `${forbiddenDirectory}/evidence.json`;
    writeFileSync(aggregateFile, JSON.stringify(forbidden));
    const guardedSource = `
      import fs from "node:fs";
      import { syncBuiltinESMExports } from "node:module";
      for (const name of ["existsSync", "lstatSync", "statSync", "readFileSync", "openSync", "realpathSync"]) {
        const original = fs[name];
        fs[name] = (file, ...args) => {
          if (String(file).split("/").includes(${JSON.stringify(forbiddenDirectory)})) {
            process.stderr.write("forbidden filesystem probe");
            process.exit(97);
          }
          return original(file, ...args);
        };
      }
      syncBuiltinESMExports();
      process.argv = ${JSON.stringify([process.execPath, "tools/harness/diagnostics/explain-run-cli.mjs", "--results-dir", path.dirname(measurementEvidenceRoot), "--run-id", path.basename(measurementEvidenceRoot), "--detail", "measurement"])};
      await import(${JSON.stringify(new URL("../../diagnostics/explain-run-cli.mjs", import.meta.url).href)});
    `;
    const guarded = spawnSync(process.execPath, ["--input-type=module", "-e", guardedSource], { cwd: root, encoding: "utf8", timeout: 30000 });
    assert.equal(guarded.status, 11);
    assert.equal(guarded.stdout, "");
    assert.equal(guarded.stderr.includes("forbidden filesystem probe"), false);
    assert.match(guarded.stderr, /^invalid measurement artifact/u);
    for (const schema_id of ["cartulary.frontend_measurement_aggregate.v1", "cartulary.frontend_measurement_aggregate.v2", "unsupported-private-value"]) {
        const bytes = JSON.stringify({ schema_id, summaries: [] });
        writeFileSync(aggregateFile, bytes);
        const rejected = inspect();
        assert.equal(rejected.status, 11, rejected.stderr);
        assert.equal(rejected.stdout, "");
        assert.match(rejected.stderr, /^unsupported measurement artifact;/u);
        assert.ok(rejected.stderr.length < 160);
        assert.equal(rejected.stderr.includes(schema_id), false);
        assert.equal(readFileSync(aggregateFile, "utf8"), bytes, "inspection changed retired bytes");
    }
    rmSync(aggregateDirectory, { recursive: true });
});
