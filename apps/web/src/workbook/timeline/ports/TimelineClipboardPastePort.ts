import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  SameFieldConflictPayload,
  TimelineApiRow,
} from "../models/workbookTimelineModel";

export type TimelineClipboardPasteTarget =
  | { readonly kind: "create" }
  | {
      readonly baseRowVersion: number;
      readonly kind: "record";
      readonly recordId: string;
    };

export type TimelineClipboardPasteAccepted = {
  readonly conflicts: readonly SameFieldConflictPayload[];
  readonly rows: readonly TimelineApiRow[];
};

export interface TimelineClipboardPastePort {
  paste(input: {
    readonly clientTxnId: string;
    readonly clipboardText: string;
    readonly columns: readonly string[];
    readonly format: "csv" | "tsv";
    readonly startFieldKey: string;
    readonly targets: readonly TimelineClipboardPasteTarget[];
  }): Promise<WorkbookOperationOutcome<TimelineClipboardPasteAccepted>>;
}
