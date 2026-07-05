import {
  defaultReasonRequiredLayers,
  validRuntimeBinaries,
} from "./phase-manifest-constants.mjs";
import {
  validBackendEvidenceClasses,
  validBackendLayers,
  validDefaultCheckKinds,
  validDefaultCheckReasonCodes,
  validWarmLocalCostClasses,
} from "./phase-manifest-shape.mjs";

export function validateExecutionFamily(entry, label) {
  if (typeof entry.execution_family !== "string" || entry.execution_family.trim() === "") {
    throw new Error(`${label} must declare execution_family`);
  }
  if (!/^[a-z][a-z0-9-]*$/.test(entry.execution_family)) {
    throw new Error(`${label} execution_family must be a lowercase hyphenated identifier`);
  }
  if (typeof entry.execution_label !== "string" || entry.execution_label.trim() === "") {
    throw new Error(`${label} must declare execution_label`);
  }
}

export function runtimeBinaries(entry, label) {
  if (entry.runtime_binaries === undefined) {
    return [];
  }
  if (!Array.isArray(entry.runtime_binaries) || entry.runtime_binaries.length === 0) {
    throw new Error(`${label} runtime_binaries must be a non-empty string array when present`);
  }
  const seen = new Set();
  const result = [];
  for (const [index, raw] of entry.runtime_binaries.entries()) {
    if (typeof raw !== "string" || raw.trim() === "") {
      throw new Error(`${label} runtime_binaries[${index + 1}] must be a non-empty string`);
    }
    const id = raw.trim();
    if (!validRuntimeBinaries.has(id)) {
      throw new Error(`${label} runtime_binaries[${index + 1}] has unknown runtime binary ${id}`);
    }
    if (seen.has(id)) {
      throw new Error(`${label} runtime_binaries contains duplicate ${id}`);
    }
    seen.add(id);
    result.push(id);
  }
  return result;
}

export function validateEvidencePlacement(entry, label) {
  if (typeof entry.evidence_class !== "string" || !validBackendEvidenceClasses.has(entry.evidence_class)) {
    throw new Error(`${label} must declare closed evidence_class`);
  }
  if (typeof entry.layer !== "string" || !validBackendLayers.has(entry.layer)) {
    throw new Error(`${label} must declare closed layer`);
  }
  if (typeof entry.default_check_required !== "boolean") {
    throw new Error(`${label} must declare default_check_required as a boolean`);
  }
  if (
    typeof entry.default_check_kind !== "string" ||
    !validDefaultCheckKinds.has(entry.default_check_kind)
  ) {
    throw new Error(`${label} must declare closed default_check_kind`);
  }
  if (
    typeof entry.default_check_reason_code !== "string" ||
    !validDefaultCheckReasonCodes.has(entry.default_check_reason_code)
  ) {
    throw new Error(`${label} must declare closed default_check_reason_code`);
  }
  if (
    typeof entry.primary_evidence_owner !== "string" ||
    entry.primary_evidence_owner.trim() === ""
  ) {
    throw new Error(`${label} must declare primary_evidence_owner`);
  }
  if (
    entry.duplicate_of !== null &&
    (typeof entry.duplicate_of !== "string" || entry.duplicate_of.trim() === "")
  ) {
    throw new Error(`${label} must declare duplicate_of as null or a non-empty string`);
  }
  if (typeof entry.evidence_delta !== "string" || entry.evidence_delta.trim() === "") {
    throw new Error(`${label} must declare evidence_delta`);
  }
  if (
    typeof entry.warm_local_cost_class !== "string" ||
    !validWarmLocalCostClasses.has(entry.warm_local_cost_class)
  ) {
    throw new Error(`${label} must declare closed warm_local_cost_class`);
  }
  if (entry.default_check_required === true && entry.default_check_kind === "explicit_only") {
    throw new Error(`${label} default_check_required=true cannot use default_check_kind=explicit_only`);
  }
  if (entry.default_check_required === false && entry.default_check_kind === "primary_local_evidence") {
    throw new Error(`${label} default_check_required=false cannot use primary_local_evidence`);
  }
  const requiresReason =
    entry.default_check_required === true &&
    (entry.evidence_class !== "product_conformance" ||
      defaultReasonRequiredLayers.has(entry.layer));
  if (requiresReason) {
    if (typeof entry.default_check_reason !== "string" || entry.default_check_reason.trim() === "") {
      throw new Error(`${label} must declare default_check_reason for non-obvious default check inclusion`);
    }
  }
}
