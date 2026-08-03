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
    JSON.stringify(["capacity_override_schema", "resources", "runner_worker_inputs", "schema_id"])
  ) {
    throw new Error(`${label} has unexpected fields`);
  }
  if (registry.schema_id !== "cartulary.scheduler_resource_registry.v5") {
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
    JSON.stringify(["PLAYWRIGHT_WORKERS", "VITEST_MAX_WORKERS"])
  ) {
    throw new Error(`${label} has unsupported runner worker inputs`);
  }
  return registry;
}

export function loadSchedulerResourceRegistry(file = defaultFile) {
  return validateSchedulerResourceRegistrySemantics(file, file);
}
