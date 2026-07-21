#!/usr/bin/env node
import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  loadBrowserBatchManifest,
  validateBrowserBatchManifestShape,
} from "../browser/browser-batch-manifest.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import {
  executionTopologySchemaID,
  loadExecutionTopology,
  taskSurfaceOwnerSchemaID,
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
import {
  validateHarnessHelperOwnership,
  validateSchemaAttachmentPolicy,
} from "./schema-attachment-validation.mjs";
import { validateSchedulerManifestShape } from "../scheduler/scheduler-manifest.mjs";
import {
  loadSchedulerResourceRegistry,
  validateSchedulerResourceRegistrySemantics,
} from "../scheduler/scheduler-resources.mjs";
import {
  collectTaskSurfaceManifestErrors,
  loadTaskSurfaceManifest,
} from "./task-surface/index.mjs";
import { quickCheckRenderIndex } from "./render-execution-topology-artifacts.mjs";
import { validateVerificationContracts } from "../test-catalog/verification-contracts.mjs";
import { validateTestCatalog } from "../test-catalog/test-catalog.mjs";
import { validateTestCatalogImportBoundary } from "../test-catalog/import-boundary.mjs";
import { scanExecutableDocumentationReads } from "../test-catalog/documentation-boundary.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");
const generatedArtifactPolicySchemaID =
  "cartulary.generated_artifact_policy.v1";
const contractFamilyRegistrySchemaID = "cartulary.contract_family_registry.v1";
const frontendImportBoundariesSchemaID =
  "cartulary.frontend_import_boundaries.v2";
const bootstrapAdminSchemaID = "cartulary.bootstrap_admin.v1";
const serviceBackedMakeTargetBaselineSchemaID =
  "cartulary.scheduler_work_unit_duration_baselines.v2";
const toolRunSummarySchemaID = "cartulary.tool_run_summary.v5";
const fallowReachabilityOwnerSchemaID =
  "cartulary.fallow_reachability_owner.v1";
const fallowStaticSummarySchemaID = "cartulary.fallow_static_summary.v2";
const agentFinalizeSummarySchemaID = "cartulary.agent_finalize_summary.v3";
const testSupportInventorySchemaID = "cartulary.test_support_inventory.v1";
const projectionProviderManifestSchemaID =
  "cartulary.projection_provider_manifest.v4";
const graphProjectionConformanceMatrixSchemaID =
  "cartulary.graph_projection_conformance_matrix.v1";
const graphProjectionFixtureCorpusSchemaID =
  "cartulary.graph_projection_fixture_corpus.v1";
const graphProjectionFixtureManifestSchemaID =
  "cartulary.graph_projection_fixture_manifest.v1";
const networkFlowFixtureManifestSchemaID =
  "cartulary.network_flow_fixture_manifest.v1";
const networkFlowActivityAccountingSchemaID =
  "cartulary.network_flow_activity_accounting.v2";
const networkFlowContractIndexSchemaID =
  "cartulary.network_flow_contract_index.v1";
const networkFlowTimezoneRulesetProvenanceSchemaID =
  "cartulary.network_flow_timezone_ruleset_provenance.v1";
const frontendVisualFixtureRegistrySchemaID =
  "cartulary.frontend_visual_fixture_registry.v4";
const schedulerSummaryCommonSchemaID = "cartulary.scheduler_summary.common.v10";
const schedulerSummaryCommonSchemaIDs = new Set([schedulerSummaryCommonSchemaID]);

const makeTargetPattern = /^[A-Za-z0-9_.-]+$/;
const snakeIDPattern = /^[a-z][a-z0-9_]*$/;
const sha256Pattern = /^[a-f0-9]{64}$/;
const networkFlowFixtureIDPattern = /^NF-FIX-\d{3}-[a-z0-9][a-z0-9-]*$/;
const networkFlowFixtureBaseIDPattern = /^NF-FIX-\d{3}$/;
const networkFlowAcceptanceIDPattern = /^NF-AC-\d{3}$/;
const manifestRelativePathPattern =
  /^(?!\/)(?!.*(?:^|\/)\.\.(?:\/|$))(?!.*\/\/)[A-Za-z0-9._@+-]+(?:\/[A-Za-z0-9._@+-]+)*$/;
const topologyTopLevelKeys = new Set([
  "schema_id",
  "runtime_profiles",
  "resource_profiles",
  "fixture_profiles",
  "generated_outputs",
  "runtime_binaries",
  "execution_dependencies",
  "go_targets",
  "task_surface_owner",
  "check_schedules",
  "service_backed_schedules",
  "sequence_resource_profiles",
  "sequence_schedules",
  "browser_e2e_batch",
]);
const generatedArtifactPolicyKeys = new Set([
  "schema_id",
  "ignored_sentinel_filenames",
  "generated_roots",
  "generated_files",
  "lint_scope_checks",
]);
const contractFamilyRegistryKeys = new Set([
  "$schema",
  "schema_id",
  "registry_id",
  "note",
  "families",
]);
const contractFamilyEntryKeys = new Set([
  "family_id",
  "contract_root",
  "generation_status",
  "go_name",
  "ts_name",
  "output_order",
  "owner_document",
  "owner_sections",
  "generated_outputs",
  "typescript_runtime_artifact_prefixes",
  "activation_dependency_ids",
  "description",
]);
const networkFlowAccountingKeys = new Set([
  "$schema",
  "schema_id",
  "profile_id",
  "owner_id",
  "verification_ids",
  "contract_registry",
  "fixture_accounting",
  "acceptance_accounting",
  "drift_accounting",
]);
const networkFlowContractRegistryAccountingKeys = new Set([
  "path",
  "family_id",
  "contract_root",
  "planned_activation_dependency_ids",
  "generated_symbol_markers",
]);
const networkFlowDependencyLocatorAccountingKeys = new Set([
  "table_caption",
  "expected_count",
  "blocked_tokens",
  "required_dependencies",
  "required_locator_fragments",
]);
const networkFlowDependencyLocatorFragmentKeys = new Set([
  "dependency",
  "fragments",
]);
const networkFlowFixtureAccountingKeys = new Set([
  "expected_count",
  "first_id",
  "last_id",
  "manifest_root",
]);
const networkFlowAcceptanceAccountingKeys = new Set([
  "expected_count",
  "first_id",
  "last_id",
  "selector_prefix",
  "tracker_row_prefix",
  "matrix_source",
  "rows",
]);
const networkFlowDriftAccountingKeys = new Set([
  "scratch_manifest",
  "required_copy_paths",
  "required_public_targets",
]);
const testSupportInventoryKeys = new Set([
  "schema_id",
  "go_support_roots",
  "shared_data_roots",
]);
const goSupportRootKeys = new Set([
  "path",
  "owner",
  "posture",
  "runtime_scan",
  "support_scan",
  "service_starting",
  "rationale",
]);
const sharedDataRootKeys = new Set([
  "path",
  "owner",
  "posture",
  "data_kind",
  "file_roles",
  "owner_semantic_data_policy",
  "retained_path_policy",
  "rationale",
]);
const networkFlowContractIndexKeys = new Set([
  "$schema",
  "schema_id",
  "profile_id",
  "contract_major",
  "document_version",
  "family_id",
  "owner_id",
  "verification_ids",
  "contract_files",
  "public_schema_ids",
  "closure_policy",
]);
const networkFlowContractFilesKeys = new Set([
  "routes",
  "schemas",
  "errors",
  "timezone_provenance",
  "key_rings",
  "mapping_registry",
  "presentation",
]);
const networkFlowClosurePolicyKeys = new Set([
  "objects_closed_by_default",
  "dynamic_maps_must_name_key_pattern",
  "raw_source_values_forbidden_outside_diagnostics",
  "generated_outputs_blocked_until",
]);
const networkFlowRouteContractKeys = new Set([
  "schema_id",
  "profile_id",
  "contract_major",
  "route_root",
  "import_integration",
  "routes",
]);
const networkFlowImportIntegrationKeys = new Set([
  "target_kind",
  "target_table_schema_id",
  "resource_ref_kind",
  "default_source_profile_id",
  "default_parser_profile_id",
  "default_unknown_column_policy",
  "owner_facade",
  "mapping_preview_route",
  "owner_facade_operations",
]);
const networkFlowFacadeOperationKeys = new Set([
  "operation",
  "request_schema_id",
  "success_schema_id",
]);
const networkFlowRouteEntryKeys = new Set([
  "route_id",
  "method",
  "path",
  "auth_context",
  "request_schema_id",
  "continuation_schema_id",
  "success_schema_id",
  "success_http_statuses",
  "idempotency",
  "primary_errors",
  "audit_event",
]);
const networkFlowErrorContractKeys = new Set([
  "schema_id",
  "profile_id",
  "contract_major",
  "errors",
  "reason_registries",
  "retry_actions",
]);
const networkFlowErrorEntryKeys = new Set([
  "code",
  "scope",
  "http_status",
  "retry_action",
]);
const networkFlowReasonRegistryKeys = new Set(["error_code", "reason_codes"]);
const frontendBoundaryKeys = new Set([
  "schema_id",
  "scan_roots",
  "scan_excludes",
  "singleton_imports",
  "rules",
  "raw_design_token_literal_checks",
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
  "source_record_types",
  "source_authority_modules",
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
  "projection_provider_descriptor.v3";
const projectionProviderSourceAuthorityModules = new Set([
  "assessments",
  "artifacts",
  "auth",
  "collaboration",
  "database_migrations",
  "deployment_admin",
  "entities",
  "evidence",
  "graphprojection",
  "harness_support",
  "imports",
  "incidentbundles",
  "incidents",
  "indicators",
  "jobapi",
  "links",
  "networkflow",
  "parties",
  "platform_jobs",
  "projections",
  "recovery",
  "reference_data",
  "reportcomposition",
  "reporting",
  "revisions",
  "savedviews",
  "tasksdecisions",
  "timeline",
  "viewschemas",
  "workbook",
]);
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
const graphProjectionFixtureCorpusKeys = new Set([
  "schema_id",
  "spec_path",
  "fixtures",
]);
const graphProjectionCorpusFixtureKeys = new Set([
  "fixture_id",
  "coverage",
  "input_kind",
]);
const networkFlowFixtureManifestKeys = new Set([
  "schema_id",
  "manifest_version",
  "fixture_id",
  "profile_id",
  "freeze",
  "verification_ids",
  "source_files",
  "expected_artifacts",
  "transcript_files",
  "acceptance_ids",
  "execution_selectors",
  "source_bundle_sha256",
  "expected_bundle_sha256",
  "extensions",
]);
const networkFlowFixtureFreezeKeys = new Set([
  "status",
  "revision",
  "change_policy",
]);
const networkFlowTimezoneProvenanceKeys = new Set([
  "schema_id",
  "ruleset_id",
  "profile_id",
  "iana_version",
  "release",
  "source_archive",
  "detached_signature",
  "license",
  "embedded_file_hashes",
  "verification_ids",
  "conformance_policy",
]);
const networkFlowTimezoneReleaseKeys = new Set([
  "released_at",
  "release_date",
  "release_index_url",
  "release_archive_index_url",
]);
const networkFlowTimezoneArchiveKeys = new Set([
  "distribution",
  "file_name",
  "url",
  "media_type",
  "size_bytes",
  "sha256",
]);
const networkFlowTimezoneSignatureKeys = new Set([
  "file_name",
  "url",
  "media_type",
  "size_bytes",
  "sha256",
  "openpgp_fingerprint",
  "openpgp_key_id",
  "signature_created_at",
]);
const networkFlowTimezoneLicenseKeys = new Set([
  "source_path",
  "summary",
  "sha256",
  "data_only_distribution_requires_bsd_3_clause_exception",
]);
const networkFlowTimezoneEmbeddedFileKeys = new Set([
  "path",
  "size_bytes",
  "sha256",
]);
const networkFlowTimezoneConformancePolicyKeys = new Set([
  "host_timezone_database_authoritative",
  "host_locale_authoritative",
  "latest_url_allowed",
  "verification_required_before_use",
  "allowed_internal_ruleset_substitution",
]);
const networkFlowFixtureSourceFileKeys = new Set([
  "logical_path",
  "media_type",
  "size_bytes",
  "sha256",
  "role",
  "newline_policy",
]);
const networkFlowFixtureExpectedArtifactKeys = new Set([
  "logical_path",
  "media_type",
  "size_bytes",
  "sha256",
  "role",
]);
const networkFlowFixtureTranscriptFileKeys = new Set([
  "logical_path",
  "media_type",
  "size_bytes",
  "sha256",
  "transcript_kind",
]);
const graphProjectionAcceptanceCount = 69;
const graphProjectionFixtureCount = 36;
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
  "step_accounting",
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
const toolRunFileArtifactKeys = new Set(["role", "path_kind", "format", "path"]);
const toolRunDirectoryArtifactKeys = new Set(["role", "path_kind", "path"]);
const toolRunArtifactPathKinds = new Set(["file", "directory"]);
const toolRunArtifactFormats = new Set([
  "json",
  "jsonl",
  "log",
  "markdown",
  "sarif",
  "text",
]);
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
  "test_accounting_unmapped",
  "artifact_error",
  "cleanup_error",
  "duration_baseline_drift",
  "timeout_failure",
  "cancelled_or_interrupted",
  "unknown_failure",
]);
const toolRunCountKeys = new Set([
  "steps",
  "tests",
  "failed",
  "non_test",
  "non_test_failed",
  "packages",
]);
const toolRunStepAccountingKeys = new Set([
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
const goSupportPostures = new Set(["shared", "owner_local", "platform_facade"]);
const scanTreatments = new Set(["included", "excluded"]);
const supportScanTreatments = new Set(["included", "not_applicable"]);
const sharedDataKinds = new Set([
  "bootstrap_manifest",
  "http_envelope",
  "otel_evidence",
  "platform_config",
  "platform_diagnostics",
  "websocket_protocol",
]);
const sharedDataPostures = new Set(["shared"]);
const sharedDataPolicies = new Set([
  "reject_unclassified",
  "adopted_external_evidence",
]);
const retainedPathPolicies = new Set([
  "stable",
  "move_only_with_owner_accounting",
]);
const sharedDataFileRoles = new Set(["fixture", "golden", "manifest", "placeholder"]);
const discoveredGoSupportRoots = Object.freeze([
  "internal/platform/contracttest",
  "internal/testutil",
  "tools",
]);
const sharedDataFacadeRoots = Object.freeze([
  "internal/testutil/fixtures",
  "internal/testutil/golden",
]);
const phaseShapedSupportPathPattern = /(?:^|\/)phase\d+(?:store)?test(?:\/|$)/;
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
  requireString(topology.task_surface_owner, `${file}.task_surface_owner`);
}

function validateTaskSurfaceOwnerShape(file) {
  const taskSurface = readShapeFile(file, file);
  validateSchemaSync(taskSurfaceOwnerSchemaID, taskSurface);
  const targets = requireObjectArray(
    taskSurface.targets,
    `${file}.targets`,
    {
      nonEmpty: true,
    },
  );
  assertUnique(
    targets.map((entry, index) =>
      requireString(
        entry.name,
        `${file}.targets[${index + 1}].name`,
        {
          pattern: makeTargetPattern,
        },
      ),
    ),
    `${file}.targets.name`,
  );
  const { schema_id: _schemaID, ...projection } = taskSurface;
  const errors = collectTaskSurfaceManifestErrors({
    schema_id: taskSurfaceSchemaID,
    ...projection,
  }, { requireSequenceTopology: false });
  if (errors.length > 0) {
    throw new Error(
      `${file} is invalid:\n${errors.map((error) => `  - ${error}`).join("\n")}`,
    );
  }
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

function validateContractFamilyRegistryShape(file) {
  const registry = readShapeFile(file, file);
  validateSchemaSync(contractFamilyRegistrySchemaID, registry);
  assertObjectKeys(registry, contractFamilyRegistryKeys, file);
  assertRequiredKeys(registry, contractFamilyRegistryKeys, file);
  requireSchemaID(registry, contractFamilyRegistrySchemaID, file);
  requireExact(
    requireString(registry.registry_id, `${file}.registry_id`),
    "cartulary.contract_families.v1",
    `${file}.registry_id`,
  );
  requireString(registry.note, `${file}.note`);

  const familyIDs = [];
  const contractRoots = [];
  const goNames = [];
  const tsNames = [];
  const outputOrders = [];
  const activeIDsByOrder = [];
  const plannedIDs = [];
  validateObjectArray(
    registry.families,
    `${file}.families`,
    {
      nonEmpty: true,
      keys: contractFamilyEntryKeys,
      requiredKeys: contractFamilyEntryKeys,
    },
    (entry, label) => {
      const familyID = requireString(entry.family_id, `${label}.family_id`, {
        pattern: /^[a-z][a-z0-9-]*$/,
      });
      familyIDs.push(familyID);
      const contractRoot = requireRepoRelativePath(
        entry.contract_root,
        `${label}.contract_root`,
      );
      if (!contractRoot.startsWith("contracts/")) {
        throw new Error(`${label}.contract_root must start with contracts/`);
      }
      contractRoots.push(contractRoot);
      if (!existsSync(repoFile(repoRoot, contractRoot))) {
        throw new Error(`${label}.contract_root does not exist: ${contractRoot}`);
      }
      goNames.push(
        requireString(entry.go_name, `${label}.go_name`, {
          pattern: /^[A-Z][A-Za-z0-9]*$/,
        }),
      );
      tsNames.push(
        requireString(entry.ts_name, `${label}.ts_name`, {
          pattern: /^[a-z][A-Za-z0-9]*$/,
        }),
      );
      const outputOrder = requireInteger(entry.output_order, `${label}.output_order`, {
        min: 0,
      });
      outputOrders.push(String(outputOrder).padStart(4, "0"));
      const ownerDocument = requireRepoRelativePath(
        entry.owner_document,
        `${label}.owner_document`,
        { extension: ".md" },
      );
      void ownerDocument;
      requireStringArray(entry.owner_sections, `${label}.owner_sections`, {
        nonEmpty: true,
      });
      const generatedOutputs = requireStringArray(
        entry.generated_outputs,
        `${label}.generated_outputs`,
        { nonEmpty: true },
      );
      for (const generatedOutput of generatedOutputs) {
        requireRepoRelativePath(generatedOutput, `${label}.generated_outputs[]`);
        if (
          !generatedOutput.startsWith("internal/gen/contracts/") &&
          !generatedOutput.startsWith("packages/protocol-ts/src/generated/")
        ) {
          throw new Error(
            `${label}.generated_outputs[] must target the contract generated roots`,
          );
        }
      }
      const typeScriptRuntimePrefixes = requireStringArray(
        entry.typescript_runtime_artifact_prefixes,
        `${label}.typescript_runtime_artifact_prefixes`,
        { nonEmpty: true },
      );
      assertUnique(
        typeScriptRuntimePrefixes,
        `${label}.typescript_runtime_artifact_prefixes`,
      );
      for (const prefix of typeScriptRuntimePrefixes) {
        if (
          !prefix.startsWith(`${contractRoot}/`) ||
          prefix.includes("..") ||
          !/^contracts\/[A-Za-z0-9._@+-]+(?:\/[A-Za-z0-9._@+-]+)*\/?$/u.test(prefix)
        ) {
          throw new Error(
            `${label}.typescript_runtime_artifact_prefixes[] must stay within ${contractRoot}`,
          );
        }
      }
      const activationDependencies = requireStringArray(
        entry.activation_dependency_ids,
        `${label}.activation_dependency_ids`,
      );
      requireString(entry.description, `${label}.description`);
      const generationStatus = requireEnum(
        entry.generation_status,
        `${label}.generation_status`,
        new Set(["active", "planned"]),
      );
      if (generationStatus === "active") {
        if (activationDependencies.length !== 0) {
          throw new Error(
            `${label}.activation_dependency_ids must be empty for active families`,
          );
        }
        activeIDsByOrder[outputOrder] = familyID;
      } else {
        if (activationDependencies.length === 0) {
          throw new Error(
            `${label}.activation_dependency_ids must not be empty for planned families`,
          );
        }
        plannedIDs.push(familyID);
      }
    },
  );

  assertUnique(familyIDs, `${file}.families.family_id`);
  assertUnique(contractRoots, `${file}.families.contract_root`);
  assertUnique(goNames, `${file}.families.go_name`);
  assertUnique(tsNames, `${file}.families.ts_name`);
  assertUnique(outputOrders, `${file}.families.output_order`);

  if (!familyIDs.includes("network-flow")) {
    throw new Error(`${file}.families must declare network-flow`);
  }
  const expectedBaseActiveIDs = [
    "openapi",
    "ws",
    "view-schemas",
    "errors",
    "extensions",
  ];
  const expectedActiveIDVariants = [
    expectedBaseActiveIDs.join("\n"),
    [...expectedBaseActiveIDs, "network-flow"].join("\n"),
  ];
  if (!expectedActiveIDVariants.includes(activeIDsByOrder.join("\n"))) {
    throw new Error(
      `${file}.families active output_order must be ${expectedBaseActiveIDs.join(", ")} with optional network-flow`,
    );
  }
  if (plannedIDs.length > 1 || (plannedIDs.length === 1 && plannedIDs[0] !== "network-flow")) {
    throw new Error(`${file}.families may declare only network-flow as planned`);
  }
}

function expectedNetworkFlowIDs(prefix, count) {
  return Array.from(
    { length: count },
    (_entry, index) => `${prefix}${String(index + 1).padStart(3, "0")}`,
  );
}

function assertExactIDSet(actualIDs, expectedIDs, label) {
  const actual = [...actualIDs].sort();
  const expected = [...expectedIDs].sort();
  if (actual.join("\n") !== expected.join("\n")) {
    const missing = expected.filter((id) => !actualIDs.has(id));
    const extra = actual.filter((id) => !expectedIDs.has(id));
    throw new Error(
      `${label} mismatch; missing=${missing.join(",") || "none"} extra=${extra.join(",") || "none"}`,
    );
  }
}

function extractUniqueMatches(source, pattern) {
  return new Set([...source.matchAll(pattern)].map((match) => match[0]));
}

function extractNetworkFlowFixtureBaseIDs(source) {
  const fixtureIDs = extractUniqueMatches(
    source,
    /NF-FIX-\d{3}-[a-z0-9][a-z0-9-]*/gu,
  );
  const baseIDs = new Set();
  const seenFullByBase = new Map();
  for (const fixtureID of fixtureIDs) {
    if (!networkFlowFixtureIDPattern.test(fixtureID)) {
      throw new Error(`invalid Network Flow fixture ID ${fixtureID}`);
    }
    const baseID = fixtureID.slice(0, "NF-FIX-000".length);
    if (seenFullByBase.has(baseID) && seenFullByBase.get(baseID) !== fixtureID) {
      throw new Error(
        `Network Flow fixture base ID ${baseID} maps to multiple fixture slugs`,
      );
    }
    seenFullByBase.set(baseID, fixtureID);
    baseIDs.add(baseID);
  }
  return baseIDs;
}

function extractMarkdownTableByCaption(source, caption) {
  const marker = `**${caption}**`;
  const markerIndex = source.indexOf(marker);
  if (markerIndex === -1) {
    throw new Error(`missing Markdown table caption ${caption}`);
  }
  const lines = source.slice(markerIndex + marker.length).split(/\r?\n/u);
  const tableLines = [];
  let inTable = false;
  for (const line of lines) {
    if (line.trim().startsWith("|")) {
      inTable = true;
      tableLines.push(line.trim());
      continue;
    }
    if (inTable) {
      break;
    }
  }
  if (tableLines.length < 3) {
    throw new Error(`Markdown table ${caption} is missing header or data rows`);
  }
  return tableLines.slice(2).map((line) =>
    line
      .slice(1, -1)
      .split("|")
      .map((cell) => cell.trim()),
  );
}

function pathIsCoveredByAny(pathValue, candidates) {
  return candidates.some(
    (candidate) => pathValue === candidate || pathValue.startsWith(`${candidate}/`),
  );
}

function generatedOutputsContainMarker(root, generatedOutputs, markers) {
  for (const generatedOutput of generatedOutputs) {
    const outputPath = repoFile(root, generatedOutput);
    if (!existsSync(outputPath) || !lstatSync(outputPath).isFile()) {
      throw new Error(`contract generated output missing: ${generatedOutput}`);
    }
    const output = readFileSync(outputPath, "utf8");
    if (markers.some((marker) => output.includes(marker))) {
      return true;
    }
  }
  return false;
}

function taskSurfaceTargetNames(taskSurface) {
  const names = new Set();
  for (const target of requireObjectArray(taskSurface.targets, "task_surface.targets")) {
    names.add(requireString(target.name, "task_surface.targets[].name"));
  }
  for (const recipeName of Object.keys(requireObject(taskSurface.make_recipes, "task_surface.make_recipes"))) {
    names.add(recipeName);
  }
  return names;
}

function validateNetworkFlowActivityAccountingShape(file) {
  const accounting = readShapeFile(file, file);
  validateSchemaSync(networkFlowActivityAccountingSchemaID, accounting);
  assertObjectKeys(accounting, networkFlowAccountingKeys, file);
  assertRequiredKeys(accounting, networkFlowAccountingKeys, file);
  requireSchemaID(accounting, networkFlowActivityAccountingSchemaID, file);
  requireExact(accounting.profile_id, "network_flow_activity", `${file}.profile_id`);

  requireExact(accounting.owner_id, "module.networkflow", `${file}.owner_id`);
  requireExactArray(
    requireStringArray(accounting.verification_ids, `${file}.verification_ids`, {
      nonEmpty: true,
    }),
    ["module.networkflow.verification.contract_accounting"],
    `${file}.verification_ids`,
  );
  const contractRegistry = requireObject(
    accounting.contract_registry,
    `${file}.contract_registry`,
  );
  assertObjectKeys(
    contractRegistry,
    networkFlowContractRegistryAccountingKeys,
    `${file}.contract_registry`,
  );
  assertRequiredKeys(
    contractRegistry,
    networkFlowContractRegistryAccountingKeys,
    `${file}.contract_registry`,
  );
  const registryPath = requireRepoRelativePath(
    contractRegistry.path,
    `${file}.contract_registry.path`,
  );
  requireExact(
    contractRegistry.family_id,
    "network-flow",
    `${file}.contract_registry.family_id`,
  );
  requireExact(
    contractRegistry.contract_root,
    "contracts/network-flow",
    `${file}.contract_registry.contract_root`,
  );
  const plannedActivationIDs = requireStringArray(
    contractRegistry.planned_activation_dependency_ids,
    `${file}.contract_registry.planned_activation_dependency_ids`,
    { nonEmpty: true },
  );
  const generatedMarkers = requireStringArray(
    contractRegistry.generated_symbol_markers,
    `${file}.contract_registry.generated_symbol_markers`,
    { nonEmpty: true },
  );

  const fixtureAccounting = requireObject(
    accounting.fixture_accounting,
    `${file}.fixture_accounting`,
  );
  assertObjectKeys(
    fixtureAccounting,
    networkFlowFixtureAccountingKeys,
    `${file}.fixture_accounting`,
  );
  assertRequiredKeys(
    fixtureAccounting,
    networkFlowFixtureAccountingKeys,
    `${file}.fixture_accounting`,
  );
  const expectedFixtureCount = requireInteger(
    fixtureAccounting.expected_count,
    `${file}.fixture_accounting.expected_count`,
    { min: 1 },
  );
  const fixtureFirstID = requireString(
    fixtureAccounting.first_id,
    `${file}.fixture_accounting.first_id`,
    { pattern: networkFlowFixtureBaseIDPattern },
  );
  const fixtureLastID = requireString(
    fixtureAccounting.last_id,
    `${file}.fixture_accounting.last_id`,
    { pattern: networkFlowFixtureBaseIDPattern },
  );
  requireExact(fixtureFirstID, "NF-FIX-001", `${file}.fixture_accounting.first_id`);
  requireExact(fixtureLastID, "NF-FIX-028", `${file}.fixture_accounting.last_id`);
  requireExact(
    fixtureAccounting.manifest_root,
    "fixtures/network-flow",
    `${file}.fixture_accounting.manifest_root`,
  );

  const acceptanceAccounting = requireObject(
    accounting.acceptance_accounting,
    `${file}.acceptance_accounting`,
  );
  assertObjectKeys(
    acceptanceAccounting,
    networkFlowAcceptanceAccountingKeys,
    `${file}.acceptance_accounting`,
  );
  assertRequiredKeys(
    acceptanceAccounting,
    networkFlowAcceptanceAccountingKeys,
    `${file}.acceptance_accounting`,
  );
  const expectedAcceptanceCount = requireInteger(
    acceptanceAccounting.expected_count,
    `${file}.acceptance_accounting.expected_count`,
    { min: 1 },
  );
  const acceptanceFirstID = requireString(
    acceptanceAccounting.first_id,
    `${file}.acceptance_accounting.first_id`,
    { pattern: networkFlowAcceptanceIDPattern },
  );
  const acceptanceLastID = requireString(
    acceptanceAccounting.last_id,
    `${file}.acceptance_accounting.last_id`,
    { pattern: networkFlowAcceptanceIDPattern },
  );
  requireExact(
    acceptanceFirstID,
    "NF-AC-001",
    `${file}.acceptance_accounting.first_id`,
  );
  requireExact(
    acceptanceLastID,
    "NF-AC-107",
    `${file}.acceptance_accounting.last_id`,
  );
  requireExact(
    acceptanceAccounting.selector_prefix,
    "network-flow/acceptance/",
    `${file}.acceptance_accounting.selector_prefix`,
  );
  requireExact(
    acceptanceAccounting.tracker_row_prefix,
    "NF-AC-",
    `${file}.acceptance_accounting.tracker_row_prefix`,
  );
  const matrixSource = requireRepoRelativePath(
    acceptanceAccounting.matrix_source,
    `${file}.acceptance_accounting.matrix_source`,
  );
  requireExact(
    matrixSource,
    "tools/test_families/module.networkflow.json",
    `${file}.acceptance_accounting.matrix_source`,
  );
  const acceptanceRows = requireObjectArray(
    acceptanceAccounting.rows,
    `${file}.acceptance_accounting.rows`,
    { nonEmpty: true },
  );
  if (acceptanceRows.length !== expectedAcceptanceCount) {
    throw new Error(
      `${file}.acceptance_accounting.rows must contain ${expectedAcceptanceCount} rows`,
    );
  }

  const driftAccounting = requireObject(
    accounting.drift_accounting,
    `${file}.drift_accounting`,
  );
  assertObjectKeys(
    driftAccounting,
    networkFlowDriftAccountingKeys,
    `${file}.drift_accounting`,
  );
  assertRequiredKeys(
    driftAccounting,
    networkFlowDriftAccountingKeys,
    `${file}.drift_accounting`,
  );
  const scratchManifestPath = requireRepoRelativePath(
    driftAccounting.scratch_manifest,
    `${file}.drift_accounting.scratch_manifest`,
  );
  const requiredCopyPaths = requireStringArray(
    driftAccounting.required_copy_paths,
    `${file}.drift_accounting.required_copy_paths`,
    { nonEmpty: true },
  );
  const requiredPublicTargets = requireStringArray(
    driftAccounting.required_public_targets,
    `${file}.drift_accounting.required_public_targets`,
    { nonEmpty: true },
  );

  const expectedFixtureIDs = new Set(
    expectedNetworkFlowIDs("NF-FIX-", expectedFixtureCount),
  );
  const expectedAcceptanceIDs = new Set(
    expectedNetworkFlowIDs("NF-AC-", expectedAcceptanceCount),
  );
  const matrix = readShapeFile(repoFile(repoRoot, matrixSource), matrixSource);
  const matrixRows = requireObjectArray(matrix.rows, `${matrixSource}.rows`, {
    nonEmpty: true,
  }).filter((row) => row.status === "active");
  const accountedAcceptanceIDs = new Set();
  for (const [index, row] of acceptanceRows.entries()) {
    const label = `${file}.acceptance_accounting.rows[${index}]`;
    const acceptanceID = requireString(row.acceptance_id, `${label}.acceptance_id`, {
      pattern: networkFlowAcceptanceIDPattern,
    });
    if (accountedAcceptanceIDs.has(acceptanceID)) {
      throw new Error(`${label}.acceptance_id duplicates ${acceptanceID}`);
    }
    accountedAcceptanceIDs.add(acceptanceID);
    requireStringArray(
      row.owner_requirements,
      `${label}.owner_requirements`,
      { nonEmpty: true },
    );
    const behaviorClass = requireEnum(
      row.behavior_class,
      `${label}.behavior_class`,
      new Set(["product_runtime", "normative_structure", "fixture_integrity", "evidence_policy"]),
    );
    const requiredEvidenceKinds = new Set(
      requireStringArray(row.required_evidence_kinds, `${label}.required_evidence_kinds`, {
        nonEmpty: true,
      }),
    );
    const selectors = requireObjectArray(row.exact_selectors, `${label}.exact_selectors`, {
      nonEmpty: true,
    });
    const actualKinds = new Set();
    for (const [selectorIndex, selector] of selectors.entries()) {
      const selectorLabel = `${label}.exact_selectors[${selectorIndex}]`;
      const kind = requireEnum(
        selector.kind,
        `${selectorLabel}.kind`,
        new Set(["unit", "store", "process", "browser"]),
      );
      actualKinds.add(kind);
      const selectorFile = requireRepoRelativePath(
        selector.file,
        `${selectorLabel}.file`,
      );
      const selectorPath = repoFile(repoRoot, selectorFile);
      if (!existsSync(selectorPath) || !lstatSync(selectorPath).isFile()) {
        throw new Error(`${selectorLabel}.file does not resolve: ${selectorFile}`);
      }
      const symbolOrTitle = selector.symbol ?? selector.title;
      const selectorSource = readFileSync(selectorPath, "utf8");
      if (!symbolOrTitle || !selectorSource.includes(symbolOrTitle)) {
        throw new Error(`${selectorLabel} does not resolve in ${selectorFile}`);
      }
      const catalogRows = matrixRows.filter((candidate) => {
        const catalogSelector = candidate.selector;
        if (!catalogSelector || typeof catalogSelector !== "object") {
          return false;
        }
        if (selector.title) {
          return (
            candidate.runner === "playwright" &&
            catalogSelector.file === selectorFile &&
            Array.isArray(catalogSelector.titles) &&
            catalogSelector.titles.includes(selector.title)
          );
        }
        return (
          candidate.runner === "go" &&
          catalogSelector.package === selector.package &&
          Array.isArray(catalogSelector.tests) &&
          catalogSelector.tests.includes(selector.symbol)
        );
      });
      if (catalogRows.length !== 1) {
        throw new Error(
          `${selectorLabel} must resolve to exactly one active ${acceptanceID} catalog row; found ${catalogRows.length}`,
        );
      }
    }
    assertExactIDSet(actualKinds, requiredEvidenceKinds, `${label}.required_evidence_kinds`);
    if (behaviorClass === "product_runtime" && actualKinds.size === 0) {
      throw new Error(`${label} product_runtime evidence must execute a runtime selector`);
    }
    const fixtureIDs = requireStringArray(
      row.supplemental_fixture_ids,
      `${label}.supplemental_fixture_ids`,
    );
    for (const fixtureID of fixtureIDs) {
      const manifestPath = `fixtures/network-flow/${fixtureID}/manifest.json`;
      if (!existsSync(repoFile(repoRoot, manifestPath))) {
        throw new Error(`${label} references missing supplemental fixture ${fixtureID}`);
      }
      const manifest = readShapeFile(repoFile(repoRoot, manifestPath), manifestPath);
      const fixtureAcceptanceIDs = new Set(manifest.acceptance_ids ?? []);
      if (!fixtureAcceptanceIDs.has(acceptanceID)) {
        throw new Error(`${manifestPath} does not claim supplemental coverage for ${acceptanceID}`);
      }
    }
  }
  assertExactIDSet(
    accountedAcceptanceIDs,
    expectedAcceptanceIDs,
    `${file}.acceptance_accounting.rows acceptance IDs`,
  );

  const registry = readShapeFile(repoFile(repoRoot, registryPath), registryPath);
  validateSchemaSync(contractFamilyRegistrySchemaID, registry);
  const family = requireObjectArray(registry.families, `${registryPath}.families`).find(
    (entry) => entry.family_id === contractRegistry.family_id,
  );
  if (!family) {
    throw new Error(`${registryPath} is missing network-flow family`);
  }
  requireExact(
    family.contract_root,
    contractRegistry.contract_root,
    `${registryPath}.families[network-flow].contract_root`,
  );
  const generatedOutputs = requireStringArray(
    family.generated_outputs,
    `${registryPath}.families[network-flow].generated_outputs`,
    { nonEmpty: true },
  );
  const generationStatus = requireEnum(
    family.generation_status,
    `${registryPath}.families[network-flow].generation_status`,
    new Set(["active", "planned"]),
  );
  const familyActivationIDs = requireStringArray(
    family.activation_dependency_ids,
    `${registryPath}.families[network-flow].activation_dependency_ids`,
  );
  const markerPresent = generatedOutputsContainMarker(
    repoRoot,
    generatedOutputs,
    generatedMarkers,
  );
  if (generationStatus === "planned") {
    assertExactIDSet(
      new Set(familyActivationIDs),
      new Set(plannedActivationIDs),
      `${registryPath}.families[network-flow].activation_dependency_ids`,
    );
    if (markerPresent) {
      throw new Error(
        `${registryPath} marks network-flow planned but generated outputs contain Network Flow markers`,
      );
    }
  } else {
    if (familyActivationIDs.length !== 0) {
      throw new Error(
        `${registryPath} marks network-flow active but activation dependencies remain`,
      );
    }
    if (!markerPresent) {
      throw new Error(
        `${registryPath} marks network-flow active but generated outputs lack Network Flow markers`,
      );
    }
  }

  const scratchManifest = readShapeFile(
    repoFile(repoRoot, scratchManifestPath),
    scratchManifestPath,
  );
  const copyPaths = new Set(
    requireStringArray(
      scratchManifest.copy_paths,
      `${scratchManifestPath}.copy_paths`,
      { nonEmpty: true },
    ),
  );
  for (const requiredPath of requiredCopyPaths) {
    if (!copyPaths.has(requiredPath)) {
      throw new Error(
        `${scratchManifestPath}.copy_paths must include ${requiredPath}`,
      );
    }
  }
  const generatedPaths = requireStringArray(
    scratchManifest.generated_paths,
    `${scratchManifestPath}.generated_paths`,
    { nonEmpty: true },
  );
  for (const generatedOutput of generatedOutputs) {
    if (!pathIsCoveredByAny(generatedOutput, generatedPaths)) {
      throw new Error(
        `${scratchManifestPath}.generated_paths does not cover ${generatedOutput}`,
      );
    }
  }

  const taskSurface = readShapeFile(
    repoFile(repoRoot, "tools/task_surface_manifest.json"),
    "tools/task_surface_manifest.json",
  );
  const targetNames = taskSurfaceTargetNames(taskSurface);
  for (const target of requiredPublicTargets) {
    if (!targetNames.has(target)) {
      throw new Error(`task surface is missing public target ${target}`);
    }
  }
}

function validateNetworkFlowContractIndexShape(file) {
  const contractIndex = readShapeFile(file, file);
  validateSchemaSync(networkFlowContractIndexSchemaID, contractIndex);
  assertObjectKeys(contractIndex, networkFlowContractIndexKeys, file);
  assertRequiredKeys(contractIndex, networkFlowContractIndexKeys, file);
  requireSchemaID(contractIndex, networkFlowContractIndexSchemaID, file);
  requireExact(contractIndex.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(contractIndex.contract_major, 2, `${file}.contract_major`);
  requireExact(contractIndex.document_version, "2.0.0", `${file}.document_version`);
  requireExact(contractIndex.family_id, "network-flow", `${file}.family_id`);
  requireExact(contractIndex.owner_id, "module.networkflow", `${file}.owner_id`);
  requireExactArray(
    requireStringArray(
      contractIndex.verification_ids,
      `${file}.verification_ids`,
      { nonEmpty: true },
    ),
    ["module.networkflow.verification.contract_accounting"],
    `${file}.verification_ids`,
  );

  const contractFiles = requireObject(contractIndex.contract_files, `${file}.contract_files`);
  assertObjectKeys(contractFiles, networkFlowContractFilesKeys, `${file}.contract_files`);
  assertRequiredKeys(contractFiles, networkFlowContractFilesKeys, `${file}.contract_files`);
  const routeFile = networkFlowContractRepoPath(contractFiles.routes, `${file}.contract_files.routes`);
  const schemaFile = networkFlowContractRepoPath(contractFiles.schemas, `${file}.contract_files.schemas`);
  const errorFile = networkFlowContractRepoPath(contractFiles.errors, `${file}.contract_files.errors`);
  const timezoneFile = networkFlowContractRepoPath(
    contractFiles.timezone_provenance,
    `${file}.contract_files.timezone_provenance`,
  );
  const keyRingsFile = networkFlowContractRepoPath(
    contractFiles.key_rings,
    `${file}.contract_files.key_rings`,
  );
  const mappingRegistryFile = networkFlowContractRepoPath(
    contractFiles.mapping_registry,
    `${file}.contract_files.mapping_registry`,
  );
  const presentationFile = networkFlowContractRepoPath(
    contractFiles.presentation,
    `${file}.contract_files.presentation`,
  );
  requireExact(
    contractFiles.timezone_provenance,
    "contracts/network-flow/timezone/tzdb-2026c.provenance.json",
    `${file}.contract_files.timezone_provenance`,
  );
  requireExact(
    contractFiles.key_rings,
    "contracts/network-flow/key-rings.v1.schema.json",
    `${file}.contract_files.key_rings`,
  );
  requireExact(
    contractFiles.mapping_registry,
    "contracts/network-flow/mapping-registry.v1.json",
    `${file}.contract_files.mapping_registry`,
  );
  requireExact(
    contractFiles.presentation,
    "contracts/network-flow/presentation.v1.json",
    `${file}.contract_files.presentation`,
  );
  for (const referencedPath of [routeFile, schemaFile, errorFile, timezoneFile, keyRingsFile, mappingRegistryFile, presentationFile]) {
    if (!existsSync(repoFile(repoRoot, referencedPath))) {
      throw new Error(`${file} references missing Network Flow contract file ${referencedPath}`);
    }
  }

  const publicSchemaIDs = new Set(
    requireStringArray(contractIndex.public_schema_ids, `${file}.public_schema_ids`, {
      nonEmpty: true,
    }),
  );
  const closurePolicy = requireObject(contractIndex.closure_policy, `${file}.closure_policy`);
  assertObjectKeys(closurePolicy, networkFlowClosurePolicyKeys, `${file}.closure_policy`);
  assertRequiredKeys(closurePolicy, networkFlowClosurePolicyKeys, `${file}.closure_policy`);
  for (const [key, expected] of [
    ["objects_closed_by_default", true],
    ["dynamic_maps_must_name_key_pattern", true],
    ["raw_source_values_forbidden_outside_diagnostics", true],
  ]) {
    requireExact(closurePolicy[key], expected, `${file}.closure_policy.${key}`);
  }
  assertExactIDSet(
    new Set(
      requireStringArray(
        closurePolicy.generated_outputs_blocked_until,
        `${file}.closure_policy.generated_outputs_blocked_until`,
        { nonEmpty: true },
      ),
    ),
    new Set(["NFA-GEN-003", "NFA-GEN-004"]),
    `${file}.closure_policy.generated_outputs_blocked_until`,
  );

  validateNetworkFlowRouteContractsShape(repoFile(repoRoot, routeFile), publicSchemaIDs);
  const errorCodes = validateNetworkFlowErrorContractsShape(repoFile(repoRoot, errorFile));
  validateNetworkFlowPublicSchemaBundle(repoFile(repoRoot, schemaFile), publicSchemaIDs);
  validateNetworkFlowTimezoneRulesetProvenanceShape(repoFile(repoRoot, timezoneFile));
  validateNetworkFlowMappingRegistryShape(repoFile(repoRoot, mappingRegistryFile));
  validateNetworkFlowPresentationShape(repoFile(repoRoot, presentationFile));

  const routes = readShapeFile(repoFile(repoRoot, routeFile), routeFile);
  for (const route of routes.routes) {
    for (const code of route.primary_errors) {
      if (!errorCodes.has(code)) {
        throw new Error(`${routeFile}.${route.route_id}.primary_errors includes unknown ${code}`);
      }
    }
  }
}

function networkFlowContractRepoPath(value, label) {
  const relativePath = requireRepoRelativePath(value, label, { extension: ".json" });
  if (!relativePath.startsWith("contracts/network-flow/")) {
    throw new Error(`${label} must be under contracts/network-flow`);
  }
  return relativePath;
}

function validateNetworkFlowPresentationShape(file) {
  const presentation = readShapeFile(file, file);
  const rootKeys = new Set(["$schema", "schema_id", "profile_id", "document_version", "grid_schemas"]);
  assertObjectKeys(presentation, rootKeys, file);
  assertRequiredKeys(presentation, rootKeys, file);
  requireSchemaID(presentation, "cartulary.network_flow.presentation.v1", file);
  requireExact(presentation.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(presentation.document_version, "2.0.0", `${file}.document_version`);
  const grids = requireObjectArray(presentation.grid_schemas, `${file}.grid_schemas`);
  requireExactArrayLength(grids, 3, `${file}.grid_schemas`);
  const expectedIDs = [
    "network_flow.accepted_rows.v1",
    "network_flow.rejected_rows.v1",
    "network_flow.graph_contributors.v1",
  ];
  for (const [index, grid] of grids.entries()) {
    const label = `${file}.grid_schemas[${index + 1}]`;
    const contributor = grid.grid_schema_id === "network_flow.graph_contributors.v1";
    const keys = new Set([
      "grid_schema_id",
      "resource_kind",
      "grouping",
      "server_order_only",
      contributor ? "columns_from_grid_schema_id" : "columns",
    ]);
    assertObjectKeys(grid, keys, label);
    assertRequiredKeys(grid, keys, label);
    requireExact(grid.grid_schema_id, expectedIDs[index], `${label}.grid_schema_id`);
    requireString(grid.resource_kind, `${label}.resource_kind`);
    requireString(grid.grouping, `${label}.grouping`);
    requireBoolean(grid.server_order_only, `${label}.server_order_only`);
    if (contributor) {
      requireExact(
        grid.columns_from_grid_schema_id,
        "network_flow.accepted_rows.v1",
        `${label}.columns_from_grid_schema_id`,
      );
      continue;
    }
    const columns = requireObjectArray(grid.columns, `${label}.columns`);
    const seenFields = new Set();
    const seenOrders = new Set();
    for (const [columnIndex, column] of columns.entries()) {
      const columnLabel = `${label}.columns[${columnIndex + 1}]`;
      const columnKeys = new Set([
        "field_key", "label_key", "value_kind", "renderer_kind",
        "filter_operators", "sortable", "copyable", "link_contexts",
        "default_visible", "default_order", "default_width_px",
        "minimum_width_px", "inspector_only",
      ]);
      assertObjectKeys(column, columnKeys, columnLabel);
      assertRequiredKeys(column, columnKeys, columnLabel);
      const fieldKey = requireString(column.field_key, `${columnLabel}.field_key`);
      if (seenFields.has(fieldKey)) {
        throw new Error(`${columnLabel}.field_key duplicates ${fieldKey}`);
      }
      seenFields.add(fieldKey);
      requireString(column.label_key, `${columnLabel}.label_key`);
      requireString(column.value_kind, `${columnLabel}.value_kind`);
      requireString(column.renderer_kind, `${columnLabel}.renderer_kind`);
      for (const [arrayKey, values] of [["filter_operators", column.filter_operators], ["link_contexts", column.link_contexts]]) {
        for (const [valueIndex, value] of requireArray(values, `${columnLabel}.${arrayKey}`).entries()) {
          requireString(value, `${columnLabel}.${arrayKey}[${valueIndex + 1}]`);
        }
      }
      for (const booleanKey of ["sortable", "copyable", "default_visible", "inspector_only"]) {
        requireBoolean(column[booleanKey], `${columnLabel}.${booleanKey}`);
      }
      const order = requireInteger(column.default_order, `${columnLabel}.default_order`);
      if (order < 0 || seenOrders.has(order)) {
        throw new Error(`${columnLabel}.default_order must be unique and non-negative`);
      }
      seenOrders.add(order);
      const width = requireInteger(column.default_width_px, `${columnLabel}.default_width_px`);
      const minimum = requireInteger(column.minimum_width_px, `${columnLabel}.minimum_width_px`);
      if (width < 0 || minimum < 0 || column.inspector_only === false && width < minimum) {
        throw new Error(`${columnLabel} has invalid width bounds`);
      }
    }
  }
}

function validateNetworkFlowMappingRegistryShape(file) {
  const registry = readShapeFile(file, file);
  const rootKeys = new Set([
    "$schema", "schema_id", "profile_id", "document_version", "target_kind",
    "target_table_schema_id", "system_derivations", "source_profiles",
  ]);
  assertObjectKeys(registry, rootKeys, file);
  assertRequiredKeys(registry, rootKeys, file);
  requireSchemaID(registry, "cartulary.network_flow.mapping_registry.v1", file);
  requireExact(registry.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(registry.document_version, "2.0.0", `${file}.document_version`);
  requireExact(registry.target_kind, "network_flow_table", `${file}.target_kind`);
  requireExact(registry.target_table_schema_id, "cartulary.network_flow_table.v1", `${file}.target_table_schema_id`);
  const derivations = requireObjectArray(registry.system_derivations, `${file}.system_derivations`);
  requireExactArrayLength(derivations, 1, `${file}.system_derivations`);
  const profiles = requireObjectArray(registry.source_profiles, `${file}.source_profiles`);
  requireExactArrayLength(profiles, 1, `${file}.source_profiles`);
  const profile = profiles[0];
  const profileKeys = new Set([
    "source_profile_id", "display_name", "conformance_status", "parser_profile_id",
    "default_unknown_column_policy", "supported_unknown_column_policies",
    "default_timestamp_profile", "supported_timestamp_modes", "fields",
  ]);
  assertObjectKeys(profile, profileKeys, `${file}.source_profiles[1]`);
  assertRequiredKeys(profile, profileKeys, `${file}.source_profiles[1]`);
  requireExact(profile.source_profile_id, "cisco_sna_netflow_csv_v1", `${file}.source_profiles[1].source_profile_id`);
  const fields = requireObjectArray(profile.fields, `${file}.source_profiles[1].fields`);
  requireExactArrayLength(fields, 15, `${file}.source_profiles[1].fields`);
  const fieldKeys = new Set();
  for (const [index, field] of fields.entries()) {
    const label = `${file}.source_profiles[1].fields[${index + 1}]`;
    const keys = new Set(["field_key", "requirement", "transform_id", "empty_value_policy", "aliases"]);
    assertObjectKeys(field, keys, label);
    assertRequiredKeys(field, keys, label);
    const fieldKey = requireString(field.field_key, `${label}.field_key`);
    if (fieldKeys.has(fieldKey)) {
      throw new Error(`${label}.field_key duplicates ${fieldKey}`);
    }
    fieldKeys.add(fieldKey);
    requireEnum(
      field.requirement,
      `${label}.requirement`,
      new Set(["required", "optional_map_when_present", "not_supported", "system_derived"]),
    );
    if (field.transform_id !== null) requireString(field.transform_id, `${label}.transform_id`);
    if (field.empty_value_policy !== null) requireString(field.empty_value_policy, `${label}.empty_value_policy`);
    for (const [aliasIndex, alias] of requireArray(field.aliases, `${label}.aliases`).entries()) {
      requireString(alias, `${label}.aliases[${aliasIndex + 1}]`);
    }
  }
}

function validateNetworkFlowRouteContractsShape(file, publicSchemaIDs) {
  const routeContracts = readShapeFile(file, file);
  assertObjectKeys(routeContracts, networkFlowRouteContractKeys, file);
  assertRequiredKeys(routeContracts, networkFlowRouteContractKeys, file);
  requireSchemaID(routeContracts, "cartulary.network_flow_route_contracts.v1", file);
  requireExact(routeContracts.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(routeContracts.contract_major, 2, `${file}.contract_major`);
  requireExact(
    routeContracts.route_root,
    "/api/v1/incidents/{incident_id}/network-flow",
    `${file}.route_root`,
  );

  const integration = requireObject(routeContracts.import_integration, `${file}.import_integration`);
  assertObjectKeys(integration, networkFlowImportIntegrationKeys, `${file}.import_integration`);
  assertRequiredKeys(integration, networkFlowImportIntegrationKeys, `${file}.import_integration`);
  requireExact(integration.target_kind, "network_flow_table", `${file}.import_integration.target_kind`);
  requireExact(
    integration.target_table_schema_id,
    "cartulary.network_flow_table.v1",
    `${file}.import_integration.target_table_schema_id`,
  );
  requireExact(
    integration.resource_ref_kind,
    "network_flow_table",
    `${file}.import_integration.resource_ref_kind`,
  );
  requireExact(
    integration.default_source_profile_id,
    "cisco_sna_netflow_csv_v1",
    `${file}.import_integration.default_source_profile_id`,
  );
  requireExact(
    integration.default_parser_profile_id,
    "rfc4180_headered_csv_v1",
    `${file}.import_integration.default_parser_profile_id`,
  );
  requireExact(
    integration.default_unknown_column_policy,
    "preserve_unmapped_raw",
    `${file}.import_integration.default_unknown_column_policy`,
  );
  requireExact(
    integration.owner_facade,
    "network_flow_import_facade_v1",
    `${file}.import_integration.owner_facade`,
  );
  requireExact(
    integration.mapping_preview_route,
    "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview",
    `${file}.import_integration.mapping_preview_route`,
  );
  const facadeOperations = validateObjectArray(
    integration.owner_facade_operations,
    `${file}.import_integration.owner_facade_operations`,
    {
      nonEmpty: true,
      keys: networkFlowFacadeOperationKeys,
      requiredKeys: networkFlowFacadeOperationKeys,
    },
  );
  const expectedFacadeOperations = [
    [
      "preview",
      "cartulary.network_flow.import_preview_request.v1",
      "cartulary.network_flow.import_preview_result.v1",
    ],
    [
      "apply",
      "cartulary.network_flow.import_apply_request.v1",
      "cartulary.network_flow.import_unit_result.v1",
    ],
  ];
  requireExactArrayLength(
    facadeOperations,
    expectedFacadeOperations.length,
    `${file}.import_integration.owner_facade_operations`,
  );
  for (const [index, expected] of expectedFacadeOperations.entries()) {
    const operation = facadeOperations[index];
    const label = `${file}.import_integration.owner_facade_operations[${index + 1}]`;
    requireExact(operation.operation, expected[0], `${label}.operation`);
    requirePublicSchemaID(operation.request_schema_id, publicSchemaIDs, `${label}.request_schema_id`);
    requireExact(operation.request_schema_id, expected[1], `${label}.request_schema_id`);
    requirePublicSchemaID(operation.success_schema_id, publicSchemaIDs, `${label}.success_schema_id`);
    requireExact(operation.success_schema_id, expected[2], `${label}.success_schema_id`);
  }

  const expectedRoutes = [
    {
      route_id: "nf.source_profiles.list",
      method: "GET",
      path: "/api/v1/incidents/{incident_id}/network-flow/source-profiles",
      auth_context: "viewer",
      request_schema_id: null,
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.source_profile_list.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: ["network_flow_invalid_request"],
      audit_event: null,
    },
    {
      route_id: "nf.tables.list",
      method: "GET",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables",
      auth_context: "viewer",
      request_schema_id: null,
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.table_list.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: ["network_flow_invalid_request"],
      audit_event: null,
    },
    {
      route_id: "nf.tables.get",
      method: "GET",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}",
      auth_context: "viewer",
      request_schema_id: null,
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.table_get.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: ["network_flow_table_not_found", "network_flow_table_not_active"],
      audit_event: null,
    },
    {
      route_id: "nf.tables.patch",
      method: "PATCH",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}",
      auth_context: "editor",
      request_schema_id: "cartulary.network_flow.table_rename_request.v1",
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.table_mutation_result.v1",
      success_http_statuses: [200],
      idempotency: "client_txn_id_required",
      primary_errors: ["network_flow_table_version_conflict", "network_flow_invalid_display_name"],
      audit_event: "network_flow_table_renamed",
    },
    {
      route_id: "nf.tables.delete",
      method: "DELETE",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}",
      auth_context: "reviewer",
      request_schema_id: "cartulary.network_flow.table_soft_delete_request.v1",
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.table_mutation_result.v1",
      success_http_statuses: [200],
      idempotency: "client_txn_id_required",
      primary_errors: ["network_flow_table_version_conflict", "network_flow_table_not_active"],
      audit_event: "network_flow_table_soft_deleted",
    },
    {
      route_id: "nf.tables.query",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}/query",
      auth_context: "viewer",
      request_schema_id: "cartulary.network_flow.table_query_request.v1",
      continuation_schema_id: "cartulary.network_flow.table_query_continuation.v1",
      success_schema_id: "cartulary.network_flow.table_query_result.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: [
        "network_flow_invalid_filter",
        "network_flow_invalid_sort",
        "network_flow_cursor_invalid",
      ],
      audit_event: null,
    },
    {
      route_id: "nf.rejected_rows.query",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/tables/{network_flow_table_id}/rejected-rows/query",
      auth_context: "viewer",
      request_schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
      continuation_schema_id: "cartulary.network_flow.rejected_rows_query_continuation.v1",
      success_schema_id: "cartulary.network_flow.rejected_rows_query_result.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: ["network_flow_invalid_filter", "network_flow_cursor_invalid"],
      audit_event: null,
    },
    {
      route_id: "nf.rows.query",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/rows/query",
      auth_context: "viewer",
      request_schema_id: "cartulary.network_flow.rows_query_request.v1",
      continuation_schema_id: "cartulary.network_flow.rows_query_continuation.v1",
      success_schema_id: "cartulary.network_flow.rows_query_result.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: [
        "network_flow_invalid_table_scope",
        "network_flow_invalid_filter",
        "network_flow_invalid_sort",
        "network_flow_cursor_invalid",
      ],
      audit_event: null,
    },
    {
      route_id: "nf.graphs.query",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/graphs/query",
      auth_context: "viewer",
      request_schema_id: "cartulary.network_flow.graph_query_request.v1",
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.graph_query_result.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: [
        "network_flow_invalid_table_scope",
        "network_flow_invalid_time_range",
        "network_flow_invalid_limit_override",
        "network_flow_graph_limit_exceeded",
        "network_flow_counter_sum_limit_exceeded",
        "network_flow_graph_projection_failed",
      ],
      audit_event: "network_flow_graph_query_executed",
    },
    {
      route_id: "nf.graphs.contributors.query",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/graphs/contributors/query",
      auth_context: "viewer",
      request_schema_id: "cartulary.network_flow.graph_contributor_query_request.v1",
      continuation_schema_id: "cartulary.network_flow.graph_contributor_query_continuation.v1",
      success_schema_id: "cartulary.network_flow.graph_contributor_query_result.v1",
      success_http_statuses: [200],
      idempotency: "read_route",
      primary_errors: [
        "network_flow_graph_query_stale",
        "network_flow_invalid_table_scope",
        "network_flow_cursor_invalid",
        "network_flow_invalid_limit",
      ],
      audit_event: null,
    },
    {
      route_id: "nf.indicator_links.create",
      method: "POST",
      path: "/api/v1/incidents/{incident_id}/network-flow/indicator-links",
      auth_context: "editor",
      request_schema_id: "cartulary.network_flow.indicator_link_request.v1",
      continuation_schema_id: null,
      success_schema_id: "cartulary.network_flow.indicator_link_result.v1",
      success_http_statuses: [200, 201],
      idempotency: "client_txn_id_required",
      primary_errors: [
        "network_flow_indicator_link_ambiguous",
        "network_flow_invalid_indicator_selector",
        "network_flow_invalid_indicator_target",
        "network_flow_indicator_link_forbidden",
      ],
      audit_event: "network_flow_indicator_binding_created_or_reused",
    },
  ];
  const routes = validateObjectArray(routeContracts.routes, `${file}.routes`, {
    nonEmpty: true,
    keys: networkFlowRouteEntryKeys,
    requiredKeys: networkFlowRouteEntryKeys,
  });
  requireExactArrayLength(routes, expectedRoutes.length, `${file}.routes`);
  for (const [index, expected] of expectedRoutes.entries()) {
    const route = routes[index];
    const label = `${file}.routes[${index + 1}]`;
    for (const key of ["route_id", "method", "path", "auth_context", "idempotency", "audit_event"]) {
      requireExact(route[key], expected[key], `${label}.${key}`);
    }
    requireNullOrPublicSchemaID(route.request_schema_id, publicSchemaIDs, `${label}.request_schema_id`);
    requireExact(route.request_schema_id, expected.request_schema_id, `${label}.request_schema_id`);
    requireNullOrPublicSchemaID(
      route.continuation_schema_id,
      publicSchemaIDs,
      `${label}.continuation_schema_id`,
    );
    requireExact(
      route.continuation_schema_id,
      expected.continuation_schema_id,
      `${label}.continuation_schema_id`,
    );
    requirePublicSchemaID(route.success_schema_id, publicSchemaIDs, `${label}.success_schema_id`);
    requireExact(route.success_schema_id, expected.success_schema_id, `${label}.success_schema_id`);
    requireExactArray(
      requireArray(route.success_http_statuses, `${label}.success_http_statuses`, { nonEmpty: true }),
      expected.success_http_statuses,
      `${label}.success_http_statuses`,
    );
    assertExactIDSet(
      new Set(requireStringArray(route.primary_errors, `${label}.primary_errors`, { nonEmpty: true })),
      new Set(expected.primary_errors),
      `${label}.primary_errors`,
    );
  }
}

function requirePublicSchemaID(value, publicSchemaIDs, label) {
  const schemaID = requireString(value, label);
  if (!publicSchemaIDs.has(schemaID)) {
    throw new Error(`${label} must reference a Network Flow public schema ID`);
  }
  return schemaID;
}

function requireNullOrPublicSchemaID(value, publicSchemaIDs, label) {
  if (value === null) {
    return null;
  }
  return requirePublicSchemaID(value, publicSchemaIDs, label);
}

function requireExactArrayLength(value, expected, label) {
  if (value.length !== expected) {
    throw new Error(`${label} must contain exactly ${expected} entries`);
  }
}

function requireExactArray(value, expected, label) {
  if (value.length !== expected.length) {
    throw new Error(`${label} must contain exactly ${expected.length} entries`);
  }
  for (const [index, expectedValue] of expected.entries()) {
    requireExact(value[index], expectedValue, `${label}[${index + 1}]`);
  }
}

function validateNetworkFlowErrorContractsShape(file) {
  const errorContracts = readShapeFile(file, file);
  assertObjectKeys(errorContracts, networkFlowErrorContractKeys, file);
  assertRequiredKeys(errorContracts, networkFlowErrorContractKeys, file);
  requireSchemaID(errorContracts, "cartulary.network_flow_error_contracts.v1", file);
  requireExact(errorContracts.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(errorContracts.contract_major, 2, `${file}.contract_major`);
  assertExactIDSet(
    new Set(requireStringArray(errorContracts.retry_actions, `${file}.retry_actions`, { nonEmpty: true })),
    new Set([
      "correct_request",
      "refresh_resource",
      "restart_query",
      "reduce_scope_or_limits",
      "retry_with_backoff",
      "do_not_retry",
    ]),
    `${file}.retry_actions`,
  );

  const expectedErrors = new Map(
    [
      ["network_flow_invalid_request", "route", 400, "correct_request"],
      ["network_flow_unsupported_source_profile", "route", 400, "correct_request"],
      ["network_flow_invalid_utf8", "route", 400, "correct_request"],
      ["network_flow_csv_empty_file", "route", 400, "correct_request"],
      ["network_flow_invalid_header", "route", 400, "correct_request"],
      ["network_flow_no_data_rows", "route", 400, "correct_request"],
      ["network_flow_csv_malformed_quote", "route", 400, "correct_request"],
      ["network_flow_source_changed", "route", 409, "refresh_resource"],
      ["network_flow_csv_field_count_mismatch", "row_diagnostic", null, "correct_request"],
      ["network_flow_mapping_required", "route", 400, "correct_request"],
      ["network_flow_mapping_conflict", "route", 400, "correct_request"],
      ["network_flow_invalid_timestamp", "row_diagnostic", null, "correct_request"],
      ["network_flow_end_before_start", "row_diagnostic", null, "correct_request"],
      ["network_flow_invalid_ip", "row_diagnostic", null, "correct_request"],
      ["network_flow_invalid_port", "row_diagnostic", null, "correct_request"],
      ["network_flow_invalid_protocol", "row_diagnostic", null, "correct_request"],
      ["network_flow_invalid_counter", "row_diagnostic", null, "correct_request"],
      ["network_flow_all_rows_rejected", "route", 400, "correct_request"],
      ["network_flow_table_limit_exceeded", "route", 409, "reduce_scope_or_limits"],
      ["network_flow_resource_limit_exceeded", "route_or_row_diagnostic", 413, "reduce_scope_or_limits"],
      ["network_flow_table_name_exhausted", "route", 409, "correct_request"],
      ["network_flow_table_not_found", "route", 404, "refresh_resource"],
      ["network_flow_table_not_active", "route", 409, "refresh_resource"],
      ["network_flow_table_version_conflict", "route", 409, "refresh_resource"],
      ["network_flow_invalid_display_name", "route", 400, "correct_request"],
      ["network_flow_invalid_table_scope", "route", 400, "correct_request"],
      ["network_flow_invalid_filter", "route", 400, "correct_request"],
      ["network_flow_invalid_sort", "route", 400, "correct_request"],
      ["network_flow_invalid_limit", "route", 400, "correct_request"],
      ["network_flow_cursor_invalid", "route", 400, "restart_query"],
      ["network_flow_invalid_time_range", "route", 400, "correct_request"],
      ["network_flow_invalid_limit_override", "route", 400, "correct_request"],
      ["network_flow_graph_limit_exceeded", "route", 413, "reduce_scope_or_limits"],
      ["network_flow_counter_sum_limit_exceeded", "route", 413, "reduce_scope_or_limits"],
      ["network_flow_graph_projection_failed", "route", 502, "do_not_retry"],
      ["network_flow_graph_query_stale", "route", 409, "refresh_resource"],
      ["network_flow_indicator_link_ambiguous", "route", 400, "correct_request"],
      ["network_flow_invalid_indicator_selector", "route", 400, "correct_request"],
      ["network_flow_invalid_indicator_target", "route", 400, "correct_request"],
      ["network_flow_indicator_link_forbidden", "route", 403, "do_not_retry"],
      ["network_flow_external_enrichment_forbidden", "route", 400, "do_not_retry"],
      ["network_flow_id_generation_failed", "route", 500, "do_not_retry"],
    ].map(([code, scope, httpStatus, retryAction]) => [
      code,
      { scope, http_status: httpStatus, retry_action: retryAction },
    ]),
  );
  const errors = validateObjectArray(errorContracts.errors, `${file}.errors`, {
    nonEmpty: true,
    keys: networkFlowErrorEntryKeys,
    requiredKeys: networkFlowErrorEntryKeys,
  });
  requireExactArrayLength(errors, expectedErrors.size, `${file}.errors`);
  const errorCodes = new Set();
  for (const [index, error] of errors.entries()) {
    const label = `${file}.errors[${index + 1}]`;
    const code = requireString(error.code, `${label}.code`);
    if (errorCodes.has(code)) {
      throw new Error(`${file}.errors contains duplicate ${code}`);
    }
    errorCodes.add(code);
    const expected = expectedErrors.get(code);
    if (!expected) {
      throw new Error(`${label}.code is not a Network Flow Table 21-A error code`);
    }
    requireExact(error.scope, expected.scope, `${label}.scope`);
    requireExact(error.http_status, expected.http_status, `${label}.http_status`);
    requireExact(error.retry_action, expected.retry_action, `${label}.retry_action`);
  }
  assertExactIDSet(errorCodes, new Set(expectedErrors.keys()), `${file}.errors.code`);

  const reasonRegistries = validateObjectArray(
    errorContracts.reason_registries,
    `${file}.reason_registries`,
    {
      nonEmpty: true,
      keys: networkFlowReasonRegistryKeys,
      requiredKeys: networkFlowReasonRegistryKeys,
    },
  );
  const reasonFamilies = new Set();
  for (const [index, registry] of reasonRegistries.entries()) {
    const label = `${file}.reason_registries[${index + 1}]`;
    const errorCode = requireString(registry.error_code, `${label}.error_code`);
    if (reasonFamilies.has(errorCode)) {
      throw new Error(`${file}.reason_registries contains duplicate ${errorCode}`);
    }
    reasonFamilies.add(errorCode);
    requireStringArray(registry.reason_codes, `${label}.reason_codes`, { nonEmpty: true });
  }
  assertExactIDSet(
    new Set([...reasonFamilies].filter((entry) =>
      [
        "network_flow_invalid_request",
        "network_flow_source_changed",
        "network_flow_mapping_conflict",
        "network_flow_invalid_filter",
        "network_flow_cursor_invalid",
        "network_flow_graph_projection_failed",
        "indicator-link errors",
        "network_flow_table_limit_exceeded,network_flow_resource_limit_exceeded",
      ].includes(entry),
    )),
    new Set([
      "network_flow_invalid_request",
      "network_flow_source_changed",
      "network_flow_mapping_conflict",
      "network_flow_invalid_filter",
      "network_flow_cursor_invalid",
      "network_flow_graph_projection_failed",
      "indicator-link errors",
      "network_flow_table_limit_exceeded,network_flow_resource_limit_exceeded",
    ]),
    `${file}.reason_registries required families`,
  );
  return errorCodes;
}

function validateNetworkFlowPublicSchemaBundle(file, publicSchemaIDs) {
  const bundle = readShapeFile(file, file);
  const bundleKeys = new Set(["$schema", "$id", "schema_id", "profile_id", "contract_major", "$defs"]);
  assertObjectKeys(bundle, bundleKeys, file);
  assertRequiredKeys(bundle, bundleKeys, file);
  requireExact(bundle.$schema, "https://json-schema.org/draft/2020-12/schema", `${file}.$schema`);
  requireExact(bundle.$id, "cartulary.network_flow_public_schemas.v1", `${file}.$id`);
  requireSchemaID(bundle, "cartulary.network_flow_public_schemas.v1", file);
  requireExact(bundle.profile_id, "network_flow_activity", `${file}.profile_id`);
  requireExact(bundle.contract_major, 2, `${file}.contract_major`);
  const defs = requireObject(bundle.$defs, `${file}.$defs`);
  const actualSchemaIDs = new Set();
  for (const [defName, def] of Object.entries(defs)) {
    const label = `${file}.$defs.${defName}`;
    requireObject(def, label);
    if (Object.hasOwn(def, "x_schema_id")) {
      const schemaID = requireString(def.x_schema_id, `${label}.x_schema_id`);
      if (!publicSchemaIDs.has(schemaID)) {
        throw new Error(`${label}.x_schema_id is not listed by the Network Flow contract index`);
      }
      if (actualSchemaIDs.has(schemaID)) {
        throw new Error(`${file} duplicates public schema ID ${schemaID}`);
      }
      actualSchemaIDs.add(schemaID);
      validateNetworkFlowPublicSchemaIDConstants(def, label, schemaID);
    }
    validateNetworkFlowSchemaClosure(def, label);
  }
  assertExactIDSet(actualSchemaIDs, publicSchemaIDs, `${file} public schema IDs`);
}

function validateNetworkFlowPublicSchemaIDConstants(node, label, schemaID) {
  if (Array.isArray(node)) {
    for (const [index, entry] of node.entries()) {
      validateNetworkFlowPublicSchemaIDConstants(entry, `${label}[${index + 1}]`, schemaID);
    }
    return;
  }
  if (!node || typeof node !== "object") {
    return;
  }
  const declaredSchemaID = node.properties?.schema_id?.const;
  if (declaredSchemaID !== undefined && declaredSchemaID !== schemaID) {
    throw new Error(`${label}.properties.schema_id.const must match ${schemaID}`);
  }
  for (const [key, value] of Object.entries(node)) {
    if (["x_schema_id", "$ref", "const", "enum"].includes(key)) {
      continue;
    }
    validateNetworkFlowPublicSchemaIDConstants(value, `${label}.${key}`, schemaID);
  }
}

function validateNetworkFlowSchemaClosure(node, label) {
  if (Array.isArray(node)) {
    for (const [index, entry] of node.entries()) {
      validateNetworkFlowSchemaClosure(entry, `${label}[${index + 1}]`);
    }
    return;
  }
  if (!node || typeof node !== "object") {
    return;
  }
  if (node.additionalProperties === true) {
    throw new Error(`${label}.additionalProperties must not be true`);
  }
  const objectLike =
    node.type === "object" ||
    Object.hasOwn(node, "properties") ||
    Object.hasOwn(node, "required") ||
    Object.hasOwn(node, "propertyNames");
  if (objectLike) {
    const closed =
      node.additionalProperties === false ||
      node.unevaluatedProperties === false ||
      (Object.hasOwn(node, "propertyNames") &&
        Object.hasOwn(node, "additionalProperties") &&
        node.additionalProperties !== true);
    if (!closed) {
      throw new Error(`${label} must be closed or declare an explicit dynamic-map key pattern`);
    }
  }
  for (const [key, value] of Object.entries(node)) {
    if (["x_schema_id", "$ref", "const", "enum"].includes(key)) {
      continue;
    }
    validateNetworkFlowSchemaClosure(value, `${label}.${key}`);
  }
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
  return `${artifact.role}\u0000${artifact.path_kind}\u0000${artifact.format ?? ""}\u0000${artifact.path}`;
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
    requireEnum(
      artifact.path_kind,
      `${artifactLabel}.path_kind`,
      toolRunArtifactPathKinds,
    );
    const keys = artifact.path_kind === "file"
      ? toolRunFileArtifactKeys
      : toolRunDirectoryArtifactKeys;
    assertObjectKeys(artifact, keys, artifactLabel);
    assertRequiredKeys(artifact, keys, artifactLabel);
    requireString(artifact.role, `${artifactLabel}.role`);
    if (artifact.path_kind === "file") {
      requireEnum(artifact.format, `${artifactLabel}.format`, toolRunArtifactFormats);
    }
    requireRepoRelativePath(artifact.path, `${artifactLabel}.path`);
  }
  requireSorted(artifacts, label, artifactStableKey, "role, path kind, format, and path");
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
    summary.step_accounting,
    `${file}.step_accounting`,
    toolRunStepAccountingKeys,
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

function validateTestSupportInventoryShape(file, root = repoRoot) {
  const inventory = readShapeFile(file, file);
  assertObjectKeys(inventory, testSupportInventoryKeys, file);
  assertRequiredKeys(inventory, testSupportInventoryKeys, file);
  validateSchemaSync(testSupportInventorySchemaID, inventory);

  const goRootPaths = [];
  const goRoots = validateObjectArray(
    inventory.go_support_roots,
    `${file}.go_support_roots`,
    {
      nonEmpty: true,
      keys: goSupportRootKeys,
      requiredKeys: goSupportRootKeys,
    },
    (entry, label) => {
      const relative = requireRepoRelativePath(entry.path, `${label}.path`);
      requireString(entry.owner, `${label}.owner`);
      requireEnum(entry.posture, `${label}.posture`, goSupportPostures);
      const runtimeScan = requireEnum(entry.runtime_scan, `${label}.runtime_scan`, scanTreatments);
      const supportScan = requireEnum(
        entry.support_scan,
        `${label}.support_scan`,
        supportScanTreatments,
      );
      requireBoolean(entry.service_starting, `${label}.service_starting`);
      requireString(entry.rationale, `${label}.rationale`);
      requireDirectory(root, relative, `${label}.path`);
      if (phaseShapedSupportPathPattern.test(relative)) {
        throw new Error(`${label}.path must not use a phase-shaped helper root`);
      }
      if (runtimeScan === "excluded" && supportScan !== "included") {
        throw new Error(`${label} is excluded from runtime scans but not included in support scans`);
      }
      validateServiceStartingClassification(root, relative, entry.service_starting, label);
      goRootPaths.push(relative);
    },
  );
  requireSorted(goRoots, `${file}.go_support_roots`, (entry) => entry.path, "path");
  assertUnique(goRootPaths, `${file}.go_support_roots.path`);

  const registeredGoRootSet = new Set(goRootPaths);
  for (const discovered of discoverGoSupportRoots(root)) {
    if (!registeredGoRootSet.has(discovered)) {
      throw new Error(`${file}.go_support_roots missing discovered support root ${discovered}`);
    }
  }

  const sharedDataRootPaths = [];
  const sharedDataRoots = validateObjectArray(
    inventory.shared_data_roots,
    `${file}.shared_data_roots`,
    {
      nonEmpty: true,
      keys: sharedDataRootKeys,
      requiredKeys: sharedDataRootKeys,
    },
    (entry, label) => {
      const relative = requireRepoRelativePath(entry.path, `${label}.path`);
      requireString(entry.owner, `${label}.owner`);
      requireEnum(entry.posture, `${label}.posture`, sharedDataPostures);
      requireEnum(entry.data_kind, `${label}.data_kind`, sharedDataKinds);
      for (const role of requireStringArray(entry.file_roles, `${label}.file_roles`, {
        nonEmpty: true,
      })) {
        requireEnum(role, `${label}.file_roles`, sharedDataFileRoles);
      }
      requireEnum(
        entry.owner_semantic_data_policy,
        `${label}.owner_semantic_data_policy`,
        sharedDataPolicies,
      );
      requireEnum(
        entry.retained_path_policy,
        `${label}.retained_path_policy`,
        retainedPathPolicies,
      );
      requireString(entry.rationale, `${label}.rationale`);
      requireDirectory(root, relative, `${label}.path`);
      if (!sharedDataFacadeRoots.some((facadeRoot) => pathContains(facadeRoot, relative))) {
        throw new Error(`${label}.path must be under a shared fixture or golden facade root`);
      }
      sharedDataRootPaths.push(relative);
    },
  );
  requireSorted(sharedDataRoots, `${file}.shared_data_roots`, (entry) => entry.path, "path");
  assertUnique(sharedDataRootPaths, `${file}.shared_data_roots.path`);
  assertNoOverlappingRoots(sharedDataRootPaths, `${file}.shared_data_roots.path`);
  validateSharedDataCoverage(root, sharedDataRootPaths, file);
  validateAdoptedOtelRoot(sharedDataRoots, file);
}

function requireDirectory(root, relative, label) {
  const absolute = repoFile(root, relative);
  if (!existsSync(absolute)) {
    throw new Error(`${label} does not exist: ${relative}`);
  }
  const stat = lstatSync(absolute);
  if (!stat.isDirectory()) {
    throw new Error(`${label} must be a directory: ${relative}`);
  }
}

function pathContains(rootPath, candidatePath) {
  return candidatePath === rootPath || candidatePath.startsWith(`${rootPath}/`);
}

function repoRelativePath(root, absolute) {
  return path.relative(root, absolute).split(path.sep).join("/");
}

function discoverGoSupportRoots(root) {
  const roots = new Set(
    discoveredGoSupportRoots.filter((relative) => existsSync(repoFile(root, relative))),
  );
  const internalRoot = repoFile(root, "internal");
  if (existsSync(internalRoot)) {
    collectTestSupportDirs(root, internalRoot, roots);
  }
  return [...roots].sort((left, right) => left.localeCompare(right));
}

function collectTestSupportDirs(root, directory, roots) {
  for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) =>
    left.name.localeCompare(right.name),
  )) {
    const absolute = path.join(directory, entry.name);
    if (entry.isSymbolicLink()) {
      throw new Error(`${absolute} must not be a symlink`);
    }
    if (!entry.isDirectory()) {
      continue;
    }
    const relative = repoRelativePath(root, absolute);
    if (entry.name === "testsupport") {
      roots.add(relative);
      continue;
    }
    collectTestSupportDirs(root, absolute, roots);
  }
}

function validateServiceStartingClassification(root, relative, serviceStarting, label) {
  const files = collectFilesUnder(root, relative, { includeGoOnly: true });
  const startsServices = files.some((file) =>
    /(?:pgtest|s3test)\.Start(?:Owned|Shared|With|\()|apptestsupport\.StartRuntime\(|httptestx\.StartServer\(|func\s+Start(?:Runtime|Server|Store)\s*\(/.test(
      readFileSync(repoFile(root, file), "utf8"),
    ),
  );
  if (startsServices && !serviceStarting) {
    throw new Error(`${label}.service_starting must be true because ${relative} starts services`);
  }
}

function collectFilesUnder(root, relative, { includeGoOnly = false, skipGo = false } = {}) {
  const absolute = repoFile(root, relative);
  const files = [];
  collectFilesUnderDir(root, absolute, files, { includeGoOnly, skipGo });
  return files;
}

function collectFilesUnderDir(root, directory, files, options) {
  for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) =>
    left.name.localeCompare(right.name),
  )) {
    const absolute = path.join(directory, entry.name);
    if (entry.isSymbolicLink()) {
      throw new Error(`${absolute} must not be a symlink`);
    }
    if (entry.isDirectory()) {
      collectFilesUnderDir(root, absolute, files, options);
      continue;
    }
    if (!entry.isFile()) {
      continue;
    }
    const relative = repoRelativePath(root, absolute);
    if (options.includeGoOnly && !relative.endsWith(".go")) {
      continue;
    }
    if (options.skipGo && relative.endsWith(".go")) {
      continue;
    }
    files.push(relative);
  }
}

function assertNoOverlappingRoots(paths, label) {
  for (let index = 0; index < paths.length; index += 1) {
    for (let next = index + 1; next < paths.length; next += 1) {
      if (pathContains(paths[index], paths[next]) || pathContains(paths[next], paths[index])) {
        throw new Error(`${label} contains overlapping roots ${paths[index]} and ${paths[next]}`);
      }
    }
  }
}

function validateSharedDataCoverage(root, sharedDataRootPaths, file) {
  const files = sharedDataFacadeRoots.flatMap((facadeRoot) => {
    if (!existsSync(repoFile(root, facadeRoot))) {
      return [];
    }
    return collectFilesUnder(root, facadeRoot, { skipGo: true });
  });
  for (const relative of files) {
    const matches = sharedDataRootPaths.filter((rootPath) => pathContains(rootPath, relative));
    if (matches.length !== 1) {
      throw new Error(
        `${file}.shared_data_roots must classify ${relative} exactly once; matched ${matches.length}`,
      );
    }
  }
}

function validateAdoptedOtelRoot(sharedDataRoots, file) {
  const otel = sharedDataRoots.find((entry) => entry.path === "internal/testutil/golden/otel");
  if (!otel) {
    throw new Error(`${file}.shared_data_roots must retain internal/testutil/golden/otel`);
  }
  if (
    otel.data_kind !== "otel_evidence" ||
    otel.owner_semantic_data_policy !== "adopted_external_evidence" ||
    otel.retained_path_policy !== "stable"
  ) {
    throw new Error(`${file}.shared_data_roots OTel entry must preserve adopted stable evidence posture`);
  }
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

  const sourceOwnerModule = requireString(entry.source_owner_module, `${label}.source_owner_module`, {
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
    entry.source_record_types,
    `${label}.source_record_types`,
    { nonEmpty: true, pattern: snakeIDPattern },
  );

  const sourceAuthorityModules = requireStringArray(
    entry.source_authority_modules,
    `${label}.source_authority_modules`,
    { nonEmpty: true, pattern: snakeIDPattern },
  );
  for (const module of sourceAuthorityModules) {
    if (!projectionProviderSourceAuthorityModules.has(module)) {
      throw new Error(`${label}.source_authority_modules contains unknown owner ${module}`);
    }
  }
  if (!sourceAuthorityModules.includes(sourceOwnerModule)) {
    throw new Error(
      `${label}.source_authority_modules must include source_owner_module ${sourceOwnerModule}`,
    );
  }

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

function graphProjectionIDRange(prefix, count) {
  return Array.from({ length: count }, (_, index) =>
    `${prefix}-${String(index + 1).padStart(3, "0")}`,
  );
}

function validateGraphProjectionFixtureCorpusShape(file, expectedFixtureIDs) {
  const corpus = readShapeFile(file, file);
  assertObjectKeys(corpus, graphProjectionFixtureCorpusKeys, file);
  assertRequiredKeys(corpus, graphProjectionFixtureCorpusKeys, file);
  requireSchemaID(corpus, graphProjectionFixtureCorpusSchemaID, file);
  requireRepoRelativePath(corpus.spec_path, `${file}.spec_path`, {
    extension: ".md",
  });

  const corpusFixtureIDs = [];
  validateObjectArray(
    corpus.fixtures,
    `${file}.fixtures`,
    {
      nonEmpty: true,
      keys: graphProjectionCorpusFixtureKeys,
      requiredKeys: graphProjectionCorpusFixtureKeys,
    },
    (entry, label) => {
      const fixtureID = requireString(entry.fixture_id, `${label}.fixture_id`, {
        pattern: /^GP-FIX-\d{3}$/,
      });
      corpusFixtureIDs.push(fixtureID);
      requireString(entry.coverage, `${label}.coverage`);
      requireString(entry.input_kind, `${label}.input_kind`);
    },
  );
  assertUnique(corpusFixtureIDs, `${file}.fixtures.fixture_id`);
  requireSorted(
    corpusFixtureIDs,
    `${file}.fixtures.fixture_id`,
    (entry) => entry,
    "GP-FIX identifier",
  );
  if (corpusFixtureIDs.join("\n") !== expectedFixtureIDs.join("\n")) {
    throw new Error(
      `${file}.fixtures must list ${expectedFixtureIDs[0]} through ${
        expectedFixtureIDs[expectedFixtureIDs.length - 1]
      }`,
    );
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
      const selectors = requireStringArray(entry.evidence_selectors, `${label}.evidence_selectors`, {
        nonEmpty: true,
      });
      for (const [selectorIndex, selector] of selectors.entries()) {
        validateGraphProjectionEvidenceSelector(
          selector,
          `${label}.evidence_selectors[${selectorIndex + 1}]`,
        );
      }
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
  const expectedAcceptanceIDs = graphProjectionIDRange(
    "GP-AC",
    graphProjectionAcceptanceCount,
  );
  if (acceptanceIDs.join("\n") !== expectedAcceptanceIDs.join("\n")) {
    throw new Error(
      `${file}.acceptance_criteria must list ${expectedAcceptanceIDs[0]} through ${
        expectedAcceptanceIDs[expectedAcceptanceIDs.length - 1]
      }`,
    );
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
  const expectedFixtureIDs = graphProjectionIDRange(
    "GP-FIX",
    graphProjectionFixtureCount,
  );
  if (fixtureIDs.join("\n") !== expectedFixtureIDs.join("\n")) {
    throw new Error(
      `${file}.fixture_registry must list ${expectedFixtureIDs[0]} through ${
        expectedFixtureIDs[expectedFixtureIDs.length - 1]
      }`,
    );
  }
  for (const fixtureID of seenFixtureIDs) {
    if (!fixtureIDs.includes(fixtureID)) {
      throw new Error(`${file}.acceptance_criteria references unknown fixture ${fixtureID}`);
    }
  }
  validateGraphProjectionFixtureCorpusShape(
    repoFile(repoRoot, "contracts/graph-projection/fixtures/corpus.v1.json"),
    expectedFixtureIDs,
  );
}

function validateGraphProjectionEvidenceSelector(selector, label) {
  const separator = selector.indexOf("::");
  const selectedPath = separator === -1 ? selector : selector.slice(0, separator);
  const selectedSymbol = separator === -1 ? "" : selector.slice(separator + 2);
  const absolutePath = path.resolve(repoRoot, selectedPath);
  if (!absolutePath.startsWith(`${repoRoot}${path.sep}`) || !existsSync(absolutePath)) {
    throw new Error(`${label} selects missing path ${selectedPath}`);
  }
  if (selectedSymbol === "") {
    return;
  }
  if (lstatSync(absolutePath).isDirectory()) {
    const testSources = readdirSync(absolutePath, { withFileTypes: true })
      .filter((entry) => entry.isFile() && entry.name.endsWith("_test.go"))
      .map((entry) => readFileSync(path.join(absolutePath, entry.name), "utf8"))
      .join("\n");
    if (!testSources.includes(`func ${selectedSymbol}(`)) {
      throw new Error(`${label} selects missing Go test symbol ${selectedSymbol}`);
    }
    return;
  }
  const contents = readFileSync(absolutePath, "utf8");
  if (!contents.includes(selectedSymbol)) {
    throw new Error(`${label} selects missing anchor ${selectedSymbol}`);
  }
}

function validateGraphProjectionFixtureManifests(root) {
  const fixtureRoot = repoFile(root, "contracts/graph-projection/fixtures");
  const matrix = readShapeFile(
    repoFile(root, "contracts/graph-projection/conformance_matrix.v1.json"),
  );
  const matrixFixtureAcceptance = new Map();
  for (const criterion of matrix.acceptance_criteria) {
    for (const fixtureID of criterion.fixture_ids) {
      if (!matrixFixtureAcceptance.has(fixtureID)) {
        matrixFixtureAcceptance.set(fixtureID, new Set());
      }
      matrixFixtureAcceptance.get(fixtureID).add(criterion.id);
    }
  }
  for (const entry of readdirSync(fixtureRoot, { withFileTypes: true })) {
    if (!entry.isDirectory() || !/^GP-FIX-\d{3}$/.test(entry.name)) {
      continue;
    }
    const fixtureID = entry.name;
    const fixtureDir = path.join(fixtureRoot, fixtureID);
    const manifestPath = path.join(fixtureDir, "fixture.json");
    if (!existsSync(manifestPath)) {
      throw new Error(`${fixtureDir} is missing fixture.json`);
    }
    const manifest = readShapeFile(manifestPath, manifestPath);
    validateSchemaSync(graphProjectionFixtureManifestSchemaID, manifest);
    requireExact(manifest.fixture_id, fixtureID, `${manifestPath}.fixture_id`);
    const declaredAcceptance = new Set(manifest.acceptance_ids);
    const matrixAcceptance = matrixFixtureAcceptance.get(fixtureID) ?? new Set();
    for (const acceptanceID of declaredAcceptance) {
      if (!matrixAcceptance.has(acceptanceID)) {
        throw new Error(
          `${manifestPath}.acceptance_ids contains ${acceptanceID}, but the matrix does not attach ${fixtureID} to it`,
        );
      }
    }
    for (const acceptanceID of matrixAcceptance) {
      if (!declaredAcceptance.has(acceptanceID)) {
        throw new Error(
          `${manifestPath}.acceptance_ids omits matrix attachment ${acceptanceID}`,
        );
      }
    }
    validateGraphProjectionEvidenceSelector(
      `internal/modules/graphprojection::${manifest.test_symbol}`,
      `${manifestPath}.test_symbol`,
    );
    const artifacts = requireArray(manifest.artifacts, `${manifestPath}.artifacts`, {
      nonEmpty: true,
    });
    const paths = new Set();
    for (const [index, artifact] of artifacts.entries()) {
      const label = `${manifestPath}.artifacts[${index + 1}]`;
      const logicalPath = requireManifestRelativePath(artifact.path, `${label}.path`);
      if (paths.has(logicalPath)) {
        throw new Error(`${manifestPath} duplicates artifact ${logicalPath}`);
      }
      paths.add(logicalPath);
      const artifactPath = path.resolve(fixtureDir, logicalPath);
      if (!artifactPath.startsWith(`${fixtureDir}${path.sep}`) || !existsSync(artifactPath)) {
        throw new Error(`${label}.path is missing or escapes the fixture directory`);
      }
      if (lstatSync(artifactPath).isSymbolicLink()) {
        throw new Error(`${label}.path must not be a symlink`);
      }
      const digest = sha256Hex(readFileSync(artifactPath));
      requireExact(artifact.sha256, digest, `${label}.sha256`);
    }
  }
}

function sha256Hex(buffer) {
  return createHash("sha256").update(buffer).digest("hex");
}

function requireSHA256(value, label) {
  return requireString(value, label, { pattern: sha256Pattern });
}

function requireManifestRelativePath(value, label) {
  return requireString(value, label, {
    pattern: manifestRelativePathPattern,
  });
}

function requireNetworkFlowFixtureFiles(entries, label, manifestDir, keys) {
  const logicalPaths = [];
  for (const [index, entry] of entries.entries()) {
    const entryLabel = `${label}[${index + 1}]`;
    assertObjectKeys(entry, keys, entryLabel);
    assertRequiredKeys(entry, keys, entryLabel);
    const logicalPath = requireManifestRelativePath(
      entry.logical_path,
      `${entryLabel}.logical_path`,
    );
    logicalPaths.push(logicalPath);
    requireString(entry.media_type, `${entryLabel}.media_type`);
    const expectedSize = requireInteger(entry.size_bytes, `${entryLabel}.size_bytes`, {
      min: 0,
    });
    const expectedDigest = requireSHA256(entry.sha256, `${entryLabel}.sha256`);
    const resolved = path.resolve(manifestDir, logicalPath);
    if (!resolved.startsWith(`${manifestDir}${path.sep}`)) {
      throw new Error(`${entryLabel}.logical_path escapes fixture directory`);
    }
    const stat = lstatSync(resolved);
    if (stat.isSymbolicLink()) {
      throw new Error(`${entryLabel}.logical_path must not be a symlink`);
    }
    if (!stat.isFile()) {
      throw new Error(`${entryLabel}.logical_path must reference a regular file`);
    }
    if (stat.size !== expectedSize) {
      throw new Error(
        `${entryLabel}.size_bytes ${expectedSize} does not match actual ${stat.size}`,
      );
    }
    const actualDigest = sha256Hex(readFileSync(resolved));
    if (actualDigest !== expectedDigest) {
      throw new Error(`${entryLabel}.sha256 does not match file bytes`);
    }
  }
  assertUnique(logicalPaths, `${label}.logical_path`);
  requireSorted(
    logicalPaths,
    `${label}.logical_path`,
    (entry) => entry,
    "manifest-relative path",
  );
  return logicalPaths;
}

function networkFlowFixtureBundleHash(entries) {
  const hash = createHash("sha256");
  for (const entry of entries) {
    hash.update(entry.logical_path, "utf8");
    hash.update(Buffer.from([0]));
    hash.update(entry.sha256, "utf8");
    hash.update(Buffer.from([0]));
    hash.update(String(entry.size_bytes), "utf8");
    hash.update("\n", "utf8");
  }
  return hash.digest("hex");
}

function collectNetworkFlowFixtureFiles(fixtureDir, currentDir = fixtureDir) {
  const relativeFiles = [];
  for (const entry of readdirSync(currentDir, { withFileTypes: true }).sort(
    (left, right) => left.name.localeCompare(right.name),
  )) {
    const absolutePath = path.join(currentDir, entry.name);
    if (entry.isSymbolicLink()) {
      throw new Error(`${absolutePath} must not be a symlink`);
    }
    if (entry.isDirectory()) {
      relativeFiles.push(
        ...collectNetworkFlowFixtureFiles(fixtureDir, absolutePath),
      );
      continue;
    }
    if (!entry.isFile()) {
      throw new Error(`${absolutePath} must be a regular file or directory`);
    }
    relativeFiles.push(
      path.relative(fixtureDir, absolutePath).split(path.sep).join("/"),
    );
  }
  return relativeFiles;
}

function validateNetworkFlowFixtureManifestShape(file) {
  const manifest = readShapeFile(file, file);
  validateSchemaSync(networkFlowFixtureManifestSchemaID, manifest);
  assertObjectKeys(manifest, networkFlowFixtureManifestKeys, file);
  assertRequiredKeys(
    manifest,
    new Set([...networkFlowFixtureManifestKeys].filter((key) => key !== "extensions")),
    file,
  );
  requireSchemaID(manifest, networkFlowFixtureManifestSchemaID, file);
  const fixtureID = requireString(manifest.fixture_id, `${file}.fixture_id`, {
    pattern: networkFlowFixtureIDPattern,
  });
  if (path.basename(path.dirname(file)) !== fixtureID) {
    throw new Error(`${file}.fixture_id must match its fixture directory name`);
  }
  if (path.basename(file) !== "manifest.json") {
    throw new Error(`${file} must be named manifest.json`);
  }
  requirePositiveInteger(manifest.manifest_version, `${file}.manifest_version`);
  if (manifest.profile_id !== "network_flow_activity") {
    throw new Error(`${file}.profile_id must be network_flow_activity`);
  }

  const freeze = requireObject(manifest.freeze, `${file}.freeze`);
  assertObjectKeys(freeze, networkFlowFixtureFreezeKeys, `${file}.freeze`);
  assertRequiredKeys(freeze, networkFlowFixtureFreezeKeys, `${file}.freeze`);
  requireEnum(freeze.status, `${file}.freeze.status`, new Set(["draft", "frozen"]));
  requirePositiveInteger(freeze.revision, `${file}.freeze.revision`);
  if (freeze.change_policy !== "new_fixture_revision_required") {
    throw new Error(
      `${file}.freeze.change_policy must be new_fixture_revision_required`,
    );
  }

  requireExactArray(
    requireStringArray(manifest.verification_ids, `${file}.verification_ids`, {
      nonEmpty: true,
    }),
    ["module.networkflow.verification.contract_accounting"],
    `${file}.verification_ids`,
  );

  const manifestDir = path.dirname(file);
  const sourceFiles = requireObjectArray(manifest.source_files, `${file}.source_files`, {
    nonEmpty: true,
  });
  const sourceLogicalPaths = requireNetworkFlowFixtureFiles(
    sourceFiles,
    `${file}.source_files`,
    manifestDir,
    networkFlowFixtureSourceFileKeys,
  );
  const expectedArtifacts = requireObjectArray(
    manifest.expected_artifacts,
    `${file}.expected_artifacts`,
    { nonEmpty: true },
  );
  const expectedLogicalPaths = requireNetworkFlowFixtureFiles(
    expectedArtifacts,
    `${file}.expected_artifacts`,
    manifestDir,
    networkFlowFixtureExpectedArtifactKeys,
  );
  const transcriptFiles = requireObjectArray(
    manifest.transcript_files,
    `${file}.transcript_files`,
    { nonEmpty: true },
  );
  const transcriptLogicalPaths = requireNetworkFlowFixtureFiles(
    transcriptFiles,
    `${file}.transcript_files`,
    manifestDir,
    networkFlowFixtureTranscriptFileKeys,
  );
  const listedPaths = new Set([
    ...sourceLogicalPaths,
    ...expectedLogicalPaths,
    ...transcriptLogicalPaths,
    "manifest.json",
  ]);
  for (const actualPath of collectNetworkFlowFixtureFiles(manifestDir)) {
    if (!listedPaths.has(actualPath)) {
      throw new Error(`${file} contains unlisted fixture file ${actualPath}`);
    }
  }

  const acceptanceIDs = requireStringArray(
    manifest.acceptance_ids,
    `${file}.acceptance_ids`,
    { nonEmpty: true },
  );
  for (const id of acceptanceIDs) {
    if (!networkFlowAcceptanceIDPattern.test(id)) {
      throw new Error(`${file}.acceptance_ids contains invalid ${id}`);
    }
  }
  assertUnique(acceptanceIDs, `${file}.acceptance_ids`);
  requireSorted(acceptanceIDs, `${file}.acceptance_ids`, (entry) => entry, "NF-AC ID");

  const executionSelectors = requireStringArray(
    manifest.execution_selectors,
    `${file}.execution_selectors`,
    { nonEmpty: true },
  );
  assertUnique(executionSelectors, `${file}.execution_selectors`);
  requireSorted(
    executionSelectors,
    `${file}.execution_selectors`,
    (entry) => entry,
    "execution selector",
  );

  const sourceBundleDigest = requireSHA256(
    manifest.source_bundle_sha256,
    `${file}.source_bundle_sha256`,
  );
  const expectedBundleDigest = requireSHA256(
    manifest.expected_bundle_sha256,
    `${file}.expected_bundle_sha256`,
  );
  const actualSourceDigest = networkFlowFixtureBundleHash(sourceFiles);
  if (actualSourceDigest !== sourceBundleDigest) {
    throw new Error(`${file}.source_bundle_sha256 does not match source files`);
  }
  const actualExpectedDigest = networkFlowFixtureBundleHash([
    ...expectedArtifacts,
    ...transcriptFiles,
  ]);
  if (actualExpectedDigest !== expectedBundleDigest) {
    throw new Error(
      `${file}.expected_bundle_sha256 does not match expected artifacts and transcripts`,
    );
  }
}

function validateNetworkFlowFixtureManifests(root) {
  const fixtureRoot = repoFile(root, "fixtures/network-flow");
  if (!existsSync(fixtureRoot)) {
    return;
  }
  for (const entry of readdirSync(fixtureRoot, { withFileTypes: true }).sort(
    (left, right) => left.name.localeCompare(right.name),
  )) {
    if (!entry.isDirectory()) {
      continue;
    }
    const manifestPath = path.join(fixtureRoot, entry.name, "manifest.json");
    if (existsSync(manifestPath)) {
      validateNetworkFlowFixtureManifestShape(manifestPath);
    }
  }
}

function requireExact(value, expected, label) {
  if (value !== expected) {
    throw new Error(`${label} must be ${expected}`);
  }
}

function validateNetworkFlowTimezoneRulesetProvenanceShape(file) {
  const provenance = readShapeFile(file, file);
  validateSchemaSync(networkFlowTimezoneRulesetProvenanceSchemaID, provenance);
  assertObjectKeys(provenance, networkFlowTimezoneProvenanceKeys, file);
  assertRequiredKeys(provenance, networkFlowTimezoneProvenanceKeys, file);
  requireSchemaID(provenance, networkFlowTimezoneRulesetProvenanceSchemaID, file);
  requireExact(provenance.ruleset_id, "tzdb-2026c", `${file}.ruleset_id`);
  requireExact(
    provenance.profile_id,
    "network_flow_activity",
    `${file}.profile_id`,
  );
  requireExact(provenance.iana_version, "2026c", `${file}.iana_version`);

  const release = requireObject(provenance.release, `${file}.release`);
  assertObjectKeys(release, networkFlowTimezoneReleaseKeys, `${file}.release`);
  assertRequiredKeys(release, networkFlowTimezoneReleaseKeys, `${file}.release`);
  requireRFC3339Timestamp(release.released_at, `${file}.release.released_at`);
  requireExact(
    release.released_at,
    "2026-07-08T17:23:58Z",
    `${file}.release.released_at`,
  );
  requireExact(release.release_date, "2026-07-08", `${file}.release.release_date`);
  requireExact(
    release.release_index_url,
    "https://www.iana.org/time-zones",
    `${file}.release.release_index_url`,
  );
  requireExact(
    release.release_archive_index_url,
    "https://ftp.iana.org/tz/releases/",
    `${file}.release.release_archive_index_url`,
  );

  const sourceArchive = requireObject(
    provenance.source_archive,
    `${file}.source_archive`,
  );
  assertObjectKeys(
    sourceArchive,
    networkFlowTimezoneArchiveKeys,
    `${file}.source_archive`,
  );
  assertRequiredKeys(
    sourceArchive,
    networkFlowTimezoneArchiveKeys,
    `${file}.source_archive`,
  );
  requireExact(
    sourceArchive.distribution,
    "data_only",
    `${file}.source_archive.distribution`,
  );
  requireExact(
    sourceArchive.file_name,
    "tzdata2026c.tar.gz",
    `${file}.source_archive.file_name`,
  );
  requireExact(
    sourceArchive.url,
    "https://data.iana.org/time-zones/releases/tzdata2026c.tar.gz",
    `${file}.source_archive.url`,
  );
  requireExact(
    sourceArchive.media_type,
    "application/gzip",
    `${file}.source_archive.media_type`,
  );
  const sourceSize = requireInteger(
    sourceArchive.size_bytes,
    `${file}.source_archive.size_bytes`,
    { min: 1 },
  );
  if (sourceSize !== 475694) {
    throw new Error(`${file}.source_archive.size_bytes must be 475694`);
  }
  requireExact(
    requireSHA256(sourceArchive.sha256, `${file}.source_archive.sha256`),
    "e4a178a4477f3d0ea77cc31828ff72aa38feff8d61aa13e7e99e142e9d902be4",
    `${file}.source_archive.sha256`,
  );
  if (sourceArchive.url.includes("latest")) {
    throw new Error(`${file}.source_archive.url must not use a latest alias`);
  }

  const signature = requireObject(
    provenance.detached_signature,
    `${file}.detached_signature`,
  );
  assertObjectKeys(
    signature,
    networkFlowTimezoneSignatureKeys,
    `${file}.detached_signature`,
  );
  assertRequiredKeys(
    signature,
    networkFlowTimezoneSignatureKeys,
    `${file}.detached_signature`,
  );
  requireExact(
    signature.file_name,
    "tzdata2026c.tar.gz.asc",
    `${file}.detached_signature.file_name`,
  );
  requireExact(
    signature.url,
    "https://data.iana.org/time-zones/releases/tzdata2026c.tar.gz.asc",
    `${file}.detached_signature.url`,
  );
  requireExact(
    signature.media_type,
    "application/pgp-signature",
    `${file}.detached_signature.media_type`,
  );
  const signatureSize = requireInteger(
    signature.size_bytes,
    `${file}.detached_signature.size_bytes`,
    { min: 1 },
  );
  if (signatureSize !== 833) {
    throw new Error(`${file}.detached_signature.size_bytes must be 833`);
  }
  requireExact(
    requireSHA256(signature.sha256, `${file}.detached_signature.sha256`),
    "26cd02e034eed682aa911d224bca3247ff15914df317e3bb0b1a01dc557b46fe",
    `${file}.detached_signature.sha256`,
  );
  requireExact(
    signature.openpgp_fingerprint,
    "7E3792A9D8ACF7D633BC1588ED97E90E62AA7E34",
    `${file}.detached_signature.openpgp_fingerprint`,
  );
  requireExact(
    signature.openpgp_key_id,
    "ED97E90E62AA7E34",
    `${file}.detached_signature.openpgp_key_id`,
  );
  requireRFC3339Timestamp(
    signature.signature_created_at,
    `${file}.detached_signature.signature_created_at`,
  );
  requireExact(
    signature.signature_created_at,
    "2026-07-08T17:38:53Z",
    `${file}.detached_signature.signature_created_at`,
  );

  const license = requireObject(provenance.license, `${file}.license`);
  assertObjectKeys(license, networkFlowTimezoneLicenseKeys, `${file}.license`);
  assertRequiredKeys(license, networkFlowTimezoneLicenseKeys, `${file}.license`);
  requireExact(license.source_path, "LICENSE", `${file}.license.source_path`);
  requireExact(
    license.summary,
    "public_domain_except_optional_bsd_3_clause_code_files",
    `${file}.license.summary`,
  );
  requireExact(
    requireSHA256(license.sha256, `${file}.license.sha256`),
    "0613408568889f5739e5ae252b722a2659c02002839ad970a63dc5e9174b27cf",
    `${file}.license.sha256`,
  );
  if (license.data_only_distribution_requires_bsd_3_clause_exception !== false) {
    throw new Error(
      `${file}.license.data_only_distribution_requires_bsd_3_clause_exception must be false`,
    );
  }

  const embeddedPathKeys = [];
  validateObjectArray(
    provenance.embedded_file_hashes,
    `${file}.embedded_file_hashes`,
    {
      nonEmpty: true,
      keys: networkFlowTimezoneEmbeddedFileKeys,
      requiredKeys: networkFlowTimezoneEmbeddedFileKeys,
    },
    (entry, label) => {
      const logicalPath = requireEnum(
        entry.path,
        `${label}.path`,
        new Set(["LICENSE", "NEWS", "version"]),
      );
      embeddedPathKeys.push(logicalPath);
      requireInteger(entry.size_bytes, `${label}.size_bytes`, { min: 1 });
      requireSHA256(entry.sha256, `${label}.sha256`);
    },
  );
  requireSorted(
    embeddedPathKeys,
    `${file}.embedded_file_hashes.path`,
    (entry) => entry,
    "embedded file path",
  );
  const expectedEmbedded = new Map([
    [
      "LICENSE",
      {
        size_bytes: 252,
        sha256:
          "0613408568889f5739e5ae252b722a2659c02002839ad970a63dc5e9174b27cf",
      },
    ],
    [
      "NEWS",
      {
        size_bytes: 254018,
        sha256:
          "09bdfd57206fe221a3d71b15160b0ac0805209c757c258902a96b228961428c6",
      },
    ],
    [
      "version",
      {
        size_bytes: 6,
        sha256:
          "b8b066b540bc2870e6f1f3cd76f1b0e6c3629b2e3a12f14ba9e47085a1abb781",
      },
    ],
  ]);
  for (const entry of provenance.embedded_file_hashes) {
    const expected = expectedEmbedded.get(entry.path);
    if (!expected) {
      throw new Error(`${file}.embedded_file_hashes contains unexpected path`);
    }
    if (entry.size_bytes !== expected.size_bytes || entry.sha256 !== expected.sha256) {
      throw new Error(`${file}.embedded_file_hashes.${entry.path} does not match pinned bytes`);
    }
  }

  requireExactArray(
    requireStringArray(
      provenance.verification_ids,
      `${file}.verification_ids`,
      { nonEmpty: true },
    ),
    ["module.networkflow.verification.contract_accounting"],
    `${file}.verification_ids`,
  );

  const policy = requireObject(
    provenance.conformance_policy,
    `${file}.conformance_policy`,
  );
  assertObjectKeys(
    policy,
    networkFlowTimezoneConformancePolicyKeys,
    `${file}.conformance_policy`,
  );
  assertRequiredKeys(
    policy,
    networkFlowTimezoneConformancePolicyKeys,
    `${file}.conformance_policy`,
  );
  if (
    policy.host_timezone_database_authoritative !== false ||
    policy.host_locale_authoritative !== false ||
    policy.latest_url_allowed !== false ||
    policy.verification_required_before_use !== true
  ) {
    throw new Error(`${file}.conformance_policy has invalid authority booleans`);
  }
  requireExact(
    policy.allowed_internal_ruleset_substitution,
    "later_ruleset_only_when_all_tzdb_2026c_fixture_transitions_are_byte_identical",
    `${file}.conformance_policy.allowed_internal_ruleset_substitution`,
  );
}

function validateFallowReachabilityOwnerShape(file) {
  validateSchemaSync(
    fallowReachabilityOwnerSchemaID,
    readShapeFile(file, file),
  );
}

function validateKind(kind, file, root = repoRoot) {
  switch (kind) {
    case "execution-topology":
      validateExecutionTopologyShape(file);
      return;
    case "task-surface":
      validateTaskSurfaceShape(file);
      return;
    case "task-surface-owner":
      validateTaskSurfaceOwnerShape(file);
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
    case "contract-family-registry":
      validateContractFamilyRegistryShape(file);
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
    case "fallow-reachability-owner":
      validateFallowReachabilityOwnerShape(file);
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
    case "test-support-inventory":
      validateTestSupportInventoryShape(file, root);
      return;
    case "projection-provider-manifest":
      validateProjectionProviderManifestShape(file);
      return;
    case "graph-projection-conformance-matrix":
      validateGraphProjectionConformanceMatrixShape(file);
      return;
    case "network-flow-fixture-manifest":
      validateNetworkFlowFixtureManifestShape(file);
      return;
    case "network-flow-activity-accounting":
      validateNetworkFlowActivityAccountingShape(file);
      return;
    case "network-flow-contract-index":
      validateNetworkFlowContractIndexShape(file);
      return;
    case "network-flow-timezone-provenance":
      validateNetworkFlowTimezoneRulesetProvenanceShape(file);
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
  validateHarnessHelperOwnership(root);
  validateVerificationContracts(root);
  const testCatalog = validateTestCatalog(root);
  validateTestCatalogImportBoundary(root);
  scanExecutableDocumentationReads(root);
  const visualFixtureRegistry = readShapeFile(
    repoFile(root, "tools/frontend_visual_fixture_registry.json"),
  );
  validateSchemaSync(
    frontendVisualFixtureRegistrySchemaID,
    visualFixtureRegistry,
  );
  for (const fixture of visualFixtureRegistry.fixtures) {
    for (const rowID of fixture.catalog_row_ids) {
      const row = testCatalog.rowByID.get(rowID);
      if (!row || row.runner !== "playwright" || row.evidence_class !== "visual") {
        throw new Error(
          `visual fixture ${fixture.fixture_id} references non-visual catalog row ${rowID}`,
        );
      }
    }
  }

  validateExecutionTopologyShape(
    repoFile(root, "tools/execution_topology_manifest.json"),
  );
  validateTaskSurfaceOwnerShape(repoFile(root, "tools/task_surface_owner.json"));
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
  validateFallowReachabilityOwnerShape(
    repoFile(root, "tools/fallow/reachability_owner.json"),
  );
  validateContractFamilyRegistryShape(repoFile(root, "contracts/index.json"));
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
  validateSchemaSync(
    "cartulary.harness_public_target_duration_baselines.v2",
    readShapeFile(repoFile(root, "tools/harness_public_target_duration_baselines.json")),
  );
  validateBootstrapAdminShape(
    repoFile(root, "configs/dev/bootstrap-admin.json"),
  );
  validateTestSupportInventoryShape(
    repoFile(root, "tools/test_support_inventory.json"),
    root,
  );
  validateProjectionProviderManifestShape(
    repoFile(root, "contracts/projection-providers/index.json"),
  );
  validateGraphProjectionConformanceMatrixShape(
    repoFile(root, "contracts/graph-projection/conformance_matrix.v1.json"),
  );
  validateGraphProjectionFixtureManifests(root);
  validateNetworkFlowActivityAccountingShape(
    repoFile(root, "tools/network_flow_activity_accounting.json"),
  );
  validateNetworkFlowContractIndexShape(
    repoFile(root, "contracts/network-flow/index.json"),
  );
  validateNetworkFlowFixtureManifests(root);
  validateNetworkFlowTimezoneRulesetProvenanceShape(
    repoFile(root, "contracts/network-flow/timezone/tzdb-2026c.provenance.json"),
  );
  validateMigrationHistory(root);
  validateSchemaObjectOwnership(root);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.kind) {
    validateKind(options.kind, options.file, options.root);
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
