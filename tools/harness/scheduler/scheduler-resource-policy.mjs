const goCPUResource = "go_cpu";
const goIOResource = "go_io";
const hostCPUResource = "host_cpu";
const hostIOResource = "host_io";
const postgresResetResource = "postgres_reset";
const postgresCloneResource = "postgres_clone";
const schedulerResourceFamilies = new Set(["check", "service_backed", "phase_slice"]);

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

export function resourceClaimsObject(value) {
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
  if (!schedulerResourceFamilies.has(scheduler)) {
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
