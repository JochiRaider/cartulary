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
const repoRoot = path.resolve(scriptDir, "../../../..");
const nodeBin = process.env.NODE_BIN || process.execPath;
const makeHelper = process.env.MAKE || "make";

const taskGuide = path.join(repoRoot, "tools", "harness", "diagnostics", "task-guide-cli.mjs");
const explainPhase = path.join(repoRoot, "tools", "harness", "diagnostics", "explain-phase-cli.mjs");
const explainTarget = path.join(repoRoot, "tools", "harness", "diagnostics", "explain-target-cli.mjs");
const explainRun = path.join(repoRoot, "tools", "harness", "diagnostics", "explain-run-cli.mjs");
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
    maxBuffer: 16 * 1024 * 1024,
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
  writeJSON(path.join(resultsDir, "run-i", "release-readiness-evidence", "target-summary.json"), {
    target: "release-readiness-evidence",
    status: "pass",
  });
  writeJSON(path.join(resultsDir, "run-i", "release-readiness-evidence", "release-readiness-evidence.json"), {
    schema_id: "cartulary.release_readiness_evidence.v2",
    status: "passed",
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
    schema_id: "cartulary.check_scheduler_summary.v10",
    target: "check",
    status: "pass",
    scheduler_kind: "check",
    completed_work_units: 2,
    total_work_units: 2,
    failed_work_unit: null,
    observed_failed_work_units: [],
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

function createContext(scenario) {
  const fixture = createFixture(scenario);
  return {
    ...fixture,
    manifest: parseJSON(readFileSync(taskSurfaceManifest, "utf8"), "task surface manifest"),
  };
}

function withContext(scenario, fn) {
  const fixture = createContext(scenario);
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
  assertContains(defaultOutput, "make task-guide ROLE=feature-dev", "default task-guide bounded role example");
  assertContains(defaultOutput, "use ROLE=<role> and PHASE=phaseN", "default task-guide narrowing hint");

  for (const role of ["local-dev", "feature-dev", "phase-author", "ci-investigator", "release"]) {
    const roleOutput = nodeScript(taskGuide, ["--role", role], resultsEnv);
    assertContains(roleOutput, `role=${role}`, "task-guide role header");
    assertContains(roleOutput, `${role}:`, "task-guide role section");
  }

  const phaseRoleOutput = nodeScript(taskGuide, ["--role", "phase-author", "--phase", "phase1"], resultsEnv);
  assertContains(phaseRoleOutput, "phase focus: phase1", "task-guide phase focus");
  assertContains(phaseRoleOutput, "make explain-phase PHASE=phase1", "task-guide phase command");
  const guideJSON = parseJSON(nodeScript(taskGuide, ["--role", "feature-dev", "--json"]), "task-guide JSON");
  if (guideJSON.schema_id !== "cartulary.task_guide.v1") {
    fail("task-guide JSON must expose schema_id cartulary.task_guide.v1");
  }
}

function scenarioTaskGuidePhaseSlices(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };
  const phase4FeatureOutput = nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase4"], resultsEnv);
  assertContains(phase4FeatureOutput, "minimal phase slice: manifest rows that cover phase4", "phase4 feature-dev minimal tier");
  assertContains(phase4FeatureOutput, "service-backed slice: service-backed manifest rows that cover phase4", "phase4 feature-dev service tier");
  assertContains(phase4FeatureOutput, "general hygiene: useful non-phase checks", "phase4 feature-dev hygiene tier");
  assertContains(phase4FeatureOutput, "make phase-slice PHASE=phase4 | selected phase manifest-row slice | phase_relevance=phase_slice", "phase4 public phase slice");
  assertContains(phase4FeatureOutput, "make service-backed-slice PHASE=phase4 | selected phase service-backed manifest-row slice | phase_relevance=service_backed_slice", "phase4 public service-backed slice");
  assertContains(phase4FeatureOutput, "execution=public evidence: 6 targets, support/internal evidence: 1 target", "phase4 public phase execution summary");
  assertContains(phase4FeatureOutput, "public evidence:", "phase4 public evidence group");
  assertContains(phase4FeatureOutput, "support/internal evidence:", "phase4 support/internal evidence group");
  assertNotContains(phase4FeatureOutput, "scheduler=", "task-guide must not print flat scheduler owner fields");
  assertNotContains(phase4FeatureOutput, "make backend-unit | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-unit not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-store | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-store not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-integration | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-integration not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make backend-integration-support | selected phase execution dependency | phase_relevance=phase_slice", "phase4 backend-integration-support not direct phase slice");
  assertNotContains(phase4FeatureOutput, "make browser-e2e-webserver-backed | selected phase execution dependency | phase_relevance=phase_slice", "phase4 browser not direct phase slice");
  assertContains(phase4FeatureOutput, "make frontend-unit services=none coverage=phases=phase4", "phase4 frontend-unit phase evidence");
  assertContains(phase4FeatureOutput, "make lint | general hygiene outside the selected phase slice | phase_relevance=general_hygiene", "phase4 lint hygiene");
  assertNotContains(phase4FeatureOutput, "make lint | selected phase execution dependency | phase_relevance=phase_slice", "phase4 lint not phase slice");

  const phase4Guide = parseJSON(
    nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase4", "--json"]),
    "phase4 task-guide JSON",
  );
  const manifest = fixture.manifest;
  if (phase4Guide.schema_id !== "cartulary.task_guide.v1") {
    fail("phase4 task-guide JSON must expose task guide schema_id");
  }
  if (phase4Guide.execution_map?.schema_id !== "cartulary.task_execution_map.v1") {
    fail("phase4 task-guide JSON must expose execution_map schema_id");
  }
  if (JSON.stringify(phase4Guide).includes("scheduler_owner")) {
    fail("task-guide JSON must not expose legacy scheduler_owner");
  }
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
  const targetClassByTarget = new Map(manifest.targets.map((target) => [target.name, target.target_class]));
  for (const tier of role.recommendation_tiers) {
    for (const item of tier.recommendations) {
      const commandTarget = item.target.split(/\s+/)[0];
      if (targetClassByTarget.get(commandTarget) !== "public") {
        fail(`task-guide recommendation ${item.target} must reference a public target`);
      }
      if (!("execution_summary" in item)) {
        fail(`task-guide recommendation ${item.target} must expose execution_summary`);
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
    "frontend-unit",
    "browser-e2e-webserver-backed",
    "browser-e2e-visual",
  ]) {
    if (!phaseSliceChildren.has(target)) {
      fail(`phase4 minimal slice children missing ${target}`);
    }
  }
  const supportChild = phaseSlice.child_targets.find((item) => item.target === "backend-integration-support");
  if (!supportChild || supportChild.target_class !== "check_internal") {
    fail("phase4 support child must remain check_internal");
  }
  for (const target of ["lint"]) {
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
    "browser-e2e-visual",
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
  if (hygieneTargets.has("frontend-unit") || !hygieneTargets.has("lint")) {
    fail("phase4 hygiene must include lint and exclude frontend-unit");
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
  assertContains(phaseOutput, "execution map:", "explain-phase execution map");
  assertContains(phaseOutput, "public evidence:", "explain-phase public evidence group");
  assertContains(phaseOutput, "support/internal evidence:", "explain-phase support evidence group");
  assertContains(phaseOutput, "make backend-store", "explain-phase backend-store target");
  assertNotContains(phaseOutput, "make backend-integration-support", "explain-phase does not recommend internal support target");
  assertContains(phaseOutput, "internal target backend-integration-support target_class=check_internal", "explain-phase shows internal support coverage");
  assertContains(phaseOutput, "ledger: docs/testing/phase1_coverage_ledger.md", "explain-phase ledger");
  assertNotContains(phaseOutput, "scheduler=", "explain-phase must not print flat scheduler owner fields");

  const phaseJSON = parseJSON(nodeScript(explainPhase, ["--phase", "phase1", "--json"]), "explain-phase JSON");
  if (phaseJSON.phase !== "phase1" || !Array.isArray(phaseJSON.targets) || phaseJSON.targets.length === 0) {
    fail("explain-phase JSON must expose phase1 targets");
  }
  if (phaseJSON.schema_id !== "cartulary.phase_guidance.v1") {
    fail("explain-phase JSON must expose phase guidance schema_id");
  }
  if (phaseJSON.execution_map?.schema_id !== "cartulary.task_execution_map.v1") {
    fail("explain-phase JSON must expose execution_map schema_id");
  }
  if (JSON.stringify(phaseJSON).includes("scheduler_owner")) {
    fail("explain-phase JSON must not expose legacy scheduler_owner");
  }

  const phase5JSON = parseJSON(nodeScript(explainPhase, ["--phase", "phase5", "--json"]), "phase5 explain-phase JSON");
  if (
    phase5JSON.phase !== "phase5" ||
    phase5JSON.status !== "active" ||
    phase5JSON.targets.length === 0 ||
    phase5JSON.execution_map.children.length === 0
  ) {
    fail("phase5 explain-phase JSON must expose active executable targets");
  }

  const unknownPhaseOutput = nodeScriptFails(explainPhase, ["--phase", "phase99"]);
  assertContains(unknownPhaseOutput, "unknown phase phase99", "unknown phase error");
}

function scenarioExplainTargetArtifacts(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };
  const backendSummary = nodeScript(explainTarget, ["--target", "backend-store"], resultsEnv);
  assertContains(backendSummary, "Cartulary target guidance: backend-store", "backend-store explain-target header");
  assertContains(backendSummary, "services: Postgres,object store", "backend-store service requirements");
  assertContains(backendSummary, "scheduler paths:", "backend-store scheduler paths");
  assertContains(backendSummary, "latest_artifact: tmp/task-guidance", "backend-store latest artifact");
  assertNotContains(backendSummary, "scheduler=", "explain-target must not print flat scheduler owner fields");

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
  assertContains(browserOutput, "webserver-backed browser batch", "browser explain-target scheduler path");
  assertContains(browserOutput, "stage=webserver-backed", "browser explain-target scheduler stage");
  assertContains(browserOutput, "expected_artifacts:", "browser explain-target expected artifacts");

  const checkOutput = nodeScript(explainTarget, ["--target", "check"]);
  assertContains(checkOutput, "check scheduler", "check explain-target scheduler");
  assertContains(checkOutput, "phase_coverage:", "check explain-target phase coverage");

  const lintOutput = nodeScript(explainTarget, ["--target", "lint"]);
  assertContains(lintOutput, "sequence_steps:", "lint explain-target sequence steps");
  assertContains(lintOutput, "sequence_summary_groups:", "lint explain-target summary groups");
  assertContains(lintOutput, "blocking-lint: blocking targets=", "lint blocking child semantics");
  assertContains(lintOutput, "lint-markdown", "lint markdown child appears in blocking group");
  assertContains(lintOutput, "lint-shell", "lint shell child appears in blocking group");

  const cleanOutput = nodeScript(explainTarget, ["--target", "clean"]);
  assertContains(cleanOutput, "inputs: CARTULARY_CLEANUP_DRY_RUN", "clean explain-target input contract");
  assertContains(cleanOutput, "dry_run: CARTULARY_CLEANUP_DRY_RUN=1 make clean", "clean explain-target dry-run usage");

  const servicesDownOutput = nodeScript(explainTarget, ["--target", "services-down"]);
  assertContains(
    servicesDownOutput,
    "inputs: CARTULARY_CLEANUP_DRY_RUN",
    "services-down explain-target input contract",
  );
  assertContains(
    servicesDownOutput,
    "dry_run: CARTULARY_CLEANUP_DRY_RUN=1 make services-down",
    "services-down explain-target dry-run usage",
  );

  const dbResetOutput = nodeScript(explainTarget, ["--target", "db-reset"]);
  assertContains(
    dbResetOutput,
    "inputs: CARTULARY_CLEANUP_DRY_RUN,CARTULARY_DESTRUCTIVE_CONFIRM",
    "db-reset explain-target destructive input contract",
  );
  assertContains(
    dbResetOutput,
    "dry_run: CARTULARY_CLEANUP_DRY_RUN=1 make db-reset",
    "db-reset explain-target dry-run usage",
  );

  const objectStoreResetOutput = nodeScript(explainTarget, ["--target", "object-store-reset"]);
  assertContains(
    objectStoreResetOutput,
    "inputs: CARTULARY_CLEANUP_DRY_RUN,CARTULARY_DESTRUCTIVE_CONFIRM",
    "object-store-reset explain-target destructive input contract",
  );
  assertContains(
    objectStoreResetOutput,
    "dry_run: CARTULARY_CLEANUP_DRY_RUN=1 make object-store-reset",
    "object-store-reset explain-target dry-run usage",
  );

  const serviceBackedOutput = nodeScript(explainTarget, ["--target", "test-service-backed"]);
  assertContains(
    serviceBackedOutput,
    "services: Postgres,object store,browser stack",
    "test-service-backed service requirements",
  );

  const fastServiceBackedOutput = nodeScript(explainTarget, ["--target", "test-fast-service-backed"]);
  assertContains(
    fastServiceBackedOutput,
    "services: Postgres,object store",
    "test-fast-service-backed service requirements",
  );
  assertNotContains(
    fastServiceBackedOutput,
    "browser stack",
    "test-fast-service-backed does not require browser stack",
  );

  const testArtifacts = nodeScript(explainTarget, ["--target", "test", "--detail", "artifacts"]);
  assertContains(testArtifacts, "<run-id>/test/target-summary.json", "test sequence expected target summary");
  assertContains(testArtifacts, "<run-id>/run-summary.json", "test sequence expected run summary");
  assertNotContains(testArtifacts, "<run-id>/test/scheduler-summary.json", "test sequence no scheduler summary");
  assertNotContains(testArtifacts, "<run-id>/test/progress-summary.log", "test sequence no progress summary");

  const checkArtifacts = nodeScript(explainTarget, ["--target", "check", "--detail", "artifacts"]);
  assertContains(checkArtifacts, "<run-id>/check/target-summary.json", "check expected target summary");
  assertContains(checkArtifacts, "<run-id>/check/scheduler-summary.json", "check expected scheduler summary");
  assertContains(checkArtifacts, "<run-id>/check/progress-summary.log", "check expected progress summary");
  assertContains(checkArtifacts, "<run-id>/run-summary.json", "check expected run summary");

  const targetArtifacts = nodeScript(explainTarget, ["--target", "backend-store", "--detail", "artifacts"]);
  assertContains(targetArtifacts, "expected:", "target artifact expected paths");
  assertContains(targetArtifacts, "<run-id>/backend-store/target-summary.json", "target artifact expected summary");

  const helperArtifacts = nodeScript(explainTarget, ["--target", "migration-drift", "--detail", "artifacts"], resultsEnv);
  assertContains(helperArtifacts, "phase_summary: tmp/task-guidance", "target artifact discovered phase summary");
  assertContains(helperArtifacts, "<phase-label>/phase-summary.json", "target artifact expected phase summary");

  const releaseReadinessArtifacts = nodeScript(
    explainTarget,
    ["--target", "release-readiness-evidence", "--detail", "artifacts"],
    resultsEnv,
  );
  assertContains(
    releaseReadinessArtifacts,
    "release_readiness_evidence: tmp/task-guidance",
    "release readiness evidence artifact discovered",
  );
  assertContains(
    releaseReadinessArtifacts,
    "<run-id>/release-readiness-evidence/release-readiness-evidence.json",
    "release readiness evidence expected artifact",
  );

  const browserJSON = parseJSON(
    nodeScript(explainTarget, ["--target", "browser-e2e-webserver-backed", "--json"]),
    "explain-target JSON",
  );
  if (browserJSON.schema_id !== "cartulary.target_guidance.v1") {
    fail("explain-target JSON must expose target guidance schema_id");
  }
  if (browserJSON.execution_map?.schema_id !== "cartulary.task_execution_map.v1") {
    fail("explain-target JSON must expose execution_map schema_id");
  }
  if (JSON.stringify(browserJSON).includes("scheduler_owner")) {
    fail("explain-target JSON must not expose legacy scheduler_owner");
  }

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
  assertContains(detailZeroOutput, "DETAIL must be one of summary, rows, artifacts", "DETAIL=0 rejected");
}

function scenarioGuidanceCore(fixture) {
  const resultsEnv = { CARTULARY_TEST_RESULTS_DIR: fixture.resultsDir };

  const phase4Guide = parseJSON(
    nodeScript(taskGuide, ["--role", "feature-dev", "--phase", "phase4", "--json"], resultsEnv),
    "core task-guide JSON",
  );
  if (phase4Guide.schema_id !== "cartulary.task_guide.v1") {
    fail("core task-guide JSON must expose task guide schema_id");
  }
  if (phase4Guide.role !== "feature-dev" || phase4Guide.phase !== "phase4") {
    fail("core task-guide JSON must preserve feature-dev phase4 selection");
  }
  const featureRole = phase4Guide.roles.find((entry) => entry.role === "feature-dev");
  const tierNames = new Set((featureRole?.recommendation_tiers ?? []).map((tier) => tier.name));
  for (const tier of ["minimal phase slice", "service-backed slice", "general hygiene"]) {
    if (!tierNames.has(tier)) {
      fail(`core task-guide missing tier ${tier}`);
    }
  }
  if (JSON.stringify(phase4Guide).includes("scheduler_owner")) {
    fail("core task-guide JSON must not expose legacy scheduler_owner");
  }

  const makeTaskGuideJSON = parseJSON(
    makeTarget(["task-guide", "ROLE=feature-dev", "PHASE=phase4", "JSON=1"], resultsEnv),
    "core make task-guide JSON",
  );
  if (makeTaskGuideJSON.role !== "feature-dev" || makeTaskGuideJSON.phase !== "phase4") {
    fail("core make task-guide JSON must preserve role and phase");
  }

  const phaseJSON = parseJSON(nodeScript(explainPhase, ["--phase", "phase1", "--json"], resultsEnv), "core explain-phase JSON");
  if (phaseJSON.schema_id !== "cartulary.phase_guidance.v1" || phaseJSON.phase !== "phase1") {
    fail("core explain-phase JSON must expose phase1 guidance");
  }

  const backendSummary = nodeScript(explainTarget, ["--target", "backend-store"], resultsEnv);
  assertContains(backendSummary, "Cartulary target guidance: backend-store", "core explain-target header");
  assertContains(backendSummary, "latest_artifact: tmp/task-guidance", "core explain-target latest artifact");

  const explainRunProgressSummary = nodeScript(explainRun, [
    "--results-dir",
    path.join(fixture.resultsDir, "run-h"),
    "--target",
    "check",
  ]);
  assertContains(explainRunProgressSummary, "[PROGRESS-SNAPSHOT] [PROGRESS] check 1/2", "core explain-run progress");
}

const scenarios = new Map([
  ["guidance-core", scenarioGuidanceCore],
  ["task-guide-roles", scenarioTaskGuideRoles],
  ["task-guide-phase-slices", scenarioTaskGuidePhaseSlices],
  ["explain-phase", scenarioExplainPhase],
  ["explain-target-artifacts", scenarioExplainTargetArtifacts],
  ["explain-run-progress", scenarioExplainRunProgress],
  ["guidance-make-wrappers", scenarioGuidanceMakeWrappers],
]);

const suites = new Map([
  ["core", ["guidance-core"]],
  [
    "matrix",
    [
      "task-guide-roles",
      "task-guide-phase-slices",
      "explain-phase",
      "explain-target-artifacts",
      "explain-run-progress",
      "guidance-make-wrappers",
    ],
  ],
]);

function runScenarioSet(label, names) {
  withContext(label, (fixture) => {
    for (const name of names) {
      scenarios.get(name)(fixture);
    }
  });
}

function main(argv) {
  const scenario = argv[0] ?? "";
  if (argv.length === 0) {
    runScenarioSet("guidance-matrix", suites.get("matrix"));
    return;
  }
  if (argv.length !== 1) {
    const names = [...suites.keys(), ...scenarios.keys()].join("|");
    fail(`usage: test-task-guidance.mjs [${names}]`);
  }
  if (suites.has(scenario)) {
    runScenarioSet(`guidance-${scenario}`, suites.get(scenario));
    return;
  }
  if (!scenarios.has(scenario)) {
    const names = [...suites.keys(), ...scenarios.keys()].join("|");
    fail(`usage: test-task-guidance.mjs [${names}]`);
  }
  runScenarioSet(scenario, [scenario]);
}

try {
  main(process.argv.slice(2));
} catch (error) {
  console.error(error instanceof Error ? error.message : String(error));
  process.exit(1);
}
