import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { WorkbookTimelineCaptureActionOwner } from "./WorkbookTimelineCaptureActionOwner";

/** Concrete Timeline construction stays outside the common runtime's import boundary. */
export function timelineCaptureOwnerFor(runtime: WorkbookMutationRuntime) {
  return runtime.retainTimelineActions(
    (ids) =>
      new WorkbookTimelineCaptureActionOwner(runtime.scope.incidentId, ids, {
        remember: (id) => runtime.rememberClientTransaction(id),
        settle: (id) => {
          runtime.resolveSocketClientTxn(id);
        },
        accepted: (id, version) => runtime.history.acceptVersion(id, version),
      }),
  );
}
