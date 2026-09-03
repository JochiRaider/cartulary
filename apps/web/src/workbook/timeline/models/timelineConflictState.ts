import type { WorkbookSameFieldConflictPayload } from "../../runtime/workbookConflictModel";

export type SameFieldConflictPayload = WorkbookSameFieldConflictPayload;

export type LocalConflictState = {
  readonly key: string;
  readonly anchor: {
    readonly record_id: string;
    readonly field_key: string;
    readonly base_row_version: number;
    readonly current_row_version?: number;
  };
  readonly conflict: SameFieldConflictPayload;
  readonly focusKey: string;
  readonly localValue: unknown;
  readonly mergedDraft: string;
};

export type PasteConflictGroupState = {
  readonly keys: string[];
};
