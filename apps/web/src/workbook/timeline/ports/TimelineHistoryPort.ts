import type {
  RecordHistoryData,
  RecordHistoryRollbackTarget,
} from "../../inspector/workbookRecordHistoryModel";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

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
    readonly target: RecordHistoryRollbackTarget;
  }): Promise<WorkbookOperationOutcome<TimelineHistoryMutationAccepted>>;
}
