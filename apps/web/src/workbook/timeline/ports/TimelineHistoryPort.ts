import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { RecordHistoryData } from "../models/timelineHistoryModel";

export type TimelineHistoryMutationAccepted = {
  readonly recordId: string;
  readonly rowVersion: number;
};

export interface TimelineHistoryPort {
  load(input: {
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<RecordHistoryData>>;
  deleteOrRestore(input: {
    readonly baseRowVersion: number;
    readonly clientTxnId: string;
    readonly operation: "delete" | "restore";
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<TimelineHistoryMutationAccepted>>;
  rollback(input: {
    readonly baseRowVersion: number;
    readonly clientTxnId: string;
    readonly recordId: string;
    readonly target: Record<string, unknown>;
  }): Promise<WorkbookOperationOutcome<TimelineHistoryMutationAccepted>>;
}
