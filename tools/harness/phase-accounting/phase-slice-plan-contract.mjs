import {
  assertKnownResource,
  compareResources,
} from "../scheduler/scheduler-resources.mjs";

export const phaseSlicePlanSchemaID = "cartulary.phase_slice_plan.v1";

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requireArray(value, label) {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value;
}

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export function resourceLimitObject(resourceLimits) {
  if (resourceLimits && typeof resourceLimits.entries === "function") {
    return Object.fromEntries(
      Array.from(resourceLimits.entries()).sort((left, right) =>
        compareResources(left[0], right[0]),
      ),
    );
  }
  return Object.fromEntries(
    Object.entries(requireObject(resourceLimits ?? {}, "resource_limits")).sort(
      ([left], [right]) => compareResources(left, right),
    ),
  );
}

export function serializePhaseSliceWorkUnit(unit) {
  const { resourceClaims, weightMs: _weightMs, ...rest } = unit;
  return {
    ...rest,
    weight_ms: unit.weightMs,
    resource_claims: resourceLimitObject(resourceClaims ?? new Map()),
  };
}

function validateResourceAmount(value, label, { allowAuto = false } = {}) {
  if (Number.isInteger(value) && value >= 1) {
    return;
  }
  if (value === "limit" || (allowAuto && value === "auto")) {
    return;
  }
  throw new Error(
    `${label} must be a positive integer${allowAuto ? ', "auto"' : ""}, or "limit"`,
  );
}

function validateResourceMap(value, label, { limits = null, allowAuto = false } = {}) {
  const entries = Object.entries(requireObject(value, label));
  for (const [resource, amount] of entries) {
    assertKnownResource(resource, `${label}.${resource}`, {
      scheduler: "phase_slice",
    });
    validateResourceAmount(amount, `${label}.${resource}`, { allowAuto });
    if (limits && !Object.hasOwn(limits, resource)) {
      throw new Error(`${label}.${resource} is not declared in resource_limits`);
    }
  }
  return entries;
}

function unitCompletionKeys(unit) {
  if (Array.isArray(unit.completionKeys) && unit.completionKeys.length > 0) {
    return unit.completionKeys;
  }
  if (Array.isArray(unit.completion_keys) && unit.completion_keys.length > 0) {
    return unit.completion_keys;
  }
  return [unit.id];
}

function validateStringArray(value, label) {
  const entries = requireArray(value ?? [], label);
  return entries.map((entry, index) => requireString(entry, `${label}[${index + 1}]`));
}

function validateOptionalString(value, label) {
  if (value === undefined) {
    return;
  }
  requireString(value, label);
}

function validateOptionalBoolean(value, label) {
  if (value === undefined || typeof value === "boolean") {
    return;
  }
  throw new Error(`${label} must be a boolean`);
}

function validateOptionalNonnegativeInteger(value, label) {
  if (value === undefined || (Number.isInteger(value) && value >= 0)) {
    return;
  }
  throw new Error(`${label} must be a nonnegative integer`);
}

export function validatePhaseSlicePlanContract(plan, label = "phase-slice plan") {
  requireObject(plan, label);
  if (plan.schema_id !== phaseSlicePlanSchemaID) {
    throw new Error(`${label}.schema_id must be ${phaseSlicePlanSchemaID}`);
  }
  if (!["phase-slice", "service-backed-slice"].includes(plan.target)) {
    throw new Error(`${label}.target must be phase-slice or service-backed-slice`);
  }
  if (!["phase", "service_backed"].includes(plan.mode)) {
    throw new Error(`${label}.mode must be phase or service_backed`);
  }
  const expectedServiceBackedOnly = plan.mode === "service_backed";
  if (
    plan.service_backed_only !== undefined &&
    plan.service_backed_only !== expectedServiceBackedOnly
  ) {
    throw new Error(`${label}.service_backed_only must match mode`);
  }
  if (plan.target === "service-backed-slice" && !expectedServiceBackedOnly) {
    throw new Error(`${label}.target service-backed-slice requires service_backed mode`);
  }
  if (plan.target === "phase-slice" && expectedServiceBackedOnly) {
    throw new Error(`${label}.target phase-slice requires phase mode`);
  }

  const resourceLimits = resourceLimitObject(plan.resource_limits);
  validateResourceMap(resourceLimits, `${label}.resource_limits`, {
    allowAuto: true,
  });

  const workUnits = requireArray(plan.work_units, `${label}.work_units`);
  const ids = new Set();
  const completionKeys = new Map();
  let counted = 0;
  let finalizers = 0;
  for (const [index, unit] of workUnits.entries()) {
    const unitLabel = `${label}.work_units[${index + 1}]`;
    requireObject(unit, unitLabel);
    const id = requireString(unit.id, `${unitLabel}.id`);
    if (ids.has(id)) {
      throw new Error(`${label}.work_units contains duplicate id ${id}`);
    }
    ids.add(id);
    requireString(unit.label, `${unitLabel}.label`);
    requireString(unit.kind, `${unitLabel}.kind`);
    requireString(unit.target, `${unitLabel}.target`);
    for (const field of [
      "type",
      "class",
      "aggregateTarget",
      "group",
      "browserStage",
      "shard",
      "schedulerProfile",
      "unblockLabel",
    ]) {
      validateOptionalString(unit[field], `${unitLabel}.${field}`);
    }
    validateOptionalNonnegativeInteger(unit.weight_ms, `${unitLabel}.weight_ms`);
    validateOptionalBoolean(unit.countInTotal, `${unitLabel}.countInTotal`);
    validateOptionalBoolean(unit.countsStarted, `${unitLabel}.countsStarted`);
    validateOptionalBoolean(unit.completeOnFailure, `${unitLabel}.completeOnFailure`);
    validateStringArray(unit.needs ?? [], `${unitLabel}.needs`);
    validateStringArray(unit.runningDependencyKeys ?? [], `${unitLabel}.runningDependencyKeys`);
    validateStringArray(unit.shardNames ?? [], `${unitLabel}.shardNames`);
    validateStringArray(unit.failureKeys ?? unit.failure_keys ?? [], `${unitLabel}.failureKeys`);
    validateResourceMap(unit.resource_claims ?? {}, `${unitLabel}.resource_claims`, {
      limits: resourceLimits,
    });
    if (unit.countInTotal === false) {
      finalizers += 1;
    } else {
      counted += 1;
    }
    if (unit.kind === "finalizer") {
      finalizers += unit.countInTotal === false ? 0 : 1;
    }
    for (const key of unitCompletionKeys(unit)) {
      const normalized = requireString(key, `${unitLabel}.completionKeys[]`);
      if (completionKeys.has(normalized)) {
        throw new Error(
          `${label}.work_units completion key ${normalized} is produced by both ${completionKeys.get(normalized)} and ${id}`,
        );
      }
      completionKeys.set(normalized, id);
    }
  }

  for (const [index, unit] of workUnits.entries()) {
    for (const need of validateStringArray(unit.needs ?? [], `${label}.work_units[${index + 1}].needs`)) {
      if (!completionKeys.has(need)) {
        throw new Error(`${label}.work_units[${index + 1}].needs references unknown completion key ${need}`);
      }
    }
  }

  if (plan.total_work_units !== undefined && plan.total_work_units !== counted) {
    throw new Error(`${label}.total_work_units must match counted work units`);
  }
  if (plan.finalizer_count !== undefined && plan.finalizer_count !== finalizers) {
    throw new Error(`${label}.finalizer_count must match finalizer work units`);
  }
  if (plan.no_op !== (workUnits.length === 0)) {
    throw new Error(`${label}.no_op must match work_units emptiness`);
  }

  if (Array.isArray(plan.child_target_names)) {
    const sorted = [...plan.child_target_names].sort(compareStrings);
    if (plan.phase_namespace !== "frontend" && sorted.join("\u0000") !== plan.child_target_names.join("\u0000")) {
      throw new Error(`${label}.child_target_names must be sorted`);
    }
  }
  return plan;
}
