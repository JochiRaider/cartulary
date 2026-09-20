import type {
  WorkbookBatchAdmission,
  WorkbookBatchSnapshot,
} from "../../runtime/workbookBatchOperation";
export type TimelineBulkTagAdmission =
  | { readonly kind: "admitted"; readonly operationId: string }
  | { readonly kind: "rejected"; readonly message: string };
export interface TimelineBulkTagCommandPort {
  readonly subscribe: (listener: () => void) => () => void;
  readonly getSnapshot: () => WorkbookBatchSnapshot;
  assignTag(
    input: {
      readonly tagName: string;
      readonly targets: readonly {
        readonly recordId: string;
        readonly baseRowVersion: number;
      }[];
    },
    admission: WorkbookBatchAdmission,
  ): TimelineBulkTagAdmission;
}
