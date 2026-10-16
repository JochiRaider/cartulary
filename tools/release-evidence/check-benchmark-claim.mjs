#!/usr/bin/env node

import { createHash } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import {
  assertMachineEvidencePath,
  canonicalJSONString,
  parseStrictJSON,
  readRetainedArtifactBytes,
  validateSchemaSync,
} from "../harness/contract/index.mjs";
import {
  collectFinalizedMeasurementSummaries,
  collectFrontendMeasurementObservations,
} from "../harness/browser/frontend-measurement-evidence.mjs";
import {
  activePerformanceFixtureProfile,
  loadPerformanceFixtureSnapshotRegistry,
  performanceFixturePredicateIDsForRows,
  postgresMigrationDigest,
  snapshotKey,
  validateProfileMeasurementObservation,
} from "../harness/performance-fixture/index.mjs";
import { adaptPlaywrightReport } from "../harness/execution/runners/playwright.mjs";
import { loadTestCatalog, resolvePlaywrightSelection } from "../harness/test-catalog/index.mjs";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const defaultManifest = path.join(repoRoot, ".cartulary/benchmark/benchmark_manifest.json");
const envelopeLimit = 1024 * 1024;
const artifactLimit = 8 * 1024 * 1024;
const reportLimit = 32 * 1024 * 1024;
const hash = bytes => `sha256:${createHash("sha256").update(bytes).digest("hex")}`;
const same = (left, right) => canonicalJSONString(left) === canonicalJSONString(right);

function admittedPath(root, relative) {
  if (typeof relative !== "string" || path.isAbsolute(relative) || relative.includes("\\") ||
      relative.includes("\0") || relative.split("/").some(part => !part || part === "." || part === "..")) {
    throw new Error("benchmark artifact path must be normalized and relative");
  }
  const file = path.join(root, relative);
  assertMachineEvidencePath(file);
  return file;
}

function readJSON(root, relative, schemaID, limit = artifactLimit) {
  admittedPath(root, relative);
  const bytes = readRetainedArtifactBytes(root, relative, limit);
  const value = parseStrictJSON(bytes.toString("utf8"), "benchmark machine evidence");
  if (schemaID) validateSchemaSync(schemaID, value);
  return { value, bytes };
}

function readReference(root, reference, schemaID, limit = artifactLimit) {
  const result = readJSON(root, reference.path, schemaID, limit);
  if (hash(result.bytes) !== reference.sha256) throw new Error("benchmark artifact digest mismatch");
  return result;
}

export function loadBenchmarkProfile(root = repoRoot) {
  return readJSON(root, "contracts/claim-publication/benchmark-profile.v3.json", "cartulary.benchmark_profile.v1", envelopeLimit).value;
}

// Derive publication membership from the existing product contracts and active
// fixture bindings. Neither titles nor an independent publication catalog own it.
function currentPredicates(root, registry) {
  const predicates = new Map();
  for (const profile of registry.profiles.values()) {
    if (profile.status !== "active") continue;
    const candidates = new Map();
    for (const ref of profile.source_contract_refs) {
      const contract = readJSON(root, ref.path, ref.schema_id).value;
      const criteria = contract.criterion_ids ?? (contract.criterion_id ? [contract.criterion_id] : []);
      if (!criteria.length) continue;
      for (const predicate of contract.predicates ?? (contract.predicate_id ? [contract] : [])) {
        if (candidates.has(predicate.predicate_id)) throw new Error("ambiguous typed predicate owner");
        candidates.set(predicate.predicate_id, criteria);
      }
    }
    for (const binding of profile.verification_bindings) {
      const criteria = candidates.get(binding.predicate_id);
      if (!criteria || predicates.has(binding.predicate_id)) throw new Error("unresolved or duplicate current predicate binding");
      predicates.set(binding.predicate_id, { criteria, profile });
    }
  }
  return predicates;
}

function validateBuild(profile, build, lease, migrationDigest) {
  const expectedReceipts = profile.contributions.map(contribution => ({
    contribution_id: contribution.contribution_id,
    version: contribution.version,
    owner_id: contribution.owner_id,
    counts: contribution.expected_receipt_counts.map(count => ({ count_id: count.count_id, actual: count.exact })),
  }));
  const expectedValidation = {
    counts: profile.semantic_expectations.counts.map(count => ({ expectation_id: count.expectation_id, actual: count.exact, passed: true })),
    conditions: profile.semantic_expectations.conditions.map(condition => ({ expectation_id: condition.expectation_id, actual: condition.required, passed: true })),
    connections_closed: true,
  };
  if (build.fixture_version !== profile.fixture_version || build.seed !== profile.seed ||
      build.source_contract_digest !== profile.source_contract_digest || build.migration_digest !== migrationDigest ||
      build.snapshot_key !== snapshotKey(profile, migrationDigest) || !same(build.contribution_receipts, expectedReceipts) ||
      !same(build.validation, expectedValidation) || lease.creation_state !== "created" ||
      build.builder_unit_id !== lease.builder_unit_id || build.redaction_policy_id !== profile.redaction_policy.policy_id ||
      lease.redaction_policy_id !== profile.redaction_policy.policy_id) {
    throw new Error("benchmark physical snapshot differs from current owner contracts");
  }
}

function reportResults(suites) {
  return suites.flatMap(suite => [
    ...(suite.specs ?? []).flatMap(spec => (spec.tests ?? []).map(test => test)),
    ...reportResults(suite.suites ?? []),
  ]);
}

function validateBrowserResult(report, receipt, receiptBytes, profile) {
  const tests = reportResults(report.suites ?? []);
  if (tests.length !== 1 || tests[0].results?.length !== 1 || tests[0].results[0].status !== "passed" ||
      (tests[0].results[0].retry ?? 0) !== 0 || report.config?.version !== receipt.playwright_version) {
    throw new Error("benchmark requires one passed browser execution without retry");
  }
  const attachments = (tests[0].results[0].attachments ?? []).filter(item => item.name === receipt.schema_id);
  if (attachments.length !== 1 || typeof attachments[0].body !== "string" ||
      attachments[0].body.length > envelopeLimit * 2 || attachments[0].path !== undefined ||
      !Buffer.from(attachments[0].body, "base64").equals(receiptBytes)) {
    throw new Error("browser environment receipt is not bound to the measured result");
  }
  if (receipt.browser_engine !== profile.environment.browser_engine ||
      receipt.browser_build !== profile.environment.browser_build ||
      receipt.browser_mode !== profile.environment.browser_mode ||
      receipt.playwright_version !== profile.environment.benchmark_harness_version || receipt.platform !== "linux" || receipt.architecture !== "x64") {
    throw new Error("actual browser environment differs from the benchmark profile");
  }
}

export function validateBenchmarkManifest(manifest, manifestDir, { root = repoRoot } = {}) {
  try {
    validateSchemaSync("cartulary.benchmark_manifest.v2", manifest);
    const profile = loadBenchmarkProfile(root);
    if (manifest.benchmark_profile_id !== profile.benchmark_profile_id || !same(manifest.environment, profile.environment)) {
      throw new Error("benchmark environment differs from the exact current profile");
    }
    const bundleFile = admittedPath(manifestDir, manifest.bundle.path);
    const bundle = readReference(manifestDir, manifest.bundle, "cartulary.benchmark_evidence_bundle.v1", envelopeLimit).value;
    const registry = loadPerformanceFixtureSnapshotRegistry(root);
    const predicates = currentPredicates(root, registry);
    const catalog = loadTestCatalog(root);
    const migrationDigest = postgresMigrationDigest(root);
    const pins = readJSON(root, "tools/toolchain_pins.json", null).value;
    if (pins.schema_id !== "cartulary.toolchain_pins.v1" || !/^\d+\.\d+\.\d+$/u.test(pins.node_version) ||
        pins.ui_review?.playwright !== profile.environment.benchmark_harness_version) {
      throw new Error("benchmark runtime projection differs from the toolchain pin owner");
    }
    const expected = [...predicates].filter(([, entry]) => entry.criteria.some(criterion => manifest.criterion_ids.includes(criterion))).map(([id]) => id).sort();
    if (manifest.criterion_ids.some(criterion => ![...predicates.values()].some(entry => entry.criteria.includes(criterion)))) {
      throw new Error("criterion has no admitted current exercising binding");
    }
    const observed = [];
    const roots = new Set();
    const runIDs = new Set();
    for (const run of bundle.runs) {
      const runRoot = admittedPath(path.dirname(bundleFile), run.root);
      if (roots.has(runRoot)) throw new Error("duplicate benchmark evidence root");
      roots.add(runRoot);
      if (run.manifest.path !== "run-manifest.json" || run.summary.path !== "run-summary.json") throw new Error("noncanonical run evidence paths");
      const identity = readReference(runRoot, run.manifest, "cartulary.harness_run_manifest.v2").value;
      const result = readReference(runRoot, run.summary, "cartulary.harness_run_summary.v1").value;
      if (runIDs.has(identity.run_id) || identity.run_id !== result.run_id || identity.target !== result.target ||
          identity.source_digest !== manifest.source_digest || result.status !== "pass" || result.cache.hit !== 0 ||
          result.unit_counts.total === 0 || result.unit_counts.passed !== result.unit_counts.total ||
          result.unit_counts.failed !== 0 || result.unit_counts.skipped !== 0 || result.unit_counts.cancelled !== 0 || result.failure_class !== null || result.failure_reason !== null) {
        throw new Error("benchmark run is incomplete, failed, cached or source-incompatible");
      }
      runIDs.add(identity.run_id);
      for (const measurement of run.measurements) {
        const summary = readReference(runRoot, measurement.summary, "cartulary.frontend_measurement_summary.v5").value;
        const binding = predicates.get(summary.rollup.predicate_id);
        if (!binding || !binding.criteria.some(criterion => manifest.criterion_ids.includes(criterion)) || observed.includes(summary.rollup.predicate_id)) {
          throw new Error("unknown, undeclared or duplicate measurement predicate");
        }
        const fixture = activePerformanceFixtureProfile(registry, summary.fixture_profile_id);
        const row = catalog.rowByID.get(summary.row_id);
        if (fixture !== binding.profile || !row || row.status !== "active" ||
            !same(performanceFixturePredicateIDsForRows(root, [row], { registry }), [summary.rollup.predicate_id])) {
          throw new Error("measurement does not match its active owner selection");
        }
        const policy = fixture.artifact_policy;
        const [entry] = collectFinalizedMeasurementSummaries({
          buildSchemaID: policy.build_schema_id, leaseSchemaID: policy.lease_schema_id,
          observationSchemaID: policy.observation_schema_id, summarySchemaID: policy.summary_schema_id,
          objectPopulation: fixture.object_population, expectedPredicateIDs: [summary.rollup.predicate_id],
          runRoot, summaryPaths: [admittedPath(runRoot, measurement.summary.path)],
          validateObservation: observation => validateProfileMeasurementObservation(root, fixture, observation),
        });
        if (entry.summary.qualification_outcome !== "qualified" || entry.observation.artifact.outcome !== "passed" ||
            entry.lease.artifact.predicate_id !== summary.rollup.predicate_id) throw new Error("measurement did not qualify");
        validateBuild(fixture, entry.build.artifact, entry.lease.artifact, migrationDigest);
        const report = readReference(runRoot, measurement.report, null, reportLimit).value;
        const selected = resolvePlaywrightSelection(root, [row]);
        if (adaptPlaywrightReport([row], report, 0, null, selected)[0].terminal_state !== "passed") {
          throw new Error("browser result differs from its exact catalog selection");
        }
        const environment = readReference(runRoot, measurement.browser_environment, "cartulary.browser_execution_environment.v1", envelopeLimit);
        if (environment.value.node_version !== `v${pins.node_version}`) throw new Error("browser runner Node version differs from its pin");
        validateBrowserResult(report, environment.value, environment.bytes, profile);
        const [observation] = collectFrontendMeasurementObservations({
          expectedPredicateIDs: [summary.rollup.predicate_id], observationSchemaID: policy.observation_schema_id,
          reportPaths: [admittedPath(runRoot, measurement.report.path)], runRoot,
        });
        if (!same(observation, entry.observation.artifact)) throw new Error("browser observation differs from finalized evidence");
        observed.push(summary.rollup.predicate_id);
      }
    }
    if (!same(observed.sort(), expected)) throw new Error("published criterion has missing or extra current predicates");
    return [];
  } catch (error) {
    return [error.message];
  }
}

function main() {
  const file = path.resolve(repoRoot, process.argv[2] ?? process.env.BENCHMARK_MANIFEST ?? defaultManifest);
  assertMachineEvidencePath(file);
  let manifest;
  try {
    manifest = readJSON(path.dirname(file), path.basename(file), "cartulary.benchmark_manifest.v2", envelopeLimit).value;
  } catch (error) {
    if (error.code === "ENOENT" && file === defaultManifest) {
      console.log("benchmark claim manifest absent: no claim-bearing benchmark publication requested");
      return;
    }
    if (error.code === "ENOENT") throw new Error("benchmark manifest missing", { cause: error });
    throw error;
  }
  const errors = validateBenchmarkManifest(manifest, path.dirname(file));
  if (errors.length) throw new Error(errors.join("; "));
  console.log("benchmark evidence package valid; publication prerequisite review remains separate");
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try { main(); } catch (error) { console.error(`benchmark claim check failed: ${error.message}`); process.exitCode = 1; }
}
