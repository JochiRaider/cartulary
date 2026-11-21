import path from "node:path";

import { intervalUnion } from "../evidence-accounting/index.mjs";
import { rowEstimateMs } from "../scheduler/work-graph/cohorts.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

// TH-HARNESS-REQ-866: an advisory audit of one exact retained invocation's Go
// units against the current authored planning inputs. It never gates, never
// writes, and reports an unqualified run as rejected rather than auditing it.
export const estimateAuditSchemaID = "cartulary.harness_estimate_audit.v1";
export const maximumEstimateFindings = 50;
export const maximumDisplayedEstimateFindings = 20;
const maximumFindingRows = 20;
const heavyFixtureCapabilities = new Set(["postgres_dedicated", "postgres_migration"]);
const compareASCII = (left, right) => (left < right ? -1 : left > right ? 1 : 0);

export function estimateAuditPolicy(owner) {
  return {
    go_cohort_packing_bound_ms: owner.go_cohort_packing_bound_ms,
    go_startup_ms: owner.invocation_policies.go.startup_ms,
    go_cleanup_ms: owner.invocation_policies.go.cleanup_ms,
    drift_ratio: owner.estimate_audit.drift_ratio,
    heavy_lane_threshold_ms: owner.estimate_audit.heavy_lane_threshold_ms,
  };
}

// The loader is an explicit port so adversarial and synthetic retained runs
// can be audited without a current checkout producing them.
export async function auditEstimates({ runDir, owner, catalogRows, sourceDigest, loadRun = validateCanonicalRun }) {
  const policy = estimateAuditPolicy(owner);
  const reasons = [];
  let run = null;
  try {
    run = await loadRun(runDir);
  } catch {
    reasons.push("canonical_evidence_invalid");
  }
  if (run) {
    if (run.summary.status !== "pass" || run.summary.failure_class !== null) reasons.push("run_not_passed");
    if (run.manifest.source_digest !== sourceDigest) reasons.push("source_digest_mismatch");
  }
  const report = {
    schema_id: estimateAuditSchemaID,
    run_id: run?.manifest.run_id ?? path.basename(runDir),
    target: run?.manifest.target ?? null,
    qualification: { status: reasons.length ? "rejected" : "accepted", reasons: reasons.sort(compareASCII) },
    policy,
    unit_count: 0,
    findings: [],
    findings_omitted: 0,
  };
  if (reasons.length) return report;

  const rows = new Map(catalogRows.map((row) => [row.row_id, row]));
  const runnerIntervals = new Map();
  for (const phase of run.phases.values()) {
    if (phase.phase !== "runner" || phase.end_ms === null) continue;
    const intervals = runnerIntervals.get(phase.unit_id) ?? [];
    intervals.push({ start: phase.start_ms, end: phase.end_ms });
    runnerIntervals.set(phase.unit_id, intervals);
  }
  const findings = [];
  for (const [unitID, registration] of [...run.registrations].sort(([left], [right]) => compareASCII(left, right))) {
    const rowIDs = [...(registration.row_ids ?? [])].sort(compareASCII);
    const unitRows = rowIDs.map((id) => rows.get(id));
    // Only launched Go row units are planned by cohort packing.
    if (registration.kind !== "runner" || rowIDs.length === 0 || unitRows.some((row) => row?.runner !== "go")) continue;
    const intervals = runnerIntervals.get(unitID);
    if (!intervals) continue;
    report.unit_count += 1;
    const observed = intervalUnion(intervals);
    const planned = policy.go_startup_ms + policy.go_cleanup_ms +
      unitRows.reduce((total, row) => total + rowEstimateMs(owner, row), 0);
    const finding = (code) => findings.push({
      unit_id: unitID, code, observed_ms: observed, planned_ms: planned,
      row_ids: rowIDs.slice(0, maximumFindingRows), row_ids_omitted: Math.max(0, rowIDs.length - maximumFindingRows),
    });
    if (observed > policy.go_cohort_packing_bound_ms) finding("over_packing_bound");
    const reviewed = unitRows.some((row) => row.estimated_work_ms !== undefined);
    if (observed > planned * policy.drift_ratio || (reviewed && observed * policy.drift_ratio < planned)) finding("estimate_drift");
    if (observed > policy.heavy_lane_threshold_ms &&
        unitRows.some((row) => heavyFixtureCapabilities.has(row.fixture_capability)) &&
        !((registration.resource_claims?.postgres_heavy ?? 0) > 0)) finding("heavy_lane_candidate");
  }
  findings.sort((left, right) => compareASCII(left.unit_id, right.unit_id) || compareASCII(left.code, right.code));
  report.findings = findings.slice(0, maximumEstimateFindings);
  report.findings_omitted = Math.max(0, findings.length - maximumEstimateFindings);
  return report;
}

export function formatEstimateAudit(report) {
  const lines = [`[ESTIMATES] run_id=${report.run_id} target=${report.target ?? "none"} qualification=${report.qualification.status} units=${report.unit_count} findings=${report.findings.length + report.findings_omitted}`];
  if (report.qualification.status === "rejected") lines.push(`[ESTIMATES-REJECTED] reasons=${report.qualification.reasons.join(",")}`);
  for (const finding of report.findings.slice(0, maximumDisplayedEstimateFindings)) {
    const shown = finding.row_ids.slice(0, 2).join(",");
    const more = finding.row_ids.length - 2 + finding.row_ids_omitted;
    lines.push(`[ESTIMATE] code=${finding.code} unit=${finding.unit_id} observed_ms=${finding.observed_ms} planned_ms=${finding.planned_ms} rows=${shown}${more > 0 ? ` (+${more})` : ""}`);
  }
  const hidden = report.findings.length + report.findings_omitted - Math.min(report.findings.length, maximumDisplayedEstimateFindings);
  if (hidden > 0) lines.push(`[ESTIMATES-OMITTED] findings=${hidden} (JSON=1 lists up to ${maximumEstimateFindings})`);
  return `${lines.join("\n")}\n`;
}
