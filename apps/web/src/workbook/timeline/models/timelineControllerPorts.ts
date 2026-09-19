import type {
  GridCellAnchor,
  GridEditCommitOutcome,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import type { SheetRef } from "../../../shared/sheetRef";
import type {
  CollectionDraftKey,
  CollectionFieldKey,
  FocusFieldKey,
  RowValues,
  TimelineScalarEditorSurface,
} from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineMutableRef<T> = {
  current: T;
};

export type TimelineRowsUpdater = (current: WorkbookRow[]) => WorkbookRow[];

export type TimelineCommittedRecordIdleResult = {
  readonly row: WorkbookRow | null;
  readonly rowVersion: number;
};

export type TimelineCommittedInspectorRecords = {
  readonly currentCommittedTimelineRow: (
    recordId: string,
  ) => WorkbookRow | null;
  readonly retainInspectorRecord: (recordId: string | null) => void;
};

export type TimelineRowContextMenuPosition = {
  readonly x: number;
  readonly y: number;
};

export type TimelineRowStoreCommands = {
  readonly replaceRows: (rows: WorkbookRow[]) => void;
  readonly updateRows: (updater: TimelineRowsUpdater) => void;
};

export type TimelinePasteTargetResolution = {
  readonly anchor: GridCellAnchor | null;
  readonly targetResolution: GridPasteTargetResolution;
};

export type TimelineScalarSaveOptions = {
  readonly allowZeroFieldCreate?: boolean | undefined;
  readonly continueOnFreshDraft: boolean;
  readonly preserveInputFocus: boolean;
  readonly surface: TimelineScalarEditorSurface;
};

export type TimelineRowMutationEditorPort = {
  readonly activateEdit: (input: {
    readonly fieldKey: string;
    readonly recordId: string;
    readonly value: unknown;
    readonly selectionRange?:
      | { readonly start: number; readonly end: number }
      | undefined;
  }) => void;
  readonly cancelEdit: (input: {
    readonly fieldKey: string;
    readonly recordId: string;
  }) => void;
  readonly focus: (input: {
    readonly fieldKey: string;
    readonly recordId: string;
  }) => void;
  readonly focusInput: (focusKey: string) => void;
  readonly reveal: (input: {
    readonly fieldKey: string;
    readonly recordId: string;
  }) => void;
};

export type TimelineReplayContext = {
  draftRevisions?: ReadonlyMap<string, number>;
  sheetRef: SheetRef;
  focusField: FocusFieldKey;
  focusKey: string;
  surface: TimelineScalarEditorSurface;
  rowSnapshot: WorkbookRow;
  continueOnFreshDraft: boolean;
  detectAutoResolution: boolean;
  promoteToCommittedRowInspect: boolean;
  viewportContinuityToken: number | undefined;
};

/** Source-owned editor settlement, shared by keyboard departure and local Find. */
export type TimelineQueueScalarSave = (
  rowKey: string,
  field: keyof RowValues,
  options: TimelineScalarSaveOptions,
  currentValue?: string,
  onSettled?: (outcome: GridEditCommitOutcome) => void,
) => void;
export type TimelineQueueCollectionSave = (
  rowKey: string,
  field: CollectionFieldKey,
  draft: CollectionDraftKey,
  currentValue?: string,
  surface?: TimelineScalarEditorSurface,
  onSettled?: (outcome: GridEditCommitOutcome) => void,
) => void;
