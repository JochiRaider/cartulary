import { existsSync } from "node:fs";
import path from "node:path";

import { validExecutionDependencies } from "../scheduler/execution-dependencies.mjs";
import { assertObjectKeys, readJsonObject } from "../contract/json-shape.mjs";
import { validCoverage, validGoSections } from "./phase-manifest-constants.mjs";
import { phaseManifestRoot, phaseRegistryEntries } from "./phase-registry.mjs";
import { phaseNumberFromPhase } from "./phase-manifest-loader.mjs";

const phasePolicyExceptionsSchemaID = "cartulary.phase_policy_exceptions.v1";
const validPhasePolicyExceptionTypes = new Set(["allowed_empty_go_manifest_selection"]);
const phasePolicyExceptionKeys = new Set([
  "id",
  "type",
  "owner",
  "reason",
  "expires_before_phase",
  "expires_on",
  "selection",
]);
const emptyGoSelectionExceptionKeys = new Set([
  "phase",
  "section",
  "coverage",
  "execution_dependency",
  "package_patterns",
]);

function phasePolicyExceptionsPath(root) {
  if (process.env.CARTULARY_PHASE_POLICY_EXCEPTIONS) {
    return path.resolve(process.env.CARTULARY_PHASE_POLICY_EXCEPTIONS);
  }
  const manifestRoot = process.env.CARTULARY_PHASE_MANIFEST_ROOT
    ? phaseManifestRoot(root)
    : root;
  return path.join(manifestRoot, "tools", "phase_policy_exceptions.json");
}

function requireNonEmptyString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function maxKnownPhaseNumber(root) {
  const phaseNumbers = phaseRegistryEntries(root).map((entry) =>
    Number.parseInt(entry.phase.replace(/^phase/, ""), 10),
  );
  return phaseNumbers.length === 0 ? -1 : Math.max(...phaseNumbers);
}

function validatePolicyExceptionExpiration(root, entry, label) {
  const hasPhaseExpiration = entry.expires_before_phase !== undefined;
  const hasDateExpiration = entry.expires_on !== undefined;
  if (hasPhaseExpiration === hasDateExpiration) {
    throw new Error(`${label} must declare exactly one of expires_before_phase or expires_on`);
  }

  if (hasPhaseExpiration) {
    const expiresBeforePhase = requireNonEmptyString(
      entry.expires_before_phase,
      `${label}.expires_before_phase`,
    );
    const expiresBeforePhaseNumber = Number.parseInt(phaseNumberFromPhase(expiresBeforePhase), 10);
    if (maxKnownPhaseNumber(root) >= expiresBeforePhaseNumber) {
      throw new Error(`${label} expired before ${expiresBeforePhase}`);
    }
    return;
  }

  const expiresOn = requireNonEmptyString(entry.expires_on, `${label}.expires_on`);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(expiresOn)) {
    throw new Error(`${label}.expires_on must use YYYY-MM-DD`);
  }
  const expiry = Date.parse(`${expiresOn}T00:00:00Z`);
  if (!Number.isFinite(expiry)) {
    throw new Error(`${label}.expires_on must be a valid date`);
  }
  const today = process.env.CARTULARY_PHASE_POLICY_TODAY ?? "";
  const now = today === "" ? Date.now() : Date.parse(`${today}T00:00:00Z`);
  if (!Number.isFinite(now)) {
    throw new Error("CARTULARY_PHASE_POLICY_TODAY must use YYYY-MM-DD when set");
  }
  if (now >= expiry) {
    throw new Error(`${label} expired on ${expiresOn}`);
  }
}

function validateEmptyGoSelectionException(entry, label) {
  if (!entry.selection || typeof entry.selection !== "object" || Array.isArray(entry.selection)) {
    throw new Error(`${label}.selection must be an object`);
  }
  const selection = entry.selection;
  assertObjectKeys(selection, emptyGoSelectionExceptionKeys, `${label}.selection`);
  phaseNumberFromPhase(requireNonEmptyString(selection.phase, `${label}.selection.phase`));
  const section = requireNonEmptyString(selection.section, `${label}.selection.section`);
  if (!validGoSections.has(section)) {
    throw new Error(`${label}.selection.section must be unit|integration|e2e`);
  }
  const coverage = requireNonEmptyString(selection.coverage, `${label}.selection.coverage`);
  if (!validCoverage.has(coverage)) {
    throw new Error(`${label}.selection.coverage must be authoritative|supplemental`);
  }
  const executionDependency =
    selection.execution_dependency === undefined
      ? ""
      : String(selection.execution_dependency).trim();
  if (executionDependency !== "" && !validExecutionDependencies.has(executionDependency)) {
    throw new Error(
      `${label}.selection.execution_dependency has invalid value ${executionDependency}`,
    );
  }
  if (!Array.isArray(selection.package_patterns) || selection.package_patterns.length === 0) {
    throw new Error(`${label}.selection.package_patterns must be a non-empty array`);
  }
  for (const [index, pattern] of selection.package_patterns.entries()) {
    if (typeof pattern !== "string" || pattern.trim() === "") {
      throw new Error(`${label}.selection.package_patterns[${index}] must be a non-empty string`);
    }
  }
}

function validatePhasePolicyException(root, entry, index) {
  const label = `phase_policy_exceptions[${index + 1}]`;
  if (!entry || typeof entry !== "object" || Array.isArray(entry)) {
    throw new Error(`${label} must be an object`);
  }
  assertObjectKeys(entry, phasePolicyExceptionKeys, label);
  const id = requireNonEmptyString(entry.id, `${label}.id`);
  if (!/^[a-z][a-z0-9_.-]*$/.test(id)) {
    throw new Error(`${label}.id must be a lowercase identifier`);
  }
  const type = requireNonEmptyString(entry.type, `${label}.type`);
  if (!validPhasePolicyExceptionTypes.has(type)) {
    throw new Error(`${label}.type has unsupported value ${type}`);
  }
  requireNonEmptyString(entry.owner, `${label}.owner`);
  requireNonEmptyString(entry.reason, `${label}.reason`);
  validatePolicyExceptionExpiration(root, entry, label);

  if (type === "allowed_empty_go_manifest_selection") {
    validateEmptyGoSelectionException(entry, label);
  }
}

export function loadPhasePolicyExceptions(root) {
  const manifestPath = phasePolicyExceptionsPath(root);
  if (!existsSync(manifestPath)) {
    return {
      manifestPath,
      manifest: {
        schema_id: phasePolicyExceptionsSchemaID,
        exceptions: [],
      },
    };
  }

  const manifest = readJsonObject(manifestPath, manifestPath);
  assertObjectKeys(manifest, new Set(["schema_id", "exceptions"]), manifestPath);
  if (manifest.schema_id !== phasePolicyExceptionsSchemaID) {
    throw new Error(`${manifestPath} must declare schema_id ${phasePolicyExceptionsSchemaID}`);
  }
  if (!Array.isArray(manifest.exceptions)) {
    throw new Error(`${manifestPath} must declare exceptions[]`);
  }
  const seen = new Set();
  for (const [index, entry] of manifest.exceptions.entries()) {
    validatePhasePolicyException(root, entry, index);
    if (seen.has(entry.id)) {
      throw new Error(`${manifestPath} declares duplicate exception id ${entry.id}`);
    }
    seen.add(entry.id);
  }
  return { manifestPath, manifest };
}

function packagePatternsEqual(left, right) {
  if (left.length !== right.length) {
    return false;
  }
  return left.every((value, index) => value === right[index]);
}

export function emptyGoManifestSelectionAllowed(
  root,
  phase,
  section,
  coverage,
  executionDependency,
  packagePatterns,
) {
  if (packagePatterns.length === 0) {
    throw new Error("empty go manifest selection lookup requires at least one package pattern");
  }
  const { manifest } = loadPhasePolicyExceptions(root);
  return manifest.exceptions.some((entry) => {
    if (entry.type !== "allowed_empty_go_manifest_selection") {
      return false;
    }
    const selection = entry.selection;
    return (
      selection.phase === phase &&
      selection.section === section &&
      selection.coverage === coverage &&
      (selection.execution_dependency ?? "") === executionDependency &&
      packagePatternsEqual(selection.package_patterns, packagePatterns)
    );
  });
}
