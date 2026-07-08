import { existsSync, readdirSync, readFileSync } from "node:fs";
import path from "node:path";

import {
  loadTaskSurfaceManifest,
  targetEntryMap,
} from "../../generated-artifacts/index.mjs";
import { readJsonObject } from "../../contract/json-shape.mjs";
import {
  phaseFromMapPath,
  phaseNumber,
  repoPath,
} from "./common.mjs";
import { frontendEvidenceFreshnessDigest, sha256File } from "./freshness.mjs";
import { collectFrontendGuideTargetRestatementErrors } from "./guide-restatements.mjs";
import {
  loadFrontendPhaseRegistry,
} from "./registry-loader.mjs";
import { validateFrontendPhaseMap } from "./phase-map-validation.mjs";
import { validateFrontendBrowserScenarioTitleOwnership } from "./row-validation.mjs";
import { validateFrontendVisualFixtureRegistry } from "./visual-fixture-registry.mjs";

function validateFrontendGuideTargetRestatements(root, registry, rowTargetNames) {
  const absoluteGuidePath = repoPath(root, registry.guide_path);
  if (!existsSync(absoluteGuidePath)) {
    throw new Error(`frontend guide missing: ${registry.guide_path}`);
  }
  const errors = collectFrontendGuideTargetRestatementErrors(
    readFileSync(absoluteGuidePath, "utf8"),
    rowTargetNames,
    registry.guide_path,
  );
  if (errors.length > 0) {
    throw new Error(
      `frontend guide target restatement drift:\n${errors.join("\n")}`,
    );
  }
}

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
  const targetEntriesByName =
    options.targetEntriesByName ??
    targetEntryMap(
      loadTaskSurfaceManifest(
        path.join(root, "tools", "task_surface_manifest.json"),
      ).manifest,
    );
  const phaseStates = new Map();
  const rowTargetNames = new Map();
  const frontendBrowserTitleOwners = new Map();
  const expectedGuideDigest = sha256File(root, registry.guide_path);
  if (
    checkFreshness &&
    expectedGuideDigest &&
    registry.guide_digest !== expectedGuideDigest
  ) {
    throw new Error(
      `${registry.path}.guide_digest must match ${registry.guide_path}`,
    );
  }
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
      rowTargetNames.set(
        row.id,
        new Set(row.targets.map((target) => target.target_name)),
      );
    }
    if (
      checkFreshness &&
      expectedGuideDigest &&
      manifest.guide_digest !== expectedGuideDigest
    ) {
      throw new Error(
        `${entry.manifest_path}.guide_digest must match ${registry.guide_path}`,
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
    const expectedLedgerDigest = sha256File(root, entry.ledger_path);
    if (
      checkFreshness &&
      expectedLedgerDigest &&
      entry.ledger_digest !== expectedLedgerDigest
    ) {
      throw new Error(
        `${entry.phase_id}.ledger_digest must match ${entry.ledger_path}`,
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
  validateFrontendGuideTargetRestatements(root, registry, rowTargetNames);
  validateFrontendVisualFixtureRegistry(root);
}
