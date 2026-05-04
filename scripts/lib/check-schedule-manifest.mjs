import { checkScheduleSchemaID } from "./execution-topology.mjs";
import {
  readJsonObject,
  requireObject,
  requirePositiveInteger,
  requireSchemaID,
  requireString,
  validateObjectArray,
  validateObjectShape,
} from "./json-shape.mjs";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const envNamePattern = /^[A-Z][A-Z0-9_]*$/;
const checkScheduleKeys = new Set(["schema_id", "schedules"]);
const checkScheduleEntryKeys = new Set([
  "target",
  "capacity_profile",
  "resource_limits",
  "summary_groups",
  "work_units",
]);
const checkWorkUnitKeys = new Set([
  "id",
  "kind",
  "target",
  "label",
  "aggregate_target",
  "weight",
  "needs",
  "produces_summary_targets",
  "completion_keys",
  "failure_keys",
  "running_dependency_keys",
  "resource_claims",
  "retained_resource_claims",
  "release_retained_resource_claims",
  "make_jobs",
  "env",
  "service_session",
  "browser_stage",
  "browser_group",
  "shard",
  "shard_names",
  "scheduler_profile",
  "count_in_total",
  "counts_started",
  "complete_on_failure",
  "unblock_label",
]);

function manifestValue(fileOrManifest, label) {
  return typeof fileOrManifest === "string"
    ? readJsonObject(fileOrManifest, label)
    : requireObject(fileOrManifest, label);
}

export function validateCheckScheduleManifestShape(fileOrManifest, label = fileOrManifest) {
  const manifest = manifestValue(fileOrManifest, label);
  validateObjectShape(manifest, label, { keys: checkScheduleKeys });
  requireSchemaID(manifest, checkScheduleSchemaID, label);
  validateObjectArray(
    manifest.schedules,
    `${label}.schedules`,
    { nonEmpty: true, keys: checkScheduleEntryKeys },
    (schedule, scheduleLabel) => {
      requireString(schedule.target, `${scheduleLabel}.target`, {
        pattern: makeTargetPattern,
      });
      requireString(schedule.capacity_profile, `${scheduleLabel}.capacity_profile`);
      validateObjectArray(
        schedule.work_units,
        `${scheduleLabel}.work_units`,
        { nonEmpty: true, keys: checkWorkUnitKeys },
        (unit, unitLabel) => {
          requireString(unit.target, `${unitLabel}.target`, {
            pattern: makeTargetPattern,
          });
          requirePositiveInteger(unit.weight, `${unitLabel}.weight`);
          if (unit.env !== undefined) {
            for (const name of Object.keys(requireObject(unit.env, `${unitLabel}.env`))) {
              requireString(name, `${unitLabel}.env key`, {
                pattern: envNamePattern,
              });
            }
          }
        },
      );
    },
  );
  return manifest;
}

export { checkScheduleSchemaID };
