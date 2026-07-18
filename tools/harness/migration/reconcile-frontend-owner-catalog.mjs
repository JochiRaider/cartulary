#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { evidenceTargetForCatalogRow } from "../evidence-accounting/index.mjs";
import { loadTestCatalog } from "../test-catalog/test-catalog.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(root, relativePath), "utf8"));
}

function fail(message) {
  throw new Error(`frontend reconciliation failed: ${message}`);
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function semanticFrontendTitle(title) {
  return title
    .replace(/^support\s+Phase\s+\d+\s+/iu, "support ")
    .replace(/^Phase\s+\d+\s+/iu, "")
    .replace(/^Sprint\s+\d+\s+/iu, "")
    .replace(/^(?:FE-[A-Z]+-P\d+(?:-[A-Z0-9]+)*|[UIEV]-\d+(?:-[A-Z0-9]+)+)\s+/iu, "");
}

function sorted(values) {
  return [...values].sort(asciiCompare);
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

function expectedOwner(phase, row) {
  if (phase === 0) {
    return {
      "FE-S-P0-01": "harness.generated_artifacts",
      "FE-S-P0-02": "web.architecture",
      "FE-S-P0-03": "harness.test_catalog",
      "FE-S-P0-04": "harness.generated_artifacts",
      "FE-U-P0-01": "package.protocol_ts",
      "FE-U-P0-02": "package.view_contracts",
      "FE-U-P0-03": "package.ui",
    }[row.id];
  }
  if (phase === 1) return "module.auth";
  if (phase === 2) return row.id === "FE-E-P2-01" ? "module.savedviews" : "module.workbook";
  if (phase === 3) {
    if (["FE-U-P3-01", "FE-U-P3-02", "FE-U-P3-03", "FE-U-P3-04", "FE-V-P3-01", "FE-A11Y-P3-01"].includes(row.id)) {
      return "package.grid_adapter";
    }
    if (row.id === "FE-B-P3-01") return "module.timeline";
    return row.id === "FE-S-P3-01" ? "web.architecture" : "module.workbook";
  }
  if (phase === 4) return "module.workbook";
  if (phase === 5) return "module.entities";
  if (phase === 6) return "module.evidence";
  if (phase === 7) return "module.collaboration";
  if (phase === 8) {
    if (row.id === "FE-U-P8-01") return "platform.viewquery";
    if (row.id === "FE-B-P8-01") return "module.workbook";
    return "module.savedviews";
  }
  if (phase === 9) return ["FE-I-P9-01", "FE-E-P9-01"].includes(row.id) ? "module.revisions" : "module.workbook";
  if (phase === 10) return row.id === "FE-B-P10-02" ? "module.timeline" : "module.workbook";
  if (phase === 11) return row.id.startsWith("FE-S-") ? "harness.test_catalog" : "web.design";
  if (phase === 12) return "module.networkflow";
  fail(`unclassified frontend owner ${phase}:${row.id}`);
}

function expectedVerification(ownerID) {
  return ownerID === "web.design"
    ? "web.design.verification.readiness_direction"
    : `${ownerID}.verification.behavior_contract`;
}

function expectedEvidence(row) {
  if (row.runner === "shell") return "static";
  if (row.runner === "vitest") return "unit";
  const stages = new Set(row.targets.map((entry) => entry.target_name));
  if (stages.has("browser-e2e-measurement")) return "measurement";
  if (stages.has("browser-e2e-a11y")) return "accessibility";
  if (stages.has("browser-e2e-visual")) return "visual";
  return "browser";
}

function expectedStageForFile(row, file) {
  const targets = new Set(row.targets.map((entry) => entry.target_name));
  if (targets.has("browser-e2e-measurement")) return "measurement";
  if (targets.has("browser-e2e-a11y")) return "accessibility";
  if (targets.has("browser-e2e-visual")) return "visual";
  if (targets.has("browser-e2e-support") && file.includes(".support.spec.")) return "support";
  if (targets.has("browser-e2e-stateful")) return "stateful";
  return "webserver_backed";
}

function sourceRow(frozen) {
  const source = readJSON(frozen.source_path);
  const row = source.rows.find((entry) => entry.id === frozen.legacy_row_id);
  if (!row) fail(`${frozen.legacy_row_id} is absent from ${frozen.source_path}`);
  return {
    ...row,
    runner: frozen.runner,
  };
}

const catalog = loadTestCatalog(root);
const baseline = readJSON("tools/test_migration_baseline.json");
const crosswalk = readJSON("tools/test_migration_crosswalk.json");
const taskSurface = readJSON("tools/task_surface_manifest.json");
const commandTargetByID = new Map(taskSurface.targets.map((entry) => [entry.command_id, entry.name]));
const frontendBaseline = baseline.identities.filter((entry) => entry.source_registry_id === "frontend_phase_maps");
const frontendDispositions = crosswalk.dispositions.filter((entry) => entry.source_registry_id === "frontend_phase_maps");
const dispositionByID = new Map(frontendDispositions.map((entry) => [entry.legacy_row_id, entry]));
if (frontendBaseline.length !== 87 || new Set(frontendBaseline.map((entry) => entry.legacy_row_id)).size !== 87) {
  fail("frozen frontend population differs from 87");
}
if (frontendDispositions.length !== 87 || dispositionByID.size !== 87) fail("expected 87 unique frontend dispositions");
if (crosswalk.pending_baseline_keys.length !== 0 || crosswalk.dispositions.length !== 548) {
  fail("authoritative crosswalk is not fully terminal");
}

const allAuthorizedRows = new Set([
  ...crosswalk.dispositions.filter((entry) => entry.disposition === "migrated").map((entry) => entry.row_id),
  ...crosswalk.new_rows.map((entry) => entry.row_id),
]);
const frontendNewRows = crosswalk.new_rows.filter((entry) => entry.provenance.startsWith("frontend:"));
const dispositionCounts = {};
const runnerCounts = {};
const ownerCounts = {};
const evidenceCounts = {};
const runtimeCounts = {};
const resourceCounts = {};
const fixtureCounts = {};
const defaultCheckCounts = {};
const claimPostureCounts = {};
const referencedCatalogRows = new Set();
let legacyTitleAtoms = 0;
let resolvedTitleSelectors = 0;
let legacyCommandAtoms = 0;
let resolvedCommandSelectors = 0;

for (const frozen of frontendBaseline) {
  const disposition = dispositionByID.get(frozen.legacy_row_id);
  if (!disposition) fail(`${frozen.legacy_row_id} has no terminal disposition`);
  const legacy = sourceRow(frozen);
  const phaseMatch = /^FE-P(\d+)$/u.exec(legacy.phase_id);
  if (!phaseMatch) fail(`${legacy.id} has invalid source phase identity`);
  const phase = Number(phaseMatch[1]);
  const ownerID = expectedOwner(phase, legacy);
  increment(dispositionCounts, disposition.disposition);
  increment(runnerCounts, frozen.runner);
  const titles = (frozen.selector.scenario_titles ?? []).map(semanticFrontendTitle);
  const commands = frozen.selector.command_ids ?? [];
  legacyTitleAtoms += titles.length;
  legacyCommandAtoms += commands.length;

  const splitAuthorizations = frontendNewRows.filter((entry) =>
    entry.provenance.startsWith(`frontend:${legacy.id} `),
  );
  if (disposition.disposition === "deleted") {
    if (!["FE-S-P0-03", "FE-S-P11-01", "FE-S-P11-02", "FE-S-P11-03"].includes(legacy.id)) {
      fail(`${legacy.id} has an unauthorized deletion`);
    }
    if (splitAuthorizations.length !== 0) fail(`${legacy.id} deletion retains split row authorizations`);
    continue;
  }

  if (disposition.owner_id !== ownerID || disposition.review.governing_owner !== ownerID) {
    fail(`${legacy.id} disposition owner drift`);
  }
  const verificationID = expectedVerification(ownerID);
  compare(disposition.verification_ids, [verificationID], `${legacy.id} disposition verifications`);

  const directRowIDs = disposition.disposition === "migrated"
    ? [disposition.row_id, ...splitAuthorizations.map((entry) => entry.row_id)]
    : [disposition.surviving_row_id];
  if (disposition.disposition === "consolidated") {
    for (const row of catalog.rows) {
      if (
        row.runner === frozen.runner &&
        (row.selector.titles ?? []).some((title) => titles.includes(title))
      ) {
        directRowIDs.push(row.row_id);
      }
      if (
        row.runner === "shell" &&
        commands.includes(evidenceTargetForCatalogRow(row, { commandTargetByID }))
      ) {
        directRowIDs.push(row.row_id);
      }
    }
  }
  const directRows = [...new Set(directRowIDs)].map((rowID) => {
    const row = catalog.rowByID.get(rowID);
    if (!row) fail(`${legacy.id} references absent catalog row ${rowID}`);
    if (!allAuthorizedRows.has(rowID)) fail(`${legacy.id} references unauthorized catalog row ${rowID}`);
    referencedCatalogRows.add(rowID);
    return row;
  });

  for (const title of titles) {
    const exactMatches = catalog.rows.filter(
      (row) => row.runner === frozen.runner && (row.selector.titles ?? []).includes(title),
    );
    if (exactMatches.length === 0) fail(`${legacy.id} title is absent from the catalog: ${title}`);
    if (!directRows.some((row) => (row.selector.titles ?? []).includes(title))) {
      fail(`${legacy.id} direct rows do not preserve title: ${title}`);
    }
    for (const row of exactMatches) {
      if (!allAuthorizedRows.has(row.row_id)) fail(`${legacy.id} title resolves through unauthorized ${row.row_id}`);
      referencedCatalogRows.add(row.row_id);
    }
    resolvedTitleSelectors += exactMatches.length;
  }
  for (const targetName of commands) {
    const exactMatches = catalog.rows.filter(
      (row) => row.runner === "shell" && evidenceTargetForCatalogRow(row, { commandTargetByID }) === targetName,
    );
    if (exactMatches.length !== 1) fail(`${legacy.id} command ${targetName} resolves ${exactMatches.length} rows`);
    if (!directRows.some((row) => row.row_id === exactMatches[0].row_id)) {
      fail(`${legacy.id} direct rows do not preserve command ${targetName}`);
    }
    referencedCatalogRows.add(exactMatches[0].row_id);
    resolvedCommandSelectors += 1;
  }

  for (const row of directRows) {
    if (disposition.disposition === "migrated" && row.owner_id !== ownerID) {
      fail(`${legacy.id} row ${row.row_id} owner drift`);
    }
    compare(
      row.verification_ids,
      [expectedVerification(row.owner_id)],
      `${legacy.id} row ${row.row_id} verifications`,
    );
    if (row.runner !== frozen.runner) fail(`${legacy.id} row ${row.row_id} runner drift`);
    if (/(?:phase_?\d+|fe_[a-z_]*p_?\d+|sprint_?\d+)/iu.test(row.row_id)) {
      fail(`${legacy.id} row ID retains delivery identity`);
    }
    const evidence = expectedEvidence(legacy);
    if (row.evidence_class !== evidence) fail(`${legacy.id} row ${row.row_id} evidence-class drift`);
    if (row.runner === "playwright") {
      if (disposition.disposition === "consolidated") {
        const target = evidenceTargetForCatalogRow(row, { commandTargetByID });
        if (!legacy.targets.some((entry) => entry.target_name === target)) {
          fail(`${legacy.id} row ${row.row_id} consolidated stage drift`);
        }
      } else if (row.selector.stage !== expectedStageForFile(legacy, row.selector.file)) {
        fail(`${legacy.id} row ${row.row_id} stage drift`);
      }
      if (row.fixture_profile_id !== "service_stack" || row.resource_profile_id !== "browser_exclusive") {
        fail(`${legacy.id} row ${row.row_id} browser profile drift`);
      }
      const runtime = ownerID === "module.networkflow" ? "network_flow_claimed" : "default";
      if (row.runtime_profile_id !== runtime) fail(`${legacy.id} row ${row.row_id} runtime-profile drift`);
    } else if (
      row.runtime_profile_id !== "none" ||
      row.resource_profile_id !== "none" ||
      row.fixture_profile_id !== "none"
    ) {
      fail(`${legacy.id} row ${row.row_id} non-runtime profile drift`);
    }
    const informative = ["accessibility", "measurement", "visual"].includes(evidence);
    if (row.claim_posture !== (informative ? "informative" : "implementation")) {
      fail(`${legacy.id} row ${row.row_id} claim-posture drift`);
    }
    if (
      disposition.disposition !== "consolidated" &&
      row.default_check !== (informative ? false : legacy.default_check_required)
    ) {
      fail(`${legacy.id} row ${row.row_id} default-check drift`);
    }
    increment(ownerCounts, row.owner_id);
    increment(evidenceCounts, row.evidence_class);
    increment(runtimeCounts, row.runtime_profile_id);
    increment(resourceCounts, row.resource_profile_id);
    increment(fixtureCounts, row.fixture_profile_id);
    increment(defaultCheckCounts, String(row.default_check));
    increment(claimPostureCounts, row.claim_posture);
  }
}

compare(sortedCounts(dispositionCounts), { consolidated: 6, deleted: 4, migrated: 77 }, "frontend dispositions");
compare(sortedCounts(runnerCounts), { playwright: 56, shell: 6, vitest: 25 }, "frontend frozen runners");
if (frontendNewRows.length !== 16) fail(`expected 16 frontend split-row authorizations, found ${frontendNewRows.length}`);
if (legacyTitleAtoms !== 193 || legacyCommandAtoms !== 23) fail("legacy frontend selector atom totals drifted");

const liveLegacyTitles = new Set(frontendBaseline.flatMap((entry) => entry.selector.scenario_titles ?? []));
const frontendAuthorizedRowIDs = new Set([
  ...frontendDispositions.filter((entry) => entry.disposition === "migrated").map((entry) => entry.row_id),
  ...frontendNewRows.map((entry) => entry.row_id),
]);
for (const row of catalog.rows) {
  if (!frontendAuthorizedRowIDs.has(row.row_id)) continue;
  for (const title of row.selector.titles ?? []) {
    if (/^FE-(?:A11Y|B|E|I|S|U|V)-P\d+-/u.test(title) && !liveLegacyTitles.has(title)) {
      fail(`catalog contains unaccounted frontend legacy title ${title}`);
    }
  }
}

const summary = {
  schema_id: "cartulary.test_frontend_reconciliation_summary.v1",
  status: "pass",
  authoritative_population: frontendBaseline.length,
  authoritative_title_atoms: legacyTitleAtoms,
  resolved_title_selectors: resolvedTitleSelectors,
  authoritative_command_atoms: legacyCommandAtoms,
  resolved_command_selectors: resolvedCommandSelectors,
  referenced_catalog_rows: referencedCatalogRows.size,
  split_row_authorizations: frontendNewRows.length,
  total_terminal_dispositions: crosswalk.dispositions.length,
  remaining_pending: crosswalk.pending_baseline_keys.length,
  counts_by_disposition: sortedCounts(dispositionCounts),
  counts_by_runner: sortedCounts(runnerCounts),
  counts_by_owner_row_reference: sortedCounts(ownerCounts),
  counts_by_evidence_class: sortedCounts(evidenceCounts),
  counts_by_runtime_profile: sortedCounts(runtimeCounts),
  counts_by_resource_profile: sortedCounts(resourceCounts),
  counts_by_fixture_profile: sortedCounts(fixtureCounts),
  counts_by_default_check: sortedCounts(defaultCheckCounts),
  counts_by_claim_posture: sortedCounts(claimPostureCounts),
  catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
  verification_semantic_digest: catalog.summary.verification_semantic_digest,
};
process.stdout.write(`${JSON.stringify(summary)}\n`);
