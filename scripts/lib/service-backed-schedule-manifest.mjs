import { serviceBackedScheduleSchemaID } from "./execution-topology.mjs";
import {
  readJsonObject,
  requireEnum,
  requireObject,
  requirePositiveInteger,
  requireSchemaID,
  requireString,
  validateObjectArray,
  validateObjectShape,
} from "./json-shape.mjs";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const serviceScheduleKeys = new Set(["schema_id", "generated", "schedules"]);
const serviceScheduleEntryKeys = new Set([
  "target",
  "capacity_profile",
  "resource_limits",
  "work_unit_sources",
]);
const serviceSourceKeys = new Set([
  "type",
  "class",
  "target",
  "needs",
  "weight",
  "resource_claims",
  "browser_stage",
]);
const serviceGeneratedKeys = new Set([
  "generator",
  "topology",
  "browser_batch_manifest",
  "make_target_duration_baseline",
]);

function manifestValue(fileOrManifest, label) {
  return typeof fileOrManifest === "string"
    ? readJsonObject(fileOrManifest, label)
    : requireObject(fileOrManifest, label);
}

export function validateServiceBackedScheduleManifestShape(
  fileOrManifest,
  label = fileOrManifest,
) {
  const manifest = manifestValue(fileOrManifest, label);
  validateObjectShape(manifest, label, { keys: serviceScheduleKeys });
  requireSchemaID(manifest, serviceBackedScheduleSchemaID, label);
  const generated = validateObjectShape(manifest.generated, `${label}.generated`, {
    keys: serviceGeneratedKeys,
  });
  for (const key of serviceGeneratedKeys) {
    requireString(generated[key], `${label}.generated.${key}`);
  }
  validateObjectArray(
    manifest.schedules,
    `${label}.schedules`,
    { nonEmpty: true, keys: serviceScheduleEntryKeys },
    (schedule, scheduleLabel) => {
      requireString(schedule.target, `${scheduleLabel}.target`, {
        pattern: makeTargetPattern,
      });
      requireString(schedule.capacity_profile, `${scheduleLabel}.capacity_profile`);
      validateObjectArray(
        schedule.work_unit_sources,
        `${scheduleLabel}.work_unit_sources`,
        { nonEmpty: true, keys: serviceSourceKeys },
        (source, sourceLabel) => {
          requireEnum(
            source.type,
            `${sourceLabel}.type`,
            new Set(["go_shards", "make_target"]),
          );
          requireEnum(
            source.class,
            `${sourceLabel}.class`,
            new Set(["backend", "browser"]),
          );
          requireString(source.target, `${sourceLabel}.target`, {
            pattern: makeTargetPattern,
          });
          if (source.type === "make_target") {
            requirePositiveInteger(source.weight, `${sourceLabel}.weight`);
          }
        },
      );
    },
  );
  return manifest;
}

export { serviceBackedScheduleSchemaID };
