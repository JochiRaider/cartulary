import { serviceBackedScheduleSchemaID } from "./execution-topology.mjs";
import {
  readJsonObject,
  requireEnum,
  requireInteger,
  requireObject,
  requirePositiveInteger,
  requireSchemaID,
  requireString,
  requireStringArray,
  validateObjectArray,
  validateObjectShape,
} from "./json-shape.mjs";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const serviceScheduleKeys = new Set(["schema_id", "generated", "schedules"]);
const serviceScheduleEntryKeys = new Set([
  "target",
  "scheduler_kind",
  "capacity_profile",
  "resource_limits",
  "work_unit_sources",
]);
const serviceSourceKeys = new Set([
  "type",
  "class",
  "target",
  "needs",
  "priority",
  "weight_ms",
  "resource_claims",
  "browser_stage",
  "groups",
]);
const serviceBrowserGroupKeys = new Set([
  "id",
  "name",
  "kind",
  "target",
  "aggregate_target",
  "coverage",
  "execution_dependency",
  "shard_name",
  "shard_index",
  "shard_count",
  "phases",
  "entry_ids",
  "priority",
  "weight_ms",
  "resource_claims",
]);
const browserGroupKinds = new Set(["functional_shard", "support", "stateful", "measurement", "visual"]);
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
      requireString(schedule.scheduler_kind, `${scheduleLabel}.scheduler_kind`);
      if (schedule.scheduler_kind !== "service_backed") {
        throw new Error(`${scheduleLabel}.scheduler_kind must be service_backed`);
      }
      requireString(schedule.capacity_profile, `${scheduleLabel}.capacity_profile`);
      validateObjectArray(
        schedule.work_unit_sources,
        `${scheduleLabel}.work_unit_sources`,
        { nonEmpty: true, keys: serviceSourceKeys },
        (source, sourceLabel) => {
          requireEnum(
            source.type,
            `${sourceLabel}.type`,
            new Set(["go_shards", "make_target", "browser_stage"]),
          );
          requireEnum(
            source.class,
            `${sourceLabel}.class`,
            new Set(["backend", "browser"]),
          );
          requireString(source.target, `${sourceLabel}.target`, {
            pattern: makeTargetPattern,
          });
          if (source.type === "make_target" || source.type === "browser_stage") {
            requirePositiveInteger(source.weight_ms, `${sourceLabel}.weight_ms`);
          }
          if (source.priority !== undefined) {
            requireInteger(source.priority, `${sourceLabel}.priority`, { min: 0 });
          }
          if (source.type === "browser_stage") {
            validateObjectArray(
              source.groups,
              `${sourceLabel}.groups`,
              { nonEmpty: true, keys: serviceBrowserGroupKeys },
              (group, groupLabel) => {
                requireString(group.id, `${groupLabel}.id`);
                requireString(group.name, `${groupLabel}.name`);
                requireEnum(group.kind, `${groupLabel}.kind`, browserGroupKinds);
                requireString(group.target, `${groupLabel}.target`, {
                  pattern: makeTargetPattern,
                });
                requireString(group.aggregate_target, `${groupLabel}.aggregate_target`, {
                  pattern: makeTargetPattern,
                });
                requirePositiveInteger(group.weight_ms, `${groupLabel}.weight_ms`);
                if (group.priority !== undefined) {
                  requireInteger(group.priority, `${groupLabel}.priority`, { min: 0 });
                }
                if (group.kind === "functional_shard") {
                  requireString(group.shard_name, `${groupLabel}.shard_name`);
                  requireInteger(group.shard_index, `${groupLabel}.shard_index`, { min: 0 });
                  requirePositiveInteger(group.shard_count, `${groupLabel}.shard_count`);
                  requireStringArray(group.entry_ids, `${groupLabel}.entry_ids`, { nonEmpty: true });
                }
              },
            );
          }
        },
      );
    },
  );
  return manifest;
}

export { serviceBackedScheduleSchemaID };
