// Preparation recipes only. The Go review owner acquires and releases fixtures.
import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { readFileSync, mkdirSync } from "node:fs";
import path from "node:path";
import { buildSourceSnapshot } from "../test-catalog/index.mjs";
import { coreReadiness, frontendBuildReadiness, browserReadiness, goReadiness, serviceImageReadiness } from "../readiness/installed-readiness.mjs";
import { operationBindingNames, operationDefinition, operationExecutionDigest, readOperationOwner } from "../execution/operations.mjs";
import { childEnvironment } from "../runtime/child-environment.mjs";
import { runtimeHandleEnvironment } from "../runtime/session-client.mjs";
import { runtimeBinaryEnvironment } from "../contract/runtime-binary-registry.mjs";
import { reviewChild } from "./review-child.mjs";
import { seedDesignReview } from "./design-review-seed.mjs";
import { CommandFailure } from "../runtime/command-failure.mjs";
import { ownedProcess } from "../runtime/owned-process.mjs";
import { execFileSync } from "../workspace/child-process.mjs";
import { resourceDependenciesEnvironment, runContextEnvironment } from "../contract/run-context.mjs";
import { browserBindingEnvironment } from "./browser-binding.mjs";

const preparationBuilds = Object.freeze([
  ["build-web", "frontend", "build_web"], ["embedded-web-assets", "embedded_assets", "build_server_harness"],
  ["build-server-harness", "server_harness", "build_server_harness"], ["build-migrate", "migrate", "build_migrate"],
  ["testservices-build", "test_services", "testservices_build"],
  ["test-service-images", "test_service_images", "test_service_images"], ["playwright-install", "chromium", "playwright_install"],
]);
// reviewForwarded is what review preparation forwards from its parent by name:
// the Make variables operation executions name (the operation runner resolves
// them) except public target inputs, which reach review only through its
// request, and Playwright's own browser location. Everything else a recipe
// child receives is a process or tool basic, or set below (TH-HARNESS 5.5).
export function reviewForwarded(root) {
  const manifest = JSON.parse(readFileSync(path.join(root, "tools/task_surface_manifest.json"), "utf8"));
  const publicInputs = new Set(manifest.targets.flatMap((target) => target.input_contract?.inputs ?? []).map((input) => input.name));
  return [...operationBindingNames(readOperationOwner(root)).filter((name) => !publicInputs.has(name)), "PLAYWRIGHT_BROWSERS_PATH"];
}
export function recipeEnvironment(root, runRoot, runtime, input, environment = process.env) {
  // Nothing else crosses from the parent: no inherited Go root, presentation
  // mode, public review input or harness context of an enclosing graph.
  const base = childEnvironment(environment, { pass: reviewForwarded(root), set: runtimeBinaryEnvironment(root, environment) });
  Object.assign(base, { ...resourceDependenciesEnvironment([]), CARTULARY_PREPARATION_POLICY: input.preparation_policy, 
    ...runContextEnvironment({ resultsDir: path.dirname(runRoot), runID: runtime.runID, target: input.presentation === "design" ? "browser-design-review" : "ui-review", prepared: true }), ...runtimeHandleEnvironment(runtime),
    NODE_RUNTIME_DIR: path.join(root, workspaceLayout.node_runtime), NODE_BIN: process.execPath, PNPM: path.join(root, workspaceLayout.node_runtime, "bin/pnpm"),
    ...browserBindingEnvironment({ runtimeProfile: input.profile ?? "network_flow_claimed", serviceRequirement: "test-services" }),
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
  return () => { runtime.record({ kind: "helper_process", target: proof, state: "released" }); };
}
export async function buildReviewProducer({ target, root, recipe, input, runtime, signal }) {
  const build = preparationBuilds.find(([name]) => name === target);
  if (!build || (input.preparation_policy !== "ensure" && ["test-service-images", "playwright-install"].includes(target))) throw new CommandFailure("unsupported preparation producer", { failure_class: "config", failure_reason: "usage_error" });
  if (input.preparation_policy === "installed_only" && target === "build-web") {
    coreReadiness(root); frontendBuildReadiness(root, { environment: recipe }); browserReadiness(root); recipe.GO = goReadiness(recipe, root);
  }
  const operationID = `harness.${target.replaceAll("-", "_")}`;
  const operation = operationDefinition(readOperationOwner(root), operationID);
  const [, subject_id, recovery_id] = build;
    const context = { phase: "build", subject_id, condition: "child_failed", recovery_id };
    await reviewChild({ root, command: process.execPath, args: ["tools/harness/execution/operation-runner-cli.mjs", operationID], environment: { ...recipe, ...runContextEnvironment({ operationDigest: operationExecutionDigest(operation), target }) }, signal, output: "ignore",
      context, onChildProcess: (pid) => processRegistration(runtime, pid, JSON.parse(recipe.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES)) });
  if (input.preparation_policy === "installed_only" && target === "testservices-build") serviceImageReadiness(recipe, root);
  return recipe;
}
export async function seedReview({ root, runRoot, runtime, recipe, input, sourceDigest, signal }) {
  const assignment = execFileSync(process.execPath, ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", recipe.CARTULARY_WEB_E2E_STACK_JSON_FILE], { cwd: root, env: recipe, encoding: "utf8", stdio: ["ignore", "pipe", "ignore"] });
  const attached = { ...recipe, ...JSON.parse(assignment) };
  const privateDirectory = runtime.privatePath("review-access"); mkdirSync(privateDirectory, { recursive: true, mode: 0o700 });
  let review;
  try { review = await seedDesignReview({ root, environment: attached, privateDirectory, runRoot, profile: input.profile, signal, verifySamples: input.verify_samples,
    registerSecret: (value) => runtime.registerSecret(value), onOwnedResource: (resource) => runtime.record({ ...resource, depends_on: JSON.parse(attached.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES) }) }); }
  catch (error) { if (!error.failure_reason && error !== signal.reason) Object.assign(error, { failure_class: "harness", failure_reason: "fixture_error" }); throw error; }
  if (buildSourceSnapshot(root).digest !== sourceDigest) throw new CommandFailure("source changed during preparation", { failure_class: "artifact", failure_reason: "artifact_error", phase: "build", subject_id: "source_snapshot", condition: "invalid_artifact", recovery_id: "inspect_source" });
  return { attached, privateDirectory, runRoot, review, check: () => checkReview({ root, recipe: attached, runtime, signal }) };
}
export async function checkReview({ root, recipe, runtime, signal }) {
  await reviewChild({ root, command: process.execPath, args: ["tools/harness/browser/browser-session-evidence.mjs", "attach-json", recipe.CARTULARY_WEB_E2E_STACK_JSON_FILE], environment: recipe, signal, output: "ignore", onChildProcess: (pid) => processRegistration(runtime, pid, JSON.parse(recipe.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES)), context: { phase: "service_readiness", subject_id: "browser_stack", condition: "child_failed", recovery_id: "inspect_failure" } });
}
