import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

export type TimelineBulkTagAccepted = {
  readonly affectedRowCount: number;
  readonly changeSetId: string | null;
  readonly conflictCount: number;
};

export interface TimelineBulkTagCommandPort {
  assignTag(input: {
    readonly tagName: string;
    readonly targets: readonly {
      readonly recordId: string;
      readonly baseRowVersion: number;
    }[];
  }): Promise<WorkbookOperationOutcome<TimelineBulkTagAccepted>>;
}
