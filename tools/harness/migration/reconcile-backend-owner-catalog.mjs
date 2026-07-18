#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { loadTestCatalog } from "../test-catalog/test-catalog.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(root, relativePath), "utf8"));
}

function fail(message) {
  throw new Error(`backend reconciliation failed: ${message}`);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function exactValues(row, plural, singular) {
  return [...new Set([...(row[plural] ?? []), ...(row[singular] ? [row[singular]] : [])])].sort(asciiCompare);
}

function packageOwner(packagePath) {
  if (packagePath === "./internal/app/serverprocess") return "app.server";
  const match = /^\.\/internal\/(app|modules|platform)\/([^/]+)/u.exec(packagePath);
  if (!match) fail(`unclassified package ${packagePath}`);
  return `${match[1] === "modules" ? "module" : match[1]}.${match[2]}`;
}

function nonGoOwner(phase, row) {
  if (phase === 1) return ["E-1-09", "E-1-10"].includes(row.id) ? "module.incidents" : "module.auth";
  if (phase === 2) return "module.incidents";
  if (phase === 3) return "module.timeline";
  if (phase === 4) {
    if (["U-4-WB-01", "U-4-WB-02", "E-4-05"].includes(row.id)) return "module.assessments";
    if (["U-4-WB-03", "V-4-GRID-02"].includes(row.id)) return "module.evidence";
    if (["U-4-WB-04", "U-4-WB-05", "E-4-06", "V-4-GRID-03"].includes(row.id)) return "module.workbook";
    return "module.entities";
  }
  if (phase === 5) return "module.evidence";
  if (phase === 6) return "module.collaboration";
  if (phase === 7) return "module.revisions";
  if (phase === 8) {
    if (["U-8-GRID-01", "E-8-03", "E-8-04"].includes(row.id)) return "platform.viewquery";
    if (["E-8-02", "E-8-SUPPORT-01"].includes(row.id)) return "module.workbook";
    return "module.savedviews";
  }
  if (phase === 9) {
    if (row.id === "E-9-03") return "module.links";
    if (row.id === "E-9-04") return "module.parties";
    if (row.id === "E-9-05") return "module.assessments";
    if (row.id === "E-9-TASKDECISION-06") return "module.tasksdecisions";
    return "module.workbook";
  }
  if (phase === 10) return "module.recovery";
  if (phase === 11) {
    return row.id.includes("REFERENCE-PACK") || row.file.includes("reference-pack")
      ? "module.reference_data"
      : "module.auth";
  }
  if (phase === 12) return "module.networkflow";
  fail(`unclassified non-Go row ${phase}:${row.id}`);
}

function expectedOwner(phase, row) {
  return (row.runner ?? "go_test") === "go_test" ? packageOwner(row.package) : nonGoOwner(phase, row);
}

function expectedFamily(row, support = false) {
  if (support) return row.layer === "backend_integration_support" ? "support_integration" : "support_unit";
  const family = {
    backend_integration: "integration",
    backend_process: "process",
    backend_store: "store",
    backend_unit: "unit",
    browser_functional: "browser",
    browser_measurement: "measurement",
    browser_stateful: "browser_stateful",
    browser_support: "browser_support",
    browser_visual: "visual",
    frontend_unit: "frontend",
  }[row.execution_dependency];
  if (!family) fail(`unknown execution dependency ${row.execution_dependency}`);
  return family;
}

function expectedFixture(row, support = false) {
  if (!support && row.runner === "playwright") return "service_stack";
  return {
    group_clone: "postgres_group_clone",
    migration_scratch: "postgres_migration_scratch",
    template_clone: "postgres_template_clone",
    transaction: "postgres_transaction",
  }[row.fixture_policy?.postgres] ?? "none";
}

function expectedRuntime(row, fixtureID, ownerID) {
  if (row.runner === "playwright") {
    if (ownerID === "module.networkflow" && row.id !== "E-12-NFAC001-01") return "network_flow_claimed";
    return row.runtime_profile_id ?? row.runtime_profile ?? "default";
  }
  return row.execution_dependency === "backend_process" || fixtureID !== "none" ? "default" : "none";
}

function expectedResource(row, fixtureID) {
  if (row.runner === "playwright") return "browser_exclusive";
  if (row.runner === "vitest") return "none";
  if (fixtureID === "postgres_transaction") return "go_transaction_heavy";
  if (["postgres_group_clone", "postgres_template_clone"].includes(fixtureID)) return "go_clone_heavy";
  if (fixtureID === "postgres_migration_scratch") return "go_io_heavy";
  return ["backend_process", "backend_integration"].includes(row.execution_dependency)
    ? "go_io_heavy"
    : "go_balanced";
}

function expectedEvidence(row, support = false) {
  if (support) return row.layer === "backend_integration_support" ? "integration" : "unit";
  if (row.execution_dependency === "browser_measurement") return "measurement";
  if (row.execution_dependency === "browser_visual") return "visual";
  if (row.runner === "playwright") return "browser";
  if (row.runner === "vitest") return "unit";
  return row.execution_dependency === "backend_unit" ? "unit" : "integration";
}

function expectedWebOwner(file) {
  if (file.includes("/networkFlow/") || file.includes("network-flow")) return "web.networkflow";
  if (file.includes("collaboration") || file.includes("session-recovery")) return "web.collaboration";
  if (file.includes("/workbook/") || file.includes("Workbook") || file.includes("workbook.")) return "web.workbook";
  return "web.application";
}

function expectedCollaborators(row, ownerID) {
  if ((row.runner ?? "go_test") === "go_test") return [];
  const collaborators = new Set([expectedWebOwner(row.file)]);
  if (row.runner === "playwright") collaborators.add("harness.browser");
  if (row.execution_dependency === "browser_visual") {
    collaborators.add("harness.visual");
    collaborators.add("web.design");
  }
  collaborators.delete(ownerID);
  return [...collaborators].sort(asciiCompare);
}

function compare(actual, expected, label) {
  if (JSON.stringify(actual) !== JSON.stringify(expected)) fail(`${label} differs`);
}

function increment(counts, key, amount = 1) {
  counts[key] = (counts[key] ?? 0) + amount;
}

function sortedCounts(counts) {
  return Object.fromEntries(Object.entries(counts).sort(([left], [right]) => asciiCompare(left, right)));
}

const catalog = loadTestCatalog(root);
const baseline = readJSON("tools/test_migration_baseline.json");
const crosswalk = readJSON("tools/test_migration_crosswalk.json");
const backendDispositions = crosswalk.dispositions.filter((entry) => entry.source_registry_id === "backend_phase_maps");
const dispositionByID = new Map(backendDispositions.map((entry) => [entry.legacy_row_id, entry]));
if (backendDispositions.length !== 456 || dispositionByID.size !== 456) fail("expected 456 unique backend dispositions");
if (backendDispositions.some((entry) => entry.disposition !== "migrated")) fail("every backend disposition must be migrated");
if (crosswalk.pending_baseline_keys.some((entry) => entry.source_registry_id === "backend_phase_maps")) {
  fail("backend pending keys remain");
}

const baselineBackendByID = new Map(
  baseline.identities
    .filter((entry) => entry.source_registry_id === "backend_phase_maps")
    .map((entry) => [entry.legacy_row_id, entry]),
);
if (baselineBackendByID.size !== 456) fail("frozen backend population differs from 456");

const supportCandidates = baseline.auxiliary_candidates.backend_support;
const supportDispositionByID = new Map(
  crosswalk.auxiliary_dispositions
    .filter((entry) => entry.candidate_id.startsWith("phase"))
    .map((entry) => [entry.candidate_id, entry]),
);
if (supportCandidates.length !== 37 || supportDispositionByID.size !== 37) fail("support population differs from 37");

const supportAuthorizationByID = new Map();
for (const candidate of supportCandidates) {
  const authorizations = crosswalk.new_rows.filter((entry) => entry.provenance.startsWith(`${candidate.candidate_id} `));
  if (authorizations.length !== 1) fail(`${candidate.candidate_id} must authorize exactly one row`);
  supportAuthorizationByID.set(candidate.candidate_id, authorizations[0]);
}

const ownerCounts = {};
const phaseCounts = {};
const runnerCounts = {};
const evidenceCounts = {};
const fixtureCounts = {};
const runtimeCounts = {};
const resourceCounts = {};
const defaultCheckCounts = {};
const claimPostureCounts = {};
let authoritativeRows = 0;
let authoritativeSelectorAtoms = 0;
let supportRows = 0;
let supportSelectorAtoms = 0;
const backendCatalogRowIDs = new Set();

for (let phase = 0; phase <= 12; phase += 1) {
  const sourcePath = `tools/phase${phase}_test_map.json`;
  const source = readJSON(sourcePath);
  const legacyRows = [
    ...(source.unit ?? []),
    ...(source.integration ?? []),
    ...(source.e2e ?? []),
    ...(source.visual ?? []),
  ];
  for (const legacy of legacyRows) {
    const frozen = baselineBackendByID.get(legacy.id);
    const disposition = dispositionByID.get(legacy.id);
    if (!frozen || frozen.source_path !== sourcePath) fail(`${legacy.id} frozen source identity drift`);
    if (!disposition) fail(`${legacy.id} has no disposition`);
    const row = catalog.rowByID.get(disposition.row_id);
    if (!row) fail(`${legacy.id} row ${disposition.row_id} is absent`);
    if (backendCatalogRowIDs.has(row.row_id)) fail(`${row.row_id} represents multiple backend identities`);
    backendCatalogRowIDs.add(row.row_id);

    const ownerID = expectedOwner(phase, legacy);
    const familyID = `${ownerID}.${expectedFamily(legacy)}`;
    const verificationIDs = [`${ownerID}.verification.behavior_contract`];
    const collaborators = expectedCollaborators(legacy, ownerID);
    const fixtureID = expectedFixture(legacy);
    const evidence = expectedEvidence(legacy);
    const runner = legacy.runner === "go_test" || legacy.runner === undefined ? "go" : legacy.runner;
    const selectorValues = runner === "go"
      ? exactValues(legacy, "symbols", "symbol")
      : exactValues(legacy, "titles", "title");

    if (row.owner_id !== ownerID || disposition.owner_id !== ownerID) fail(`${legacy.id} owner drift`);
    if (row.family_id !== familyID || !row.row_id.startsWith(`${familyID}.`)) fail(`${legacy.id} family drift`);
    compare(row.verification_ids, verificationIDs, `${legacy.id} row verifications`);
    compare(disposition.verification_ids, verificationIDs, `${legacy.id} disposition verifications`);
    compare(row.collaborator_ids, collaborators, `${legacy.id} row collaborators`);
    compare(disposition.collaborator_ids, collaborators, `${legacy.id} disposition collaborators`);
    if (row.runner !== runner) fail(`${legacy.id} runner drift`);
    compare(row.selector.tests ?? row.selector.titles, selectorValues, `${legacy.id} selector atoms`);
    if (runner === "go" && row.selector.package !== legacy.package) fail(`${legacy.id} package drift`);
    if (runner !== "go" && row.selector.file !== legacy.file) fail(`${legacy.id} file drift`);
    if (row.evidence_class !== evidence) fail(`${legacy.id} evidence-class drift`);
    if (row.fixture_profile_id !== fixtureID) fail(`${legacy.id} fixture-profile drift`);
    if (row.runtime_profile_id !== expectedRuntime(legacy, fixtureID, ownerID)) fail(`${legacy.id} runtime-profile drift`);
    if (row.resource_profile_id !== expectedResource(legacy, fixtureID)) fail(`${legacy.id} resource-profile drift`);
    const defaultCheck = ["visual", "measurement"].includes(evidence) ? false : legacy.default_check_required;
    if (row.default_check !== defaultCheck) fail(`${legacy.id} default-check drift`);
    const claimPosture = ["visual", "measurement"].includes(evidence) ? "informative" : "implementation";
    if (row.claim_posture !== claimPosture) fail(`${legacy.id} claim-posture drift`);
    if (/(?:phase_?\d+|fe_[a-z_]*p_?\d+|sprint_?\d+)/iu.test(row.row_id)) fail(`${legacy.id} row ID retains delivery identity`);

    authoritativeRows += 1;
    authoritativeSelectorAtoms += selectorValues.length;
    increment(ownerCounts, ownerID);
    increment(phaseCounts, `phase${phase}`);
    increment(runnerCounts, runner);
    increment(evidenceCounts, evidence);
    increment(fixtureCounts, fixtureID);
    increment(runtimeCounts, row.runtime_profile_id);
    increment(resourceCounts, row.resource_profile_id);
    increment(defaultCheckCounts, String(row.default_check));
    increment(claimPostureCounts, row.claim_posture);
  }

  const frozenSupport = supportCandidates.filter((candidate) => candidate.source_path === sourcePath);
  if (frozenSupport.length !== (source.support_go_targets ?? []).length) fail(`${sourcePath} support cardinality drift`);
  for (const [index, legacy] of (source.support_go_targets ?? []).entries()) {
    const candidate = frozenSupport[index];
    const disposition = supportDispositionByID.get(candidate.candidate_id);
    const authorization = supportAuthorizationByID.get(candidate.candidate_id);
    if (!disposition || disposition.disposition !== "retained") fail(`${candidate.candidate_id} is not retained`);
    const row = catalog.rowByID.get(authorization.row_id);
    if (!row) fail(`${candidate.candidate_id} row is absent`);
    if (backendCatalogRowIDs.has(row.row_id)) fail(`${row.row_id} overlaps another backend row`);
    backendCatalogRowIDs.add(row.row_id);
    const ownerID = packageOwner(legacy.package);
    const familyID = `${ownerID}.${expectedFamily(legacy, true)}`;
    const fixtureID = expectedFixture(legacy, true);
    const selectors = exactValues(legacy, "symbols", "symbol");
    if (row.owner_id !== ownerID || authorization.owner_id !== ownerID) fail(`${candidate.candidate_id} owner drift`);
    if (row.family_id !== familyID || !row.row_id.startsWith(`${familyID}.`)) fail(`${candidate.candidate_id} family drift`);
    compare(row.selector.tests, selectors, `${candidate.candidate_id} selectors`);
    if (row.selector.package !== legacy.package || row.runner !== "go") fail(`${candidate.candidate_id} Go selector drift`);
    if (row.default_check || row.claim_posture !== "implementation") fail(`${candidate.candidate_id} support posture drift`);
    if (row.evidence_class !== expectedEvidence(legacy, true)) fail(`${candidate.candidate_id} evidence drift`);
    if (row.fixture_profile_id !== fixtureID) fail(`${candidate.candidate_id} fixture drift`);
    if (row.runtime_profile_id !== expectedRuntime(legacy, fixtureID, ownerID)) fail(`${candidate.candidate_id} runtime drift`);
    if (row.resource_profile_id !== expectedResource(legacy, fixtureID)) fail(`${candidate.candidate_id} resource drift`);
    supportRows += 1;
    supportSelectorAtoms += selectors.length;
  }
}

compare(
  { authoritativeRows, authoritativeSelectorAtoms, supportRows, supportSelectorAtoms },
  { authoritativeRows: 456, authoritativeSelectorAtoms: 550, supportRows: 37, supportSelectorAtoms: 118 },
  "closed backend totals",
);
if (backendCatalogRowIDs.size !== 493) fail("backend catalog row population differs from 493");
compare(sortedCounts(runnerCounts), { go: 335, playwright: 92, vitest: 29 }, "backend runner totals");

const summary = {
  schema_id: "cartulary.test_backend_reconciliation_summary.v1",
  status: "pass",
  authoritative_population: authoritativeRows,
  authoritative_selector_atoms: authoritativeSelectorAtoms,
  support_population: supportRows,
  support_selector_atoms: supportSelectorAtoms,
  backend_catalog_rows: backendCatalogRowIDs.size,
  remaining_pending_by_source: sortedCounts(
    Object.fromEntries(
      [...new Set(crosswalk.pending_baseline_keys.map((entry) => entry.source_registry_id))]
        .map((sourceID) => [sourceID, crosswalk.pending_baseline_keys.filter((entry) => entry.source_registry_id === sourceID).length]),
    ),
  ),
  counts_by_phase: sortedCounts(phaseCounts),
  counts_by_owner: sortedCounts(ownerCounts),
  counts_by_runner: sortedCounts(runnerCounts),
  counts_by_evidence_class: sortedCounts(evidenceCounts),
  counts_by_fixture_profile: sortedCounts(fixtureCounts),
  counts_by_runtime_profile: sortedCounts(runtimeCounts),
  counts_by_resource_profile: sortedCounts(resourceCounts),
  counts_by_default_check: sortedCounts(defaultCheckCounts),
  counts_by_claim_posture: sortedCounts(claimPostureCounts),
  catalog_semantic_digest: catalog.semantic_digest,
  verification_semantic_digest: catalog.verification.semantic_digest,
};
process.stdout.write(`${JSON.stringify(summary)}\n`);
