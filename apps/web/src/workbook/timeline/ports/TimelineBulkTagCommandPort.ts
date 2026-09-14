import type { WorkbookBatchAdmission } from "../../runtime/workbookBatchOperation";
export interface TimelineBulkTagCommandPort {
  assignTag(
    input: {
      readonly tagName: string;
      readonly targets: readonly {
        readonly recordId: string;
        readonly baseRowVersion: number;
      }[];
    },
    admission: WorkbookBatchAdmission,
  ): string | null;
}
