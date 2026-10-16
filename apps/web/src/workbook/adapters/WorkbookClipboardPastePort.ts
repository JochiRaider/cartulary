import type { PasteWorkbookClipboardRequest } from "@cartulary/protocol-ts/http";
import type {
  WorkbookBatchAdmission,
  WorkbookBatchReceipt,
} from "../runtime/workbookBatchOperation";
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
  /** Observe one retained receipt after source settlement; never resubmit work. */
  observeReceipt(
    id: string,
    accepted: (receipt: WorkbookBatchReceipt | null) => void,
  ): () => void;
}
