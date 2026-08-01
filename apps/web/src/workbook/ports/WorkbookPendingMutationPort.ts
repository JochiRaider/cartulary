import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { PendingReplayUnitState } from "../utils/workbookPendingQueue";

export type WorkbookPendingMutationAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow & { readonly view_schema_id: string };
  readonly viewSchemaId: string;
};

export interface WorkbookPendingMutationPort {
  execute(input: {
    readonly committedRowVersion: number | null;
    readonly unit: PendingReplayUnitState;
  }): Promise<WorkbookOperationOutcome<WorkbookPendingMutationAccepted>>;
}
