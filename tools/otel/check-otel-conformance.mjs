#!/usr/bin/env node

import { createHash } from "node:crypto";
import { mkdirSync, readdirSync, readFileSync, statSync, writeFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { enforcePrivateProcessUmask } from "../harness/runtime/private-process.mjs";
import {
  expectedBrowserRuntimeProbe,
  otelGeneratorSourceRef,
  validateOtelGeneratorReference,
  validateOtelImportBoundaryContractShape,
} from "./generate-otel-contracts.mjs";

// This process writes only conformance evidence below its result subtree; it
// never generates repository source. Raw captures, normalized captures,
// comparison evidence, and the summary therefore inherit owner-only modes.
enforcePrivateProcessUmask();

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "..");
const snapshotPath = "contracts/otel/otel_source_snapshot.v1.json";
const generatedConstantsManifestPath = "contracts/otel/generated_constants_manifest.json";
const importBoundaryPath = "contracts/otel/import_boundary.json";
const errorClassRegistryPath = "contracts/otel/error_class_registry.json";
const telemetryConfigSchemaPath = "contracts/otel/telemetry_config_schema.v2.json";
const configHazardMatrixPath = "contracts/otel/config_hazard_fixture_matrix.v2.json";
const corpusManifestPath = "internal/testutil/golden/otel/corpus_manifest.json";
const dependencyClassificationPath = "internal/testutil/golden/otel/dependency_update_classification.json";
const verificationContractPath = "contracts/verification/owners/platform.telemetry.json";
const publicErrorRegistryPath = "contracts/errors/index.json";
const generatedArtifactPolicyPath = "tools/generated_artifact_policy.json";
const expectedDigest = "3f8f80a2ed04521dfe29e50fcddd7f7de70145a6aee01959f985a65fbb4c8632";
const otelCommit = "d4a91bddb53b4c308df3e40171a60059183efd88";
const semconvCommit = "e018fe6f91862f5ed63c082f87697cddac596784";

const expectedTopLevelKeys = [
  "schema_id",
  "otel_spec_version",
  "otel_spec_ref",
  "otel_spec_commit_sha",
  "semconv_version",
  "semconv_ref",
  "semconv_commit_sha",
  "semconv_model_digest_algorithm",
  "semconv_model_digest",
  "sampler_profile_review_after",
  "sampler_profile_current_fractional",
  "probability_sampler_status",
  "semconv_generated_constants",
  "language_sdk_versions",
  "created_at",
  "created_by_tool",
];

const requiredPackageFamilies = new Set([
  "API",
  "SDK",
  "Trace SDK",
  "Metrics SDK",
  "Logs SDK or bridge",
  "OTLP HTTP exporters",
  "OTLP gRPC exporters",
  "Semantic-convention constants",
  "Instrumentation adapters",
  "Resource-detector packages",
  "Autoconfiguration or declarative-config packages",
  "Browser-side OTel packages",
]);

const expectedCorpusCases = [
  ["OTEL-CORPUS-001", "No-SDK mode"],
  ["OTEL-CORPUS-002", "Source baseline"],
  ["OTEL-CORPUS-003", "Cartulary environment binding parser"],
  ["OTEL-CORPUS-004", "Hostile environment"],
  ["OTEL-CORPUS-005", "Hostile declarative configuration"],
  ["OTEL-CORPUS-006", "HTTP route shape"],
  ["OTEL-CORPUS-007", "Workbook query and mutation"],
  ["OTEL-CORPUS-008", "WebSocket collaboration"],
  ["OTEL-CORPUS-009", "Background jobs"],
  ["OTEL-CORPUS-010", "Postgres dependency"],
  ["OTEL-CORPUS-011", "Object storage"],
  ["OTEL-CORPUS-012", "Resource identity"],
  ["OTEL-CORPUS-013", "Attribute boundary"],
  ["OTEL-CORPUS-014", "Logs"],
  ["OTEL-CORPUS-015", "Metrics"],
  ["OTEL-CORPUS-016", "Exporter"],
  ["OTEL-CORPUS-017", "Product-boundary runtime invariance"],
  ["OTEL-CORPUS-018", "Redaction"],
];

const expectedSpanFamilies = [
  "http_server",
  "workbook_query",
  "workbook_mutation",
  "workbook_projection",
  "websocket_lifecycle",
  "websocket_event_send",
  "job_enqueue",
  "job_run",
  "postgres_dependency",
  "objectstore_dependency",
];

const expectedMetricNames = [
  "cartulary.http.server.request.duration",
  "cartulary.workbook.query.duration",
  "cartulary.workbook.mutation.duration",
  "cartulary.workbook.rows.returned",
  "cartulary.collaboration.connections.active",
  "cartulary.collaboration.events.sent",
  "cartulary.jobs.active",
  "cartulary.jobs.duration",
  "cartulary.jobs.attempts",
  "cartulary.jobs.expired",
  "cartulary.jobs.lease_renewal.failures",
  "cartulary.postgres.operation.duration",
  "cartulary.objectstore.operation.duration",
  "cartulary.objectstore.transfer.bytes",
  "cartulary.telemetry.export.failure",
  "cartulary.telemetry.item.dropped",
  "cartulary.telemetry.queue.depth",
];

const expectedDurationBuckets = [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30];
const expectedByteBuckets = [1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456];
const expectedRowBuckets = [0, 1, 5, 10, 25, 50, 100, 250, 500];
const expectedRowBucketFixtureValues = [0, 1, 5, 6, 500, 501];

const expectedFixedTraceIDSamplerCorpus = [
  { trace_id: "10000000000000000000000000000000", sample_ratio: 0, allow: false },
  { trace_id: "10000000000000000000000000000000", sample_ratio: 1, allow: true },
  { trace_id: "10000000000000000000000000000000", sample_ratio: 0.25, allow: true },
  { trace_id: "10000000000000004000000000000000", sample_ratio: 0.25, allow: false },
  { trace_id: "1000000000000000ffffffffffffffff", sample_ratio: 0.5, allow: false },
];

const expectedNoSDKPathEvidence = [
  { path: "http", evidence: "internal/platform/telemetry/accessors_test.go::TestHTTPMiddlewareNoSDK" },
  { path: "workbook", evidence: "internal/modules/workbook/telemetry_test.go::TestWorkbookTelemetryNoSDK" },
  { path: "job", evidence: "internal/platform/jobs/telemetry_test.go::TestJobTelemetryNoSDK" },
  { path: "postgres", evidence: "internal/platform/postgres/telemetry_test.go::TestTelemetryDBPreservesDBBehaviorNoSDK" },
  { path: "object-store", evidence: "internal/platform/objectstore/telemetry_test.go::TestTelemetryStorePreservesStoreBehaviorNoSDK" },
  { path: "log-call", evidence: "internal/platform/telemetry/logs_test.go::TestLogCorrelationSafeFields" },
  { path: "websocket", evidence: "internal/modules/collaboration/hub_telemetry_test.go::TestWebSocketEventSendTelemetryNoSDK" },
];

const expectedHostileDeclarativeConfigAttempts = [
  "otlp_http_exporter",
  "otlp_grpc_per_signal_endpoint",
  "resource_detector_host_process_container_cloud_service",
  "sampler_adaptive_remote_plugin",
  "metric_reader_export_interval",
  "view_stream_rename_or_attribute_change",
  "exemplars",
  "log_bridge_or_log_exporter",
  "plugin_component_provider",
];

const expectedNonTransferAbsenceRules = [
  "required_collector_deployment",
  "collector_side_privacy_enforcement",
  "baggage_correlation",
  "environment_autoconfiguration",
  "declarative_plugin_providers",
  "default_otlp_localhost_export",
  "per_signal_exporters_endpoints",
  "prometheus_scrape_exporter",
  "zipkin_export",
  "jaeger_native_export",
  "vendor_native_export",
  "sql_commenter_propagation",
  "resource_detectors",
  "resource_schema_url_merge",
  "s3_semantic_attributes",
  "exemplars",
  "metric_bind_bypass",
  "logrecord_event_name",
  "logs_api_exception_parameter",
  "log_to_span_bridge",
  "probability_sampler",
];

const expectedLogSeverityMapping = {
  trace: { number: 1, text: "TRACE" },
  debug: { number: 5, text: "DEBUG" },
  info: { number: 9, text: "INFO" },
  warn: { number: 13, text: "WARN" },
  error: { number: 17, text: "ERROR" },
  fatal: { number: 21, text: "FATAL" },
  unknown: { number: 9, text: "INFO" },
};

const expectedForbiddenValueActionMatrix = [
  { family: "incident_authored_content", owner_literal_fixture: "timeline summary", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "incident_authored_content" },
  { family: "evidence_identity", owner_literal_fixture: "filename_hint", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "evidence_identity" },
  { family: "stable_identifier", owner_literal_fixture: "incident_id", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "stable_identifier" },
  { family: "security_material", owner_literal_fixture: "exporter header value", default_treatment: "omit", replacement_allowed: false, diagnostic_family: "security_material" },
  { family: "request_content", owner_literal_fixture: "raw query string", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "request_content" },
  { family: "database_content", owner_literal_fixture: "sql text", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "database_content" },
  { family: "infrastructure_identity", owner_literal_fixture: "client ip", default_treatment: "omit", replacement_allowed: true, diagnostic_family: "infrastructure_identity" },
  { family: "exception_detail", owner_literal_fixture: "exception.stacktrace", default_treatment: "replace_with_closed_class", replacement_allowed: true, diagnostic_family: "exception_detail" },
  { family: "baggage_trace_state", owner_literal_fixture: "tracestate", default_treatment: "omit", replacement_allowed: false, diagnostic_family: "baggage_trace_state" },
];

const expectedRuntimeInvarianceSurfaces = [
  "http_request",
  "workbook_query",
  "workbook_mutation",
  "websocket_send",
  "evidence_access",
  "background_job_transition",
];

const expectedRuntimeInvarianceFailures = [
  "exporter_failure",
  "exporter_timeout",
  "queue_overflow",
  "redaction_rejection",
];

const expectedResourceAttributes = [
  "service.name",
  "service.namespace",
  "service.version",
  "service.instance.id",
  "deployment.environment.name",
  "cartulary.deployment.profile",
  "cartulary.profile.claims",
];

const expectedForbiddenResourcePrefixes = [
  "host.",
  "process.",
  "os.",
  "container.",
  "k8s.",
  "cloud.",
  "faas.",
  "browser.",
  "device.",
];

const expectedNullOmissionSignalFamilies = [
  "spans",
  "metrics",
  "logs",
  "resource",
  "self_diagnostics",
];

const normalizedSignalSchemas = {
  normalized_traces: "cartulary.otel_normalized_traces.v2",
  normalized_metrics: "cartulary.otel_normalized_metrics.v2",
  normalized_logs: "cartulary.otel_normalized_logs.v2",
};

const expectedTelemetryConfigKeys = [
  "telemetry.enabled",
  "telemetry.otel_env_passthrough",
  "telemetry.exporter.kind",
  "telemetry.exporter.endpoint",
  "telemetry.exporter.headers",
  "telemetry.exporter.protocol",
  "telemetry.exporter.compression",
  "telemetry.exporter.retry.enabled",
  "telemetry.exporter.retry.max_elapsed_ms",
  "telemetry.exporter.retry.initial_interval_ms",
  "telemetry.exporter.retry.max_interval_ms",
  "telemetry.exporter.retry.multiplier",
  "telemetry.traces.enabled",
  "telemetry.traces.sample_ratio",
  "telemetry.traces.sampler_profile",
  "telemetry.traces.accept_remote_context",
  "telemetry.metrics.enabled",
  "telemetry.metrics.temporality_profile",
  "telemetry.metrics.exemplars.enabled",
  "telemetry.logs.bridge_enabled",
  "telemetry.logs.body_max_chars",
  "telemetry.processor.max_queue_size",
  "telemetry.processor.max_export_batch_size",
  "telemetry.processor.traces.schedule_delay_ms",
  "telemetry.processor.metrics.schedule_delay_ms",
  "telemetry.processor.logs.schedule_delay_ms",
  "telemetry.processor.export_timeout_ms",
  "telemetry.processor.overflow_policy",
  "telemetry.shutdown.flush_timeout_ms",
  "telemetry.self_diagnostics.enabled",
  "telemetry.self_diagnostics.recursion_guard",
  "telemetry.resource.service_name",
  "telemetry.resource.service_namespace",
  "telemetry.resource.service_version",
  "telemetry.resource.service_instance_id",
  "telemetry.resource.deployment_environment_name",
  "telemetry.attribute.incident_correlation",
  "telemetry.attribute.hmac_secret_ref",
];

const expectedCrossKeyRuleIDs = [
  "OTEL-CFG-001",
  "OTEL-CFG-002",
  "OTEL-CFG-003",
  "OTEL-CFG-004",
  "OTEL-CFG-005",
  "OTEL-CFG-006",
  "OTEL-CFG-007",
  "OTEL-CFG-008",
  "OTEL-CFG-008A",
  "OTEL-CFG-009",
  "OTEL-CFG-010",
  "OTEL-CFG-011",
  "OTEL-CFG-012",
  "OTEL-CFG-013",
  "OTEL-CFG-014",
  "OTEL-CFG-015",
  "OTEL-CFG-016",
  "OTEL-CFG-017",
  "OTEL-CFG-018",
];

const expectedHazardFixtureIDs = [
  "OTEL-ENV-001",
  "OTEL-ENV-002",
  "OTEL-ENV-003",
  "OTEL-ENV-004",
  "OTEL-ENV-005",
  "OTEL-ENV-006",
  "OTEL-ENV-007",
  "OTEL-ENV-008",
  "OTEL-ENV-009",
  "OTEL-ENV-010",
  "OTEL-ENV-011",
  "OTEL-ENV-012",
  "OTEL-ENV-013",
  "OTEL-ENV-014",
  "OTEL-ENV-015",
  "OTEL-ENV-016",
  "OTEL-ENV-017",
  "OTEL-ENV-018",
  "OTEL-ENV-019",
];

const expectedHostileEnvironmentFamilies = [
  "OTEL_ATTRIBUTE_COUNT_LIMIT",
  "OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT",
  "OTEL_BLRP_*",
  "OTEL_BSP_*",
  "OTEL_CONFIG_FILE",
  "OTEL_ENTITIES",
  "OTEL_EXPORTER_OTLP_*",
  "OTEL_EXPORTER_OTLP_ENDPOINT",
  "OTEL_EXPORTER_OTLP_LOGS_*",
  "OTEL_EXPORTER_OTLP_METRICS_*",
  "OTEL_EXPORTER_OTLP_TRACES_*",
  "OTEL_EXPORTER_PROMETHEUS_*",
  "OTEL_EXPORTER_ZIPKIN_*",
  "OTEL_EXPERIMENTAL_CONFIG_FILE",
  "OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT",
  "OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT",
  "OTEL_LOGS_EXPORTER",
  "OTEL_METRIC_EXPORT_INTERVAL",
  "OTEL_METRIC_EXPORT_TIMEOUT",
  "OTEL_METRICS_EXEMPLAR_FILTER",
  "OTEL_METRICS_EXPORTER",
  "OTEL_PROPAGATORS",
  "OTEL_RESOURCE_ATTRIBUTES",
  "OTEL_SDK_DISABLED",
  "OTEL_SEMCONV_STABILITY_OPT_IN",
  "OTEL_SERVICE_NAME",
  "OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT",
  "OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT",
  "OTEL_SPAN_EVENT_COUNT_LIMIT",
  "OTEL_SPAN_LINK_COUNT_LIMIT",
  "OTEL_TRACES_EXPORTER",
  "OTEL_TRACES_SAMPLER",
  "OTEL_TRACES_SAMPLER_ARG",
  "OTEL_LOG_LEVEL",
  "OTEL_{LANGUAGE}_{FEATURE}",
];

function repoPath(relativePath) {
  return path.join(repoRoot, relativePath);
}

const textCache = new Map();

function readText(relativePath) {
  if (!textCache.has(relativePath)) {
    textCache.set(relativePath, readFileSync(repoPath(relativePath), "utf8"));
  }
  return textCache.get(relativePath);
}

function readJSON(relativePath) {
  return JSON.parse(readText(relativePath));
}

function canonicalize(value) {
  if (Array.isArray(value)) {
    return value.map((entry) => canonicalize(entry));
  }
  if (value && typeof value === "object") {
    const out = {};
    for (const key of Object.keys(value).sort()) {
      out[key] = canonicalize(value[key]);
    }
    return out;
  }
  return value;
}

function canonicalJSON(value) {
  return `${JSON.stringify(canonicalize(value))}\n`;
}

function sha256Text(text) {
  return createHash("sha256").update(text).digest("hex");
}

function fileExists(relativePath) {
  try {
    return statSync(repoPath(relativePath)).isFile();
  } catch {
    return false;
  }
}

function absoluteFileExists(file) {
  try {
    return statSync(file).isFile();
  } catch {
    return false;
  }
}

function evidenceOK(actual, expected) {
  return (
    Array.isArray(actual) &&
    expected.every((entry) => actual.includes(entry)) &&
    actual.every((entry) => typeof entry === "string" && fileExists(entry.split("::")[0]))
  );
}

function assert(condition, message, checks, id) {
  if (!condition) {
    checks.push({ id, status: "fail", message });
    return false;
  }
  checks.push({ id, status: "pass", message });
  return true;
}

function publicErrorCodes() {
  return new Set(
    (readJSON(publicErrorRegistryPath).errors ?? []).map((entry) => entry.code),
  );
}

function validateSnapshot(snapshot, checks) {
  const keys = Object.keys(snapshot).sort();
  const expectedKeys = [...expectedTopLevelKeys].sort();
  assert(
    JSON.stringify(keys) === JSON.stringify(expectedKeys),
    "source snapshot has exactly the adopted top-level keys",
    checks,
    "snapshot.top_level_keys",
  );
  assert(snapshot.schema_id === "cartulary.otel_source_snapshot.v2", "schema_id is adopted", checks, "snapshot.schema_id");
  assert(snapshot.otel_spec_version === "1.57.0", "OTel spec version is pinned", checks, "snapshot.otel_spec_version");
  assert(snapshot.otel_spec_ref === "v1.57.0", "OTel spec ref is immutable", checks, "snapshot.otel_spec_ref");
  assert(snapshot.otel_spec_commit_sha === otelCommit, "OTel spec commit SHA is full and pinned", checks, "snapshot.otel_spec_commit_sha");
  assert(snapshot.semconv_version === "1.41.0", "semantic-conventions version is pinned", checks, "snapshot.semconv_version");
  assert(snapshot.semconv_ref === "v1.41.0", "semantic-conventions ref is immutable", checks, "snapshot.semconv_ref");
  assert(snapshot.semconv_commit_sha === semconvCommit, "semantic-conventions commit SHA is full and pinned", checks, "snapshot.semconv_commit_sha");
  assert(snapshot.semconv_model_digest_algorithm === "semconv_model_digest_v1", "digest algorithm is adopted", checks, "snapshot.digest_algorithm");
  assert(snapshot.semconv_model_digest === expectedDigest, "semantic-conventions model digest is concrete", checks, "snapshot.digest");
  assert(snapshot.sampler_profile_review_after === "2027-01-01", "sampler review metadata is pinned", checks, "snapshot.sampler.review_after");
  assert(
    snapshot.sampler_profile_current_fractional === "cartulary.sampler.traceidratio_compat.v1",
    "current fractional sampler profile metadata is pinned",
    checks,
    "snapshot.sampler.current_fractional",
  );
  assert(
    snapshot.probability_sampler_status === "not_adopted_development",
    "ProbabilitySampler non-adoption metadata is pinned",
    checks,
    "snapshot.sampler.probability_status",
  );

  const constants = snapshot.semconv_generated_constants ?? {};
  assert(constants.source_kind === "repo_codegen", "generated constants source kind is closed", checks, "snapshot.constants.source_kind");
  assert(constants.input_model_digest === expectedDigest, "generated constants bind to model digest", checks, "snapshot.constants.digest_binding");
  const generatorErrors = validateOtelGeneratorReference({
    root: repoRoot,
    generatorSourceRef: constants.generator_source_ref,
    generatorSourceSHA: constants.generator_source_sha,
    requireTracked: true,
  });
  assert(
    generatorErrors.length === 0,
    `generated constants source SHA matches the tracked OTel contract generator${generatorErrors.length ? `; ${generatorErrors.join("; ")}` : ""}`,
    checks,
    "snapshot.constants.source_sha",
  );

  const families = new Set((snapshot.language_sdk_versions ?? []).map((row) => row.package_family));
  const missingFamilies = [...requiredPackageFamilies].filter((family) => !families.has(family));
  assert(missingFamilies.length === 0, `language SDK package-family rows are exhaustive${missingFamilies.length ? `; missing ${missingFamilies.join(", ")}` : ""}`, checks, "snapshot.package_families");
}

function validateGeneratedConstantsManifest(manifest, checks) {
  assert(
    manifest.schema_id === "cartulary.otel_generated_constants_manifest.v2",
    "generated constants manifest uses the adopted schema",
    checks,
    "constants_manifest.schema_id",
  );
  assert(manifest.source_kind === "repo_codegen", "generated constants source kind is repo_codegen", checks, "constants_manifest.source_kind");
  assert(manifest.input_model_digest === expectedDigest, "generated constants manifest binds to the adopted model digest", checks, "constants_manifest.digest");
  assert(manifest.generator_source_ref === otelGeneratorSourceRef, "generated constants manifest names the OTel contract generator", checks, "constants_manifest.generator_source_ref");
  assert(manifest.output_path === "contracts/otel/semantic_conventions_constants.v1.json", "generated constants manifest names the repo-control output", checks, "constants_manifest.output_path");
}

function validateTelemetryConfigSchema(schema, checks) {
  assert(
    schema.schema_id === "cartulary.otel_telemetry_config_schema.v2",
    "telemetry config schema uses the adopted schema identifier",
    checks,
    "telemetry_config.schema_id",
  );
  assert(
    schema.source_owner === "platform.telemetry.verification.current_conformance",
    "telemetry config schema names its machine verification owner",
    checks,
    "telemetry_config.source_owner",
  );
  assert(schema.error_code === "invalid_deployment_config", "telemetry config schema records the deployment error code", checks, "telemetry_config.error_code");
  assert(schema.telemetry_reason_code === "invalid_telemetry_config", "telemetry config schema records the telemetry reason code", checks, "telemetry_config.reason_code");
  assert(
    typeof schema.environment_binding === "string" &&
      schema.environment_binding.includes("CARTULARY__") &&
      schema.environment_binding.includes("empty"),
    "telemetry config schema records Core04 overlay and empty-env behavior",
    checks,
    "telemetry_config.environment_binding",
  );

  const entries = Array.isArray(schema.entries) ? schema.entries : [];
  const keys = entries.map((entry) => entry.key);
  assert(
    JSON.stringify(keys) === JSON.stringify(expectedTelemetryConfigKeys),
    `telemetry config schema key order matches the adopted table${JSON.stringify(keys) !== JSON.stringify(expectedTelemetryConfigKeys) ? `; saw ${keys.join(", ")}` : ""}`,
    checks,
    "telemetry_config.keys",
  );
  const entriesByKey = new Map(entries.map((entry) => [entry.key, entry]));
  for (const key of expectedTelemetryConfigKeys) {
    const entry = entriesByKey.get(key);
    if (!entry) {
      continue;
    }
    const hasRequiredShape =
      typeof entry.type === "string" &&
      Object.hasOwn(entry, "default") &&
      Object.hasOwn(entry, "allowed") &&
      typeof entry.omitted === "string" &&
      typeof entry.explicit_null === "string" &&
      entry.empty_env === "omit" &&
      typeof entry.secret === "boolean" &&
      entry.failure === "invalid_deployment_config";
    assert(hasRequiredShape, `${key} records type/default/bounds/null/empty-env/secret/failure behavior`, checks, `telemetry_config.${key}.shape`);
  }

  const headers = entriesByKey.get("telemetry.exporter.headers") ?? {};
  const headerAllowed = String(headers.allowed ?? "");
  assert(headers.secret === true, "telemetry exporter headers are marked secret-bearing", checks, "telemetry_config.headers.secret");
  assert(
    headers.type === "map_string_secret_ref_v1" &&
      headerAllowed.includes("at_most_16_headers") &&
      headerAllowed.includes("resolved_value_bytes_lte_4096") &&
      headerAllowed.includes("configured_header_block_bytes_lte_8192") &&
      headerAllowed.includes("non_protocol_owned"),
    "telemetry exporter header schema records secret_ref_v1 count, value, total-size, duplicate, and protocol-owned-header rules",
    checks,
    "telemetry_config.headers.rules",
  );

  const hmac = entriesByKey.get("telemetry.attribute.hmac_secret_ref") ?? {};
  assert(hmac.secret === true && hmac.type === "secret_ref_v1_or_null", "telemetry HMAC secret ref is marked as nullable secret_ref_v1", checks, "telemetry_config.hmac_secret_ref");

  const crossRuleIDs = (schema.cross_key_rules ?? []).map((row) => row.id);
  assert(
    JSON.stringify(crossRuleIDs) === JSON.stringify(expectedCrossKeyRuleIDs),
    `telemetry config schema records every cross-key rule in owner order${JSON.stringify(crossRuleIDs) !== JSON.stringify(expectedCrossKeyRuleIDs) ? `; saw ${crossRuleIDs.join(", ")}` : ""}`,
    checks,
    "telemetry_config.cross_key_rules",
  );

  const hostileFamilies = schema.hostile_environment_families ?? [];
  assert(
    JSON.stringify(hostileFamilies) === JSON.stringify(expectedHostileEnvironmentFamilies),
    `telemetry config schema records every hostile OTel environment family in owner order${JSON.stringify(hostileFamilies) !== JSON.stringify(expectedHostileEnvironmentFamilies) ? `; saw ${hostileFamilies.join(", ")}` : ""}`,
    checks,
    "telemetry_config.hostile_env",
  );
}

function validateConfigHazardMatrix(matrix, checks) {
  assert(
    matrix.schema_id === "cartulary.otel_config_hazard_fixture_matrix.v2",
    "config/hazard matrix uses the adopted schema identifier",
    checks,
    "config_hazard.schema_id",
  );
  assert(
    matrix.source_owner === "platform.telemetry.verification.current_conformance",
    "config/hazard matrix names its machine verification owner",
    checks,
    "config_hazard.source_owner",
  );
  assert(matrix.default_error_code === "invalid_deployment_config", "config/hazard matrix records the deployment error code", checks, "config_hazard.error_code");
  assert(matrix.default_reason_code === "invalid_telemetry_config", "config/hazard matrix records the telemetry reason code", checks, "config_hazard.reason_code");

  const crossRules = Array.isArray(matrix.cross_key_rules) ? matrix.cross_key_rules : [];
  const crossRuleIDs = crossRules.map((row) => row.id);
  assert(
    JSON.stringify(crossRuleIDs) === JSON.stringify(expectedCrossKeyRuleIDs),
    `config/hazard matrix records every OTEL-CFG row in owner order${JSON.stringify(crossRuleIDs) !== JSON.stringify(expectedCrossKeyRuleIDs) ? `; saw ${crossRuleIDs.join(", ")}` : ""}`,
    checks,
    "config_hazard.cross_key_ids",
  );
  for (const row of crossRules) {
    const rowOK =
      typeof row.fixture_id === "string" &&
      row.fixture_id.length > 0 &&
      Array.isArray(row.inputs) &&
      row.inputs.length > 0 &&
      typeof row.forbidden_effect_assertion === "string" &&
      row.forbidden_effect_assertion.length > 0 &&
      Array.isArray(row.evidence) &&
      row.evidence.length > 0;
    assert(rowOK, `${row.id} has fixture input, forbidden-effect assertion, and evidence`, checks, `config_hazard.cross_key.${row.id}.shape`);
  }

  const hazardFixtures = Array.isArray(matrix.hazard_fixtures) ? matrix.hazard_fixtures : [];
  const hazardIDs = hazardFixtures.map((row) => row.id);
  assert(
    JSON.stringify(hazardIDs) === JSON.stringify(expectedHazardFixtureIDs),
    `config/hazard matrix records every OTEL-ENV row in owner order${JSON.stringify(hazardIDs) !== JSON.stringify(expectedHazardFixtureIDs) ? `; saw ${hazardIDs.join(", ")}` : ""}`,
    checks,
    "config_hazard.hazard_ids",
  );
  for (const row of hazardFixtures) {
    const rowOK =
      typeof row.environment_family === "string" &&
      row.environment_family.length > 0 &&
      typeof row.members_or_pattern === "string" &&
      row.members_or_pattern.length > 0 &&
      Array.isArray(row.representative_inputs) &&
      row.representative_inputs.length > 0 &&
      typeof row.forbidden_effect === "string" &&
      row.forbidden_effect.length > 0 &&
      typeof row.assertion === "string" &&
      row.assertion.length > 0 &&
      Array.isArray(row.evidence) &&
      row.evidence.length > 0;
    assert(rowOK, `${row.id} has representative inputs, forbidden-effect assertion, and evidence`, checks, `config_hazard.hazard.${row.id}.shape`);
  }

  const declarative = hazardFixtures.find((row) => row.id === "OTEL-ENV-017") ?? {};
  const declarativeInputs = (declarative.representative_inputs ?? []).join(" ");
  assert(
    declarativeInputs.includes("OTEL_CONFIG_FILE") && declarativeInputs.includes("OTEL_EXPERIMENTAL_CONFIG_FILE"),
    "declarative-config hazard fixture covers both stable and experimental config env names",
    checks,
    "config_hazard.declarative_env_names",
  );
}

let repositoryFileIndex = null;

function collectFiles(root, excludedDirectories, files = []) {
  const absoluteRoot = repoPath(root);
  let entries;
  try {
    entries = readdirSync(absoluteRoot, { withFileTypes: true });
  } catch {
    return files;
  }
  for (const entry of entries) {
    const relative = path.join(root, entry.name);
    if (entry.isDirectory()) {
      if (excludedDirectories.has(entry.name)) {
        continue;
      }
      collectFiles(relative, excludedDirectories, files);
      continue;
    }
    if (entry.isFile()) {
      files.push(relative);
    }
  }
  return files;
}

function walkFiles(root, predicate, files = []) {
  repositoryFileIndex ??= collectFiles(
    ".",
    new Set([
      ".git",
      "node_modules",
      "tmp",
      ".cartulary",
      ".cache",
      ".pnpm-store",
      "dist",
    ]),
  );
  const prefix = root === "." ? "" : `${root.replace(/\/$/u, "")}/`;
  for (const relative of repositoryFileIndex) {
    if ((prefix === "" || relative.startsWith(prefix)) && predicate(relative)) {
      files.push(relative);
    }
  }
  return files;
}

const buildFileIndexes = new Map();

function walkFilesIncludingBuild(root, predicate, files = []) {
  if (!buildFileIndexes.has(root)) {
    buildFileIndexes.set(
      root,
      collectFiles(
        root,
        new Set([".git", "node_modules", "tmp", ".cartulary", ".cache", ".pnpm-store"]),
      ),
    );
  }
  for (const relative of buildFileIndexes.get(root)) {
    if (predicate(relative)) {
      files.push(relative);
    }
  }
  return files;
}

function goImports(source) {
  const imports = [];
  for (const match of source.matchAll(/import\s+"([^"]+)"/g)) {
    imports.push(match[1]);
  }
  for (const block of source.matchAll(/import\s*\(([\s\S]*?)\)/g)) {
    for (const line of block[1].split("\n")) {
      const match = line.match(/"([^"]+)"/);
      if (match) {
        imports.push(match[1]);
      }
    }
  }
  return imports;
}

function underAny(relativePath, roots) {
  return roots.some((root) => relativePath === root || relativePath.startsWith(`${root}/`));
}

function hasSuffixAny(relativePath, suffixes) {
  return suffixes.some((suffix) => relativePath.endsWith(suffix));
}

function frontendSourceFile(relativePath, excludeSuffixes = []) {
  if (hasSuffixAny(relativePath, excludeSuffixes)) {
    return false;
  }
  return [".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"].includes(path.extname(relativePath));
}

function scanForbiddenPatterns(files, patterns) {
  const hits = [];
  for (const file of files) {
    const text = readText(file);
    const lower = text.toLowerCase();
    for (const row of patterns) {
      const pattern = typeof row === "string" ? row : row.pattern;
      const id = typeof row === "string" ? row : row.id;
      if (lower.includes(pattern.toLowerCase())) {
        hits.push(`${file}:${id}`);
      }
    }
  }
  return hits;
}

function dynamicImportViolations(files, patterns) {
  const violations = [];
  for (const file of files) {
    const source = readText(file);
    let literalCount = 0;
    for (const match of source.matchAll(/import\s*\(\s*(["'`])([^"'`]+)\1\s*\)/g)) {
      literalCount += 1;
      const specifier = match[2].toLowerCase();
      for (const row of patterns) {
        const pattern = typeof row === "string" ? row : row.pattern;
        const id = typeof row === "string" ? row : row.id;
        if (specifier.includes(pattern.toLowerCase())) {
          violations.push(`${file}:${id}:${match[2]}`);
        }
      }
    }
    const importCallCount = [...source.matchAll(/import\s*\(/g)].length;
    if (importCallCount !== literalCount) {
      violations.push(`${file}:non_literal_dynamic_import`);
    }
  }
  return violations;
}

function validateImportBoundary(boundary, checks) {
  assert(boundary.schema_id === "cartulary.otel_import_boundary.v1", "import-boundary manifest uses the adopted schema", checks, "import_boundary.schema_id");
  const shapeErrors = validateOtelImportBoundaryContractShape(repoRoot, boundary, { requireEvidenceFile: true });
  assert(
    shapeErrors.length === 0,
    `import-boundary manifest has the closed contract shape and live evidence references${shapeErrors.length ? `; ${shapeErrors.join("; ")}` : ""}`,
    checks,
    "import_boundary.shape",
  );

  const allowedAPIs = new Set(boundary.allowed_ordinary_otel_imports ?? []);
  const forbiddenPrefixes = boundary.forbidden_ordinary_otel_import_prefixes ?? [];
  const telemetryBootstrapRoots = boundary.telemetry_bootstrap_roots ?? [];
  const goFiles = walkFiles(".", (relative) => relative.endsWith(".go"));
  const violations = [];
  for (const file of goFiles) {
    if (file.startsWith("tools/") || file.startsWith("tmp/")) {
      continue;
    }
    for (const importPath of goImports(readText(file))) {
      if (!importPath.startsWith("go.opentelemetry.io/")) {
        continue;
      }
      const inTelemetryBootstrap = underAny(file, telemetryBootstrapRoots);
      if (inTelemetryBootstrap) {
        continue;
      }
      const allowed = allowedAPIs.has(importPath);
      const forbidden = forbiddenPrefixes.some((prefix) => importPath === prefix || importPath.startsWith(prefix));
      if (forbidden || !allowed) {
        violations.push(`${file}:${importPath}`);
      }
    }
  }
  assert(
    violations.length === 0,
    `authored Go imports obey telemetry boundary${violations.length ? `; violations ${violations.join(", ")}` : ""}`,
    checks,
    "import_boundary.go_authored",
  );

  const packageFiles = walkFiles(".", (relative) =>
    relative.endsWith("package.json") &&
    ((boundary.browser_package_manifest_roots ?? []).some((root) => relative.startsWith(`${root}/`))),
  );
  const patterns = boundary.forbidden_browser_package_patterns ?? [];
  const browserViolations = [];
  for (const file of packageFiles) {
    const pkg = readJSON(file);
    const deps = {
      ...(pkg.dependencies ?? {}),
      ...(pkg.devDependencies ?? {}),
      ...(pkg.peerDependencies ?? {}),
      ...(pkg.optionalDependencies ?? {}),
    };
    for (const name of Object.keys(deps)) {
      if (patterns.some((pattern) => name.toLowerCase().includes(pattern.toLowerCase()))) {
        browserViolations.push(`${file}:${name}`);
      }
    }
  }
  assert(
    browserViolations.length === 0,
    `browser package manifests contain no forbidden telemetry exporters or vendor SDKs${browserViolations.length ? `; violations ${browserViolations.join(", ")}` : ""}`,
    checks,
    "import_boundary.browser_packages",
  );

  const browserExcludeSuffixes = boundary.browser_source_exclude_suffixes ?? [];
  const runtimePatterns = boundary.forbidden_browser_runtime_patterns ?? [];
  const generatedArtifactPolicy = readJSON(generatedArtifactPolicyPath);
  const generatedRoots = (generatedArtifactPolicy.generated_roots ?? []).map((entry) => entry.path);
  const browserSourceFiles = [];
  for (const root of boundary.browser_source_roots ?? []) {
    walkFiles(
      root,
      (relative) => frontendSourceFile(relative, browserExcludeSuffixes) && !underAny(relative, generatedRoots),
      browserSourceFiles,
    );
  }
  const browserSourceViolations = scanForbiddenPatterns(browserSourceFiles, runtimePatterns);
  assert(
    browserSourceViolations.length === 0,
    `browser authored source contains no forbidden exporter, vendor SDK, analytics, or telemetry-authority strings${browserSourceViolations.length ? `; violations ${browserSourceViolations.join(", ")}` : ""}`,
    checks,
    "import_boundary.browser_source",
  );

  const runtimeSourceFiles = [];
  for (const root of boundary.browser_runtime_source_roots ?? []) {
    walkFiles(
      root,
      (relative) => frontendSourceFile(relative, browserExcludeSuffixes) && !underAny(relative, generatedRoots),
      runtimeSourceFiles,
    );
  }
  const dynamicViolations = dynamicImportViolations(runtimeSourceFiles, runtimePatterns);
  assert(
    dynamicViolations.length === 0,
    `browser runtime dynamic imports are literal and contain no forbidden telemetry packages${dynamicViolations.length ? `; violations ${dynamicViolations.join(", ")}` : ""}`,
    checks,
    "import_boundary.browser_dynamic_imports",
  );

  const bundleExtensions = new Set(boundary.browser_text_bundle_extensions ?? []);
  const bundleFiles = [];
  const sourceMapFiles = [];
  for (const root of boundary.browser_built_bundle_roots ?? []) {
    walkFilesIncludingBuild(root, (relative) => bundleExtensions.has(path.extname(relative)), bundleFiles);
    walkFilesIncludingBuild(root, (relative) => relative.endsWith(".map"), sourceMapFiles);
  }
  const bundleViolations = scanForbiddenPatterns(bundleFiles, runtimePatterns);
  assert(
    bundleFiles.length > 0,
    "browser built bundle text artifacts are present for static inspection",
    checks,
    "import_boundary.browser_bundle_present",
  );
  assert(
    bundleViolations.length === 0,
    `browser built bundle contains no forbidden exporter, vendor SDK, analytics, or telemetry-authority strings${bundleViolations.length ? `; violations ${bundleViolations.join(", ")}` : ""}`,
    checks,
    "import_boundary.browser_bundle",
  );
  assert(
    sourceMapFiles.length > 0,
    "browser source-map artifacts are present for closed source-map inspection",
    checks,
    "import_boundary.browser_source_maps",
  );

  const runtimeProbe = boundary.browser_runtime_probe ?? {};
  const runtimeProbeOK =
    runtimeProbe.make_target === expectedBrowserRuntimeProbe.make_target &&
    JSON.stringify(runtimeProbe.state_sources ?? []) === JSON.stringify(expectedBrowserRuntimeProbe.state_sources) &&
    JSON.stringify(runtimeProbe.forbidden_effects ?? []) === JSON.stringify(expectedBrowserRuntimeProbe.forbidden_effects) &&
    shapeErrors.length === 0;
  assert(
    runtimeProbeOK,
    "browser runtime probe evidence covers state sources, forbidden effects, and a live frontend-unit test",
    checks,
    "import_boundary.browser_runtime_probe",
  );
  assert(
    JSON.stringify(boundary.non_transfer_absence_rules ?? []) === JSON.stringify(expectedNonTransferAbsenceRules),
    "browser and runtime non-transfer absence rules match the adopted OTel registry",
    checks,
    "import_boundary.non_transfer_absence_rules",
  );
}

function validateVerificationOwner(checks) {
  const contract = readJSON(verificationContractPath);
  assert(
    contract.schema_id === "cartulary.verification_contract.v3" &&
      contract.owner_id === "platform.telemetry",
    "telemetry verification contract has the current machine owner identity",
    checks,
    "verification.owner_identity",
  );
  const verification = (contract.verifications ?? []).find(
    (entry) =>
      entry.verification_id ===
      "platform.telemetry.verification.current_conformance",
  );
  assert(
    verification?.profile === "support" &&
      (verification?.evidence_kinds ?? []).includes("static_check"),
    "telemetry conformance has a current support-profile machine verification",
    checks,
    "verification.current_conformance",
  );

  for (const dir of [
    "internal/testutil/golden/otel",
    "internal/testutil/golden/otel/source-snapshot",
    "internal/testutil/golden/otel/signals",
  ]) {
    assert(statSync(repoPath(dir)).isDirectory(), `${dir} exists`, checks, `golden.${dir}`);
  }
}

function signalPayloadField(field) {
  if (field === "normalized_traces") {
    return "spans";
  }
  if (field === "normalized_metrics") {
    return "metrics";
  }
  return "logs";
}

function signalKind(field) {
  if (field === "normalized_traces") {
    return "traces";
  }
  if (field === "normalized_metrics") {
    return "metrics";
  }
  return "logs";
}

function normalizeRawCapture(raw, schemaID) {
  const payloadField = raw.payload_field;
  return {
    schema_id: schemaID,
    case_id: raw.case_id,
    [payloadField]: raw.payload,
  };
}

function runRootRelative(absolutePath) {
  return path.relative(repoRoot, absolutePath);
}

function emitAndCompareCorpusCaptures(manifest, cases, checks) {
  const outDir = targetDir();
  const rawRoot = path.join(outDir, "raw");
  const normalizedRoot = path.join(outDir, "normalized");
  mkdirSync(rawRoot, { recursive: true, mode: 0o700 });
  mkdirSync(normalizedRoot, { recursive: true, mode: 0o700 });

  const rows = [];
  const mismatches = [];
  const nonCanonicalGoldens = [];
  for (const row of cases) {
    for (const [field, schemaID] of Object.entries(normalizedSignalSchemas)) {
      const payloadField = signalPayloadField(field);
      const signalPath = `${manifest.normalized_golden_root}/${row[field]}`;
      const signal = readJSON(signalPath);
      const canonicalGolden = canonicalJSON(signal);
      const committedGolden = readText(signalPath);
      if (committedGolden !== canonicalGolden) {
        nonCanonicalGoldens.push(signalPath);
      }

      const raw = {
        schema_id: "cartulary.otel_raw_capture.v2",
        corpus_revision: manifest.corpus_revision,
        case_id: row.case_id,
        signal_kind: signalKind(field),
        captured_at: new Date().toISOString(),
        resource: {
          schema_url: "",
          attributes: {
            "service.version": "0.0.0+unknown",
          },
        },
        instrumentation_scope: {
          name: "cartulary.otel-conformance",
          version: "0.0.0+unknown",
          schema_url: "",
          attributes: {},
        },
        payload_field: payloadField,
        payload: signal[payloadField],
      };
      const rawPath = path.join(rawRoot, row.case_id, `${signalKind(field)}.json`);
      mkdirSync(path.dirname(rawPath), { recursive: true, mode: 0o700 });
      writeFileSync(rawPath, canonicalJSON(raw));

      const normalized = normalizeRawCapture(raw, schemaID);
      const canonicalNormalized = canonicalJSON(normalized);
      const normalizedPath = path.join(normalizedRoot, row.case_id, `${signalKind(field)}.json`);
      mkdirSync(path.dirname(normalizedPath), { recursive: true, mode: 0o700 });
      writeFileSync(normalizedPath, canonicalNormalized);

      const rawSHA = sha256Text(canonicalJSON(raw));
      const normalizedSHA = sha256Text(canonicalNormalized);
      const goldenSHA = sha256Text(canonicalGolden);
      if (normalizedSHA !== goldenSHA) {
        mismatches.push(`${row.case_id}:${field}`);
      }
      rows.push({
        case_id: row.case_id,
        signal_kind: signalKind(field),
        raw_capture: runRootRelative(rawPath),
        normalized_capture: runRootRelative(normalizedPath),
        committed_golden: signalPath,
        raw_sha256: rawSHA,
        normalized_sha256: normalizedSHA,
        committed_golden_sha256: goldenSHA,
        comparison: normalizedSHA === goldenSHA ? "match" : "mismatch",
      });
    }
  }

  const comparison = {
    schema_id: "cartulary.otel_corpus_comparison.v1",
    status: mismatches.length === 0 && nonCanonicalGoldens.length === 0 ? "pass" : "fail",
    classification: mismatches.length === 0 && nonCanonicalGoldens.length === 0 ? "registry_equivalent" : "breaking_shape_change",
    corpus_manifest: corpusManifestPath,
    corpus_revision: manifest.corpus_revision,
    normalized_golden_root: manifest.normalized_golden_root,
    raw_capture_root: runRootRelative(rawRoot),
    normalized_capture_root: runRootRelative(normalizedRoot),
    case_count: cases.length,
    signal_file_count: rows.length,
    mismatches,
    non_canonical_goldens: nonCanonicalGoldens,
    rows,
  };
  const comparisonPath = path.join(outDir, "otel-corpus-comparison.json");
  writeFileSync(comparisonPath, canonicalJSON(comparison));

  assert(rows.length === expectedCorpusCases.length * Object.keys(normalizedSignalSchemas).length, "raw capture comparison covers every corpus case and signal family", checks, "golden.corpus.capture_count");
  assert(mismatches.length === 0, `normalized run captures match committed goldens${mismatches.length ? `; mismatches ${mismatches.join(", ")}` : ""}`, checks, "golden.corpus.normalized_comparison");
  assert(
    nonCanonicalGoldens.length === 0,
    `committed normalized goldens are byte-identical to canonical reserialization${nonCanonicalGoldens.length ? `; non-canonical ${nonCanonicalGoldens.join(", ")}` : ""}`,
    checks,
    "golden.corpus.canonical_reserialization",
  );
  assert(fileExists(runRootRelative(comparisonPath)), "corpus comparison summary is retained under the harness run root", checks, "golden.corpus.comparison_artifact");
  assert(statSync(rawRoot).isDirectory(), "raw captures are retained under the harness run root", checks, "golden.corpus.raw_capture_artifact");
}

function validateGoldenCorpus(manifest, classification, checks) {
  assert(
    manifest.schema_id === "cartulary.otel_corpus_manifest.v2",
    "golden corpus manifest uses the adopted schema",
    checks,
    "golden.corpus.schema_id",
  );
  assert(
    manifest.normalized_golden_root === "internal/testutil/golden/otel",
    "golden corpus root is the adopted normalized root",
    checks,
    "golden.corpus.root",
  );
  assert(
    manifest.raw_capture_policy?.retained_root === "harness_run_root/otel-conformance/raw",
    "raw OTel captures are retained under the harness run root",
    checks,
    "golden.corpus.raw_retained_root",
  );
  assert(
    manifest.raw_capture_policy?.forbidden_committed_root === "internal/testutil/golden/otel",
    "raw OTel captures are forbidden under the committed golden root",
    checks,
    "golden.corpus.raw_forbidden_root",
  );

  const normalization = manifest.canonical_normalization ?? {};
  const normalizationOK =
    normalization.service_version_token === "SERVICE_VERSION" &&
    normalization.scope_version_token === "SCOPE_VERSION" &&
    normalization.resource_schema_url === "" &&
    normalization.sort_object_keys === true &&
    normalization.canonical_numbers === true;
  assert(normalizationOK, "golden corpus canonical normalization settings are closed", checks, "golden.corpus.normalization");

  const dependencyGate = manifest.dependency_update_gate ?? {};
  assert(
    dependencyGate.manifest === dependencyClassificationPath,
    "golden corpus names the dependency-update classification manifest",
    checks,
    "golden.corpus.dependency_manifest",
  );
  assert(
    JSON.stringify(dependencyGate.allowed_without_nlspec_revision ?? []) === JSON.stringify(["registry_equivalent"]),
    "only registry-equivalent dependency updates may avoid an NLSpec revision",
    checks,
    "golden.corpus.dependency_allowed_without_revision",
  );
  assert(
    JSON.stringify(dependencyGate.requires_nlspec_revision ?? []) ===
      JSON.stringify(["additive_non_breaking", "privacy_tightening", "breaking_shape_change"]),
    "non-equivalent dependency updates require an NLSpec revision",
    checks,
    "golden.corpus.dependency_requires_revision",
  );

  const rawPathHits = walkFilesIncludingBuild("internal/testutil/golden/otel", () => true).filter((file) => {
    const segments = file.split(path.sep);
    return segments.some((segment) => ["raw", "raw-capture", "raw-captures", "captures"].includes(segment)) || file.endsWith(".raw");
  });
  assert(
    rawPathHits.length === 0,
    `committed golden corpus contains no raw-capture artifacts${rawPathHits.length ? `; found ${rawPathHits.join(", ")}` : ""}`,
    checks,
    "golden.corpus.no_committed_raw",
  );

  const cases = Array.isArray(manifest.cases) ? manifest.cases : [];
  assert(cases.length === expectedCorpusCases.length, "golden corpus has exactly the adopted case count", checks, "golden.corpus.case_count");
  const casesByID = new Map(cases.map((row) => [row.case_id, row]));
  const expectedIDs = expectedCorpusCases.map(([caseID]) => caseID);
  const actualIDs = cases.map((row) => row.case_id);
  assert(
    JSON.stringify(actualIDs) === JSON.stringify(expectedIDs),
    `golden corpus case order matches the adopted registry${JSON.stringify(actualIDs) !== JSON.stringify(expectedIDs) ? `; saw ${actualIDs.join(", ")}` : ""}`,
    checks,
    "golden.corpus.case_order",
  );

  for (const [caseID, title] of expectedCorpusCases) {
    const row = casesByID.get(caseID);
    if (
      !assert(
        row !== undefined,
        `${caseID} is present in the golden corpus manifest`,
        checks,
        `golden.corpus.${caseID}.present`,
      )
    ) {
      continue;
    }
    assert(row.title === title, `${caseID} title matches the adopted corpus title`, checks, `golden.corpus.${caseID}.title`);
    const expectedInput = `cases/${caseID}/input.json`;
    assert(row.input === expectedInput, `${caseID} input path is canonical`, checks, `golden.corpus.${caseID}.input_path`);
    const inputPath = `${manifest.normalized_golden_root}/${expectedInput}`;
    if (assert(fileExists(inputPath), `${caseID} input fixture exists`, checks, `golden.corpus.${caseID}.input_exists`)) {
      const input = readJSON(inputPath);
      assert(input.schema_id === "cartulary.otel_corpus_input.v2", `${caseID} input schema is current`, checks, `golden.corpus.${caseID}.input_schema`);
      assert(input.case_id === caseID, `${caseID} input fixture is bound to the case ID`, checks, `golden.corpus.${caseID}.input_case_id`);
      assert(input.title === title, `${caseID} input title matches the manifest`, checks, `golden.corpus.${caseID}.input_title`);
      if (caseID === "OTEL-CORPUS-001") {
        const noSDK = input.no_sdk_assertions ?? {};
        const paths = noSDK.product_paths ?? [];
        assert(
          noSDK.sdk_provider_status === "absent" &&
            noSDK.expected_emitted_telemetry === "none" &&
            noSDK.static_import_boundary === importBoundaryPath &&
            noSDK.scope_evidence === "internal/platform/telemetry/accessors_test.go::TestAccessorsNoSDKRegisteredScopes" &&
            JSON.stringify(paths) === JSON.stringify(expectedNoSDKPathEvidence),
          "OTEL-CORPUS-001 records no-SDK product paths, static import-boundary evidence, and registered-scope evidence",
          checks,
          "golden.corpus.OTEL-CORPUS-001.no_sdk_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-002") {
        const sourceBaseline = input.source_baseline_assertions ?? {};
        assert(
          sourceBaseline.snapshot === snapshotPath &&
            sourceBaseline.generated_constants_manifest === generatedConstantsManifestPath &&
            sourceBaseline.sampler_profile_review_after === "2027-01-01" &&
            sourceBaseline.sampler_profile_current_fractional === "cartulary.sampler.traceidratio_compat.v1" &&
            sourceBaseline.probability_sampler_status === "not_adopted_development" &&
            JSON.stringify(sourceBaseline.fixed_trace_id_corpus ?? []) === JSON.stringify(expectedFixedTraceIDSamplerCorpus) &&
            evidenceOK(sourceBaseline.evidence, [
              "internal/platform/telemetry/bootstrap_test.go::TestBootstrapNoSDKExportDisabled",
              "internal/platform/telemetry/sampler_test.go::TestResolveSamplerProfile",
              "internal/platform/telemetry/sampler_test.go::TestFixedTraceIDRatioCompatCorpus",
            ]),
          "OTEL-CORPUS-002 records source snapshot, generated constants, sampler metadata, and fixed TraceID sampler corpus evidence",
          checks,
          "golden.corpus.OTEL-CORPUS-002.source_baseline_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-004") {
        const hostile = input.hostile_environment_assertions ?? {};
        assert(
          hostile.hazard_matrix === configHazardMatrixPath &&
            JSON.stringify(hostile.fixture_ids ?? []) === JSON.stringify(expectedHazardFixtureIDs) &&
            hostile.exporter_kind === "none" &&
            hostile.network_export_enabled === false &&
            hostile.default_localhost_endpoint_contacted === false &&
            hostile.resource_schema_url === "" &&
            hostile.remote_context_accepted === false &&
            hostile.exemplars_enabled === false &&
            hostile.log_bridge_enabled === false &&
            hostile.declarative_config_component_active === false &&
            Array.isArray(hostile.evidence) &&
            hostile.evidence.includes("internal/platform/config/config_otel_test.go::TestOpenTelemetryEnvironmentBindingParser") &&
            hostile.evidence.includes("internal/platform/telemetry/bootstrap_test.go::TestBootstrapIgnoresHostileOTelEnvironment"),
          "OTEL-CORPUS-004 records every hostile environment fixture and closed effective behavior",
          checks,
          "golden.corpus.OTEL-CORPUS-004.hostile_environment_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-005") {
        const hostileConfig = input.hostile_declarative_config_assertions ?? {};
        assert(
          hostileConfig.hazard_matrix === configHazardMatrixPath &&
            JSON.stringify(hostileConfig.component_attempts ?? []) === JSON.stringify(expectedHostileDeclarativeConfigAttempts) &&
            hostileConfig.required_result === "no_externally_observable_telemetry_effect" &&
            hostileConfig.network_export_enabled === false &&
            hostileConfig.per_signal_endpoint_used === false &&
            hostileConfig.resource_detector_invoked === false &&
            hostileConfig.sampler_profile_changed === false &&
            hostileConfig.metric_reader_schedule_changed === false &&
            hostileConfig.unregistered_metric_stream_exported === false &&
            hostileConfig.exemplars_enabled === false &&
            hostileConfig.log_bridge_authority_changed === false &&
            hostileConfig.plugin_provider_loaded === false &&
            Array.isArray(hostileConfig.evidence) &&
            hostileConfig.evidence.includes("contracts/otel/import_boundary.json") &&
            hostileConfig.evidence.includes("internal/platform/config/config_otel_test.go::TestOpenTelemetryEnvironmentBindingParser"),
          "OTEL-CORPUS-005 records every hostile declarative-config attempt and contained result",
          checks,
          "golden.corpus.OTEL-CORPUS-005.hostile_declarative_config_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-006") {
        const httpShape = input.http_route_shape_assertions ?? {};
        assert(
          JSON.stringify(httpShape.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(0, 1)) &&
            JSON.stringify(httpShape.span_names ?? []) === JSON.stringify(["<HTTP_METHOD> <route_template>"]) &&
            JSON.stringify(httpShape.span_kinds ?? []) === JSON.stringify(["server"]) &&
            JSON.stringify(httpShape.metrics ?? []) === JSON.stringify(["cartulary.http.server.request.duration"]) &&
            JSON.stringify(httpShape.required_attributes ?? []) ===
              JSON.stringify(["http.request.method", "http.route", "http.response.status_code", "cartulary.route_family", "cartulary.result"]) &&
            httpShape.route_template_only === true &&
            httpShape.remote_context_accepted === false &&
            ["url.full", "url.path", "url.query", "http.request.header.*", "http.response.header.*", "cookie", "user_agent", "client.ip", "request.body", "response.body", "route_parameter", "stable_id"].every((entry) =>
              (httpShape.forbidden_values_absent ?? []).includes(entry),
            ) &&
            evidenceOK(httpShape.evidence, [
              "internal/platform/telemetry/accessors_test.go::TestHTTPMiddlewareNoSDK",
              "internal/platform/telemetry/privacy_test.go::TestSafeAttributesOmitsForbiddenValueFamiliesBeforeRecording",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-006 asserts HTTP route-template span/metric shape and forbidden request value absence",
          checks,
          "golden.corpus.OTEL-CORPUS-006.http_route_shape_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-007") {
        const workbookSignals = input.workbook_signal_assertions ?? {};
        assert(
          JSON.stringify(workbookSignals.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(1, 4)) &&
            JSON.stringify(workbookSignals.span_names ?? []) ===
              JSON.stringify(["cartulary.workbook.query", "cartulary.workbook.mutation", "cartulary.workbook.projection"]) &&
            JSON.stringify(workbookSignals.span_kinds ?? []) === JSON.stringify(["internal"]) &&
            JSON.stringify(workbookSignals.metrics ?? []) ===
              JSON.stringify(["cartulary.workbook.query.duration", "cartulary.workbook.mutation.duration", "cartulary.workbook.rows.returned"]) &&
            JSON.stringify(workbookSignals.safe_operations ?? []) === JSON.stringify(["query", "create", "patch", "rebuild"]) &&
            JSON.stringify(workbookSignals.row_bucket_fixture_values ?? []) === JSON.stringify(expectedRowBucketFixtureValues) &&
            ["saved_view_id", "filters", "search_text", "row.values", "projection_table_name", "record_id", "row_version", "client_txn_id", "field_values"].every((entry) =>
              (workbookSignals.forbidden_values_absent ?? []).includes(entry),
            ) &&
            evidenceOK(workbookSignals.evidence, [
              "internal/modules/workbook/telemetry_test.go::TestWorkbookTelemetrySafeMappings",
              "internal/modules/workbook/telemetry_test.go::TestWorkbookAPIErrorTelemetry",
              "internal/modules/workbook/telemetry_test.go::TestWorkbookTelemetryNoSDK",
              "internal/modules/projections/internal/runtime/telemetry_test.go::TestProjectionTelemetrySafeVocabulary",
              "internal/modules/projections/internal/runtime/telemetry_test.go::TestProjectionTelemetryNoSDK",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-007 asserts workbook query/mutation/projection signal registry, row bucket fixtures, and safe vocabulary",
          checks,
          "golden.corpus.OTEL-CORPUS-007.workbook_signal_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-008") {
        const websocketSignals = input.websocket_signal_assertions ?? {};
        assert(
          JSON.stringify(websocketSignals.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(4, 6)) &&
            JSON.stringify(websocketSignals.span_names ?? []) ===
              JSON.stringify(["cartulary.collaboration.websocket", "cartulary.collaboration.event_send"]) &&
            JSON.stringify(websocketSignals.span_kinds ?? []) === JSON.stringify(["internal"]) &&
            JSON.stringify(websocketSignals.metrics ?? []) ===
              JSON.stringify(["cartulary.collaboration.connections.active", "cartulary.collaboration.events.sent"]) &&
            JSON.stringify(websocketSignals.safe_lifecycle_operations ?? []) === JSON.stringify(["connect"]) &&
            JSON.stringify(websocketSignals.safe_results ?? []) ===
              JSON.stringify(["success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped"]) &&
            ["record_changed", "extension_resource_changed", "job_progress", "presence_delta", "presence_snapshot", "hello_ack", "resume_ack", "ping", "session_revoked", "error", "other"].every((entry) =>
              (websocketSignals.safe_event_types ?? []).includes(entry),
            ) &&
            ["connection_id", "user_id", "incident_id", "payload", "record_id"].every((entry) => (websocketSignals.forbidden_values_absent ?? []).includes(entry)) &&
            evidenceOK(websocketSignals.evidence, [
              "internal/modules/collaboration/telemetry_test.go::TestWebSocketLifecycleTelemetryClassifiesPublicErrors",
              "internal/modules/collaboration/telemetry_test.go::TestWebSocketLifecycleTelemetryClosesVocabulary",
              "internal/modules/collaboration/hub_telemetry_test.go::TestWebSocketTelemetrySafeVocabulary",
              "internal/modules/collaboration/hub_telemetry_test.go::TestWebSocketEventSendTelemetryNoSDK",
              "internal/modules/collaboration/hub_telemetry_test.go::TestActiveConnectionTelemetryGaugeNoSDK",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-008 asserts WebSocket lifecycle/event-send signal registry and safe vocabulary",
          checks,
          "golden.corpus.OTEL-CORPUS-008.websocket_signal_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-009") {
        const jobSignals = input.job_signal_assertions ?? {};
        assert(
          JSON.stringify(jobSignals.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(6, 8)) &&
            JSON.stringify(jobSignals.span_names ?? []) === JSON.stringify(["cartulary.jobs.enqueue", "cartulary.jobs.run"]) &&
            JSON.stringify(jobSignals.span_kinds ?? []) === JSON.stringify(["internal"]) &&
            JSON.stringify(jobSignals.metrics ?? []) === JSON.stringify([
              "cartulary.jobs.active",
              "cartulary.jobs.duration",
              "cartulary.jobs.attempts",
              "cartulary.jobs.expired",
              "cartulary.jobs.lease_renewal.failures",
            ]) &&
            JSON.stringify(jobSignals.active_job_kinds ?? []) === JSON.stringify(["import.discovery_v1", "reference_pack.refresh_v1", "unknown"]) &&
            JSON.stringify(jobSignals.terminal_results ?? []) === JSON.stringify(["success", "canceled", "failed", "conflict"]) &&
            JSON.stringify(jobSignals.terminal_statuses ?? []) === JSON.stringify(["succeeded", "canceled", "failed"]) &&
            [
              "job_id",
              "incident_id",
              "artifact_path",
              "evidence_id",
              "request.body",
              "attempt_token",
              "progress_unit_id",
              "handler_payload",
              "raw_handler_error",
              "panic_value",
              "filesystem_path",
              "secret",
            ].every((entry) => (jobSignals.forbidden_values_absent ?? []).includes(entry)) &&
            evidenceOK(jobSignals.evidence, [
              "internal/platform/jobs/telemetry_test.go::TestJobTelemetryVocabularyHelpers",
              "internal/platform/jobs/telemetry_test.go::TestSafeJobTelemetryToken",
              "internal/platform/jobs/telemetry_test.go::TestJobTelemetryNoSDK",
              "internal/platform/jobs/telemetry_integration_test.go::TestJobAttemptTelemetryUsesCatalogKindsAndClosedOutcomes_Integration",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-009 asserts background job span/metric registry, active gauge labels, terminal statuses, and safe vocabulary",
          checks,
          "golden.corpus.OTEL-CORPUS-009.job_signal_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-010") {
        const postgresSignals = input.postgres_dependency_assertions ?? {};
        assert(
          JSON.stringify(postgresSignals.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(8, 9)) &&
            JSON.stringify(postgresSignals.span_names ?? []) === JSON.stringify(["cartulary.postgres.operation"]) &&
            JSON.stringify(postgresSignals.span_kinds ?? []) === JSON.stringify(["client"]) &&
            JSON.stringify(postgresSignals.metrics ?? []) === JSON.stringify(["cartulary.postgres.operation.duration"]) &&
            JSON.stringify(postgresSignals.required_attributes ?? []) === JSON.stringify(["db.system.name", "cartulary.operation", "cartulary.result"]) &&
            postgresSignals.db_system_name === "postgresql" &&
            ["timeout", "dependency_unavailable", "serialization_conflict", "constraint_violation", "internal_error"].every((entry) =>
              (postgresSignals.safe_error_classes ?? []).includes(entry),
            ) &&
            ["db.statement", "db.query.text", "db.query.summary", "db.namespace", "db.collection.name", "server.address", "server.port", "projection_table_name", "bind_values"].every((entry) =>
              (postgresSignals.forbidden_values_absent ?? []).includes(entry),
            ) &&
            evidenceOK(postgresSignals.evidence, [
              "internal/platform/postgres/telemetry_test.go::TestTelemetryDBPreservesDBBehaviorNoSDK",
              "internal/platform/postgres/telemetry_test.go::TestPostgresErrorClass",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-010 asserts Postgres dependency signal shape, safe error classes, and forbidden database value absence",
          checks,
          "golden.corpus.OTEL-CORPUS-010.postgres_dependency_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-011") {
        const objectStoreSignals = input.objectstore_dependency_assertions ?? {};
        assert(
          JSON.stringify(objectStoreSignals.span_families ?? []) === JSON.stringify(expectedSpanFamilies.slice(9, 10)) &&
            JSON.stringify(objectStoreSignals.span_names ?? []) === JSON.stringify(["cartulary.objectstore.operation"]) &&
            JSON.stringify(objectStoreSignals.span_kinds ?? []) === JSON.stringify(["client"]) &&
            JSON.stringify(objectStoreSignals.metrics ?? []) ===
              JSON.stringify(["cartulary.objectstore.operation.duration", "cartulary.objectstore.transfer.bytes"]) &&
            JSON.stringify(objectStoreSignals.safe_operations ?? []) ===
              JSON.stringify(["create_upload_target", "complete_upload_target", "put_object", "get_object", "get_object_range", "head_object", "list_prefix", "delete_object", "ensure_bucket_for_dev_test"]) &&
            JSON.stringify(objectStoreSignals.safe_error_classes ?? []) === JSON.stringify(["dependency_unavailable", "timeout", "internal_error"]) &&
            ["bucket", "key", "filename", "object_hash", "upload_id", "copy_source", "storage_ref", "aws.s3.*"].every((entry) =>
              (objectStoreSignals.forbidden_values_absent ?? []).includes(entry),
            ) &&
            evidenceOK(objectStoreSignals.evidence, [
              "internal/platform/objectstore/telemetry_test.go::TestTelemetryStorePreservesStoreBehaviorNoSDK",
              "internal/platform/objectstore/telemetry_test.go::TestObjectStoreErrorClass",
              "internal/platform/telemetry/registry_test.go::TestSpanRegistryClosed",
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
            ]),
          "OTEL-CORPUS-011 asserts object-store dependency signal shape, transfer metric, safe operations, and forbidden object value absence",
          checks,
          "golden.corpus.OTEL-CORPUS-011.objectstore_dependency_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-012") {
        const resourceAssertions = input.resource_identity_assertions ?? {};
        const optionalEnvironment = resourceAssertions.optional_deployment_environment ?? {};
        const detectorSuppression = resourceAssertions.detector_suppression ?? {};
        const conflictingSchema = resourceAssertions.conflicting_schema_url ?? {};
        const profileClaims = resourceAssertions.profile_claims ?? {};
        assert(
          resourceAssertions.schema_url === "" &&
            JSON.stringify(resourceAssertions.allowed_attributes ?? []) === JSON.stringify(expectedResourceAttributes) &&
            optionalEnvironment.omitted_when_null === true &&
            JSON.stringify(optionalEnvironment.allowed_values ?? []) === JSON.stringify(["development", "test", "staging", "production", "custom-token"]) &&
            detectorSuppression.resource_detectors_invoked === false &&
            JSON.stringify(detectorSuppression.forbidden_prefixes ?? []) === JSON.stringify(expectedForbiddenResourcePrefixes) &&
            detectorSuppression.otel_resource_attributes_merged === false &&
            detectorSuppression.otel_service_name_merged === false &&
            detectorSuppression.otel_entities_merged === false &&
            conflictingSchema.non_empty_schema_url_rejected_before_provider_activation === true &&
            profileClaims.base_only === "base" &&
            profileClaims.duplicate_and_sorted === "base,import,incident_portability,reference_pack,snapshot_reporting" &&
            profileClaims.separator === "," &&
            profileClaims.spaces === false &&
            profileClaims.arrays === false &&
            Array.isArray(resourceAssertions.evidence) &&
            resourceAssertions.evidence.includes("internal/platform/telemetry/resource_test.go::TestResourceIdentityClosedRegistry") &&
            resourceAssertions.evidence.includes("internal/platform/telemetry/resource_test.go::TestResourceIdentityOmitsOptionalNullDeploymentEnvironment") &&
            resourceAssertions.evidence.includes("internal/platform/telemetry/resource_test.go::TestExternalResourceContributionRejectsSchemaURLAndDetectorAttributes"),
          "OTEL-CORPUS-012 asserts closed resource attributes, empty schema URL, detector suppression, optional null omission, and profile-claims serialization",
          checks,
          "golden.corpus.OTEL-CORPUS-012.resource_identity_assertions",
        );
        const instanceAssertions = input.instance_id_assertions ?? {};
        assert(
          instanceAssertions.field === "service.instance.id" &&
            instanceAssertions.default_generator_predicate === "canonical_lowercase_non_nil_uuid_v4_per_process_start" &&
            instanceAssertions.normalization_placeholder === "SERVICE_INSTANCE_ID_1" &&
            Array.isArray(instanceAssertions.configured_accepts) &&
            instanceAssertions.configured_accepts.includes("10000000-0000-4000-8000-000000000001") &&
            Array.isArray(instanceAssertions.configured_rejects) &&
            ["arbitrary_string", "nil_uuid", "non_v4_uuid", "uppercase_uuid"].every((entry) => instanceAssertions.configured_rejects.includes(entry)) &&
            typeof instanceAssertions.provenance_invariant === "string" &&
            instanceAssertions.provenance_invariant.includes("structural UUID-v4 predicate"),
          "OTEL-CORPUS-012 asserts service.instance.id UUID-v4 default generation, normalization, configured acceptance/rejection, and provenance invariant",
          checks,
          "golden.corpus.OTEL-CORPUS-012.instance_id_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-013") {
        const nullOmission = input.null_omission_assertions ?? {};
        assert(
          nullOmission.omission_point === "before_otel_api_call" &&
            JSON.stringify(nullOmission.null_like_values ?? []) === JSON.stringify(["", "null", "invalid_attribute_value"]) &&
            JSON.stringify(nullOmission.signal_families ?? []) === JSON.stringify(expectedNullOmissionSignalFamilies) &&
            nullOmission.setter_calls_with_null === 0 &&
            nullOmission.unknown_attributes_dropped === true &&
            Array.isArray(nullOmission.evidence) &&
            nullOmission.evidence.includes("internal/platform/telemetry/privacy_test.go::TestSafeAttributesOmitNullEquivalentAndUnknownCartularyKeys") &&
            nullOmission.evidence.includes("internal/platform/telemetry/privacy_test.go::TestNullLikeValuesOmittedBeforeRecordingForAllSignalFamilies") &&
            nullOmission.evidence.includes("internal/platform/telemetry/resource_test.go::TestResourceIdentityOmitsOptionalNullDeploymentEnvironment"),
          "OTEL-CORPUS-013 asserts null-like values are omitted before OTel API calls across all adopted signal families",
          checks,
          "golden.corpus.OTEL-CORPUS-013.null_omission_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-014") {
        const logBridge = input.log_bridge_assertions ?? {};
        assert(
          logBridge.disabled_exports === false &&
            logBridge.enabled_mapping === true &&
            JSON.stringify(logBridge.severity_mapping ?? {}) === JSON.stringify(expectedLogSeverityMapping) &&
            logBridge.body_bounds?.exact_bound === "truncate_after_redaction_by_unicode_scalar_count" &&
            logBridge.body_bounds?.zero_bound === "empty_body" &&
            logBridge.exception_reduction === "safe_low_cardinality_fields_only" &&
            logBridge.event_name_omitted === true &&
            logBridge.optional_exception_parameter === false &&
            logBridge.span_event_bridge === false &&
            Array.isArray(logBridge.evidence) &&
            logBridge.evidence.includes("internal/platform/telemetry/logs_test.go::TestLogBridgeDisabledDoesNotExport") &&
            logBridge.evidence.includes("internal/platform/telemetry/logs_test.go::TestLogBridgeEnabledMapping") &&
            logBridge.evidence.includes("internal/platform/telemetry/logs_test.go::TestLogBridgeBodyBoundsAndExceptionReduction") &&
            logBridge.evidence.includes("internal/platform/telemetry/logs_test.go::TestLogBridgeOmitsEventNameExceptionParameterAndSpanEvents"),
          "OTEL-CORPUS-014 records disabled/enabled LogRecord bridge mapping, severity, body bounds, exception reduction, EventName omission, and no span-event bridge",
          checks,
          "golden.corpus.OTEL-CORPUS-014.log_bridge_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-015") {
        const metrics = input.metric_registry_assertions ?? {};
        assert(
          JSON.stringify(metrics.metric_names ?? []) === JSON.stringify(expectedMetricNames) &&
            metrics.temporality_profile === "cumulative_current_profile" &&
            JSON.stringify(metrics.duration_buckets ?? []) === JSON.stringify(expectedDurationBuckets) &&
            JSON.stringify(metrics.byte_buckets ?? []) === JSON.stringify(expectedByteBuckets) &&
            JSON.stringify(metrics.row_buckets ?? []) === JSON.stringify(expectedRowBuckets) &&
            JSON.stringify(metrics.row_bucket_fixture_values ?? []) === JSON.stringify(expectedRowBucketFixtureValues) &&
            metrics.duplicate_metric_name_policy === "case_insensitive_reject" &&
            metrics.view_stream_rename === "rejected_or_absent" &&
            metrics.exemplars_enabled === false &&
            metrics.bind_bypass_present === false &&
            metrics.overflow_drop_reason === "metric_overflow" &&
            evidenceOK(metrics.evidence, [
              "internal/platform/telemetry/registry_test.go::TestMetricRegistryClosed",
              "internal/platform/telemetry/exporter_test.go::TestOfferProcessorQueueDropNewOverflow",
              "internal/platform/telemetry/privacy_test.go::TestRedactionDropMetricRecordsOnlyWhenNonRecursive",
            ]),
          "OTEL-CORPUS-015 asserts the closed metric registry, bucket fixtures, temporality, View/exemplar/Bind absence, and overflow behavior",
          checks,
          "golden.corpus.OTEL-CORPUS-015.metric_registry_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-016") {
        const retryMatrix = input.retry_classification_matrix ?? {};
        const expectedHTTPTransient = [429, 502, 503, 504];
        const expectedHTTPPermanent = [400, 401, 403, 404, 422];
        const expectedGRPCTransient = [
          "CANCELLED",
          "DEADLINE_EXCEEDED",
          "RESOURCE_EXHAUSTED_WITH_RETRY_INFO",
          "ABORTED",
          "OUT_OF_RANGE",
          "UNAVAILABLE",
          "DATA_LOSS",
        ];
        const expectedGRPCPermanent = [
          "RESOURCE_EXHAUSTED_WITHOUT_RETRY_INFO",
          "INVALID_ARGUMENT",
          "UNAUTHENTICATED",
          "PERMISSION_DENIED",
          "NOT_FOUND",
          "UNIMPLEMENTED",
        ];
        assert(
          retryMatrix.pinned_exporter_version_source === snapshotPath &&
            JSON.stringify(retryMatrix.http_transient) === JSON.stringify(expectedHTTPTransient) &&
            JSON.stringify(retryMatrix.http_permanent) === JSON.stringify(expectedHTTPPermanent) &&
            JSON.stringify(retryMatrix.grpc_transient) === JSON.stringify(expectedGRPCTransient) &&
            JSON.stringify(retryMatrix.grpc_permanent) === JSON.stringify(expectedGRPCPermanent) &&
            retryMatrix.retry_envelope === "full_jitter_bounds_only" &&
            retryMatrix.permanent_drop_reason === "exporter_permanent_discard",
          "OTEL-CORPUS-016 records the pinned retry classification matrix and permanent-discard drop reason",
          checks,
          "golden.corpus.OTEL-CORPUS-016.retry_classification_matrix",
        );
        const exporterRuntime = input.exporter_runtime_assertions ?? {};
        const httpPaths = exporterRuntime.otlp_http_paths ?? {};
        const grpc = exporterRuntime.otlp_grpc ?? {};
        const headerPolicy = exporterRuntime.header_policy ?? {};
        const userAgent = exporterRuntime.user_agent ?? {};
        const timeout = exporterRuntime.timeout ?? {};
        const queueOverflow = exporterRuntime.queue_overflow ?? {};
        const queueDropMetric = queueOverflow.drop_metric ?? {};
        const exportFailureMetric = exporterRuntime.export_failure_metric ?? {};
        const shutdown = exporterRuntime.shutdown ?? {};
        const shutdownDropMetric = shutdown.drop_metric ?? {};
        const selfDiagnostics = exporterRuntime.self_diagnostics ?? {};
        const recursionDropMetric = selfDiagnostics.drop_metric_when_possible ?? {};
        assert(
          exporterRuntime.export_disabled_default === true &&
            httpPaths.traces === "/v1/traces" &&
            httpPaths.metrics === "/v1/metrics" &&
            httpPaths.logs === "/v1/logs" &&
            httpPaths.per_signal_endpoint_divergence === false &&
            grpc.one_configured_target === true &&
            grpc.per_signal_endpoint_divergence === false &&
            grpc.https_selects_tls === true &&
            grpc.http_selects_insecure === true &&
            headerPolicy.configured_source === "telemetry.exporter.headers.secret_ref_v1" &&
            JSON.stringify(headerPolicy.protocol_required_headers ?? []) === JSON.stringify(["content-type", "user-agent"]) &&
            headerPolicy.configured_values_redacted_in_diagnostics === true &&
            headerPolicy.secret_values_absent_from_retained_artifacts === true &&
            headerPolicy.otel_environment_headers_ignored === true &&
            userAgent.grammar === "Cartulary/<SERVICE_VERSION> OTel-OTLP-Exporter-go/<EXPORTER_VERSION>" &&
            userAgent.forbidden_extra_segments === false &&
            userAgent.forbidden_values === false &&
            userAgent.pinned_exporter_version_source === snapshotPath &&
            timeout.attempt_timeout_key === "telemetry.processor.export_timeout_ms" &&
            timeout.transient_only_when_transport_transient === true &&
            timeout.product_hot_path_blocking === false &&
            queueOverflow.policy === "drop_new" &&
            queueOverflow.retains_queued_items === true &&
            queueOverflow.drops_new_item === true &&
            queueOverflow.product_hot_path_blocking === false &&
            queueOverflow.queue_depth_metric === "cartulary.telemetry.queue.depth" &&
            queueDropMetric.name === "cartulary.telemetry.item.dropped" &&
            queueDropMetric.drop_reason === "queue_full" &&
            queueDropMetric.records_when_non_recursive === true &&
            queueDropMetric.records_when_recursive === false &&
            exportFailureMetric.name === "cartulary.telemetry.export.failure" &&
            JSON.stringify(exportFailureMetric.attributes ?? []) ===
              JSON.stringify(["cartulary.signal_kind", "cartulary.telemetry.exporter_kind", "cartulary.error_class"]) &&
            exportFailureMetric.records_when_non_recursive === true &&
            exportFailureMetric.records_when_recursive === false &&
            shutdown.flush_timeout_key === "telemetry.shutdown.flush_timeout_ms" &&
            shutdown.continues_after_timeout === true &&
            shutdown.shutdown_calls_per_active_provider === 1 &&
            shutdown.repeated_shutdown_idempotent === true &&
            shutdownDropMetric.name === "cartulary.telemetry.item.dropped" &&
            shutdownDropMetric.drop_reason === "shutdown_timeout" &&
            shutdownDropMetric.records_when_non_recursive === true &&
            shutdownDropMetric.records_when_recursive === false &&
            selfDiagnostics.bounded === true &&
            selfDiagnostics.recursive_item_dropped === true &&
            recursionDropMetric.name === "cartulary.telemetry.item.dropped" &&
            recursionDropMetric.drop_reason === "recursion_guard" &&
            Array.isArray(exporterRuntime.evidence) &&
            exporterRuntime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestBuildExporterRequestHeadersRedactsConfiguredSecrets") &&
            exporterRuntime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestPlanExporterAttemptTimeoutClassification") &&
            exporterRuntime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestOfferProcessorQueueDropNewOverflow") &&
            exporterRuntime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestPlanShutdownTimeoutAndIdempotence") &&
            exporterRuntime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestPlanSelfDiagnosticRecursionGuard"),
          "OTEL-CORPUS-016 records exporter endpoint, header, User-Agent, retry, timeout, queue overflow, shutdown, and self-diagnostic runtime assertions",
          checks,
          "golden.corpus.OTEL-CORPUS-016.exporter_runtime_assertions",
        );
      }
      if (caseID === "OTEL-CORPUS-017") {
        const runtime = input.runtime_invariance_matrix ?? {};
        assert(
          JSON.stringify(runtime.surfaces ?? []) === JSON.stringify(expectedRuntimeInvarianceSurfaces) &&
            JSON.stringify(runtime.failure_modes ?? []) === JSON.stringify(expectedRuntimeInvarianceFailures) &&
            runtime.scenario_count === expectedRuntimeInvarianceSurfaces.length * expectedRuntimeInvarianceFailures.length &&
            runtime.product_response === "match_no_export_baseline" &&
            runtime.committed_state === "match_no_export_baseline" &&
            runtime.product_hot_path_blocking === false &&
            JSON.stringify(runtime.startup_failures_excluded ?? []) === JSON.stringify(["invalid_telemetry_config", "sdk_provider_construction_failure"]) &&
            runtime.allowed_telemetry_effect === "bounded_local_diagnostics_and_self_metrics_only" &&
            Array.isArray(runtime.evidence) &&
            runtime.evidence.includes("internal/platform/telemetry/exporter_test.go::TestRuntimeInvarianceMatrixMatchesNoExportBaseline"),
          "OTEL-CORPUS-017 records every product surface and runtime telemetry failure mode with no-export-baseline behavior",
          checks,
          "golden.corpus.OTEL-CORPUS-017.runtime_invariance_matrix",
        );
      }
      if (caseID === "OTEL-CORPUS-018") {
        const actionMatrix = input.forbidden_value_action_matrix ?? [];
        const actionMatrixOK =
          Array.isArray(actionMatrix) &&
          actionMatrix.length === expectedForbiddenValueActionMatrix.length &&
          actionMatrix.every((row, index) => {
            const expected = expectedForbiddenValueActionMatrix[index];
            const dropMetric = row.drop_metric ?? {};
            return (
              row.family === expected.family &&
              row.owner_literal_fixture === expected.owner_literal_fixture &&
              row.default_treatment === expected.default_treatment &&
              row.replacement_allowed === expected.replacement_allowed &&
              row.diagnostic_family === expected.diagnostic_family &&
              dropMetric.name === "cartulary.telemetry.item.dropped" &&
              dropMetric.drop_reason === "redaction_rejected" &&
              dropMetric.records_when_non_recursive === true &&
              dropMetric.records_when_recursive === false
            );
          });
        assert(
          actionMatrixOK,
          "OTEL-CORPUS-018 records deterministic action and non-recursive drop-metric behavior for every forbidden-value family",
          checks,
          "golden.corpus.OTEL-CORPUS-018.forbidden_value_action_matrix",
        );
      }
    }

    for (const [field, schemaID] of Object.entries(normalizedSignalSchemas)) {
      const expectedSignalPath = `cases/${caseID}/${field}.json`;
      assert(
        row[field] === expectedSignalPath,
        `${caseID} ${field} path is canonical`,
        checks,
        `golden.corpus.${caseID}.${field}_path`,
      );
      const signalPath = `${manifest.normalized_golden_root}/${expectedSignalPath}`;
      if (assert(fileExists(signalPath), `${caseID} ${field} exists`, checks, `golden.corpus.${caseID}.${field}_exists`)) {
        const signal = readJSON(signalPath);
        assert(signal.schema_id === schemaID, `${caseID} ${field} schema is adopted`, checks, `golden.corpus.${caseID}.${field}_schema`);
        assert(signal.case_id === caseID, `${caseID} ${field} is bound to the case ID`, checks, `golden.corpus.${caseID}.${field}_case_id`);
        const payloadField = field === "normalized_traces" ? "spans" : field === "normalized_metrics" ? "metrics" : "logs";
        assert(Array.isArray(signal[payloadField]), `${caseID} ${field} payload is an array`, checks, `golden.corpus.${caseID}.${field}_payload`);
        if (caseID === "OTEL-CORPUS-001") {
          assert(
            signal[payloadField].length === 0,
            `OTEL-CORPUS-001 ${field} remains empty under no-SDK mode`,
            checks,
            `golden.corpus.OTEL-CORPUS-001.${field}_no_sdk_empty`,
          );
        }
      }
    }
  }

  emitAndCompareCorpusCaptures(manifest, cases, checks);

  assert(
    classification.schema_id === "cartulary.otel_dependency_update_classification.v2",
    "dependency-update classification manifest uses the adopted schema",
    checks,
    "golden.dependency.schema_id",
  );
  assert(
    classification.corpus_manifest === corpusManifestPath,
    "dependency-update classification is bound to the golden corpus manifest",
    checks,
    "golden.dependency.corpus_manifest",
  );
  const expectedClasses = new Map([
    ["registry_equivalent", false],
    ["additive_non_breaking", true],
    ["privacy_tightening", true],
    ["breaking_shape_change", true],
  ]);
  const classRows = Array.isArray(classification.change_classes) ? classification.change_classes : [];
  const classNames = classRows.map((row) => row.class);
  assert(
    JSON.stringify(classNames) === JSON.stringify([...expectedClasses.keys()]),
    `dependency-update classes exactly match the adopted registry${JSON.stringify(classNames) !== JSON.stringify([...expectedClasses.keys()]) ? `; saw ${classNames.join(", ")}` : ""}`,
    checks,
    "golden.dependency.classes",
  );
  for (const row of classRows) {
    if (!expectedClasses.has(row.class)) {
      continue;
    }
    assert(
      row.nlspec_revision_required === expectedClasses.get(row.class),
      `${row.class} NLSpec-revision requirement is correct`,
      checks,
      `golden.dependency.${row.class}.revision_requirement`,
    );
    assert(
      typeof row.merge_behavior === "string" && row.merge_behavior.length > 0,
      `${row.class} merge behavior is recorded`,
      checks,
      `golden.dependency.${row.class}.merge_behavior`,
    );
  }
  assert(
    classification.latest_update?.classification === "registry_equivalent" &&
      classification.latest_update?.corpus_manifest === corpusManifestPath &&
      classification.latest_update?.comparison_schema === "cartulary.otel_corpus_comparison.v1" &&
      classification.latest_update?.nlspec_revision_required === false,
    "latest dependency-update classification records a registry-equivalent corpus comparison",
    checks,
    "golden.dependency.latest_classification",
  );
}

function validateNoRepoAdoptionTODOs(checks) {
  const visibleFiles = [
    snapshotPath,
    generatedConstantsManifestPath,
    importBoundaryPath,
    errorClassRegistryPath,
    telemetryConfigSchemaPath,
    configHazardMatrixPath,
    "contracts/otel/semantic_conventions_constants.v1.json",
    corpusManifestPath,
    dependencyClassificationPath,
  ];
  const hits = [];
  const goldenJSONFiles = walkFilesIncludingBuild("internal/testutil/golden/otel", (relative) => relative.endsWith(".json"));
  for (const file of [...new Set([...visibleFiles, ...goldenJSONFiles])]) {
    const text = readText(file);
    if (text.includes("TODO(repo-adoption)")) {
      hits.push(file);
    }
  }
  assert(hits.length === 0, `conformance-visible artifacts contain no TODO(repo-adoption) placeholders${hits.length ? `; found ${hits.join(", ")}` : ""}`, checks, "repo_adoption_todos.none_visible");
}

function validateErrorMapping(checks) {
  const publicCodes = publicErrorCodes();
  const mappedCounts = new Map();
  for (const row of readJSON(errorClassRegistryPath).public_error_classes ?? []) {
    for (const code of row.error_codes ?? []) {
      mappedCounts.set(code, (mappedCounts.get(code) ?? 0) + 1);
    }
  }
  const missing = [...publicCodes].filter((code) => !mappedCounts.has(code));
  const duplicated = [...mappedCounts.entries()].filter(([, count]) => count !== 1).map(([code]) => code);
  const unknown = [...mappedCounts.keys()].filter((code) => !publicCodes.has(code));
  assert(missing.length === 0, `all machine-owned public error codes are mapped${missing.length ? `; missing ${missing.join(", ")}` : ""}`, checks, "errors.mapping_complete");
  assert(duplicated.length === 0, `no public error code is mapped more than once${duplicated.length ? `; duplicated ${duplicated.join(", ")}` : ""}`, checks, "errors.mapping_unique");
  assert(unknown.length === 0, `mapping contains no unknown public error code${unknown.length ? `; unknown ${unknown.join(", ")}` : ""}`, checks, "errors.mapping_known");
}

function validateErrorClassRegistry(registry, checks) {
  assert(registry.schema_id === "cartulary.otel_error_class_registry.v1", "error-class registry uses the adopted schema", checks, "errors.registry_schema_id");
  const publicCodes = publicErrorCodes();
  const mappedCounts = new Map();
  const classNames = new Set();
  for (const row of registry.public_error_classes ?? []) {
    classNames.add(row.error_class);
    for (const code of row.error_codes ?? []) {
      mappedCounts.set(code, (mappedCounts.get(code) ?? 0) + 1);
    }
  }
  for (const row of registry.additional_error_classes ?? []) {
    classNames.add(row.error_class);
  }
  const missing = [...publicCodes].filter((code) => !mappedCounts.has(code));
  const duplicated = [...mappedCounts.entries()].filter(([, count]) => count !== 1).map(([code]) => code);
  const unknown = [...mappedCounts.keys()].filter((code) => !publicCodes.has(code));
  assert(missing.length === 0, `error-class registry maps every Core 01 public error code${missing.length ? `; missing ${missing.join(", ")}` : ""}`, checks, "errors.registry_mapping_complete");
  assert(duplicated.length === 0, `error-class registry maps every public error code once${duplicated.length ? `; duplicated ${duplicated.join(", ")}` : ""}`, checks, "errors.registry_mapping_unique");
  assert(unknown.length === 0, `error-class registry contains no unknown public error code${unknown.length ? `; unknown ${unknown.join(", ")}` : ""}`, checks, "errors.registry_mapping_known");
  assert(classNames.has("internal_error") && classNames.has("timeout") && classNames.has("dependency_unavailable"), "error-class registry includes required runtime/dependency classes", checks, "errors.registry_runtime_classes");
}

function targetDir() {
  const resultsRoot = process.env.CARTULARY_TEST_RESULTS_DIR || path.join(repoRoot, ".cartulary", "test-results");
  const runId = process.env.CARTULARY_TEST_RUN_ID || "manual";
  return path.join(resultsRoot, runId, "otel-conformance");
}

function writeSummary(checks, status) {
  const outDir = targetDir();
  mkdirSync(outDir, { recursive: true, mode: 0o700 });
  const summary = {
    schema_id: "cartulary.otel_conformance_summary.v1",
    target: "otel-conformance",
    status,
    checked_at: new Date().toISOString(),
    source_snapshot_path: snapshotPath,
    checks,
  };
  writeFileSync(path.join(outDir, "otel-conformance-summary.json"), `${JSON.stringify(summary, null, 2)}\n`);
}

function validateHarnessTelemetryBoundary(checks) {
  const profileFile = "tools/harness/observability/observability.mjs";
  const exportFile = "tools/harness/observability/otel-export-cli.mjs";
  const profile = readText(profileFile);
  const exporter = readText(exportFile);
  const rootPackage = readJSON("package.json");
  const dependencies = {
    ...(rootPackage.dependencies ?? {}),
    ...(rootPackage.devDependencies ?? {}),
  };
  assert(
    profile.includes('export const harnessScope = "cartulary.harness.execution"') &&
      !profile.includes("OTEL_") &&
      !profile.includes("internal/platform/telemetry") &&
      !profile.includes("fetch("),
    "harness diagnostics use only their delegated scope, ignore inherited OTEL variables, do not import application telemetry, and cannot export during local reconstruction",
    checks,
    "import_boundary.harness_profile",
  );
  assert(
    Object.keys(dependencies).every((name) => !name.startsWith("@opentelemetry/")) &&
      exporter.includes("HARNESS_OTLP_ENDPOINT") &&
      exporter.includes('redirect: "error"'),
    "harness diagnostics add no JavaScript OTel SDK and network delivery remains explicit with redirects disabled",
    checks,
    "import_boundary.harness_dependencies_and_export",
  );
  const applicationRoots = ["internal", "apps", "packages"];
  const scopeViolations = [];
  for (const root of applicationRoots) {
    const files = [];
    walkFiles(root, () => true, files);
    for (const file of files) {
      if (
        [".go", ".ts", ".tsx", ".js", ".mjs"].some((suffix) => file.endsWith(suffix)) &&
        readText(file).includes("cartulary.harness.execution")
      ) {
        scopeViolations.push(file);
      }
    }
  }
  assert(
    scopeViolations.length === 0,
    `application sources cannot emit the harness scope${scopeViolations.length ? `; violations ${scopeViolations.join(", ")}` : ""}`,
    checks,
    "import_boundary.application_harness_scope",
  );
}

function validateRuntimeBehavior(checks) {
  const resultsRoot = process.env.CARTULARY_TEST_RESULTS_DIR;
  const runID = process.env.CARTULARY_TEST_RUN_ID;
  const manifest = readJSON("tools/test_families/platform.telemetry.json");
  const activeRows = manifest.rows.filter((row) => row.status === "active");
  const runRoot = resultsRoot && runID
    ? path.resolve(repoRoot, resultsRoot, runID)
    : "";
  const failures = [];
  for (const row of activeRows) {
    const resultFile = path.join(runRoot, "rows", `${row.row_id}.json`);
    if (!runRoot || !absoluteFileExists(resultFile)) {
      failures.push(`${row.row_id}:missing`);
      continue;
    }
    const result = JSON.parse(readFileSync(resultFile, "utf8"));
    if (result.schema_id !== "cartulary.harness_row_result.v2" || result.terminal_state !== "passed") {
      failures.push(`${row.row_id}:${result.terminal_state ?? "invalid"}`);
    }
  }
  assert(
    activeRows.length > 0 && failures.length === 0,
    `current platform.telemetry runtime, privacy, exporter-failure, retry, queue, and shutdown canonical row evidence passes${failures.length === 0 ? "" : `; ${failures.join(", ")}`}`,
    checks,
    "runtime.current_behavior",
  );
}

function main() {
  const checks = [];
  try {
    validateRuntimeBehavior(checks);
    validateSnapshot(readJSON(snapshotPath), checks);
    validateGeneratedConstantsManifest(readJSON(generatedConstantsManifestPath), checks);
    validateTelemetryConfigSchema(readJSON(telemetryConfigSchemaPath), checks);
    validateConfigHazardMatrix(readJSON(configHazardMatrixPath), checks);
    validateImportBoundary(readJSON(importBoundaryPath), checks);
    validateHarnessTelemetryBoundary(checks);
    validateVerificationOwner(checks);
    validateGoldenCorpus(readJSON(corpusManifestPath), readJSON(dependencyClassificationPath), checks);
    validateNoRepoAdoptionTODOs(checks);
    validateErrorMapping(checks);
    validateErrorClassRegistry(readJSON(errorClassRegistryPath), checks);
  } catch (error) {
    checks.push({ id: "otel_conformance.exception", status: "fail", message: error.message });
  }

  const status = checks.every((check) => check.status === "pass") ? "pass" : "fail";
  writeSummary(checks, status);
  if (status !== "pass") {
    for (const check of checks.filter((entry) => entry.status !== "pass")) {
      process.stderr.write(`${check.id}: ${check.message}\n`);
    }
    process.exit(1);
  }
}

main();
