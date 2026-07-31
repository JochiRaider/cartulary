import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

export type TimelineRecordActionName = "mark-reviewed" | "supersede";

export type TimelineRecordActionAccepted = {
  readonly captureState: string;
  readonly changeSetId: string;
  readonly incidentId: string;
  readonly reason: string | null;
  readonly recordId: string;
  readonly replacementRecordId: string | null;
  readonly rowVersion: number;
};

export interface TimelineRecordActionPort {
  execute(input: {
    readonly action: TimelineRecordActionName;
    readonly baseRowVersion: number;
    readonly clientTxnId: string;
    readonly recordId: string;
    readonly replacementRecordId: string | null;
  }): Promise<WorkbookOperationOutcome<TimelineRecordActionAccepted>>;
}
