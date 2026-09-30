export const failureMappings = Object.freeze({
  invalid_request: ["config", "usage_error", 2],
  tool_configuration: ["config", "configuration_error", 2],
  session_mismatch: ["config", "configuration_error", 2],
  target_unavailable: ["harness", "tool_diagnostic_failure", 1],
  navigation_boundary: ["harness", "tool_diagnostic_failure", 1],
  environment_unavailable: ["infra", "preflight_error", 3],
  startup_failed: ["infra", "service_start_error", 3],
  readiness_expired: ["infra", "service_readiness_timeout", 3],
  capacity_exceeded: ["infra", "resource_conflict", 4],
  invalid_artifact: ["artifact", "artifact_error", 11],
  unstable_capture: ["artifact", "artifact_error", 11],
  observation_limit: ["artifact", "artifact_error", 11],
  analysis_failed: ["harness", "tool_diagnostic_failure", 1],
  operation_expired: ["timing", "timeout_failure", 13],
  session_lost: ["infra", "service_start_error", 3],
  unsafe_artifact: ["artifact", "artifact_error", 11],
  input_boundary: ["artifact", "artifact_error", 11],
  cleanup_failed: ["harness", "cleanup_error", 12],
  interrupted: ["interrupted", "cancelled_or_interrupted", 15],
  build_failed: ["harness", "tool_diagnostic_failure", 1],
  fixture_failed: ["harness", "fixture_error", 3],
  diagnostic_invalid: ["harness", "scheduler_accounting_error", 11],
  preparation_failed: ["unknown", "unknown_failure", 1],
});
export const preparationContext = Object.freeze({
  phase: ["prerequisites", "build", "service_acquisition", "service_readiness", "seeding", "browser_start", "cleanup"],
  subject_id: ["node", "pnpm", "frontend_dependencies", "chromium", "host_lock", "go", "docker", "test_service_images", "frontend", "embedded_assets", "server_harness", "migrate", "test_services", "browser_stack", "fixture_seed", "preparation_child", "source_snapshot", "browser"],
  condition: ["missing", "incompatible", "stale_installation", "invalid_artifact", "preflight_failed", "child_failed", "timeout", "resource_conflict", "invalid_diagnostic", "cancelled", "unknown"],
  recovery_id: ["bootstrap_node", "frontend_toolchain", "frontend_install", "playwright_install", "bootstrap", "test_service_images", "doctor", "build_web", "build_server_harness", "build_migrate", "testservices_build", "inspect_source", "inspect_failure", "exact_stop"],
});
export const recoveryGuidance = Object.freeze({
  bootstrap_node: "run make bootstrap-node-runtime",
  frontend_toolchain: "run make frontend-toolchain",
  frontend_install: "run make frontend-install",
  playwright_install: "run make playwright-install",
  bootstrap: "run make bootstrap",
  test_service_images: "run make test-service-images",
  doctor: "run make doctor and repair the reported prerequisite",
  build_web: "inspect make build-web",
  build_server_harness: "inspect make build-server-harness",
  build_migrate: "inspect make build-migrate",
  testservices_build: "inspect make testservices-build",
  inspect_source: "finish source changes, then start a new review",
  inspect_failure: "inspect the classified phase; retry only after repairing its condition",
  exact_stop: "stop through the exact UI_SESSION locator and inspect cleanup",
});
function contextFields(context) {
  if (context && Object.keys(context).some((key) => !Object.hasOwn(preparationContext, key))) throw new Error("unknown preparation context field");
  const value = {};
  for (const [key, allowed] of Object.entries(preparationContext)) {
    value[key] = context?.[key] ?? null;
    if (value[key] !== null && !allowed.includes(value[key])) throw new Error("invalid preparation context");
  }
  if (Object.values(value).some((item) => item === null) && Object.values(value).some((item) => item !== null)) throw new Error("partial preparation context");
  return value;
}
export class ReviewFailure extends Error {
  constructor(code, options = {}) {
    super(code, options);
    if (!Object.hasOwn(failureMappings, code)) throw new Error("unknown review diagnostic");
    this.diagnostic = code;
    this.exitCode = options.exitCode ?? failureMappings[code][2];
    this.context = contextFields(options.context);
    if (["build_failed", "fixture_failed", "diagnostic_invalid", "preparation_failed"].includes(code) && this.context.phase === null) throw new Error("preparation context required");
  }
}
export function failureRecord(error, fallback = "unsafe_artifact") {
  const code = error instanceof ReviewFailure ? error.diagnostic : fallback;
  const [failure_class, failure_reason] = failureMappings[code];
  return { failure_class, failure_reason, diagnostic_code: code, ...contextFields(error instanceof ReviewFailure ? error.context : null) };
}

// Only owner-authored structured fields cross this boundary. Error messages,
// stacks, process output and private artifact references never do.
export function preparationFailure(error, context = { phase: "prerequisites", subject_id: "preparation_child", condition: "unknown", recovery_id: "inspect_failure" }) {
  if (error instanceof ReviewFailure && error.context.phase !== null) return error;
  const mapping = error instanceof ReviewFailure ? error.diagnostic : Object.keys(failureMappings).find((code) => failureMappings[code][0] === error?.failure_class && failureMappings[code][1] === error?.failure_reason);
  const fields = { ...context, ...Object.fromEntries(Object.keys(preparationContext).filter((key) => error?.[key] !== undefined).map((key) => [key, error[key]])) };
  if (error?.failure_reason === "service_readiness_timeout") fields.phase = "service_readiness";
  if (error?.failure_reason === "timeout_failure" || error?.failure_reason === "service_readiness_timeout") fields.condition = "timeout";
  if (error?.failure_reason === "resource_conflict") fields.condition = "resource_conflict";
  if (error?.failure_reason === "scheduler_accounting_error") fields.condition = "invalid_diagnostic";
  if (error?.failure_reason === "cleanup_error") Object.assign(fields, { phase: "cleanup", condition: "child_failed", recovery_id: "exact_stop" });
  if (!mapping || error?.failure_reason === "unknown_failure") fields.condition = "unknown";
  const code = error?.failure_reason === "tool_diagnostic_failure" ? "build_failed" : mapping ?? "preparation_failed";
  const value = new ReviewFailure(code, { cause: error, context: fields, ...(error instanceof ReviewFailure ? { exitCode: error.exitCode } : {}) });
  if (error?.cleanupFailures?.length) value.cleanupFailures = error.cleanupFailures;
  return value;
}
export function failureFromRecord(record) {
  const expected = failureMappings[record?.diagnostic_code];
  if (!expected || expected[0] !== record.failure_class || expected[1] !== record.failure_reason || Object.keys(record).sort().join(",") !== "condition,diagnostic_code,failure_class,failure_reason,phase,recovery_id,subject_id") throw new Error("invalid failure record");
  return new ReviewFailure(record.diagnostic_code, { context: contextFields(Object.fromEntries(Object.keys(preparationContext).map((key) => [key, record[key]]))) });
}
