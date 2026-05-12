#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { chmodSync, copyFileSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
process.chdir(root);

const topologyModule = await import(pathToFileURL(path.join(root, "scripts/lib/execution-topology.mjs")));
const targetPlanModule = await import(pathToFileURL(path.join(root, "scripts/lib/target-plan.mjs")));
const serviceRendererModule = await import(
  pathToFileURL(path.join(root, "scripts/render-service-backed-schedule-manifest.mjs"))
);
const checkServiceBackedExpansionModule = await import(
  pathToFileURL(path.join(root, "scripts/lib/check-service-backed-expansion.mjs"))
);
const topologyRendererModule = await import(
  pathToFileURL(path.join(root, "scripts/render-execution-topology-artifacts.mjs"))
);
const taskSurfaceModule = await import(pathToFileURL(path.join(root, "scripts/lib/task-surface.mjs")));

const {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
  topologySummary,
} = topologyModule;

const topology = loadExecutionTopology();
const summary = topologySummary(topology);
assert.equal(summary.schema_id, "cartulary.execution_topology.v3");
assert.ok(summary.execution_dependencies >= 10);
assert.ok(summary.go_targets >= 5);
assert.ok(summary.check_schedules >= 1);
assert.ok(summary.service_backed_schedules >= 3);

const renderedTaskSurface = renderTaskSurfaceManifest(topology);
const renderedBrowserBatch = renderBrowserBatchManifest(topology);
const renderedServiceBacked = serviceRendererModule.renderServiceBackedScheduleManifest({
  topology: "tools/execution_topology_manifest.json",
  topologyObject: topology,
});
const renderedCheckServiceBacked = renderedServiceBacked.schedules.find((schedule) => schedule.target === "check-service-backed");
const renderedTestServiceBacked = renderedServiceBacked.schedules.find((schedule) => schedule.target === "test-service-backed");
assert.ok(
  !(renderedCheckServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "browser-e2e-measurement"),
  "check-service-backed must exclude ordinary measurement from default local check",
);
assert.ok(
  (renderedTestServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "browser-e2e-measurement"),
  "test-service-backed must retain ordinary measurement evidence",
);
const renderedTaskSurfaceErrors = taskSurfaceModule.collectTaskSurfaceManifestErrors(renderedTaskSurface, {
  browserBatchManifest: renderedBrowserBatch,
  serviceBackedScheduleManifest: renderedServiceBacked,
});
assert.deepEqual(renderedTaskSurfaceErrors, [], "rendered task surface must satisfy task-surface validation");
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "migration-drift")?.service_requirements,
  ["postgres"],
  "migration-drift must declare its Postgres service requirement",
);
assert.equal(
  renderedTaskSurface.targets.find((target) => target.name === "agent-finalize")?.classification,
  "public",
  "agent-finalize must be a public workflow target",
);
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "agent-finalize")?.backing_scripts,
  ["scripts/agent-finalize.sh", "scripts/agent-finalize.mjs"],
  "agent-finalize must be backed by the structured end-of-run maintenance orchestrator",
);
assert.equal(
  renderedTaskSurface.make_recipes["agent-finalize"]?.type,
  "phase_command",
  "agent-finalize must be generated as a summarized Make command",
);

const taskSurfaceFixture = () => JSON.parse(JSON.stringify(renderedTaskSurface));
const taskSurfaceErrors = (manifest) =>
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(manifest, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n");
const taskSurfaceErrorList = (manifest) =>
  taskSurfaceModule.collectTaskSurfaceManifestErrors(manifest, {
    browserBatchManifest: renderedBrowserBatch,
    serviceBackedScheduleManifest: renderedServiceBacked,
  });

const invalidHarnessReference = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidHarnessReference.harness_tiers.fast.checks.push("harness-smoke-missing");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidHarnessReference, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness_tiers\.fast\.checks references unknown harness check harness-smoke-missing/,
  "task-surface validation must reject harness tiers that reference unknown checks",
);

const invalidHarnessBackingScript = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidHarnessBackingScript.harness_checks.push({
  name: "harness-smoke-missing-script",
  backing_scripts: ["scripts/missing-harness-smoke.sh"],
});
invalidHarnessBackingScript.harness_tiers.fast.checks.push("harness-smoke-missing-script");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidHarnessBackingScript, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness-smoke-missing-script backing script missing: scripts\/missing-harness-smoke\.sh/,
  "task-surface validation must reject harness checks with missing backing scripts",
);

const invalidServiceRequirement = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidServiceRequirement.targets
  .find((target) => target.name === "migration-drift")
  .service_requirements.push("legacy-db");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidServiceRequirement, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /migration-drift\.service_requirements\[2\] has invalid service requirement "legacy-db"/,
  "task-surface validation must reject unknown service requirements",
);

const invalidExportsShape = taskSurfaceFixture();
invalidExportsShape.make_recipes.help.exports = [];
assert.match(
  taskSurfaceErrors(invalidExportsShape),
  /make_recipes\.help\.exports must be an object/,
  "task-surface validation must reject non-object Make exports",
);

const invalidExportsKey = taskSurfaceFixture();
invalidExportsKey.make_recipes.help.exports = { "bad-name": "1" };
assert.match(
  taskSurfaceErrors(invalidExportsKey),
  /make_recipes\.help\.exports\.bad-name must be a safe Make variable name/,
  "task-surface validation must reject unsafe Make export names",
);

const invalidExportsValue = taskSurfaceFixture();
invalidExportsValue.make_recipes.help.exports = { SAFE_NAME: "unsafe`value" };
assert.match(
  taskSurfaceErrors(invalidExportsValue),
  /make_recipes\.help\.exports\.SAFE_NAME must be a safe Make value/,
  "task-surface validation must reject unsafe Make export values",
);

const obsoleteTestTargetExport = taskSurfaceFixture();
obsoleteTestTargetExport.make_recipes.help.exports = {
  CARTULARY_TEST_TARGET: "help",
};
assert.match(
  taskSurfaceErrors(obsoleteTestTargetExport),
  /make_recipes\.help\.exports\.CARTULARY_TEST_TARGET is obsolete; use test_target: "self"/,
  "task-surface validation must reject obsolete CARTULARY_TEST_TARGET exports",
);

const invalidEnvShape = taskSurfaceFixture();
invalidEnvShape.make_recipes.doctor.env = [];
assert.match(
  taskSurfaceErrors(invalidEnvShape),
  /make_recipes\.doctor\.env must be an object/,
  "task-surface validation must reject non-object command env",
);

const invalidEnvKey = taskSurfaceFixture();
invalidEnvKey.make_recipes.doctor.env = { "bad-name": "1" };
assert.match(
  taskSurfaceErrors(invalidEnvKey),
  /make_recipes\.doctor\.env\.bad-name must be a safe environment variable name/,
  "task-surface validation must reject unsafe command env names",
);

const invalidEnvValue = taskSurfaceFixture();
invalidEnvValue.make_recipes.doctor.env = { SAFE_NAME: "unsafe`value" };
assert.match(
  taskSurfaceErrors(invalidEnvValue),
  /make_recipes\.doctor\.env\.SAFE_NAME must be a safe Make value/,
  "task-surface validation must reject unsafe command env values",
);

const unknownRecipeType = taskSurfaceFixture();
unknownRecipeType.make_recipes.help.type = "future_recipe";
assert.match(
  taskSurfaceErrors(unknownRecipeType),
  /make_recipes\.help\.type must be one of alias, cleanup, print_help, sequence, check_schedule, go_target, service_backed_target, service_backed_schedule, browser_batch, phase_command, summary_target, node_tool/,
  "task-surface validation must reject unknown Make recipe types with the registry order",
);

const aliasWithoutTypeSpecificFields = taskSurfaceFixture();
aliasWithoutTypeSpecificFields.make_recipes.doctor = {
  type: "alias",
  prerequisites: [],
};
assert.deepEqual(
  taskSurfaceErrorList(aliasWithoutTypeSpecificFields),
  [],
  "alias recipes must not require type-specific fields beyond common preflight",
);

// Future Make recipe types must be added to makeRecipeValidators and covered here.
const recipeValidationCases = [
  {
    type: "cleanup",
    target: "clean",
    mutate: (recipe) => {
      recipe.scope = "scrub";
    },
    pattern: /make_recipes\.clean\.scope must be clean or distclean/,
  },
  {
    type: "print_help",
    target: "help",
    mutate: (recipe) => {
      recipe.scope = "wide";
    },
    pattern: /make_recipes\.help\.scope must be compact or all/,
  },
  {
    type: "sequence",
    target: "test-fast",
    mutate: (recipe) => {
      recipe.sequence = "missing-sequence";
    },
    pattern: /make_recipes\.test-fast\.sequence references unknown sequence "missing-sequence"/,
  },
  {
    type: "check_schedule",
    target: "check",
    mutate: (recipe) => {
      recipe.manifest_variable = "bad-name";
    },
    pattern: /make_recipes\.check\.manifest_variable must be a safe Make variable name/,
  },
  {
    type: "go_target",
    target: "backend-unit",
    mutate: (recipe) => {
      recipe.env = [];
    },
    pattern: /make_recipes\.backend-unit\.env must be an object/,
  },
  {
    type: "service_backed_target",
    target: "backend-store",
    mutate: (recipe) => {
      recipe.env = [];
    },
    pattern: /make_recipes\.backend-store\.env must be an object/,
  },
  {
    type: "service_backed_schedule",
    target: "test-service-backed",
    mutate: (recipe) => {
      recipe.phase_label = "";
    },
    pattern: /make_recipes\.test-service-backed\.phase_label must be a non-empty string/,
  },
  {
    type: "browser_batch",
    target: "browser-e2e",
    mutate: (recipe) => {
      recipe.stage = "bad`stage";
    },
    pattern: /make_recipes\.browser-e2e\.stage must be a safe browser stage name/,
  },
  {
    type: "phase_command",
    target: "doctor",
    mutate: (recipe) => {
      recipe.mode = "shell";
    },
    pattern: /make_recipes\.doctor\.mode must be run_phase, node, or command/,
  },
  {
    type: "summary_target",
    target: "check-harness-smoke",
    mutate: (recipe) => {
      recipe.child_target = "missing-target";
    },
    pattern: /make_recipes\.check-harness-smoke\.child_target references unknown target "missing-target"/,
  },
  {
    type: "node_tool",
    target: "doctor",
    mutate: (_recipe, manifest) => {
      manifest.make_recipes.doctor = {
        type: "node_tool",
        prerequisites: [],
      };
    },
    pattern: /make_recipes\.doctor has no scripts\/lib\/make-node-tools\.mjs registry entry/,
  },
];
for (const validationCase of recipeValidationCases) {
  const invalidRecipe = taskSurfaceFixture();
  validationCase.mutate(
    invalidRecipe.make_recipes[validationCase.target],
    invalidRecipe,
  );
  assert.match(
    taskSurfaceErrors(invalidRecipe),
    validationCase.pattern,
    `task-surface validation must reject invalid ${validationCase.type} recipe fields`,
  );
}

const invalidCheckScheduleOrder = taskSurfaceFixture();
Object.assign(invalidCheckScheduleOrder.make_recipes.check, {
  target: "missing-schedule",
  summary_profile: "legacy-summary",
  manifest_variable: "bad-name",
  schedule_manifest: "bad`path.json",
  resource_limits: { cpu: 1 },
});
const checkScheduleErrors = taskSurfaceErrorList(invalidCheckScheduleOrder);
assert.deepEqual(
  checkScheduleErrors.filter((error) => error.startsWith("make_recipes.check.")),
  [
    'make_recipes.check.target references unknown schedule target "missing-schedule"',
    "make_recipes.check.summary_profile is obsolete; summary targets derive from the check schedule",
    "make_recipes.check.manifest_variable must be a safe Make variable name",
    "make_recipes.check.schedule_manifest must be a safe repo-local JSON path",
    "make_recipes.check.resource_limits is obsolete; scheduler capacity overrides come from the resource registry",
  ],
  "check_schedule recipe validation errors must remain deterministic",
);

const serialize = (value) => `${JSON.stringify(value, null, 2)}\n`;
const artifactSnapshot = () => ({
  taskSurface: serialize(renderTaskSurfaceManifest(topology)),
  browserBatch: serialize(renderBrowserBatchManifest(topology)),
  checkSchedule: serialize(renderCheckScheduleManifest(topology)),
  serviceBacked: serialize(
    serviceRendererModule.renderServiceBackedScheduleManifest({
      topology: "tools/execution_topology_manifest.json",
      topologyObject: topology,
    }),
  ),
  taskSurfaceMake: taskSurfaceModule.renderTaskSurfaceMake(renderTaskSurfaceManifest(topology)),
});
assert.deepEqual(artifactSnapshot(), artifactSnapshot(), "topology artifact rendering must be deterministic");

const renderedCheckSchedule = renderCheckScheduleManifest(topology);
const checkSchedule = renderedCheckSchedule.schedules.find((schedule) => schedule.target === "check");
assert.ok(checkSchedule, "rendered check schedule must include check");
assert.equal(checkSchedule.work_units.length, 35, "check schedule must render the current check work-unit set");
assert.deepEqual(
  checkSchedule.work_units.find((unit) => unit.target === "lint-shell")?.env,
  { LINT_SHELL_STRICT: "1" },
  "scheduled lint-shell must run ShellCheck in strict mode",
);
assert.deepEqual(
  checkSchedule.work_units.map((unit) => [unit.target, unit.priority]),
  [
    ["toolchain-drift", 50000],
    ["codegen-toolchain", 49900],
    ["go-lint-toolchain", 49800],
    ["govulncheck-toolchain", 49700],
    ["gosec-toolchain", 49600],
    ["shell-lint-toolchain", 49500],
    ["check-frontend-install", 49400],
    ["build-server", 40000],
    ["build-migrate", 39000],
    ["test-service-images", 38000],
    ["check-service-backed", 30000],
    ["migration-drift", 27000],
    ["deployable-shape", 26000],
    ["backend-unit", 25000],
    ["frontend-typecheck", 24000],
    ["lint-go", 23000],
    ["go-vulncheck", 22000],
    ["go-gosec-targeted", 21900],
    ["go-gosec-audit", 21800],
    ["frontend-unit", 15000],
    ["check-harness-smoke", 14000],
    ["lint-biome", 13000],
    ["harness-contract-tests", 12980],
    ["frontend-import-boundary-check", 12950],
    ["json-shape-check", 12925],
    ["lint-scripts", 12900],
    ["lint-shell", 12850],
    ["phase-test-name-check", 12000],
    ["phase-map-check", 11500],
    ["go-test-duration-baseline-coverage", 11400],
    ["phase-ledger-drift", 11300],
    ["phase-schedule-drift", 11200],
    ["service-backed-unit-check", 11100],
    ["generated-artifact-policy-check", 11050],
    ["generate-drift", 11000],
  ],
  "profile-expanded check schedule must preserve setup fanout and DAG priority order",
);
assert.ok(
  checkSchedule.work_units.every((unit) => Number.isInteger(unit.weight_ms) && unit.weight_ms > 0),
  "profile-expanded check schedule must assign advisory weight_ms values",
);

const renderedExpandedCheckSchedule = renderCheckScheduleManifest(topology, {
  serviceBackedScheduleManifest: renderedServiceBacked,
  expandServiceBackedScheduleForCheck: checkServiceBackedExpansionModule.expandServiceBackedScheduleForCheck,
});
const expandedCheckSchedule = renderedExpandedCheckSchedule.schedules.find((schedule) => schedule.target === "check");
assert.ok(expandedCheckSchedule, "expanded check schedule must include check");
const expandedUnit = (id) => expandedCheckSchedule.work_units.find((unit) => unit.id === id);
const serviceSessionUnit = expandedUnit("check-service-backed:service-session");
assert.deepEqual(
  serviceSessionUnit?.needs,
  ["test-service-images"],
  "check service session must start after service images and before build artifacts finish",
);
const webserverStageSession = expandedUnit("check-service-backed:browser-stage-session:webserver-backed");
assert.deepEqual(
  webserverStageSession?.needs,
  ["service_session:check-service-backed", "build-server", "build-migrate"],
  "browser stage sessions must wait for service readiness and browser build artifacts",
);
const backendShard = expandedCheckSchedule.work_units.find(
  (unit) => unit.kind === "go_shard" && unit.target === "backend-store",
);
assert.deepEqual(
  backendShard?.needs,
  ["service_session:check-service-backed"],
  "backend service-backed shards must depend only on the ready service session",
);
assert.deepEqual(
  webserverStageSession?.retained_resource_claims,
  {
    browser_stack: 1,
    browser_stage_webserver_backed: 1,
    process: 1,
  },
  "browser stage retained claims must model only live browser stack ownership",
);
const measurementStageSession = expandedUnit("check-service-backed:browser-stage-session:measurement");
assert.equal(
  measurementStageSession,
  undefined,
  "default local check must not expand the ordinary measurement browser stage",
);

const rows = targetPlanModule.collectTargetPlanRows(root);
const scheduledBackendTargets = (rowsToUse) =>
  Array.from(
    new Set(
      rowsToUse
        .filter((row) => row.runner_family === "go_test" && row.service_backed && row.check_service_backed_safe)
        .map((row) => row.target),
    ),
  ).sort();
const originalBackendTargets = scheduledBackendTargets(rows);
const existingRow = rows.find((row) => row.execution_dependency === "backend_store");
assert.ok(existingRow, "test requires an existing backend_store row");
const futureRows = [
  ...rows,
  {
    ...existingRow,
    id: "U-99-01",
    manifest_phase: "phase99",
  },
];
assert.deepEqual(
  scheduledBackendTargets(futureRows),
  originalBackendTargets,
  "a future phase row under an existing execution dependency must not change service-backed schedule targets",
);

const tempDir = mkdtempSync(path.join(os.tmpdir(), "cartulary-topology-test-"));
copyFileSync(
  path.join(root, "tools/service_backed_make_target_duration_baselines.json"),
  path.join(tempDir, "service_backed_make_target_duration_baselines.json"),
);
const topologyFixture = () => JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"));
const writeTopologyFixture = (name, value) => {
  const fixturePath = path.join(tempDir, name);
  writeFileSync(fixturePath, `${JSON.stringify(value, null, 2)}\n`);
  return fixturePath;
};

const cliRenderDir = mkdtempSync(path.join(root, "tmp", "topology-cli-render-test-"));
const cliRenderOutputPath = (name) => path.relative(root, path.join(cliRenderDir, name)).split(path.sep).join("/");
const cliRenderTopology = topologyFixture();
cliRenderTopology.generated_outputs = {
  ...cliRenderTopology.generated_outputs,
  task_surface_manifest: cliRenderOutputPath("task_surface_manifest.json"),
  task_surface_make: cliRenderOutputPath("task_surface.generated.mk"),
  scheduler_manifest: cliRenderOutputPath("scheduler_manifest.json"),
  browser_e2e_batch_manifest: cliRenderOutputPath("browser_e2e_batch_manifest.json"),
  execution_topology_render_index: cliRenderOutputPath("execution_topology_render_index.json"),
};
const cliRenderTopologyPath = path.join(cliRenderDir, "execution_topology_manifest.json");
copyFileSync(
  path.join(root, "tools/service_backed_make_target_duration_baselines.json"),
  path.join(cliRenderDir, "service_backed_make_target_duration_baselines.json"),
);
writeFileSync(cliRenderTopologyPath, `${JSON.stringify(cliRenderTopology, null, 2)}\n`);
const renderScript = path.join(root, "scripts/render-execution-topology-artifacts.mjs");
const firstRenderOutput = execFileSync(process.execPath, [renderScript, "--topology", cliRenderTopologyPath], {
  encoding: "utf8",
});
assert.match(
  firstRenderOutput,
  /phase-schedules: updated 5 files/,
  "phase-schedules must report updated generated artifacts",
);
assert.doesNotThrow(
  () => topologyRendererModule.quickCheckRenderIndex({ topology: cliRenderTopologyPath }),
  "phase-schedule drift must pass after rendering generated artifacts",
);
const secondRenderOutput = execFileSync(process.execPath, [renderScript, "--topology", cliRenderTopologyPath], {
  encoding: "utf8",
});
assert.equal(
  secondRenderOutput.trim(),
  "phase-schedules: unchanged",
  "phase-schedules must be idempotent on a second run",
);

const renderIndexDir = mkdtempSync(path.join(root, "tmp", "topology-render-index-test-"));
const renderOutputPath = (name) => path.relative(root, path.join(renderIndexDir, name)).split(path.sep).join("/");
const renderIndexTopology = topologyFixture();
renderIndexTopology.generated_outputs = {
  ...renderIndexTopology.generated_outputs,
  task_surface_manifest: renderOutputPath("task_surface_manifest.json"),
  task_surface_make: renderOutputPath("task_surface.generated.mk"),
  scheduler_manifest: renderOutputPath("scheduler_manifest.json"),
  browser_e2e_batch_manifest: renderOutputPath("browser_e2e_batch_manifest.json"),
  execution_topology_render_index: renderOutputPath("execution_topology_render_index.json"),
};
const renderIndexBaselinePath = path.join(renderIndexDir, "service_backed_make_target_duration_baselines.json");
copyFileSync(path.join(root, "tools/service_backed_make_target_duration_baselines.json"), renderIndexBaselinePath);
renderIndexTopology.service_backed_schedules = {
  ...renderIndexTopology.service_backed_schedules,
  defaults: {
    ...renderIndexTopology.service_backed_schedules.defaults,
    make_target_duration_baseline: renderIndexBaselinePath,
  },
};
const renderIndexTopologyPath = path.join(renderIndexDir, "execution_topology_manifest.json");
writeFileSync(renderIndexTopologyPath, `${JSON.stringify(renderIndexTopology, null, 2)}\n`);
const renderArtifacts = [
  { file: renderIndexTopology.generated_outputs.task_surface_manifest, content: "not json\n" },
  { file: renderIndexTopology.generated_outputs.task_surface_make, content: "# generated make fixture\n" },
  { file: renderIndexTopology.generated_outputs.scheduler_manifest, content: "not json\n" },
  { file: renderIndexTopology.generated_outputs.browser_e2e_batch_manifest, content: "not json\n" },
];
for (const artifact of renderArtifacts) {
  writeFileSync(path.join(root, artifact.file), artifact.content);
}
const phaseRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-render-phase-root-"));
mkdirSync(path.join(phaseRoot, "tools"), { recursive: true });
copyFileSync(path.join(root, "tools/phase_registry.json"), path.join(phaseRoot, "tools/phase_registry.json"));
for (const entry of JSON.parse(readFileSync(path.join(root, "tools/phase_registry.json"), "utf8")).phases) {
  if (entry.status === "active") {
    copyFileSync(path.join(root, entry.manifest_path), path.join(phaseRoot, entry.manifest_path));
  }
}
const oldPhaseRoot = process.env.CARTULARY_PHASE_MANIFEST_ROOT;
process.env.CARTULARY_PHASE_MANIFEST_ROOT = phaseRoot;
try {
  const writeRenderIndex = () => {
    const inputInfo = topologyRendererModule.collectRenderInputs({ topology: renderIndexTopologyPath });
    writeFileSync(
      path.join(root, renderIndexTopology.generated_outputs.execution_topology_render_index),
      `${JSON.stringify(topologyRendererModule.buildRenderIndex({ inputInfo, artifacts: renderArtifacts }), null, 2)}\n`,
    );
  };
  writeRenderIndex();
  assert.doesNotThrow(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath }),
    "quick phase-schedule drift must accept a fresh index without rendering output content",
  );
  writeFileSync(path.join(root, renderArtifacts[0].file), "changed\n");
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath }),
    /stale; run make phase-schedules/,
    "quick phase-schedule drift must reject changed generated outputs by hash",
  );
  writeFileSync(path.join(root, renderArtifacts[0].file), renderArtifacts[0].content);
  const changedTopology = { ...renderIndexTopology, schema_id: "cartulary.execution_topology.v3" };
  changedTopology.task_surface = {
    ...changedTopology.task_surface,
    compact_help: {
      ...changedTopology.task_surface.compact_help,
      entries: [
        ...changedTopology.task_surface.compact_help.entries,
        { target: "future", description: "future fixture" },
      ],
    },
  };
  writeFileSync(renderIndexTopologyPath, `${JSON.stringify(changedTopology, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath }),
    /phase schedule inputs are stale.*execution_topology_manifest\.json changed.*run make phase-schedules/,
    "quick phase-schedule drift must reject changed topology input by digest",
  );
  writeFileSync(renderIndexTopologyPath, `${JSON.stringify(renderIndexTopology, null, 2)}\n`);
  const baselineFixture = JSON.parse(readFileSync(renderIndexBaselinePath, "utf8"));
  baselineFixture.default_work_unit_weight_ms += 1;
  writeFileSync(renderIndexBaselinePath, `${JSON.stringify(baselineFixture, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath }),
    /phase schedule inputs are stale.*service_backed_make_target_duration_baselines\.json changed.*run make phase-schedules/,
    "quick phase-schedule drift must reject changed duration baselines by digest",
  );
  copyFileSync(path.join(root, "tools/service_backed_make_target_duration_baselines.json"), renderIndexBaselinePath);
  const phase1FixturePath = path.join(phaseRoot, "tools/phase1_test_map.json");
  const phase1Fixture = JSON.parse(readFileSync(phase1FixturePath, "utf8"));
  phase1Fixture.notes = [...(phase1Fixture.notes ?? []), "render index drift fixture"];
  writeFileSync(phase1FixturePath, `${JSON.stringify(phase1Fixture, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath }),
    /phase schedule inputs are stale.*phase1_test_map\.json changed.*run make phase-schedules/,
    "quick phase-schedule drift must reject changed active phase maps by digest",
  );
} finally {
  if (oldPhaseRoot === undefined) {
    delete process.env.CARTULARY_PHASE_MANIFEST_ROOT;
  } else {
    process.env.CARTULARY_PHASE_MANIFEST_ROOT = oldPhaseRoot;
  }
}

const invalidTopology = topologyFixture();
invalidTopology.execution_dependencies.push({ ...invalidTopology.execution_dependencies[0] });
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("invalid-topology.json", invalidTopology) }),
  /duplicate execution dependency/,
  "topology validation must reject duplicate execution dependency IDs",
);

const legacyFlatTopology = topologyFixture();
legacyFlatTopology.check_schedules = renderedCheckSchedule.schedules;
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("legacy-flat-topology.json", legacyFlatTopology) }),
  /flat check_schedules\[\] work_units are no longer supported/,
  "topology validation must reject the obsolete flat check schedule source shape",
);

const unknownProfileTopology = topologyFixture();
unknownProfileTopology.task_surface.targets.find((target) => target.name === "generate-drift").check_schedule.profile =
  "missing_profile";
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("unknown-check-profile-topology.json", unknownProfileTopology),
    }),
  /generate-drift\.check_schedule\.profile references unknown profile missing_profile/,
  "topology validation must reject unknown check schedule profiles",
);

const profileNeedsTopology = topologyFixture();
profileNeedsTopology.check_schedules.defaults.resource_profiles.after_setup_cpu.needs = ["toolchain-drift"];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("profile-needs-topology.json", profileNeedsTopology),
    }),
  /check_schedules\.defaults\.resource_profiles\.after_setup_cpu has unknown key needs/,
  "topology validation must keep check schedule dependency edges on targets instead of profiles",
);

const unknownNeedTopology = topologyFixture();
unknownNeedTopology.task_surface.targets.find((target) => target.name === "backend-unit").check_schedule.needs = [
  "missing-setup",
];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("unknown-check-need-topology.json", unknownNeedTopology),
    }),
  /check schedule check work unit backend-unit depends on unknown missing-setup/,
  "topology validation must reject check schedule needs that reference unknown work units",
);

const duplicateOrderTopology = topologyFixture();
duplicateOrderTopology.task_surface.targets.find((target) => target.name === "phase-schedule-drift").check_schedule.order =
  0;
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("duplicate-check-order-topology.json", duplicateOrderTopology),
    }),
  /duplicate priority order drift_validation:0/,
  "topology validation must reject duplicate check schedule priority orders within a band",
);

const mismatchedSummaryTargetTopology = topologyFixture();
mismatchedSummaryTargetTopology.task_surface.targets.find(
  (target) => target.name === "frontend-unit",
).check_schedule.produces_summary_targets = ["frontend-typecheck"];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("mismatched-summary-target-topology.json", mismatchedSummaryTargetTopology),
    }),
  /frontend-unit\.check_schedule\.produces_summary_targets must include owning target frontend-unit/,
  "topology validation must reject check work units that omit their own summary target",
);

const missingServiceBoundaryTopology = topologyFixture();
delete missingServiceBoundaryTopology.check_schedules.defaults.resource_profiles.post_build_migration_scratch_postgres
  .resource_claims.migration_scratch_postgres;
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("missing-service-boundary-topology.json", missingServiceBoundaryTopology),
    }),
  /migration-drift\.check_schedule target declares service_requirements and must claim a check service boundary resource/,
  "topology validation must require a boundary resource for service-backed check schedule profiles",
);

const schedulerOwnedEnvTopology = topologyFixture();
schedulerOwnedEnvTopology.task_surface.targets.find((target) => target.name === "lint-shell").check_schedule.env = {
  CARTULARY_TEST_TARGET: "lint-shell",
};
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("scheduler-owned-env-topology.json", schedulerOwnedEnvTopology),
    }),
  /lint-shell\.check_schedule\.env\.CARTULARY_TEST_TARGET is scheduler-owned and cannot be overridden/,
  "topology validation must reject scheduler-owned check work-unit env",
);

const futureCheckTargetTopology = topologyFixture();
futureCheckTargetTopology.task_surface.targets.push({
  name: "future-phase-check-leaf",
  classification: "check_internal",
  included_in: ["check"],
  check_schedule: {
    schedules: ["check"],
    profile: "after_setup_cpu",
    priority_band: "phase_validation",
    order: 700,
  },
});
const futureCheckSchedule = renderCheckScheduleManifest(
  loadExecutionTopology({
    manifestPath: writeTopologyFixture("future-check-target-topology.json", futureCheckTargetTopology),
  }),
).schedules.find((schedule) => schedule.target === "check");
assert.ok(
  futureCheckSchedule.work_units.some((unit) => unit.target === "future-phase-check-leaf"),
  "new check-scheduled targets must be included through metadata without adding flat work units",
);
EOF
