import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { TimelineApiRow, WorkbookRow } from "../models/timelineRowModel";

export type TimelineEvidenceAttachmentAccepted = {
  readonly evidenceRecordId: string;
  readonly row: TimelineApiRow;
  readonly viewSchemaId: string;
};

export type TimelineEvidenceCreated = {
  readonly evidenceRecordId: string;
};

export interface TimelineEvidenceAttachmentPort {
  createEvidence(input: {
    readonly file: File;
  }): Promise<WorkbookOperationOutcome<TimelineEvidenceCreated>>;
  attachEvidence(input: {
    readonly evidenceRecordId: string;
    readonly onTimelineClientTxnId: (clientTxnId: string) => void;
    readonly target: WorkbookRow;
  }): Promise<{
    readonly clientTxnId: string | null;
    readonly outcome: WorkbookOperationOutcome<TimelineEvidenceAttachmentAccepted>;
  }>;
}
