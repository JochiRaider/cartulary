import { accessSync, constants, readFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import { semanticJSONDigest } from "../../contract/index.mjs";
import { loadSchedulerResourceRegistry } from "../scheduler-resources.mjs";

function readPositiveInteger(file) {
  try {
    const value = readFileSync(file, "utf8").trim();
    if (value === "max") return null;
    const parsed = Number.parseInt(value, 10);
    return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : null;
  } catch {
    return null;
  }
}

// The effective process limit is the smaller of the soft per-user limit and a
// positive cgroup pids.max; null when neither is available.
function processLimit() {
  const cgroup = readPositiveInteger("/sys/fs/cgroup/pids.max");
  const user = userProcessLimit();
  if (cgroup === null) return user;
  return user === null ? cgroup : Math.min(cgroup, user);
}

function userProcessLimit() {
  try {
    const line = readFileSync("/proc/self/limits", "utf8")
      .split("\n")
      .find((entry) => entry.startsWith("Max processes"));
    if (!line) return null;
    const value = line.trim().split(/\s+/u)[2];
    if (value === "unlimited") return null;
    const parsed = Number.parseInt(value, 10);
    return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : null;
  } catch {
    return null;
  }
}

function parseOverride(value, root) {
  if (value === undefined || value === "") return null;
  let override;
  try {
    override = typeof value === "string"
      ? JSON.parse(readFileSync(path.resolve(root ?? process.cwd(), value), "utf8"))
      : value;
  } catch (error) {
    throw new Error(`invalid capacity override declaration: ${error.message}`);
  }
  validateSchemaSync("cartulary.harness_capacity_override.v1", override);
  return override;
}

export function cpuCapacityWithSafetyMargin(cpuTokens, policy) {
  const minimum = policy.minimum;
  const marginPercent = policy.safety_margin_percent ?? 0;
  if (
    !Number.isSafeInteger(cpuTokens) ||
    cpuTokens < 1 ||
    !Number.isSafeInteger(minimum) ||
    minimum < 1 ||
    !Number.isSafeInteger(marginPercent) ||
    marginPercent < 0 ||
    marginPercent > 100
  ) {
    throw new Error("invalid CPU capacity safety-margin inputs");
  }
  return Math.max(
    minimum,
    Math.floor((cpuTokens * 100) / (100 + marginPercent)),
  );
}

// This bounds worker planning inside an admitted parent without rewriting the
// physical host snapshot. Native admission independently validates the lease.
export function availableGoWorkerBudget(cpuTokens, environment) {
  if (!Number.isSafeInteger(cpuTokens) || cpuTokens < 1) throw new Error("invalid Go CPU budget");
  if (environment.CARTULARY_HOST_ADMISSION_LEASE === undefined) return cpuTokens;
  const inherited = environment.CARTULARY_UNIT_CPU_TOKENS;
  if (!environment.CARTULARY_HOST_ADMISSION_LEASE || typeof inherited !== "string" ||
      !/^[1-9][0-9]*$/u.test(inherited) || !Number.isSafeInteger(Number(inherited))) {
    throw new Error("invalid inherited Go CPU budget");
  }
  return Math.min(cpuTokens, Number(inherited));
}

export function retainedGoWorkerBudget(manifest) {
  const budget = manifest.go_worker_budget, capacity = manifest.capability_snapshot?.cpu_tokens;
  if (!Number.isSafeInteger(capacity) || capacity < 1 ||
      !Number.isSafeInteger(budget) || budget < 1 || budget > capacity) {
    throw new Error("invalid retained Go worker budget");
  }
  return budget;
}

// Process slots bound operating-system processes, including idle resident
// service processes. CPU contention is bounded separately by cpu_tokens.
export function processSlotsFor(policy, cpuTokens, detectedProcessLimit) {
  if (detectedProcessLimit === null || detectedProcessLimit === undefined) {
    return Math.max(policy.minimum, cpuTokens * policy.fallback_cpu_multiplier);
  }
  return Math.max(policy.minimum, Math.floor((detectedProcessLimit * (100 - policy.safety_margin_percent)) / 100));
}

export function automaticCapacities({ registry, cpuTokens, detectedMemory, detectedProcessLimit }) {
  const policies = registry.capacity_policies;
  const fixed = (field) => {
    const policy = policies[field];
    return Math.max(
      policy.minimum,
      Math.min(policy.default, cpuTokens * policy.maximum_cpu_multiplier),
    );
  };
  return {
    cpu_tokens: Math.max(policies.cpu_tokens.minimum, cpuTokens),
    memory_bytes: Math.max(policies.memory_bytes.minimum, detectedMemory),
    process_slots: processSlotsFor(policies.process_slots, cpuTokens, detectedProcessLimit),
    io_tokens: Math.max(
      policies.io_tokens.minimum,
      cpuTokens * policies.io_tokens.cpu_multiplier,
    ),
    postgres_lanes: fixed("postgres_lanes"),
    // At most the PostgreSQL lanes; the caller bounds it by the resolved value.
    postgres_heavy_lanes: Math.max(
      policies.postgres_heavy_lanes.minimum,
      Math.min(policies.postgres_heavy_lanes.default, fixed("postgres_lanes")),
    ),
    object_store_lanes: fixed("object_store_lanes"),
    port_lanes: fixed("port_lanes"),
  };
}

function resolveIntegerCapacity({ field, override, automatic, maximum }) {
  const value = override?.[field] ?? automatic;
  if (!Number.isSafeInteger(value) || value < 1 || value > maximum) {
    throw new Error(
      `capacity override ${field}=${String(value)} exceeds the detected policy bound ${maximum}`,
    );
  }
  return value;
}

export function captureCapabilitySnapshot({
  root,
  override: overrideInput,
  services = {},
} = {}) {
  const detectedCPUTokens = Math.max(
    1,
    os.availableParallelism?.() ?? os.cpus().length ?? 1,
  );
  const cgroupMemory = readPositiveInteger("/sys/fs/cgroup/memory.max");
  const hostMemory = Math.max(1, os.totalmem());
  const detectedMemory = cgroupMemory ? Math.min(cgroupMemory, hostMemory) : hostMemory;
  const detectedProcessLimit = processLimit();
  const override = parseOverride(overrideInput, root);
  const registry = loadSchedulerResourceRegistry(
    root ? path.join(root, "tools/scheduler_resource_registry.json") : undefined,
  );
  const cpuTokens = cpuCapacityWithSafetyMargin(
    detectedCPUTokens,
    registry.capacity_policies.cpu_tokens,
  );
  const automatic = automaticCapacities({
    registry,
    cpuTokens,
    detectedMemory,
    detectedProcessLimit,
  });
  const resolvedCPU = resolveIntegerCapacity({
    field: "cpu_tokens",
    override,
    automatic: automatic.cpu_tokens,
    maximum: automatic.cpu_tokens,
  });
  const resolvedMemory = resolveIntegerCapacity({
    field: "memory_bytes",
    override,
    automatic: automatic.memory_bytes,
    maximum: automatic.memory_bytes,
  });
  const boundedAutomatic = automaticCapacities({
    registry,
    cpuTokens: resolvedCPU,
    detectedMemory,
    detectedProcessLimit,
  });
  const processMaximum = processSlotsFor(registry.capacity_policies.process_slots, resolvedCPU, detectedProcessLimit);
  const ioMaximum = resolvedCPU * registry.capacity_policies.io_tokens.cpu_multiplier;
  const fixedMaximum = (field) =>
    resolvedCPU * registry.capacity_policies[field].maximum_cpu_multiplier;
  const postgresLanes = resolveIntegerCapacity({
    field: "postgres_lanes",
    override,
    automatic: boundedAutomatic.postgres_lanes,
    maximum: fixedMaximum("postgres_lanes"),
  });
  let writableVolume = true;
  try {
    accessSync(root ?? process.cwd(), constants.W_OK);
  } catch {
    writableVolume = false;
  }
  const snapshot = {
    schema_id: "cartulary.harness_capability_snapshot.v1",
    cpu_tokens: resolvedCPU,
    memory_bytes: resolvedMemory,
    process_slots: resolveIntegerCapacity({
      field: "process_slots",
      override,
      automatic: boundedAutomatic.process_slots,
      maximum: processMaximum,
    }),
    io_tokens: resolveIntegerCapacity({
      field: "io_tokens",
      override,
      automatic: boundedAutomatic.io_tokens,
      maximum: ioMaximum,
    }),
    postgres_lanes: postgresLanes,
    postgres_heavy_lanes: resolveIntegerCapacity({
      field: "postgres_heavy_lanes",
      override,
      automatic: Math.min(boundedAutomatic.postgres_heavy_lanes, postgresLanes),
      maximum: postgresLanes,
    }),
    object_store_lanes: resolveIntegerCapacity({
      field: "object_store_lanes",
      override,
      automatic: boundedAutomatic.object_store_lanes,
      maximum: fixedMaximum("object_store_lanes"),
    }),
    writable_volume: override?.writable_volume ?? writableVolume,
    port_lanes: resolveIntegerCapacity({
      field: "port_lanes",
      override,
      automatic: boundedAutomatic.port_lanes,
      maximum: fixedMaximum("port_lanes"),
    }),
    services: { ...services, ...(override?.services ?? {}) },
    sources: {
      cpu_tokens: override?.cpu_tokens === undefined ? "host" : "override",
      io_tokens: override?.io_tokens === undefined ? "host" : "override",
      memory_bytes:
        override?.memory_bytes === undefined
          ? cgroupMemory
            ? "cgroup"
            : "host"
          : "override",
      object_store_lanes:
        override?.object_store_lanes === undefined ? "fallback" : "override",
      port_lanes: override?.port_lanes === undefined ? "host" : "override",
      postgres_heavy_lanes:
        override?.postgres_heavy_lanes === undefined ? "fallback" : "override",
      postgres_lanes:
        override?.postgres_lanes === undefined ? "fallback" : "override",
      process_slots:
        override?.process_slots === undefined
          ? detectedProcessLimit
            ? "host"
            : "fallback"
          : "override",
      services: override?.services === undefined ? "host" : "override",
      writable_volume:
        override?.writable_volume === undefined ? "host" : "override",
    },
    snapshot_digest: "",
  };
  const semantic = { ...snapshot };
  delete semantic.snapshot_digest;
  snapshot.snapshot_digest = semanticJSONDigest(semantic);
  validateSchemaSync(snapshot.schema_id, snapshot);
  return snapshot;
}

export function resourceCapacities(snapshot) {
  return new Map([
    ["browser_stack", snapshot.services.browser === false ? 0 : snapshot.port_lanes],
    ["cpu", snapshot.cpu_tokens],
    ["io", snapshot.io_tokens],
    ["memory_mb", Math.max(1, Math.floor(snapshot.memory_bytes / 1048576))],
    ["object_store", snapshot.services.object_store === false ? 0 : snapshot.object_store_lanes],
    ["port_lane", snapshot.port_lanes],
    ["postgres", snapshot.services.postgres === false ? 0 : snapshot.postgres_lanes],
    // Capture always resolves the authored lanes (minimum one). A historical
    // snapshot recorded before the lane existed is read as recorded, without it.
    ...(snapshot.postgres_heavy_lanes === undefined ? [] : [["postgres_heavy", snapshot.postgres_heavy_lanes]]),
    ["process", snapshot.process_slots],
    ["service_stack", snapshot.services.service_stack === false ? 0 : snapshot.port_lanes],
    ["volume", snapshot.writable_volume ? 1 : 0],
  ]);
}
