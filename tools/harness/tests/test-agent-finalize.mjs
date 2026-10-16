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
import { restoreTrackedWorkspace, snapshotTrackedWorkspace } from "../finalization/tracked-workspace-snapshot.mjs";
import { captureFinalizationSubstep } from "../finalization/captured-substep.mjs";
import { assertRetainedFinalizerPreflight } from "./retained-finalizer-cases.mjs";

await assertRetainedFinalizerPreflight();

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

  for (const scenario of ["success", "failure", "failure-and-cleanup", "cleanup-failure", "capture-failure"]) {
    const childFailed = scenario === "failure" || scenario === "failure-and-cleanup";
    const events = [];
    const logDirectory = path.join(fixture, scenario);
    const result = await captureFinalizationSubstep({
      command: "make", args: ["json-shape-check"], repoRoot: root,
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

