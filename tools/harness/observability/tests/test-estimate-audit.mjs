import assert from "node:assert/strict";
import test from "node:test";

import { validateSchemaSync } from "../../contract/index.mjs";
import { auditEstimates, formatEstimateAudit, maximumEstimateFindings } from "../estimate-audit.mjs";

// TH-HARNESS-REQ-866: the advisory estimate audit qualifies one exact retained
// run, then reports bounded findings for its Go units. Synthetic runs enter
// through the loader port, so no retained root or current checkout is needed.
const owner = {
  go_cohort_packing_bound_ms: 300000,
  invocation_policies: { go: { startup_ms: 5000, cleanup_ms: 250 } },
  estimate_audit: { drift_ratio: 2, heavy_lane_threshold_ms: 120000 },
  evidence_estimates_ms: { unit: 1000, integration: 5000 },
};
const sourceDigest = "sha256:" + "a".repeat(64);
const goRow = (rowID, extra = {}) => ({ row_id: rowID, runner: "go", evidence_class: "integration", fixture_capability: "none", ...extra });

function syntheticRun({ status = "pass", source = sourceDigest, units }) {
  const registrations = new Map();
  const phases = new Map();
  let phaseID = 0;
  for (const unit of units) {
    registrations.set(unit.id, { unit_id: unit.id, kind: unit.kind ?? "runner", row_ids: unit.rows, resource_claims: unit.claims ?? { cpu: 1 } });
    for (const [phase, start, end] of unit.phases ?? []) {
      phaseID += 1;
      phases.set(`phase:${phaseID}`, { unit_id: unit.id, phase, start_ms: start, end_ms: end, outcome: "completed" });
    }
  }
  return {
    manifest: { run_id: "tcr-estimates-fixture", target: "check", source_digest: source },
    summary: { status, failure_class: status === "pass" ? null : "product" },
    registrations, phases,
  };
}

const audit = (run, catalogRows = [], durations = {}) => auditEstimates({
  runDir: "/retained/tcr-estimates-fixture", owner, catalogRows, sourceDigest, readRow: (_root, id) => {
    if (!Object.hasOwn(durations, id)) throw new Error("row timing unavailable");
    return { duration_ms: durations[id] };
  }, loadRun: async () => {
    if (run instanceof Error) throw run;
    return run;
  },
});

test("an unqualified retained run is rejected with sorted reasons and no findings", async () => {
  const failed = await audit(syntheticRun({ status: "fail", source: "sha256:" + "b".repeat(64), units: [] }));
  assert.equal(failed.qualification.status, "rejected");
  assert.deepEqual(failed.qualification.reasons, ["run_not_passed", "source_digest_mismatch"]);
  assert.deepEqual(failed.findings, []);
  validateSchemaSync(failed.schema_id, failed);
  const unreadable = await audit(new Error("canonical roster does not close"));
  assert.deepEqual(unreadable.qualification, { status: "rejected", reasons: ["canonical_evidence_invalid"] });
  assert.equal(unreadable.run_id, "tcr-estimates-fixture");
  assert.equal(unreadable.target, null);
  validateSchemaSync(unreadable.schema_id, unreadable);
  assert.match(formatEstimateAudit(unreadable), /\[ESTIMATES-REJECTED\] reasons=canonical_evidence_invalid/u);
});

test("a qualified run reports packing, drift and heavy-lane findings from runner phases only", async () => {
  const rows = [
    goRow("a.heavy", { estimated_work_ms: 100000, fixture_capability: "postgres_dedicated" }),
    goRow("b.overestimated", { estimated_work_ms: 100000 }),
    goRow("c.class_default", { evidence_class: "unit" }),
    goRow("d.heavy_lane", { estimated_work_ms: 100000, fixture_capability: "postgres_migration" }),
    { row_id: "e.vitest", runner: "vitest", evidence_class: "unit", fixture_capability: "none" },
    goRow("f.not_launched"),
  ];
  const report = await audit(syntheticRun({ units: [
    // Overlapping runner intervals are unioned; admission and fixture time are excluded.
    { id: "go:a", rows: ["a.heavy"], phases: [["host_admission", 0, 50000], ["fixture_acquire", 50000, 90000], ["runner", 90000, 290000], ["runner", 240000, 400000]] },
    { id: "go:b", rows: ["b.overestimated"], phases: [["runner", 0, 20000]] },
    { id: "go:c", rows: ["c.class_default"], phases: [["runner", 0, 300]] },
    { id: "go:d", rows: ["d.heavy_lane"], claims: { cpu: 1, postgres_heavy: 1 }, phases: [["runner", 0, 150000]] },
    { id: "vitest_cohort:e", rows: ["e.vitest"], phases: [["runner", 0, 900000]] },
    { id: "go:f", rows: ["f.not_launched"] },
  ] }), rows, { "a.heavy": 130000 });
  assert.equal(report.qualification.status, "accepted");
  assert.equal(report.unit_count, 4, "only launched Go row units are audited");
  assert.deepEqual(report.findings.map((finding) => [finding.unit_id, finding.code, finding.observed_ms, finding.planned_ms]), [
    ["go:a", "estimate_drift", 310000, 105250],
    ["go:a", "heavy_lane_candidate", 130000, 100000],
    ["go:a", "over_packing_bound", 310000, 105250],
    ["go:b", "estimate_drift", 20000, 105250],
  ]);
  assert.equal(report.findings_omitted, 0);
  validateSchemaSync(report.schema_id, report);
  const human = formatEstimateAudit(report);
  assert.match(human, /^\[ESTIMATES\] run_id=tcr-estimates-fixture target=check qualification=accepted units=4 findings=4\n/u);
  assert.match(human, /\[ESTIMATE\] code=heavy_lane_candidate unit=go:a observed_ms=130000 planned_ms=100000 rows=a\.heavy\n/u);
});

test("findings are sorted and bounded, and the human view states the remainder", async () => {
  const count = maximumEstimateFindings + 10;
  const rows = Array.from({ length: count }, (_, index) => goRow(`row.${String(index).padStart(3, "0")}`, { estimated_work_ms: 200000 }));
  const report = await audit(syntheticRun({ units: rows.map((row, index) => ({
    id: `go:${String(count - index).padStart(3, "0")}`, rows: [row.row_id], phases: [["runner", 0, 300001]],
  })) }), rows);
  assert.equal(report.findings.length, maximumEstimateFindings);
  assert.equal(report.findings_omitted, 10);
  const ids = report.findings.map((finding) => finding.unit_id);
  assert.deepEqual(ids, [...ids].sort());
  validateSchemaSync(report.schema_id, report);
  const lines = formatEstimateAudit(report).trimEnd().split("\n");
  assert.equal(lines.filter((line) => line.startsWith("[ESTIMATE] ")).length, 20);
  assert.equal(lines.at(-1), `[ESTIMATES-OMITTED] findings=${count - 20} (JSON=1 lists up to ${maximumEstimateFindings})`);
});

test("a unit's row list is bounded with an omission count", async () => {
  const rows = Array.from({ length: 25 }, (_, index) => goRow(`row.${String(index).padStart(2, "0")}`, { estimated_work_ms: 1000 }));
  const report = await audit(syntheticRun({ units: [{ id: "go:wide", rows: rows.map((row) => row.row_id), phases: [["runner", 0, 300001]] }] }), rows);
  const [finding] = report.findings;
  assert.equal(finding.code, "estimate_drift");
  assert.equal(finding.row_ids.length, 20);
  assert.equal(finding.row_ids_omitted, 5);
  validateSchemaSync(report.schema_id, report);
});

test("a slow cohort cannot substitute its duration for eligible PostgreSQL rows", async () => {
  const rows = ["first", "second", "missing", "plain"].map((id) =>
    goRow(id, { fixture_capability: id === "plain" ? "none" : "postgres_dedicated" }));
  const run = syntheticRun({ units: [{ id: "go:group", rows: rows.map((row) => row.row_id), phases: [["runner", 0, 240000]] }] });
  const report = await audit(run, rows, { first: 120000, second: 60000, plain: 180000 });
  assert.deepEqual(report.findings.filter((finding) => finding.code !== "estimate_drift"), [{
    unit_id: "go:group", code: "row_timing_unavailable", observed_ms: null, planned_ms: 5000, row_ids: ["missing"], row_ids_omitted: 0,
  }]);
  validateSchemaSync(report.schema_id, report);
});
