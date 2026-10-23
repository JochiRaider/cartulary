// Preparation recipes only. The Go review owner acquires and releases fixtures.
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { readFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import { buildSourceSnapshot } from "../test-catalog/index.mjs";
import { coreReadiness, frontendBuildReadiness, browserReadiness, goReadiness, serviceImageReadiness } from "../readiness/installed-readiness.mjs";
import { operationDefinition, operationExecutionDigest, readOperationOwner } from "../scheduler/work-graph/operations.mjs";
import { reviewChild } from "./review-child.mjs";
import { seedDesignReview } from "./design-review-seed.mjs";
import { CommandFailure } from "../runtime/command-failure.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";
import { execFileSync } from "../workspace/child-process.mjs";

const preparationBuilds = Object.freeze([
  ["build-web", "frontend", "build_web"], ["embedded-web-assets", "embedded_assets", "build_server_harness"],
  ["build-server-harness", "server_harness", "build_server_harness"], ["build-migrate", "migrate", "build_migrate"],
  ["testservices-build", "test_services", "testservices_build"],
  ["test-service-images", "test_service_images", "test_service_images"], ["playwright-install", "chromium", "playwright_install"],
]);
export function recipeEnvironment(root, runRoot, runtime, input, environment = process.env) {
  const base = { ...environment };
  // A Go test parent may export its own toolchain root. Resolve from GO instead.
  delete base.GOROOT;
  const topology = JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json")));
  for (const binary of topology.runtime_binaries) base[binary.output_make_variable] = path.resolve(root, base[binary.output_make_variable] || binary.default_output_path);
  for (const name of Object.keys(base)) if (/^(?:MAKEFLAGS|MAKEOVERRIDES|MFLAGS|REVIEW_PROFILE|UI_|OTEL_|CARTULARY_(?:MAKE_|TEST_|WEB_E2E_|BROWSER_|HARNESS_|PGTEST_|S3_)|CARTULARY__)/u.test(name)) delete base[name];
  Object.assign(base, { CARTULARY_HARNESS_RESOURCE_DEPENDENCIES: "[]", CARTULARY_PREPARATION_POLICY: input.preparation_policy, CARTULARY_TEST_RESULTS_DIR: path.dirname(runRoot), CARTULARY_TEST_RUN_ID: runtime.runID,
    CARTULARY_TEST_TARGET: input.presentation === "design" ? "browser-design-review" : "ui-review", CARTULARY_HARNESS_IDENTITY_PREPARED: "1", CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: runtime.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: runtime.leaseID, CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runtime.runID,
    CARTULARY_SERVER_HARNESS_BIN: path.join(root, workspaceLayout.binaries, "server-harness"), CARTULARY_MIGRATE_BIN: path.join(root, workspaceLayout.binaries, "migrate"),
    CARTULARY_TEST_SERVICES_BIN: path.join(root, workspaceLayout.toolbin, "cartulary-test-services"), CARTULARY_HARNESS_RUNTIME_BIN: path.join(root, workspaceLayout.toolbin, "cartulary-harness"),
    NODE_RUNTIME_DIR: path.join(root, workspaceLayout.node_runtime), NODE_BIN: process.execPath, PNPM: path.join(root, workspaceLayout.node_runtime, "bin/pnpm"),
    CARTULARY_BROWSER_RUNTIME_PROFILE_ID: input.profile ?? "network_flow_claimed", CARTULARY_BROWSER_SERVICE_REQUIREMENT: "test-services",
    CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED: "false", CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED: "false" });
  if (input.preparation_policy === "installed_only") {
    const pins = JSON.parse(readFileSync(path.join(root, "tools/toolchain_pins.json")));
    const machine = environment.CARTULARY_MACHINE_CACHE_DIR || path.join(environment.XDG_CACHE_HOME || path.join(environment.HOME, ".cache"), "cartulary");
    Object.assign(base, { GO: environment.GO || "go", GO_TOOLCHAIN: pins.go_toolchain,
      GO_CACHE_DIR: environment.GO_CACHE_DIR || path.join(machine, "go/build"), GO_MOD_CACHE_DIR: environment.GO_MOD_CACHE_DIR || path.join(machine, "go/mod"), GO_TMP_DIR: environment.GO_TMP_DIR || path.join(machine, "go/tmp"),
      GOTOOLCHAIN: "local", GOPROXY: "off", GONOPROXY: "none", GOSUMDB: "off", GOTELEMETRY: "off", COREPACK_ENABLE_NETWORK: "0", npm_config_offline: "true" });
  }
  return base;
}
function processRegistration(runtime, pid, dependencies) {
  const proof = ownedProcess(pid); runtime.record({ kind: "helper_process", target: proof, depends_on: dependencies });
  return async () => { await stopOwnedProcess(proof); runtime.record({ kind: "helper_process", target: proof, state: "released" }); };
}
export async function buildReviewProducer({ target, root, environment, input, runtime, signal }) {
  const recipe = preparationBuilds.find(([name]) => name === target);
  if (!recipe || (input.preparation_policy !== "ensure" && ["test-service-images", "playwright-install"].includes(target))) throw new CommandFailure("unsupported preparation producer", { failure_class: "config", failure_reason: "usage_error" });
  if (input.preparation_policy === "installed_only" && target === "build-web") {
    coreReadiness(root); frontendBuildReadiness(root, { environment }); browserReadiness(root); environment.GO = goReadiness(environment, root);
  }
  const operationID = `harness.${target.replaceAll("-", "_")}`;
  const operation = operationDefinition(readOperationOwner(root), operationID);
  const [, subject_id, recovery_id] = recipe;
    const context = { phase: "build", subject_id, condition: "child_failed", recovery_id };
    await reviewChild({ root, command: process.execPath, args: ["tools/harness/execution/operation-runner-cli.mjs", operationID], environment: { ...environment, CARTULARY_OPERATION_DIGEST: operationExecutionDigest(operation), CARTULARY_TEST_TARGET: target,
      CARTULARY_HARNESS_GRAPH_CHILD: "1", CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD: "1", CARTULARY_HARNESS_SKIP_PREREQUISITES: "1" }, signal, output: "ignore",
      context, onChildProcess: (pid) => processRegistration(runtime, pid, JSON.parse(environment.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES)) });
  if (input.preparation_policy === "installed_only" && target === "testservices-build") serviceImageReadiness(environment, root);
  return environment;
}
export async function seedReview({ root, runRoot, runtime, environment, input, sourceDigest, signal }) {
  const assignment = execFileSync(process.execPath, ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", environment.CARTULARY_WEB_E2E_STACK_JSON_FILE], { cwd: root, env: environment, encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] });
  const attached = { ...environment, ...JSON.parse(assignment) };
  const privateDirectory = runtime.privatePath("review-access"); mkdirSync(privateDirectory, { recursive: true, mode: 0o700 });
  let review;
  try { review = await seedDesignReview({ root, environment: attached, privateDirectory, runRoot, profile: input.profile, signal, verifySamples: input.verify_samples,
    registerSecret: (value) => runtime.registerSecret(value), onOwnedResource: (resource) => runtime.record({ ...resource, depends_on: JSON.parse(attached.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES) }) }); }
  catch (error) { if (!error.failure_reason && error !== signal.reason) Object.assign(error, { failure_class: "harness", failure_reason: "fixture_error" }); throw error; }
  if (buildSourceSnapshot(root).digest !== sourceDigest) throw new CommandFailure("source changed during preparation", { failure_class: "artifact", failure_reason: "artifact_error", phase: "build", subject_id: "source_snapshot", condition: "invalid_artifact", recovery_id: "inspect_source" });
  return { attached, privateDirectory, runRoot, review, check: () => checkReview({ root, environment: attached, runtime, signal }) };
}
export async function checkReview({ root, environment, runtime, signal }) {
  await reviewChild({ root, command: process.execPath, args: ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", environment.CARTULARY_WEB_E2E_STACK_JSON_FILE], environment, signal, output: "ignore", onChildProcess: (pid) => processRegistration(runtime, pid, JSON.parse(environment.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES)), context: { phase: "service_readiness", subject_id: "browser_stack", condition: "child_failed", recovery_id: "inspect_failure" } });
}
