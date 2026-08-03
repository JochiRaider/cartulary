#!/usr/bin/env node

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import { semanticJSONSHA256 } from "../../test-catalog/index.mjs";
import {
  assertBaselineClosure,
  compareQualifiedBaselines,
  median,
  nearestRankP90,
  performanceRoster,
} from "../canonical-performance.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const surface = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
const roster = performanceRoster(surface);
assert.equal(roster.size, 47, "the measured public-target roster must remain exact");
assert.deepEqual(
  [...roster.entries()].filter(([, policy]) => policy.gate === "required_improvement").map(([target]) => target),
  [
    "agent-finalize",
    "browser-e2e",
    "browser-e2e-a11y",
    "browser-e2e-measurement",
    "browser-e2e-stateful",
    "browser-e2e-visual",
    "browser-e2e-webserver-backed",
    "check",
    "ci",
    "go-vulncheck",
    "harness-contract",
    "release-check",
    "test",
    "test-fast",
  ],
);
assert.equal(median([1, 9, 3, 5, 7]), 5);
assert.equal(nearestRankP90([1, 2, 3, 4, 5]), 5);

const evidenceFixture = {
  schema_id: "cartulary.harness_performance_evidence_roots.v3",
  mode: "comparison",
  reference_windows: [{
    window_id: "reference",
    provider_target: "check",
    cold_root: "cold",
    warmup_root: "warmup",
    measured_roots: ["r1", "r2", "r3", "r4", "r5"],
  }],
  reference_bindings: [{
    target: "check",
    window_id: "reference",
    timing_source: "public_invocation_envelope",
  }],
  candidate_windows: [{
    window_id: "candidate",
    provider_target: "check",
    cold_root: "cold-candidate",
    warmup_root: "warmup-candidate",
    measured_roots: ["c1", "c2", "c3", "c4", "c5"],
  }],
  candidate_bindings: [{
    target: "check",
    window_id: "candidate",
    timing_source: "public_invocation_envelope",
  }],
};
validateSchemaSync(evidenceFixture.schema_id, evidenceFixture);

function qualifiedBaseline({ candidate = false } = {}) {
  const targets = [...roster.entries()].map(([target, policy]) => {
    const required = policy.gate === "required_improvement";
    const value = candidate ? (required ? 80 : 101) : 100;
    const executionPolicy = { target, graph_digest: `sha256:${"a".repeat(64)}` };
    return {
      target,
      gate: policy.gate,
      command_id: policy.command_id,
      measurement_profile_id: policy.measurement_profile_id,
      canonical_inputs: policy.canonical_inputs,
      timing_source: "canonical_unit_interval_union",
      source_commit: (candidate ? "b" : "a").repeat(40),
      source_snapshot_sha256: (candidate ? "b" : "a").repeat(64),
      system_profile_sha256: "c".repeat(64),
      toolchain_profile_sha256: "d".repeat(64),
      workload_evidence_profile_sha256: "e".repeat(64),
      execution_policy: executionPolicy,
      execution_policy_sha256: semanticJSONSHA256(executionPolicy),
      ...(policy.allowed_policy_transition
        ? { allowed_policy_transition: policy.allowed_policy_transition }
        : {}),
      sample_provider_target: target,
      cold_root: `${target}-${candidate ? "candidate" : "reference"}-cold`,
      cold_ms: value + 10,
      warmup_root: `${target}-${candidate ? "candidate" : "reference"}-warmup`,
      sample_roots: [1, 2, 3, 4, 5].map((index) => `${target}-${candidate ? "c" : "r"}${index}`),
      sample_count: 5,
      samples_ms: [value, value, value, value, value],
      p50_ms: value,
      p90_ms: value,
      mad_ms: 0,
      timing_accounting: {
        critical_path_p50_ms: candidate && required ? 80 : 100,
        resource_blocking_p50_ms: 10,
        setup_p50_ms: 1,
        fixture_p50_ms: 2,
        execution_p50_ms: value - 5,
        collation_p50_ms: 1,
        wrapper_p50_ms: 1,
        unattributed_p50_ms: 0,
        process_count_p50: 3,
      },
    };
  });
  const sourceCommit = (candidate ? "b" : "a").repeat(40);
  const sourceSnapshot = (candidate ? "b" : "a").repeat(64);
  return {
    schema_id: "cartulary.harness_public_target_duration_baselines.v3",
    status: "qualified",
    timing_authority: "canonical_unit_events",
    sample_rule: "one_cold_one_discarded_warmup_five_warm_sixth_only_if_unstable",
    targets,
    internal_diagnostics: [],
    source_windows: [{
      source_commit: sourceCommit,
      source_snapshot_sha256: sourceSnapshot,
      targets: [...roster.keys()],
    }],
    public_entrypoint_portfolio: {
      target_count: targets.length,
      targets: [...roster.keys()],
      total_p50_ms: targets.reduce((sum, row) => sum + row.p50_ms, 0),
    },
    rejected_roots: [],
  };
}

const reference = qualifiedBaseline();
const candidate = qualifiedBaseline({ candidate: true });
assertBaselineClosure(reference, surface);
assertBaselineClosure(candidate, surface);
const comparison = compareQualifiedBaselines(reference, candidate, surface);
assert.deepEqual(comparison.failures, []);
assert.equal(comparison.rows.length, 47);

const regressed = structuredClone(candidate);
const row = regressed.targets.find((entry) => entry.gate === "no_regression");
row.samples_ms = [200, 200, 200, 200, 200];
row.p50_ms = 200;
row.p90_ms = 200;
regressed.public_entrypoint_portfolio.total_p50_ms = regressed.targets.reduce((sum, entry) => sum + entry.p50_ms, 0);
assert.ok(compareQualifiedBaselines(reference, regressed, surface).failures.includes(row.target));
