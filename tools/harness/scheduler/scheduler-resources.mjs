import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const defaultFile = path.join(root, "tools/scheduler_resource_registry.json");
const expectedResources = [
  "browser_stack",
  "cpu",
  "io",
  "memory_mb",
  "object_store",
  "port_lane",
  "postgres",
  "process",
  "service_stack",
  "volume",
];

function value(fileOrValue) {
  return typeof fileOrValue === "string"
    ? JSON.parse(readFileSync(fileOrValue, "utf8"))
    : fileOrValue;
}

export function validateSchedulerResourceRegistrySemantics(
  fileOrValue,
  label = "scheduler resource registry",
) {
  const registry = value(fileOrValue);
  const keys = Object.keys(registry ?? {}).sort();
  if (
    JSON.stringify(keys) !==
    JSON.stringify(["capacity_override_schema", "capacity_policies", "resources", "runner_worker_inputs", "schema_id"])
  ) {
    throw new Error(`${label} has unexpected fields`);
  }
  if (registry.schema_id !== "cartulary.scheduler_resource_registry.v7") {
    throw new Error(`${label} has the wrong schema ID`);
  }
  if (registry.capacity_override_schema !== "cartulary.harness_capacity_override.v1") {
    throw new Error(`${label} must bind the canonical capacity override`);
  }
  if (
    !Array.isArray(registry.resources) ||
    JSON.stringify(registry.resources.map((entry) => entry.name)) !==
      JSON.stringify(expectedResources)
  ) {
    throw new Error(`${label} must declare the sorted canonical resource roster`);
  }
  const expectedPolicyFields = [
    "cpu_tokens",
    "io_tokens",
    "memory_bytes",
    "object_store_lanes",
    "port_lanes",
    "postgres_lanes",
    "process_slots",
    "writable_volume",
  ];
  if (
    JSON.stringify(Object.keys(registry.capacity_policies ?? {}).sort()) !==
    JSON.stringify(expectedPolicyFields)
  ) {
    throw new Error(`${label} must declare the closed capacity policy roster`);
  }
  const expectedSources = {
    cpu_tokens: "host_parallelism",
    io_tokens: "cpu_multiple",
    memory_bytes: "host_memory_cgroup_bounded",
    object_store_lanes: "fixed_cpu_bounded",
    port_lanes: "fixed_cpu_bounded",
    postgres_lanes: "fixed_cpu_bounded",
    process_slots: "process_limit_bounded_cpu_multiple",
    writable_volume: "writable_root",
  };
  for (const field of expectedPolicyFields) {
    if (registry.capacity_policies[field]?.source !== expectedSources[field]) {
      throw new Error(`${label}.capacity_policies.${field} has an unsupported source`);
    }
  }
  const cpuSafetyMargin = registry.capacity_policies.cpu_tokens.safety_margin_percent ?? 0;
  if (
    !Number.isSafeInteger(cpuSafetyMargin) ||
    cpuSafetyMargin < 0 ||
    cpuSafetyMargin > 100
  ) {
    throw new Error(
      `${label}.capacity_policies.cpu_tokens.safety_margin_percent must be an integer from 0 through 100`,
    );
  }
  for (const field of ["postgres_lanes", "object_store_lanes", "port_lanes"]) {
    const policy = registry.capacity_policies[field];
    if (!Number.isSafeInteger(policy.default) || policy.default < 1 || policy.default > 1024 ||
        policy.minimum !== 1 || !Number.isSafeInteger(policy.maximum_cpu_multiplier) ||
        policy.maximum_cpu_multiplier < 1 || policy.maximum_cpu_multiplier > 64) {
      throw new Error(`${label}.capacity_policies.${field} must declare bounded positive integral capacity`);
    }
  }
  for (const resource of registry.resources) {
    if (
      Object.keys(resource).sort().join(",") !== "name,snapshot_field,unit" ||
      typeof resource.snapshot_field !== "string" ||
      typeof resource.unit !== "string"
    ) {
      throw new Error(`${label} resource ${resource.name ?? "unknown"} is invalid`);
    }
  }
  if (
    JSON.stringify(registry.runner_worker_inputs) !==
    JSON.stringify(["VITEST_MAX_WORKERS"])
  ) {
    throw new Error(`${label} has unsupported runner worker inputs`);
  }
  return registry;
}

export function loadSchedulerResourceRegistry(file = defaultFile) {
  return validateSchedulerResourceRegistrySemantics(file, file);
}
