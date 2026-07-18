#!/usr/bin/env node
import { createHash } from "node:crypto";
import { mkdirSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "..", "..", "..");
const runnerPath = process.argv[2];
if (!runnerPath) {
  throw new Error("usage: build-owner-catalog.mjs <successful-vitest-runner.json>");
}

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.resolve(root, relativePath), "utf8"));
}

function writeJSON(relativePath, value) {
  writeFileSync(path.resolve(root, relativePath), `${JSON.stringify(value, null, 2)}\n`);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sourceFiles(directory, recursive) {
  const result = [];
  for (const entry of readdirSync(path.resolve(root, directory), { withFileTypes: true })) {
    const relative = path.posix.join(directory, entry.name);
    if (entry.isDirectory() && recursive) {
      result.push(...sourceFiles(relative, true));
    } else if (entry.isFile() && entry.name.endsWith(".test.ts")) {
      result.push(relative);
    }
  }
  return result.sort(asciiCompare);
}

function expandClassification(entry) {
  if (entry.file) {
    return [entry.file];
  }
  if (entry.file_pattern === "apps/web/src/workbook/hooks/useWorkbook*Controller.test.tsx") {
    return readdirSync(path.join(root, "apps/web/src/workbook/hooks"))
      .filter((name) => /^useWorkbook.*Controller\.test\.tsx$/u.test(name))
      .map((name) => `apps/web/src/workbook/hooks/${name}`)
      .sort(asciiCompare);
  }
  if (entry.file_pattern === "apps/web/e2e/*.test.ts") {
    return sourceFiles("apps/web/e2e", false);
  }
  if (entry.file_pattern === "apps/web/e2e/support/**/*.test.ts") {
    return sourceFiles("apps/web/e2e/support", true);
  }
  throw new Error(`unsupported frozen Vitest pattern ${entry.file_pattern}`);
}

function ownerForFile(file) {
  if (file.startsWith("apps/web/e2e/")) return "harness.browser";
  if (file.startsWith("apps/web/src/collaboration/")) return "web.collaboration";
  if (file.startsWith("apps/web/src/networkFlow/")) return "web.networkflow";
  if (file.startsWith("apps/web/src/workbook/")) return "web.workbook";
  if (file.startsWith("apps/web/src/testing/")) return "web.architecture";
  if (
    file.startsWith("apps/web/src/app/") ||
    file.startsWith("apps/web/src/services/") ||
    file.startsWith("apps/web/src/shared/")
  ) return "web.application";
  if (file.startsWith("packages/grid-adapter/")) return "package.grid_adapter";
  if (file.startsWith("packages/protocol-ts/")) return "package.protocol_ts";
  if (file.startsWith("packages/test-utils/")) return "package.test_utils";
  if (file.startsWith("packages/ui-contracts/")) return "package.ui";
  if (file.startsWith("packages/view-contracts/")) return "package.view_contracts";
  throw new Error(`no durable owner for ${file}`);
}

const verificationByOwner = Object.freeze({
  "harness.browser": "harness.browser.verification.support_contract",
  "package.grid_adapter": "package.grid_adapter.verification.adapter_contract",
  "package.protocol_ts": "package.protocol_ts.verification.shared_contract",
  "package.test_utils": "package.test_utils.verification.shared_contract",
  "package.ui": "package.ui.verification.shared_contract",
  "package.view_contracts": "package.view_contracts.verification.shared_contract",
  "web.application": "web.application.verification.frontend_regression",
  "web.architecture": "web.architecture.verification.test_boundaries",
  "web.collaboration": "web.collaboration.verification.session_regression",
  "web.networkflow": "web.networkflow.verification.extension_regression",
  "web.workbook": "web.workbook.verification.frontend_regression",
});

function rowSegment(file, title) {
  const basis = `${file}\0${title ?? "all"}`;
  const semanticSlug = (value) => value
    .replace(/\.(?:test|spec)\.[^.]+(?:x)?$/u, "")
    .replace(/[^a-zA-Z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "")
    .toLowerCase()
    .replace(/(?:^|_)(?:phase_?[0-9]+|fe_(?:[a-z]_)?p_?[0-9]+)(?=_|$)/gu, "_")
    .replace(/^_+|_+$/gu, "") || "behavior";
  const stem = semanticSlug(path.basename(file));
  const titleStem = title ? semanticSlug(title) : "suite";
  const compact = `${stem}_${titleStem}`.slice(0, 48).replace(/_+$/u, "");
  const digest = createHash("sha256").update(basis).digest("hex").slice(0, 10);
  return `${compact}_${digest}`;
}

const runner = JSON.parse(readFileSync(path.resolve(root, runnerPath), "utf8"));
const titleIndex = new Map();
for (const fileResult of runner.testResults) {
  const file = path.relative(root, fileResult.name).split(path.sep).join("/");
  titleIndex.set(file, fileResult.assertionResults.map((entry) => ({
    fullName: entry.fullName,
    title: entry.title,
  })));
}

const classification = readJSON("tools/test_accounting_classification.json");
const baseline = readJSON("tools/test_migration_baseline.json");
const individuallyClaimedTitles = new Map();
for (const entry of classification.vitest) {
  if (!entry.file || !entry.title) continue;
  const discovered = titleIndex.get(entry.file) ?? [];
  const matches = discovered.filter((item) => item.title === entry.title || item.fullName === entry.title);
  if (matches.length === 1) {
    if (!individuallyClaimedTitles.has(entry.file)) individuallyClaimedTitles.set(entry.file, new Set());
    individuallyClaimedTitles.get(entry.file).add(matches[0].fullName);
  }
}
const manifests = new Map();
const auxiliaryDispositions = [];
const newRows = [];

function appendRow(ownerID, row) {
  if (!manifests.has(ownerID)) manifests.set(ownerID, []);
  manifests.get(ownerID).push(row);
}

for (const [index, entry] of classification.vitest.entries()) {
  const candidate = baseline.auxiliary_candidates.vitest_classifications[index];
  if (!candidate || candidate.candidate_id !== `vitest:${String(index + 1).padStart(3, "0")}`) {
    throw new Error(`frozen Vitest candidate ${index + 1} is inconsistent`);
  }
  if (entry.file === "apps/web/src/Unmapped.frontend-unit.test.tsx") {
    auxiliaryDispositions.push({
      candidate_id: candidate.candidate_id,
      disposition: "deleted",
      reason: "The absent Unmapped file is a deliberate legacy accounting fixture, not production evidence.",
      owner_review_evidence: "WS-03 T-017 explicit non-production deletion",
    });
    continue;
  }
  if (
    entry.file === "apps/web/src/networkFlow/networkFlowClient.test.ts" &&
    entry.title === "rejects an empty graph table scope before transport"
  ) {
    auxiliaryDispositions.push({
      candidate_id: candidate.candidate_id,
      disposition: "deleted",
      reason: "The classified title has no executable assertion and describes compile-time typing already enforced by frontend typecheck.",
      owner_review_evidence: "WS-03 T-017 explicit obsolete duplicate deletion",
    });
    continue;
  }
  const files = expandClassification(entry);
  auxiliaryDispositions.push({
    candidate_id: candidate.candidate_id,
    disposition: "retained",
    reason: `Retained as ${files.length} exact owner-catalog row${files.length === 1 ? "" : "s"}; glob and catch-all ownership are removed.`,
    owner_review_evidence: "WS-03 T-017 durable owner adjudication",
  });
  for (const file of files) {
    const discovered = titleIndex.get(file);
    if (!discovered || discovered.length === 0) {
      throw new Error(`successful Vitest evidence has no assertions for ${file}`);
    }
    let titles;
    if (entry.title) {
      const matches = discovered.filter((item) => item.title === entry.title || item.fullName === entry.title);
      if (matches.length !== 1) {
        throw new Error(`${candidate.candidate_id} title resolves ${matches.length} times in ${file}`);
      }
      titles = [matches[0].fullName];
    } else {
      const individuallyClaimed = individuallyClaimedTitles.get(file) ?? new Set();
      titles = discovered
        .map((item) => item.fullName)
        .filter((title) => !individuallyClaimed.has(title))
        .sort(asciiCompare);
      if (titles.length === 0) {
        throw new Error(`${candidate.candidate_id} has no selectors after exact-title partitioning`);
      }
    }
    const ownerID = ownerForFile(file);
    const familySegment = entry.coverage === "tooling_support" ? "boundary_support" : "regression";
    const rowID = `${ownerID}.${familySegment}.${rowSegment(file, entry.title)}`;
    const row = {
      row_id: rowID,
      owner_id: ownerID,
      family_id: `${ownerID}.${familySegment}`,
      collaborator_ids: [],
      verification_ids: [verificationByOwner[ownerID]],
      runner: "vitest",
      selector: { file, titles },
      evidence_class: entry.coverage === "tooling_support" ? "static" : "unit",
      runtime_profile_id: "none",
      resource_profile_id: "none",
      fixture_profile_id: "none",
      default_check: true,
      claim_posture: "implementation",
      status: "active",
    };
    appendRow(ownerID, row);
    newRows.push({
      owner_id: ownerID,
      row_id: rowID,
      verification_ids: [verificationByOwner[ownerID]],
      provenance: `${candidate.candidate_id} frozen Vitest classification; exact file/title selectors from ${runnerPath}`,
    });
  }
}

const graphRows = [
  {
    legacy: "GP-UNIT-001",
    row_id: "module.graphprojection.engine.canonical_behavior",
    family_id: "module.graphprojection.engine",
    package: "./internal/modules/graphprojection",
    evidence_class: "unit",
    resource_profile_id: "go_balanced",
    fixture_profile_id: "none",
  },
  {
    legacy: "GP-UNIT-002",
    row_id: "module.graphprojection.fixtures.fixture_verifier",
    family_id: "module.graphprojection.fixtures",
    package: "./internal/modules/graphprojection/fixturetest",
    evidence_class: "unit",
    resource_profile_id: "go_balanced",
    fixture_profile_id: "none",
  },
  {
    legacy: "GP-STORE-001",
    row_id: "module.graphprojection.storage.lifecycle",
    family_id: "module.graphprojection.storage",
    package: "./internal/modules/graphprojection",
    evidence_class: "integration",
    resource_profile_id: "go_transaction_heavy",
    fixture_profile_id: "postgres_transaction",
  },
  {
    legacy: "GP-MIGRATION-001",
    row_id: "module.graphprojection.storage.migration_reset",
    family_id: "module.graphprojection.storage",
    package: "./internal/platform/postgres",
    evidence_class: "integration",
    resource_profile_id: "go_io_heavy",
    fixture_profile_id: "postgres_migration_scratch",
    collaborator_ids: ["platform.postgres"],
  },
  {
    legacy: "GP-BINDING-001",
    row_id: "module.graphprojection.storage.transaction_binding",
    family_id: "module.graphprojection.storage",
    package: "./internal/modules/graphprojection/postgresbinding",
    evidence_class: "integration",
    resource_profile_id: "go_transaction_heavy",
    fixture_profile_id: "postgres_transaction",
  },
];
const graphSource = readJSON("tools/subsystem_test_maps/graph_projection_test_map.json");
const graphEntries = [...graphSource.unit, ...graphSource.integration];
const graphDispositions = [];
for (const definition of graphRows) {
  const legacy = graphEntries.find((entry) => entry.id === definition.legacy);
  if (!legacy) throw new Error(`missing Graph Projection row ${definition.legacy}`);
  const symbols = legacy.symbols ?? [legacy.symbol];
  const collaboratorIDs = definition.collaborator_ids ?? [];
  appendRow("module.graphprojection", {
    row_id: definition.row_id,
    owner_id: "module.graphprojection",
    family_id: definition.family_id,
    collaborator_ids: collaboratorIDs,
    verification_ids: ["module.graphprojection.verification.engine_and_storage"],
    runner: "go",
    selector: { package: definition.package, tests: [...symbols].sort(asciiCompare) },
    evidence_class: definition.evidence_class,
    runtime_profile_id: "none",
    resource_profile_id: definition.resource_profile_id,
    fixture_profile_id: definition.fixture_profile_id,
    default_check: true,
    claim_posture: "implementation",
    status: "active",
  });
  graphDispositions.push({
    source_registry_id: "graph_projection_subsystem",
    legacy_row_id: definition.legacy,
    disposition: "migrated",
    owner_id: "module.graphprojection",
    row_id: definition.row_id,
    verification_ids: ["module.graphprojection.verification.engine_and_storage"],
    adjudication_rule: "normative_postcondition",
    collaborator_ids: collaboratorIDs,
    review: {
      governing_owner: "module.graphprojection",
      governing_requirement: "Adopted Graph Projection behavior and boundary postconditions",
      review_revision: "b0ae8400",
      owner_review_evidence: "WS-03 T-016 Graph Projection owner absorption",
    },
  });
}

mkdirSync(path.join(root, "tools/test_families"), { recursive: true });
for (const [ownerID, rows] of [...manifests.entries()].sort(([left], [right]) => asciiCompare(left, right))) {
  rows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
  writeJSON(`tools/test_families/${ownerID}.json`, {
    schema_id: "cartulary.test_family_manifest.v1",
    owner_id: ownerID,
    rows,
  });
}
const ownerIDs = [...manifests.keys()].sort(asciiCompare);
writeJSON("tools/test_catalog_owner.json", {
  schema_id: "cartulary.test_owner_registry.v1",
  owners: ownerIDs.map((ownerID) => ({
    owner_id: ownerID,
    manifest_path: `tools/test_families/${ownerID}.json`,
    status: "active",
  })),
});

classification.vitest = classification.vitest.map((entry) => ({
  ...entry,
  coverage: entry.coverage === "unowned_regression" ? "support" : entry.coverage,
  reason: entry.coverage === "unowned_regression"
    ? `${entry.reason} Ownership is now explicit in the unified test catalog.`
    : entry.reason,
}));
writeJSON("tools/test_accounting_classification.json", classification);

const crosswalk = readJSON("tools/test_migration_crosswalk.json");
const graphKeys = new Set(graphRows.map((entry) => entry.legacy));
crosswalk.pending_baseline_keys = crosswalk.pending_baseline_keys.filter(
  (entry) => entry.source_registry_id !== "graph_projection_subsystem" || !graphKeys.has(entry.legacy_row_id),
);
crosswalk.dispositions = crosswalk.dispositions.filter(
  (entry) => entry.source_registry_id !== "graph_projection_subsystem" || !graphKeys.has(entry.legacy_row_id),
);
crosswalk.dispositions.push(...graphDispositions);
crosswalk.dispositions.sort((left, right) =>
  asciiCompare(`${left.source_registry_id}:${left.legacy_row_id}`, `${right.source_registry_id}:${right.legacy_row_id}`),
);
crosswalk.auxiliary_dispositions = auxiliaryDispositions;
crosswalk.new_rows = newRows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
writeJSON("tools/test_migration_crosswalk.json", crosswalk);

process.stdout.write(`${JSON.stringify({
  owner_count: ownerIDs.length,
  row_count: [...manifests.values()].reduce((sum, rows) => sum + rows.length, 0),
  graph_rows: graphRows.length,
  vitest_candidates: auxiliaryDispositions.length,
  vitest_new_rows: newRows.length,
})}\n`);
