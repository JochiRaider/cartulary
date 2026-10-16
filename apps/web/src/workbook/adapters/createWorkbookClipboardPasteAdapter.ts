import type { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import type { WorkbookClipboardPastePort } from "./WorkbookClipboardPastePort";
export function createWorkbookClipboardPasteAdapter(
  owner: Pick<
    WorkbookBatchOperationOwner,
    "admit" | "subscribe" | "getSnapshot"
  >,
): WorkbookClipboardPastePort {
  return {
    observeReceipt(id, accepted) {
      let active = true;
      let unsubscribe = () => {};
      const observe = () => {
        const entry = owner
          .getSnapshot()
          .entries.find((candidate) => candidate.id === id);
        const receipt = entry?.receipt ?? null;
        if (entry && entry.phase !== "rejected" && receipt === null) return;
        unsubscribe();
        // Receipt publication precedes source application in the retained
        // owner. Presentation follows that synchronous settlement, and can be
        // cancelled independently without changing the operation's outcome.
        queueMicrotask(() => {
          if (!active) return;
          active = false;
          accepted(receipt);
        });
      };
      unsubscribe = owner.subscribe(observe);
      observe();
      return () => {
        active = false;
        unsubscribe();
      };
    },
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
