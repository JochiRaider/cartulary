#!/usr/bin/env node
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { collectSourceTestTitleCounts } from "../test-catalog/selector-resolution.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const phaseArgument = process.argv.find((argument) => argument.startsWith("--phases="));
if (!phaseArgument) throw new Error("usage: build-frontend-owner-catalog.mjs --phases=<n[,n]|n-m>");

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(root, relativePath), "utf8"));
}

function writeJSON(relativePath, value) {
  writeFileSync(path.join(root, relativePath), `${JSON.stringify(value, null, 2)}\n`);
}

function phaseNumbers(value) {
  const selected = new Set();
  for (const token of value.split(",")) {
    const range = /^(\d+)-(\d+)$/u.exec(token);
    if (range) {
      const first = Number(range[1]);
      const last = Number(range[2]);
      if (first > last) throw new Error(`invalid descending phase range ${token}`);
      for (let phase = first; phase <= last; phase += 1) selected.add(phase);
    } else if (/^\d+$/u.test(token)) {
      selected.add(Number(token));
    } else {
      throw new Error(`invalid phase selector ${token}`);
    }
  }
  const result = [...selected].sort((left, right) => left - right);
  if (result.length === 0 || result.some((phase) => phase < 0 || phase > 12)) {
    throw new Error("phase selectors must be within 0..12");
  }
  return result;
}

const selectedPhases = phaseNumbers(phaseArgument.slice("--phases=".length));
const selectedSourcePaths = new Set(
  selectedPhases.map((phase) => `tools/frontend_phase_maps/fe_p${phase}_test_map.json`),
);
const reviewRevision = execFileSync("git", ["rev-parse", "--short=12", "HEAD"], {
  cwd: root,
  encoding: "utf8",
}).trim();

function sourceFiles(directory, pattern) {
  const result = [];
  for (const entry of readdirSync(path.join(root, directory), { withFileTypes: true })) {
    const relative = path.posix.join(directory, entry.name);
    if (entry.isDirectory()) result.push(...sourceFiles(relative, pattern));
    else if (entry.isFile() && pattern.test(entry.name)) result.push(relative);
  }
  return result.sort(asciiCompare);
}

const runnerFiles = {
  playwright: sourceFiles("apps/web/e2e", /\.spec\.tsx?$/u),
  vitest: [
    ...sourceFiles("apps/web", /\.test\.tsx?$/u),
    ...sourceFiles("packages", /\.test\.tsx?$/u),
  ].sort(asciiCompare),
};
const titleIndex = new Map();

function titlesForFile(runner, file) {
  const key = `${runner}\0${file}`;
  if (!titleIndex.has(key)) {
    titleIndex.set(
      key,
      collectSourceTestTitleCounts({
        root,
        file,
        approvedRoots: runner === "playwright" ? ["apps/web/e2e"] : ["apps/web", "packages"],
      }),
    );
  }
  return titleIndex.get(key);
}

function resolveTitleFiles(runner, title) {
  const matches = runnerFiles[runner].filter((file) => (titlesForFile(runner, file).get(title) ?? 0) === 1);
  if (matches.length === 0) {
    throw new Error(`${runner} title ${JSON.stringify(title)} does not resolve in any source file`);
  }
  return matches;
}

function rowRunner(row) {
  const targets = row.targets.map((target) => target.target_name);
  if (targets.some((target) => target.startsWith("browser-e2e"))) return "playwright";
  if ((row.scenario_titles ?? []).length > 0) return "vitest";
  return "shell";
}

function primaryOwner(phase, row) {
  if (phase === 0) {
    return {
      "FE-U-P0-01": "package.protocol_ts",
      "FE-U-P0-02": "package.view_contracts",
      "FE-U-P0-03": "package.ui",
      "FE-S-P0-01": "harness.generated_artifacts",
      "FE-S-P0-02": "web.architecture",
      "FE-S-P0-04": "harness.generated_artifacts",
    }[row.id];
  }
  if (phase === 1) return "module.auth";
  if (phase === 2) return row.id === "FE-E-P2-01" ? "module.savedviews" : "module.workbook";
  if (phase === 3) {
    if (["FE-U-P3-01", "FE-U-P3-02", "FE-U-P3-03", "FE-U-P3-04", "FE-V-P3-01", "FE-A11Y-P3-01"].includes(row.id)) {
      return "package.grid_adapter";
    }
    if (row.id === "FE-S-P3-01") return "web.architecture";
    return "module.workbook";
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
  if (phase === 9) {
    if (["FE-I-P9-01", "FE-E-P9-01"].includes(row.id)) return "module.revisions";
    return "module.workbook";
  }
  if (phase === 10) return "module.workbook";
  if (phase === 11) return "web.design";
  if (phase === 12) return "module.networkflow";
  throw new Error(`no frontend owner for phase ${phase} row ${row.id}`);
}

function webOwnerForFile(file) {
  if (file.startsWith("packages/grid-adapter/")) return "package.grid_adapter";
  if (file.startsWith("packages/protocol-ts/")) return "package.protocol_ts";
  if (file.startsWith("packages/test-utils/")) return "package.test_utils";
  if (file.startsWith("packages/ui-contracts/")) return "package.ui";
  if (file.startsWith("packages/view-contracts/")) return "package.view_contracts";
  if (file.includes("/networkFlow/") || file.includes("network-flow")) return "web.networkflow";
  if (file.includes("collaboration") || file.includes("session-recovery")) return "web.collaboration";
  if (file.includes("/workbook/") || file.includes("Workbook") || file.includes("workbook.")) return "web.workbook";
  return "web.application";
}

function collaborators(ownerID, runner, file, evidenceClass) {
  if (runner === "shell") return [];
  const result = new Set([webOwnerForFile(file)]);
  if (runner === "playwright") result.add("harness.browser");
  if (["visual", "accessibility"].includes(evidenceClass)) result.add("web.design");
  if (evidenceClass === "visual") result.add("harness.visual");
  result.delete(ownerID);
  return [...result].sort(asciiCompare);
}

function semanticSlug(value) {
  return value
    .replace(/([a-z0-9])([A-Z])/gu, "$1 $2")
    .replace(/\b(?:phase|sprint)\s*[-_]?\d+(?!\d)/giu, " ")
    .replace(/\bfe[-_](?:a11y|u|i|e|v|b|s)[-_]p[-_]?\d+(?:[-_][a-z0-9]+)*(?![a-z0-9])/giu, " ")
    .replace(/\bfe[-_]p[-_]?\d+(?:[-_]\d+)?(?!\d)/giu, " ")
    .replace(/[^a-zA-Z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "")
    .toLowerCase()
    .replace(/_+/gu, "_") || "behavior";
}

function rowSegment(sourceKey, statement, file) {
  const stem = semanticSlug(statement || path.basename(file)).slice(0, 48).replace(/_+$/u, "");
  const digest = createHash("sha256").update(sourceKey).digest("hex").slice(0, 10);
  return `${stem || "behavior"}_${digest}`;
}

function stageFor(row, file) {
  const targets = new Set(row.targets.map((target) => target.target_name));
  if (targets.has("browser-e2e-measurement")) return "measurement";
  if (targets.has("browser-e2e-a11y")) return "accessibility";
  if (targets.has("browser-e2e-visual")) return "visual";
  if (targets.has("browser-e2e-support") && file.includes(".support.spec.")) return "support";
  if (targets.has("browser-e2e-stateful")) return "stateful";
  return "webserver_backed";
}

function evidenceClass(row, runner, file) {
  if (runner === "shell") return "static";
  if (runner === "vitest") return "unit";
  const stage = stageFor(row, file);
  if (stage === "accessibility") return "accessibility";
  if (stage === "visual") return "visual";
  if (stage === "measurement") return "measurement";
  return "browser";
}

function familySegment(row, runner, evidence, file) {
  if (runner === "shell") return "support";
  if (evidence === "accessibility") return "accessibility";
  if (evidence === "visual") return "visual";
  if (evidence === "measurement") return "measurement";
  if (runner === "playwright") {
    const stage = stageFor(row, file);
    if (stage === "support") return "browser_support";
    return stage === "stateful" ? "browser_stateful" : "browser";
  }
  return {
    integration: "frontend_integration",
    unit: "frontend_unit",
  }[row.layer];
}

function catalogTitleRow({ phase, legacy, ownerID, runner, file, titles, sourceKey }) {
  const evidence = evidenceClass(legacy, runner, file);
  const familyID = `${ownerID}.${familySegment(legacy, runner, evidence, file)}`;
  const stage = runner === "playwright" ? stageFor(legacy, file) : null;
  const selector = runner === "vitest"
    ? { file, titles }
    : {
        file,
        project_id: "chromium",
        stage,
        scenario_ids: titles
          .map((title) => `scenario_${createHash("sha256").update(`${file}\0${title}`).digest("hex").slice(0, 12)}`)
          .sort(asciiCompare),
        titles,
      };
  return {
    row_id: `${familyID}.${rowSegment(sourceKey, legacy.claim?.statement, file)}`,
    owner_id: ownerID,
    family_id: familyID,
    collaborator_ids: collaborators(ownerID, runner, file, evidence),
    verification_ids: [`${ownerID}.verification.behavior_contract`],
    runner,
    selector,
    evidence_class: evidence,
    runtime_profile_id: runner === "playwright"
      ? ownerID === "module.networkflow" ? "network_flow_claimed" : "default"
      : "none",
    resource_profile_id: runner === "playwright" ? "browser_exclusive" : "none",
    fixture_profile_id: runner === "playwright" ? "service_stack" : "none",
    default_check: ["visual", "accessibility", "measurement"].includes(evidence)
      ? false
      : legacy.default_check_required,
    claim_posture: ["visual", "accessibility", "measurement"].includes(evidence)
      ? "informative"
      : "implementation",
    status: "active",
  };
}

const commandIDs = Object.fromEntries(
  readJSON("tools/task_surface_manifest.json").targets.map((entry) => [entry.name, entry.command_id]),
);

function shellCommands(row) {
  return {
    "FE-S-P0-01": ["generated-artifact-policy-check", "generate-drift"],
    "FE-S-P0-02": ["frontend-import-boundary-check"],
    "FE-S-P0-04": ["generate-drift"],
    "FE-S-P3-01": ["frontend-import-boundary-check"],
  }[row.id] ?? [];
}

function catalogShellRow({ legacy, ownerID, target, sourceKey }) {
  const familyID = `${ownerID}.support`;
  return {
    row_id: `${familyID}.${rowSegment(sourceKey, legacy.claim?.statement, target)}`,
    owner_id: ownerID,
    family_id: familyID,
    collaborator_ids: target === "generate-drift" ? ["web.design"] : [],
    verification_ids: [`${ownerID}.verification.behavior_contract`],
    runner: "shell",
    selector: { command_id: commandIDs[target] },
    evidence_class: "static",
    runtime_profile_id: "none",
    resource_profile_id: "none",
    fixture_profile_id: "none",
    default_check: legacy.default_check_required,
    claim_posture: "implementation",
    status: "active",
  };
}

function verificationDefinition(ownerID, evidenceKinds) {
  const profile = ownerID === "module.networkflow"
    ? "extension.network_flow_activity"
    : ownerID === "module.reporting"
      ? "extension.snapshot_reporting"
      : ownerID === "module.reference_data"
        ? "extension.reference_pack"
        : ownerID.startsWith("module.")
          ? "base"
          : "support";
  return {
    verification_id: `${ownerID}.verification.behavior_contract`,
    behavior_class: ownerID === "module.auth"
      ? "security"
      : ownerID.startsWith("module.")
        ? "product"
        : ownerID.startsWith("harness.")
          ? "harness"
          : "architecture",
    profile,
    requirement: `${ownerID} frontend postconditions remain exact across retained selectors and declared execution profiles.`,
    evidence_kinds: evidenceKinds,
    status: "active",
  };
}

const baseline = readJSON("tools/test_migration_baseline.json");
const crosswalk = readJSON("tools/test_migration_crosswalk.json");
const selectedLegacyIDs = new Set(
  baseline.identities
    .filter((entry) => entry.source_registry_id === "frontend_phase_maps" && selectedSourcePaths.has(entry.source_path))
    .map((entry) => entry.legacy_row_id),
);
const selectedDispositions = crosswalk.dispositions.filter(
  (entry) => entry.source_registry_id === "frontend_phase_maps" && selectedLegacyIDs.has(entry.legacy_row_id),
);
const selectedNewRows = crosswalk.new_rows.filter((entry) =>
  [...selectedLegacyIDs].some((legacyID) => entry.provenance.startsWith(`frontend:${legacyID} `)),
);
const ownedRowIDs = new Set([
  ...selectedDispositions.filter((entry) => entry.disposition === "migrated").map((entry) => entry.row_id),
  ...selectedNewRows.map((entry) => entry.row_id),
]);

const manifests = new Map();
for (const file of readdirSync(path.join(root, "tools/test_families")).filter((name) => name.endsWith(".json"))) {
  const manifest = readJSON(`tools/test_families/${file}`);
  manifests.set(manifest.owner_id, manifest.rows.filter((row) => !ownedRowIDs.has(row.row_id)));
}
const catalogRows = () => [...manifests.values()].flat();
const existingTitleRow = (runner, file, title) => catalogRows().find(
  (row) => row.runner === runner && row.selector.file === file && (row.selector.titles ?? []).includes(title),
);
const existingShellRow = (commandID) => catalogRows().find(
  (row) => row.runner === "shell" && row.selector.command_id === commandID,
);

const dispositions = crosswalk.dispositions.filter(
  (entry) => entry.source_registry_id !== "frontend_phase_maps" || !selectedLegacyIDs.has(entry.legacy_row_id),
);
const newRows = crosswalk.new_rows.filter(
  (entry) => ![...selectedLegacyIDs].some((legacyID) => entry.provenance.startsWith(`frontend:${legacyID} `)),
);
const generatedRows = [];

function appendRow(row) {
  if (!manifests.has(row.owner_id)) manifests.set(row.owner_id, []);
  manifests.get(row.owner_id).push(row);
  generatedRows.push(row);
}

function review(ownerID, legacy, sourcePath) {
  return {
    governing_owner: ownerID,
    governing_requirement: legacy.claim?.statement ?? `Retained frontend behavior for ${legacy.id}`,
    review_revision: reviewRevision,
    owner_review_evidence: `WS-05 ${sourcePath} owner-slice adjudication`,
  };
}

function migratedDisposition(legacy, row, sourcePath) {
  return {
    source_registry_id: "frontend_phase_maps",
    legacy_row_id: legacy.id,
    disposition: "migrated",
    owner_id: row.owner_id,
    row_id: row.row_id,
    verification_ids: row.verification_ids,
    adjudication_rule: "normative_postcondition",
    collaborator_ids: row.collaborator_ids,
    review: review(row.owner_id, legacy, sourcePath),
  };
}

function consolidatedDisposition(legacy, rows, sourcePath) {
  const sorted = [...new Map(rows.map((row) => [row.row_id, row])).values()]
    .sort((left, right) => asciiCompare(left.row_id, right.row_id));
  const survivor = sorted[0];
  return {
    source_registry_id: "frontend_phase_maps",
    legacy_row_id: legacy.id,
    disposition: "consolidated",
    owner_id: survivor.owner_id,
    row_id: survivor.row_id,
    verification_ids: survivor.verification_ids,
    adjudication_rule: "normative_postcondition",
    collaborator_ids: survivor.collaborator_ids,
    review: review(survivor.owner_id, legacy, sourcePath),
    surviving_row_id: survivor.row_id,
    assertion_preservation_evidence: `All exact legacy selectors are retained by existing owner rows: ${sorted.map((row) => row.row_id).join(", ")}`,
  };
}

for (const phase of selectedPhases) {
  const sourcePath = `tools/frontend_phase_maps/fe_p${phase}_test_map.json`;
  const source = readJSON(sourcePath);
  for (const legacy of source.rows) {
    if (!selectedLegacyIDs.has(legacy.id)) throw new Error(`unfrozen frontend row ${sourcePath}:${legacy.id}`);
    if (["FE-S-P0-03", "FE-S-P11-01", "FE-S-P11-02", "FE-S-P11-03"].includes(legacy.id)) {
      dispositions.push({
        source_registry_id: "frontend_phase_maps",
        legacy_row_id: legacy.id,
        disposition: "deleted",
        reason: legacy.id === "FE-S-P0-03" ? "obsolete" : "duplicate",
        review: review("harness.test_catalog", legacy, sourcePath),
      });
      continue;
    }

    const ownerID = primaryOwner(phase, legacy);
    if (!ownerID) throw new Error(`missing owner for ${legacy.id}`);
    const runner = rowRunner(legacy);
    const existingRows = [];
    const createdRows = [];
    if (runner === "shell") {
      const targets = shellCommands(legacy);
      if (targets.length === 0) throw new Error(`unsupported target-only frontend row ${legacy.id}`);
      for (const [index, target] of targets.entries()) {
        const commandID = commandIDs[target];
        if (!commandID) throw new Error(`unknown shell target ${target}`);
        const existing = existingShellRow(commandID);
        if (existing) {
          existingRows.push(existing);
          continue;
        }
        const row = catalogShellRow({
          legacy,
          ownerID,
          target,
          sourceKey: `${sourcePath}\0${legacy.id}\0${target}`,
        });
        appendRow(row);
        createdRows.push(row);
        if (index > 0) {
          newRows.push({
            owner_id: row.owner_id,
            row_id: row.row_id,
            verification_ids: row.verification_ids,
            provenance: `frontend:${legacy.id} split target ${target}`,
          });
        }
      }
    } else {
      const groups = new Map();
      for (const title of [...new Set(legacy.scenario_titles)].sort(asciiCompare)) {
        for (const file of resolveTitleFiles(runner, title)) {
          const existing = existingTitleRow(runner, file, title);
          if (existing) {
            existingRows.push(existing);
            continue;
          }
          if (!groups.has(file)) groups.set(file, []);
          groups.get(file).push(title);
        }
      }
      for (const [index, [file, titles]] of [...groups.entries()].sort(([left], [right]) => asciiCompare(left, right)).entries()) {
        const row = catalogTitleRow({
          phase,
          legacy,
          ownerID,
          runner,
          file,
          titles: titles.sort(asciiCompare),
          sourceKey: `${sourcePath}\0${legacy.id}\0${file}`,
        });
        appendRow(row);
        createdRows.push(row);
        if (index > 0) {
          newRows.push({
            owner_id: row.owner_id,
            row_id: row.row_id,
            verification_ids: row.verification_ids,
            provenance: `frontend:${legacy.id} split file ${file}`,
          });
        }
      }
    }

    if (createdRows.length > 0) {
      dispositions.push(migratedDisposition(legacy, createdRows[0], sourcePath));
      for (const [index, row] of createdRows.entries()) {
        if (index === 0) continue;
        if (!newRows.some((entry) => entry.row_id === row.row_id)) {
          newRows.push({
            owner_id: row.owner_id,
            row_id: row.row_id,
            verification_ids: row.verification_ids,
            provenance: `frontend:${legacy.id} split selector group`,
          });
        }
      }
    } else if (existingRows.length > 0) {
      dispositions.push(consolidatedDisposition(legacy, existingRows, sourcePath));
    } else {
      throw new Error(`${legacy.id} produced no catalog or consolidation row`);
    }
  }
}

for (const [ownerID, rows] of [...manifests]) {
  const unique = new Map();
  for (const row of rows) {
    if (unique.has(row.row_id)) throw new Error(`duplicate frontend row ${row.row_id}`);
    unique.set(row.row_id, row);
  }
  manifests.set(ownerID, [...unique.values()].sort((left, right) => asciiCompare(left.row_id, right.row_id)));
}

const evidenceByOwner = new Map();
for (const row of generatedRows) {
  if (!evidenceByOwner.has(row.owner_id)) evidenceByOwner.set(row.owner_id, new Set());
  evidenceByOwner.get(row.owner_id).add({ playwright: "playwright", shell: "shell", vitest: "vitest" }[row.runner]);
}
const verificationRegistry = readJSON("contracts/verification/registry.json");
const verificationOwners = new Map(verificationRegistry.owners.map((entry) => [entry.owner_id, entry]));
for (const [ownerID, evidenceKinds] of evidenceByOwner) {
  const contractPath = `contracts/verification/owners/${ownerID}.json`;
  const contract = readJSON(contractPath);
  const previous = contract.verifications.find(
    (entry) => entry.verification_id === `${ownerID}.verification.behavior_contract`,
  );
  for (const kind of previous?.evidence_kinds ?? []) evidenceKinds.add(kind);
  contract.verifications = contract.verifications.filter(
    (entry) => entry.verification_id !== `${ownerID}.verification.behavior_contract`,
  );
  contract.verifications.push(previous
    ? { ...previous, evidence_kinds: [...evidenceKinds].sort(asciiCompare) }
    : verificationDefinition(ownerID, [...evidenceKinds].sort(asciiCompare)));
  contract.verifications.sort((left, right) => asciiCompare(left.verification_id, right.verification_id));
  writeJSON(contractPath, contract);
  verificationOwners.set(ownerID, { owner_id: ownerID, contract_path: contractPath, status: "active" });
}
verificationRegistry.owners = [...verificationOwners.values()].sort((left, right) => asciiCompare(left.owner_id, right.owner_id));
writeJSON("contracts/verification/registry.json", verificationRegistry);

for (const [ownerID, rows] of manifests) {
  if (rows.length === 0) throw new Error(`owner ${ownerID} would have zero rows`);
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

crosswalk.pending_baseline_keys = crosswalk.pending_baseline_keys.filter(
  (entry) => entry.source_registry_id !== "frontend_phase_maps" || !selectedLegacyIDs.has(entry.legacy_row_id),
);
crosswalk.dispositions = dispositions.sort((left, right) =>
  asciiCompare(`${left.source_registry_id}:${left.legacy_row_id}`, `${right.source_registry_id}:${right.legacy_row_id}`),
);
crosswalk.new_rows = newRows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
writeJSON("tools/test_migration_crosswalk.json", crosswalk);

process.stdout.write(`${JSON.stringify({
  phases: selectedPhases,
  baseline_rows: selectedLegacyIDs.size,
  generated_rows: generatedRows.length,
  consolidated_rows: dispositions.filter(
    (entry) => entry.source_registry_id === "frontend_phase_maps" && selectedLegacyIDs.has(entry.legacy_row_id) && entry.disposition === "consolidated",
  ).length,
  deleted_rows: dispositions.filter(
    (entry) => entry.source_registry_id === "frontend_phase_maps" && selectedLegacyIDs.has(entry.legacy_row_id) && entry.disposition === "deleted",
  ).length,
  catalog_rows: [...manifests.values()].reduce((sum, rows) => sum + rows.length, 0),
  pending_authoritative: crosswalk.pending_baseline_keys.length,
})}\n`);
