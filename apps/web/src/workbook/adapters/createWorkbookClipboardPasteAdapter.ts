import type { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import type { WorkbookClipboardPastePort } from "./WorkbookClipboardPastePort";
export function createWorkbookClipboardPasteAdapter(
  owner: Pick<WorkbookBatchOperationOwner, "admit">,
): WorkbookClipboardPastePort {
  return {
    paste(input, admission) {
      const entityType =
        input.view_schema_id === "cartulary.view.hosts.v1"
          ? "host"
          : input.view_schema_id === "cartulary.view.identities.v1"
            ? "identity"
            : undefined;
      return owner.admit(
        {
          operation: "pasteWorkbookClipboard",
          request: input,
          recordIds: input.targets.flatMap((target) =>
            target.kind === "record" ? [target.record_id] : [],
          ),
          ...(entityType ? { entityType } : {}),
        },
        admission,
      );
    },
  };
}
