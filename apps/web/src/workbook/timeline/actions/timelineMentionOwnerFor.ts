import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { WorkbookTimelineMentionOperationOwner } from "./WorkbookTimelineMentionOperationOwner";

/** Timeline constructs its concrete owner; the workbook runtime retains only its lifecycle bridge. */
export function timelineMentionOwnerFor(runtime: WorkbookMutationRuntime) {
  return runtime.retainTimelineMentionOperations(
    (ids) =>
      new WorkbookTimelineMentionOperationOwner(runtime.scope.incidentId, ids, {
        remember: (id) => runtime.rememberClientTransaction(id),
        settle: (id) => {
          runtime.resolveSocketClientTxn(id);
        },
        accepted: (id, version) => runtime.observeTimelineVersion(id, version),
        entityAccepted: (id, version) =>
          runtime.acceptEntityVersion(id, version),
      }),
  );
}
