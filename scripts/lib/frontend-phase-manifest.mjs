import { existsSync, readdirSync } from "node:fs";
import path from "node:path";

import {
  assertObjectKeys,
  assertUnique,
  readJsonObject,
  requireEnum,
  requireObjectArray,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "./json-shape.mjs";

export const frontendPhaseNamespace = "frontend";
export const frontendPhaseRegistrySchemaID =
  "cartulary.frontend_phase_registry.v1";
export const frontendPhaseTestMapSchemaID =
  "cartulary.frontend_phase_test_map.v1";

const registryKeys = new Set([
  "schema_id",
  "phase_namespace",
  "guide_path",
  "phases",
]);
const registryEntryKeys = new Set([
  "phase_id",
  "status",
  "manifest_path",
  "ledger_path",
  "owner_refs",
  "depends_on",
]);
const mapKeys = new Set(["schema_id", "phase_namespace", "phase_id", "rows"]);
const rowKeys = new Set([
  "id",
  "layer",
  "evidence_class",
  "owner_refs",
  "core_req_ids",
  "core_ac_ids",
  "support_or_design_ac_ids",
  "targets",
  "scenario_titles",
  "claim_status",
  "claim",
  "out_of_scope",
]);
const validStatuses = new Set(["planned", "active", "retired"]);
const validEvidenceClasses = new Set([
  "product_conformance",
  "design_direction",
  "implementation_support",
  "claim_publication_boundary",
  "TODO_owner_lookup",
]);
const validClaimStatuses = new Set([
  "implemented",
  "blocked",
  "not_applicable",
]);
const validLayers = new Set([
  "unit",
  "integration",
  "browser_integration",
  "e2e",
  "visual_regression",
  "accessibility",
  "support",
]);
const phaseIDPattern = /^FE-P(?:0|[1-9]\d*)$/;
const phaseMapFilenamePattern = /^fe_p(0|[1-9]\d*)_test_map\.json$/;
const phaseLedgerFilenamePattern = /^fe_p(0|[1-9]\d*)_coverage_ledger\.md$/;
const rowIDPattern = /^FE-(?:U|I|B|E|V|A11Y|S)-P(?:0|[1-9]\d*)-\d{2}$/;
const core03SortingFilteringGroupingReqIDs = new Set(
  Array.from({ length: 13 }, (_, index) => `REQ-03-${223 + index}`),
);
const core03Section14OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])14(?:[^\d.]|$)/;
const core03Section48OwnerRefPattern =
  /^Core 03 Section(?:s)?\b.*(?:^|[^\d.])4\.8(?:[^\d.]|$)/;

function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

function phaseNumber(phaseID) {
  const match = /^FE-P(0|[1-9]\d*)$/.exec(phaseID);
  if (!match) {
    throw new Error(`frontend phase id ${phaseID} must match FE-P<N>`);
  }
  return match[1];
}

function phaseFromMapPath(manifestPath, label) {
  const match = phaseMapFilenamePattern.exec(path.posix.basename(manifestPath));
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_test_map.json`);
  }
  return `FE-P${match[1]}`;
}

function phaseFromLedgerPath(ledgerPath, label) {
  const match = phaseLedgerFilenamePattern.exec(
    path.posix.basename(ledgerPath),
  );
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_coverage_ledger.md`);
  }
  return `FE-P${match[1]}`;
}

function requirePhaseID(value, label) {
  return requireString(value, label, { pattern: phaseIDPattern });
}

function validateRowMetadata(row, label) {
  const evidenceClass = row.evidence_class;
  if (evidenceClass === "product_conformance") {
    if (row.core_req_ids.length === 0 || row.core_ac_ids.length === 0) {
      throw new Error(
        `${label} product_conformance rows must declare core_req_ids[] and core_ac_ids[]`,
      );
    }
    return;
  }
  if (
    evidenceClass === "design_direction" ||
    evidenceClass === "implementation_support"
  ) {
    if (row.core_req_ids.length !== 0 || row.core_ac_ids.length !== 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must not declare Core requirement or acceptance IDs`,
      );
    }
    if (row.support_or_design_ac_ids.length === 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must declare support_or_design_ac_ids[]`,
      );
    }
    return;
  }
  if (evidenceClass === "claim_publication_boundary") {
    for (const id of row.core_req_ids) {
      if (!id.startsWith("REQ-05-")) {
        throw new Error(
          `${label} claim_publication_boundary Core IDs must be Core 05 IDs`,
        );
      }
    }
    return;
  }
  if (evidenceClass === "TODO_owner_lookup") {
    if (row.core_req_ids.length !== 0 || row.core_ac_ids.length !== 0) {
      throw new Error(
        `${label} TODO_owner_lookup rows must not declare Core IDs`,
      );
    }
    if (row.claim_status !== "blocked") {
      throw new Error(
        `${label} TODO_owner_lookup rows must declare claim_status=blocked`,
      );
    }
  }
}

function validateCore03SortingFilteringGroupingOwnerRefs(row, label) {
  const coversSortingFilteringGrouping = row.core_req_ids.some((id) =>
    core03SortingFilteringGroupingReqIDs.has(id),
  );
  if (!coversSortingFilteringGrouping) {
    return;
  }

  const citesCore03Section14 = row.owner_refs.some((ownerRef) =>
    core03Section14OwnerRefPattern.test(ownerRef),
  );
  const citesCore03Section48 = row.owner_refs.some((ownerRef) =>
    core03Section48OwnerRefPattern.test(ownerRef),
  );
  if (!citesCore03Section14 || citesCore03Section48) {
    throw new Error(
      `${label} rows covering REQ-03-223..REQ-03-235 must cite Core 03 Section 14 and must not cite Core 03 Section 4.8`,
    );
  }
}

export function frontendRegistryPath(root = process.cwd()) {
  return repoPath(root, "tools/frontend_phase_registry.json");
}

export function loadFrontendPhaseRegistry(root = process.cwd()) {
  const file = frontendRegistryPath(root);
  const registry = readJsonObject(file, file);
  assertObjectKeys(registry, registryKeys, file);
  requireSchemaID(registry, frontendPhaseRegistrySchemaID, file);
  if (registry.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${file}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  requireRepoRelativePath(registry.guide_path, `${file}.guide_path`, {
    extension: ".md",
  });

  const rawPhases = requireObjectArray(registry.phases, `${file}.phases`, {
    nonEmpty: true,
  });
  if (rawPhases.length !== 12) {
    throw new Error(`${file}.phases must declare exactly FE-P0 through FE-P11`);
  }

  const phases = rawPhases.map((entry, index) => {
    const label = `${file}.phases[${index + 1}]`;
    assertObjectKeys(entry, registryEntryKeys, label);
    const phaseID = requirePhaseID(entry.phase_id, `${label}.phase_id`);
    const status = requireEnum(entry.status, `${label}.status`, validStatuses);
    const manifestPath = requireRepoRelativePath(
      entry.manifest_path,
      `${label}.manifest_path`,
      { extension: ".json" },
    );
    const ledgerPath = requireRepoRelativePath(
      entry.ledger_path,
      `${label}.ledger_path`,
      { extension: ".md" },
    );
    if (phaseFromMapPath(manifestPath, `${label}.manifest_path`) !== phaseID) {
      throw new Error(`${label}.manifest_path must match ${phaseID}`);
    }
    if (phaseFromLedgerPath(ledgerPath, `${label}.ledger_path`) !== phaseID) {
      throw new Error(`${label}.ledger_path must match ${phaseID}`);
    }
    return {
      phase_id: phaseID,
      status,
      manifest_path: manifestPath,
      ledger_path: ledgerPath,
      owner_refs: requireStringArray(entry.owner_refs, `${label}.owner_refs`, {
        nonEmpty: true,
      }),
      depends_on: requireStringArray(entry.depends_on, `${label}.depends_on`),
    };
  });

  assertUnique(
    phases.map((entry) => entry.phase_id),
    `${file}.phases.phase_id`,
  );
  const expected = Array.from({ length: 12 }, (_, index) => `FE-P${index}`);
  const actual = phases.map((entry) => entry.phase_id).sort();
  if (actual.join(",") !== expected.sort().join(",")) {
    throw new Error(
      `${file}.phases must contain exactly ${expected.join(", ")}`,
    );
  }
  const phaseIDs = new Set(actual);
  for (const entry of phases) {
    for (const dependency of entry.depends_on) {
      if (!phaseIDs.has(dependency)) {
        throw new Error(
          `${file} ${entry.phase_id}.depends_on references unknown ${dependency}`,
        );
      }
      if (
        Number(phaseNumber(dependency)) >= Number(phaseNumber(entry.phase_id))
      ) {
        throw new Error(
          `${file} ${entry.phase_id}.depends_on must reference earlier phases`,
        );
      }
    }
  }

  return {
    path: file,
    phase_namespace: registry.phase_namespace,
    guide_path: registry.guide_path,
    phases: phases.sort(
      (left, right) =>
        Number(phaseNumber(left.phase_id)) -
        Number(phaseNumber(right.phase_id)),
    ),
  };
}

export function loadFrontendPhaseMap(root, phaseID) {
  const registry = loadFrontendPhaseRegistry(root);
  const entry = registry.phases.find(
    (candidate) => candidate.phase_id === phaseID,
  );
  if (!entry) {
    throw new Error(`unknown frontend phase ${phaseID}`);
  }
  const file = repoPath(root, entry.manifest_path);
  const manifest = readJsonObject(file, file);
  validateFrontendPhaseMap(manifest, file, phaseID);
  return { path: file, registryEntry: entry, manifest };
}

export function validateFrontendPhaseMap(
  manifest,
  label,
  expectedPhaseID = "",
) {
  assertObjectKeys(manifest, mapKeys, label);
  requireSchemaID(manifest, frontendPhaseTestMapSchemaID, label);
  if (manifest.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${label}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  const phaseID = requirePhaseID(manifest.phase_id, `${label}.phase_id`);
  if (expectedPhaseID && phaseID !== expectedPhaseID) {
    throw new Error(`${label}.phase_id must be ${expectedPhaseID}`);
  }
  const rows = requireObjectArray(manifest.rows, `${label}.rows`, {
    nonEmpty: true,
  });
  const ids = [];
  for (const [index, row] of rows.entries()) {
    const rowLabel = `${label}.rows[${index + 1}]`;
    assertObjectKeys(row, rowKeys, rowLabel);
    ids.push(
      requireString(row.id, `${rowLabel}.id`, { pattern: rowIDPattern }),
    );
    requireEnum(row.layer, `${rowLabel}.layer`, validLayers);
    requireEnum(
      row.evidence_class,
      `${rowLabel}.evidence_class`,
      validEvidenceClasses,
    );
    requireStringArray(row.owner_refs, `${rowLabel}.owner_refs`, {
      nonEmpty: true,
    });
    requireStringArray(row.core_req_ids, `${rowLabel}.core_req_ids`);
    requireStringArray(row.core_ac_ids, `${rowLabel}.core_ac_ids`);
    requireStringArray(
      row.support_or_design_ac_ids,
      `${rowLabel}.support_or_design_ac_ids`,
    );
    requireStringArray(row.targets, `${rowLabel}.targets`, { nonEmpty: true });
    requireStringArray(row.scenario_titles, `${rowLabel}.scenario_titles`);
    requireEnum(
      row.claim_status,
      `${rowLabel}.claim_status`,
      validClaimStatuses,
    );
    requireString(row.claim, `${rowLabel}.claim`);
    requireString(row.out_of_scope, `${rowLabel}.out_of_scope`);
    if (!row.id.includes(`-${phaseID.replace("FE-", "")}-`)) {
      throw new Error(`${rowLabel}.id must belong to ${phaseID}`);
    }
    if (
      row.targets.some((target) => target.startsWith("make browser-e2e")) &&
      row.scenario_titles.length === 0
    ) {
      throw new Error(
        `${rowLabel}.scenario_titles must be non-empty for browser-backed rows`,
      );
    }
    validateRowMetadata(row, rowLabel);
    validateCore03SortingFilteringGroupingOwnerRefs(row, rowLabel);
  }
  assertUnique(ids, `${label}.rows.id`);
}

export function validateFrontendPhaseArtifacts(root = process.cwd()) {
  const registry = loadFrontendPhaseRegistry(root);
  for (const entry of registry.phases) {
    if (!existsSync(repoPath(root, entry.manifest_path))) {
      throw new Error(`frontend phase map missing: ${entry.manifest_path}`);
    }
    const manifest = readJsonObject(
      repoPath(root, entry.manifest_path),
      entry.manifest_path,
    );
    validateFrontendPhaseMap(manifest, entry.manifest_path, entry.phase_id);
    if (
      entry.status === "active" &&
      manifest.rows.some((row) => row.claim_status === "blocked")
    ) {
      throw new Error(
        `${entry.phase_id} is active but contains blocked frontend rows`,
      );
    }
  }
  const mapDir = repoPath(root, "tools/frontend_phase_maps");
  for (const filename of readdirSync(mapDir).filter((name) =>
    name.endsWith(".json"),
  )) {
    const file = path.posix.join("tools/frontend_phase_maps", filename);
    const phaseID = phaseFromMapPath(file, file);
    if (!registry.phases.some((entry) => entry.phase_id === phaseID)) {
      throw new Error(`unregistered frontend phase map: ${file}`);
    }
  }
}

export function frontendLedgerOutputPath(entry) {
  return entry.ledger_path;
}

export function renderFrontendPhaseLedger(root, phaseID) {
  const { registryEntry, manifest } = loadFrontendPhaseMap(root, phaseID);
  const lines = [
    `# ${phaseID} Frontend Coverage Ledger`,
    "",
    `This ledger is generated from \`${registryEntry.manifest_path}\`. Update the frontend phase map first, then regenerate this file.`,
    "",
    `- Namespace: \`${frontendPhaseNamespace}\``,
    `- Status: \`${registryEntry.status}\``,
    `- Owner refs: ${registryEntry.owner_refs.map((owner) => `\`${owner}\``).join(", ")}`,
    `- Depends on: ${
      registryEntry.depends_on.length === 0
        ? "`none`"
        : registryEntry.depends_on.map((phase) => `\`${phase}\``).join(", ")
    }`,
    "- Authority: frontend phase maps are implementation-readiness inputs. This rendered ledger does not own product behavior.",
    "",
    "## Rows",
    "",
    "| Row | Layer | Evidence class | Claim status | Targets | Owner refs | Core REQs | Core ACs | Support/design ACs | Scenario titles | Claim | Out of scope |",
    "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |",
  ];

  for (const row of manifest.rows) {
    lines.push(
      `| \`${row.id}\` | \`${row.layer}\` | \`${row.evidence_class}\` | \`${row.claim_status}\` | ${row.targets.map((target) => `\`${target}\``).join("<br>")} | ${row.owner_refs.map((owner) => `\`${owner}\``).join("<br>")} | ${row.core_req_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.core_ac_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.support_or_design_ac_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.scenario_titles.map((title) => `\`${title}\``).join("<br>") || "`none`"} | ${row.claim} | ${row.out_of_scope} |`,
    );
  }

  return `${lines.join("\n")}\n`;
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const command = process.argv[2] ?? "";
  const root = process.cwd();
  if (command === "validate") {
    validateFrontendPhaseArtifacts(root);
    console.log("frontend phase artifacts verified");
  } else if (command === "phases") {
    for (const entry of loadFrontendPhaseRegistry(root).phases) {
      console.log(entry.phase_id);
    }
  } else {
    console.error("usage: frontend-phase-manifest.mjs validate|phases");
    process.exit(2);
  }
}
