import { primaryPublicFailure, publicExitCodeForFailure } from "../contract/index.mjs";
import { CommandFailure, reportCommandFailure } from "../runtime/command-failure.mjs";

// The retained action receipts and the owning operation share one primary
// cause. A generic child exit must not turn a finalization failure into a test
// assertion when it crosses the private command boundary.
export function publishFinalizationFailure(repoRoot, failures, report = reportCommandFailure) {
  const primary = primaryPublicFailure(failures, {
    failure_class: "unknown", failure_reason: "unknown_failure",
  });
  const failure = report(repoRoot, new CommandFailure("finalization failed", primary), primary);
  return publicExitCodeForFailure(failure);
}
