import { lifecyclePolicy } from "../../contract/generated/lifecycle-policy.mjs";
export const failureMappings = Object.freeze(lifecyclePolicy.review_failures);
export const preparationContext = Object.freeze(lifecyclePolicy.preparation_context);
export const recoveryGuidance = Object.freeze({
  bootstrap_node: "run make bootstrap PART=node",
  frontend_toolchain: "run make bootstrap PART=frontend",
  frontend_install: "run make bootstrap PART=frontend",
  playwright_install: "run make bootstrap PART=browser",
  bootstrap: "run make bootstrap",
  test_service_images: "run make bootstrap PART=test-images",
  doctor: "run make doctor and repair the reported prerequisite",
  build_web: "inspect make build COMPONENT=web",
  build_server_harness: "inspect make build COMPONENT=server-harness",
  build_migrate: "inspect make build COMPONENT=migrate",
  testservices_build: "inspect make bootstrap",
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
