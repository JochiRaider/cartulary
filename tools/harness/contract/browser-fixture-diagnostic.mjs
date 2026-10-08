import { validateSchemaSync } from "./harness-contract.mjs";

export const fixtureLifecycleAttachment = "cartulary-browser-fixture-lifecycle";

export function browserFixtureDiagnostic(result) {
  const attachments = (result?.attachments ?? []).filter((entry) => entry.name === fixtureLifecycleAttachment);
  if (attachments.length === 0) return null;
  try {
    if (attachments.length !== 1) throw new Error("ambiguous fixture evidence");
    const attachment = attachments[0];
    if (attachment.contentType !== "application/json" || typeof attachment.body !== "string" || attachment.body.length > 1_048_576)
      throw new Error("invalid fixture attachment");
    const diagnostic = JSON.parse(Buffer.from(attachment.body, "base64").toString("utf8"));
    validateSchemaSync("cartulary.browser_fixture_lifecycle.v1", diagnostic);
    if (diagnostic.events.some((entry) => entry.attempt_id !== diagnostic.attempt_id))
      throw new Error("fixture attempt mismatch");
    const cleanupFailed = diagnostic.process.forced || !diagnostic.process.closed || diagnostic.process.exit_code !== 0 || diagnostic.events.some((entry) => entry.phase === "cleanup" && entry.outcome === "failed");
    if (cleanupFailed && !diagnostic.failures.some((entry) => entry.phase === "cleanup"))
      throw new Error("unaccounted fixture cleanup failure");
    for (const failure of diagnostic.failures) {
      const expected = failure.phase === "cleanup" ? ["harness", "cleanup_error"] : failure.phase === "startup" ? ["harness", "tool_diagnostic_failure"] : failure.failure_class === "interrupted" ? ["interrupted", "cancelled_or_interrupted"] : ["product", "test_assertion_failure"];
      if (failure.failure_class !== expected[0] || failure.failure_reason !== expected[1])
        throw new Error("fixture failure classification mismatch");
    }
    if (diagnostic.failures.length > 0 && result.status === "passed")
      throw new Error("passing test has failed fixture evidence");
    return { diagnostic, failures: diagnostic.failures };
  } catch {
    return { diagnostic: null, failures: [{ failure_class: "artifact", failure_reason: "artifact_error", message: "Invalid browser fixture lifecycle evidence" }] };
  }
}
