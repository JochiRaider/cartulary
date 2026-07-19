#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { chmodSync, copyFileSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root] = process.argv.slice(2);
process.chdir(root);

const topologyModule = await import(pathToFileURL(path.join(root, "tools/harness/generated-artifacts/execution-topology.mjs")));
const targetPlanModule = await import(pathToFileURL(path.join(root, "tools/harness/backend/backend-target-plan.mjs")));
const goShardPlanModule = await import(pathToFileURL(path.join(root, "tools/harness/backend/backend-shard-plan.mjs")));
const serviceRendererModule = await import(
  pathToFileURL(path.join(root, "tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs"))
);
const checkServiceBackedExpansionModule = await import(
  pathToFileURL(path.join(root, "tools/harness/execution/service-backed/schedule-planning.mjs"))
);
const topologyRendererModule = await import(
  pathToFileURL(path.join(root, "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs"))
);
const taskSurfaceModule = await import(pathToFileURL(path.join(root, "tools/harness/generated-artifacts/task-surface/index.mjs")));
const runtimeBinaryRegistryModule = await import(pathToFileURL(path.join(root, "tools/harness/runtime-binary-registry.mjs")));

const {
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderCheckScheduleManifest,
  renderTaskSurfaceManifest,
  topologySummary,
} = topologyModule;

const topology = loadExecutionTopology();
const summary = topologySummary(topology);
assert.equal(summary.schema_id, "cartulary.execution_topology.v4");
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
assert.doesNotThrow(
  () => checkServiceBackedExpansionModule.validateServiceBackedScheduleManifestShape(renderedServiceBacked, "rendered service-backed schedule"),
  "rendered service-backed sources must satisfy the service-backed source validator",
);
const operatorRuntimeRecord = topology.runtimeBinaries.find((entry) => entry.id === "operator");
assert.equal(
  operatorRuntimeRecord?.default_output_path,
  "operator",
  "operator runtime binary registry row must declare its scheduler-owned default output path",
);
const syntheticRuntimeRegistry = runtimeBinaryRegistryModule.runtimeBinaryRegistry([
  ...topology.runtimeBinaries,
  {
    id: "synthetic-helper",
    producer_target: "build-migrate",
    output_make_variable: "MIGRATE_BIN",
    consumer_env: "CARTULARY_SYNTHETIC_HELPER_BIN",
    default_output_path: "migrate",
  },
]);
assert.deepEqual(
  runtimeBinaryRegistryModule.runtimeBinaryProducerTargetsForIDs(
    syntheticRuntimeRegistry,
    ["synthetic-helper"],
  ),
  ["build-migrate"],
  "runtime-binary producer derivation must be data-driven for future registry rows",
);
assert.deepEqual(
  runtimeBinaryRegistryModule.runtimeBinaryDefaultEnvForIDs(
    syntheticRuntimeRegistry,
    ["synthetic-helper"],
  ),
  { CARTULARY_SYNTHETIC_HELPER_BIN: "migrate" },
  "runtime-binary env derivation must use registry default_output_path for future rows",
);
const camelCaseBrowserSource = JSON.parse(JSON.stringify(renderedServiceBacked));
camelCaseBrowserSource.schedules[0].work_unit_sources.push({
  type: "browser_stage",
  class: "browser",
  target: "browser-e2e-webserver-backed",
  browser_stage: "webserver-backed",
  browserSessionGroup: "legacy-camel-case",
  priority: 1,
  weight_ms: 1,
  resource_claims: { browser_stack: 1 },
  groups: [
    {
      id: "camel-case-fixture",
      name: "camel-case-fixture",
      kind: "support",
      target: "browser-e2e-webserver-backed",
      aggregate_target: "browser-e2e-webserver-backed",
      weight_ms: 1,
    },
  ],
});
assert.throws(
  () => checkServiceBackedExpansionModule.validateServiceBackedScheduleManifestShape(camelCaseBrowserSource, "camelCase fixture"),
  /unknown key browserSessionGroup/,
  "service-backed source validator must reject camelCase browser session aliases",
);
assert.ok(
  !(renderedCheckServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "browser-e2e-measurement"),
  "check-service-backed must exclude ordinary measurement from default local check",
);
assert.ok(
  !(renderedCheckServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "backend-integration-support"),
  "check-service-backed must exclude support-only backend integration evidence from default local check",
);
assert.ok(
  (renderedTestServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "browser-e2e-measurement"),
  "test-service-backed must retain ordinary measurement evidence",
);
const renderedTestMeasurement = (renderedTestServiceBacked?.work_unit_sources ?? []).find(
  (source) => source.browser_stage === "measurement",
);
assert.deepEqual(
  renderedTestMeasurement?.needs,
  [
    "browser-e2e-webserver-backed",
    "browser-e2e-stateful",
    "browser-e2e-visual",
    "browser-e2e-a11y",
  ],
  "ordinary service-backed measurement evidence must depend on every selected full-fidelity peer browser stage",
);
assert.ok(
  (renderedTestServiceBacked?.work_unit_sources ?? []).some((source) => source.target === "backend-integration-support"),
  "test-service-backed must retain explicit support-only backend integration evidence",
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
for (const targetName of ["backend-integration-support", "test-fast-service-backed", "browser-e2e"]) {
  assert.ok(
    !(renderedTaskSurface.targets.find((target) => target.name === targetName)?.default_inclusion_sets ?? []).includes("check"),
    `${targetName} must not advertise default check membership unless check actually schedules it`,
  );
}
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-webserver-backed")?.check_projection,
  {
    mode: "projection",
    schedule: "check-service-backed",
    stage: "webserver-backed",
    evidence: "default_local_cross_stack_conformance",
    evidence_class: "product_conformance",
    reason_code: "lower_layer_gap",
    full_target: "browser-e2e-webserver-backed",
    full_target_equivalent: false,
  },
  "browser-e2e-webserver-backed must label its default-check cross-stack projection",
);
assert.ok(
  !(renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-webserver-backed")?.default_inclusion_sets ?? []).includes("check"),
  "browser-e2e-webserver-backed must not advertise direct check membership when check uses a projection",
);
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-stateful")?.check_projection,
  {
    mode: "direct",
    schedule: "check-service-backed",
    stage: "stateful",
    evidence: "stateful_browser_conformance",
    evidence_class: "product_conformance",
    reason_code: "full_target_equivalent_stateful",
    full_target: "browser-e2e-stateful",
    full_target_equivalent: true,
  },
  "browser-e2e-stateful must label its full-target-equivalent default-check evidence",
);
assert.ok(
  (renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-stateful")?.default_inclusion_sets ?? []).includes("check"),
  "browser-e2e-stateful must advertise direct check membership for full-target-equivalent scheduler evidence",
);
assert.ok(
  renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-visual")?.check_projection === undefined,
  "browser-e2e-visual must remain explicit-only for default local check",
);
assert.ok(
  renderedTaskSurface.targets.find((target) => target.name === "browser-e2e-a11y")?.check_projection === undefined,
  "browser-e2e-a11y must remain explicit-only for default local check",
);
assert.equal(
  renderedTaskSurface.targets.find((target) => target.name === "agent-finalize")?.target_class,
  "public",
  "agent-finalize must be a public workflow target",
);
assert.deepEqual(
  renderedTaskSurface.targets.find((target) => target.name === "agent-finalize")?.backing_scripts,
  ["tools/harness/finalization/agent-finalize-cli.mjs", "tools/harness/finalization/agent-finalize-action-cache.mjs"],
  "agent-finalize must be backed by the structured end-of-run maintenance orchestrator",
);
assert.equal(
  renderedTaskSurface.make_recipes["agent-finalize"]?.type,
  "step_command",
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
  backing_scripts: ["tools/harness/generated-artifacts/tests/missing-harness-smoke.sh"],
});
invalidHarnessBackingScript.harness_tiers.fast.checks.push("harness-smoke-missing-script");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidHarnessBackingScript, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness-smoke-missing-script backing script missing: tools\/harness\/generated-artifacts\/tests\/missing-harness-smoke\.sh/,
  "task-surface validation must reject harness checks with missing backing scripts",
);

const invalidRootHarnessBackingScript = JSON.parse(JSON.stringify(renderedTaskSurface));
invalidRootHarnessBackingScript.harness_checks.push({
  name: "harness-smoke-root-script",
  backing_scripts: ["scripts/root-harness-smoke.sh"],
});
invalidRootHarnessBackingScript.harness_tiers.fast.checks.push("harness-smoke-root-script");
assert.match(
  taskSurfaceModule
    .collectTaskSurfaceManifestErrors(invalidRootHarnessBackingScript, {
      browserBatchManifest: renderedBrowserBatch,
      serviceBackedScheduleManifest: renderedServiceBacked,
    })
    .join("\n"),
  /harness-smoke-root-script\.backing_scripts must not reference retired root scripts\/ path scripts\/root-harness-smoke\.sh/,
  "task-surface validation must reject root scripts backing paths",
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
  /make_recipes\.help\.type must be one of artifact_binding, aggregate, readiness_projection, cleanup, print_help, sequence, check_schedule, go_target, service_backed_target, service_backed_schedule, browser_batch, step_command, owner_command, summary_target, node_tool/,
  "task-surface validation must reject unknown Make recipe types with the registry order",
);

const unsupportedAlias = taskSurfaceFixture();
unsupportedAlias.make_recipes.doctor = {
  type: "alias",
  prerequisites: [],
};
assert.match(
  taskSurfaceErrors(unsupportedAlias),
  /make_recipes\.doctor\.type must be one of artifact_binding/,
  "legacy alias recipes must be rejected",
);

const emptyArtifactBinding = taskSurfaceFixture();
emptyArtifactBinding.make_recipes["frontend-install"].prerequisites = [];
assert.match(
  taskSurfaceErrors(emptyArtifactBinding),
  /make_recipes\.frontend-install\.prerequisites must name at least one artifact producer/,
  "artifact bindings must name an artifact producer",
);

const unknownAggregateChild = taskSurfaceFixture();
unknownAggregateChild.make_recipes.build.prerequisites = ["missing-child"];
assert.match(
  taskSurfaceErrors(unknownAggregateChild),
  /make_recipes\.build\.prerequisites references unknown aggregate child missing-child/,
  "aggregates must reference declared children",
);

const publicReadinessProjection = taskSurfaceFixture();
publicReadinessProjection.make_recipes.doctor = {
  type: "readiness_projection",
  prerequisites: ["frontend-install"],
};
assert.match(
  taskSurfaceErrors(publicReadinessProjection),
  /make_recipes\.doctor readiness_projection must remain internal/,
  "readiness projections must remain internal",
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
      recipe.step_label = "";
    },
    pattern: /make_recipes\.test-service-backed\.step_label must be a non-empty string/,
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
    type: "step_command",
    target: "doctor",
    mutate: (recipe) => {
      recipe.mode = "shell";
    },
    pattern: /make_recipes\.doctor\.mode must be run_step, node, or command/,
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
    pattern: /make_recipes\.doctor has no tools\/harness\/command-surface\/make-node-tools\.mjs registry entry/,
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
  taskSurfaceRuntimeMake: taskSurfaceModule.renderTaskSurfaceMakeRuntime(
    renderTaskSurfaceManifest(topology),
  ),
});
assert.deepEqual(artifactSnapshot(), artifactSnapshot(), "topology artifact rendering must be deterministic");
const density = taskSurfaceModule.taskSurfaceMakeDensity(renderTaskSurfaceManifest(topology));
assert.equal(density.synthetic_target_count, 25, "density guard must model 25 future targets");
assert.ok(
  density.synthetic_average_growth_bytes <= 512,
  `future target growth ${density.synthetic_average_growth_bytes} must remain <= 512 bytes per target`,
);
assert.ok(density.maximum_line_bytes <= 512, "generated Make lines must remain bounded");

const renderedCheckSchedule = renderCheckScheduleManifest(topology);
const checkSchedule = renderedCheckSchedule.schedules.find((schedule) => schedule.target === "check");
assert.ok(checkSchedule, "rendered check schedule must include check");
const expectedCheckWorkUnitPriorities = [
  ["toolchain-drift", 50000],
  ["codegen-toolchain", 49900],
  ["go-lint-toolchain", 49800],
  ["govulncheck-toolchain", 49700],
  ["gosec-toolchain", 49600],
  ["shell-lint-toolchain", 49500],
  ["check-frontend-install", 49400],
  ["build-server", 40000],
  ["build-server-harness", 39900],
  ["embedded-web-assets", 39750],
  ["build-web", 39500],
  ["build-migrate", 39000],
  ["build-operator", 38500],
  ["testservices-build", 38050],
  ["test-service-images", 38000],
  ["check-service-backed", 37000],
  ["migration-scratch-apply", 27000],
  ["backend-unit", 25000],
  ["frontend-typecheck", 24000],
  ["lint-go", 23000],
  ["go-vulncheck", 22000],
  ["go-gosec-targeted", 21900],
  ["frontend-unit", 15000],
  ["check-harness-smoke", 14000],
  ["lint-biome", 13000],
  ["frontend-import-boundary-check", 12950],
  ["backend-module-boundary-check", 12945],
  ["otel-conformance", 12940],
  ["json-shape-check", 12925],
  ["lint-scripts", 12900],
  ["lint-markdown", 12875],
  ["lint-shell", 12850],
  ["semantic-identity-check", 12000],
  ["test-catalog-check", 11500],
  ["go-test-duration-baseline-coverage", 11400],
  ["generated-artifact-policy-check", 11050],
  ["generate-drift", 11000],
  ["migration-input-drift", 10900],
];
assert.equal(
  checkSchedule.work_units.length,
  expectedCheckWorkUnitPriorities.length,
  "check schedule must render the current check work-unit set",
);
assert.ok(
  checkSchedule.work_units.every((unit) => unit.local_input_stamp === undefined),
  "check schedule must not render retired local_input_stamp metadata",
);
assert.ok(
  checkSchedule.work_units.every((unit) => ["run", "skip"].includes(unit.make_prerequisite_policy)),
  "check schedule must render explicit make prerequisite policy values",
);
const checkUnitByTarget = new Map(checkSchedule.work_units.map((unit) => [unit.target, unit]));
assert.deepEqual(
  checkUnitByTarget.get("build-server")?.needs,
  ["embedded-web-assets"],
  "scheduled build-server must depend on embedded web asset readiness",
);
assert.equal(
  checkUnitByTarget.get("build-server")?.make_prerequisite_policy,
  "run",
  "scheduled build-server must run its Make prerequisites to prove the server binary",
);
assert.deepEqual(
  checkUnitByTarget.get("build-server-harness")?.needs,
  ["embedded-web-assets"],
  "scheduled build-server-harness must depend on its complete embedded asset input",
);
assert.equal(
  checkUnitByTarget.get("build-server-harness")?.make_prerequisite_policy,
  "run",
  "scheduled build-server-harness must run its Make prerequisites to prove the harness binary",
);
assert.deepEqual(
  checkUnitByTarget.get("embedded-web-assets")?.needs,
  ["build-web"],
  "scheduled embedded-web-assets must depend on build-web as its frontend dist producer",
);
assert.equal(
  checkUnitByTarget.get("embedded-web-assets")?.make_prerequisite_policy,
  "run",
  "scheduled embedded-web-assets must run its Make prerequisites to publish the archive",
);
assert.deepEqual(
  checkUnitByTarget.get("build-web")?.needs,
  ["check-frontend-install"],
  "scheduled build-web must depend on frontend install readiness",
);
assert.equal(
  checkUnitByTarget.get("build-web")?.make_prerequisite_policy,
  "run",
  "scheduled build-web must run its Make prerequisites to produce apps/web/dist",
);
assert.deepEqual(
  checkUnitByTarget.get("otel-conformance")?.needs,
  ["toolchain-drift", "build-web"],
  "scheduled otel-conformance must wait for toolchain drift and the built web bundle",
);
assert.equal(
  checkUnitByTarget.get("otel-conformance")?.make_prerequisite_policy,
  "skip",
  "scheduled otel-conformance must rely on scheduler-modeled build-web readiness",
);
assert.deepEqual(
  checkUnitByTarget.get("backend-unit")?.needs,
  ["toolchain-drift", "embedded-web-assets"],
  "scheduled backend-unit must wait for embedded web asset readiness before Go compilation",
);
assert.deepEqual(
  checkSchedule.work_units.find((unit) => unit.target === "lint-shell")?.env,
  { LINT_SHELL_STRICT: "1" },
  "scheduled lint-shell must run ShellCheck in strict mode",
);
const frontendUnitWorkUnit = checkSchedule.work_units.find((unit) => unit.target === "frontend-unit");
assert.deepEqual(
  frontendUnitWorkUnit?.resource_claims,
  { host_cpu: 2 },
  "scheduled frontend-unit must claim the CPU capacity consumed by Vitest workers",
);
assert.deepEqual(
  frontendUnitWorkUnit?.env,
  { VITEST_MAX_WORKERS: "2" },
  "scheduled frontend-unit must pin its Vitest worker budget independently of the direct target default",
);
assert.equal(
  frontendUnitWorkUnit?.make_jobs,
  "host_cpu",
  "scheduled frontend-unit make jobs must stay tied to the claimed host_cpu budget",
);
assert.deepEqual(
  checkSchedule.work_units.map((unit) => [unit.target, unit.priority]),
  expectedCheckWorkUnitPriorities,
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
const webserverStageSession = expandedUnit("check-service-backed:browser-stage-session:default-check-browser-shared");
const serviceCompleteUnit = expandedUnit("check-service-backed:complete");
assert.equal(
  serviceSessionUnit?.priority,
  37000,
  "check service-backed session must carry the service-backed readiness priority",
);
assert.equal(
  webserverStageSession?.priority,
  36000,
  "check service-backed browser children must outrank static and drift work",
);
assert.deepEqual(
  webserverStageSession?.needs,
  [
    "service_session:check-service-backed",
    "build-web",
    "build-server-harness",
    "build-migrate",
  ],
  "browser stage sessions must wait for service readiness and their declared runtime artifacts",
);
assert.equal(
  expandedCheckSchedule.work_units.filter((unit) => unit.target === "build-server-harness").length,
  1,
  "expanded check must retain exactly one harness-server producer",
);
const backendShard = expandedCheckSchedule.work_units.find(
  (unit) => unit.kind === "go_shard" && unit.target === "backend-store",
);
assert.equal(
  backendShard?.priority,
  35000,
  "check service-backed backend shards must outrank static and drift work",
);
assert.equal(
  serviceCompleteUnit?.priority,
  34900,
  "check service-backed completion must remain above post-build local evidence",
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
    id: "module.example.unit.synthetic-row",
    owner_id: "module.future",
  },
];
assert.deepEqual(
  scheduledBackendTargets(futureRows),
  originalBackendTargets,
  "a future owner row under an existing execution dependency must not change service-backed schedule targets",
);

const tempDir = mkdtempSync(path.join(os.tmpdir(), "cartulary-topology-test-"));
copyFileSync(
  path.join(root, "tools/service_backed_make_target_duration_baselines.json"),
  path.join(tempDir, "service_backed_make_target_duration_baselines.json"),
);
const topologyFixture = () => JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"));
const taskSurfaceOwnerFixture = () => JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
const writeTopologyFixture = (name, value) => {
  const fixturePath = path.join(tempDir, name);
  writeFileSync(fixturePath, `${JSON.stringify(value, null, 2)}\n`);
  return fixturePath;
};
const writeTaskSurfaceOwnerFixture = (name, value) => {
  const fixturePath = path.join(tempDir, name);
  writeFileSync(fixturePath, `${JSON.stringify(value, null, 2)}\n`);
  return path.relative(root, fixturePath).split(path.sep).join("/");
};

const missingA11yGeneratedNeedsTopology = topologyFixture();
missingA11yGeneratedNeedsTopology.service_backed_schedules.defaults.browser_stage_generated_needs.measurement.selected_peer_stages = [
  "webserver-backed",
  "stateful",
  "visual",
];
assert.throws(
  () =>
    serviceRendererModule.renderServiceBackedScheduleManifest({
      topology: writeTopologyFixture("missing-a11y-generated-needs-topology.json", missingA11yGeneratedNeedsTopology),
    }),
  /test-service-backed browser measurement isolation must explicitly account for newly selected stage a11y/,
  "service-backed measurement policy must fail closed when a selected peer browser stage is missing",
);

const serviceRenderScript = path.join(root, "tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs");
const serviceRenderWithoutOutput = spawnSync(process.execPath, [serviceRenderScript], {
  encoding: "utf8",
});
assert.equal(
  serviceRenderWithoutOutput.status,
  2,
  "standalone service-backed schedule rendering without --output must be a usage error",
);
assert.match(
  serviceRenderWithoutOutput.stderr,
  /usage: render-service-backed-schedule-manifest\.mjs .*--output <path>/,
  "standalone service-backed schedule renderer usage must name the explicit output path",
);

const cliRenderDir = mkdtempSync(path.join(root, "tmp", "topology-cli-render-test-"));
const cliRenderOutputPath = (name) => path.relative(root, path.join(cliRenderDir, name)).split(path.sep).join("/");
const cliRenderTopology = topologyFixture();
cliRenderTopology.generated_outputs = {
  ...cliRenderTopology.generated_outputs,
  task_surface_manifest: cliRenderOutputPath("task_surface_manifest.json"),
  task_surface_make: cliRenderOutputPath("task_surface.generated.mk"),
  task_surface_runtime_make: cliRenderOutputPath("task_surface.runtime.generated.mk"),
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
const renderScript = path.join(root, "tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs");
const firstRenderOutput = execFileSync(process.execPath, [renderScript, "--topology", cliRenderTopologyPath], {
  encoding: "utf8",
});
assert.match(
  firstRenderOutput,
  /generated-topology: updated 6 files/,
  "topology generation must report updated generated artifacts",
);
assert.doesNotThrow(
  () => topologyRendererModule.quickCheckRenderIndex({ topology: cliRenderTopologyPath }),
  "generated topology drift must pass after rendering generated artifacts",
);
const secondRenderOutput = execFileSync(process.execPath, [renderScript, "--topology", cliRenderTopologyPath], {
  encoding: "utf8",
});
assert.equal(
  secondRenderOutput.trim(),
  "generated-topology: unchanged",
  "topology generation must be idempotent on a second run",
);

const renderIndexDir = mkdtempSync(path.join(root, "tmp", "topology-render-index-test-"));
const renderOutputPath = (name) => path.relative(root, path.join(renderIndexDir, name)).split(path.sep).join("/");
const renderIndexTopology = topologyFixture();
renderIndexTopology.generated_outputs = {
  ...renderIndexTopology.generated_outputs,
  task_surface_manifest: renderOutputPath("task_surface_manifest.json"),
  task_surface_make: renderOutputPath("task_surface.generated.mk"),
  task_surface_runtime_make: renderOutputPath("task_surface.runtime.generated.mk"),
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
  {
    file: renderIndexTopology.generated_outputs.task_surface_runtime_make,
    content: "# generated runtime Make fixture\n",
  },
  { file: renderIndexTopology.generated_outputs.scheduler_manifest, content: "not json\n" },
  { file: renderIndexTopology.generated_outputs.browser_e2e_batch_manifest, content: "not json\n" },
];
for (const artifact of renderArtifacts) {
  writeFileSync(path.join(root, artifact.file), artifact.content);
}
const catalogRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-render-catalog-root-"));
const copyCatalogFile = (relativeFile) => {
  const destination = path.join(catalogRoot, relativeFile);
  mkdirSync(path.dirname(destination), { recursive: true });
  copyFileSync(path.join(root, relativeFile), destination);
};
copyCatalogFile("tools/test_catalog_owner.json");
copyCatalogFile("tools/test_runner_registry.json");
const catalogOwnerRegistry = JSON.parse(
  readFileSync(path.join(root, "tools/test_catalog_owner.json"), "utf8"),
);
for (const owner of catalogOwnerRegistry.owners) copyCatalogFile(owner.manifest_path);
copyCatalogFile("contracts/verification/registry.json");
const verificationRegistry = JSON.parse(
  readFileSync(path.join(root, "contracts/verification/registry.json"), "utf8"),
);
for (const owner of verificationRegistry.owners) copyCatalogFile(owner.contract_path);
{
  const writeRenderIndex = () => {
    const inputInfo = topologyRendererModule.collectRenderInputs({
      topology: renderIndexTopologyPath,
      catalogRoot,
    });
    writeFileSync(
      path.join(root, renderIndexTopology.generated_outputs.execution_topology_render_index),
      `${JSON.stringify(topologyRendererModule.buildRenderIndex({ inputInfo, artifacts: renderArtifacts }), null, 2)}\n`,
    );
  };
  writeRenderIndex();
  assert.doesNotThrow(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath, catalogRoot }),
    "quick topology drift must accept a fresh owner-catalog index without rendering output content",
  );
  writeFileSync(path.join(root, renderArtifacts[0].file), "changed\n");
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath, catalogRoot }),
    /stale; run make generate/,
    "quick generated-topology drift must reject changed outputs by hash",
  );
  writeFileSync(path.join(root, renderArtifacts[0].file), renderArtifacts[0].content);
  const changedTopology = {
    ...renderIndexTopology,
    schema_id: "cartulary.execution_topology.v4",
    check_schedules: {
      ...renderIndexTopology.check_schedules,
      defaults: {
        ...renderIndexTopology.check_schedules.defaults,
        priority_bands: {
          ...renderIndexTopology.check_schedules.defaults.priority_bands,
          setup: renderIndexTopology.check_schedules.defaults.priority_bands.setup + 1,
        },
      },
    },
  };
  writeFileSync(renderIndexTopologyPath, `${JSON.stringify(changedTopology, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath, catalogRoot }),
    /generated topology inputs are stale.*execution_topology_manifest\.json changed.*run make generate/,
    "quick generated-topology drift must reject changed topology input by digest",
  );
  writeFileSync(renderIndexTopologyPath, `${JSON.stringify(renderIndexTopology, null, 2)}\n`);
  const baselineFixture = JSON.parse(readFileSync(renderIndexBaselinePath, "utf8"));
  baselineFixture.default_work_unit_weight_ms += 1;
  writeFileSync(renderIndexBaselinePath, `${JSON.stringify(baselineFixture, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath, catalogRoot }),
    /generated topology inputs are stale.*service_backed_make_target_duration_baselines\.json changed.*run make generate/,
    "quick generated-topology drift must reject changed duration baselines by digest",
  );
  copyFileSync(path.join(root, "tools/service_backed_make_target_duration_baselines.json"), renderIndexBaselinePath);
  const familyFixturePath = path.join(catalogRoot, catalogOwnerRegistry.owners[0].manifest_path);
  const familyFixture = JSON.parse(readFileSync(familyFixturePath, "utf8"));
  familyFixture.rows[0].documentation_refs = ["render index drift fixture"];
  writeFileSync(familyFixturePath, `${JSON.stringify(familyFixture, null, 2)}\n`);
  assert.throws(
    () => topologyRendererModule.quickCheckRenderIndex({ topology: renderIndexTopologyPath, catalogRoot }),
    /generated topology inputs are stale.*app\.operator\.json changed.*run make generate/,
    "quick topology drift must reject changed owner family manifests by digest",
  );
}

const invalidTopology = topologyFixture();
invalidTopology.execution_dependencies.push({ ...invalidTopology.execution_dependencies[0] });
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("invalid-topology.json", invalidTopology) }),
  /duplicate execution dependency/,
  "topology validation must reject duplicate execution dependency IDs",
);

const unknownFamilyRuntimeBinaryTopology = topologyFixture();
unknownFamilyRuntimeBinaryTopology.go_targets.family_runtime_binaries[0].runtime_binary_ids = [
  "missing-binary",
];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture(
        "unknown-family-runtime-binary-topology.json",
        unknownFamilyRuntimeBinaryTopology,
      ),
    }),
  /family_runtime_binaries\[1\]\.runtime_binary_ids references unknown missing-binary/,
  "topology validation must reject unknown family-scoped runtime binaries",
);

const duplicateFamilyRuntimeBinaryTopology = topologyFixture();
duplicateFamilyRuntimeBinaryTopology.go_targets.family_runtime_binaries.splice(
  1,
  0,
  structuredClone(duplicateFamilyRuntimeBinaryTopology.go_targets.family_runtime_binaries[0]),
);
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture(
        "duplicate-family-runtime-binary-topology.json",
        duplicateFamilyRuntimeBinaryTopology,
      ),
    }),
  /family_runtime_binaries must be sorted by unique family_id/,
  "topology validation must reject duplicate family-scoped runtime binary mappings",
);

const legacyFlatTopology = topologyFixture();
legacyFlatTopology.check_schedules = renderedCheckSchedule.schedules;
assert.throws(
  () => loadExecutionTopology({ manifestPath: writeTopologyFixture("legacy-flat-topology.json", legacyFlatTopology) }),
  /flat check_schedules\[\] work_units are no longer supported/,
  "topology validation must reject the obsolete flat check schedule source shape",
);

const localInputStampTopology = topologyFixture();
localInputStampTopology.check_schedules.target_profiles["generate-drift"].local_input_stamp = {
  profile: "generate_drift",
};
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("retired-local-input-stamp-topology.json", localInputStampTopology),
    }),
  /check_schedules\.target_profiles\.generate-drift has unknown key local_input_stamp/,
  "topology validation must reject retired local_input_stamp metadata",
);

const unknownProfileTopology = topologyFixture();
unknownProfileTopology.check_schedules.target_profiles["generate-drift"].profile =
  "missing_profile";
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("unknown-check-profile-topology.json", unknownProfileTopology),
    }),
  /check_schedules\.target_profiles\.generate-drift\.profile references unknown profile missing_profile/,
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
unknownNeedTopology.check_schedules.target_profiles["backend-unit"].needs = [
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
duplicateOrderTopology.check_schedules.target_profiles["generate-drift"].order =
  250;
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("duplicate-check-order-topology.json", duplicateOrderTopology),
    }),
  /duplicate priority order drift_validation:250/,
  "topology validation must reject duplicate check schedule priority orders within a band",
);

const mismatchedSummaryTargetTopology = topologyFixture();
mismatchedSummaryTargetTopology.check_schedules.target_profiles[
  "frontend-unit"
].produces_summary_targets = ["frontend-typecheck"];
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("mismatched-summary-target-topology.json", mismatchedSummaryTargetTopology),
    }),
  /check_schedules\.target_profiles\.frontend-unit\.produces_summary_targets must include owning target frontend-unit/,
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
  /check_schedules\.target_profiles\.migration-scratch-apply target declares service_requirements and must claim a check service boundary resource/,
  "topology validation must require a boundary resource for service-backed check schedule profiles",
);

const schedulerOwnedEnvTopology = topologyFixture();
schedulerOwnedEnvTopology.check_schedules.target_profiles["lint-shell"].env = {
  CARTULARY_TEST_TARGET: "lint-shell",
};
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture("scheduler-owned-env-topology.json", schedulerOwnedEnvTopology),
    }),
  /check_schedules\.target_profiles\.lint-shell\.env\.CARTULARY_TEST_TARGET is scheduler-owned and cannot be overridden/,
  "topology validation must reject scheduler-owned check work-unit env",
);

const invalidMakePrerequisitePolicyTopology = topologyFixture();
invalidMakePrerequisitePolicyTopology.check_schedules.target_profiles[
  "backend-unit"
].make_prerequisite_policy = "sometimes";
assert.throws(
  () =>
    loadExecutionTopology({
      manifestPath: writeTopologyFixture(
        "invalid-make-prerequisite-policy-topology.json",
        invalidMakePrerequisitePolicyTopology,
      ),
    }),
  /check_schedules\.target_profiles\.backend-unit\.make_prerequisite_policy must be one of run, skip/,
  "topology validation must reject unknown check work-unit make prerequisite policies",
);
EOF
