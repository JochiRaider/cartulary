import { readFileSync } from "node:fs";
import path from "node:path";

import { loadTestCatalog, targetForCatalogRow } from "../test-catalog/index.mjs";
import { parseStrictJSON, semanticJSONDigest } from "../test-catalog/semantic-json.mjs";

function asciiCompare(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function sortedUnique(values, label) {
  const sorted = [...values].sort(asciiCompare);
  for (let index = 1; index < sorted.length; index += 1) {
    if (sorted[index] === sorted[index - 1]) {
      throw new Error(`${label} contains duplicate ${sorted[index]}`);
    }
  }
  return sorted;
}

function sortedDistinct(values) {
  return [...new Set(values)].sort(asciiCompare);
}

function readStrictJSON(root, relativePath) {
  const file = path.join(root, relativePath);
  return parseStrictJSON(readFileSync(file, "utf8"), file);
}

function commandTargets(root) {
  const manifest = readStrictJSON(root, "tools/task_surface_manifest.json");
  return new Map(manifest.targets.map((entry) => [entry.command_id, entry.name]));
}

function profileByID(profiles, key, id) {
  const profile = profiles.semantic[key].find((entry) => entry.id === id);
  if (!profile) throw new Error(`catalog row references unresolved ${key} ${id}`);
  return profile;
}

export { targetForCatalogRow as evidenceTargetForCatalogRow };

export function loadOwnerAccountingSelection(
  root,
  { ownerID, rowIDs = null, targetName = "" },
) {
  const normalizedRoot = path.resolve(root);
  const catalog = loadTestCatalog(normalizedRoot);
  const owner = catalog.registry.owners.find((entry) => entry.owner_id === ownerID);
  if (!owner || owner.status !== "active") throw new Error(`unknown active test owner ${ownerID}`);

  const ownerRows = catalog.rows.filter((row) => row.owner_id === ownerID);
  if (ownerRows.length === 0) throw new Error(`active test owner ${ownerID} has no rows`);
  const requested = rowIDs === null
    ? null
    : sortedUnique(rowIDs.map((rowID) => String(rowID).trim()), "row selection");
  if (requested?.some((rowID) => rowID === "")) throw new Error("row selection contains a blank row ID");

  let rows = ownerRows;
  if (requested !== null) {
    if (requested.length === 0) throw new Error("row selection must not be empty");
    rows = requested.map((rowID) => {
      const row = catalog.rowByID.get(rowID);
      if (!row) throw new Error(`unknown test row ${rowID}`);
      if (row.owner_id !== ownerID) throw new Error(`test row ${rowID} does not belong to ${ownerID}`);
      return row;
    });
  }

  const targetByCommand = commandTargets(normalizedRoot);
  const rowsWithTargets = rows.map((row) => ({
    row,
    target_name: targetForCatalogRow(row, { commandTargetByID: targetByCommand }),
  }));
  if (targetName !== "") {
    const unsupported = rowsWithTargets.find((entry) => entry.target_name !== targetName);
    if (unsupported) {
      throw new Error(`test row ${unsupported.row.row_id} is not selected by evidence target ${targetName}`);
    }
  }
  rowsWithTargets.sort((left, right) => asciiCompare(left.row.row_id, right.row.row_id));

  const runtimeProfiles = sortedDistinct(rows.map((row) => row.runtime_profile_id))
    .map((id) => profileByID(catalog.profiles, "runtime_profiles", id));
  const resourceProfiles = sortedDistinct(rows.map((row) => row.resource_profile_id))
    .map((id) => profileByID(catalog.profiles, "resource_profiles", id));
  const fixtureProfiles = sortedDistinct(rows.map((row) => row.fixture_profile_id))
    .map((id) => profileByID(catalog.profiles, "fixture_profiles", id));

  const expectedRows = rowsWithTargets.map(({ row, target_name: target }) => ({
    row_id: row.row_id,
    owner_id: row.owner_id,
    family_id: row.family_id,
    verification_ids: [...row.verification_ids],
    runner: row.runner,
    selector_digest: semanticJSONDigest(row.selector),
    evidence_class: row.evidence_class,
    runtime_profile_id: row.runtime_profile_id,
    resource_profile_id: row.resource_profile_id,
    fixture_profile_id: row.fixture_profile_id,
    default_check: row.default_check,
    claim_posture: row.claim_posture,
    target_name: target,
  }));

  return {
    owner_id: ownerID,
    selected_rows: expectedRows.map((row) => row.row_id),
    catalog_semantic_digest: catalog.summary.catalog_semantic_digest,
    verification_semantic_digest: catalog.summary.verification_semantic_digest,
    runtime_profile_digest: semanticJSONDigest(runtimeProfiles),
    resource_profile_digest: semanticJSONDigest(resourceProfiles),
    fixture_profile_digest: semanticJSONDigest(fixtureProfiles),
    expected_rows: expectedRows,
  };
}

export function accountingRowsForTarget(root, { ownerID, rowIDs = null, targetName }) {
  const selection = loadOwnerAccountingSelection(root, { ownerID, rowIDs });
  const targetRowIDs = selection.expected_rows
    .filter((row) => row.target_name === targetName)
    .map((row) => row.row_id);
  return targetRowIDs.length === 0
    ? null
    : loadOwnerAccountingSelection(root, {
        ownerID,
        rowIDs: targetRowIDs,
        targetName,
      });
}
