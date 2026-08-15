import { readFileSync } from "node:fs";
import path from "node:path";

import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import { parseStrictJSON } from "../contract/index.mjs";

const ownerIDPattern = /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}$/u;

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sortedCounts(values) {
  const counts = new Map();
  for (const value of values) counts.set(value, (counts.get(value) ?? 0) + 1);
  return Object.fromEntries([...counts.entries()].sort(([left], [right]) => asciiCompare(left, right)));
}

function commandTargetMap(root) {
  const manifest = parseStrictJSON(
    readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"),
    "tools/task_surface_manifest.json",
  );
  return new Map(manifest.targets.map((entry) => [entry.command_id, entry.name]));
}

function requireOwner(catalog, ownerID) {
  const normalized = String(ownerID ?? "").trim();
  if (!ownerIDPattern.test(normalized)) throw new Error("OWNER is missing or malformed");
  const owner = catalog.registry.owners.find((entry) => entry.owner_id === normalized && entry.status === "active");
  if (!owner) throw new Error(`unknown active test owner ${normalized}`);
  return owner;
}

function ownerProjection(root, ownerID) {
  const catalog = loadTestCatalog(root);
  const owner = requireOwner(catalog, ownerID);
  const rows = catalog.rows
    .filter((row) => row.owner_id === owner.owner_id)
    .sort((left, right) => asciiCompare(left.row_id, right.row_id));
  if (rows.length === 0) throw new Error(`active test owner ${owner.owner_id} has no rows`);
  const targetByCommand = commandTargetMap(root);
  const targetRows = rows.map((row) => ({
    row,
    target: targetForCatalogRow(row, { commandTargetByID: targetByCommand }),
  }));
  const serviceRows = rows.filter((row) => row.service_dependencies.length > 0);
  return { catalog, owner, rows, targetRows, serviceRows };
}

export function explainTestOwner(root, ownerID) {
  const projection = ownerProjection(root, ownerID);
  const familyIDs = [...new Set(projection.rows.map((row) => row.family_id))].sort(asciiCompare);
  return {
    schema_id: "cartulary.test_owner_explanation.v3",
    evidence_epoch: projection.catalog.summary.evidence_epoch,
    owner_id: projection.owner.owner_id,
    manifest_path: projection.owner.manifest_path,
    test_catalog_digest: projection.catalog.summary.test_catalog_digest,
    verification_routing_digest: projection.catalog.summary.verification_routing_digest,
    row_count: projection.rows.length,
    service_backed_row_count: projection.serviceRows.length,
    families: familyIDs.map((familyID) => ({
      family_id: familyID,
      row_count: projection.rows.filter((row) => row.family_id === familyID).length,
    })),
    runner_counts: sortedCounts(projection.rows.map((row) => row.runner)),
    evidence_counts: sortedCounts(projection.rows.map((row) => row.evidence_class)),
    runtime_profile_counts: sortedCounts(projection.rows.map((row) => row.runtime_profile_id)),
    resource_profile_counts: sortedCounts(projection.rows.map((row) => row.resource_profile_id)),
    fixture_capability_counts: sortedCounts(projection.rows.map((row) => row.fixture_capability)),
    target_counts: sortedCounts(projection.targetRows.map((entry) => entry.target)),
    minimum_tier_counts: sortedCounts(projection.rows.map((row) => row.minimum_tier)),
    commands: {
      full_owner: `make test-slice OWNER=${projection.owner.owner_id}`,
      service_backed: projection.serviceRows.length > 0
        ? `make service-backed-test-slice OWNER=${projection.owner.owner_id}`
        : null,
      exact_row_template: `make test-slice OWNER=${projection.owner.owner_id} ROWS=<row-id,...>`,
    },
  };
}

function broaderCommands(projection) {
  const commands = new Set();
  for (const { row, target } of projection.targetRows) {
    if (row.runner === "shell" && target) commands.add(`make ${target}`);
    if (["unit", "integration"].includes(row.evidence_class)) commands.add("make test-fast");
    if (row.runner === "playwright" && target) commands.add(`make ${target}`);
    if (row.evidence_class === "security") commands.add("make go-gosec-targeted");
    if (row.evidence_class === "release") commands.add("make release-check");
  }
  return [...commands].sort(asciiCompare);
}

export function buildModuleAuthorTaskGuide(root, ownerID, role) {
  if (role !== "module-author") throw new Error("ROLE must be exact module-author");
  const projection = ownerProjection(root, ownerID);
  const evidenceClasses = new Set(projection.rows.map((row) => row.evidence_class));
  const generatedCommands = [];
  if (
    projection.owner.owner_id === "harness.generated_artifacts" ||
    projection.rows.some((row) => row.runner === "shell" && row.selector.command_id.includes("generate"))
  ) {
    generatedCommands.push("make generate-drift", "make generated-artifact-policy-check");
  }
  const releaseRequired = evidenceClasses.has("release");
  return {
    schema_id: "cartulary.task_guide_summary.v3",
    evidence_epoch: projection.catalog.summary.evidence_epoch,
    role,
    owner_id: projection.owner.owner_id,
    test_catalog_digest: projection.catalog.summary.test_catalog_digest,
    verification_routing_digest: projection.catalog.summary.verification_routing_digest,
    focused_commands: [
      `make test-slice OWNER=${projection.owner.owner_id}`,
      ...(projection.serviceRows.length > 0
        ? [`make service-backed-test-slice OWNER=${projection.owner.owner_id}`]
        : []),
    ],
    generated_commands: generatedCommands.sort(asciiCompare),
    broader_commands: broaderCommands(projection),
    release_gate: releaseRequired ? "make release-check" : null,
  };
}
