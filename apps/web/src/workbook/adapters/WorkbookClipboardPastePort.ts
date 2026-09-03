import type {
  PasteWorkbookClipboardRequest,
  PasteWorkbookClipboardResponse,
} from "@cartulary/protocol-ts/http";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";

export type WorkbookClipboardPasteInput = Omit<
  PasteWorkbookClipboardRequest,
  "client_txn_id"
> & {
  readonly onClientTxnId?: ((clientTxnId: string) => void) | undefined;
};

export type WorkbookClipboardPasteAccepted = {
  readonly changeSetId: string | null;
  readonly conflicts: NonNullable<
    PasteWorkbookClipboardResponse["data"]["conflicts"]
  >;
  readonly rows: PasteWorkbookClipboardResponse["data"]["rows"];
  readonly viewSchemaId: PasteWorkbookClipboardRequest["view_schema_id"];
};

export type WorkbookClipboardPasteResult = {
  readonly clientTxnId: string | null;
  readonly outcome: WorkbookOperationOutcome<WorkbookClipboardPasteAccepted>;
};

/** Exact private transport boundary for the base Workbook paste operation. */
export interface WorkbookClipboardPastePort {
  paste(
    input: WorkbookClipboardPasteInput,
  ): Promise<WorkbookClipboardPasteResult>;
}
