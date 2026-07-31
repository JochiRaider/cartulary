import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  TimelineApiRow,
  WorkbookRow,
} from "../models/workbookTimelineModel";

export type TimelineEvidenceAttachmentAccepted = {
  readonly evidenceRecordId: string;
  readonly row: TimelineApiRow;
  readonly viewSchemaId: string;
};

export interface TimelineEvidenceAttachmentPort {
  attach(input: {
    readonly file: File;
    readonly onTimelineClientTxnId: (clientTxnId: string) => void;
    readonly target: WorkbookRow;
  }): Promise<{
    readonly clientTxnId: string | null;
    readonly outcome: WorkbookOperationOutcome<TimelineEvidenceAttachmentAccepted>;
  }>;
}
