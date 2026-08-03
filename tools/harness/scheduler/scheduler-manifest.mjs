import { readFileSync } from "node:fs";

const schemaID = "cartulary.scheduler_manifest.v3";

export function validateSchedulerManifestShape(fileOrValue) {
  const value =
    typeof fileOrValue === "string"
      ? JSON.parse(readFileSync(fileOrValue, "utf8"))
      : fileOrValue;
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error("scheduler manifest must be an object");
  }
  const keys = Object.keys(value).sort();
  if (JSON.stringify(keys) !== JSON.stringify(["generated", "schedules", "schema_id"])) {
    throw new Error("scheduler manifest has unexpected fields");
  }
  if (value.schema_id !== schemaID) {
    throw new Error(`scheduler manifest must use ${schemaID}`);
  }
  if (!value.generated || typeof value.generated !== "object" || Array.isArray(value.generated)) {
    throw new Error("scheduler manifest generated metadata must be an object");
  }
  if (!Array.isArray(value.schedules) || value.schedules.length !== 0) {
    throw new Error("scheduler manifest v3 forbids legacy authored schedules");
  }
  return value;
}
