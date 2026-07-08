#!/usr/bin/env node
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadBrowserBatchManifest,
  validateBrowserBatchManifestShape,
} from "../browser/browser-batch-manifest.mjs";
import {
  activePhaseRegistryEntries,
  loadPhasePolicyExceptions,
  phaseRegistrySchemaID,
  validateFrontendPhaseArtifacts,
  validateManifest as validatePhaseManifestSemantics,
  validatePhaseManifestShapeFile,
  validatePhaseRegistry,
} from "../phase-accounting/index.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import {
  executionTopologySchemaID,
  loadExecutionTopology,
  taskSurfaceSchemaID,
} from "./execution-topology.mjs";
import {
  assertObjectKeys,
  assertRequiredKeys,
  assertUnique,
  readJsonObject,
  requireArray,
  requireBoolean,
  requireEnum,
  requireInteger,
  requireNullableEnum,
  requireObject,
  requireObjectArray,
  requirePositiveInteger,
  requireRepoRelativePath,
  requireRFC3339Timestamp,
  requireSchemaID,
  requireSorted,
  requireString,
  requireStringArray,
  validateObjectArray,
} from "../contract/json-shape.mjs";
import {
  validateMigrationHistory,
  validateMigrationHistoryManifestShape,
  validateSchemaObjectOwnership,
  validateSchemaObjectOwnershipManifestShape,
} from "./database-contract-drift/index.mjs";
import { validateSchedulerManifestShape } from "../scheduler/scheduler-manifest.mjs";
import {
  loadSchedulerResourceRegistry,
  validateSchedulerResourceRegistrySemantics,
} from "../scheduler/scheduler-resources.mjs";
import {
  collectTaskSurfaceManifestErrors,
  loadTaskSurfaceManifest,
} from "./task-surface.mjs";
import { quickCheckRenderIndex } from "./render-execution-topology-artifacts.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const phasePolicyExceptionsSchemaID = "cartulary.phase_policy_exceptions.v1";
const generatedArtifactPolicySchemaID =
  "cartulary.generated_artifact_policy.v1";
const frontendImportBoundariesSchemaID =
  "cartulary.frontend_import_boundaries.v2";
const bootstrapAdminSchemaID = "cartulary.bootstrap_admin.v1";
const serviceBackedMakeTargetBaselineSchemaID =
  "cartulary.scheduler_work_unit_duration_baselines.v2";
const toolRunSummarySchemaID = "cartulary.tool_run_summary.v3";
const fallowStaticSummarySchemaID = "cartulary.fallow_static_summary.v1";
const agentFinalizeSummarySchemaID = "cartulary.agent_finalize_summary.v3";
const frontendPhaseRegistrySchemaID = "cartulary.frontend_phase_registry.v5";
const frontendPhaseTestMapSchemaID = "cartulary.frontend_phase_test_map.v4";
const testAccountingClassificationSchemaID =
  "cartulary.test_accounting_classification.v2";
const projectionProviderManifestSchemaID =
  "cartulary.projection_provider_manifest.v3";
const graphProjectionConformanceMatrixSchemaID =
  "cartulary.graph_projection_conformance_matrix.v1";
const frontendVisualFixtureRegistrySchemaID =
  "cartulary.frontend_visual_fixture_registry.v3";
const sharedExtensionsRef = "cartulary.harness.defs.v1#/$defs/extensions";
const schedulerSummaryCommonSchemaID = "cartulary.scheduler_summary.common.v10";
const schedulerSummaryCommonSchemaIDs = new Set([
  "cartulary.scheduler_summary.common.v9",
  schedulerSummaryCommonSchemaID,
]);

const phaseStatusValues = new Set(["active", "planned", "retired"]);
const phaseNamePattern = /^phase(?:0|[1-9]\d*)$/;
const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const snakeIDPattern = /^[a-z][a-z0-9_]*$/;
const topologyTopLevelKeys = new Set([
  "schema_id",
  "generated_outputs",
  "runtime_binaries",
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
  "singleton_imports",
  "rules",
  "raw_design_token_literal_checks",
]);
const testAccountingClassificationKeys = new Set([
  "schema_id",
  "vitest",
  "go_packages",
  "go_tests",
  "playwright",
]);
const projectionProviderManifestKeys = new Set([
  "schema_id",
  "manifest_version",
  "authority",
  "source_registry",
  "import_policy",
  "providers",
]);
const projectionProviderImportPolicyKeys = new Set([
  "approved_root_importers",
  "approved_adapter_packages",
  "approved_contract_packages",
]);
const projectionProviderEntryKeys = new Set([
  "provider_id",
  "schema_version",
  "source_owner_module",
  "projection_storage_owner_module",
  "view_schema_ids",
  "projection_table_ids",
  "source_authorities",
  "capabilities",
  "restore_rebuild",
  "status",
  "facade_packages",
  "rebuild_after",
  "characterization_refs",
]);
const projectionProviderCapabilityKeys = new Set([
  "query",
  "refresh_row",
  "restore_rebuild",
  "incident_rebuild",
]);
const projectionProviderAuthority =
  "validation_only_code_backed_registry_authoritative";
const projectionProviderDescriptorVersion =
  "projection_provider_descriptor.v2";
const projectionProviderStatusValues = new Set([
  "active",
  "deprecated",
  "experimental",
]);
const projectionProviderRestoreRebuildValues = new Set([
  "required",
  "nonparticipating",
  "unsupported",
]);
const graphProjectionMatrixKeys = new Set([
  "schema_id",
  "spec_path",
  "spec_status",
  "matrix_version",
  "authority",
  "acceptance_criteria",
  "fixture_registry",
]);
const graphProjectionAcceptanceKeys = new Set([
  "id",
  "owner",
  "coverage_status",
  "areas",
  "evidence_selectors",
  "fixture_ids",
]);
const graphProjectionFixtureKeys = new Set([
  "fixture_id",
  "fixture_path",
  "coverage",
]);
const graphProjectionCoverageStatuses = new Set([
  "planned",
  "implemented",
  "deferred",
]);
const graphProjectionAreas = new Set([
  "specification",
  "implementation",
  "tests",
  "documentation",
  "contracts",
  "migration",
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
  "result_root",
  "run_id",
  "run_root",
  "summary_artifacts",
  "log_artifacts",
  "work_units",
  "evidence_targets",
  "helper_units",
  "counts",
  "phase_accounting",
  "failure_class",
  "failure_reason",
  "failures",
  "slowest",
  "warnings",
  "rerun_commands",
  "scheduler_timing",
  "extensions",
]);
const toolRunCommandKeys = new Set(["cwd", "argv", "make_target", "env"]);
const toolRunArtifactKeys = new Set(["role", "kind", "path"]);
const toolRunStatusValues = new Set(["pass", "fail"]);
const toolRunOutputModes = new Set([
  "quiet",
  "summary",
  "ci",
  "verbose",
  "debug",
  "machine",
]);
const toolRunFailureClasses = new Set([
  "product",
  "config",
  "infra",
  "harness",
  "artifact",
  "timing",
  "interrupted",
  "unknown",
]);
const toolRunFailureReasons = new Set([
  "usage_error",
  "configuration_error",
  "preflight_error",
  "service_start_error",
  "service_readiness_timeout",
  "fixture_error",
  "resource_conflict",
  "test_assertion_failure",
  "security_finding",
  "child_target_failure",
  "tool_diagnostic_failure",
  "scheduler_accounting_error",
  "frontend_row_accounting",
  "test_accounting_unmapped",
  "artifact_error",
  "cleanup_error",
  "duration_baseline_drift",
  "timeout_failure",
  "cancelled_or_interrupted",
  "unknown_failure",
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
const toolRunEvidenceTargetKeys = new Set(["target", "status", "run_root"]);
const toolRunHelperUnitKeys = new Set(["target", "status", "run_root"]);
const toolRunSlowestKeys = new Set(["id", "duration_ms", "kind"]);
const extensionKeyPattern =
  /^(?:cartulary\.[A-Za-z0-9_.-]+|[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+){2,})$/u;
const generatedArtifactEntryKeys = new Set([
  "path",
  "allowed_extensions",
  "required_marker",
]);
const lintScopeCheckKeys = new Set([
  "shell_sources",
  "biome",
  "frontend_import_boundaries",
  "markdownlint",
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
const lintMarkdownlintKeys = new Set([
  "path",
  "required_globs",
  "required_ignores",
  "forbidden_globs",
  "required_rules",
  "disabled_rules",
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
const frontendBoundaryRawDesignLiteralCheckKeys = new Set([
  "id",
  "level",
  "message",
  "design_document",
  "token_namespaces",
  "applies_to",
]);
const restrictedImportKeys = new Set([
  "kind",
  "name",
  "names",
  "path",
  "package_roots",
  "include_subpaths",
]);
const singletonImportKeys = new Set([
  "id",
  "level",
  "message",
  "specifier",
  "required_count",
  "allowed_importers",
]);
const frontendBoundaryLevelValues = new Set(["error", "warning"]);
const restrictedImportKindValues = new Set([
  "package",
  "path_prefix",
  "node_builtin",
  "workspace_package_facade",
]);
const supportSchemaIDs = new Set([
  "cartulary.harness.defs.v1",
  ...schedulerSummaryCommonSchemaIDs,
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

function validateFrontendSchemaArtifacts(root) {
  validateSchemaSync(
    frontendPhaseRegistrySchemaID,
    readShapeFile(repoFile(root, "tools/frontend_phase_registry.json")),
  );
  const mapDir = repoFile(root, "tools/frontend_phase_maps");
  for (const filename of readdirSync(mapDir).filter((name) =>
    name.endsWith(".json"),
  )) {
    validateSchemaSync(
      frontendPhaseTestMapSchemaID,
      readShapeFile(path.join(mapDir, filename)),
    );
  }
  validateSchemaSync(
    frontendVisualFixtureRegistrySchemaID,
    readShapeFile(repoFile(root, "tools/frontend_visual_fixture_registry.json")),
  );
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
  validatePhaseManifestShapeFile(file);
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

function validateBrowserBatchShape(file) {
  validateBrowserBatchManifestShape(file, file);
}

function validateGeneratedArtifactEntry(entry, label) {
  requireRepoRelativePath(entry.path, `${label}.path`);
  requireStringArray(entry.allowed_extensions, `${label}.allowed_extensions`, {
    nonEmpty: true,
  });
  requireString(entry.required_marker, `${label}.required_marker`);
}

function validateGeneratedArtifactPolicyShape(file) {
  const policy = readShapeFile(file, file);
  assertObjectKeys(policy, generatedArtifactPolicyKeys, file);
  requireSchemaID(policy, generatedArtifactPolicySchemaID, file);
  requireStringArray(
    policy.ignored_sentinel_filenames ?? [],
    `${file}.ignored_sentinel_filenames`,
  );
  validateObjectArray(
    policy.generated_roots ?? [],
    `${file}.generated_roots`,
    { keys: generatedArtifactEntryKeys },
    validateGeneratedArtifactEntry,
  );
  validateObjectArray(
    policy.generated_files ?? [],
    `${file}.generated_files`,
    { keys: generatedArtifactEntryKeys },
    validateGeneratedArtifactEntry,
  );
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
  const markdownlint = requireObject(
    lintScope.markdownlint,
    `${file}.lint_scope_checks.markdownlint`,
  );
  assertObjectKeys(
    markdownlint,
    lintMarkdownlintKeys,
    `${file}.lint_scope_checks.markdownlint`,
  );
  requireRepoRelativePath(
    markdownlint.path,
    `${file}.lint_scope_checks.markdownlint.path`,
  );
  requireStringArray(
    markdownlint.required_globs,
    `${file}.lint_scope_checks.markdownlint.required_globs`,
    { nonEmpty: true },
  );
  requireStringArray(
    markdownlint.required_ignores,
    `${file}.lint_scope_checks.markdownlint.required_ignores`,
    { nonEmpty: true },
  );
  requireStringArray(
    markdownlint.forbidden_globs ?? [],
    `${file}.lint_scope_checks.markdownlint.forbidden_globs`,
  );
  requireStringArray(
    markdownlint.required_rules,
    `${file}.lint_scope_checks.markdownlint.required_rules`,
    { nonEmpty: true },
  );
  requireStringArray(
    markdownlint.disabled_rules,
    `${file}.lint_scope_checks.markdownlint.disabled_rules`,
    { nonEmpty: true },
  );
}

function requireFrontendBoundaryLevel(value, label) {
  return requireEnum(value, label, frontendBoundaryLevelValues);
}

function validateFrontendBoundaryAppliesTo(value, label) {
  const appliesTo = requireObject(value, label);
  assertObjectKeys(appliesTo, frontendBoundaryAppliesToKeys, label);
  requireStringArray(appliesTo.include, `${label}.include`, {
    nonEmpty: true,
  });
  requireStringArray(appliesTo.exclude ?? [], `${label}.exclude`);
}

function validateRestrictedImport(restrictedImport, importLabel) {
  const kind = requireEnum(
    restrictedImport.kind,
    `${importLabel}.kind`,
    restrictedImportKindValues,
  );
  if (kind === "package") {
    requireString(restrictedImport.name, `${importLabel}.name`);
  }
  if (kind === "path_prefix") {
    requireRepoRelativePath(restrictedImport.path, `${importLabel}.path`);
  }
  if (kind === "node_builtin") {
    requireStringArray(restrictedImport.names ?? [], `${importLabel}.names`);
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

function validateSingletonImport(singletonImport, label) {
  requireString(singletonImport.id, `${label}.id`);
  requireFrontendBoundaryLevel(singletonImport.level, `${label}.level`);
  requireString(singletonImport.message, `${label}.message`);
  requireString(singletonImport.specifier, `${label}.specifier`);
  requirePositiveInteger(
    singletonImport.required_count,
    `${label}.required_count`,
  );
  requireStringArray(
    singletonImport.allowed_importers,
    `${label}.allowed_importers`,
    { nonEmpty: true },
  );
}

function validateFrontendBoundaryRule(rule, label) {
  requireString(rule.id, `${label}.id`);
  requireFrontendBoundaryLevel(rule.level, `${label}.level`);
  requireString(rule.message, `${label}.message`);
  validateFrontendBoundaryAppliesTo(rule.applies_to, `${label}.applies_to`);
  requireStringArray(rule.allowed_importers, `${label}.allowed_importers`);
  validateObjectArray(
    rule.restricted_imports,
    `${label}.restricted_imports`,
    { nonEmpty: true, keys: restrictedImportKeys },
    validateRestrictedImport,
  );
}

function validateRawDesignTokenLiteralCheck(check, label) {
  requireString(check.id, `${label}.id`);
  requireFrontendBoundaryLevel(check.level, `${label}.level`);
  requireString(check.message, `${label}.message`);
  requireRepoRelativePath(check.design_document, `${label}.design_document`);
  requireStringArray(check.token_namespaces, `${label}.token_namespaces`, {
    nonEmpty: true,
  });
  validateFrontendBoundaryAppliesTo(check.applies_to, `${label}.applies_to`);
}

function validateFrontendImportBoundariesShape(file) {
  const config = readShapeFile(file, file);
  assertObjectKeys(config, frontendBoundaryKeys, file);
  requireSchemaID(config, frontendImportBoundariesSchemaID, file);
  requireStringArray(config.scan_roots, `${file}.scan_roots`, {
    nonEmpty: true,
  });
  requireStringArray(config.scan_excludes ?? [], `${file}.scan_excludes`);
  validateObjectArray(
    config.singleton_imports ?? [],
    `${file}.singleton_imports`,
    { keys: singletonImportKeys },
    validateSingletonImport,
  );
  validateObjectArray(
    config.rules,
    `${file}.rules`,
    { nonEmpty: true, keys: frontendBoundaryRuleKeys },
    validateFrontendBoundaryRule,
  );
  validateObjectArray(
    config.raw_design_token_literal_checks ?? [],
    `${file}.raw_design_token_literal_checks`,
    { keys: frontendBoundaryRawDesignLiteralCheckKeys },
    validateRawDesignTokenLiteralCheck,
  );
}

function validateSchedulerResourceRegistryShape(file) {
  validateSchedulerResourceRegistrySemantics(file, file);
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
    baseline.default_work_unit_weight_ms,
    `${file}.default_work_unit_weight_ms`,
  );
  const workUnits = requireObject(baseline.work_units, `${file}.work_units`);
  for (const [key, entry] of Object.entries(workUnits)) {
    requireObject(entry, `${file}.work_units.${key}`);
    const expectedKey = [
      requireString(
        entry.scheduler_kind,
        `${file}.work_units.${key}.scheduler_kind`,
      ),
      requireString(
        entry.schedule_target,
        `${file}.work_units.${key}.schedule_target`,
      ),
      requireString(
        entry.work_unit_id,
        `${file}.work_units.${key}.work_unit_id`,
      ),
      requireString(
        entry.aggregate_target,
        `${file}.work_units.${key}.aggregate_target`,
      ),
    ].join("|");
    if (key !== expectedKey) {
      throw new Error(
        `${file}.work_units.${key} must match scheduler context key ${expectedKey}`,
      );
    }
    requirePositiveInteger(
      entry.weight_ms,
      `${file}.work_units.${key}.weight_ms`,
    );
  }
}

function artifactStableKey(artifact) {
  return `${artifact.role}\u0000${artifact.kind}\u0000${artifact.path}`;
}

function failureStableKey(failure) {
  return [
    failure.failure_class ?? "",
    failure.failure_reason ?? "",
    failure.target ?? "",
    failure.work_unit ?? "",
    failure.child_target ?? "",
    failure.label ?? "",
    failure.kind ?? "",
    failure.message ?? failure.headline ?? "",
  ].join("\u0000");
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
    if (target.run_root !== undefined) {
      requireString(target.run_root, `${targetLabel}.run_root`);
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
  requireString(summary.result_root, `${file}.result_root`);
  requireString(summary.run_id, `${file}.run_id`);
  requireString(summary.run_root, `${file}.run_root`);
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
    summary.failure_reason,
    `${file}.failure_reason`,
    toolRunFailureReasons,
  );
  if (summary.status === "fail") {
    if (summary.failure_class === null) {
      throw new Error(`${file}.failure_class must be non-null when status is fail`);
    }
    if (summary.failure_reason === null) {
      throw new Error(`${file}.failure_reason must be non-null when status is fail`);
    }
  }
  if (summary.status === "pass") {
    if (summary.failure_class !== null) {
      throw new Error(`${file}.failure_class must be null when status is pass`);
    }
    if (summary.failure_reason !== null) {
      throw new Error(`${file}.failure_reason must be null when status is pass`);
    }
  }
  requireSorted(
    requireObjectArray(summary.failures, `${file}.failures`),
    `${file}.failures`,
    failureStableKey,
    "failure class and target",
  );
  validateToolRunSlowest(summary.slowest, `${file}.slowest`);
  requireObjectArray(summary.warnings, `${file}.warnings`);
  requireStringArray(summary.rerun_commands, `${file}.rerun_commands`);
  if (summary.scheduler_timing !== null) {
    requireObject(summary.scheduler_timing, `${file}.scheduler_timing`);
  }
  const extensions = requireObject(summary.extensions, `${file}.extensions`);
  for (const key of Object.keys(extensions)) {
    if (!extensionKeyPattern.test(key)) {
      throw new Error(`${file}.extensions has invalid extension key ${key}`);
    }
  }
  validateSchemaSync(toolRunSummarySchemaID, summary);
}

function validateTestAccountingClassificationShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, testAccountingClassificationKeys, file);
  assertRequiredKeys(manifest, testAccountingClassificationKeys, file);
  validateSchemaSync(testAccountingClassificationSchemaID, manifest);
}

function validateProjectionProviderEntry(entry, label, seen) {
  const providerID = requireString(entry.provider_id, `${label}.provider_id`, {
    pattern: snakeIDPattern,
  });
  seen.providerIDs.push(providerID);

  const schemaVersion = requireString(
    entry.schema_version,
    `${label}.schema_version`,
  );
  if (schemaVersion !== projectionProviderDescriptorVersion) {
    throw new Error(
      `${label}.schema_version must be ${projectionProviderDescriptorVersion}`,
    );
  }

  requireString(entry.source_owner_module, `${label}.source_owner_module`, {
    pattern: snakeIDPattern,
  });
  requireString(entry.projection_storage_owner_module, `${label}.projection_storage_owner_module`, {
    pattern: snakeIDPattern,
  });

  const viewSchemaIDs = requireStringArray(
    entry.view_schema_ids,
    `${label}.view_schema_ids`,
    {
      nonEmpty: true,
      pattern: /^cartulary\.view\.[a-z0-9_]+\.v[0-9]+$/,
    },
  );
  for (const viewSchemaID of viewSchemaIDs) {
    if (seen.viewSchemaIDs.has(viewSchemaID)) {
      throw new Error(
        `${label}.view_schema_ids duplicates view schema ${viewSchemaID}`,
      );
    }
    seen.viewSchemaIDs.add(viewSchemaID);
  }

  const projectionTableIDs = requireStringArray(
    entry.projection_table_ids,
    `${label}.projection_table_ids`,
    { nonEmpty: true, pattern: snakeIDPattern },
  );
  seen.projectionTableIDs.push(...projectionTableIDs);

  requireStringArray(
    entry.source_authorities,
    `${label}.source_authorities`,
    { nonEmpty: true, pattern: snakeIDPattern },
  );

  const capabilities = requireObject(
    entry.capabilities,
    `${label}.capabilities`,
  );
  assertObjectKeys(
    capabilities,
    projectionProviderCapabilityKeys,
    `${label}.capabilities`,
  );
  assertRequiredKeys(
    capabilities,
    projectionProviderCapabilityKeys,
    `${label}.capabilities`,
  );
  for (const capability of projectionProviderCapabilityKeys) {
    requireBoolean(capabilities[capability], `${label}.capabilities.${capability}`);
  }

  const restoreRebuild = requireEnum(
    entry.restore_rebuild,
    `${label}.restore_rebuild`,
    projectionProviderRestoreRebuildValues,
  );
  if (capabilities.restore_rebuild !== (restoreRebuild === "required")) {
    throw new Error(
      `${label}.capabilities.restore_rebuild must match restore_rebuild`,
    );
  }

  requireEnum(
    entry.status,
    `${label}.status`,
    projectionProviderStatusValues,
  );

  for (const [index, facadePackage] of requireStringArray(
    entry.facade_packages,
    `${label}.facade_packages`,
    { nonEmpty: true },
  ).entries()) {
    requireRepoRelativePath(
      facadePackage,
      `${label}.facade_packages[${index + 1}]`,
    );
  }

  requireStringArray(entry.rebuild_after, `${label}.rebuild_after`, {
    pattern: snakeIDPattern,
  });

  for (const [index, ref] of requireStringArray(
    entry.characterization_refs,
    `${label}.characterization_refs`,
  ).entries()) {
    requireRepoRelativePath(ref, `${label}.characterization_refs[${index + 1}]`, {
      extension: ".go",
    });
  }
}

function validateProjectionProviderManifestShape(file) {
  const manifest = readShapeFile(file, file);
  assertObjectKeys(manifest, projectionProviderManifestKeys, file);
  assertRequiredKeys(manifest, projectionProviderManifestKeys, file);
  requireSchemaID(manifest, projectionProviderManifestSchemaID, file);
  requirePositiveInteger(manifest.manifest_version, `${file}.manifest_version`);

  const authority = requireString(manifest.authority, `${file}.authority`);
  if (authority !== projectionProviderAuthority) {
    throw new Error(`${file}.authority must be ${projectionProviderAuthority}`);
  }

  requireRepoRelativePath(manifest.source_registry, `${file}.source_registry`, {
    extension: ".go",
  });

  const importPolicy = requireObject(
    manifest.import_policy,
    `${file}.import_policy`,
  );
  assertObjectKeys(
    importPolicy,
    projectionProviderImportPolicyKeys,
    `${file}.import_policy`,
  );
  assertRequiredKeys(
    importPolicy,
    projectionProviderImportPolicyKeys,
    `${file}.import_policy`,
  );

  const approvedRootImporters = requireStringArray(
    importPolicy.approved_root_importers,
    `${file}.import_policy.approved_root_importers`,
  );
  if (approvedRootImporters.length !== 0) {
    throw new Error(
      `${file}.import_policy.approved_root_importers must be empty`,
    );
  }

  const approvedAdapterPackages = requireStringArray(
    importPolicy.approved_adapter_packages,
    `${file}.import_policy.approved_adapter_packages`,
    { nonEmpty: true },
  );
  validateProjectionImportPolicyPackages(
    approvedAdapterPackages,
    `${file}.import_policy.approved_adapter_packages`,
  );

  const approvedContractPackages = requireStringArray(
    importPolicy.approved_contract_packages,
    `${file}.import_policy.approved_contract_packages`,
    { nonEmpty: true },
  );
  validateProjectionImportPolicyPackages(
    approvedContractPackages,
    `${file}.import_policy.approved_contract_packages`,
  );

  const seen = {
    providerIDs: [],
    projectionTableIDs: [],
    viewSchemaIDs: new Set(),
  };
  validateObjectArray(
    manifest.providers,
    `${file}.providers`,
    {
      nonEmpty: true,
      keys: projectionProviderEntryKeys,
      requiredKeys: projectionProviderEntryKeys,
    },
    (entry, label) => validateProjectionProviderEntry(entry, label, seen),
  );
  assertUnique(seen.providerIDs, `${file}.providers.provider_id`);
  assertUnique(seen.projectionTableIDs, `${file}.providers.projection_table_ids`);
}

function validateProjectionImportPolicyPackages(packagePaths, label) {
  requireSorted(packagePaths, label, (entry) => entry, "repo-relative package");
  for (const [index, packagePath] of packagePaths.entries()) {
    requireRepoRelativePath(packagePath, `${label}[${index + 1}]`);
  }
}

function validateGraphProjectionConformanceMatrixShape(file) {
  const matrix = readShapeFile(file, file);
  assertObjectKeys(matrix, graphProjectionMatrixKeys, file);
  assertRequiredKeys(matrix, graphProjectionMatrixKeys, file);
  requireSchemaID(matrix, graphProjectionConformanceMatrixSchemaID, file);
  requireRepoRelativePath(matrix.spec_path, `${file}.spec_path`, {
    extension: ".md",
  });
  if (matrix.spec_status !== "adopted/current") {
    throw new Error(`${file}.spec_status must be adopted/current`);
  }
  requirePositiveInteger(matrix.matrix_version, `${file}.matrix_version`);
  if (matrix.authority !== "adopted_graph_projection_nlspec") {
    throw new Error(`${file}.authority must be adopted_graph_projection_nlspec`);
  }

  const acceptanceIDs = [];
  const seenFixtureIDs = new Set();
  validateObjectArray(
    matrix.acceptance_criteria,
    `${file}.acceptance_criteria`,
    {
      nonEmpty: true,
      keys: graphProjectionAcceptanceKeys,
      requiredKeys: graphProjectionAcceptanceKeys,
    },
    (entry, label) => {
      const id = requireString(entry.id, `${label}.id`, {
        pattern: /^GP-AC-\d{3}$/,
      });
      acceptanceIDs.push(id);
      requireString(entry.owner, `${label}.owner`, {
        pattern: /^[a-z][a-z0-9_]*$/,
      });
      requireEnum(
        entry.coverage_status,
        `${label}.coverage_status`,
        graphProjectionCoverageStatuses,
      );
      const areas = requireStringArray(entry.areas, `${label}.areas`, {
        nonEmpty: true,
      });
      for (const area of areas) {
        if (!graphProjectionAreas.has(area)) {
          throw new Error(`${label}.areas contains invalid area ${area}`);
        }
      }
      requireStringArray(entry.evidence_selectors, `${label}.evidence_selectors`, {
        nonEmpty: true,
      });
      for (const fixtureID of requireStringArray(
        entry.fixture_ids,
        `${label}.fixture_ids`,
      )) {
        if (!/^GP-FIX-\d{3}$/.test(fixtureID)) {
          throw new Error(`${label}.fixture_ids contains invalid ${fixtureID}`);
        }
        seenFixtureIDs.add(fixtureID);
      }
    },
  );
  assertUnique(acceptanceIDs, `${file}.acceptance_criteria.id`);
  requireSorted(
    acceptanceIDs,
    `${file}.acceptance_criteria.id`,
    (entry) => entry,
    "GP-AC identifier",
  );
  const expectedAcceptanceIDs = Array.from({ length: 68 }, (_, index) =>
    `GP-AC-${String(index + 1).padStart(3, "0")}`,
  );
  if (acceptanceIDs.join("\n") !== expectedAcceptanceIDs.join("\n")) {
    throw new Error(`${file}.acceptance_criteria must list GP-AC-001 through GP-AC-068`);
  }

  const fixtureIDs = [];
  validateObjectArray(
    matrix.fixture_registry,
    `${file}.fixture_registry`,
    {
      nonEmpty: true,
      keys: graphProjectionFixtureKeys,
      requiredKeys: graphProjectionFixtureKeys,
    },
    (entry, label) => {
      const fixtureID = requireString(entry.fixture_id, `${label}.fixture_id`, {
        pattern: /^GP-FIX-\d{3}$/,
      });
      fixtureIDs.push(fixtureID);
      const fixturePath = requireRepoRelativePath(
        entry.fixture_path,
        `${label}.fixture_path`,
        { extension: ".json" },
      );
      const resolvedFixturePath = path.resolve(repoRoot, fixturePath);
      if (!existsSync(resolvedFixturePath)) {
        throw new Error(`${label}.fixture_path does not exist: ${fixturePath}`);
      }
      requireString(entry.coverage, `${label}.coverage`);
    },
  );
  assertUnique(fixtureIDs, `${file}.fixture_registry.fixture_id`);
  requireSorted(
    fixtureIDs,
    `${file}.fixture_registry.fixture_id`,
    (entry) => entry,
    "GP-FIX identifier",
  );
  const expectedFixtureIDs = Array.from({ length: 23 }, (_, index) =>
    `GP-FIX-${String(index + 1).padStart(3, "0")}`,
  );
  if (fixtureIDs.join("\n") !== expectedFixtureIDs.join("\n")) {
    throw new Error(`${file}.fixture_registry must list GP-FIX-001 through GP-FIX-023`);
  }
  for (const fixtureID of seenFixtureIDs) {
    if (!fixtureIDs.includes(fixtureID)) {
      throw new Error(`${file}.acceptance_criteria references unknown fixture ${fixtureID}`);
    }
  }
}

function schemaIDFromFile(file) {
  const base = path.basename(file);
  if (!base.endsWith(".schema.json")) {
    throw new Error(`${file} must end with .schema.json`);
  }
  return base.slice(0, -".schema.json".length);
}

function extensionSchemaIsShared(value) {
  return (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    value.$ref === sharedExtensionsRef &&
    Object.keys(value).length === 1
  );
}

function schemaDeclaresSchemaID(schema, schemaID) {
  if (schema?.properties?.schema_id?.const === schemaID) {
    return true;
  }
  return (schema?.allOf ?? []).some(
    (entry) => entry?.properties?.schema_id?.const === schemaID,
  );
}

function schemaRequiresSchemaID(schema) {
  if ((schema?.required ?? []).includes("schema_id")) {
    return true;
  }
  return (schema?.allOf ?? []).some((entry) =>
    (entry?.required ?? []).includes("schema_id"),
  );
}

function schemaIsClosed(schema) {
  if (schema?.additionalProperties === false) {
    return true;
  }
  return (schema?.allOf ?? []).some(
    (entry) => schedulerSummaryCommonSchemaIDs.has(entry?.$ref),
  );
}

function schemaIsAliasOnly(schema) {
  const keys = Object.keys(schema).sort();
  return (
    keys.length === 3 &&
    keys[0] === "$id" &&
    keys[1] === "$ref" &&
    keys[2] === "$schema"
  );
}

function validateExtensionProperties(schema, label) {
  if (!schema || typeof schema !== "object" || Array.isArray(schema)) {
    return;
  }
  if (schema.properties?.extensions !== undefined) {
    if (!extensionSchemaIsShared(schema.properties.extensions)) {
      throw new Error(
        `${label}.properties.extensions must reference ${sharedExtensionsRef}`,
      );
    }
  }
  for (const [key, value] of Object.entries(schema)) {
    if (key === "properties" && value && typeof value === "object") {
      for (const [propertyName, propertySchema] of Object.entries(value)) {
        validateExtensionProperties(propertySchema, `${label}.properties.${propertyName}`);
      }
      continue;
    }
    if (Array.isArray(value)) {
      value.forEach((entry, index) => {
        validateExtensionProperties(entry, `${label}.${key}[${index + 1}]`);
      });
      continue;
    }
    validateExtensionProperties(value, `${label}.${key}`);
  }
}

function validateSchemaAttachmentPolicy(root) {
  const schemaDir = repoFile(root, "tools/schemas");
  for (const name of readdirSync(schemaDir).sort((left, right) =>
    left.localeCompare(right),
  )) {
    if (!name.endsWith(".schema.json")) {
      continue;
    }
    const file = path.join(schemaDir, name);
    const schema = readShapeFile(file, file);
    const schemaID = schemaIDFromFile(file);
    if (schema.$id !== schemaID) {
      throw new Error(`${file} $id must match ${schemaID}`);
    }
    validateExtensionProperties(schema, file);
    if (supportSchemaIDs.has(schemaID)) {
      continue;
    }
    if (schemaIsAliasOnly(schema)) {
      throw new Error(`${file} must not be an alias-only public schema`);
    }
    if (!schemaRequiresSchemaID(schema)) {
      throw new Error(`${file} must require schema_id`);
    }
    if (!schemaDeclaresSchemaID(schema, schemaID)) {
      throw new Error(`${file} must constrain schema_id to ${schemaID}`);
    }
    if (!schemaIsClosed(schema)) {
      throw new Error(`${file} must be closed at the top level`);
    }
  }
}

function validateHarnessRequirementIDs(root) {
  const file = repoFile(root, "docs/testing-harness-nlspec.md");
  const source = readFileSync(file, "utf8");
  const seen = new Set();
  const duplicates = new Set();
  for (const match of source.matchAll(/\*\*(TH-HARNESS-REQ-\d+)\*\*/g)) {
    const id = match[1];
    if (seen.has(id)) {
      duplicates.add(id);
    }
    seen.add(id);
  }
  if (duplicates.size > 0) {
    throw new Error(`${file} contains duplicate requirement IDs: ${[...duplicates].sort().join(", ")}`);
  }
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
    case "scheduler-manifest":
      validateSchedulerManifestShape(file);
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
    case "fallow-static-summary":
      validateSchemaSync(
        fallowStaticSummarySchemaID,
        readShapeFile(file, file),
      );
      return;
    case "agent-finalize-summary":
      validateSchemaSync(
        agentFinalizeSummarySchemaID,
        readShapeFile(file, file),
      );
      return;
    case "test-accounting-classification":
      validateTestAccountingClassificationShape(file);
      return;
    case "projection-provider-manifest":
      validateProjectionProviderManifestShape(file);
      return;
    case "graph-projection-conformance-matrix":
      validateGraphProjectionConformanceMatrixShape(file);
      return;
    case "migration-history":
      validateMigrationHistoryManifestShape(file);
      return;
    case "schema-object-ownership":
      validateSchemaObjectOwnershipManifestShape(file);
      return;
    default:
      throw new Error(`unknown json shape kind ${kind}`);
  }
}

function validateAll(root) {
  validateSchemaAttachmentPolicy(root);
  validateHarnessRequirementIDs(root);
  validatePhaseRegistryShape(repoFile(root, "tools/phase_registry.json"));
  validatePhaseRegistry(root);
  validateFrontendPhaseArtifacts(root);
  validateFrontendSchemaArtifacts(root);
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
  loadExecutionTopology({
    root,
    manifestPath: repoFile(root, "tools/execution_topology_manifest.json"),
  });
  quickCheckRenderIndex({
    topology: repoFile(root, "tools/execution_topology_manifest.json"),
  });

  validateTaskSurfaceShape(repoFile(root, "tools/task_surface_manifest.json"));
  loadTaskSurfaceManifest(repoFile(root, "tools/task_surface_manifest.json"));
  validateSchedulerManifestShape(
    repoFile(root, "tools/scheduler_manifest.json"),
  );
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
  validateTestAccountingClassificationShape(
    repoFile(root, "tools/test_accounting_classification.json"),
  );
  validateProjectionProviderManifestShape(
    repoFile(root, "contracts/projection-providers/index.json"),
  );
  validateGraphProjectionConformanceMatrixShape(
    repoFile(root, "contracts/graph-projection/conformance_matrix.v1.json"),
  );
  validateMigrationHistory(root);
  validateSchemaObjectOwnership(root);
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
