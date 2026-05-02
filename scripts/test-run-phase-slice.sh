#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const [root] = process.argv.slice(2);
const nodeBin = process.env.NODE_BIN || process.execPath;
const script = path.join(root, "scripts/run-phase-slice.mjs");

function run(args, env = {}, options = {}) {
  const result = spawnSync(nodeBin, [script, ...args], {
    cwd: root,
    env: {
      ...process.env,
      ...env,
    },
    encoding: "utf8",
  });
  if (options.allowFailure !== true && result.status !== 0) {
    throw new Error(
      `run-phase-slice failed status=${result.status}\nstdout=${result.stdout}\nstderr=${result.stderr}`,
    );
  }
  return result;
}

function plan(phase, mode) {
  const result = run(["--phase", phase, "--mode", mode, "--json"]);
  return JSON.parse(result.stdout);
}

function targets(plan) {
  return new Set(plan.child_targets);
}

function workTargets(plan, kind = "") {
  return new Set(
    plan.work_units
      .filter((unit) => kind === "" || unit.kind === kind)
      .map((unit) => unit.target),
  );
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function writePhaseRegistry(root, phase) {
  const phaseNumber = phase.replace(/^phase/, "");
  writeFileSync(
    path.join(root, "tools/phase_registry.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.phase_registry.v1",
        phases: [
          {
            phase,
            order: Number.parseInt(phaseNumber, 10),
            status: "active",
            label: `Phase ${phaseNumber}`,
            manifest_path: `tools/${phase}_test_map.json`,
            ledger_path: `docs/testing/${phase}_coverage_ledger.md`,
            scope: `synthetic ${phase} scope.`,
            normative_owners: "Synthetic owner.",
          },
        ],
      },
      null,
      2,
    )}\n`,
  );
}

const phase4 = plan("phase4", "phase");
assert.equal(phase4.schema_id, "cartulary.phase_slice_plan.v1");
assert.equal(phase4.target, "phase-slice");
assert.equal(phase4.no_op, false);
for (const target of [
  "backend-unit",
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "browser-e2e-webserver-backed",
]) {
  assert.ok(targets(phase4).has(target), `phase4 full slice must include ${target}`);
}
assert.ok(workTargets(phase4, "go_target").has("backend-unit"), "phase4 full slice must include pure backend unit work");
assert.ok(workTargets(phase4, "go_shard").has("backend-store"), "phase4 full slice must include service-backed store Go shards");
assert.ok(workTargets(phase4, "go_shard").has("backend-integration"), "phase4 full slice must include service-backed integration Go shards");
assert.ok(workTargets(phase4, "go_shard").has("backend-integration-support"), "phase4 full slice must include support Go shards");
assert.ok(workTargets(phase4, "browser_target").has("browser-e2e-webserver-backed"), "phase4 full slice must include browser functional work");
assert.deepEqual(phase4.service_requirements, ["browser_stack", "minio", "postgres"]);
assert.ok(phase4.resource_limits.browser_stage_webserver_backed >= 1, "phase4 browser stage resource must be declared");
assert.ok(
  phase4.work_units.some((unit) => unit.id.includes("phase4-backend-store")),
  "phase4 Go shard names must carry the phase selection",
);

const phase4Service = plan("phase4", "service-backed");
assert.equal(phase4Service.target, "service-backed-slice");
assert.equal(phase4Service.mode, "service_backed");
assert.ok(!targets(phase4Service).has("backend-unit"), "service-backed phase4 slice must exclude backend-unit child target");
assert.ok(!workTargets(phase4Service).has("backend-unit"), "service-backed phase4 slice must exclude backend-unit work");
assert.ok(
  !phase4Service.row_groups.some((group) => group.execution_dependency === "backend_unit"),
  "service-backed phase4 slice must exclude pure backend unit rows",
);
for (const target of [
  "backend-store",
  "backend-integration",
  "backend-integration-support",
  "browser-e2e-webserver-backed",
]) {
  assert.ok(targets(phase4Service).has(target), `phase4 service-backed slice must include ${target}`);
}

const phase3 = plan("phase3", "phase");
const frontendGroups = phase3.row_groups.filter((group) => group.target === "frontend-unit");
assert.equal(frontendGroups.length, 1, "phase3 must have one frontend-unit row group");
assert.equal(frontendGroups[0].runner, "vitest");
assert.equal(frontendGroups[0].execution_dependency, "frontend_unit");
assert.equal(frontendGroups[0].coverage, "authoritative");
assert.deepEqual(frontendGroups[0].ids, [
  "U-3-05",
  "U-3-12",
  "U-3-13",
  "U-3-GRID-01",
  "U-3-GRID-02",
  "U-3-GRID-03",
]);
assert.ok(workTargets(phase3, "frontend_unit").has("frontend-unit"), "phase3 must schedule frontend-unit through Vitest phase work");

const tempRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-phase-slice-test-"));
try {
  mkdirSync(path.join(tempRoot, "tools"), { recursive: true });
  writePhaseRegistry(tempRoot, "phase100");
  writeFileSync(
    path.join(tempRoot, "tools/phase100_test_map.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.phase_test_map.v1",
        phase: "phase100",
        note: "Synthetic phase-slice fixture.",
        ledger: {
          title: "Phase 100 Coverage Ledger",
          notes: "Synthetic phase-slice fixture.",
          authoritative_execution: "make phase-slice PHASE=phase100",
          support_execution_extras: [],
          sections: [],
          shared_harness: [],
          support_only: [],
        },
        expected_ids: ["U-100-01"],
        support_go_targets: [],
        unit: [
          {
            id: "U-100-01",
            coverage: "authoritative",
            runner: "go_test",
            package: "./internal/modules/auth",
            file: "internal/modules/auth/phase100_unit_test.go",
            symbol: "TestPhase100_UnitOnly_U_100_01",
            execution_dependency: "backend_unit",
            execution_family: "backend-unit",
            execution_label: "backend-unit phase100 authoritative",
          },
        ],
        integration: [],
        e2e: [],
      },
      null,
      2,
    )}\n`,
  );

  const noopResults = path.join(tempRoot, "noop-results");
  const noop = run(["--phase", "phase100", "--mode", "service-backed"], {
    CARTULARY_PHASE_MANIFEST_ROOT: tempRoot,
    CARTULARY_TEST_RESULTS_DIR: noopResults,
    CARTULARY_TEST_RUN_ID: "noop",
  });
  assert.match(noop.stdout, /\[NOOP\] service-backed-slice phase=phase100 mode=service-backed children=0/);
  const noopSummary = readJSON(path.join(noopResults, "noop/service-backed-slice/target-summary.json"));
  assert.equal(noopSummary.status, "pass", "pure-only service-backed no-op must write a passing target summary");

  const fakeRunner = path.join(tempRoot, "fake-cartulary-runner.mjs");
  writeFileSync(
    fakeRunner,
    [
      "#!/usr/bin/env node",
      "if (process.argv[2] === 'go-target' && process.argv[3] === 'backend-unit') process.exit(17);",
      "process.exit(0);",
      "",
    ].join("\n"),
  );
  chmodSync(fakeRunner, 0o755);
  const failureResults = path.join(tempRoot, "failure-results");
  const failed = run(["--phase", "phase100", "--mode", "phase", "--inside-service-wrapper"], {
    CARTULARY_PHASE_MANIFEST_ROOT: tempRoot,
    CARTULARY_RUNNER_SCRIPT: fakeRunner,
    CARTULARY_TEST_RESULTS_DIR: failureResults,
    CARTULARY_TEST_RUN_ID: "failure",
  }, { allowFailure: true });
  assert.notEqual(failed.status, 0, "failed selected work must fail the phase scheduler");
  assert.match(failed.stdout, /\[SUMMARY\] target=phase-slice status=fail /);
  const schedulerSummaryPath = path.join(failureResults, "failure/phase-slice/scheduler-summary.json");
  assert.ok(existsSync(schedulerSummaryPath), "failed phase slice must write scheduler summary");
  const schedulerSummary = readJSON(schedulerSummaryPath);
  assert.equal(schedulerSummary.status, "fail");
  assert.equal(schedulerSummary.failed_work_unit, "backend-unit");
  const targetSummary = readJSON(path.join(failureResults, "failure/phase-slice/target-summary.json"));
  assert.equal(targetSummary.status, "fail", "failed phase slice must write failing target summary");
} finally {
  rmSync(tempRoot, { recursive: true, force: true });
}

const unknown = run(["--phase", "phase404", "--mode", "phase"], {}, { allowFailure: true });
assert.notEqual(unknown.status, 0, "unknown phase must fail");
assert.match(`${unknown.stdout}\n${unknown.stderr}`, /unknown phase phase404/);
EOF
