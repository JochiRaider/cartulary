import path from "node:path";

import {
  assertObjectKeys,
  assertUnique,
  readJsonObject,
  requireEnum,
  requireInteger,
  requireObjectArray,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";
import {
  phaseFromLedgerPath,
  phaseFromMapPath,
  phaseNumber,
  repoPath,
  requirePhaseID,
} from "./common.mjs";
import {
  frontendPhaseNamespace,
  frontendPhaseRegistrySchemaID,
  registryEntryKeys,
  registryKeys,
  validRowRollupStates,
  validStatuses,
} from "./constants.mjs";
import { validateBlocker, validateOwnerRef } from "./owner-refs.mjs";
import { validateFrontendPhaseMap } from "./phase-map-validation.mjs";

export function frontendRegistryPath(root = process.cwd()) {
  return repoPath(root, "tools/frontend_phase_registry.json");
}

export function frontendVisualFixtureRegistryPath(root = process.cwd()) {
  return repoPath(root, "tools/frontend_visual_fixture_registry.json");
}

export function loadFrontendPhaseRegistry(root = process.cwd()) {
  const normalizedRoot = path.resolve(root);
  const file = frontendRegistryPath(normalizedRoot);
  const registry = readJsonObject(file, file);
  assertObjectKeys(registry, registryKeys, file);
  requireSchemaID(registry, frontendPhaseRegistrySchemaID, file);
  if (registry.phase_namespace !== frontendPhaseNamespace) {
    throw new Error(
      `${file}.phase_namespace must be ${frontendPhaseNamespace}`,
    );
  }
  requireInteger(registry.schema_version, `${file}.schema_version`, { min: 3 });
  requireRepoRelativePath(registry.guide_path, `${file}.guide_path`, {
    extension: ".md",
  });
  requireString(registry.guide_digest, `${file}.guide_digest`);

  const rawPhases = requireObjectArray(registry.phases, `${file}.phases`, {
    nonEmpty: true,
  });

  const phases = rawPhases.map((entry, index) => {
    const label = `${file}.phases[${index + 1}]`;
    assertObjectKeys(entry, registryEntryKeys, label);
    const phaseID = requirePhaseID(entry.phase_id, `${label}.phase_id`);
    const status = requireEnum(entry.status, `${label}.status`, validStatuses);
    const rowRollupState = requireEnum(
      entry.row_rollup_state,
      `${label}.row_rollup_state`,
      validRowRollupStates,
    );
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
      row_rollup_state: rowRollupState,
      manifest_path: manifestPath,
      manifest_digest: requireString(
        entry.manifest_digest,
        `${label}.manifest_digest`,
      ),
      ledger_path: ledgerPath,
      ledger_digest: requireString(entry.ledger_digest, `${label}.ledger_digest`),
      owner_refs: requireObjectArray(entry.owner_refs, `${label}.owner_refs`, {
        nonEmpty: true,
      }).map((ownerRef, ownerIndex) =>
        validateOwnerRef(
          normalizedRoot,
          ownerRef,
          `${label}.owner_refs[${ownerIndex + 1}]`,
          {
            claim_status: "implemented",
          },
        ),
      ),
      depends_on: requireStringArray(entry.depends_on, `${label}.depends_on`),
      activation_blockers: requireObjectArray(
        entry.activation_blockers,
        `${label}.activation_blockers`,
      ).map((blocker, blockerIndex) =>
        validateBlocker(
          blocker,
          `${label}.activation_blockers[${blockerIndex + 1}]`,
        ),
      ),
      evidence_freshness_digest: requireString(
        entry.evidence_freshness_digest,
        `${label}.evidence_freshness_digest`,
      ),
    };
  });

  assertUnique(
    phases.map((entry) => entry.phase_id),
    `${file}.phases.phase_id`,
  );
  const expected = Array.from(
    { length: phases.length },
    (_, index) => `FE-P${index}`,
  );
  const actual = phases
    .map((entry) => entry.phase_id)
    .sort(
      (left, right) =>
        Number(phaseNumber(left)) - Number(phaseNumber(right)),
    );
  if (actual.join(",") !== expected.join(",")) {
    throw new Error(
      `${file}.phases must contain contiguous frontend phases ${expected.join(", ")}`,
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
    guide_digest: registry.guide_digest,
    phases: phases.sort(
      (left, right) =>
        Number(phaseNumber(left.phase_id)) -
        Number(phaseNumber(right.phase_id)),
    ),
  };
}

export function loadFrontendPhaseMap(root, phaseID) {
  const normalizedRoot = path.resolve(root);
  const registry = loadFrontendPhaseRegistry(normalizedRoot);
  const entry = registry.phases.find(
    (candidate) => candidate.phase_id === phaseID,
  );
  if (!entry) {
    throw new Error(`unknown frontend phase ${phaseID}`);
  }
  const file = repoPath(normalizedRoot, entry.manifest_path);
  const manifest = readJsonObject(file, file);
  validateFrontendPhaseMap(manifest, file, phaseID, { root: normalizedRoot });
  return { path: file, registryEntry: entry, manifest };
}
