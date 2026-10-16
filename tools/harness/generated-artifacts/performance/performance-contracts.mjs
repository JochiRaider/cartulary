import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";

import {
  parseStrictJSON,
  validateSchemaSync,
} from "../../contract/index.mjs";
import {
  canonicalSnapshotKeyInput,
  loadPerformanceFixtureSnapshotRegistry,
  snapshotKey,
  snapshotKeyEnvelope,
} from "../../performance-fixture/index.mjs";

const performanceSchemaID = "cartulary.performance_interaction_contract.v2";
const policySchemaID = "cartulary.measurement_policy.v1";
const generatorID = "cartulary.performance_contract_generation.v1";

function compareUnicodeCodePoints(left, right) {
  const leftCodePoints = Array.from(left, (value) => value.codePointAt(0));
  const rightCodePoints = Array.from(right, (value) => value.codePointAt(0));
  const sharedLength = Math.min(leftCodePoints.length, rightCodePoints.length);
  for (let index = 0; index < sharedLength; index += 1) {
    const difference = leftCodePoints[index] - rightCodePoints[index];
    if (difference !== 0) return difference;
  }
  return leftCodePoints.length - rightCodePoints.length;
}

function loadJSON(file, schemaID) {
  const bytes = readFileSync(file);
  const value = parseStrictJSON(
    new TextDecoder("utf-8", { fatal: true }).decode(bytes),
    file,
  );
  validateSchemaSync(schemaID, value);
  return { bytes, value };
}

export function validateViewChangeCapabilities(contract, viewSchema) {
  const fields = new Map(viewSchema.fields.map(field => [field.field_key, field]));
  const {sort, filter, group} = contract.workloads;
  if (contract.view_schema_id !== viewSchema.view_schema_id ||
    !fields.get(sort.field_key)?.sortable ||
    !fields.get(filter.field_key)?.filter_ops.includes(filter.operator) ||
    !fields.get(group.field_key)?.groupable ||
    !fields.get(group.field_key)?.enum_values?.includes(group.value)) {
    throw new Error("view-change workload exceeds its admitted view capabilities");
  }
}

export function loadPerformanceContracts({ performancePath, policyPath, pastePath, fixturePath, viewChangePath, viewSchemaPath, presencePath, historyPath, anchorPath, importDiscoveryPath, importApplyPath, reportingProgressPath, evidenceVerificationPath, evidenceInspectorPath, referencePackProgressPath }) {
  const evidenceVerification = loadJSON(evidenceVerificationPath, "cartulary.performance_evidence_verification_contract.v1");
  const evidenceInspector = loadJSON(evidenceInspectorPath, "cartulary.performance_evidence_inspector_contract.v1");
  if (evidenceVerification.value.fixture_contract_id !== evidenceInspector.value.contract_id || evidenceVerification.value.workload.object_bytes !== evidenceInspector.value.fixture.object_bytes || evidenceVerification.value.workload.preview_preflight_records !== evidenceInspector.value.fixture.linked_evidence_count) throw new Error("Evidence verification fixture binding differs from its owner");
  const referencePackProgress = loadJSON(referencePackProgressPath, "cartulary.performance_reference_pack_progress_contract.v1");
  const reportingProgress = loadJSON(reportingProgressPath, "cartulary.performance_reporting_progress_contract.v1");
  const importDiscovery = loadJSON(importDiscoveryPath, "cartulary.performance_import_discovery_contract.v1");
  const importApply = loadJSON(importApplyPath, "cartulary.performance_import_apply_contract.v1");
  const performance = loadJSON(performancePath, performanceSchemaID);
  const policy = loadJSON(policyPath, policySchemaID);
  const presence = loadJSON(presencePath, "cartulary.performance_presence_contract.v2");
  const anchor = loadJSON(anchorPath, "cartulary.performance_anchor_trace_contract.v1");
  const history = loadJSON(historyPath, "cartulary.performance_history_contract.v2");
  const paste = loadJSON(pastePath, "cartulary.performance_paste_contract.v1");
  const viewSchema = loadJSON(viewSchemaPath, "cartulary.view_schema_source.v1");
  const viewChange = loadJSON(viewChangePath, "cartulary.performance_view_change_contract.v1");
  validateViewChangeCapabilities(viewChange.value, viewSchema.value);
  const fixture = loadJSON(fixturePath, "cartulary.performance_large_grid_contract.v1");
  if (referencePackProgress.value.fixture_contract_id !== fixture.value.contract_id) throw new Error("Reference Pack progress fixture reference differs from the shared owner");
  if (reportingProgress.value.fixture_contract_id !== fixture.value.contract_id) throw new Error("Reporting progress fixture reference differs from the shared owner");
  if (importApply.value.fixture_contract_id !== fixture.value.contract_id) {
    throw new Error("Import apply fixture reference does not match the shared large-grid owner");
  }
  if (importDiscovery.value.fixture_contract_id !== fixture.value.contract_id || anchor.value.fixture_contract_id !== fixture.value.contract_id || history.value.fixture_contract_id !== fixture.value.contract_id || presence.value.fixture_contract_id !== fixture.value.contract_id || viewChange.value.fixture_contract_id !== fixture.value.contract_id || viewChange.value.measurement_policy_id !== policy.value.policy_id || paste.value.fixture_contract_id !== fixture.value.contract_id || performance.value.fixture_contract_id !== fixture.value.contract_id) {
    throw new Error("performance fixture reference does not match the shared large-grid owner");
  }
  if (performance.value.measurement_policy_id !== policy.value.policy_id) {
    throw new Error(
      `${performancePath}.measurement_policy_id does not match ${policyPath}.policy_id`,
    );
  }
  const predicateIDs = performance.value.predicates.map(
    (predicate) => predicate.predicate_id,
  );
  if (new Set(predicateIDs).size !== predicateIDs.length) {
    throw new Error(`${performancePath}.predicates contains duplicate IDs`);
  }
  return {
    inputSha256: createHash("sha256")
      .update(performance.bytes)
      .update(policy.bytes)
      .update(paste.bytes)
      .update(fixture.bytes)
      .update(viewChange.bytes)
      .update(viewSchema.bytes)
      .update(presence.bytes)
      .update(history.bytes)
      .update(anchor.bytes)
      .update(importDiscovery.bytes)
      .update(importApply.bytes)
      .update(referencePackProgress.bytes)
      .update(reportingProgress.bytes)
      .update(evidenceVerification.bytes)
      .update(evidenceInspector.bytes)
      .digest("hex"),
    performance: performance.value,
    viewChange: viewChange.value,
    presence: presence.value,
    history: history.value,
    anchor: anchor.value,
    importDiscovery: importDiscovery.value,
    importApply: importApply.value,
    referencePackProgress: referencePackProgress.value,
    reportingProgress: reportingProgress.value,
    evidenceVerification: evidenceVerification.value,
    paste: paste.value,
    fixture: fixture.value,
    policy: policy.value,
  };
}

export function renderPerformanceContractsTypeScript(document) {
  const fixture = document.fixture.fixture;
  const sharedFixture = {
    fixtureId: fixture.fixture_id,
    seed: fixture.seed,
    timelineRows: fixture.timeline_rows,
    hostRows: fixture.host_rows,
    identityRows: fixture.identity_rows,
    tagAssignments: fixture.tag_assignments,
    mentionAssignments: fixture.mention_assignments,
    linkAssignments: fixture.link_assignments,
    analystSessions: fixture.analyst_sessions,
    backgroundSessions: fixture.background_sessions,
    backgroundUpdateIntervalMs: fixture.background_update_interval_ms,
    backgroundUpdatesPerSecond: fixture.background_updates_per_second,
    trafficTraceId: fixture.traffic_trace_id,
  };
  const performance = {
    contractId: document.performance.contract_id,
    criterionId: document.performance.criterion_id,
    measurementPolicyId: document.performance.measurement_policy_id,
    fixtureContractId: document.performance.fixture_contract_id,
    predicates: document.performance.predicates.map((predicate) => ({
      predicateId: predicate.predicate_id,
      viewSchemaId: predicate.view_schema_id,
      fieldKey: predicate.field_key,
      initiatingKey: predicate.initiating_key,
      startMark: predicate.start_mark,
      stopMark: predicate.stop_mark,
      thresholdMs: predicate.threshold_ms,
      startState: predicate.start_state,
      stopState: predicate.stop_state,
    })),
  };
  const policy = {
    policyId: document.policy.policy_id,
    warmupPasses: document.policy.warmup_passes,
    measuredSamples: document.policy.measured_samples,
    totalSamples:
      document.policy.warmup_passes + document.policy.measured_samples,
    percentile: document.policy.percentile,
    estimator: document.policy.estimator,
    finiteNonNegativeSamples: document.policy.finite_non_negative_samples,
    productRetries: document.policy.product_retries,
  };
  return [
    `// Code generated by ${generatorID}; DO NOT EDIT.`,
    `// Input SHA-256: ${document.inputSha256}`,
    "",
    `export const cartularyAc043PerformanceContract = ${JSON.stringify(performance, null, 2)} as const;`,
    `export const cartularyLargeGridFixture = ${JSON.stringify(sharedFixture, null, 2)} as const;`,
    "",
    `export const cartularyInteractiveP95MeasurementPolicy = ${JSON.stringify(policy, null, 2)} as const;`,
    "",
    `export const cartularyAc003PerformanceContract = ${JSON.stringify(document.paste, null, 2)} as const;`,
    "",
    `export const cartularyPresencePerformanceContract = ${JSON.stringify(document.presence, null, 2)} as const;`,
    'export type CartularyPresencePredicateId = (typeof cartularyPresencePerformanceContract.predicates)[number]["predicate_id"];',
    `export const cartularyAnchorTraceContract = ${JSON.stringify(document.anchor, null, 2)} as const;`,
    `export const cartularyImportDiscoveryPerformanceContract = ${JSON.stringify(document.importDiscovery, null, 2)} as const;`,
    'export type CartularyImportDiscoveryPredicateId = (typeof cartularyImportDiscoveryPerformanceContract.predicates)[number]["predicate_id"];',
    `export const cartularyEvidenceVerificationContract = ${JSON.stringify(document.evidenceVerification, null, 2)} as const;`,
    `export const cartularyReferencePackProgressContract = ${JSON.stringify(document.referencePackProgress, null, 2)} as const;`,
    'export type CartularyReferencePackProgressPredicateId = (typeof cartularyReferencePackProgressContract.predicates)[number]["predicate_id"];',
    `export const cartularyReportingProgressContract = ${JSON.stringify(document.reportingProgress, null, 2)} as const;`,
    'export type CartularyReportingProgressPredicateId = (typeof cartularyReportingProgressContract.predicates)[number]["predicate_id"];',
    `export const cartularyImportApplyPerformanceContract = ${JSON.stringify(document.importApply, null, 2)} as const;`,
    'export type CartularyImportApplyPredicateId = (typeof cartularyImportApplyPerformanceContract.predicates)[number]["predicate_id"];',
    `export const cartularyHistoryPerformanceContract = ${JSON.stringify(document.history, null, 2)} as const;`,
    'export type CartularyHistoryPredicateId = (typeof cartularyHistoryPerformanceContract.predicates)[number]["predicate_id"];',
    `export const cartularyAc044PerformanceContract = ${JSON.stringify(document.viewChange, null, 2)} as const;`,
    "export type CartularyAc044PredicateId = (typeof cartularyAc044PerformanceContract.predicates)[number][\"predicate_id\"];",
    "",
    "export type CartularyAc043Predicate = (typeof cartularyAc043PerformanceContract.predicates)[number];",
    "export type CartularyAc043PredicateId = CartularyAc043Predicate[\"predicateId\"];",
    "",
  ].join("\n");
}

export function loadPerformanceFixtureProfiles(root) {
  const loaded = loadPerformanceFixtureSnapshotRegistry(root);
  return orderedFixtureProfiles(loaded.profiles.values());
}

function orderedFixtureProfiles(profiles) {
  return [...profiles].sort((left, right) =>
    compareUnicodeCodePoints(left.fixture_profile_id, right.fixture_profile_id),
  );
}

function descriptorProjection(profile) {
  return {
    fixture_profile_id: profile.fixture_profile_id,
    status: profile.status,
    fixture_version: profile.fixture_version,
    seed: profile.seed,
    source_contract_digest: profile.source_contract_digest,
    verification_bindings: profile.verification_bindings,
    source_contract_refs: profile.source_contract_refs.map((reference) => ({
      schema_id: reference.schema_id,
      contract_id: reference.contract_id,
    })),
    contributions: profile.contributions,
    semantic_expectations: profile.semantic_expectations,
    runtime_credential_sets: profile.runtime_credential_sets,
    object_population: profile.object_population,
    artifact_policy: profile.artifact_policy,
    redaction_policy: profile.redaction_policy,
  };
}

export function renderPerformanceFixtureProfilesGo(profiles) {
  const catalog = orderedFixtureProfiles(profiles).map(descriptorProjection);
  return [
    `// Code generated by ${generatorID}; DO NOT EDIT.`,
    "",
    "package performancefixtureprofile",
    "",
    'import "encoding/json"',
    "",
    "type CountExpectation struct {",
    '\tCountID string `json:"count_id"`',
    '\tExact int `json:"exact"`',
    "}",
    "",
    "type SemanticCountExpectation struct {",
    '\tExpectationID string `json:"expectation_id"`',
    '\tExact int `json:"exact"`',
    "}",
    "",
    "type SemanticConditionExpectation struct {",
    '\tExpectationID string `json:"expectation_id"`',
    '\tRequired bool `json:"required"`',
    "}",
    "",
    "type VerificationBinding struct {",
    '\tVerificationID string `json:"verification_id"`',
    '\tPredicateID string `json:"predicate_id"`',
    "}",
    "",
    "type SourceContractRef struct {",
    '\tSchemaID string `json:"schema_id"`',
    '\tContractID string `json:"contract_id"`',
    "}",
    "",
    "type Contribution struct {",
    '\tContributionID string `json:"contribution_id"`',
    '\tOwnerID string `json:"owner_id"`',
    '\tVersion string `json:"version"`',
    '\tDependencies []string `json:"dependencies"`',
    '\tSourceContractIDs []string `json:"source_contract_ids"`',
    '\tExpectedReceiptCounts []CountExpectation `json:"expected_receipt_counts"`',
    "}",
    "",
    "type SemanticExpectations struct {",
    '\tCounts []SemanticCountExpectation `json:"counts"`',
    '\tConditions []SemanticConditionExpectation `json:"conditions"`',
    "}",
    "",
    "type RuntimeCredentialSet struct {",
    '\tSetID string `json:"set_id"`',
    '\tCredentialKind string `json:"credential_kind"`',
    '\tAccountCount int `json:"account_count"`',
    "}",
    "",
    "type ObjectPopulation struct {",
    '\tObjectCount int `json:"object_count"`',
    '\tTotalBytes int64 `json:"total_bytes"`',
    "}",
    "",
    "type ArtifactPolicy struct {",
    '\tSnapshotKeySchemaID string `json:"snapshot_key_schema_id"`',
    '\tBuildSchemaID string `json:"build_schema_id"`',
    '\tLeaseSchemaID string `json:"lease_schema_id"`',
    '\tRuntimeSchemaID string `json:"runtime_schema_id"`',
    '\tObservationSchemaID string `json:"observation_schema_id"`',
    '\tSummarySchemaID string `json:"summary_schema_id"`',
    '\tAggregateSchemaID string `json:"aggregate_schema_id"`',
    "}",
    "",
    "type RedactionPolicy struct {",
    '\tPolicyID string `json:"policy_id"`',
    '\tForbiddenFields []string `json:"forbidden_fields"`',
    "}",
    "",
    "type Profile struct {",
    '\tFixtureProfileID string `json:"fixture_profile_id"`',
    '\tStatus string `json:"status"`',
    '\tFixtureVersion string `json:"fixture_version"`',
    '\tSeed int `json:"seed"`',
    '\tSourceContractDigest string `json:"source_contract_digest"`',
    '\tVerificationBindings []VerificationBinding `json:"verification_bindings"`',
    '\tSourceContractRefs []SourceContractRef `json:"source_contract_refs"`',
    '\tContributions []Contribution `json:"contributions"`',
    '\tSemanticExpectations SemanticExpectations `json:"semantic_expectations"`',
    '\tRuntimeCredentialSets []RuntimeCredentialSet `json:"runtime_credential_sets"`',
    '\tArtifactPolicy ArtifactPolicy `json:"artifact_policy"`',
    '\tObjectPopulation ObjectPopulation `json:"object_population"`',
    '\tRedactionPolicy RedactionPolicy `json:"redaction_policy"`',
    "}",
    "",
    `const catalogJSON = \`${JSON.stringify(catalog)}\``,
    "",
    "func Profiles() []Profile {",
    "\tvar profiles []Profile",
    "\tif err := json.Unmarshal([]byte(catalogJSON), &profiles); err != nil {",
    '\t\tpanic("decode generated performance fixture profile catalog: " + err.Error())',
    "\t}",
    "\treturn profiles",
    "}",
    "",
    "func Lookup(profileID string) (Profile, bool) {",
    "\tfor _, profile := range Profiles() {",
    "\t\tif profile.FixtureProfileID == profileID {",
    "\t\t\treturn profile, true",
    "\t\t}",
    "\t}",
    "\treturn Profile{}, false",
    "}",
    "",
  ].join("\n");
}

export function renderPerformanceFixtureKeyVectors(profiles) {
  const migrationDigest = "e".repeat(64);
  const sourceContractDigest = "f".repeat(64);
  const vectors = orderedFixtureProfiles(profiles).map((profile) => {
    const vectorProfile = {
      ...profile,
      source_contract_digest: sourceContractDigest,
    };
    const input = snapshotKeyEnvelope(vectorProfile, migrationDigest);
    return {
      name: profile.fixture_profile_id,
      input,
      canonical_json: canonicalSnapshotKeyInput(vectorProfile, migrationDigest),
      snapshot_key: snapshotKey(vectorProfile, migrationDigest),
    };
  });
  return `${JSON.stringify({
    schema_id: "cartulary.performance_fixture_snapshot_key_vectors.v1",
    vectors,
  }, null, 2)}\n`;
}
