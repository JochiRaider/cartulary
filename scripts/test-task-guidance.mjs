#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const nodeBin = process.env.NODE_BIN || process.execPath;
const makeHelper = process.env.MAKE || "make";

const taskGuide = path.join(repoRoot, "scripts", "print-task-guide.mjs");
const explainPhase = path.join(repoRoot, "scripts", "print-explain-phase.mjs");
const explainTarget = path.join(repoRoot, "scripts", "print-explain-target.mjs");
const explainRun = path.join(repoRoot, "scripts", "print-explain-run.mjs");
const taskSurfaceManifest = path.join(repoRoot, "tools", "task_surface_manifest.json");

function fail(message) {
  throw new Error(message);
}

function assertContains(haystack, needle, label) {
  if (!haystack.includes(needle)) {
    fail(`${label}: expected output to contain [${needle}]`);
  }
}

function assertNotContains(haystack, needle, label) {
  if (haystack.includes(needle)) {
    fail(`${label}: expected output not to contain [${needle}]`);
  }
}

function parseJSON(value, label) {
  try {
    return JSON.parse(value);
  } catch (error) {
    fail(`${label}: invalid JSON: ${error.message}`);
  }
}

function runCapture(command, args = [], { env = {}, expectFailure = false, label = command } = {}) {
  const result = spawnSync(command, args, {
    cwd: repoRoot,
    env: { ...process.env, ...env },
    encoding: "utf8",
  });
  const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;
  if (result.error) {
    fail(`${label}: failed to start: ${result.error.message}`);
  }
  if (expectFailure) {
    if ((result.status ?? 0) === 0) {
      fail(`${label}: expected failure`);
    }
    return output;
  }
  if ((result.status ?? 0) !== 0) {
    fail(`${label}: exited ${result.status}\n${output}`);
  }
  return output;
}

function writeJSON(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function writeText(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, value);
}

function countFiles(dir) {
  let total = 0;
  const stack = [dir];
  while (stack.length > 0) {
    const current = stack.pop();
    if (!existsSync(current)) {
      continue;
    }
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
      } else if (entry.isFile()) {
        total += 1;
      }
    }
  }
  return total;
}

function rel(file) {
  return path.relative(repoRoot, file).replaceAll("\\", "/");
}

function createFixture(scenario) {
  const tmpRoot = path.join(repoRoot, "tmp");
  mkdirSync(tmpRoot, { recursive: true });
  const tmpDir = mkdtempSync(path.join(tmpRoot, `task-guidance-${scenario}.`));
  const resultsDir = path.join(tmpDir, "results");

  writeJSON(path.join(resultsDir, "run-a", "backend-store", "target-summary.json"), {
    target: "backend-store",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-b", "run-summary.json"), {
    label: "check",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-c", "run-summary.json"), {
    label: "ci",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-d", "frontend-unit", "target-summary.json"), {
    target: "not-frontend-unit",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-e", "frontend-unit", "target-summary.json"), {
    target: "frontend-unit",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-f", "migration-drift", "migration-drift", "phase-summary.json"), {
    target: "migration-drift",
    label: "migration-drift",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-g", "generate-drift", "generate-drift", "phase-summary.json"), {
    target: "not-generate-drift",
    label: "generate-drift",
    status: "pass",
  });

  const progressLog = path.join(resultsDir, "run-h", "check", "progress-summary.log");
  const schedulerLogsDir = path.join(resultsDir, "run-h", "check", "scheduler-logs");
  writeJSON(path.join(resultsDir, "run-h", "run-summary.json"), {
    label: "check",
    status: "pass",
    counts: { tests: 0 },
    work_units: { completed: 2, total: 2 },
    summary_targets: { expected: [] },
    evidence_targets: { present: [] },
    helper_units: { total: 0 },
    artifacts: { dir: "tmp/task-guidance/run-h" },
  });
  writeJSON(path.join(resultsDir, "run-h", "check", "scheduler-summary.json"), {
    schema_id: "cartulary.check_scheduler_summary.v6",
    target: "check",
    status: "pass",
    scheduler_kind: "check",
    completed_work_units: 2,
    total_work_units: 2,
    failed_work_unit: null,
    slowest_work_units: [{ label: "check-service-backed", duration_ms: 60700 }],
    progress_snapshots: [
      {
        completed: 1,
        total_work_units: 2,
        running: 1,
        pending: 0,
        blocked: 0,
        slowest_running: { label: "check-service-backed", duration_ms: 60700 },
        nested_scheduler_progress: [
          {
            work_unit: "check-service-backed",
            nested_target: "check-service-backed",
            completed: 4,
            total_work_units: 6,
            running: 1,
            pending: 1,
            blocked: 0,
            active_groups: { "backend-store": 1 },
            slowest_running: { label: "backend-store", duration_ms: 45000 },
          },
        ],
        line: "[PROGRESS] check 1/2: check 1/2, running check-service-backed, slowest check-service-backed 60.70s, logs tmp/task-guidance/run-h/check; bottleneck check-service-backed 4/6, running backend-store, slowest backend-store 45.00s",
      },
    ],
    slowest_running_observations: [
      {
        source: "nested",
        work_unit: "check-service-backed",
        nested_target: "check-service-backed",
        label: "backend-store",
        duration_ms: 45000,
      },
    ],
    artifacts: {
      scheduler_logs_dir: rel(schedulerLogsDir),
      progress_summary_log: rel(progressLog),
    },
  });
  writeText(
    progressLog,
    [
      "[CHECK-SCHEDULER] check start work_units=2 capacity={host_cpu:2}",
      "[PROGRESS] check 1/2: check 1/2, running check-service-backed, slowest check-service-backed 60.70s, logs tmp/task-guidance/run-h/check; bottleneck check-service-backed 4/6, running backend-store, slowest backend-store 45.00s",
      "[CHECK-SCHEDULER] check summary status=pass completed_work_units=2/2 failed=none slowest=check-service-backed:60.70s",
      "",
    ].join("\n"),
  );

  return {
    tmpDir,
    resultsDir,
    checkProgressRel: rel(progressLog),
    expectedResultsFiles: countFiles(resultsDir),
  };
}

function withFixture(scenario, fn) {
  const fixture = createFixture(scenario);
  try {
    fn(fixture);
    const actualFiles = countFiles(fixture.resultsDir);
    if (actualFiles !== fixture.expectedResultsFiles) {
      fail(
        `guidance commands must not create test report artifacts: expected ${fixture.expectedResultsFiles}, got ${actualFiles}`,
      );
    }
  } finally {
    rmSync(fixture.tmpDir, { recursive: true, force: true });
  }
}

function nodeScript(script, args = [], env = {}) {
  return runCapture(nodeBin, [script, ...args], { env, label: `${path.basename(script)} ${args.join(" ")}` });
}

function nodeScriptFails(script, args = [], env = {}) {
  return runCapture(nodeBin, [script, ...args], {
    env,
    expectFailure: true,
    label: `${path.basename(script)} ${args.join(" ")}`,
  });
}

function makeTarget(args = [], env = {}) {
  return runCapture(makeHelper, ["--no-print-directory", "-C", repoRoot, ...args], {
    env,
    label: `make ${args.join(" ")}`,
  });
}

function makeTargetFails(args = [], env = {}) {
  return runCapture(makeHelper, ["--no-print-directory", "-C", repoRoot, ...args], {
    env,
    expectFailure: true,
    label: `make ${args.join(" ")}`,
  });
}

function scenarioTaskGuideRoles(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };
  const defaultOutput = nodeScript(taskGuide, [], resultsEnv);
  assertContains(defaultOutput, "Cartulary task guide", "default task-guide header");
  assertContains(defaultOutput, "local-dev:", "default task-guide local-dev role");
  assertContains(defaultOutput, "feature-dev:", "default task-guide feature-dev role");
  assertContains(defaultOutput, "latest_artifact=none", "default task-guide reports missing artifacts");

  for (const role of ["local-dev", "feature-dev", "phase-author", "ci-investigator", "release"]) {
    const roleOutput = nodeScript(taskGuide, ["--role", role], resultsEnv);
    assertContains(roleOutput, `role=${role}`, "task-guide role header");
    assertContains(roleOutput, `${role}:`, "task-guide role section");
  }

  const phaseRoleOutput = nodeScript(taskGuide, ["--role", "phase-author", "--phase", "phase1"], resultsEnv);
  assertContains(phaseRoleOutput, "phase focus: phase1", "task-guide phase focus");
  assertContains(phaseRoleOutput, "make explain-phase PHASE=phase1", "task-guide phase command");
  parseJSON(nodeScript(taskGuide, ["--role", "feature-dev", "--json"]), "task-guide JSON");
}

function scenarioTaskGuidePhaseSlices(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };
  const phase4FeatureOutput = nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase4"], resultsEnv);
  assertContains(phase4FeatureOutput, "minimal phase slice: manifest rows that cover phase4", "phase4 feature-dev minimal tier");
  assertContains(phase4FeatureOutput, "service-backed slice: service-backed manifest rows that cover phase4", "phase4 feature-dev service tier");
  assertContains(phase4FeatureOutput, "general hygiene: useful non-phase checks", "phase4 feature-dev hygiene tier");
  assertContains(phase4FeatureOutput, "make phase-slice PHASE=phase4 | selected phase manifest-row slice | phase_relevance=phase_slice", "phase4 public phase slice");
  assertContains(phase4FeatureOutput, "make service-backed-slice PHASE=phase4 | selected phase service-backed manifest-row slice | phase_relevance=service_backed_slice", "phase4 public service-backed slice");
  assertContains(phase4FeatureOutput, "scheduler=Make node_tool;phase scheduler", "phase4 public phase scheduler owner");
  assertNotContains(phase4FeatureOutput, "make backend-unit | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-unit not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-store | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-store not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-integration | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-integration not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-integration-support | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-integration-support not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make browser-e2e-webserver-backed | selected phase execution dependency | phase_relevance=phase_slice", "phase4 browser not direct phase slice");
  assertContains(phase4FeatureOutput, "make frontend-unit | general hygiene outside the selected phase slice | phase_relevance=general_hygiene", "phase4 frontend-unit hygiene");
  assertContains(phase4FeatureOutput, "make lint | general hygiene outside the selected phase slice | phase_relevance=general_hygiene", "phase4 lint hygiene");
  assertNotContains(phase4FeatureOutput, "make frontend-unit | selected phase execution dependency | phase_relevance=phase_slice", "phase4 frontend-unit not phase slice");
  assertNotContains(phase4FeatureOutput, "make lint | selected phase execution dependency | phase_relevance=phase_slice", "phase4 lint not phase slice");

  const phase4Guide = parseJSON(
    nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase4", "--json"]),
    "phase4 task-guide JSON",
  );
  const manifest = parseJSON(readFileSync(taskSurfaceManifest, "utf8"), "task surface manifest");
  const role = phase4Guide.roles.find((entry) => entry.role === "feature-dev");
  if (!role || !Array.isArray(role.recommendation_tiers)) {
    fail("feature-dev must expose recommendation_tiers");
  }
  if ("recommendations" in role) {
    fail("legacy flat recommendations must not be present");
  }
  const tierByName = new Map(role.recommendation_tiers.map((tier) => [tier.name, tier]));
  for (const name of ["minimal phase slice", "service-backed slice", "full local gate", "general hygiene"]) {
    if (!tierByName.has(name)) {
      fail(`missing tier ${name}`);
    }
  }
  const classificationByTarget = new Map(manifest.targets.map((target) => [target.name, target.classification]));
  for (const tier of role.recommendation_tiers) {
    for (const item of tier.recommendations) {
      const commandTarget = item.target.split(/\s+/)[0];
      if (classificationByTarget.get(commandTarget) !== "public") {
        fail(`task-guide recommendation ${item.target} must reference a public target`);
      }
    }
  }
  const minimalTargets = new Set(tierByName.get("minimal phase slice").recommendations.map((item) => item.target));
  if (minimalTargets.size !== 1 || !minimalTargets.has("phase-slice PHASE=phase4")) {
    fail("phase4 minimal slice must recommend only phase-slice PHASE=phase4");
  }
  const phaseSlice = tierByName.get("minimal phase slice").recommendations[0];
  const phaseSliceChildren = new Set(phaseSlice.child_targets.map((item) => item.target));
  for (const target of [
    "backend-unit",
    "backend-store",
    "backend-integration",
    "backend-integration-support",
    "browser-e2e-webserver-backed",
  ]) {
    if (!phaseSliceChildren.has(target)) {
      fail(`phase4 minimal slice children missing ${target}`);
    }
  }
  const supportChild = phaseSlice.child_targets.find((item) => item.target === "backend-integration-support");
  if (!supportChild || supportChild.classification !== "check_internal") {
    fail("phase4 support child must remain check_internal");
  }
  for (const target of ["frontend-unit", "lint"]) {
    if (phaseSliceChildren.has(target)) {
      fail(`phase4 minimal slice children must not include ${target}`);
    }
  }
  for (const item of tierByName.get("minimal phase slice").recommendations) {
    if (item.phase_relevance !== "phase_slice") {
      fail(`minimal slice item ${item.target} has phase_relevance=${item.phase_relevance}`);
    }
  }
  const serviceTargets = new Set(tierByName.get("service-backed slice").recommendations.map((item) => item.target));
  if (serviceTargets.size !== 1 || !serviceTargets.has("service-backed-slice PHASE=phase4")) {
    fail("phase4 service-backed slice must recommend only service-backed-slice PHASE=phase4");
  }
  const serviceSlice = tierByName.get("service-backed slice").recommendations[0];
  const serviceChildren = new Set(serviceSlice.child_targets.map((item) => item.target));
  for (const target of [
    "backend-store",
    "backend-integration",
    "backend-integration-support",
    "browser-e2e-webserver-backed",
  ]) {
    if (!serviceChildren.has(target)) {
      fail(`phase4 service-backed slice children missing ${target}`);
    }
  }
  if (serviceChildren.has("backend-unit")) {
    fail("backend-unit must not be in the service-backed slice");
  }
  const gateTargets = new Set(tierByName.get("full local gate").recommendations.map((item) => item.target));
  if (!gateTargets.has("test-fast") || !gateTargets.has("check")) {
    fail("full local gate must include test-fast and check");
  }
  const hygieneTargets = new Set(tierByName.get("general hygiene").recommendations.map((item) => item.target));
  if (!hygieneTargets.has("frontend-unit") || !hygieneTargets.has("lint")) {
    fail("phase4 hygiene must include frontend-unit and lint");
  }
  for (const item of tierByName.get("general hygiene").recommendations) {
    if (item.phase_relevance !== "general_hygiene") {
      fail(`hygiene item ${item.target} has phase_relevance=${item.phase_relevance}`);
    }
  }

  const phase3Guide = parseJSON(
    nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase3", "--json"]),
    "phase3 task-guide JSON",
  );
  const phase3Role = phase3Guide.roles.find((entry) => entry.role === "feature-dev");
  const phase3TierByName = new Map(phase3Role.recommendation_tiers.map((tier) => [tier.name, tier]));
  const phase3Slice = phase3TierByName.get("minimal phase slice").recommendations[0];
  const phase3SliceChildren = new Set(phase3Slice.child_targets.map((item) => item.target));
  if (!phase3SliceChildren.has("frontend-unit")) {
    fail("phase3 frontend-unit evidence must stay in the minimal phase slice children");
  }
  const phase3HygieneTargets = new Set(phase3TierByName.get("general hygiene").recommendations.map((item) => item.target));
  if (phase3HygieneTargets.has("frontend-unit")) {
    fail("phase3 frontend-unit must not be general hygiene");
  }
}

function scenarioExplainPhase() {
  const phaseOutput = nodeScript(explainPhase, ["--phase", "phase1"]);
  assertContains(phaseOutput, "Cartulary phase guidance: phase1", "explain-phase header");
  assertContains(phaseOutput, "targets:", "explain-phase targets");
  assertContains(phaseOutput, "make backend-store", "explain-phase backend-store target");
  assertNotContains(phaseOutput, "make backend-integration-support", "explain-phase does not recommend internal support target");
  assertContains(phaseOutput, "internal target backend-integration-support classification=check_internal", "explain-phase shows internal support coverage");
  assertContains(phaseOutput, "ledger: docs/testing/phase1_coverage_ledger.md", "explain-phase ledger");

  const phaseJSON = parseJSON(nodeScript(explainPhase, ["--phase", "phase1", "--json"]), "explain-phase JSON");
  if (phaseJSON.phase !== "phase1" || !Array.isArray(phaseJSON.targets) || phaseJSON.targets.length === 0) {
    fail("explain-phase JSON must expose phase1 targets");
  }

  const plannedJSON = parseJSON(nodeScript(explainPhase, ["--phase", "phase5", "--json"]), "planned explain-phase JSON");
  if (plannedJSON.phase !== "phase5" || plannedJSON.status !== "planned" || plannedJSON.targets.length !== 0) {
    fail("planned explain-phase JSON must expose registered metadata without executable targets");
  }

  const unknownPhaseOutput = nodeScriptFails(explainPhase, ["--phase", "phase99"]);
  assertContains(unknownPhaseOutput, "unknown phase phase99", "unknown phase error");
}

function scenarioExplainTargetArtifacts(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };
  const backendSummary = nodeScript(explainTarget, ["--target", "backend-store"], resultsEnv);
  assertContains(backendSummary, "Cartulary target guidance: backend-store", "backend-store explain-target header");
  assertContains(backendSummary, "services: Postgres,MinIO", "backend-store service requirements");
  assertContains(backendSummary, "latest_artifact: tmp/task-guidance", "backend-store latest artifact");

  const backendRows = nodeScript(explainTarget, ["--target", "backend-store", "--detail", "rows"]);
  assertContains(backendRows, "rows:", "backend-store rows");
  assertContains(backendRows, "U-1-05", "backend-store Go rows");

  const frontendOutput = nodeScript(explainTarget, ["--target", "frontend-unit"]);
  assertContains(frontendOutput, "Cartulary target guidance: frontend-unit", "frontend-unit explain-target");
  assertContains(frontendOutput, "phase_coverage:", "frontend-unit phase coverage");

  const frontendIdentityOutput = nodeScript(explainTarget, ["--target", "frontend-unit"], resultsEnv);
  assertContains(frontendIdentityOutput, "latest_artifact: tmp/task-guidance", "matching target summary latest artifact");
  assertNotContains(frontendIdentityOutput, "run-d/frontend-unit/target-summary.json", "mismatched target summary ignored");

  const testFastMismatchOutput = nodeScript(explainTarget, ["--target", "test-fast"], resultsEnv);
  assertContains(testFastMismatchOutput, "latest_artifact: none", "mismatched check run summary ignored for test-fast");

  const releaseMismatchOutput = nodeScript(explainTarget, ["--target", "release-check"], resultsEnv);
  assertContains(releaseMismatchOutput, "latest_artifact: none", "mismatched check run summary ignored for release-check");

  const ciMatchOutput = nodeScript(explainTarget, ["--target", "ci"], resultsEnv);
  assertContains(ciMatchOutput, "run-c/run-summary.json", "matching ci run summary accepted");

  const migrationOutput = nodeScript(explainTarget, ["--target", "migration-drift"], resultsEnv);
  assertContains(migrationOutput, "services: Postgres", "migration-drift service requirements");
  assertContains(migrationOutput, "latest_artifact: tmp/task-guidance", "helper phase summary latest artifact");
  assertContains(migrationOutput, "run-f/migration-drift/migration-drift/phase-summary.json", "helper phase summary artifact path");

  const generateMismatchOutput = nodeScript(explainTarget, ["--target", "generate-drift"], resultsEnv);
  assertContains(generateMismatchOutput, "latest_artifact: none", "mismatched helper phase summary ignored");
  assertNotContains(generateMismatchOutput, "run-g/generate-drift/generate-drift/phase-summary.json", "mismatched helper phase summary path ignored");

  const browserOutput = nodeScript(explainTarget, ["--target", "browser-e2e-webserver-backed"]);
  assertContains(browserOutput, "browser stack", "browser explain-target service requirements");
  assertContains(browserOutput, "webserver-backed browser batch", "browser explain-target scheduler");

  const checkOutput = nodeScript(explainTarget, ["--target", "check"]);
  assertContains(checkOutput, "check scheduler", "check explain-target scheduler");
  assertContains(checkOutput, "phase_coverage:", "check explain-target phase coverage");

  const targetArtifacts = nodeScript(explainTarget, ["--target", "backend-store", "--detail", "artifacts"]);
  assertContains(targetArtifacts, "expected:", "target artifact expected paths");
  assertContains(targetArtifacts, "<run-id>/backend-store/target-summary.json", "target artifact expected summary");

  const helperArtifacts = nodeScript(explainTarget, ["--target", "migration-drift", "--detail", "artifacts"], resultsEnv);
  assertContains(helperArtifacts, "phase_summary: tmp/task-guidance", "target artifact discovered phase summary");
  assertContains(helperArtifacts, "<phase-label>/phase-summary.json", "target artifact expected phase summary");

  parseJSON(
    nodeScript(explainTarget, ["--target", "browser-e2e-webserver-backed", "--json"]),
    "explain-target JSON",
  );

  const unknownTargetOutput = nodeScriptFails(explainTarget, ["--target", "no-such-target"]);
  assertContains(unknownTargetOutput, "unknown target no-such-target", "unknown target error");
}

function scenarioExplainRunProgress(fixture) {
  const explainRunProgressSummary = nodeScript(explainRun, [
    "--results-dir",
    path.join(fixture.resultsDir, "run-h"),
    "--target",
    "check",
  ]);
  assertContains(explainRunProgressSummary, `[PROGRESS-LOG] check ${fixture.checkProgressRel}`, "explain-run summary progress artifact");
  assertContains(explainRunProgressSummary, "[PROGRESS-SNAPSHOT] [PROGRESS] check 1/2", "explain-run summary retained progress snapshot");
  assertContains(explainRunProgressSummary, "[SLOWEST-RUNNING] check check-service-backed:backend-store(45.00s)", "explain-run summary slowest retained running");

  const explainRunProgress = nodeScript(explainRun, [
    "--results-dir",
    path.join(fixture.resultsDir, "run-h"),
    "--target",
    "check",
    "--detail",
    "progress",
  ]);
  assertContains(explainRunProgress, `[PROGRESS-LOG] check ${fixture.checkProgressRel}`, "explain-run progress artifact header");
  assertContains(explainRunProgress, "[PROGRESS] check 1/2", "explain-run progress retained human progress");
  assertNotContains(explainRunProgress, "[LOG] check", "explain-run progress does not print child logs");
}

function scenarioGuidanceMakeWrappers(fixture) {
  const makeTaskGuide = makeTarget(["task-guide", "ROLE=feature-dev"]);
  assertContains(makeTaskGuide, "role=feature-dev", "make task-guide role");

  const makeTaskGuideJSON = parseJSON(
    makeTarget(["task-guide", "ROLE=feature-dev", "PHASE=phase4", "JSON=1"]),
    "make task-guide JSON",
  );
  if (makeTaskGuideJSON.role !== "feature-dev" || makeTaskGuideJSON.phase !== "phase4") {
    fail("make task-guide JSON must preserve role and phase");
  }

  const makePhase = makeTarget(["explain-phase", "PHASE=phase1"]);
  assertContains(makePhase, "Cartulary phase guidance: phase1", "make explain-phase");

  const makePhaseJSON = parseJSON(makeTarget(["explain-phase", "PHASE=phase1", "JSON=1"]), "make explain-phase JSON");
  if (makePhaseJSON.phase !== "phase1") {
    fail("make explain-phase JSON must preserve phase1");
  }

  const makeTargetRows = makeTarget(["explain-target", "TARGET=backend-store", "DETAIL=rows"], {
    CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir,
  });
  assertContains(makeTargetRows, "U-1-05", "make explain-target rows");

  const fixtureReportJSON = parseJSON(
    makeTarget(["fixture-report", "JSON=1", `RESULTS_DIR=${fixture.resultsDir}`]),
    "make fixture-report JSON",
  );
  if (fixtureReportJSON.schema_id !== "cartulary.fixture_report.v1") {
    fail("make fixture-report JSON must expose fixture report schema");
  }

  const detailZeroOutput = makeTargetFails(["explain-target", "TARGET=backend-store", "DETAIL=0"]);
  assertContains(detailZeroOutput, "usage: print-explain-target.mjs", "DETAIL=0 rejected");
}

const scenarios = new Map([
  ["task-guide-roles", scenarioTaskGuideRoles],
  ["task-guide-phase-slices", scenarioTaskGuidePhaseSlices],
  ["explain-phase", scenarioExplainPhase],
  ["explain-target-artifacts", scenarioExplainTargetArtifacts],
  ["explain-run-progress", scenarioExplainRunProgress],
  ["guidance-make-wrappers", scenarioGuidanceMakeWrappers],
]);

function main(argv) {
  const scenario = argv[0] ?? "";
  if (!scenario || !scenarios.has(scenario) || argv.length !== 1) {
    const names = Array.from(scenarios.keys()).join("|");
    fail(`usage: test-task-guidance.mjs <${names}>`);
  }
  withFixture(scenario, scenarios.get(scenario));
}

try {
  main(process.argv.slice(2));
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
