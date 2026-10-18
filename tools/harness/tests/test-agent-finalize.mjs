#!/usr/bin/env node

import assert from "node:assert/strict";
import fs, { readFileSync } from "node:fs";
import childProcess from "node:child_process";
import { syncBuiltinESMExports } from "node:module";
import { tmpdir } from "node:os";
import path from "node:path";

import {
  selectedActionDefinitions,
  selectedActions,
} from "../finalization/agent-finalize-action-plan.mjs";
import { changedTrackedWorkspace, restoreTrackedWorkspace, snapshotTrackedWorkspace, withTrackedWorkspaceTransaction } from "../finalization/tracked-workspace-snapshot.mjs";
import { createValidatedContext, buildSourceSnapshot } from "../test-catalog/index.mjs";
import { captureFinalizationSubstep } from "../finalization/captured-substep.mjs";
import { publishFinalizationFailure } from "../finalization/failure-publication.mjs";
import { assertRetainedFinalizerPreflight } from "./retained-finalizer-cases.mjs";

await assertRetainedFinalizerPreflight();

for (const [failures, reason, status] of [
  [[{ failure_class: "config", failure_reason: "configuration_error" }, { failure_class: "harness", failure_reason: "cleanup_error" }], "configuration_error", 2],
  [[{ failure_class: "harness", failure_reason: "cleanup_error" }], "cleanup_error", 12],
  [[{ failure_class: "artifact", failure_reason: "artifact_error" }], "artifact_error", 11],
  [[new Error("unclassified failure")], "unknown_failure", 1],
]) {
  let published;
  const exit = publishFinalizationFailure("/fixture", failures, (_root, error) => {
    published = error;
    return error;
  });
  assert.equal(published.failure_reason, reason, "public failure preserves action taxonomy");
  assert.equal(exit, status, "the public exit follows the published primary cause");
}

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

const fixture = fs.mkdtempSync(path.join(tmpdir(), "finalizer-input-boundary."));
const restricted = "human-material";
const originalRead = fs.readFileSync;
const originalStat = fs.lstatSync;
const originalSpawn = childProcess.spawnSync;
try {
  childProcess.execFileSync("git", ["init", "-q"], { cwd: fixture });
  fs.mkdirSync(path.join(fixture, restricted));
  fs.writeFileSync(path.join(fixture, restricted, "untrusted.sh"), "preserve this material");
  fs.writeFileSync(path.join(fixture, "README"), "human material");
  fs.writeFileSync(path.join(fixture, "owner.Md"), "human material");
  fs.writeFileSync(path.join(fixture, "source.mjs"), "before", { mode: 0o750 });
  fs.chmodSync(path.join(fixture, "source.mjs"), 0o750);
  fs.symlinkSync(`${restricted}/untrusted.sh`, path.join(fixture, "opaque-link"));
  childProcess.execFileSync("git", ["add", "."], { cwd: fixture });
  const assertAdmitted = (file) => {
    if (typeof file !== "string") return;
    assert.ok(!file.startsWith(path.join(fixture, restricted)) && !/\/(?:README|owner\.Md)$/u.test(file), "finalizer probed excluded human material");
  };
  fs.readFileSync = (file, ...args) => { assertAdmitted(file); return originalRead(file, ...args); };
  fs.lstatSync = (file, ...args) => { assertAdmitted(file); return originalStat(file, ...args); };
  let checkedGitAdmission = false;
  childProcess.spawnSync = (command, args, options) => {
    if (command === "git" && args[0] === "ls-files") {
      assert.ok(args.includes(`:(top,exclude,literal)${restricted}`));
      assert.ok(args.includes(":(exclude,icase)*.md"));
      assert.ok(args.includes(":(exclude,icase)README"));
      checkedGitAdmission = true;
    }
    return originalSpawn(command, args, options);
  };
  syncBuiltinESMExports();
  const snapshot = snapshotTrackedWorkspace(fixture, { restrictedRoots: [restricted] });
  assert.equal(checkedGitAdmission, true);
  assert.deepEqual([...snapshot.keys()], ["opaque-link", "source.mjs"]);
  assert.equal(snapshot.get("opaque-link").kind, "symlink");
  fs.writeFileSync(path.join(fixture, "source.mjs"), "after");
  fs.unlinkSync(path.join(fixture, "opaque-link"));
  fs.writeFileSync(path.join(fixture, "opaque-link"), "replacement");
  const restored = restoreTrackedWorkspace(fixture, snapshot);
  assert.equal(restored.status, "restored");
  assert.deepEqual(restored.restored_files, ["opaque-link", "source.mjs"]);
  assert.equal(readFileSync(path.join(fixture, "source.mjs"), "utf8"), "before");
  assert.equal(fs.statSync(path.join(fixture, "source.mjs")).mode & 0o777, 0o750);
  assert.equal(fs.readlinkSync(path.join(fixture, "opaque-link")), `${restricted}/untrusted.sh`);
  assert.equal(restoreTrackedWorkspace(fixture, snapshot).status, "not_needed");

  const ports = { snapshot: (root) => buildSourceSnapshot(root, { restrictedRoots: [restricted] }),
    loadCatalog: () => { validations += 1; return { rows: [{ row_id: "fixture.row" }], rowByID: new Map([["fixture.row", { status: "active" }]]) }; } };
  let validations = 0;
  const context = createValidatedContext(fixture, ports);
  const first = context.catalog();
  assert.equal(context.catalog(), first);
  assert.equal(validations, 1, "unchanged input snapshot validates once");
  assert.throws(() => { first.rows[0].row_id = "changed"; }, TypeError);
  assert.throws(() => first.rowByID.set("extra", {}), /immutable/u);
  assert.throws(() => first.rowByID.valueOf().clear(), /immutable/u);
  assert.throws(() => first.rowByID.forEach((_value, _key, map) => map.delete("fixture.row")), /immutable/u);
  fs.writeFileSync(path.join(fixture, "source.mjs"), "pre-existing user edit");
  assert.throws(() => context.catalog(), /stale/u);
  const next = createValidatedContext(fixture, ports);
  next.catalog();
  assert.equal(validations, 2);
  next.invalidate();
  assert.throws(() => next.catalog(), /stale/u);
  const changing = createValidatedContext(fixture, { ...ports, loadCatalog: () => {
    fs.writeFileSync(path.join(fixture, "source.mjs"), "changed during validation");
    return {};
  } });
  assert.throws(() => changing.catalog(), /stale/u, "a validator cannot accept a changing snapshot");
  fs.writeFileSync(path.join(fixture, "source.mjs"), "pre-existing user edit");
  const primary = Object.assign(new Error("injected primary failure"), { failure_class: "config", failure_reason: "configuration_error" });
  const transaction = await withTrackedWorkspaceTransaction(fixture, async () => {
    fs.writeFileSync(path.join(fixture, "source.mjs"), "generated mutation");
    fs.writeFileSync(path.join(fixture, "untracked.mjs"), "preserve untracked output");
    throw primary;
  }, { restrictedRoots: [restricted] });
  assert.equal(transaction.error, primary, "rollback preserves primary failure identity and taxonomy");
  assert.equal(transaction.rollback.status, "restored");
  assert.deepEqual(transaction.updatedFiles, []);
  assert.equal(readFileSync(path.join(fixture, "source.mjs"), "utf8"), "pre-existing user edit");
  assert.equal(readFileSync(path.join(fixture, "untracked.mjs"), "utf8"), "preserve untracked output");
  const beforeEdit = snapshotTrackedWorkspace(fixture, { restrictedRoots: [restricted] });
  fs.writeFileSync(path.join(fixture, "source.mjs"), "second dirty edit");
  assert.deepEqual(changedTrackedWorkspace(fixture, beforeEdit), ["source.mjs"], "content changes are visible even when Git status remains modified");

  for (const scenario of ["success", "failure", "failure-and-cleanup", "cleanup-failure", "capture-failure"]) {
    const childFailed = scenario === "failure" || scenario === "failure-and-cleanup";
    const events = [];
    const logDirectory = path.join(fixture, scenario);
    const result = await captureFinalizationSubstep({
      command: "make", args: ["source-check", "CHECK=schema"], repoRoot: root,
      runRoot: fixture, env: {}, logDirectory,
    }, {
      createLaunch: () => ({ settle: (outcome) => events.push(outcome) }),
      createFailureContext: () => ({
        environment: {},
        read: () => childFailed ? { failure_class: "config", failure_reason: "configuration_error" } : null,
        close: () => events.push("context released"),
      }),
      capture: async (_command, _args, options) => {
        assert.equal(options.tailBytes, 256 * 1024);
        if (scenario === "capture-failure") throw new Error("capture unavailable");
        return {
          status: childFailed ? 2 : 0,
          stdout: "validation started\n",
          stderr: childFailed ? "Error: invalid selector\nAuthorization: Bearer fixture-secret\n" : "",
          cleanup() {
            events.push("capture released");
            if (childFailed) {
              const retained = readFileSync(path.join(logDirectory, "stderr.log"), "utf8");
              assert.ok(retained.includes("invalid selector"));
              assert.ok(!retained.includes("fixture-secret"));
            }
            if (scenario === "cleanup-failure" || scenario === "failure-and-cleanup") throw new Error("release failed");
          },
        };
      },
    });
    assert.equal(result.status === 0, scenario === "success");
    assert.ok(events.includes("context released"));
    assert.equal(events.at(-1), scenario === "success" ? "passed" : "failed");
    if (childFailed) {
      assert.equal(result.status, 2);
      assert.equal(result.failure.failure_reason, "configuration_error");
      assert.ok(result.stderr.includes("invalid selector"));
      assert.ok(!result.stderr.includes("fixture-secret"));
      assert.equal(fs.statSync(result.stderrLog).mode & 0o777, 0o600);
    }
    if (scenario === "cleanup-failure") assert.equal(result.failure.failure_reason, "cleanup_error");
    if (scenario === "failure-and-cleanup") assert.equal(result.cleanupFailure.failure_reason, "cleanup_error");
    if (scenario === "capture-failure") assert.equal(result.failure.failure_reason, "artifact_error");
  }
} finally {
  fs.readFileSync = originalRead; fs.lstatSync = originalStat; childProcess.spawnSync = originalSpawn;
  syncBuiltinESMExports();
  fs.rmSync(fixture, { recursive: true, force: true });
}

