import { validate } from "./contract.mjs";
import { recoveryGuidance } from "./failure.mjs";

// This is a terminal projection, never a run-step child. Only controller-authored
// receipts may enter retained output; transient paths never pass a tee or logger.
export function emitResult(value, mode = "summary", stream = process.stdout) {
  validate("command_result", value);
  if (mode === "machine") stream.write(`${JSON.stringify(value)}\n`);
  else if (mode !== "quiet" || value.status === "error") {
    stream.write(`[UI REVIEW] status=${value.status} state=${value.state ?? "unavailable"} session=${value.session_id ?? "unavailable"}${value.failures.length ? ` diagnostic=${value.failures[0].diagnostic_code}` : ""}\n`);
    for (const failure of value.failures) if (failure.phase !== null) {
      stream.write(`failure_class=${failure.failure_class} reason=${failure.failure_reason} phase=${failure.phase} subject=${failure.subject_id} condition=${failure.condition}; ${recoveryGuidance[failure.recovery_id]}\n`);
    }
    for (const ref of value.private_refs) stream.write(`${ref.kind}: ${ref.absolute_path}\n`);
  }
}
