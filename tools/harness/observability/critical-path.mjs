import { actualCriticalPath, intervalUnion } from "../evidence-accounting/index.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { loadRetainedObservability } from "./observability.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

const usage = (message) => Object.assign(new Error(message), { exit_code: 2 });
export function criticalPathQuery({ view = "units", unitID = null, offset = 0, limit = 20 } = {}) {
  if (!["units", "waits", "allocations"].includes(view)) throw usage("unknown critical-path VIEW");
  const integer = (value) => /^(0|[1-9][0-9]*)$/u.test(String(value)) && Number.isSafeInteger(Number(value));
  if (!integer(offset) || !integer(limit) || Number(limit) < 1 || Number(limit) > 100) throw usage("invalid critical-path OFFSET or LIMIT");
  if (unitID !== null && (typeof unitID !== "string" || !/^[A-Za-z0-9_.:/-]{1,255}$/u.test(unitID))) throw usage("invalid UNIT_ID");
  return { view, unit_id: unitID, offset: Number(offset), limit: Number(limit) };
}
const totals = (run, ids) => ({ unit_count: ids.length,
  queue_ms: ids.reduce((sum, id) => sum + (run.intervals.get(id)?.queue_ms ?? 0), 0),
  lifetime_ms: ids.reduce((sum, id) => { const i = run.intervals.get(id); return sum + (i ? i.end - i.start : 0); }, 0) });

export function criticalPathModel(run, execution, target, query) {
  const summary = run.targetSummaries.get(target || run.manifest.target);
  if (!summary) throw usage("TARGET is absent from retained target summaries");
  const ids = actualCriticalPath({ units: [...run.registrations.values()] }, run.intervals, summary.unit_ids);
  if (query.unit_id !== null && !ids.includes(query.unit_id)) throw usage("UNIT_ID is absent from selected critical path");
  for (const id of ids) {
    if (Object.values(run.waits.get(id).by_reason).reduce((sum, value) => sum + value, 0) !== run.intervals.get(id).queue_ms) throw new Error("canonical queue reason partition does not close");
  }
  const selected = ids.filter((id) => query.unit_id === null || id === query.unit_id);
  const positions = new Map(ids.map((id, index) => [id, index]));
  const lifecycle = execution?.allocation_lifecycle;
  let records;
  if (query.view === "units") records = selected.map((unit_id) => {
    const i = run.intervals.get(unit_id);
    const phases = [...run.phases.values()].filter((p) => p.unit_id === unit_id)
      .sort((a, b) => a.start_ms - b.start_ms || a.end_ms - b.end_ms || a.phase.localeCompare(b.phase))
      .map((p) => ({ ...p, duration_ms: p.end_ms - p.start_ms }));
    const covered = intervalUnion(phases.map((p) => ({ start: Math.max(i.start, p.start_ms), end: Math.min(i.end, p.end_ms) })).filter((p) => p.end >= p.start));
    return { unit_id, kind: run.registrations.get(unit_id).kind, path_index: positions.get(unit_id), queue_ms: i.queue_ms,
      queue_by_reason_ms: run.waits.get(unit_id).by_reason, lifetime_ms: i.end - i.start, start_ms: i.start, end_ms: i.end,
      phases: phases.map(({ unit_id: _, ...p }) => p), phase_union_ms: covered,
      phase_overlap_ms: phases.reduce((sum, p) => sum + p.duration_ms, 0) - intervalUnion(phases.map(p => ({start:p.start_ms,end:p.end_ms}))),
      phase_outside_lifetime_ms: intervalUnion(phases.map(p => ({start:p.start_ms,end:p.end_ms}))) - covered, uncovered_lifetime_ms: i.end - i.start - covered };
  });
  else if (query.view === "allocations") records = (lifecycle?.records ?? []).flatMap((record) => {
    const pathUnits = ids.filter((id) => record.borrows.some((borrow) => borrow.unit_id === id) || record.triggering_unit_id === id);
    if (!pathUnits.some((id) => selected.includes(id))) return [];
    return [{ path_index: positions.get(pathUnits[0]), path_unit_ids: pathUnits, lifecycle: record }];
  }).sort((a, b) => a.path_index - b.path_index || a.lifecycle.allocation_ref.localeCompare(b.lifecycle.allocation_ref));
  else records = [];
  const total = query.view === "waits" ? selected.reduce((sum, id) => sum + (run.waits.get(id)?.count ?? 0), 0) : records.length;
  return { ids, selected, positions, model: {
    schema_id: "cartulary.harness_critical_path.v1", run_id: run.manifest.run_id, source_digest: run.manifest.source_digest,
    graph_digest: run.manifest.graph_digest, target: summary.target, query,
    coverage: { canonical: "complete", allocation_lifecycle: lifecycle?.status ?? "unavailable", allocation_omissions: lifecycle?.omitted_records ?? null },
    path_totals: totals(run, ids), filtered_totals: totals(run, selected), total_count: total,
    returned_count: 0, next_offset: null, records: records.slice(query.offset, query.offset + query.limit),
    limitations: ["graph_queue_excludes_host_admission", "phases_and_blockers_may_overlap", "claims_are_not_utilization", "allocation_clock_is_independent"] } };
}

export async function criticalPathExplanation(runDir, options = {}) {
  const query = criticalPathQuery(options);
  const { run, execution } = await loadRetainedObservability(runDir);
  const { model, selected, positions } = criticalPathModel(run, execution, options.target, query);
  if (query.view === "waits" && query.offset < model.total_count) {
    // One bounded window per selected unit, interpreted by the canonical reducer.
    // Prefix counts translate global path ordering without retaining every wait.
    let prefix = 0;
    const windows = new Map();
    for (const id of selected) { windows.set(id, { prefix, seen: 0 }); prefix += run.waits.get(id)?.count ?? 0; }
    const again = await validateCanonicalRun(runDir, "", { onWaitSegment: (wait) => {
      const window = windows.get(wait.unit_id);
      if (!window) return;
      const index = window.prefix + window.seen++;
      if (index < query.offset || index >= query.offset + query.limit) return;
      model.records.push({ unit_id: wait.unit_id, path_index: positions.get(wait.unit_id), start_ms: wait.start, end_ms: wait.end,
        duration_ms: wait.end - wait.start, wait_reason: wait.wait_reason, blocking_resources: wait.blocking_resources, blocking_unit_ids: wait.blocking_unit_ids });
    } });
    if (again.eventDigest !== run.eventDigest || JSON.stringify(again.summary) !== JSON.stringify(run.summary) || JSON.stringify(again.manifest) !== JSON.stringify(run.manifest) || [...windows].some(([id, w]) => w.seen !== run.waits.get(id).count)) throw new Error("canonical evidence changed during critical-path query");
    model.records.sort((a, b) => a.path_index - b.path_index || a.start_ms - b.start_ms || a.end_ms - b.end_ms);
  }
  return model;
}

function page(model, count) {
  const records = model.records.slice(0, count);
  return { ...model, records, returned_count: records.length,
    next_offset: model.query.offset + records.length < model.total_count ? model.query.offset + records.length : null };
}
const human = (value) => [
  `[CRITICAL-PATH] run=${value.run_id} target=${value.target} source=${value.source_digest} graph=${value.graph_digest}`,
  `[QUERY] ${JSON.stringify(value.query)}`,
  `[COVERAGE] ${JSON.stringify(value.coverage)}`,
  `[PATH-TOTALS] ${JSON.stringify(value.path_totals)}`,
  `[FILTERED-TOTALS] ${JSON.stringify(value.filtered_totals)}`,
  ...value.records.map((record) => JSON.stringify(record)),
  `[PAGE] total=${value.total_count} returned=${value.returned_count} next_offset=${value.next_offset ?? "null"}`,
  `[LIMITATIONS] ${value.limitations.join(",")}`,
].join("\n") + "\n";
export function formatCriticalPath(model, { json = false, maximumBytes = json ? 262144 : 8000 } = {}) {
  for (let count = model.records.length; count >= 0; count--) {
    if (!count && model.records.length) throw new Error("one critical-path record exceeds output bound");
    const value = page(model, count);
    validateSchemaSync(value.schema_id, value);
    const output = json ? `${JSON.stringify(value)}\n` : human(value);
    if (Buffer.byteLength(output) <= maximumBytes && (json || output.split("\n").length - 1 <= 120)) return output;
  }
  throw new Error("critical-path metadata exceeds output bound");
}
