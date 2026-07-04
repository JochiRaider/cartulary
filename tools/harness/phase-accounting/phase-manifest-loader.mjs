import path from "node:path";

import { readJsonObject } from "../contract/json-shape.mjs";
import {
  activePhaseRegistryEntries,
  activePhaseRegistryEntry,
  manifestPhaseRegistryEntries,
  phaseManifestRoot,
  phaseRegistryEntries,
  phaseRegistryEntry,
  retiredPhaseStatus,
} from "./phase-registry.mjs";
import { validatePhaseManifestShape } from "./phase-manifest-shape.mjs";

export function phaseNumberFromPhase(phase) {
  const match = /^phase(0|[1-9]\d*)$/.exec(phase);
  if (!match) {
    throw new Error(`invalid phase name ${phase}; expected phase0 or phase[1-9][0-9]*`);
  }
  return match[1];
}

function phaseFromManifestFilename(manifestPath) {
  const filename = path.basename(manifestPath);
  const match = /^(phase(?:0|[1-9]\d*))_test_map\.json$/.exec(filename);
  if (!match) {
    throw new Error(
      `phase test map filename ${filename} must match phase0_test_map.json or phase[1-9][0-9]*_test_map.json`,
    );
  }
  return match[1];
}

function validateManifestIdentity(manifestPath, manifest, requestedPhase = "") {
  if (!manifest || typeof manifest !== "object" || Array.isArray(manifest)) {
    throw new Error(`manifest ${manifestPath} must be a JSON object`);
  }
  validatePhaseManifestShape(manifest, `manifest ${manifestPath}`);
  phaseNumberFromPhase(manifest.phase);
  const filenamePhase = phaseFromManifestFilename(manifestPath);
  if (manifest.phase !== filenamePhase) {
    throw new Error(
      `manifest ${manifestPath} declares phase ${manifest.phase} but filename declares ${filenamePhase}`,
    );
  }
  if (requestedPhase !== "" && manifest.phase !== requestedPhase) {
    throw new Error(
      `manifest ${manifestPath} declares phase ${manifest.phase} but was requested as ${requestedPhase}`,
    );
  }
  return manifest.phase;
}

export function loadManifest(root, phase, { allowPlanned = false } = {}) {
  phaseNumberFromPhase(phase);
  const manifestRoot = phaseManifestRoot(root);
  const registryEntry = allowPlanned ? phaseRegistryEntry(root, phase) : activePhaseRegistryEntry(root, phase);
  if (!registryEntry) {
    const known = activePhaseRegistryEntries(root).map((entry) => entry.phase);
    const registered = phaseRegistryEntries(root).find((entry) => entry.phase === phase);
    const inactiveStatus = registered ? ` (${registered.status})` : "";
    throw new Error(
      `unknown active phase ${phase}${inactiveStatus}; expected one of ${known.join(", ") || "none"}`,
    );
  }
  if (allowPlanned && registryEntry.status === retiredPhaseStatus) {
    throw new Error(`phase ${phase} is retired and has no executable manifest`);
  }
  const manifestPath = path.join(manifestRoot, registryEntry.manifest_path);
  const manifest = readJsonObject(manifestPath, manifestPath);
  validateManifestIdentity(manifestPath, manifest, phase);
  return { manifestPath, manifest, registryEntry };
}

export function phaseManifestNames(root, { includePlanned = false } = {}) {
  const entries = includePlanned
    ? manifestPhaseRegistryEntries(root)
    : activePhaseRegistryEntries(root);
  return entries.map((entry) => entry.phase);
}
