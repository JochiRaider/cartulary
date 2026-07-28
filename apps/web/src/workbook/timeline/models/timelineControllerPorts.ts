import type {
  GridCellAnchor,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import type {
  FocusFieldKey,
  TimelineScalarEditorSurface,
  WorkbookRow,
} from "./workbookTimelineModel";

export type TimelineMutableRef<T> = {
  current: T;
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

export type PendingReplayRuntimeMeta = {
  focusField: FocusFieldKey;
  focusKey: string;
  surface: TimelineScalarEditorSurface;
  rowSnapshot: WorkbookRow;
  continueOnFreshDraft: boolean;
  detectAutoResolution: boolean;
  promoteToCommittedRowInspect: boolean;
  viewportContinuityToken: number;
};
