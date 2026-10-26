import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";
import test from "node:test";
import { measurementContext, writeMeasurementArtifact } from "./measurement-fixtures.mjs";
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
    const inspect = () => spawnSync(process.execPath, ["tools/harness/diagnostics/explain-run-cli.mjs", "--results-dir", path.dirname(measurementEvidenceRoot), "--run-id", path.basename(measurementEvidenceRoot), "--detail", "measurement"], { cwd: root, encoding: "utf8", timeout: 30000 });
    writeFileSync(aggregateFile, JSON.stringify(aggregate));
    const current = inspect();
    assert.equal(current.status, 0, current.stderr);
    assert.match(current.stdout, /generation=current/u);
    assert.match(current.stdout, /MEASUREMENT-STAGE/u);
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
