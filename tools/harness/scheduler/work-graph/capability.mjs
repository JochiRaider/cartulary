import { accessSync, constants, readFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import { semanticJSONDigest } from "../../test-catalog/index.mjs";

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

function processLimit() {
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

export function captureCapabilitySnapshot({
  root,
  override: overrideInput,
  services = {},
} = {}) {
  const cpuTokens = Math.max(1, os.availableParallelism?.() ?? os.cpus().length ?? 1);
  const cgroupMemory = readPositiveInteger("/sys/fs/cgroup/memory.max");
  const hostMemory = Math.max(1, os.totalmem());
  const detectedMemory = cgroupMemory ? Math.min(cgroupMemory, hostMemory) : hostMemory;
  const detectedProcessLimit = processLimit();
  const override = parseOverride(overrideInput, root);
  let writableVolume = true;
  try {
    accessSync(root ?? process.cwd(), constants.W_OK);
  } catch {
    writableVolume = false;
  }
  const snapshot = {
    schema_id: "cartulary.harness_capability_snapshot.v1",
    cpu_tokens: override?.cpu_tokens ?? cpuTokens,
    memory_bytes: override?.memory_bytes ?? detectedMemory,
    process_slots:
      override?.process_slots ??
      Math.max(1, Math.min(detectedProcessLimit ?? cpuTokens * 4, cpuTokens * 4)),
    io_tokens: override?.io_tokens ?? Math.max(1, cpuTokens * 2),
    postgres_lanes: override?.postgres_lanes ?? Math.max(1, Math.min(cpuTokens, 8)),
    object_store_lanes:
      override?.object_store_lanes ?? Math.max(1, Math.min(cpuTokens, 4)),
    writable_volume: override?.writable_volume ?? writableVolume,
    port_lanes: override?.port_lanes ?? Math.max(1, Math.min(cpuTokens, 4)),
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
    ["process", snapshot.process_slots],
    ["service_stack", snapshot.services.service_stack === false ? 0 : snapshot.port_lanes],
    ["volume", snapshot.writable_volume ? 1 : 0],
  ]);
}
