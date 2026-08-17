import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  helpTiers,
  readJSON,
  targetEntryMap,
} from "../generated-artifacts/task-surface/model.mjs";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";
import { loadTestCatalog } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
export const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const aggregateTargets = new Set(["test-fast", "test", "check", "ci", "release-check"]);

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

function loadTaskSurface(root = repoRoot) {
  const manifestFile =
    process.env.CARTULARY_TASK_SURFACE_MANIFEST ??
    path.join(root, "tools", "task_surface_manifest.json");
  const manifest = readJSON(manifestFile);
  return { manifest, targets: targetEntryMap(manifest) };
}

function helpTierByTarget(manifest) {
  const tiers = new Map();
  for (const tier of helpTiers(manifest)) {
    for (const entry of tier.entries ?? []) tiers.set(entry.target, tier.name);
  }
  return tiers;
}

function selectedRowIDs(graph) {
  return [...new Set(graph.units.flatMap((unit) =>
    unit.current_run_evidence_outputs
      .filter((output) => output.startsWith("rows/") && output.endsWith(".json"))
      .map((output) => output.slice("rows/".length, -".json".length)),
  ))].sort(compareStrings);
}

function artifactProjection(target, root) {
  const relative = [
    "run-manifest.json",
    "unit-events.ndjson",
    "run-summary.json",
    `target-summaries/${target}.json`,
  ];
  return {
    latest: null,
    candidates: [],
    expected: relative.map((file) =>
      path.relative(root, path.join(root, ".cartulary", "test-results", "<run-id>", file)),
    ),
  };
}

function stepCoverage(rows) {
  const counts = (field) => Object.fromEntries(
    [...new Set(rows.map((row) => row[field]))]
      .sort(compareStrings)
      .map((value) => [value, rows.filter((row) => row[field] === value).length]),
  );
  return {
    total: rows.length,
    owners: [...new Set(rows.map((row) => row.owner_id))].sort(compareStrings),
    runners: counts("runner"),
    evidence_classes: counts("evidence_class"),
    minimum_tiers: counts("minimum_tier"),
  };
}

function serviceRequirements(graph) {
  const fixtureCapabilities = new Set(
    graph.units.map((unit) => unit.fixture_lease).filter((value) => value !== "none"),
  );
  const resources = new Set(
    graph.units.flatMap((unit) => Object.keys(unit.resource_claims))
      .filter((resource) => ["postgres", "object_store", "service_stack", "browser_stack"].includes(resource)),
  );
  return [...new Set([...fixtureCapabilities, ...resources])].sort(compareStrings);
}

export function targetGuidance(
  target,
  { root = repoRoot, includeExecutionMap = true, includeArtifacts = true } = {},
) {
  const { manifest, targets } = loadTaskSurface(root);
  const entry = targets.get(target);
  if (!entry) return null;

  const compiler = new WorkGraphCompiler(root);
  const compiled = aggregateTargets.has(target)
    ? compiler.compileAggregatePlan(target)
    : { graph: compiler.compile({ kind: "target", target }), projections: { [target]: [] } };
  const catalog = loadTestCatalog(root);
  const rowIDs = selectedRowIDs(compiled.graph);
  const rows = rowIDs.map((rowID) => catalog.rowByID.get(rowID));
  const artifact = includeArtifacts ? artifactProjection(target, root) : {
    latest: null,
    candidates: [],
    expected: [],
  };
  const workUnits = compiled.graph.units.map((unit) => ({
    label: unit.unit_id,
    kind: unit.kind,
    owner_id: unit.owner_id,
    needs: unit.needs,
    fixture_capability: unit.fixture_lease,
    cache_policy: unit.cache_policy,
    semantic_digest: unit.semantic_digest,
  }));
  const executionMap = includeExecutionMap ? {
    graph_digest: compiled.graph.graph_digest,
    projections: compiled.projections,
    work_unit_summary: workUnits,
    artifacts: artifact,
  } : null;

  return {
    schema_id: "cartulary.target_guidance.v3",
    target,
    target_class: entry.target_class,
    default_inclusion_sets: entry.default_inclusion_sets ?? [],
    check_projection: entry.check_projection ?? null,
    input_contract: entry.input_contract ?? null,
    help_tier: helpTierByTarget(manifest).get(target) ?? null,
    backing_scripts: entry.backing_scripts ?? [],
    graph_digest: compiled.graph.graph_digest,
    execution_map: executionMap,
    execution_summary: `units=${compiled.graph.units.length} rows=${rows.length}`,
    service_requirements: serviceRequirements(compiled.graph),
    artifact,
    step_coverage: stepCoverage(rows),
    rows,
    go_rows: [],
  };
}

export function allTargetNames({ root = repoRoot } = {}) {
  return [...loadTaskSurface(root).targets.keys()].sort(compareStrings);
}

export function formatRequirements(values) {
  return values.length === 0 ? "none" : values.join(",");
}

export function formatStepCoverage(summary) {
  if (!summary || summary.total === 0) return "none";
  const compact = (value) => Object.entries(value)
    .map(([key, count]) => `${key}:${count}`)
    .join(",");
  return [
    `owners=${summary.owners.join(",") || "none"}`,
    `runners=${compact(summary.runners) || "none"}`,
    `evidence=${compact(summary.evidence_classes) || "none"}`,
    `tiers=${compact(summary.minimum_tiers) || "none"}`,
  ].join(" ");
}
