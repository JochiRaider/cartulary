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
});
export class ReviewFailure extends Error {
  constructor(code, options = {}) {
    super(code, options);
    if (!Object.hasOwn(failureMappings, code)) throw new Error("unknown review diagnostic");
    this.diagnostic = code;
    this.exitCode = options.exitCode ?? failureMappings[code][2];
  }
}
export function failureRecord(error, fallback = "unsafe_artifact") {
  const code = error instanceof ReviewFailure ? error.diagnostic : fallback;
  const [failure_class, failure_reason] = failureMappings[code];
  return { failure_class, failure_reason, diagnostic_code: code };
}
