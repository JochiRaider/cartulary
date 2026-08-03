#!/usr/bin/env node

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  selectedActionDefinitions,
  selectedActions,
} from "../finalization/agent-finalize-action-plan.mjs";

const definitions = [
  { actionID: "generated_structure_refresh", requiresResultsDir: false, mutating: true, substeps: [] },
  { actionID: "schema_shape_validation", requiresResultsDir: false, mutating: false, substeps: [] },
  { actionID: "tier_coverage_validation", requiresResultsDir: false, mutating: false, substeps: [] },
  { actionID: "canonical_evidence_validation", requiresResultsDir: true, mutating: false, substeps: [] },
  { actionID: "scheduler_drift_validation", requiresResultsDir: true, mutating: false, substeps: [] },
];
assert.deepEqual(
  selectedActionDefinitions(definitions, "").map((entry) => entry.actionID),
  ["schema_shape_validation", "tier_coverage_validation", "generated_structure_refresh", "canonical_evidence_validation", "scheduler_drift_validation"],
);
assert.deepEqual(
  selectedActionDefinitions(definitions, "retained").map((entry) => entry.actionID),
  ["scheduler_drift_validation", "schema_shape_validation", "tier_coverage_validation", "generated_structure_refresh", "canonical_evidence_validation"],
);
const actions = selectedActions(definitions, "");
assert.equal(actions.filter((entry) => entry.execution_state === "not_selected").length, 2);
assert.equal(actions.some((entry) => Object.hasOwn(entry, "cache")), false, "finalizer action caches are retired");
const root = path.resolve(import.meta.dirname, "../../..");
const source = readFileSync(path.join(root, "tools/harness/finalization/agent-finalize-cli.mjs"), "utf8");
for (const retired of ["duration_baseline_refresh", "CARTULARY_AGENT_FINALIZE_ACTION_CACHE", "go-test-duration-baselines"]) {
  assert.equal(source.includes(retired), false, `${retired} must not survive the finalizer cutover`);
}

