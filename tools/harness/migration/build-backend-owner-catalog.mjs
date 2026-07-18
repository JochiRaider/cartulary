#!/usr/bin/env node
import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const phaseArgument = process.argv.find((argument) => argument.startsWith("--phases="));
if (!phaseArgument) {
  throw new Error("usage: build-backend-owner-catalog.mjs --phases=<n[,n]|n-m>");
}

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function readJSON(relativePath) {
  return JSON.parse(readFileSync(path.join(root, relativePath), "utf8"));
}

function writeJSON(relativePath, value) {
  writeFileSync(path.join(root, relativePath), `${JSON.stringify(value, null, 2)}\n`);
}

function canonical(value) {
  if (Array.isArray(value)) return `[${value.map(canonical).join(",")}]`;
  if (value && typeof value === "object") {
    return `{${Object.keys(value).sort(asciiCompare).map((key) => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

function semanticDigest(value) {
  return `sha256:${createHash("sha256").update(canonical(value)).digest("hex")}`;
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
      continue;
    }
    if (!/^\d+$/u.test(token)) throw new Error(`invalid phase selector ${token}`);
    selected.add(Number(token));
  }
  const result = [...selected].sort((left, right) => left - right);
  if (result.length === 0 || result.some((phase) => phase < 0 || phase > 12)) {
    throw new Error("phase selectors must be within 0..12");
  }
  return result;
}

const selectedPhases = phaseNumbers(phaseArgument.slice("--phases=".length));
const selectedSourcePaths = new Set(selectedPhases.map((phase) => `tools/phase${phase}_test_map.json`));
const reviewRevision = execFileSync("git", ["rev-parse", "--short=12", "HEAD"], {
  cwd: root,
  encoding: "utf8",
}).trim();

function packageOwner(packagePath) {
  if (packagePath === "./internal/app/serverprocess") return "app.server";
  const match = /^\.\/internal\/(app|modules|platform)\/([^/]+)/u.exec(packagePath);
  if (!match) throw new Error(`no owner mapping for package ${packagePath}`);
  const namespace = match[1] === "modules" ? "module" : match[1];
  return `${namespace}.${match[2]}`;
}

function nonGoOwner(phase, row) {
  if (phase === 1) {
    return ["E-1-09", "E-1-10"].includes(row.id) ? "module.incidents" : "module.auth";
  }
  if (phase === 2) return "module.incidents";
  if (phase === 3) return "module.timeline";
  if (phase === 4) {
    if (["U-4-WB-01", "U-4-WB-02", "E-4-05"].includes(row.id)) return "module.assessments";
    if (["U-4-WB-03", "V-4-GRID-02"].includes(row.id)) return "module.evidence";
    if (["U-4-WB-04", "U-4-WB-05", "E-4-06", "V-4-GRID-03"].includes(row.id)) {
      return "module.workbook";
    }
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
  throw new Error(`no non-Go owner mapping for phase ${phase} row ${row.id}`);
}

function ownerFor(phase, row) {
  return (row.runner ?? "go_test") === "go_test"
    ? packageOwner(row.package)
    : nonGoOwner(phase, row);
}

function webOwner(file) {
  if (file.includes("/networkFlow/") || file.includes("network-flow")) return "web.networkflow";
  if (file.includes("collaboration") || file.includes("session-recovery")) return "web.collaboration";
  if (file.includes("/workbook/") || file.includes("Workbook") || file.includes("workbook.")) {
    return "web.workbook";
  }
  return "web.application";
}

function collaboratorsFor(row, ownerID) {
  if ((row.runner ?? "go_test") === "go_test") return [];
  const collaborators = new Set([webOwner(row.file)]);
  if (row.runner === "playwright") collaborators.add("harness.browser");
  if (row.execution_dependency === "browser_visual") {
    collaborators.add("harness.visual");
    collaborators.add("web.design");
  }
  collaborators.delete(ownerID);
  return [...collaborators].sort(asciiCompare);
}

function semanticSlug(value) {
  return value
    .replace(/([a-z0-9])([A-Z])/gu, "$1 $2")
    .replace(/\b(?:phase|sprint)\s*[-_]?\d+(?!\d)/giu, " ")
    .replace(/\bfe[-_]p[-_]?\d+(?:[-_]\d+)?(?!\d)/giu, " ")
    .replace(/\b(?:fe[-_])?(?:u|i|e|v|b|s)(?:[-_][a-z]+)*[-_]p?\d+(?:[-_][a-z0-9]+)*(?![a-z0-9])/giu, " ")
    .replace(/\bnf[-_]ac[-_]\d+(?!\d)/giu, " ")
    .replace(/\b(?:test|support)\b/giu, " ")
    .replace(/[^a-zA-Z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "")
    .toLowerCase()
    .replace(/_+/gu, "_") || "behavior";
}

function rowSegment(sourceKey, row, support = false) {
  const selectorText = row.symbol ?? row.symbols?.[0] ?? row.title ?? row.titles?.[0] ?? row.file;
  const claimText = row.claim || selectorText;
  const stem = semanticSlug(`${support ? "support " : ""}${claimText}`).slice(0, 48).replace(/_+$/u, "");
  const digest = createHash("sha256").update(sourceKey).digest("hex").slice(0, 10);
  return `${stem || "behavior"}_${digest}`;
}

function familySegment(row, support = false) {
  if (support) {
    return row.layer === "backend_integration_support" ? "support_integration" : "support_unit";
  }
  return {
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
}

function fixtureProfile(row, support = false) {
  if (!support && row.runner === "playwright") return "service_stack";
  return {
    group_clone: "postgres_group_clone",
    migration_scratch: "postgres_migration_scratch",
    template_clone: "postgres_template_clone",
    transaction: "postgres_transaction",
  }[row.fixture_policy?.postgres] ?? "none";
}

function runtimeProfile(row, fixtureID, ownerID) {
  if (row.runner === "playwright") {
    if (ownerID === "module.networkflow" && row.id !== "E-12-NFAC001-01") {
      return "network_flow_claimed";
    }
    return row.runtime_profile_id ?? row.runtime_profile ?? "default";
  }
  if (row.execution_dependency === "backend_process" || fixtureID !== "none") return "default";
  return "none";
}

function resourceProfile(row, fixtureID) {
  if (row.runner === "playwright") return "browser_exclusive";
  if (row.runner === "vitest") return "none";
  if (fixtureID === "postgres_transaction") return "go_transaction_heavy";
  if (["postgres_group_clone", "postgres_template_clone"].includes(fixtureID)) return "go_clone_heavy";
  if (fixtureID === "postgres_migration_scratch") return "go_io_heavy";
  if (["backend_process", "backend_integration"].includes(row.execution_dependency)) return "go_io_heavy";
  return "go_balanced";
}

function evidenceClass(row, support = false) {
  if (support) return row.layer === "backend_integration_support" ? "integration" : "unit";
  return {
    browser_measurement: "measurement",
    browser_visual: "visual",
    playwright: "browser",
    vitest: "unit",
  }[row.execution_dependency] ?? {
    playwright: "browser",
    vitest: "unit",
  }[row.runner] ?? (row.execution_dependency === "backend_unit" ? "unit" : "integration");
}

function playwrightStage(dependency) {
  return {
    browser_functional: "webserver_backed",
    browser_measurement: "measurement",
    browser_stateful: "stateful",
    browser_support: "support",
    browser_visual: "visual",
  }[dependency];
}

function selectorFor(row) {
  if ((row.runner ?? "go_test") === "go_test") {
    return {
      package: row.package,
      tests: [...new Set([...(row.symbols ?? []), ...(row.symbol ? [row.symbol] : [])])].sort(asciiCompare),
    };
  }
  const titles = [...new Set([...(row.titles ?? []), ...(row.title ? [row.title] : [])])].sort(asciiCompare);
  if (row.runner === "vitest") return { file: row.file, titles };
  return {
    file: row.file,
    project_id: "chromium",
    stage: playwrightStage(row.execution_dependency),
    scenario_ids: titles
      .map((title) => `scenario_${createHash("sha256").update(`${row.file}\0${title}`).digest("hex").slice(0, 12)}`)
      .sort(asciiCompare),
    titles,
  };
}

function catalogRow(phase, row, sourceKey, support = false) {
  const ownerID = ownerFor(phase, row);
  const familyID = `${ownerID}.${familySegment(row, support)}`;
  const verificationID = `${ownerID}.verification.behavior_contract`;
  const fixtureID = fixtureProfile(row, support);
  const evidence = evidenceClass(row, support);
  return {
    row_id: `${familyID}.${rowSegment(sourceKey, row, support)}`,
    owner_id: ownerID,
    family_id: familyID,
    collaborator_ids: collaboratorsFor(row, ownerID),
    verification_ids: [verificationID],
    runner: row.runner === "go_test" || row.runner === undefined ? "go" : row.runner,
    selector: selectorFor(row),
    evidence_class: evidence,
    runtime_profile_id: runtimeProfile(row, fixtureID, ownerID),
    resource_profile_id: resourceProfile(row, fixtureID),
    fixture_profile_id: fixtureID,
    default_check: support ? false : evidence === "visual" || evidence === "measurement" ? false : row.default_check_required,
    claim_posture: ["visual", "measurement"].includes(evidence) ? "informative" : "implementation",
    status: "active",
  };
}

const ownerLabels = Object.freeze({
  "app.operator": "operator application composition",
  "app.server": "server application composition and process lifecycle",
  "module.assessments": "assessment creation, history, filtering, and presentation",
  "module.auth": "authentication, session, credential, and authorization",
  "module.collaboration": "live collaboration, presence, conflict, and recovery",
  "module.entities": "entity, mention, resolution, and merge",
  "module.evidence": "evidence lifecycle, access, linking, and presentation",
  "module.imports": "import lifecycle and validation",
  "module.incidentbundles": "incident bundle export and import",
  "module.incidents": "incident identity, membership, lifecycle, and discovery",
  "module.indicators": "indicator lifecycle and projection",
  "module.jobapi": "job API lifecycle and cancellation",
  "module.links": "record link lifecycle and projection",
  "module.networkflow": "Network Flow extension",
  "module.parties": "party identity, references, and workbook behavior",
  "module.recovery": "backup, restore, recovery, and integrity",
  "module.reference_data": "Reference Pack lifecycle and administration",
  "module.reporting": "snapshot reporting",
  "module.revisions": "revision history, conflict tokens, rollback, and restore",
  "module.savedviews": "saved views, query state, and startup selection",
  "module.tabularingest": "tabular ingest validation and execution",
  "module.tasksdecisions": "task and decision lifecycle and relationships",
  "module.timeline": "Timeline record, mutation, query, and workbook behavior",
  "module.workbook": "workbook registry, mutation, query, and interaction",
  "platform.bootstrap": "bootstrap orchestration",
  "platform.config": "configuration discovery and validation",
  "platform.httpruntime": "HTTP runtime decoding, errors, and safety boundaries",
  "platform.objectstore": "object-store lifecycle and access boundaries",
  "platform.postgres": "PostgreSQL bootstrap, migrations, and fixture boundaries",
  "platform.viewquery": "view-query parsing and execution boundaries",
  "platform.viewschema": "view-schema registry and validation boundaries",
  "platform.ws": "WebSocket runtime, protocol, and lifecycle boundaries",
});

function verificationDefinition(ownerID, evidenceKinds) {
  const moduleProfile = ownerID === "module.networkflow"
    ? "extension.network_flow_activity"
    : ownerID === "module.reporting"
      ? "extension.snapshot_reporting"
      : ownerID === "module.reference_data"
        ? "extension.reference_pack"
        : "base";
  return {
    verification_id: `${ownerID}.verification.behavior_contract`,
    behavior_class: ownerID === "module.auth"
      ? "security"
      : ownerID.startsWith("module.")
        ? "product"
        : "architecture",
    profile: ownerID.startsWith("module.") ? moduleProfile : "base",
    requirement: `${ownerLabels[ownerID] ?? ownerID} postconditions remain exact across every retained executable selector and declared fixture/runtime boundary.`,
    evidence_kinds: evidenceKinds,
    status: "active",
  };
}

const baseline = readJSON("tools/test_migration_baseline.json");
const crosswalk = readJSON("tools/test_migration_crosswalk.json");
for (let phase = 0; phase <= 12; phase += 1) {
  const sourcePath = `tools/phase${phase}_test_map.json`;
  const source = readJSON(sourcePath);
  for (const legacy of [
    ...(source.unit ?? []),
    ...(source.integration ?? []),
    ...(source.e2e ?? []),
    ...(source.visual ?? []),
  ].filter((row) => row.runner === "vitest")) {
    const frozen = baseline.identities.find(
      (entry) => entry.source_registry_id === "backend_phase_maps" && entry.legacy_row_id === legacy.id,
    );
    if (!frozen || frozen.source_path !== sourcePath) {
      throw new Error(`cannot repair frozen Vitest identity ${sourcePath}:${legacy.id}`);
    }
    const selector = {
      file: legacy.file,
      titles: [...new Set([...(legacy.titles ?? []), ...(legacy.title ? [legacy.title] : [])])].sort(asciiCompare),
    };
    frozen.runner = "vitest";
    frozen.selector = selector;
    frozen.selector_digest = semanticDigest(selector);
  }
}
crosswalk.baseline_digest = semanticDigest(baseline);
writeJSON("tools/test_migration_baseline.json", baseline);
const baselineByKey = new Map(
  baseline.identities.map((entry) => [`${entry.source_registry_id}\0${entry.legacy_row_id}`, entry]),
);
const supportBySource = new Map();
for (const candidate of baseline.auxiliary_candidates.backend_support) {
  if (!supportBySource.has(candidate.source_path)) supportBySource.set(candidate.source_path, []);
  supportBySource.get(candidate.source_path).push(candidate);
}

const selectedLegacyKeys = new Set(
  baseline.identities
    .filter((entry) => entry.source_registry_id === "backend_phase_maps" && selectedSourcePaths.has(entry.source_path))
    .map((entry) => `${entry.source_registry_id}\0${entry.legacy_row_id}`),
);
const selectedSupportIDs = new Set(
  baseline.auxiliary_candidates.backend_support
    .filter((entry) => selectedSourcePaths.has(entry.source_path))
    .map((entry) => entry.candidate_id),
);
const replacedRowIDs = new Set([
  ...crosswalk.dispositions
    .filter((entry) => selectedLegacyKeys.has(`${entry.source_registry_id}\0${entry.legacy_row_id}`) && entry.row_id)
    .map((entry) => entry.row_id),
  ...crosswalk.new_rows
    .filter((entry) => [...selectedSupportIDs].some((candidateID) => entry.provenance.startsWith(`${candidateID} `)))
    .map((entry) => entry.row_id),
]);

const manifests = new Map();
for (const file of readdirSync(path.join(root, "tools/test_families")).filter((name) => name.endsWith(".json"))) {
  const manifest = readJSON(`tools/test_families/${file}`);
  manifests.set(manifest.owner_id, manifest.rows.filter((row) => !replacedRowIDs.has(row.row_id)));
}

const dispositions = crosswalk.dispositions.filter(
  (entry) => !selectedLegacyKeys.has(`${entry.source_registry_id}\0${entry.legacy_row_id}`),
);
const auxiliaryDispositions = crosswalk.auxiliary_dispositions.filter(
  (entry) => !selectedSupportIDs.has(entry.candidate_id),
);
const newRows = crosswalk.new_rows.filter(
  (entry) => ![...selectedSupportIDs].some((candidateID) => entry.provenance.startsWith(`${candidateID} `)),
);
const generatedRows = [];

function appendRow(row) {
  if (!manifests.has(row.owner_id)) manifests.set(row.owner_id, []);
  manifests.get(row.owner_id).push(row);
  generatedRows.push(row);
}

for (const phase of selectedPhases) {
  const sourcePath = `tools/phase${phase}_test_map.json`;
  const manifest = readJSON(sourcePath);
  for (const group of ["unit", "integration", "e2e", "visual"]) {
    for (const legacy of manifest[group] ?? []) {
      const key = `backend_phase_maps\0${legacy.id}`;
      const frozen = baselineByKey.get(key);
      if (!frozen || frozen.source_path !== sourcePath) {
        throw new Error(`frozen backend identity is missing or inconsistent: ${sourcePath}:${legacy.id}`);
      }
      const row = catalogRow(phase, legacy, `${sourcePath}\0${legacy.id}`);
      appendRow(row);
      dispositions.push({
        source_registry_id: "backend_phase_maps",
        legacy_row_id: legacy.id,
        disposition: "migrated",
        owner_id: row.owner_id,
        row_id: row.row_id,
        verification_ids: row.verification_ids,
        adjudication_rule: "normative_postcondition",
        collaborator_ids: row.collaborator_ids,
        review: {
          governing_owner: row.owner_id,
          governing_requirement: legacy.claim || `Exact executable postconditions retained from ${legacy.id}`,
          review_revision: reviewRevision,
          owner_review_evidence: `WS-04 ${sourcePath} owner-slice adjudication`,
        },
      });
    }
  }
  const frozenSupport = supportBySource.get(sourcePath) ?? [];
  for (const [index, support] of (manifest.support_go_targets ?? []).entries()) {
    const candidate = frozenSupport[index];
    if (!candidate) throw new Error(`missing frozen support candidate ${sourcePath}:${index + 1}`);
    const row = catalogRow(phase, support, candidate.candidate_id, true);
    appendRow(row);
    auxiliaryDispositions.push({
      candidate_id: candidate.candidate_id,
      disposition: "retained",
      reason: `Retained as exact owner-catalog row ${row.row_id}; support remains explicit-only and owner-accountable.`,
      owner_review_evidence: `WS-04 ${sourcePath} support adjudication`,
    });
    newRows.push({
      owner_id: row.owner_id,
      row_id: row.row_id,
      verification_ids: row.verification_ids,
      provenance: `${candidate.candidate_id} frozen backend support candidate; exact package/test selectors retained`,
    });
  }
  if (frozenSupport.length !== (manifest.support_go_targets ?? []).length) {
    throw new Error(`${sourcePath} support candidate cardinality drift`);
  }
}

for (const [ownerID, rows] of [...manifests]) {
  const unique = new Map();
  for (const row of rows) {
    if (unique.has(row.row_id)) throw new Error(`duplicate generated row ${row.row_id}`);
    unique.set(row.row_id, row);
  }
  manifests.set(ownerID, [...unique.values()].sort((left, right) => asciiCompare(left.row_id, right.row_id)));
}

const evidenceByOwner = new Map();
for (const row of generatedRows) {
  if (!evidenceByOwner.has(row.owner_id)) evidenceByOwner.set(row.owner_id, new Set());
  evidenceByOwner.get(row.owner_id).add({ go: "go_test", playwright: "playwright", vitest: "vitest" }[row.runner]);
}
const verificationRegistry = readJSON("contracts/verification/registry.json");
const verificationOwners = new Map(verificationRegistry.owners.map((entry) => [entry.owner_id, entry]));
for (const [ownerID, evidenceKinds] of evidenceByOwner) {
  const contractPath = `contracts/verification/owners/${ownerID}.json`;
  const contract = existsSync(path.join(root, contractPath))
    ? readJSON(contractPath)
    : { schema_id: "cartulary.verification_contract.v1", owner_id: ownerID, verifications: [] };
  const previousBehavior = contract.verifications.find(
    (entry) => entry.verification_id === `${ownerID}.verification.behavior_contract`,
  );
  for (const evidenceKind of previousBehavior?.evidence_kinds ?? []) evidenceKinds.add(evidenceKind);
  contract.verifications = contract.verifications.filter(
    (entry) => entry.verification_id !== `${ownerID}.verification.behavior_contract`,
  );
  contract.verifications.push(verificationDefinition(ownerID, [...evidenceKinds].sort(asciiCompare)));
  contract.verifications.sort((left, right) => asciiCompare(left.verification_id, right.verification_id));
  writeJSON(contractPath, contract);
  verificationOwners.set(ownerID, { owner_id: ownerID, contract_path: contractPath, status: "active" });
}
verificationRegistry.owners = [...verificationOwners.values()].sort((left, right) => asciiCompare(left.owner_id, right.owner_id));
writeJSON("contracts/verification/registry.json", verificationRegistry);

for (const [ownerID, rows] of manifests) {
  if (rows.length === 0) throw new Error(`owner ${ownerID} would have zero catalog rows`);
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
  (entry) => !selectedLegacyKeys.has(`${entry.source_registry_id}\0${entry.legacy_row_id}`),
);
crosswalk.dispositions = dispositions.sort((left, right) =>
  asciiCompare(`${left.source_registry_id}:${left.legacy_row_id}`, `${right.source_registry_id}:${right.legacy_row_id}`),
);
crosswalk.auxiliary_dispositions = auxiliaryDispositions.sort((left, right) => asciiCompare(left.candidate_id, right.candidate_id));
crosswalk.new_rows = newRows.sort((left, right) => asciiCompare(left.row_id, right.row_id));
writeJSON("tools/test_migration_crosswalk.json", crosswalk);

process.stdout.write(`${JSON.stringify({
  phases: selectedPhases,
  authoritative_rows: generatedRows.length - selectedSupportIDs.size,
  support_rows: selectedSupportIDs.size,
  catalog_owners: ownerIDs.length,
  catalog_rows: [...manifests.values()].reduce((sum, rows) => sum + rows.length, 0),
  pending_authoritative: crosswalk.pending_baseline_keys.length,
})}\n`);
