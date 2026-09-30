import { ReviewFailure, failureRecord } from "./failure.mjs";
import { controlProfile, toolProfile } from "./toolchain.mjs";

process.umask(0o077);
const requested = process.argv[2];
const command = ["ui-review", "ui-review-status", "ui-browser", "ui-capture", "ui-analyze", "ui-review-report", "ui-review-stop"].includes(requested) ? requested : "ui-review";
try {
  const profile = ["ui-review-status", "ui-review-stop"].includes(command) ? controlProfile() : toolProfile();
  const { resolveInputs } = await import("./inputs.mjs");
  if (process.argv.length !== 3 || requested !== command) throw new ReviewFailure("invalid_request");
  const input = resolveInputs(command);
  const { execute } = await import("./controller.mjs");
  await execute(command, input, profile);
} catch (error) {
  // Preflight works even when an installed package is absent; no raw exception
  // message or user-provided value crosses the transient output boundary.
  const value = { schema_id: "cartulary.ui_review_command_result.v2", command_id: `cartulary.harness.command.${command.replaceAll("-", "_")}.v1`, session_id: null, operation_id: null, state: null, epoch: null, status: "error", exit_code: error instanceof ReviewFailure ? error.exitCode : 2, failures: [failureRecord(error, "tool_configuration")], receipt: null, bundle_id: null, private_refs: [] };
  process.stdout.write(`${JSON.stringify(value)}\n`);
  process.exitCode = error instanceof ReviewFailure ? error.exitCode : 2;
}
