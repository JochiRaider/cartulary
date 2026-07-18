import {
  serviceBackedScheduleSchemaID,
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
} from "../../generated-artifacts/contracts/index.mjs";
import { normalizeRuntimeBinaryEntries } from "../../runtime-binary-registry.mjs";
import { requireSchedulerCapacityProfileForFamily } from "../../scheduler/scheduler-family-contract.mjs";

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const serviceScheduleKeys = new Set(["schema_id", "generated", "schedules"]);
const serviceScheduleEntryKeys = new Set([
  "target",
  "scheduler_kind",
  "capacity_profile",
  "service_complete_priority",
  "resource_limits",
  "work_unit_sources",
]);
const serviceSourceKeys = new Set([
  "type",
  "class",
  "target",
  "needs",
  "env",
  "priority",
  "weight_ms",
  "resource_claims",
  "resource_claims_by_execution_family",
  "default_check_required",
  "runtime_binary_records",
  "runtime_binaries",
  "browser_stage",
  "browser_session_group",
  "browser_session_isolation_reason",
  "runtime_profile_id",
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
  "env",
  "workers",
  "selected_phase",
  "selected_row_ids",
  "browser_session_group",
  "browser_session_isolation_reason",
  "runtime_profile_id",
  "priority",
  "weight_ms",
  "resource_claims",
]);
const browserGroupKinds = new Set([
  "functional_shard",
  "support",
  "stateful",
  "stateful_partition",
  "measurement",
  "visual",
  "a11y",
  "duration_balanced_specs",
]);
const serviceGeneratedKeys = new Set([
  "generator",
  "topology",
  "browser_batch_manifest",
  "make_target_duration_baseline",
]);

function requireBoolean(value, label) {
  if (typeof value !== "boolean") {
    throw new Error(`${label} must be a boolean`);
  }
  return value;
}

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
      requireSchedulerCapacityProfileForFamily(
        requireString(schedule.capacity_profile, `${scheduleLabel}.capacity_profile`),
        "service_backed",
        scheduleLabel,
      );
      if (schedule.service_complete_priority !== undefined) {
        requirePositiveInteger(
          schedule.service_complete_priority,
          `${scheduleLabel}.service_complete_priority`,
        );
      }
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
          if (source.default_check_required !== undefined) {
            requireBoolean(
              source.default_check_required,
              `${sourceLabel}.default_check_required`,
            );
          }
          if (source.env !== undefined) {
            const env = requireObject(source.env, `${sourceLabel}.env`);
            for (const [name, value] of Object.entries(env)) {
              requireString(name, `${sourceLabel}.env key`);
              requireString(value, `${sourceLabel}.env.${name}`);
            }
          }
          if (source.runtime_binary_records !== undefined) {
            normalizeRuntimeBinaryEntries(
              validateObjectArray(
                source.runtime_binary_records,
                `${sourceLabel}.runtime_binary_records`,
                { nonEmpty: true },
                (record) => record,
              ),
              { label: `${sourceLabel}.runtime_binary_records` },
            );
          }
          if (source.runtime_binaries !== undefined) {
            requireStringArray(source.runtime_binaries, `${sourceLabel}.runtime_binaries`, {
              nonEmpty: true,
            });
          }
          if (source.type === "browser_stage") {
            if (source.browser_session_group !== undefined) {
              requireString(
                source.browser_session_group,
                `${sourceLabel}.browser_session_group`,
              );
            }
            if (source.browser_session_isolation_reason !== undefined) {
              requireString(
                source.browser_session_isolation_reason,
                `${sourceLabel}.browser_session_isolation_reason`,
              );
            }
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
                if (group.workers !== undefined) {
                  requireString(group.workers, `${groupLabel}.workers`);
                }
                if (group.selected_phase !== undefined) {
                  requireString(group.selected_phase, `${groupLabel}.selected_phase`);
                }
                if (group.selected_row_ids !== undefined) {
                  requireStringArray(group.selected_row_ids, `${groupLabel}.selected_row_ids`, {
                    nonEmpty: true,
                  });
                }
                if (group.env !== undefined) {
                  const env = requireObject(group.env, `${groupLabel}.env`);
                  for (const [name, value] of Object.entries(env)) {
                    requireString(name, `${groupLabel}.env key`);
                    requireString(value, `${groupLabel}.env.${name}`);
                  }
                }
                if (group.browser_session_group !== undefined) {
                  requireString(group.browser_session_group, `${groupLabel}.browser_session_group`);
                }
                if (group.browser_session_isolation_reason !== undefined) {
                  requireString(
                    group.browser_session_isolation_reason,
                    `${groupLabel}.browser_session_isolation_reason`,
                  );
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
