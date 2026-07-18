import { existsSync, readdirSync } from "node:fs";
import path from "node:path";

import { targetEntryMap } from "../../generated-artifacts/task-surface/model.mjs";
import { readJsonObject } from "../../contract/json-shape.mjs";
import {
  phaseFromMapPath,
  phaseNumber,
  repoPath,
} from "./common.mjs";
import { frontendEvidenceFreshnessDigest, sha256File } from "./freshness.mjs";
import {
  loadFrontendPhaseRegistry,
} from "./registry-loader.mjs";
import { validateFrontendPhaseMap } from "./phase-map-validation.mjs";
import { validateFrontendBrowserScenarioTitleOwnership } from "./row-validation.mjs";
import { validateFrontendVisualFixtureRegistry } from "./visual-fixture-registry.mjs";

function computeRowRollupState(entry, manifest, priorPhaseStates) {
  const implementedRows = manifest.rows.filter(
    (row) => row.claim_status === "implemented",
  );
  if (implementedRows.length !== manifest.rows.length) {
    throw new Error(
      `${entry.phase_id} current frontend maps must contain only implemented rows`,
    );
  }
  const dependenciesGreen = entry.depends_on.every((phaseID) =>
    priorPhaseStates.get(phaseID) === "active_green",
  );
  if (!dependenciesGreen) {
    throw new Error(`${entry.phase_id} dependencies must be active_green`);
  }
  return "active_green";
}

export function validateFrontendPhaseArtifacts(root = process.cwd(), options = {}) {
  const checkFreshness = options.checkFreshness !== false;
  const registry = loadFrontendPhaseRegistry(root);
  const taskSurfaceManifestPath = path.join(
    root,
    "tools",
    "task_surface_manifest.json",
  );
  const targetEntriesByName =
    options.targetEntriesByName ??
    targetEntryMap(readJsonObject(taskSurfaceManifestPath, taskSurfaceManifestPath));
  const phaseStates = new Map();
  const frontendBrowserTitleOwners = new Map();
  for (const entry of registry.phases) {
    if (!existsSync(repoPath(root, entry.manifest_path))) {
      throw new Error(`frontend phase map missing: ${entry.manifest_path}`);
    }
    const manifest = readJsonObject(
      repoPath(root, entry.manifest_path),
      entry.manifest_path,
    );
    validateFrontendPhaseMap(manifest, entry.manifest_path, entry.phase_id, {
      root,
      targetEntriesByName,
      frontendBrowserTitleOwners,
    });
    for (const row of manifest.rows) {
      validateFrontendBrowserScenarioTitleOwnership(
        root,
        row,
        `${entry.manifest_path}.rows.${row.id}`,
        frontendBrowserTitleOwners,
      );
    }
    const expectedManifestDigest = sha256File(root, entry.manifest_path);
    if (
      checkFreshness &&
      expectedManifestDigest &&
      entry.manifest_digest !== expectedManifestDigest
    ) {
      throw new Error(
        `${entry.phase_id}.manifest_digest must match ${entry.manifest_path}`,
      );
    }
    const expectedFreshnessDigest = frontendEvidenceFreshnessDigest(
      root,
      registry,
      entry,
    );
    if (
      checkFreshness &&
      entry.evidence_freshness_digest !== expectedFreshnessDigest
    ) {
      throw new Error(
        `${entry.phase_id}.evidence_freshness_digest must match frontend freshness inputs`,
      );
    }
    const rowRollupState = computeRowRollupState(entry, manifest, phaseStates);
    phaseStates.set(entry.phase_id, rowRollupState);
    if (entry.row_rollup_state !== rowRollupState) {
      throw new Error(
        `${entry.phase_id}.row_rollup_state must be ${rowRollupState}, got ${entry.row_rollup_state}`,
      );
    }
    if (entry.status === "active" && rowRollupState !== "active_green") {
      throw new Error(
        `${entry.phase_id} active phases must have row_rollup_state=active_green`,
      );
    }
    if (entry.activation_blockers.length !== 0) {
      throw new Error(
        `${entry.phase_id} active frontend phases must not declare activation_blockers[]`,
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
  validateFrontendVisualFixtureRegistry(root);
}
