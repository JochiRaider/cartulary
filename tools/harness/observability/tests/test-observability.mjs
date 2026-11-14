#!/usr/bin/env node

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";

import { validateSchemaSync } from "../../contract/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");

function event(seq, monotonicMs, kind, unitID, status, extra = {}) {
  return {
    schema_id: "cartulary.harness_unit_event.v4",
    seq,
    monotonic_ms: monotonicMs,
    event: kind,
    unit_id: unitID,
    status,
    needs: [],
    resource_claims: { cpu: 1, memory_bytes: 1024, process: 1 },
    service_dependencies: [],
    ...(kind === "queued" ? { unit_kind: "runner", row_ids: [] } : {}),
    ...extra,
  };
}

test("canonical event schemas admit representative execution records and reject invalid sequence types", () => {
  const events = [
    event(1, 0, "queued", "a", "pending"),
    event(2, 5, "started", "a", "running"),
    event(3, 7, "started", "b", "running"),
    event(4, 20, "completed", "a", "passed", { output_digest: `sha256:${"a".repeat(64)}` }),
    event(5, 25, "completed", "b", "passed", { output_digest: `sha256:${"b".repeat(64)}` }),
  ];
  for (const entry of events) validateSchemaSync(entry.schema_id, entry);
  for (const seq of [0, "1"]) assert.throws(() => validateSchemaSync(events[0].schema_id, { ...events[0], seq }));
});

test("canonical event schemas admit cache and fixture records and reject invalid lease types", () => {
  const events = [
    event(1, 0, "cache_miss", "row:a", "pending", { cache_profile_id: "go.test", cache_reason: "not_found" }),
    event(2, 1, "fixture_acquired", "row:a", "running", { fixture_lease_id: "lease:a" }),
    event(3, 10, "fixture_released", "row:a", "running", { fixture_lease_id: "lease:a" }),
    event(4, 11, "completed", "row:a", "passed"),
  ];
  for (const entry of events) validateSchemaSync(entry.schema_id, entry);
  for (const entry of events.filter(entry => entry.fixture_lease_id)) assert.throws(() => validateSchemaSync(entry.schema_id, { ...entry, fixture_lease_id: 1 }));
});

test("canonical summary schema admits a representative record and rejects negative counts", () => {
  const summary = {
    schema_id: "cartulary.harness_run_summary.v1",
    run_id: "contract-run",
    target: "check",
    status: "pass",
    failure_class: null,
    failure_reason: null,
    unit_counts: { total: 2, passed: 2, failed: 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 20,
    critical_path: ["a", "b"],
    actual_dependency_critical_path_ms: 20,
    timing_accounting: {
      setup_ms: 0,
      fixture_ms: 0,
      execution_ms: 20,
      collation_ms: 0,
      wrapper_ms: 0,
      unattributed_ms: 0,
      resource_blocking_ms: 0,
      process_count: 2,
    },
    resource_pressure: { wait_events: 0 },
    cache: { hit: 0, miss: 2, bypass: 0 },
    artifact_refs: ["run-manifest.json", "unit-events.ndjson", "target-summaries/check.json"],
  };
  validateSchemaSync(summary.schema_id, summary);
  assert.throws(() => validateSchemaSync(summary.schema_id, { ...summary, unit_counts: { ...summary.unit_counts, total: -1 } }));
});

test("retired phase timing authorities are absent from current public recipes", () => {
  const owner = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
  const text = JSON.stringify(owner.make_recipes);
  for (const retired of [
    "test-local",
    "test-fast-service-backed",
    "test-service-backed",
    "check-service-backed",
    "release-browser-readiness",
  ]) {
    assert.equal(text.includes(`\"${retired}\"`), false, `retired phase ${retired} remains in current recipes`);
  }
});
