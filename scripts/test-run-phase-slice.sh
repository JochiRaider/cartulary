#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { spawnSync } from "node:child_process";

const [root] = process.argv.slice(2);
const nodeBin = process.env.NODE_BIN || process.execPath;
const script = path.join(root, "scripts/run-phase-slice.mjs");
const { runNormalizedSchedule } = await import(pathToFileURL(path.join(root, "tools/harness/scheduler/scheduler-runner.mjs")).href);
const targetPlanModule = await import(pathToFileURL(path.join(root, "scripts/lib/target-plan.mjs")).href);

function scenarioShardSuffix(scenarioID) {
  return scenarioID.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
}

function expectedScenarioShardNames({ phase, executionFamily, target, phasePrefix = "" }) {
  const row = targetPlanModule.collectTargetPlanRows(root).find(
    (candidate) =>
      candidate.manifest_phase === phase &&
      candidate.target === target &&
      candidate.execution_family === executionFamily,
  );
  assert.ok(row, `${phase} ${executionFamily} target-plan row must exist`);
  assert.ok(
    Object.keys(row.scenario_symbols ?? {}).length > 0,
    `${phase} ${executionFamily} target-plan row must declare scenario_symbols`,
  );
  return Object.keys(row.scenario_symbols)
    .map((scenarioID) => `${phasePrefix}${executionFamily}-${scenarioShardSuffix(scenarioID)}`)
    .sort((left, right) => left.localeCompare(right));
}

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

function frontendPlan(phase, mode, extraArgs = []) {
  const result = run([
    "--phase",
    phase,
    "--phase-namespace",
    "frontend",
    "--mode",
    mode,
    ...extraArgs,
    "--json",
  ]);
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
assert.deepEqual(phase4.service_requirements, ["browser_stack", "object_store", "postgres"]);
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

const phase5Service = plan("phase5", "service-backed");
const phase5VisualUnit = phase5Service.work_units.find((unit) => unit.target === "browser-e2e-visual");
assert.ok(phase5VisualUnit, "service-backed phase5 slice must include browser visual work");
assert.ok(phase5VisualUnit.resource_claims.browser_stage_visual >= 1, "service-backed phase5 visual work must claim visual stage lane");
assert.ok(!("browser_stage_visual_smoke" in phase5VisualUnit.resource_claims), "service-backed phase5 visual work must not claim visual-smoke stage lane");
assert.ok(phase5Service.resource_limits.browser_stage_visual >= 1, "service-backed phase5 must declare visual stage capacity");
assert.ok(!("browser_stage_visual_smoke" in phase5Service.resource_limits), "service-backed phase5 must not declare visual-smoke capacity");

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

const phase9 = plan("phase9", "phase");
for (const target of ["frontend-unit", "browser-e2e-webserver-backed"]) {
  const unit = phase9.work_units.find((candidate) => candidate.target === target);
  assert.equal(unit.frontend_row_accounting_scope?.mode, "disabled", `phase9 ${target} must disable FE row accounting`);
  assert.equal(unit.frontend_row_accounting_scope?.phase_namespace, "base", `phase9 ${target} FE row accounting must be base scoped`);
  assert.equal(unit.frontend_row_accounting_scope?.phase, "phase9", `phase9 ${target} FE row accounting must retain phase`);
}

const phase10 = plan("phase10", "phase");
const phase10OperatorExecutionFamily = "backend-process-phase10-operator-inspection";
const phase10OperatorShards = phase10.work_units
  .filter(
    (unit) =>
      unit.target === "backend-process" &&
      unit.id?.startsWith(`backend-process:phase10-${phase10OperatorExecutionFamily}-scn-`),
  )
  .sort((left, right) => left.id.localeCompare(right.id));
assert.deepEqual(
  phase10OperatorShards.map((unit) => unit.id.replace(/^backend-process:/, "")),
  expectedScenarioShardNames({
    phase: "phase10",
    target: "backend-process",
    executionFamily: phase10OperatorExecutionFamily,
    phasePrefix: "phase10-",
  }),
  "phase10 slice must expose E-10-01 operator evidence as stable scenario shards",
);
for (const unit of phase10OperatorShards) {
  assert.deepEqual(unit.runtime_binaries, ["operator"], `${unit.id} must consume the operator runtime`);
  assert.equal(unit.resource_claims?.process, 1, `${unit.id} must claim one process slot`);
}

const feP3 = frontendPlan("FE-P3", "phase");
assert.equal(feP3.schema_id, "cartulary.phase_slice_plan.v1");
assert.equal(feP3.phase_namespace, "frontend");
assert.equal(feP3.phase, "FE-P3");
assert.equal(feP3.target, "phase-slice");
assert.equal(feP3.no_op, false);
for (const target of [
  "frontend-typecheck",
  "frontend-unit",
  "generate-drift",
  "generated-artifact-policy-check",
  "phase-ledger-drift",
  "browser-e2e-support",
  "browser-e2e-webserver-backed",
  "browser-e2e-visual",
  "browser-e2e-a11y",
  "frontend-import-boundary-check",
]) {
  assert.ok(targets(feP3).has(target), `FE-P3 frontend phase slice must include ${target}`);
  assert.ok(workTargets(feP3, "make_target").has(target), `FE-P3 frontend phase slice must schedule ${target}`);
}
assert.deepEqual(feP3.service_requirements, ["browser_stack", "object_store", "postgres"]);
assert.equal(feP3.phase_claim_status, "complete");
const feP3FrontendUnit = feP3.work_units.find((unit) => unit.target === "frontend-unit");
assert.equal(feP3FrontendUnit.frontend_row_accounting_scope?.mode, "selected_rows");
assert.equal(feP3FrontendUnit.frontend_row_accounting_scope?.phase, "FE-P3");
assert.ok(
  feP3FrontendUnit.frontend_row_accounting_scope.selected_row_ids.includes("FE-U-P0-01"),
  "FE-P3 selected accounting must include FE-P0 rows",
);
assert.ok(
  feP3FrontendUnit.frontend_row_accounting_scope.selected_row_ids.includes("FE-I-P3-01"),
  "FE-P3 selected accounting must include FE-P3 rows",
);
assert.ok(
  !feP3FrontendUnit.frontend_row_accounting_scope.selected_row_ids.includes("FE-U-P4-01"),
  "FE-P3 selected accounting must exclude later FE-P4 rows",
);

const feP3Service = frontendPlan("FE-P3", "service-backed");
assert.equal(feP3Service.target, "service-backed-slice");
assert.equal(feP3Service.mode, "service_backed");
for (const target of [
  "browser-e2e-support",
  "browser-e2e-webserver-backed",
  "browser-e2e-visual",
  "browser-e2e-a11y",
]) {
  assert.ok(targets(feP3Service).has(target), `FE-P3 service-backed slice must include ${target}`);
}
for (const target of ["frontend-unit", "frontend-import-boundary-check", "lint-biome"]) {
  assert.ok(!targets(feP3Service).has(target), `FE-P3 service-backed slice must exclude ${target}`);
}

const feP0Service = frontendPlan("FE-P0", "service-backed");
assert.equal(feP0Service.target, "service-backed-slice");
assert.equal(feP0Service.no_op, true);
assert.deepEqual(feP0Service.child_targets, []);

const plannedFrontend = run([
  "--phase",
  "FE-P5",
  "--phase-namespace",
  "frontend",
  "--mode",
  "phase",
  "--json",
]);
assert.equal(plannedFrontend.status, 0, "complete frontend phases must be executable");
const plannedFeP5 = JSON.parse(plannedFrontend.stdout);
assert.equal(plannedFeP5.phase_claim_status, "complete");
assert.ok(targets(plannedFeP5).has("browser-e2e-stateful"), "FE-P5 full phase slice must include stateful browser work");

const selectedFeP5 = frontendPlan("FE-P5", "phase", ["--rows", "FE-I-P5-01"]);
assert.equal(selectedFeP5.phase, "FE-P5");
assert.deepEqual(selectedFeP5.selected_row_ids, ["FE-I-P5-01"]);
assert.deepEqual(targets(selectedFeP5), new Set(["browser-e2e-webserver-backed", "frontend-unit"]));
for (const unit of selectedFeP5.work_units) {
  assert.deepEqual(
    unit.frontend_row_accounting_scope?.selected_row_ids,
    ["FE-I-P5-01"],
    `${unit.target} selected row accounting must stay scoped to FE-I-P5-01`,
  );
}

const selectedFeP5Service = frontendPlan("FE-P5", "service-backed", ["--rows", "FE-I-P5-01"]);
assert.equal(selectedFeP5Service.target, "service-backed-slice");
assert.deepEqual(selectedFeP5Service.selected_row_ids, ["FE-I-P5-01"]);
assert.deepEqual(targets(selectedFeP5Service), new Set(["browser-e2e-webserver-backed"]));
assert.deepEqual(
  selectedFeP5Service.work_units[0]?.frontend_row_accounting_scope?.selected_row_ids,
  ["FE-I-P5-01"],
  "selected FE-P5 service-backed slice must keep browser row accounting scoped",
);

const tempRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-phase-slice-test-"));
try {
  mkdirSync(path.join(tempRoot, "tools"), { recursive: true });
  writePhaseRegistry(tempRoot, "phase100");
  writeFileSync(
    path.join(tempRoot, "tools/phase100_test_map.json"),
    `${JSON.stringify(
      {
        schema_id: "cartulary.phase_test_map.v2",
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
            evidence_class: "product_conformance",
            layer: "backend_unit",
            default_check_required: true,
            default_check_kind: "primary_local_evidence",
            default_check_reason_code: "cheapest_authoritative_layer",
            primary_evidence_owner: "phase-slice-fixture",
            duplicate_of: null,
            evidence_delta: "Synthetic phase-slice fixture coverage.",
            warm_local_cost_class: "low",
            evidence_layer: "phase_slice_smoke",
            claim: "synthetic phase-slice fixture selects future phase unit work",
            out_of_scope: "product behavior and service-backed coverage",
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
  assert.equal(schedulerSummary.schema_id, "cartulary.phase_slice_scheduler_summary.v4");
  assert.equal(schedulerSummary.status, "fail");
  assert.equal(schedulerSummary.failed_work_unit, "backend-unit");
  assert.deepEqual(schedulerSummary.observed_failed_work_units.map((unit) => unit.label), ["backend-unit"]);
  const targetSummary = readJSON(path.join(failureResults, "failure/phase-slice/target-summary.json"));
  assert.equal(targetSummary.status, "fail", "failed phase slice must write failing target summary");
} finally {
  rmSync(tempRoot, { recursive: true, force: true });
}

const finalizerLogDir = mkdtempSync(path.join(os.tmpdir(), "cartulary-finalizer-log-test-"));
const previousResultsDir = process.env.CARTULARY_TEST_RESULTS_DIR;
const previousRunID = process.env.CARTULARY_TEST_RUN_ID;
const previousOutputMode = process.env.CARTULARY_OUTPUT_MODE;
try {
  process.env.CARTULARY_TEST_RESULTS_DIR = path.join(finalizerLogDir, "results");
  process.env.CARTULARY_TEST_RUN_ID = "finalizer-log";
  process.env.CARTULARY_OUTPUT_MODE = "machine";
  const target = "scheduler-finalizer-log-contract";
  const result = await runNormalizedSchedule({
    repoRoot: root,
    testOutputScript: path.join(root, "tools/harness/core/test-output.sh"),
    schedule: {
      target,
      kind: "phase-slice",
      prefix: "PHASE-SCHEDULER",
      eventSchemaID: "cartulary.scheduler_event.v6",
      summarySchemaID: "cartulary.phase_slice_scheduler_summary.v4",
      resourceLimits: new Map(),
      resourceLimitSources: new Map(),
      workUnits: [
        {
          id: "finalize:backend-store",
          label: "finalize/backend-store",
          kind: "finalizer",
          target: "backend-store",
          aggregateTarget: "backend-store",
          countInTotal: false,
          countsStarted: false,
          resourceClaims: new Map(),
          needs: [],
          command: {
            command: nodeBin,
            args: ["-e", "console.log('fake finalized backend-store')"],
          },
        },
      ],
      totalWorkUnits: 0,
      finalizerCount: 1,
      showFinalizing: true,
      validateSummaryTiming: false,
      shouldReplayLog: () => false,
    },
  });
  assert.equal(result.status, 0, "finalizer-only scheduler fixture must pass");
  const finalizerLog = path.join(
    finalizerLogDir,
    "results/finalizer-log/scheduler-finalizer-log-contract/scheduler-logs/finalize-backend-store.log",
  );
  assert.ok(existsSync(finalizerLog), "default scheduler log naming must preserve finalize-<target>.log");
  assert.match(readFileSync(finalizerLog, "utf8"), /fake finalized backend-store/);
} finally {
  if (previousResultsDir === undefined) {
    delete process.env.CARTULARY_TEST_RESULTS_DIR;
  } else {
    process.env.CARTULARY_TEST_RESULTS_DIR = previousResultsDir;
  }
  if (previousRunID === undefined) {
    delete process.env.CARTULARY_TEST_RUN_ID;
  } else {
    process.env.CARTULARY_TEST_RUN_ID = previousRunID;
  }
  if (previousOutputMode === undefined) {
    delete process.env.CARTULARY_OUTPUT_MODE;
  } else {
    process.env.CARTULARY_OUTPUT_MODE = previousOutputMode;
  }
  rmSync(finalizerLogDir, { recursive: true, force: true });
}

const unknown = run(["--phase", "phase404", "--mode", "phase"], {}, { allowFailure: true });
assert.notEqual(unknown.status, 0, "unknown phase must fail");
assert.match(`${unknown.stdout}\n${unknown.stderr}`, /unknown phase phase404/);
EOF
