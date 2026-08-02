import os from "node:os";

import {
  estimateBrowserStackAutoLimit,
  estimateCheckHostCPULimit,
  estimateCheckHostIOLimit,
  estimatePostgresCloneAutoLimit,
  estimatePostgresResetAutoLimit,
  normalizeResourceLimits,
  resolveAutoResourceLimits,
} from "./scheduler-resources.mjs";
import {
  isSchedulerFamily,
  requireSchedulerFamily,
} from "./scheduler-family-contract.mjs";

const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const hostCPUResource = "host_cpu";
const hostIOResource = "host_io";
const postgresResetResource = "postgres_reset";
const postgresCloneResource = "postgres_clone";
const postgresClusterAdvisoryLockResource = "postgres_cluster_advisory_lock";
export const testSliceDefaultCapacityProfile = "test_slice_default";
export const sequenceAdaptiveCapacityProfile = "sequence_adaptive";

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }
  return value;
}

function addClaim(claims, resource, amount) {
  if (amount === "limit") {
    claims.set(resource, amount);
    return;
  }
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`resource claim ${resource} must be a positive integer or "limit"`);
  }
  if (claims.get(resource) === "limit") {
    return;
  }
  claims.set(resource, (claims.get(resource) ?? 0) + amount);
}

function resourceClaimsObject(value) {
  return Object.fromEntries(
    Object.entries(value).sort(([left], [right]) => left.localeCompare(right)),
  );
}

export function mapServiceBackedClaimsToCheckClaims(rawClaims, { ensureHost = false } = {}) {
  const claims = new Map();
  for (const [resource, amount] of Object.entries(requireObject(rawClaims, "resource_claims"))) {
    if (resource === goCPUResource) {
      addClaim(claims, hostCPUResource, amount);
    } else if (resource === goIOResource) {
      addClaim(claims, hostIOResource, amount);
    } else {
      addClaim(claims, resource, amount);
    }
  }
  if (ensureHost) {
    if (!claims.has(hostCPUResource)) {
      claims.set(hostCPUResource, 1);
    }
    if (!claims.has(hostIOResource)) {
      claims.set(hostIOResource, 1);
    }
  }
  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

export function goShardSchedulerProfileClaims(profile, { scheduler, resourceLimits = null, shardName = "" } = {}) {
  if (!isSchedulerFamily(scheduler)) {
    throw new Error(`unsupported scheduler resource family ${scheduler}`);
  }
  const claims = new Map();
  const cpuResource = scheduler === "check" ? hostCPUResource : goCPUResource;
  const ioResource = scheduler === "check" ? hostIOResource : goIOResource;
  const addProfileClaim = (resource, amount) => addClaim(claims, resource, amount);
  const requireLimit = (resource) => {
    if (resourceLimits?.has && !resourceLimits.has(resource)) {
      const label = shardName ? `go shard ${shardName}` : "go shard";
      throw new Error(`${label} has ${profile} profile but schedule is missing ${resource}`);
    }
  };

  switch (profile) {
    case "cpu_heavy":
      addProfileClaim(cpuResource, 2);
      addProfileClaim(ioResource, 1);
      break;
    case "io_heavy":
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 2);
      break;
    case "reset_heavy":
      requireLimit(postgresResetResource);
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 2);
      addProfileClaim(postgresResetResource, 1);
      break;
    case "clone_heavy":
      requireLimit(postgresCloneResource);
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 2);
      addProfileClaim(postgresCloneResource, 1);
      break;
    case "postgres_advisory_lock_exclusive":
      requireLimit(postgresCloneResource);
      requireLimit(postgresClusterAdvisoryLockResource);
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 2);
      addProfileClaim(postgresCloneResource, 1);
      addProfileClaim(postgresClusterAdvisoryLockResource, 1);
      break;
    case "transaction_heavy":
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 1);
      break;
    default:
      addProfileClaim(cpuResource, 1);
      addProfileClaim(ioResource, 1);
      break;
  }

  return resourceClaimsObject(Object.fromEntries(claims.entries()));
}

function clampInteger(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function availableCPUCount() {
  if (typeof os.availableParallelism === "function") {
    return Math.max(1, os.availableParallelism());
  }
  return Math.max(1, os.cpus().length);
}

export function estimateSequenceHostCPULimit(availableParallelism = availableCPUCount()) {
  return Math.max(1, Math.floor(availableParallelism * 0.85));
}

export function estimateSequenceHostIOLimit(
  hostCPULimit,
  availableParallelism = availableCPUCount(),
) {
  return Math.max(hostCPULimit, availableParallelism);
}

export function estimateSequenceProcessLimit(availableParallelism = availableCPUCount()) {
  return clampInteger(Math.floor(availableParallelism / 3), 2, 8);
}

export function estimateHostProcessSlotLimit(availableParallelism = availableCPUCount()) {
  return clampInteger(Math.floor(availableParallelism / 2), 2, 12);
}

function goShardUnits(workUnits) {
  return (workUnits ?? []).filter((unit) => unit.kind === "go_shard");
}

export function estimateServiceBackedGoCPULimit(workUnits) {
  const units = goShardUnits(workUnits);
  if (units.length === 0) {
    return 1;
  }
  const totalWeight = units.reduce(
    (sum, unit) => sum + Math.max(1, unit.weightMs ?? unit.weight_ms ?? 1),
    0,
  );
  const maxWeight = Math.max(
    ...units.map((unit) => Math.max(1, unit.weightMs ?? unit.weight_ms ?? 1)),
  );
  const weightedConcurrency = Math.ceil(
    totalWeight / Math.max(30_000, maxWeight),
  );
  const cpuCount = availableCPUCount();
  const hostConcurrency =
    cpuCount <= 4 ? Math.max(2, cpuCount - 1) : Math.floor(cpuCount * 0.75);
  return clampInteger(
    Math.max(4, Math.min(hostConcurrency, weightedConcurrency)),
    4,
    16,
  );
}

export function estimateServiceBackedGoIOLimit(workUnits, goCPULimit) {
  const units = goShardUnits(workUnits);
  if (units.length === 0) {
    return 1;
  }
  const profileCount = (profile) =>
    units.filter((unit) => unit.schedulerProfile === profile).length;
  const profileConcurrency =
    profileCount("balanced") +
    profileCount("transaction_heavy") +
    profileCount("io_heavy") * 2 +
    profileCount("clone_heavy") * 2 +
    profileCount("postgres_advisory_lock_exclusive") * 2 +
    profileCount("reset_heavy") * 2 +
    Math.ceil(profileCount("cpu_heavy") / 2);
  return clampInteger(Math.max(6, goCPULimit + 2, profileConcurrency), 6, 24);
}

export function schedulerAutoLimitResolvers(scheduler, provisionalUnits = []) {
  const schedulerKind = requireSchedulerFamily(scheduler, "scheduler");
  if (schedulerKind === "check") {
    return {
      host_cpu: () => estimateCheckHostCPULimit(),
      host_io: ({ resourceLimits: currentLimits }) =>
        estimateCheckHostIOLimit(currentLimits),
      host_process_slots: () => estimateHostProcessSlotLimit(),
      service_backed_browser_stack: ({ resourceLimits: currentLimits }) =>
        estimateBrowserStackAutoLimit(provisionalUnits, currentLimits, {
          cpuResources: [hostCPUResource],
        }),
      service_backed_postgres_clone: ({ resourceLimits: currentLimits }) =>
        estimatePostgresCloneAutoLimit(currentLimits, {
          cpuResources: [hostCPUResource],
          ioResources: [hostIOResource],
        }),
      service_backed_postgres_reset: ({ resourceLimits: currentLimits }) =>
        estimatePostgresResetAutoLimit(currentLimits, {
          ioResources: [hostIOResource],
        }),
    };
  }
  if (schedulerKind === "sequence") {
    return {
      host_cpu: () => estimateSequenceHostCPULimit(),
      host_io: ({ resourceLimits: currentLimits }) =>
        estimateSequenceHostIOLimit(currentLimits.get(hostCPUResource) ?? 1),
      host_process_slots: () => estimateSequenceProcessLimit(),
    };
  }
  if (schedulerKind !== "service_backed" && schedulerKind !== "test_slice") {
    throw new Error(`unsupported scheduler auto-limit family ${schedulerKind}`);
  }
  return {
    host_process_slots: () => estimateHostProcessSlotLimit(),
    service_backed_go_cpu: () =>
      estimateServiceBackedGoCPULimit(provisionalUnits),
    service_backed_go_io: ({ resourceLimits: currentLimits }) =>
      estimateServiceBackedGoIOLimit(
        provisionalUnits,
        currentLimits.get(goCPUResource),
      ),
    service_backed_browser_stack: ({ resourceLimits: currentLimits }) =>
      estimateBrowserStackAutoLimit(provisionalUnits, currentLimits, {
        cpuResources: [goCPUResource],
      }),
    service_backed_postgres_clone: ({ resourceLimits: currentLimits }) =>
      estimatePostgresCloneAutoLimit(currentLimits, {
        cpuResources: [goCPUResource],
        ioResources: [goIOResource],
      }),
    service_backed_postgres_reset: ({ resourceLimits: currentLimits }) =>
      estimatePostgresResetAutoLimit(currentLimits, {
        ioResources: [goIOResource],
      }),
  };
}

export function schedulerCapacityProfileLimits(
  scheduler,
  capacityProfile,
  label,
  { env = process.env } = {},
) {
  return normalizeResourceLimits({}, label, {
    scheduler: requireSchedulerFamily(scheduler, "scheduler"),
    capacityProfile,
    allowAuto: true,
    env,
  });
}

function unitResourceClaimEntries(unit) {
  const claims = unit.resourceClaims ?? unit.resource_claims;
  if (!claims) {
    return [];
  }
  if (typeof claims.entries === "function") {
    return Array.from(claims.entries());
  }
  if (typeof claims === "object" && !Array.isArray(claims)) {
    return Object.entries(claims);
  }
  return [];
}

function pruneResourceLimitsToClaims(resourceLimits, resourceLimitSources, workUnits) {
  const claimedResources = new Set();
  for (const unit of workUnits ?? []) {
    for (const [resource] of unitResourceClaimEntries(unit)) {
      claimedResources.add(resource);
    }
  }
  const prunedLimits = new Map();
  const prunedSources = new Map();
  for (const [resource, limit] of resourceLimits.entries()) {
    if (!claimedResources.has(resource)) {
      continue;
    }
    prunedLimits.set(resource, limit);
    if (resourceLimitSources.has(resource)) {
      prunedSources.set(resource, resourceLimitSources.get(resource));
    }
  }
  return { resourceLimits: prunedLimits, resourceLimitSources: prunedSources };
}

function maxPolicyResourceClaims(workUnits) {
  const claims = new Map();
  for (const unit of workUnits ?? []) {
    for (const [resource, amount] of unitResourceClaimEntries(unit)) {
      if (!Number.isInteger(amount) || amount >= Number.MAX_SAFE_INTEGER) {
        continue;
      }
      claims.set(resource, Math.max(claims.get(resource) ?? 0, amount));
    }
  }
  return claims;
}

export function resolveSchedulerResourceLimits({
  scheduler,
  resourceLimits,
  resourceLimitSources,
  label,
  workUnits,
  pruneToClaims = false,
}) {
  const schedulerKind = requireSchedulerFamily(scheduler, "scheduler");
  const resolved = resolveAutoResourceLimits(
    resourceLimits,
    resourceLimitSources,
    label,
    schedulerAutoLimitResolvers(schedulerKind, workUnits),
    maxPolicyResourceClaims(workUnits),
  );
  if (!pruneToClaims) {
    return resolved;
  }
  return pruneResourceLimitsToClaims(
    resolved.resourceLimits,
    resolved.resourceLimitSources,
    workUnits,
  );
}
