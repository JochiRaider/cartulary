import type { PasteWorkbookClipboardRequest } from "@cartulary/protocol-ts/http";
import type { WorkbookBatchAdmission } from "../runtime/workbookBatchOperation";
export type WorkbookClipboardPasteInput = Omit<
  PasteWorkbookClipboardRequest,
  "client_txn_id"
>;
/** Source-owned paste planning enters the retained Workbook batch lifecycle. */
export interface WorkbookClipboardPastePort {
  paste(
    input: WorkbookClipboardPasteInput,
    admission: WorkbookBatchAdmission,
  ): string | null;
}
