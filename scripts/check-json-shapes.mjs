#!/usr/bin/env node
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  browserBatchManifestSchemaID,
  loadBrowserBatchManifest,
} from "./lib/browser-batch-manifest.mjs";
import {
  checkScheduleSchemaID,
  defaultExecutionTopologyManifestPath,
  executionTopologySchemaID,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  browserBatchManifestSchemaID as renderedBrowserBatchSchemaID,
  renderTaskSurfaceManifest,
  serviceBackedScheduleSchemaID,
  taskSurfaceSchemaID,
} from "./lib/execution-topology.mjs";
import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireArray,
  requireEnum,
  requireInteger,
  requireObject,
  requireObjectArray,
  requirePositiveInteger,
  requireRepoRelativePath,
  requireSchemaID,
  requireString,
  requireStringArray,
} from "./lib/json-shape.mjs";
import {
  loadPhasePolicyExceptions,
  validateManifest as validatePhaseManifestSemantics,
} from "./lib/phase-manifest.mjs";
import {
  activePhaseRegistryEntries,
  phaseRegistrySchemaID,
  validatePhaseRegistry,
} from "./lib/phase-registry.mjs";
import { loadSchedulerResourceRegistry } from "./lib/scheduler-resources.mjs";
import { validateServiceBackedScheduleTopology } from "./lib/service-backed-schedule-topology.mjs";
import {
  collectTaskSurfaceManifestErrors,
  loadTaskSurfaceManifest,
} from "./lib/task-surface.mjs";
import { renderServiceBackedScheduleManifest } from "./render-service-backed-schedule-manifest.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const phaseTestMapSchemaID = "cartulary.phase_test_map.v1";
const phasePolicyExceptionsSchemaID = "cartulary.phase_policy_exceptions.v1";
const generatedArtifactPolicySchemaID =
  "cartulary.generated_artifact_policy.v1";
const frontendImportBoundariesSchemaID =
  "cartulary.frontend_import_boundaries.v2";
const schedulerResourceRegistrySchemaID =
  "cartulary.scheduler_resource_registry.v3";
const bootstrapAdminSchemaID = "cartulary.bootstrap_admin.v1";
const serviceBackedMakeTargetBaselineSchemaID =
  "cartulary.service_backed_make_target_duration_baselines.v1";
const toolRunSummarySchemaID = "cartulary.tool_run_summary.v1";

const phaseStatusValues = new Set(["active", "planned", "retired"]);
const phaseNamePattern = /^phase(?:0|[1-9]\d*)$/;
const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const snakeIDPattern = /^[a-z][a-z0-9_]*$/;
const envNamePattern = /^[A-Z][A-Z0-9_]*$/;
const phaseManifestTopLevelKeys = new Set([
  "schema_id",
  "phase",
  "note",
  "ledger",
  "expected_ids",
  "forbidden_id_files",
  "support_go_targets",
  "unit",
  "integration",
  "e2e",
]);
const phaseManifestRequiredKeys = new Set([
  "note",
  "ledger",
  "expected_ids",
  "support_go_targets",
  "unit",
  "integration",
  "e2e",
]);
const phaseLedgerKeys = new Set([
  "title",
  "notes",
  "authoritative_execution",
  "support_execution_extras",
  "sections",
  "shared_harness",
  "support_only",
]);
const phaseMapEntryKeys = new Set([
  "id",
  "coverage",
  "runner",
  "package",
  "file",
  "symbol",
  "symbols",
  "title",
  "execution_dependency",
  "evidence_layer",
  "claim",
  "out_of_scope",
  "execution_family",
  "execution_label",
  "fixture_policy",
  "fixture_budget",
  "template_clone_reason",
  "migration_scratch_reason",
]);
const phaseSupportEntryKeys = new Set([
  "target",
  "section",
  "package",
  "file",
  "symbol",
  "symbols",
  "selection_pattern",
  "execution_family",
  "execution_label",
  "fixture_policy",
  "fixture_budget",
  "migration_scratch_reason",
]);
const topologyTopLevelKeys = new Set([
  "schema_id",
  "generated_outputs",
  "execution_dependencies",
  "go_targets",
  "task_surface",
  "check_schedules",
  "service_backed_schedules",
  "browser_e2e_batch",
]);
const generatedArtifactPolicyKeys = new Set([
  "schema_id",
  "ignored_sentinel_filenames",
  "generated_roots",
  "generated_files",
  "lint_scope_checks",
]);
const frontendBoundaryKeys = new Set([
  "schema_id",
  "scan_roots",
  "scan_excludes",
  "rules",
]);
const schedulerResourceRegistryKeys = new Set([
  "schema_id",
  "resources",
  "templates",
  "capacity_profiles",
  "forwarding_profiles",
]);
const bootstrapAdminKeys = new Set([
  "bootstrap_schema_id",
  "bootstrap_artifact_id",
  "email",
  "display_name",
  "initial_password",
]);
const toolRunSummaryKeys = new Set([
  "schema_id",
  "target",
  "command",
  "status",
  "exit_code",
  "started_at",
  "completed_at",
  "duration_ms",
  "output_mode",
  "artifact_root",
  "summary_artifacts",
  "log_artifacts",
  "work_units",
  "evidence_targets",
  "helper_units",
  "counts",
  "phase_accounting",
  "failure_class",
  "failure_origin",
  "failures",
  "slowest",
  "warnings",
  "rerun_commands",
  "extensions",
]);
const toolRunCommandKeys = new Set(["cwd", "argv", "make_target", "env"]);
const toolRunArtifactKeys = new Set(["role", "kind", "path"]);
const toolRunStatusValues = new Set(["pass", "fail"]);
const toolRunOutputModes = new Set([
  "summary",
  "ci",
  "verbose",
  "debug",
  "machine",
]);
const toolRunFailureClasses = new Set([
  "test",
  "infra",
  "helper",
  "timing",
  "artifact",
]);
const toolRunFailureOrigins = new Set([
  "product",
  "infrastructure",
  "helper",
  "timing",
  "artifact",
]);
const toolRunCountKeys = new Set([
  "phases",
  "tests",
  "failed",
  "non_test",
  "non_test_failed",
  "packages",
]);
const toolRunPhaseAccountingKeys = new Set([
  "authoritative",
  "support",
  "raw",
  "tooling_support",
  "unowned_regression",
  "unmapped",
  "authoritative_failed",
  "support_failed",
  "raw_failed",
  "tooling_support_failed",
  "unowned_regression_failed",
  "unmapped_failed",
  "missing",
]);
const toolRunWorkUnitKeys = new Set([
  "id",
  "completed",
  "total",
  "aborted_after",
  "status",
]);
const toolRunEvidenceTargetKeys = new Set([
  "target",
  "status",
  "artifact_root",
]);
const toolRunHelperUnitKeys = new Set(["target", "status", "artifact_root"]);
const toolRunSlowestKeys = new Set(["id", "duration_ms", "kind"]);
const rfc3339TimestampPattern =
  /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/u;
const checkScheduleKeys = new Set(["schema_id", "schedules"]);
const checkScheduleEntryKeys = new Set([
  "target",
  "capacity_profile",
  "resource_limits",
  "summary_groups",
  "work_units",
]);
const checkWorkUnitKeys = new Set([
  "target",
  "weight",
  "needs",
  "produces_summary_targets",
  "resource_claims",
  "make_jobs",
  "env",
  "nested_scheduler",
]);
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
const browserBatchKeys = new Set(["schema_id", "stages"]);
const browserStageKeys = new Set([
  "name",
  "target",
  "schedule_tags",
  "scheduler_dependency_policy",
  "summary_children",
  "groups",
]);
const browserGroupKeys = new Set([
  "name",
  "kind",
  "target",
  "project",
  "workers",
  "config",
  "coverage",
  "execution_dependency",
  "dependency_target",
  "reset_before",
  "specs",
]);
const generatedArtifactEntryKeys = new Set([
  "path",
  "allowed_extensions",
  "required_marker",
]);
const lintScopeCheckKeys = new Set([
  "shell_sources",
  "biome",
  "frontend_import_boundaries",
]);
const lintShellSourceKeys = new Set([
  "path",
  "must_contain",
  "must_not_contain",
]);
const lintBiomeKeys = new Set([
  "path",
  "required_files_includes",
  "forbidden_files_includes",
  "required_override_includes",
]);
const lintFrontendBoundaryKeys = new Set([
  "path",
  "required_scan_excludes",
  "required_restricted_paths",
]);
const frontendBoundaryRuleKeys = new Set([
  "id",
  "level",
  "message",
  "applies_to",
  "allowed_importers",
  "restricted_imports",
]);
const frontendBoundaryAppliesToKeys = new Set(["include", "exclude"]);
const restrictedImportKeys = new Set([
  "kind",
  "name",
  "names",
  "path",
  "package_roots",
  "include_subpaths",
]);
const schedulerResourceKeys = new Set([
  "name",
  "display_name",
  "schedulers",
  "display_order",
  "capacity",
]);
const schedulerTemplateKeys = new Set([
  "name",
  "prefix",
  "display_name",
  "schedulers",
  "display_order",
]);
const schedulerCapacityKeys = new Set([
  "default_limit",
  "auto_policy",
  "override_env",
]);
const schedulerCapacityProfileKeys = new Set([
  "name",
  "scheduler",
  "resources",
]);
const schedulerForwardingProfileKeys = new Set(["name", "mappings"]);
const schedulerForwardingMappingKeys = new Set([
  "source_resource",
  "target_resource",
  "env_variable",
]);

function usage() {
  throw new Error(
    "usage: check-json-shapes.mjs [--root <path>] [--kind <kind> --file <path>]",
  );
}

function parseArgs(argv) {
  const options = {
    root: repoRoot,
    kind: "",
    file: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--root") {
      options.root = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--kind") {
      options.kind = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--file") {
      options.file = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    usage();
  }
  if (!options.root) {
    usage();
  }
  if ((options.kind === "") !== (options.file === "")) {
    usage();
  }
  options.root = path.resolve(options.root);
  if (options.file) {
    options.file = path.isAbsolute(options.file)
      ? options.file
      : path.join(options.root, options.file);
  }
  return options;
}

function repoFile(root, relativePath) {
  return path.join(root, relativePath);
}

function readShapeFile(file, label = file) {
  return readJsonObject(file, label);
}

function serializeJSON(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function assertGeneratedJSONFresh(root, relativePath, rendered) {
  const file = repoFile(root, relativePath);
  const existing = readFileSync(file, "utf8");
  const expected = serializeJSON(rendered);
  if (existing !== expected) {
    throw new Error(`${relativePath} is stale; run make phase-schedules`);
  }
}

function validatePhaseRegistryShape(file) {
  const registry = readShapeFile(file, file);
  assertObjectKeys(registry, new Set(["schema_id", "phases"]), file);
  requireSchemaID(registry, phaseRegistrySchemaID, file);
  const entries = requireObjectArray(registry.phases, `${file}.phases`, {
    nonEmpty: true,
  });
  const phases = [];
  const orders = [];
  for (const [index, entry] of entries.entries()) {
    const label = `${file}.phases[${index + 1}]`;
    assertObjectKeys(
      entry,
      entry.status === "retired"
        ? new Set([
            "phase",
            "order",
            "status",
            "label",
            "manifest_path",
            "ledger_path",
            "scope",
            "normative_owners",
            "retired_reason",
            "retained_artifacts",
          ])
        : new Set([
            "phase",
            "order",
            "status",
            "label",
            "manifest_path",
            "ledger_path",
            "scope",
            "normative_owners",
          ]),
      label,
    );
    phases.push(
      requireString(entry.phase, `${label}.phase`, {
        pattern: phaseNamePattern,
      }),
    );
    orders.push(requireInteger(entry.order, `${label}.order`, { min: 0 }));
    requireEnum(entry.status, `${label}.status`, phaseStatusValues);
    requireString(entry.label, `${label}.label`);
    requireRepoRelativePath(entry.manifest_path, `${label}.manifest_path`, {
      extension: ".json",
    });
    requireRepoRelativePath(entry.ledger_path, `${label}.ledger_path`, {
      extension: ".md",
    });
    requireString(entry.scope, `${label}.scope`);
    requireString(entry.normative_owners, `${label}.normative_owners`);
  }
  assertUnique(phases, `${file}.phases.phase`);
  assertUnique(orders, `${file}.phases.order`);
}

function validatePhaseMapShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, phaseManifestTopLevelKeys, file);
  assertRequiredKeys(manifest, phaseManifestRequiredKeys, file);
  requireSchemaID(manifest, phaseTestMapSchemaID, file);
  requireString(manifest.phase, `${file}.phase`, { pattern: phaseNamePattern });
  requireString(manifest.note, `${file}.note`);
  const ledger = requireObject(manifest.ledger, `${file}.ledger`);
  assertObjectKeys(ledger, phaseLedgerKeys, `${file}.ledger`);
  requireStringArray(manifest.expected_ids, `${file}.expected_ids`, {
    nonEmpty: true,
  });
  if (manifest.forbidden_id_files !== undefined) {
    requireStringArray(
      manifest.forbidden_id_files,
      `${file}.forbidden_id_files`,
    );
  }
  for (const section of ["unit", "integration", "e2e"]) {
    const entries = requireObjectArray(
      manifest[section] ?? [],
      `${file}.${section}`,
    );
    const ids = [];
    for (const [index, entry] of entries.entries()) {
      const label = `${file}.${section}[${index + 1}]`;
      assertObjectKeys(entry, phaseMapEntryKeys, label);
      ids.push(requireString(entry.id, `${label}.id`));
      requireEnum(
        entry.coverage,
        `${label}.coverage`,
        new Set(["authoritative", "supplemental"]),
      );
      requireEnum(
        entry.runner,
        `${label}.runner`,
        new Set(["go_test", "playwright", "vitest"]),
      );
      requireString(entry.file, `${label}.file`);
      requireString(entry.evidence_layer, `${label}.evidence_layer`);
      if (entry.symbol !== undefined && entry.symbols !== undefined) {
        throw new Error(`${label} must declare symbol or symbols[], not both`);
      }
    }
    assertUnique(ids, `${file}.${section}.id`);
  }
  const supportEntries = requireObjectArray(
    manifest.support_go_targets ?? [],
    `${file}.support_go_targets`,
  );
  for (const [index, entry] of supportEntries.entries()) {
    const label = `${file}.support_go_targets[${index + 1}]`;
    assertObjectKeys(entry, phaseSupportEntryKeys, label);
    requireString(entry.target, `${label}.target`);
    requireString(entry.section, `${label}.section`);
    requireString(entry.package, `${label}.package`);
    requireString(entry.file, `${label}.file`);
    requireString(entry.selection_pattern, `${label}.selection_pattern`);
  }
}

function validatePhasePolicyExceptionsShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, new Set(["schema_id", "exceptions"]), file);
  requireSchemaID(manifest, phasePolicyExceptionsSchemaID, file);
  const entries = requireObjectArray(manifest.exceptions, `${file}.exceptions`);
  const ids = [];
  for (const [index, entry] of entries.entries()) {
    const label = `${file}.exceptions[${index + 1}]`;
    assertObjectKeys(
      entry,
      new Set([
        "id",
        "type",
        "owner",
        "reason",
        "expires_before_phase",
        "expires_on",
        "selection",
      ]),
      label,
    );
    ids.push(requireString(entry.id, `${label}.id`));
    requireString(entry.type, `${label}.type`);
    requireString(entry.owner, `${label}.owner`);
    requireString(entry.reason, `${label}.reason`);
  }
  assertUnique(ids, `${file}.exceptions.id`);
}

function validateExecutionTopologyShape(file) {
  const topology = readShapeFile(file, file);
  assertObjectKeys(topology, topologyTopLevelKeys, file);
  requireSchemaID(topology, executionTopologySchemaID, file);
  const dependencies = requireObjectArray(
    topology.execution_dependencies,
    `${file}.execution_dependencies`,
    { nonEmpty: true },
  );
  assertUnique(
    dependencies.map((entry, index) =>
      requireString(
        entry.id,
        `${file}.execution_dependencies[${index + 1}].id`,
        {
          pattern: snakeIDPattern,
        },
      ),
    ),
    `${file}.execution_dependencies.id`,
  );
  const taskSurface = requireObject(
    topology.task_surface,
    `${file}.task_surface`,
  );
  const targets = requireObjectArray(
    taskSurface.targets,
    `${file}.task_surface.targets`,
    {
      nonEmpty: true,
    },
  );
  assertUnique(
    targets.map((entry, index) =>
      requireString(
        entry.name,
        `${file}.task_surface.targets[${index + 1}].name`,
        {
          pattern: makeTargetPattern,
        },
      ),
    ),
    `${file}.task_surface.targets.name`,
  );
}

function validateTaskSurfaceShape(file) {
  const manifest = readShapeFile(file, file);
  requireSchemaID(manifest, taskSurfaceSchemaID, file);
  const errors = collectTaskSurfaceManifestErrors(manifest);
  if (errors.length > 0) {
    throw new Error(
      `${file} is invalid:\n${errors.map((error) => `  - ${error}`).join("\n")}`,
    );
  }
}

function validateCheckScheduleShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, checkScheduleKeys, file);
  requireSchemaID(manifest, checkScheduleSchemaID, file);
  for (const [scheduleIndex, schedule] of requireObjectArray(
    manifest.schedules,
    `${file}.schedules`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.schedules[${scheduleIndex + 1}]`;
    assertObjectKeys(schedule, checkScheduleEntryKeys, label);
    requireString(schedule.target, `${label}.target`, {
      pattern: makeTargetPattern,
    });
    requireString(schedule.capacity_profile, `${label}.capacity_profile`);
    for (const [unitIndex, unit] of requireObjectArray(
      schedule.work_units,
      `${label}.work_units`,
      { nonEmpty: true },
    ).entries()) {
      const unitLabel = `${label}.work_units[${unitIndex + 1}]`;
      assertObjectKeys(unit, checkWorkUnitKeys, unitLabel);
      requireString(unit.target, `${unitLabel}.target`, {
        pattern: makeTargetPattern,
      });
      requirePositiveInteger(unit.weight, `${unitLabel}.weight`);
      if (unit.env !== undefined) {
        for (const name of Object.keys(
          requireObject(unit.env, `${unitLabel}.env`),
        )) {
          requireString(name, `${unitLabel}.env key`, {
            pattern: envNamePattern,
          });
        }
      }
    }
  }
}

function validateServiceBackedScheduleShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, serviceScheduleKeys, file);
  requireSchemaID(manifest, serviceBackedScheduleSchemaID, file);
  const generated = requireObject(manifest.generated, `${file}.generated`);
  assertObjectKeys(generated, serviceGeneratedKeys, `${file}.generated`);
  for (const key of serviceGeneratedKeys) {
    requireString(generated[key], `${file}.generated.${key}`);
  }
  for (const [scheduleIndex, schedule] of requireObjectArray(
    manifest.schedules,
    `${file}.schedules`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.schedules[${scheduleIndex + 1}]`;
    assertObjectKeys(schedule, serviceScheduleEntryKeys, label);
    requireString(schedule.target, `${label}.target`, {
      pattern: makeTargetPattern,
    });
    requireString(schedule.capacity_profile, `${label}.capacity_profile`);
    for (const [sourceIndex, source] of requireObjectArray(
      schedule.work_unit_sources,
      `${label}.work_unit_sources`,
      { nonEmpty: true },
    ).entries()) {
      const sourceLabel = `${label}.work_unit_sources[${sourceIndex + 1}]`;
      assertObjectKeys(source, serviceSourceKeys, sourceLabel);
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
    }
  }
}

function validateBrowserBatchShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, browserBatchKeys, file);
  requireSchemaID(manifest, browserBatchManifestSchemaID, file);
  for (const [stageIndex, stage] of requireObjectArray(
    manifest.stages,
    `${file}.stages`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.stages[${stageIndex + 1}]`;
    assertObjectKeys(stage, browserStageKeys, label);
    requireString(stage.name, `${label}.name`);
    requireString(stage.target, `${label}.target`, {
      pattern: makeTargetPattern,
    });
    for (const [groupIndex, group] of requireObjectArray(
      stage.groups,
      `${label}.groups`,
      {
        nonEmpty: true,
      },
    ).entries()) {
      const groupLabel = `${label}.groups[${groupIndex + 1}]`;
      assertObjectKeys(group, browserGroupKeys, groupLabel);
      requireString(group.name, `${groupLabel}.name`);
      requireString(group.target, `${groupLabel}.target`, {
        pattern: makeTargetPattern,
      });
    }
  }
}

function validateGeneratedArtifactPolicyShape(file) {
  const policy = readShapeFile(file, file);
  assertObjectKeys(policy, generatedArtifactPolicyKeys, file);
  requireSchemaID(policy, generatedArtifactPolicySchemaID, file);
  requireStringArray(
    policy.ignored_sentinel_filenames ?? [],
    `${file}.ignored_sentinel_filenames`,
  );
  for (const [index, root] of requireObjectArray(
    policy.generated_roots ?? [],
    `${file}.generated_roots`,
  ).entries()) {
    const label = `${file}.generated_roots[${index + 1}]`;
    assertObjectKeys(root, generatedArtifactEntryKeys, label);
    requireRepoRelativePath(root.path, `${label}.path`);
    requireStringArray(root.allowed_extensions, `${label}.allowed_extensions`, {
      nonEmpty: true,
    });
    requireString(root.required_marker, `${label}.required_marker`);
  }
  for (const [index, generatedFile] of requireObjectArray(
    policy.generated_files ?? [],
    `${file}.generated_files`,
  ).entries()) {
    const label = `${file}.generated_files[${index + 1}]`;
    assertObjectKeys(generatedFile, generatedArtifactEntryKeys, label);
    requireRepoRelativePath(generatedFile.path, `${label}.path`);
    requireStringArray(
      generatedFile.allowed_extensions,
      `${label}.allowed_extensions`,
      {
        nonEmpty: true,
      },
    );
    requireString(generatedFile.required_marker, `${label}.required_marker`);
  }
  const lintScope = requireObject(
    policy.lint_scope_checks,
    `${file}.lint_scope_checks`,
  );
  assertObjectKeys(lintScope, lintScopeCheckKeys, `${file}.lint_scope_checks`);
  for (const [index, source] of requireObjectArray(
    lintScope.shell_sources,
    `${file}.lint_scope_checks.shell_sources`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.lint_scope_checks.shell_sources[${index + 1}]`;
    assertObjectKeys(source, lintShellSourceKeys, label);
    requireRepoRelativePath(source.path, `${label}.path`);
    requireStringArray(source.must_contain, `${label}.must_contain`, {
      nonEmpty: true,
    });
    requireStringArray(
      source.must_not_contain ?? [],
      `${label}.must_not_contain`,
    );
  }
  const biome = requireObject(
    lintScope.biome,
    `${file}.lint_scope_checks.biome`,
  );
  assertObjectKeys(biome, lintBiomeKeys, `${file}.lint_scope_checks.biome`);
  requireRepoRelativePath(biome.path, `${file}.lint_scope_checks.biome.path`);
  requireStringArray(
    biome.required_files_includes,
    `${file}.lint_scope_checks.biome.required_files_includes`,
  );
  requireStringArray(
    biome.forbidden_files_includes,
    `${file}.lint_scope_checks.biome.forbidden_files_includes`,
  );
  requireStringArray(
    biome.required_override_includes,
    `${file}.lint_scope_checks.biome.required_override_includes`,
  );
  const frontendBoundaries = requireObject(
    lintScope.frontend_import_boundaries,
    `${file}.lint_scope_checks.frontend_import_boundaries`,
  );
  assertObjectKeys(
    frontendBoundaries,
    lintFrontendBoundaryKeys,
    `${file}.lint_scope_checks.frontend_import_boundaries`,
  );
  requireRepoRelativePath(
    frontendBoundaries.path,
    `${file}.lint_scope_checks.frontend_import_boundaries.path`,
  );
  requireStringArray(
    frontendBoundaries.required_scan_excludes,
    `${file}.lint_scope_checks.frontend_import_boundaries.required_scan_excludes`,
  );
  requireStringArray(
    frontendBoundaries.required_restricted_paths,
    `${file}.lint_scope_checks.frontend_import_boundaries.required_restricted_paths`,
  );
}

function validateFrontendImportBoundariesShape(file) {
  const config = readShapeFile(file, file);
  assertObjectKeys(config, frontendBoundaryKeys, file);
  requireSchemaID(config, frontendImportBoundariesSchemaID, file);
  requireStringArray(config.scan_roots, `${file}.scan_roots`, {
    nonEmpty: true,
  });
  requireStringArray(config.scan_excludes ?? [], `${file}.scan_excludes`);
  for (const [index, rule] of requireObjectArray(
    config.rules,
    `${file}.rules`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.rules[${index + 1}]`;
    assertObjectKeys(rule, frontendBoundaryRuleKeys, label);
    requireString(rule.id, `${label}.id`);
    requireEnum(rule.level, `${label}.level`, new Set(["error", "warning"]));
    requireString(rule.message, `${label}.message`);
    const appliesTo = requireObject(rule.applies_to, `${label}.applies_to`);
    assertObjectKeys(
      appliesTo,
      frontendBoundaryAppliesToKeys,
      `${label}.applies_to`,
    );
    requireStringArray(appliesTo.include, `${label}.applies_to.include`, {
      nonEmpty: true,
    });
    requireStringArray(appliesTo.exclude ?? [], `${label}.applies_to.exclude`);
    requireStringArray(rule.allowed_importers, `${label}.allowed_importers`);
    for (const [importIndex, restrictedImport] of requireObjectArray(
      rule.restricted_imports,
      `${label}.restricted_imports`,
      { nonEmpty: true },
    ).entries()) {
      const importLabel = `${label}.restricted_imports[${importIndex + 1}]`;
      assertObjectKeys(restrictedImport, restrictedImportKeys, importLabel);
      const kind = requireEnum(
        restrictedImport.kind,
        `${importLabel}.kind`,
        new Set([
          "package",
          "path_prefix",
          "node_builtin",
          "workspace_package_facade",
        ]),
      );
      if (kind === "package") {
        requireString(restrictedImport.name, `${importLabel}.name`);
      }
      if (kind === "path_prefix") {
        requireRepoRelativePath(restrictedImport.path, `${importLabel}.path`);
      }
      if (kind === "node_builtin") {
        requireStringArray(
          restrictedImport.names ?? [],
          `${importLabel}.names`,
        );
      }
      if (kind === "workspace_package_facade") {
        for (const [rootIndex, packageRoot] of requireStringArray(
          restrictedImport.package_roots,
          `${importLabel}.package_roots`,
          { nonEmpty: true },
        ).entries()) {
          requireRepoRelativePath(
            packageRoot,
            `${importLabel}.package_roots[${rootIndex + 1}]`,
          );
        }
      }
      if (
        restrictedImport.include_subpaths !== undefined &&
        typeof restrictedImport.include_subpaths !== "boolean"
      ) {
        throw new Error(`${importLabel}.include_subpaths must be a boolean`);
      }
    }
  }
}

function validateSchedulerResourceRegistryShape(file) {
  const registry = readShapeFile(file, file);
  assertObjectKeys(registry, schedulerResourceRegistryKeys, file);
  requireSchemaID(registry, schedulerResourceRegistrySchemaID, file);
  for (const [index, resource] of requireObjectArray(
    registry.resources,
    `${file}.resources`,
    {
      nonEmpty: true,
    },
  ).entries()) {
    const label = `${file}.resources[${index + 1}]`;
    assertObjectKeys(resource, schedulerResourceKeys, label);
    requireString(resource.name, `${label}.name`);
    requireString(resource.display_name, `${label}.display_name`);
    requireStringArray(resource.schedulers, `${label}.schedulers`, {
      nonEmpty: true,
    });
    requireInteger(resource.display_order, `${label}.display_order`, {
      min: 0,
    });
    if (resource.capacity !== undefined) {
      const capacity = requireObject(resource.capacity, `${label}.capacity`);
      assertObjectKeys(capacity, schedulerCapacityKeys, `${label}.capacity`);
      if (capacity.default_limit !== undefined) {
        requirePositiveInteger(
          capacity.default_limit,
          `${label}.capacity.default_limit`,
        );
      }
      if (capacity.auto_policy !== undefined) {
        requireString(capacity.auto_policy, `${label}.capacity.auto_policy`);
      }
      if (
        (capacity.default_limit === undefined) ===
        (capacity.auto_policy === undefined)
      ) {
        throw new Error(
          `${label}.capacity must declare exactly one of default_limit or auto_policy`,
        );
      }
      if (capacity.override_env !== undefined) {
        requireString(capacity.override_env, `${label}.capacity.override_env`, {
          pattern: envNamePattern,
        });
      }
    }
  }
  for (const [index, template] of requireObjectArray(
    registry.templates,
    `${file}.templates`,
    {
      nonEmpty: true,
    },
  ).entries()) {
    const label = `${file}.templates[${index + 1}]`;
    assertObjectKeys(template, schedulerTemplateKeys, label);
    requireString(template.name, `${label}.name`);
    requireString(template.prefix, `${label}.prefix`);
    requireString(template.display_name, `${label}.display_name`);
    requireStringArray(template.schedulers, `${label}.schedulers`, {
      nonEmpty: true,
    });
    requireInteger(template.display_order, `${label}.display_order`, {
      min: 0,
    });
  }
  for (const [index, profile] of requireObjectArray(
    registry.capacity_profiles,
    `${file}.capacity_profiles`,
    { nonEmpty: true },
  ).entries()) {
    const label = `${file}.capacity_profiles[${index + 1}]`;
    assertObjectKeys(profile, schedulerCapacityProfileKeys, label);
    requireString(profile.name, `${label}.name`);
    requireString(profile.scheduler, `${label}.scheduler`);
    requireStringArray(profile.resources, `${label}.resources`, {
      nonEmpty: true,
    });
  }
  for (const [index, profile] of requireObjectArray(
    registry.forwarding_profiles,
    `${file}.forwarding_profiles`,
  ).entries()) {
    const label = `${file}.forwarding_profiles[${index + 1}]`;
    assertObjectKeys(profile, schedulerForwardingProfileKeys, label);
    requireString(profile.name, `${label}.name`);
    for (const [mappingIndex, mapping] of requireObjectArray(
      profile.mappings,
      `${label}.mappings`,
      { nonEmpty: true },
    ).entries()) {
      const mappingLabel = `${label}.mappings[${mappingIndex + 1}]`;
      assertObjectKeys(mapping, schedulerForwardingMappingKeys, mappingLabel);
      requireString(mapping.source_resource, `${mappingLabel}.source_resource`);
      requireString(mapping.target_resource, `${mappingLabel}.target_resource`);
      requireString(mapping.env_variable, `${mappingLabel}.env_variable`, {
        pattern: envNamePattern,
      });
    }
  }
}

function validateBootstrapAdminShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, bootstrapAdminKeys, file);
  if (manifest.bootstrap_schema_id !== bootstrapAdminSchemaID) {
    throw new Error(
      `${file}.bootstrap_schema_id must be ${bootstrapAdminSchemaID}`,
    );
  }
  requireString(
    manifest.bootstrap_artifact_id,
    `${file}.bootstrap_artifact_id`,
  );
  requireString(manifest.email, `${file}.email`, {
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  });
  requireString(manifest.display_name, `${file}.display_name`);
  requireString(manifest.initial_password, `${file}.initial_password`);
}

function validateDurationBaselineShape(file) {
  const baseline = readShapeFile(file, file);
  requireSchemaID(baseline, serviceBackedMakeTargetBaselineSchemaID, file);
  requirePositiveInteger(
    baseline.default_make_target_weight_ms,
    `${file}.default_make_target_weight_ms`,
  );
  requireObject(baseline.targets, `${file}.targets`);
}

function requireSorted(values, label, keyFn, orderLabel = "stable key") {
  let previous = null;
  for (const value of values) {
    const key = keyFn(value);
    if (previous !== null && key < previous) {
      throw new Error(`${label} must be sorted by ${orderLabel}`);
    }
    previous = key;
  }
}

function artifactStableKey(artifact) {
  return `${artifact.role}\u0000${artifact.kind}\u0000${artifact.path}`;
}

function failureStableKey(failure) {
  return [
    failure.failure_class ?? "",
    failure.target ?? "",
    failure.work_unit ?? "",
    failure.child_target ?? "",
    failure.label ?? "",
    failure.kind ?? "",
    failure.message ?? failure.headline ?? "",
  ].join("\u0000");
}

function requireNullableEnum(value, label, allowed) {
  if (value === null) {
    return null;
  }
  return requireEnum(value, label, allowed);
}

function requireRFC3339Timestamp(value, label) {
  const timestamp = requireString(value, label);
  if (
    !rfc3339TimestampPattern.test(timestamp) ||
    Number.isNaN(Date.parse(timestamp))
  ) {
    throw new Error(`${label} must be an RFC3339 timestamp`);
  }
  return timestamp;
}

function validateToolRunArtifacts(value, label) {
  const artifacts = requireObjectArray(value, label);
  for (const [index, artifact] of artifacts.entries()) {
    const artifactLabel = `${label}[${index + 1}]`;
    assertObjectKeys(artifact, toolRunArtifactKeys, artifactLabel);
    requireString(artifact.role, `${artifactLabel}.role`);
    requireString(artifact.kind, `${artifactLabel}.kind`);
    requireString(artifact.path, `${artifactLabel}.path`);
  }
  requireSorted(artifacts, label, artifactStableKey, "role, kind, and path");
}

function validateNonNegativeIntegerObject(value, label, keys) {
  const object = requireObject(value, label);
  assertObjectKeys(object, keys, label);
  assertRequiredKeys(object, keys, label);
  for (const key of keys) {
    requireInteger(object[key], `${label}.${key}`, { min: 0 });
  }
}

function validateToolRunWorkUnits(value, label) {
  const units = requireObjectArray(value, label);
  for (const [index, unit] of units.entries()) {
    const unitLabel = `${label}[${index + 1}]`;
    assertObjectKeys(unit, toolRunWorkUnitKeys, unitLabel);
    requireString(unit.id, `${unitLabel}.id`);
    requireInteger(unit.completed, `${unitLabel}.completed`, { min: 0 });
    requireInteger(unit.total, `${unitLabel}.total`, { min: 0 });
    requireString(unit.status, `${unitLabel}.status`);
    if (unit.aborted_after !== undefined) {
      requireString(unit.aborted_after, `${unitLabel}.aborted_after`);
    }
  }
  requireSorted(units, label, (unit) => unit.id, "work unit id");
}

function validateToolRunTargetRefs(value, label, keys) {
  const targets = requireObjectArray(value, label);
  for (const [index, target] of targets.entries()) {
    const targetLabel = `${label}[${index + 1}]`;
    assertObjectKeys(target, keys, targetLabel);
    requireString(target.target, `${targetLabel}.target`);
    if (target.status !== undefined) {
      requireString(target.status, `${targetLabel}.status`);
    }
    if (target.artifact_root !== undefined) {
      requireString(target.artifact_root, `${targetLabel}.artifact_root`);
    }
  }
  requireSorted(targets, label, (target) => target.target, "target id");
}

function validateToolRunSlowest(value, label) {
  const entries = requireObjectArray(value, label);
  for (const [index, entry] of entries.entries()) {
    const entryLabel = `${label}[${index + 1}]`;
    assertObjectKeys(entry, toolRunSlowestKeys, entryLabel);
    requireString(entry.id, `${entryLabel}.id`);
    requireInteger(entry.duration_ms, `${entryLabel}.duration_ms`, { min: 0 });
    requireString(entry.kind, `${entryLabel}.kind`);
  }
  requireSorted(
    entries,
    label,
    (entry) =>
      `${String(Number.MAX_SAFE_INTEGER - entry.duration_ms).padStart(16, "0")}\u0000${entry.id}`,
    "descending duration and id",
  );
}

function validateToolRunSummaryShape(file) {
  const summary = readShapeFile(file, file);
  assertObjectKeys(summary, toolRunSummaryKeys, file);
  assertRequiredKeys(summary, toolRunSummaryKeys, file);
  requireSchemaID(summary, toolRunSummarySchemaID, file);
  requireString(summary.target, `${file}.target`, {
    pattern: makeTargetPattern,
  });
  const command = requireObject(summary.command, `${file}.command`);
  assertObjectKeys(command, toolRunCommandKeys, `${file}.command`);
  assertRequiredKeys(command, toolRunCommandKeys, `${file}.command`);
  requireString(command.cwd, `${file}.command.cwd`);
  for (const [index, arg] of requireArray(
    command.argv,
    `${file}.command.argv`,
  ).entries()) {
    requireString(arg, `${file}.command.argv[${index + 1}]`);
  }
  if (command.make_target !== null) {
    requireString(command.make_target, `${file}.command.make_target`, {
      pattern: makeTargetPattern,
    });
  }
  requireObject(command.env, `${file}.command.env`);
  requireEnum(summary.status, `${file}.status`, toolRunStatusValues);
  requireInteger(summary.exit_code, `${file}.exit_code`);
  requireRFC3339Timestamp(summary.started_at, `${file}.started_at`);
  requireRFC3339Timestamp(summary.completed_at, `${file}.completed_at`);
  requireInteger(summary.duration_ms, `${file}.duration_ms`, { min: 0 });
  requireEnum(summary.output_mode, `${file}.output_mode`, toolRunOutputModes);
  requireString(summary.artifact_root, `${file}.artifact_root`);
  validateToolRunArtifacts(
    summary.summary_artifacts,
    `${file}.summary_artifacts`,
  );
  validateToolRunArtifacts(summary.log_artifacts, `${file}.log_artifacts`);
  validateToolRunWorkUnits(summary.work_units, `${file}.work_units`);
  validateToolRunTargetRefs(
    summary.evidence_targets,
    `${file}.evidence_targets`,
    toolRunEvidenceTargetKeys,
  );
  validateToolRunTargetRefs(
    summary.helper_units,
    `${file}.helper_units`,
    toolRunHelperUnitKeys,
  );
  validateNonNegativeIntegerObject(
    summary.counts,
    `${file}.counts`,
    toolRunCountKeys,
  );
  validateNonNegativeIntegerObject(
    summary.phase_accounting,
    `${file}.phase_accounting`,
    toolRunPhaseAccountingKeys,
  );
  requireNullableEnum(
    summary.failure_class,
    `${file}.failure_class`,
    toolRunFailureClasses,
  );
  requireNullableEnum(
    summary.failure_origin,
    `${file}.failure_origin`,
    toolRunFailureOrigins,
  );
  requireSorted(
    requireObjectArray(summary.failures, `${file}.failures`),
    `${file}.failures`,
    failureStableKey,
    "failure class and target",
  );
  validateToolRunSlowest(summary.slowest, `${file}.slowest`);
  requireObjectArray(summary.warnings, `${file}.warnings`);
  requireStringArray(summary.rerun_commands, `${file}.rerun_commands`);
  requireObject(summary.extensions, `${file}.extensions`);
}

function validateKind(kind, file) {
  switch (kind) {
    case "phase-registry":
      validatePhaseRegistryShape(file);
      return;
    case "phase-map":
      validatePhaseMapShape(file);
      return;
    case "phase-policy-exceptions":
      validatePhasePolicyExceptionsShape(file);
      return;
    case "execution-topology":
      validateExecutionTopologyShape(file);
      return;
    case "task-surface":
      validateTaskSurfaceShape(file);
      return;
    case "check-schedule":
      validateCheckScheduleShape(file);
      return;
    case "service-backed-schedule":
      validateServiceBackedScheduleShape(file);
      return;
    case "browser-batch":
      validateBrowserBatchShape(file);
      return;
    case "generated-artifact-policy":
      validateGeneratedArtifactPolicyShape(file);
      return;
    case "frontend-import-boundaries":
      validateFrontendImportBoundariesShape(file);
      return;
    case "scheduler-resource-registry":
      validateSchedulerResourceRegistryShape(file);
      return;
    case "service-backed-make-target-baseline":
      validateDurationBaselineShape(file);
      return;
    case "bootstrap-admin":
      validateBootstrapAdminShape(file);
      return;
    case "tool-run-summary":
      validateToolRunSummaryShape(file);
      return;
    default:
      throw new Error(`unknown json shape kind ${kind}`);
  }
}

function validateAll(root) {
  validatePhaseRegistryShape(repoFile(root, "tools/phase_registry.json"));
  validatePhaseRegistry(root);
  for (const entry of activePhaseRegistryEntries(root)) {
    validatePhaseMapShape(repoFile(root, entry.manifest_path));
    validatePhaseManifestSemantics(root, entry.phase);
  }
  validatePhasePolicyExceptionsShape(
    repoFile(root, "tools/phase_policy_exceptions.json"),
  );
  loadPhasePolicyExceptions(root);

  validateExecutionTopologyShape(
    repoFile(root, "tools/execution_topology_manifest.json"),
  );
  const topology = loadExecutionTopology({
    root,
    manifestPath: repoFile(root, "tools/execution_topology_manifest.json"),
  });
  const taskSurface = renderTaskSurfaceManifest(topology);
  const checkSchedule = renderCheckScheduleManifest(topology);
  const browserBatch = renderBrowserBatchManifest(topology);
  const serviceBackedSchedule = renderServiceBackedScheduleManifest({
    topology: defaultExecutionTopologyManifestPath,
    topologyObject: topology,
  });
  assertGeneratedJSONFresh(
    root,
    topology.generatedOutputs.task_surface_manifest,
    taskSurface,
  );
  assertGeneratedJSONFresh(
    root,
    topology.generatedOutputs.check_schedule_manifest,
    checkSchedule,
  );
  assertGeneratedJSONFresh(
    root,
    topology.generatedOutputs.browser_e2e_batch_manifest,
    browserBatch,
  );
  assertGeneratedJSONFresh(
    root,
    topology.generatedOutputs.service_backed_schedule_manifest,
    serviceBackedSchedule,
  );

  validateTaskSurfaceShape(repoFile(root, "tools/task_surface_manifest.json"));
  loadTaskSurfaceManifest(repoFile(root, "tools/task_surface_manifest.json"));
  validateCheckScheduleShape(
    repoFile(root, "tools/check_schedule_manifest.json"),
  );
  validateServiceBackedScheduleShape(
    repoFile(root, "tools/service_backed_schedule_manifest.json"),
  );
  validateServiceBackedScheduleTopology({
    scheduleManifestPath: repoFile(
      root,
      "tools/service_backed_schedule_manifest.json",
    ),
    topologyPath: repoFile(root, "tools/execution_topology_manifest.json"),
  });
  validateBrowserBatchShape(
    repoFile(root, "tools/browser_e2e_batch_manifest.json"),
  );
  loadBrowserBatchManifest(
    repoFile(root, "tools/browser_e2e_batch_manifest.json"),
  );

  validateGeneratedArtifactPolicyShape(
    repoFile(root, "tools/generated_artifact_policy.json"),
  );
  validateFrontendImportBoundariesShape(
    repoFile(root, "tools/frontend_import_boundaries.json"),
  );
  validateSchedulerResourceRegistryShape(
    repoFile(root, "tools/scheduler_resource_registry.json"),
  );
  loadSchedulerResourceRegistry(
    repoFile(root, "tools/scheduler_resource_registry.json"),
  );
  validateDurationBaselineShape(
    repoFile(root, "tools/service_backed_make_target_duration_baselines.json"),
  );
  validateBootstrapAdminShape(
    repoFile(root, "configs/dev/bootstrap-admin.json"),
  );

  if (browserBatchManifestSchemaID !== renderedBrowserBatchSchemaID) {
    throw new Error("browser batch schema constants diverged");
  }
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.kind) {
    validateKind(options.kind, options.file);
    console.log(`json shape check passed: ${options.kind}`);
    return;
  }
  validateAll(options.root);
  console.log("json shape check passed");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`json shape check failed: ${message}`);
  process.exit(1);
}
