import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import type { TimelineApiRow } from "../models/workbookTimelineModel";

export type TimelinePendingMutationAccepted = {
  readonly changeSetId: string;
  readonly row: TimelineApiRow;
  readonly viewSchemaId: string;
};

export type TimelinePendingMutationOutcome =
  WorkbookOperationOutcome<TimelinePendingMutationAccepted>;

export type TimelineResolvedConflictAccepted = Pick<
  TimelinePendingMutationAccepted,
  "row" | "viewSchemaId"
>;

export interface TimelinePendingMutationPort {
  normalizeResolvedConflict(input: {
    readonly expectedRecordId: string;
    readonly row: unknown;
    readonly viewSchemaId: string;
  }): WorkbookOperationOutcome<TimelineResolvedConflictAccepted>;
  execute(input: {
    readonly committedRowVersion: number | null;
    readonly unit: PendingReplayUnitState;
  }): Promise<TimelinePendingMutationOutcome>;
}
