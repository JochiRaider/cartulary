import { readFile } from "node:fs/promises";
import path from "node:path";

import {
  browserGroupWorkerSlotCount,
} from "./adapters/browser.mjs";
import {
  assertObjectKeys,
  readJsonObject,
  requireArray,
  requireEnum,
  requireInteger,
  requireObject,
  requireObjectArray,
  requirePositiveInteger,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "../contract/json-shape.mjs";
import {
  maxResourceClaims,
  normalizeResourceClaims as normalizeSchedulerResourceClaims,
  normalizeResourceLimits as normalizeSchedulerResourceLimits,
  provisionalResourceLimitsForClaims,
  resolveAutoResourceLimits,
} from "./scheduler-resources.mjs";
import {
  isSchedulerFamily,
  requireSchedulerCapacityProfileForFamily,
  schedulerFamilySet,
} from "./scheduler-family-contract.mjs";

const schedulerManifestSchemaID = "cartulary.scheduler_manifest.v2";
const envNamePattern = /^[A-Z][A-Z0-9_]*$/;
const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const makePrerequisitePolicies = new Set(["run", "skip"]);
const schedulerManifestKeys = new Set(["schema_id", "generated", "schedules"]);
const schedulerScheduleKeys = new Set([
  "target",
  "scheduler_kind",
  "capacity_profile",
  "resource_limits",
  "stop_on_first_failure",
  "progress_tick_seconds",
  "validate_timing",
  "summary_groups",
  "work_units",
  "finalizers",
]);
const schedulerWorkUnitKeys = new Set([
  "id",
  "kind",
  "type",
  "class",
  "target",
  "label",
  "aggregate_target",
  "priority",
  "weight_ms",
  "needs",
  "produces_summary_targets",
  "completion_keys",
  "failure_keys",
  "running_dependency_keys",
  "resource_claims",
  "retained_resource_claims",
  "release_retained_resource_claims",
  "make_jobs",
  "make_prerequisite_policy",
  "env",
  "runtime_binaries",
  "service_session",
  "browser_stage",
  "browser_session_group",
  "browser_session_isolation_reason",
  "browser_session_finalizer",
  "browser_group",
  "shard",
  "shard_names",
  "scheduler_profile",
  "readiness_attribution",
  "count_in_total",
  "counts_started",
  "complete_on_failure",
  "unblock_label",
  "timeout_seconds",
  "command",
]);
const schedulerCommandTypeValues = Object.freeze([
  "make_target",
  "service_session_start",
  "browser_stage_session_start",
  "browser_group",
  "browser_stage_complete",
  "browser_session_finalizer",
  "go_shard",
  "go_shard_finalize",
  "service_complete",
]);
const commandTypes = new Set(schedulerCommandTypeValues);
const readinessTimingRoles = new Set([
  "readiness",
  "provisioning",
  "build_artifact",
  "service_setup",
]);
const readinessClasses = new Set([
  "frontend_install",
  "toolchain",
  "build_artifact",
  "embedded_web_assets",
  "test_service_binary",
  "service_image",
  "browser_readiness",
  "service_readiness",
]);

const makeTargetReadinessAttribution = Object.freeze({
  "check-frontend-install": {
    timing_role: "readiness",
    readiness_class: "frontend_install",
    warm_threshold_ms: 30000,
    reason: "pnpm-managed workspace dependency readiness",
  },
  "frontend-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned frontend toolchain readiness",
  },
  "codegen-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned code-generation toolchain readiness",
  },
  "go-lint-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned Go lint toolchain readiness",
  },
  "govulncheck-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned Govulncheck toolchain readiness",
  },
  "gosec-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned Gosec toolchain readiness",
  },
  "shell-lint-toolchain": {
    timing_role: "readiness",
    readiness_class: "toolchain",
    warm_threshold_ms: 10000,
    reason: "pinned shell lint toolchain readiness",
  },
  "build-server": {
    timing_role: "build_artifact",
    readiness_class: "build_artifact",
    warm_threshold_ms: 15000,
    reason: "server binary build artifact readiness",
  },
  "build-server-harness": {
    timing_role: "build_artifact",
    readiness_class: "build_artifact",
    warm_threshold_ms: 15000,
    reason: "harness server binary build artifact readiness",
  },
  "build-migrate": {
    timing_role: "build_artifact",
    readiness_class: "build_artifact",
    warm_threshold_ms: 15000,
    reason: "migration binary build artifact readiness",
  },
  "build-operator": {
    timing_role: "build_artifact",
    readiness_class: "build_artifact",
    warm_threshold_ms: 15000,
    reason: "operator binary build artifact readiness",
  },
  "build-web": {
    timing_role: "build_artifact",
    readiness_class: "build_artifact",
    warm_threshold_ms: 15000,
    reason: "web build artifact readiness",
  },
  "embedded-web-assets": {
    timing_role: "build_artifact",
    readiness_class: "embedded_web_assets",
    warm_threshold_ms: 15000,
    reason: "embedded web asset archive readiness",
  },
  "testservices-build": {
    timing_role: "build_artifact",
    readiness_class: "test_service_binary",
    warm_threshold_ms: 15000,
    reason: "test services binary build artifact readiness",
  },
  "test-service-images": {
    timing_role: "readiness",
    readiness_class: "service_image",
    warm_threshold_ms: 15000,
    reason: "test service image warm readiness",
  },
});

export function readinessAttributionForMakeTarget(target) {
  const attribution = makeTargetReadinessAttribution[target];
  return attribution ? { ...attribution } : null;
}
const schedulerCommandShapes = Object.freeze({
  make_target: {
    required: ["target"],
    optional: ["service_target"],
  },
  service_session_start: {
    required: ["service_target"],
    optional: [],
  },
  browser_stage_session_start: {
    required: ["service_target", "browser_stage"],
    optional: [],
  },
  browser_group: {
    required: ["service_target", "browser_stage", "group_id"],
    optional: [],
  },
  browser_stage_complete: {
    required: ["service_target", "browser_stage"],
    optional: [],
  },
  browser_session_finalizer: {
    required: ["service_target", "browser_session_group"],
    optional: [],
  },
  go_shard: {
    required: ["target", "shard", "service_target"],
    optional: [],
  },
  go_shard_finalize: {
    required: ["target", "service_target"],
    optional: [],
  },
  service_complete: {
    required: ["service_target"],
    optional: [],
  },
});

function validateSchedulerCommandShape(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  if (typeof value.type !== "string" || value.type.trim() === "") {
    throw new Error(`${label}.type must be a non-empty string`);
  }
  if (!commandTypes.has(value.type)) {
    throw new Error(
      `${label}.type must be one of ${schedulerCommandTypeValues.join("|")}`,
    );
  }
  const shape = schedulerCommandShapes[value.type];
  const allowed = new Set(["type", ...shape.required, ...shape.optional]);
  for (const key of Object.keys(value)) {
    if (!allowed.has(key)) {
      throw new Error(`${label} has unknown key ${key}`);
    }
  }
  for (const field of shape.required) {
    if (typeof value[field] !== "string" || value[field].trim() === "") {
      throw new Error(`${label}.${field} must be a non-empty string`);
    }
  }
  for (const field of shape.optional) {
    if (
      value[field] !== undefined &&
      (typeof value[field] !== "string" || value[field].trim() === "")
    ) {
      throw new Error(`${label}.${field} must be a non-empty string`);
    }
  }
}

function parseWorkerEnvInteger(value, label, { min }) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be declared`);
  }
  const parsed = Number.parseInt(value, 10);
  if (!Number.isInteger(parsed) || parsed < min || String(parsed) !== value) {
    throw new Error(
      `${label} must be ${min === 0 ? "a non-negative" : "a positive"} integer`,
    );
  }
  return parsed;
}

function validateBrowserWorkerSlotDescriptors(descriptors) {
  const groupsByRuntime = new Map();
  for (const descriptor of descriptors) {
    const group = groupsByRuntime.get(descriptor.runtime) ?? [];
    group.push(descriptor);
    groupsByRuntime.set(descriptor.runtime, group);
  }

  for (const [runtime, descriptorsForRuntime] of groupsByRuntime.entries()) {
    const totalWorkerCount = descriptorsForRuntime.reduce(
      (sum, descriptor) =>
        sum + browserGroupWorkerSlotCount(descriptor.browserGroup),
      0,
    );
    const occupied = new Set();
    for (const descriptor of descriptorsForRuntime) {
      const count = parseWorkerEnvInteger(
        descriptor.workerCountValue,
        descriptor.workerCountLabel,
        { min: 1 },
      );
      if (count !== totalWorkerCount) {
        throw new Error(descriptor.workerCountMismatch(totalWorkerCount));
      }
      const offset = parseWorkerEnvInteger(
        descriptor.workerOffsetValue,
        descriptor.workerOffsetLabel,
        { min: 0 },
      );
      const slotCount = browserGroupWorkerSlotCount(descriptor.browserGroup);
      if (offset + slotCount > totalWorkerCount) {
        throw new Error(descriptor.workerRangeExceeded(totalWorkerCount));
      }
      for (let slot = offset; slot < offset + slotCount; slot += 1) {
        if (occupied.has(slot)) {
          throw new Error(descriptor.workerSlotOverlap(slot));
        }
        occupied.add(slot);
      }
    }
    if (occupied.size !== totalWorkerCount) {
      throw new Error(descriptorsForRuntime[0].workerSlotsNotContiguous(runtime));
    }
  }
}

function validateSchedulerBrowserWorkerSlots(units, label) {
  const descriptors = [];
  for (const unit of units ?? []) {
    if (unit?.kind !== "browser_group") {
      continue;
    }
    const runtime =
      typeof unit.service_session?.target === "string" &&
      unit.service_session.target.trim() !== ""
        ? unit.service_session.target.trim()
        : label;
    const unitID = unit.id ?? unit.target;
    descriptors.push({
      browserGroup: unit.browser_group,
      runtime,
      workerCountLabel: `${label}.${unitID}.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT`,
      workerCountMismatch: (total) =>
        `${label}.${unitID} worker count must equal ${total} for ${runtime}`,
      workerCountValue: unit.env?.CARTULARY_PLAYWRIGHT_WORKER_COUNT,
      workerOffsetLabel: `${label}.${unitID}.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET`,
      workerOffsetValue: unit.env?.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET,
      workerRangeExceeded: (total) =>
        `${label}.${unitID} worker slot range exceeds ${total} for ${runtime}`,
      workerSlotOverlap: (slot) =>
        `${label} ${runtime} has overlapping worker-admin slot ${slot}`,
      workerSlotsNotContiguous: () =>
        `${label} ${runtime} worker-admin slots must be contiguous`,
    });
  }
  validateBrowserWorkerSlotDescriptors(descriptors);
}

export function validateSchedulerManifestShape(file) {
  const manifest = readJsonObject(file, file);
  assertObjectKeys(manifest, schedulerManifestKeys, file);
  requireSchemaID(manifest, schedulerManifestSchemaID, file);
  requireObject(manifest.generated, `${file}.generated`);
  requireObjectArray(manifest.schedules, `${file}.schedules`, {
    nonEmpty: true,
  }).forEach((schedule, index) => {
    const label = `${file}.schedules[${index + 1}]`;
    assertObjectKeys(schedule, schedulerScheduleKeys, label);
    requireString(schedule.target, `${label}.target`, {
      pattern: makeTargetPattern,
    });
    requireEnum(schedule.scheduler_kind, `${label}.scheduler_kind`, schedulerFamilySet);
    requireSchedulerCapacityProfileForFamily(
      requireString(schedule.capacity_profile, `${label}.capacity_profile`),
      schedule.scheduler_kind,
      label,
    );
    requireObject(schedule.resource_limits, `${label}.resource_limits`);
    if (
      schedule.stop_on_first_failure !== undefined &&
      typeof schedule.stop_on_first_failure !== "boolean"
    ) {
      throw new Error(`${label}.stop_on_first_failure must be a boolean`);
    }
    if (
      schedule.validate_timing !== undefined &&
      typeof schedule.validate_timing !== "boolean"
    ) {
      throw new Error(`${label}.validate_timing must be a boolean`);
    }
    if (schedule.progress_tick_seconds !== undefined) {
      requireInteger(schedule.progress_tick_seconds, `${label}.progress_tick_seconds`, {
        min: 5,
      });
      if (schedule.progress_tick_seconds > 300) {
        throw new Error(`${label}.progress_tick_seconds must be <= 300`);
      }
    }
    requireObjectArray(schedule.work_units, `${label}.work_units`, {
      nonEmpty: true,
    }).forEach((unit, unitIndex) => {
      const unitLabel = `${label}.work_units[${unitIndex + 1}]`;
      assertObjectKeys(unit, schedulerWorkUnitKeys, unitLabel);
      requireString(unit.target, `${unitLabel}.target`, {
        pattern: makeTargetPattern,
      });
      requirePositiveInteger(unit.weight_ms, `${unitLabel}.weight_ms`);
      const command = requireObject(unit.command, `${unitLabel}.command`);
      validateSchedulerCommandShape(command, `${unitLabel}.command`);
      if (unit.shard_names !== undefined) {
        requireStringArray(unit.shard_names, `${unitLabel}.shard_names`);
      }
      if (command.type === "go_shard_finalize") {
        const shardNames = requireStringArray(
          unit.shard_names,
          `${unitLabel}.shard_names`,
          { nonEmpty: true },
        );
        const expectedNeeds = shardNames.map((shardName) => `go_shard:${shardName}`);
        const needs = requireStringArray(unit.needs ?? [], `${unitLabel}.needs`);
        for (const expectedNeed of expectedNeeds) {
          if (!needs.includes(expectedNeed)) {
            throw new Error(
              `${unitLabel}.shard_names must match needs; missing ${expectedNeed}`,
            );
          }
        }
        for (const need of needs.filter((entry) => entry.startsWith("go_shard:"))) {
          if (!expectedNeeds.includes(need)) {
            throw new Error(
              `${unitLabel}.needs includes ${need} not declared by shard_names`,
            );
          }
        }
      }
      if (unit.priority !== undefined) {
        requireInteger(unit.priority, `${unitLabel}.priority`, { min: 0 });
      }
      if (command.type === "make_target" && unit.make_prerequisite_policy === undefined) {
        throw new Error(`${unitLabel}.make_prerequisite_policy is required for make_target work units`);
      }
      if (command.type !== "make_target" && unit.make_prerequisite_policy !== undefined) {
        throw new Error(`${unitLabel}.make_prerequisite_policy is only supported for make_target work units`);
      }
      if (unit.make_prerequisite_policy !== undefined) {
        requireEnum(
          unit.make_prerequisite_policy,
          `${unitLabel}.make_prerequisite_policy`,
          makePrerequisitePolicies,
        );
      }
      if (unit.env !== undefined) {
        for (const name of Object.keys(
          requireObject(unit.env, `${unitLabel}.env`),
        )) {
          requireString(name, `${unitLabel}.env key`, {
            pattern: envNamePattern,
          });
        }
      }
      if (unit.runtime_binaries !== undefined) {
        requireStringArray(unit.runtime_binaries, `${unitLabel}.runtime_binaries`, {
          nonEmpty: true,
        });
      }
      if (unit.readiness_attribution !== undefined) {
        const readinessAttribution = requireObject(
          unit.readiness_attribution,
          `${unitLabel}.readiness_attribution`,
        );
        assertObjectKeys(
          readinessAttribution,
          new Set(["timing_role", "readiness_class", "warm_threshold_ms", "reason"]),
          `${unitLabel}.readiness_attribution`,
        );
        requireEnum(
          readinessAttribution.timing_role,
          `${unitLabel}.readiness_attribution.timing_role`,
          readinessTimingRoles,
        );
        requireEnum(
          readinessAttribution.readiness_class,
          `${unitLabel}.readiness_attribution.readiness_class`,
          readinessClasses,
        );
        requireInteger(
          readinessAttribution.warm_threshold_ms,
          `${unitLabel}.readiness_attribution.warm_threshold_ms`,
          { min: 0 },
        );
        requireString(
          readinessAttribution.reason,
          `${unitLabel}.readiness_attribution.reason`,
        );
      }
      if (unit.timeout_seconds !== undefined) {
        requireInteger(unit.timeout_seconds, `${unitLabel}.timeout_seconds`, { min: 1, max: 3600 });
      }
    });
    validateSchedulerBrowserWorkerSlots(schedule.work_units, label);
    if (schedule.finalizers !== undefined) {
      requireArray(schedule.finalizers, `${label}.finalizers`);
    }
  });
}

export function parseResourceLimitOverride(value) {
  const [resource, amountText, extra] = value.split("=");
  if (!resource || !amountText || extra !== undefined) {
    throw new Error(`--resource-limit expects <name=value>, got ${value}`);
  }
  const amount = Number.parseInt(amountText, 10);
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`--resource-limit ${resource} must be a positive integer`);
  }
  return [resource.trim(), amount];
}

export async function loadSchedulerManifest(file, { repoRoot, schemaID = schedulerManifestSchemaID } = {}) {
  const manifestPath = path.isAbsolute(file) ? file : path.join(repoRoot, file);
  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (manifest.schema_id !== schemaID) {
    throw new Error(`${manifestPath} must declare schema_id ${schemaID}; run make generate to regenerate the normalized scheduler manifest`);
  }
  if (!Array.isArray(manifest.schedules)) {
    throw new Error(`${manifestPath} must declare schedules[]`);
  }
  return { manifest, manifestPath };
}

const loadScheduleManifest = loadSchedulerManifest;

function selectSingleSchedule(manifest, target, { label = "schedule" } = {}) {
  const matches = manifest.schedules.filter((schedule) => schedule?.target === target);
  if (matches.length !== 1) {
    throw new Error(`expected exactly one ${label} for ${target}, found ${matches.length}`);
  }
  return matches[0];
}

function normalizeResourceClaims(value, label, resourceLimits, scheduler) {
  return normalizeSchedulerResourceClaims(value ?? {}, label, resourceLimits, {
    scheduler,
    allowBounded: scheduler === "check",
  });
}

function normalizePriority(value, label) {
  if (value === undefined) {
    return 0;
  }
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`${label} priority must be a non-negative integer`);
  }
  return value;
}

function normalizePositiveWeight(value, label) {
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} weight_ms must be a positive integer`);
  }
  return value;
}

function normalizeBoolean(value, label, fallback) {
  if (value === undefined) {
    return fallback;
  }
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

function normalizeProgressTickSeconds(value, label) {
  if (value === undefined) {
    return 30;
  }
  if (!Number.isInteger(value) || value < 5 || value > 300) {
    throw new Error(`${label} must be an integer between 5 and 300`);
  }
  return value;
}

function normalizeEnv(value, label) {
  if (value === undefined) {
    return {};
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} env must be an object`);
  }
  const entries = [];
  for (const [name, rawValue] of Object.entries(value)) {
    if (!envNamePattern.test(name)) {
      throw new Error(`${label} env.${name} must be a safe environment variable name`);
    }
    if (typeof rawValue !== "string" || rawValue.includes("\0") || rawValue.includes("\n") || rawValue.includes("\r")) {
      throw new Error(`${label} env.${name} must be a single-line string`);
    }
    entries.push([name, rawValue]);
  }
  return Object.fromEntries(entries.sort(([left], [right]) => left.localeCompare(right)));
}

function normalizeCommand(value, label) {
  validateSchedulerCommandShape(value, `${label} command`);
  return Object.fromEntries(
    Object.entries(value).map(([key, raw]) => [
      key,
      typeof raw === "string" ? raw.trim() : raw,
    ]),
  );
}

function normalizeRetainedResourceClaims(value, label, resourceClaims, resourceLimits, scheduler) {
  if (value === undefined) {
    return new Map();
  }
  const retained = normalizeResourceClaims(value, label, resourceLimits, scheduler);
  for (const [resource, amount] of retained.entries()) {
    if ((resourceClaims.get(resource) ?? 0) < amount) {
      throw new Error(`${label}.${resource} exceeds resource_claims`);
    }
  }
  return retained;
}

function normalizeMakeJobs(value, label, resourceClaims, scheduler) {
  if (value === undefined) {
    return 1;
  }
  if (typeof value === "string") {
    if (scheduler !== "check") {
      throw new Error(`${label} make_jobs resource references are only supported by check schedules`);
    }
    return resourceClaims.get(value) ?? 1;
  }
  if (!Number.isInteger(value) || value < 1) {
    throw new Error(`${label} make_jobs must be a positive integer or claimed resource name`);
  }
  return value;
}

function normalizeMakePrerequisitePolicy(value, label, commandType) {
  if (commandType !== "make_target") {
    if (value !== undefined) {
      throw new Error(`${label} make_prerequisite_policy is only supported for make_target work units`);
    }
    return undefined;
  }
  if (value === undefined) {
    throw new Error(`${label} make_prerequisite_policy is required for make_target work units`);
  }
  if (typeof value !== "string" || !makePrerequisitePolicies.has(value)) {
    throw new Error(`${label} make_prerequisite_policy must be one of run, skip`);
  }
  return value;
}

function normalizeReadinessAttribution(value, label) {
  if (value === undefined) {
    return null;
  }
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} readiness_attribution must be an object`);
  }
  const timingRole = typeof value.timing_role === "string" ? value.timing_role.trim() : "";
  const readinessClass = typeof value.readiness_class === "string" ? value.readiness_class.trim() : "";
  const reason = typeof value.reason === "string" ? value.reason.trim() : "";
  if (!readinessTimingRoles.has(timingRole)) {
    throw new Error(`${label} readiness_attribution.timing_role must be a known readiness timing role`);
  }
  if (!readinessClasses.has(readinessClass)) {
    throw new Error(`${label} readiness_attribution.readiness_class must be a known readiness class`);
  }
  if (!Number.isInteger(value.warm_threshold_ms) || value.warm_threshold_ms < 0) {
    throw new Error(`${label} readiness_attribution.warm_threshold_ms must be a non-negative integer`);
  }
  if (reason === "") {
    throw new Error(`${label} readiness_attribution.reason must be a non-empty string`);
  }
  return {
    timing_role: timingRole,
    readiness_class: readinessClass,
    warm_threshold_ms: value.warm_threshold_ms,
    reason,
  };
}

function normalizeWorkUnit(unit, index, scheduleLabel, scheduler, resourceLimits) {
  const label = `${scheduleLabel} work_units ${index + 1}`;
  if (!unit || typeof unit !== "object" || Array.isArray(unit)) {
    throw new Error(`${label} must be an object`);
  }
  if (typeof unit.target !== "string" || unit.target.trim() === "") {
    throw new Error(`${label} must declare target`);
  }
  const target = unit.target.trim();
  const id = typeof unit.id === "string" && unit.id.trim() !== "" ? unit.id.trim() : target;
  const kind = typeof unit.kind === "string" && unit.kind.trim() !== "" ? unit.kind.trim() : "make_target";
  const aggregateTarget =
    typeof unit.aggregate_target === "string" && unit.aggregate_target.trim() !== ""
      ? unit.aggregate_target.trim()
      : target;
  const claims = normalizeResourceClaims(unit.resource_claims, `${label} ${target}`, resourceLimits, scheduler);
  if (kind === "browser_group" && claims.get("process") !== 1) {
    throw new Error(`${label} ${target} browser_group must claim exactly one process slot`);
  }
  const retainedResourceClaims = normalizeRetainedResourceClaims(
    unit.retained_resource_claims,
    `${label} ${target} retained_resource_claims`,
    claims,
    resourceLimits,
    scheduler,
  );
  const commandSpec = normalizeCommand(unit.command, `${label} ${target}`);
  return {
    id,
    label: typeof unit.label === "string" && unit.label.trim() !== "" ? unit.label.trim() : target,
    kind,
    type: typeof unit.type === "string" && unit.type.trim() !== "" ? unit.type.trim() : kind,
    class: typeof unit.class === "string" ? unit.class.trim() : "",
    target,
    aggregateTarget,
    completionKeys: normalizeStringList(unit.completion_keys, `${label} ${target} completion_keys`).length > 0
      ? normalizeStringList(unit.completion_keys, `${label} ${target} completion_keys`)
      : [id],
    failureKeys: normalizeStringList(unit.failure_keys, `${label} ${target} failure_keys`).length > 0
      ? normalizeStringList(unit.failure_keys, `${label} ${target} failure_keys`)
      : undefined,
    runningDependencyKeys: normalizeStringList(
      unit.running_dependency_keys,
      `${label} ${target} running_dependency_keys`,
    ),
    priority: normalizePriority(unit.priority, `${label} ${target}`),
    weightMs: normalizePositiveWeight(unit.weight_ms, `${label} ${target}`),
    needs: normalizeNeeds(unit.needs, `${label} ${target}`),
    producesSummaryTargets: normalizeStringList(
      unit.produces_summary_targets,
      `${label} ${target} produces_summary_targets`,
    ),
    resourceClaims: claims,
    retainedResourceClaims,
    releaseRetainedResourceClaims: normalizeResourceClaims(
      unit.release_retained_resource_claims ?? {},
      `${label} ${target} release_retained_resource_claims`,
      resourceLimits,
      scheduler,
    ),
    makeJobs: normalizeMakeJobs(unit.make_jobs, `${label} ${target}`, claims, scheduler),
    makePrerequisitePolicy: normalizeMakePrerequisitePolicy(
      unit.make_prerequisite_policy,
      `${label} ${target}`,
      commandSpec.type,
    ),
    env: normalizeEnv(unit.env, `${label} ${target}`),
    runtimeBinaries: normalizeStringList(
      unit.runtime_binaries,
      `${label} ${target} runtime_binaries`,
    ),
    commandSpec,
    serviceSession: unit.service_session && typeof unit.service_session === "object" && !Array.isArray(unit.service_session)
      ? JSON.parse(JSON.stringify(unit.service_session))
      : null,
    browserStage: typeof unit.browser_stage === "string" ? unit.browser_stage : "",
    browserSessionGroup:
      typeof unit.browser_session_group === "string" && unit.browser_session_group.trim() !== ""
        ? unit.browser_session_group.trim()
        : "",
    browserSessionIsolationReason:
      typeof unit.browser_session_isolation_reason === "string" &&
      unit.browser_session_isolation_reason.trim() !== ""
        ? unit.browser_session_isolation_reason.trim()
        : "",
    browserSessionFinalizer:
      unit.browser_session_finalizer === undefined
        ? undefined
        : unit.browser_session_finalizer === true,
    browserGroup: unit.browser_group && typeof unit.browser_group === "object" && !Array.isArray(unit.browser_group)
      ? JSON.parse(JSON.stringify(unit.browser_group))
      : null,
    shard: typeof unit.shard === "string" ? unit.shard : "",
    shardNames: Array.isArray(unit.shard_names)
      ? unit.shard_names.filter((entry) => typeof entry === "string" && entry !== "")
      : [],
    schedulerProfile: typeof unit.scheduler_profile === "string" ? unit.scheduler_profile : "",
    readinessAttribution: normalizeReadinessAttribution(unit.readiness_attribution, `${label} ${target}`),
    countInTotal: unit.count_in_total === false ? false : undefined,
    countsStarted: unit.counts_started === false ? false : undefined,
    completeOnFailure: unit.complete_on_failure === true,
    unblockLabel: typeof unit.unblock_label === "string" && unit.unblock_label.trim() !== ""
      ? unit.unblock_label.trim()
      : undefined,
    timeoutMs: unit.timeout_seconds === undefined ? 0 : unit.timeout_seconds * 1_000,
    startDetail: {},
    order: index,
  };
}

function validateWorkUnitDependencyGraph(units, scheduleLabel) {
  const ids = new Set();
  const completionKeys = new Map();
  for (const unit of units) {
    if (ids.has(unit.id)) {
      throw new Error(`${scheduleLabel} contains duplicate work unit id ${unit.id}`);
    }
    ids.add(unit.id);
    for (const key of unit.completionKeys) {
      if (completionKeys.has(key)) {
        throw new Error(`${scheduleLabel} completion key ${key} is produced by both ${completionKeys.get(key)} and ${unit.id}`);
      }
      completionKeys.set(key, unit.id);
    }
  }
  for (const unit of units) {
    unit.failureKeys = unit.failureKeys?.length ? unit.failureKeys : unit.completionKeys;
    for (const need of unit.needs) {
      if (!completionKeys.has(need)) {
        throw new Error(`${scheduleLabel} work unit ${unit.id} depends on unknown completion key ${need}`);
      }
    }
  }
  const byID = new Map(units.map((unit) => [unit.id, unit]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (unit) => {
    if (visited.has(unit.id)) {
      return;
    }
    if (visiting.has(unit.id)) {
      throw new Error(`${scheduleLabel} has a dependency cycle at ${unit.id}`);
    }
    visiting.add(unit.id);
    for (const need of unit.needs) {
      const producer = byID.get(completionKeys.get(need));
      if (producer) {
        visit(producer);
      }
    }
    visiting.delete(unit.id);
    visited.add(unit.id);
  };
  for (const unit of units) {
    visit(unit);
  }
}

function validateRetainedClaimReleases(units, scheduleLabel) {
  const retainedTotals = new Map();
  const releaseTotals = new Map();
  for (const unit of units) {
    for (const [resource, amount] of unit.retainedResourceClaims.entries()) {
      retainedTotals.set(resource, (retainedTotals.get(resource) ?? 0) + amount);
    }
    for (const [resource, amount] of unit.releaseRetainedResourceClaims.entries()) {
      releaseTotals.set(resource, (releaseTotals.get(resource) ?? 0) + amount);
    }
  }
  for (const [resource, amount] of releaseTotals.entries()) {
    if ((retainedTotals.get(resource) ?? 0) < amount) {
      throw new Error(`${scheduleLabel} releases retained ${resource} claims that were never retained`);
    }
  }
}

function validateGoShardFinalizers(units, scheduleLabel) {
  const completionKeys = new Set(units.flatMap((unit) => unit.completionKeys));
  for (const unit of units) {
    if (unit.commandSpec.type !== "go_shard_finalize") {
      continue;
    }
    if (unit.shardNames.length === 0) {
      throw new Error(
        `${scheduleLabel} ${unit.id} go_shard_finalize must declare shard_names[]`,
      );
    }
    const expectedNeeds = unit.shardNames.map((shardName) => `go_shard:${shardName}`);
    const missingNeeds = expectedNeeds.filter((need) => !unit.needs.includes(need));
    if (missingNeeds.length > 0) {
      throw new Error(
        `${scheduleLabel} ${unit.id} go_shard_finalize shard_names must match needs[]; missing ${missingNeeds[0]}`,
      );
    }
    const extraShardNeeds = unit.needs.filter(
      (need) => need.startsWith("go_shard:") && !expectedNeeds.includes(need),
    );
    if (extraShardNeeds.length > 0) {
      throw new Error(
        `${scheduleLabel} ${unit.id} go_shard_finalize needs[] contains shard not declared in shard_names[]: ${extraShardNeeds[0]}`,
      );
    }
    const missingProducers = expectedNeeds.filter((need) => !completionKeys.has(need));
    if (missingProducers.length > 0) {
      throw new Error(
        `${scheduleLabel} ${unit.id} go_shard_finalize references unproduced shard completion key ${missingProducers[0]}`,
      );
    }
  }
}

function validateBrowserWorkerSlotRanges(units, scheduleLabel, scheduleTarget) {
  const descriptors = [];
  for (const unit of units) {
    if (unit.kind !== "browser_group") {
      continue;
    }
    const runtimeKey =
      unit.serviceSession &&
      typeof unit.serviceSession.target === "string" &&
      unit.serviceSession.target.trim() !== ""
        ? unit.serviceSession.target.trim()
        : scheduleTarget;
    descriptors.push({
      browserGroup: unit.browserGroup,
      runtime: runtimeKey,
      workerCountLabel: `${scheduleLabel} ${unit.id} env.CARTULARY_PLAYWRIGHT_WORKER_COUNT`,
      workerCountMismatch: (total) =>
        `${scheduleLabel} ${unit.id} env.CARTULARY_PLAYWRIGHT_WORKER_COUNT must equal ${total} for browser runtime ${runtimeKey}`,
      workerCountValue: unit.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT,
      workerOffsetLabel: `${scheduleLabel} ${unit.id} env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET`,
      workerOffsetValue: unit.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET,
      workerRangeExceeded: (total) =>
        `${scheduleLabel} ${unit.id} worker slot range exceeds ${total} for browser runtime ${runtimeKey}`,
      workerSlotOverlap: (slot) =>
        `${scheduleLabel} browser runtime ${runtimeKey} has overlapping worker-admin slot ${slot}`,
      workerSlotsNotContiguous: () =>
        `${scheduleLabel} browser runtime ${runtimeKey} worker-admin slots must be contiguous`,
    });
  }
  validateBrowserWorkerSlotDescriptors(descriptors);
}

export function normalizeSchedulerSchedule(manifest, target, {
  scheduler = null,
  resourceLimitOverrides = new Map(),
  autoLimitResolvers = () => ({}),
  env = process.env,
  label = "scheduler schedule",
} = {}) {
  const schedule = selectSingleSchedule(manifest, target, { label });
  const scheduleLabel = `${label} ${target}`;
  if (scheduler && schedule.scheduler_kind !== scheduler) {
    throw new Error(`${scheduleLabel} scheduler_kind must be ${scheduler}`);
  }
  const schedulerKind = schedule.scheduler_kind;
  if (!isSchedulerFamily(schedulerKind)) {
    throw new Error(`${scheduleLabel} scheduler_kind is unsupported: ${schedulerKind}`);
  }
  if (!Array.isArray(schedule.work_units) || schedule.work_units.length === 0) {
    throw new Error(`${scheduleLabel} must declare work_units[]`);
  }
  if (Array.isArray(schedule.work_unit_sources)) {
    throw new Error(`${scheduleLabel} must use normalized work_units[], not work_unit_sources[]`);
  }
  const normalizedLimits = normalizeSchedulerResourceLimits(schedule.resource_limits, scheduleLabel, {
    scheduler: schedulerKind,
    capacityProfile: schedule.capacity_profile ?? null,
    overrides: resourceLimitOverrides,
    allowAuto: true,
    env,
  });
  const provisionalLimits = provisionalResourceLimitsForClaims(normalizedLimits.limits);
  const provisionalUnits = schedule.work_units.map((unit, index) =>
    normalizeWorkUnit(unit, index, scheduleLabel, schedulerKind, provisionalLimits),
  );
  const resolvedLimits = resolveAutoResourceLimits(
    normalizedLimits.limits,
    normalizedLimits.sources,
    scheduleLabel,
    autoLimitResolvers(provisionalUnits),
    maxResourceClaims(provisionalUnits),
  );
  const units = schedule.work_units.map((unit, index) =>
    normalizeWorkUnit(unit, index, scheduleLabel, schedulerKind, resolvedLimits.resourceLimits),
  );
  validateWorkUnitDependencyGraph(units, scheduleLabel);
  validateGoShardFinalizers(units, scheduleLabel);
  validateRetainedClaimReleases(units, scheduleLabel);
  validateBrowserWorkerSlotRanges(units, scheduleLabel, target);
  const sortedUnits = units.sort(
    (left, right) =>
      right.priority - left.priority ||
      right.weightMs - left.weightMs ||
      left.order - right.order ||
      left.id.localeCompare(right.id),
  );
  return {
    target,
    schedulerKind,
    capacityProfile: schedule.capacity_profile ?? "",
    stopOnFirstFailure: normalizeBoolean(
      schedule.stop_on_first_failure,
      `${scheduleLabel}.stop_on_first_failure`,
      schedulerKind === "check",
    ),
    progressTickSeconds: normalizeProgressTickSeconds(
      schedule.progress_tick_seconds,
      `${scheduleLabel}.progress_tick_seconds`,
    ),
    validateTiming: normalizeBoolean(schedule.validate_timing, `${scheduleLabel}.validate_timing`, true),
    summaryGroups: schedule.summary_groups ?? [],
    finalizers: Array.isArray(schedule.finalizers) ? schedule.finalizers : [],
    resourceLimits: resolvedLimits.resourceLimits,
    resourceLimitSources: resolvedLimits.resourceLimitSources,
    workUnits: sortedUnits,
  };
}

function normalizeStringList(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

function normalizeNeeds(value, label) {
  if (value === undefined) {
    return [];
  }
  if (!Array.isArray(value)) {
    throw new Error(`${label} needs must be an array`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || entry.trim() === "") {
      throw new Error(`${label} needs entries must be non-empty strings`);
    }
    return entry.trim();
  });
}

function validateTargetDependencyGraph(
  nodes,
  { scheduleLabel, nodeKind, duplicateTargetsMessage = null },
) {
  const targets = new Set();
  const duplicateTargets = [];
  for (const node of nodes) {
    if (targets.has(node.target)) {
      duplicateTargets.push(node.target);
    }
    targets.add(node.target);
  }
  if (duplicateTargets.length > 0) {
    if (duplicateTargetsMessage) {
      throw new Error(duplicateTargetsMessage(duplicateTargets));
    }
    throw new Error(`${scheduleLabel} contains duplicate ${nodeKind} target ${duplicateTargets[0]}`);
  }

  for (const node of nodes) {
    for (const need of node.needs) {
      if (!targets.has(need)) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} depends on unknown target ${need}`);
      }
      if (need === node.target) {
        throw new Error(`${scheduleLabel} ${nodeKind} ${node.target} cannot depend on itself`);
      }
    }
  }

  const byTarget = new Map(nodes.map((node) => [node.target, node]));
  const visiting = new Set();
  const visited = new Set();
  const visit = (node) => {
    if (visited.has(node.target)) {
      return;
    }
    if (visiting.has(node.target)) {
      throw new Error(`${scheduleLabel} has a dependency cycle at ${node.target}`);
    }
    visiting.add(node.target);
    for (const need of node.needs) {
      visit(byTarget.get(need));
    }
    visiting.delete(node.target);
    visited.add(node.target);
  };
  for (const node of nodes) {
    visit(node);
  }
}
