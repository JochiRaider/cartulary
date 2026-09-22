import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { PendingReplayUnitState } from "../runtime/pending/workbookPendingQueue";

export type WorkbookPendingMutationAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow & { readonly view_schema_id: string };
  readonly viewSchemaId: string;
};

export interface WorkbookPendingMutationPort {
  retire?(): void;
  execute(input: {
    readonly committedRowVersion: number | null;
    readonly unit: PendingReplayUnitState;
  }): Promise<WorkbookOperationOutcome<WorkbookPendingMutationAccepted>>;
}
